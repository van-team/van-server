package role

import (
	"github.com/gin-gonic/gin"
	curd "github.com/kainonly/gin-curd"
	"gorm.io/gorm"
	"lab-api/application/common"
	"lab-api/application/common/datatype"
	"lab-api/application/model"
)

type Controller struct {
	*common.Dependency
}

type originListsBody struct {
	curd.OriginLists
}

func (c *Controller) OriginLists(ctx *gin.Context) interface{} {
	var body originListsBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	return c.Curd.Operates(
		curd.Plan(model.Role{}, body),
	).Originlists()
}

type listsBody struct {
	curd.Lists
}

func (c *Controller) Lists(ctx *gin.Context) interface{} {
	var body listsBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	return c.Curd.Operates(
		curd.Plan(model.Role{}, body),
	).Lists()
}

type getBody struct {
	curd.Get
}

func (c *Controller) Get(ctx *gin.Context) interface{} {
	var body getBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	return c.Curd.Operates(
		curd.Plan(model.Role{}, body),
	).Get()
}

type addBody struct {
	Key      string              `binding:"required"`
	Name     datatype.JSONObject `binding:"required"`
	Resource []string            `binding:"required"`
	Note     string
	Status   bool
}

func (c *Controller) Add(ctx *gin.Context) interface{} {
	var body addBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	data := model.RoleBasic{
		Key:    body.Key,
		Name:   body.Name,
		Note:   body.Note,
		Status: body.Status,
	}
	return c.Curd.Operates(
		curd.After(func(tx *gorm.DB) error {
			var assoc []model.RoleResourceRel
			for _, resourceKey := range body.Resource {
				assoc = append(assoc, model.RoleResourceRel{
					RoleKey:     body.Key,
					ResourceKey: resourceKey,
				})
			}
			if err = tx.Create(&assoc).Error; err != nil {
				return err
			}
			c.clearcache()
			return nil
		}),
	).Add(&data)
}

type editBody struct {
	curd.Edit
	Key      string              `binding:"required_if=switch false"`
	Name     datatype.JSONObject `binding:"required_if=switch false"`
	Resource []string            `binding:"required_if=switch false"`
	Note     string
	Status   bool
}

func (c *Controller) Edit(ctx *gin.Context) interface{} {
	var body editBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	data := model.RoleBasic{
		Key:    body.Key,
		Name:   body.Name,
		Note:   body.Note,
		Status: body.Status,
	}
	return c.Curd.Operates(
		curd.Plan(model.Resource{}, body),
		curd.After(func(tx *gorm.DB) error {
			if !body.Switch {
				if err = tx.Where("role_key = ?", body.Key).
					Delete(model.RoleResourceRel{}).
					Error; err != nil {
					return err
				}
				var assoc []model.RoleResourceRel
				for _, resourceKey := range body.Resource {
					assoc = append(assoc, model.RoleResourceRel{
						RoleKey:     body.Key,
						ResourceKey: resourceKey,
					})
				}
				if err = tx.Create(&assoc).Error; err != nil {
					return err
				}
			}
			c.clearcache()
			return nil
		}),
	).Edit(data)
}

type deleteBody struct {
	curd.Delete
}

func (c *Controller) Delete(ctx *gin.Context) interface{} {
	var body deleteBody
	var err error
	if err = ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	return c.Curd.Operates(
		curd.Plan(model.RoleBasic{}, body),
		curd.After(func(tx *gorm.DB) error {
			c.clearcache()
			return nil
		}),
	).Delete()
}

type validedkeyBody struct {
	Key string `binding:"required"`
}

func (c *Controller) Validedkey(ctx *gin.Context) interface{} {
	var body validedkeyBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		return err
	}
	var count int64
	c.Db.Model(&model.RoleBasic{}).
		Where("keyid = ?", body.Key).
		Count(&count)

	return gin.H{
		"exists": count != 0,
	}
}

func (c *Controller) clearcache() {
	c.Cache.Role.Clear()
	c.Cache.Admin.Clear()
}
