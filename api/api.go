package api

import (
	"context"
	"github.com/bytedance/go-tagexpr/v2/binding"
	"github.com/bytedance/go-tagexpr/v2/validator"
	"github.com/bytedance/gopkg/util/logger"
	"github.com/bytedance/sonic/decoder"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/errors"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol"
	"github.com/google/wire"
	"github.com/weplanx/server/api/index"
	"github.com/weplanx/server/api/sessions"
	"github.com/weplanx/server/api/values"
	"github.com/weplanx/server/common"
	"github.com/weplanx/transfer"
	"github.com/weplanx/utils/dsl"
	"github.com/weplanx/utils/helper"
	"net/http"
	"time"
)

var Provides = wire.NewSet(
	index.Provides,
	values.Provides,
	sessions.Provides,
	dsl.Provides,
)

type API struct {
	*common.Inject

	Hertz              *server.Hertz
	IndexController    *index.Controller
	IndexService       *index.Service
	ValuesController   *values.Controller
	ValuesService      *values.Service
	SessionsController *sessions.Controller
	SessionsService    *sessions.Service
	DSL                *dsl.Controller
}

func (x *API) Routes(h *server.Hertz) (err error) {
	auth := x.AuthGuard()
	h.GET("", x.IndexController.Index)
	h.POST("login", x.IndexController.Login)
	h.GET("verify", x.IndexController.Verify)
	h.GET("code", auth, x.IndexController.GetRefreshCode)
	h.POST("refresh_token", auth, x.IndexController.RefreshToken)
	h.POST("logout", auth, x.IndexController.Logout)
	h.GET("navs", auth, x.IndexController.GetNavs)
	h.GET("options", auth, x.IndexController.GetOptions)

	_user := h.Group("user", auth)
	{
		_user.GET("", x.IndexController.GetUser)
		_user.PATCH("", x.IndexController.SetUser)
	}

	_values := h.Group("values", auth)
	{
		_values.GET("", x.ValuesController.Get)
		_values.PATCH("", x.ValuesController.Set)
		_values.DELETE(":key", x.ValuesController.Remove)
	}

	_sessions := h.Group("sessions", auth)
	{
		_sessions.GET("", x.SessionsController.Lists)
		_sessions.DELETE(":uid", x.SessionsController.Remove)
		_sessions.DELETE("", x.SessionsController.Clear)
	}

	helper.BindDSL(h.Group("/:collection", auth), x.DSL)

	return
}

// AuthGuard 认证中间件
func (x *API) AuthGuard() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		ts := c.Cookie("access_token")
		if ts == nil {
			c.AbortWithStatusJSON(401, utils.H{
				"code":    0,
				"message": "认证已失效请重新登录",
			})
			return
		}

		claims, err := x.IndexService.Verify(ctx, string(ts))
		if err != nil {
			c.SetCookie("access_token", "", -1, "/", "", protocol.CookieSameSiteStrictMode, true, true)
			c.AbortWithStatusJSON(401, utils.H{
				"code":    0,
				"message": "认证已失效请重新登录",
			})
			return
		}

		c.Set("identity", claims)
		c.Next(ctx)
	}
}

// AccessLogs 日志
func (x *API) AccessLogs() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		c.Next(ctx)
		end := time.Now()
		latency := end.Sub(start).Microseconds
		x.Transfer.Publish(context.Background(), "access", transfer.Payload{
			Metadata: map[string]interface{}{
				"method": string(c.Request.Header.Method()),
				"host":   string(c.Request.Host()),
				"path":   string(c.Request.Path()),
				"status": c.Response.StatusCode(),
				"ip":     c.ClientIP(),
			},
			Data: map[string]interface{}{
				"user_agent": string(c.Request.Header.UserAgent()),
				"query":      c.Request.QueryString(),
				"body":       string(c.Request.Body()),
				"cost":       latency(),
			},
			Timestamp: start,
		})
	}
}

// ErrHandler 错误处理
func (x *API) ErrHandler() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		c.Next(ctx)
		err := c.Errors.Last()
		if err == nil {
			return
		}

		if err.IsType(errors.ErrorTypePublic) {
			statusCode := http.StatusBadRequest
			result := utils.H{"message": err.Error()}
			if meta, ok := err.Meta.(map[string]interface{}); ok {
				if meta["statusCode"] != nil {
					statusCode = meta["statusCode"].(int)
				}
				if meta["code"] != nil {
					result["code"] = meta["code"]
				}
			}
			c.JSON(statusCode, result)
			return
		}

		switch e := err.Err.(type) {
		case decoder.SyntaxError:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Description(),
			})
			break
		case *binding.Error:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Error(),
			})
			break
		case *validator.Error:
			c.JSON(http.StatusBadRequest, utils.H{
				"message": e.Error(),
			})
			break
		default:
			logger.Error(err)
			c.Status(http.StatusInternalServerError)
		}
	}
}

// Initialize 初始化
func (x *API) Initialize(ctx context.Context) (h *server.Hertz, err error) {
	h = x.Hertz
	h.Use(x.AccessLogs())
	h.Use(x.ErrHandler())
	// 加载自定义验证
	helper.RegValidate()
	// 订阅动态配置
	go x.ValuesService.Sync(ctx)
	// 传输指标
	if err = x.Transfer.Set(ctx, transfer.Option{
		Key:         "access",
		Description: "请求日志",
		TTL:         15552000,
	}); err != nil {
		return
	}
	return
}
