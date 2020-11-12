package role

import (
	"github.com/gin-gonic/gin"
	"github.com/kainonly/gin-curd/operates"
	"github.com/kainonly/gin-curd/typ"
	"github.com/kainonly/gin-extra/helper/res"
	"gorm.io/gorm"
	"taste-api/application/cache"
	"taste-api/application/common"
	"taste-api/application/common/types"
	"taste-api/application/model"
)

type Controller struct {
	*common.Dependency
}

type originListsBody struct {
	operates.OriginListsBody
}

func (c *Controller) OriginLists(ctx *gin.Context) interface{} {
	var body originListsBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return res.Error(err)
	}
	return c.Curd.
		Originlists(model.Role{}, body.OriginListsBody).
		OrderBy(typ.Orders{"create_time": "desc"}).
		Exec()
}

type listsBody struct {
	operates.ListsBody
}

func (c *Controller) Lists(ctx *gin.Context) interface{} {
	var body listsBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return res.Error(err)
	}
	return c.Curd.
		Lists(model.Role{}, body.ListsBody).
		OrderBy(typ.Orders{"create_time": "desc"}).
		Exec()
}

type getBody struct {
	operates.GetBody
}

func (c *Controller) Get(ctx *gin.Context) interface{} {
	var body getBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return res.Error(err)
	}
	return c.Curd.
		Get(model.Role{}, body.GetBody).
		Exec()
}

type addBody struct {
	Key      string     `binding:"required"`
	Name     types.JSON `binding:"required"`
	Resource []string   `binding:"required"`
	Note     string
	Status   bool
}

func (c *Controller) Add(ctx *gin.Context) interface{} {
	var body addBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return res.Error(err)
	}
	data := model.RoleBasic{
		Key:    body.Key,
		Name:   body.Name,
		Note:   body.Note,
		Status: body.Status,
	}
	return c.Curd.
		Add().
		After(func(tx *gorm.DB) error {
			var assoc []model.RoleResourceAssoc
			for _, resourceKey := range body.Resource {
				assoc = append(assoc, model.RoleResourceAssoc{
					RoleKey:     body.Key,
					ResourceKey: resourceKey,
				})
			}
			err = tx.Create(&assoc).Error
			if err != nil {
				return err
			}
			clearcache(c.Cache)
			return nil
		}).
		Exec(&data)
}

type editBody struct {
	operates.EditBody
	Key      string     `binding:"required_if=switch false"`
	Name     types.JSON `binding:"required_if=switch false"`
	Resource []string   `binding:"required_if=switch false"`
	Note     string
	Status   bool
}

func (c *Controller) Edit(ctx *gin.Context) interface{} {
	var body editBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return res.Error(err)
	}
	data := model.RoleBasic{
		Key:    body.Key,
		Name:   body.Name,
		Note:   body.Note,
		Status: body.Status,
	}
	return c.Curd.
		Edit(model.Resource{}, body.EditBody).
		After(func(tx *gorm.DB) error {
			if !body.Switch {
				err = tx.Where("role_key = ?", body.Key).
					Delete(model.RoleResourceAssoc{}).
					Error
				if err != nil {
					return err
				}
				var assoc []model.RoleResourceAssoc
				for _, resourceKey := range body.Resource {
					assoc = append(assoc, model.RoleResourceAssoc{
						RoleKey:     body.Key,
						ResourceKey: resourceKey,
					})
				}
				err = tx.Create(&assoc).Error
				if err != nil {
					return err
				}
			}
			clearcache(c.Cache)
			return nil
		}).
		Exec(data)
}

type deleteBody struct {
	operates.DeleteBody
}

func (c *Controller) Delete(ctx *gin.Context) interface{} {
	var body deleteBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return res.Error(err)
	}
	return c.Curd.
		Delete(model.RoleBasic{}, body.DeleteBody).
		After(func(tx *gorm.DB) error {
			clearcache(c.Cache)
			return nil
		}).
		Exec()
}

type validedkeyBody struct {
	Key string `binding:"required"`
}

func (c *Controller) Validedkey(ctx *gin.Context) interface{} {
	var body validedkeyBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return res.Error(err)
	}
	var count int64
	c.Db.Model(&model.RoleBasic{}).
		Where("keyid = ?", body.Key).
		Count(&count)
	return res.Data(count != 0)
}

func clearcache(cache *cache.Cache) {
	cache.Role.Clear()
	cache.Admin.Clear()
}
