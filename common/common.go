package common

import (
	"fmt"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/go-redis/redis/v8"
	"github.com/nats-io/nats.go"
	"github.com/weplanx/transfer"
	"github.com/weplanx/utils/passport"
	"github.com/weplanx/utils/sessions"
	"go.mongodb.org/mongo-driver/mongo"
	"strings"
	"time"
)

type Inject struct {
	Values    *Values
	Mongo     *mongo.Client
	Db        *mongo.Database
	Redis     *redis.Client
	Nats      *nats.Conn
	JetStream nats.JetStreamContext
	KeyValue  nats.KeyValue
	Transfer  *transfer.Transfer
}

type Values struct {
	// 监听地址
	Address string `env:"ADDRESS" envDefault:":3000"`

	// 命名空间
	Namespace string `env:"NAMESPACE,required"`

	// 密钥
	Key string `env:"KEY,required"`

	// 跨域
	Hosts []string `env:"HOSTS" envSeparator:","`

	// 数据库
	Database `envPrefix:"DATABASE_"`

	// NATS 配置
	Nats `envPrefix:"NATS_"`

	// 动态配置
	DynamicValues `env:"-"`

	SessionsValues *sessions.Values `env:"-"`
}

type Database struct {
	// MongoDB 连接 Uri
	Mongo string `env:"MONGO,required"`

	// MongoDB 数据库名称
	Name string `env:"NAME,required"`

	// Redis 连接 Uri
	Redis string `env:"REDIS,required"`
}

type Nats struct {
	// Nats 连接地址
	Hosts []string `env:"HOSTS,required" envSeparator:","`

	// Nats Nkey 认证
	Nkey string `env:"NKEY,required"`
}

// Name 生成空间名称
func (x Values) Name(v ...string) string {
	return fmt.Sprintf(`%s:%s`, x.Namespace, strings.Join(v, ":"))
}

// Subject 生成主题名称
func (x Values) Subject(v string) string {
	return fmt.Sprintf(`%s.events.%s`, x.Namespace, v)
}

// Queue 生成队列名称
func (x Values) Queue(v string) string {
	return fmt.Sprintf(`%s:events:%s`, x.Namespace, v)
}

// DynamicValues 动态配置
type DynamicValues struct {
	//// 会话周期（秒）
	//// 用户在 1 小时 内没有操作，将结束会话。
	//SessionTTL time.Duration `json:"session_ttl"`
	// 登录锁定时间
	// 锁定 15 分钟。
	LoginTTL time.Duration `json:"login_ttl"`
	// 用户最大登录失败次数
	// 有限时间（锁定时间）内连续登录失败 5 次，锁定帐号。
	LoginFailures int64 `json:"login_failures"`
	// IP 最大登录失败次数
	// 同 IP 连续 10 次登录失败后，锁定 IP（周期为锁定时间）。
	IpLoginFailures int64 `json:"ip_login_failures"`
	// IP 白名单
	// 白名单 IP 允许超出最大登录失败次数。
	IpWhitelist []string `json:"ip_whitelist"`
	// GetIpBlacklist IP 黑名单
	// 黑名单 IP 将禁止访问。
	IpBlacklist []string `json:"ip_blacklist"`
	// 密码强度
	// 0：无限制；
	// 1：需要大小写字母；
	// 2：需要大小写字母、数字；
	// 3：需要大小写字母、数字、特殊字符
	PwdStrategy int `json:"pwd_strategy"`
	// 密码有效期（天）
	// 密码过期后强制要求修改密码，0：永久有效
	PwdTTL time.Duration `json:"pwd_ttl"`
	// 云平台
	// tencent：腾讯云；
	Cloud string `json:"cloud"`
	// 腾讯云 API 密钥 Id
	// 建议用子账号分配需要的权限
	TencentSecretId string `json:"tencent_secret_id"`
	// 腾讯云 API 密钥 Key
	TencentSecretKey string `json:"tencent_secret_key,omitempty"`
	// 腾讯云 COS 对象存储 Bucket（存储桶名称）
	TencentCosBucket string `json:"tencent_cos_bucket,omitempty"`
	// 腾讯云 COS 对象存储所属地域，例如：ap-guangzhou
	TencentCosRegion string `json:"tencent_cos_region"`
	// 腾讯云 COS 对象存储预签名有效期，单位：秒
	TencentCosExpired int `json:"tencent_cos_expired"`
	// 腾讯云 COS 对象存储上传大小限制，单位：KB
	TencentCosLimit int64 `json:"tencent_cos_limit"`
	// 办公平台
	// feishu：飞书；
	Office string `json:"office"`
	// 飞书应用 ID
	FeishuAppId string `json:"feishu_app_id"`
	// 飞书应用密钥
	FeishuAppSecret string `json:"feishu_app_secret,omitempty"`
	// 飞书事件订阅安全校验数据密钥
	FeishuEncryptKey string `json:"feishu_encrypt_key,omitempty"`
	// 飞书事件订阅验证令牌
	FeishuVerificationToken string `json:"feishu_verification_token,omitempty"`
	// 第三方免登授权码跳转地址
	RedirectUrl string `json:"redirect_url"`
	// 公共电子邮件服务 SMTP 地址
	EmailHost string `json:"email_host"`
	// SMTP 端口号（SSL）
	EmailPort string `json:"email_port"`
	// 公共邮箱用户
	EmailUsername string `json:"email_username"`
	// 公共邮箱用户
	EmailPassword string `json:"email_password,omitempty"`
	// 开放服务地址
	OpenapiUrl string `json:"openapi_url"`
	// 开放服务应用认证 Key
	// API 网关应用认证方式 https://cloud.tencent.com/document/product/628/55088
	OpenapiKey string `json:"openapi_key"`
	// 开放服务应用认证密钥
	OpenapiSecret string `json:"openapi_secret,omitempty"`
}

// GetClaims 获取授权标识
func GetClaims(c *app.RequestContext) (claims passport.Claims) {
	value, ok := c.Get("identity")
	if !ok {
		return
	}
	return value.(passport.Claims)
}
