package v1

import (
	"blog/global"
	"blog/internal/model"
	"blog/pkg/app"
	"blog/pkg/errcode"
	"github.com/gin-gonic/gin"
)

type Tag struct {
}

func NewTag() Tag {
	return Tag{}
}

func (t Tag) Get(c *gin.Context) {

}

func (t Tag) List(c *gin.Context) {
	response := app.NewResponse(c)
	valid,errs := app.BindAndValid(c, &model.TagListRequest{})
	if !valid {
		global.Logger.Errorf("app.BindAndValid errs; %v", errs)
		errRsp:=errcode.InvalidParams.WithDetails(errs.Errors()...)
		response.ToErrorResponse(errRsp)
		return
	}
	response.ToResponse(nil)
	return
}

func (t Tag) Create(c *gin.Context) {

}

func (t Tag) Update(c *gin.Context) {

}

func (t Tag) Delete(c *gin.Context) {

}
