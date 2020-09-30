package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-api/conf"
	"go-api/middleware"
	"go-api/model"
	"go-api/serializer"
	"gopkg.in/go-playground/validator.v8"
)

// @Summary 接口调试
// @Accept json
// @Tags Ping
// @Produce  json
// @Router /api/v1/ping [post]
// @Success 200 {object} serializer.Response
func Ping(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Code: 0,
		Msg:  "Pong",
	})
}

// CurrentUser 获取当前用户
func CurrentUser(c *gin.Context) (model.User, error) {
	if claims, _ := c.Get("claims"); claims != nil {
		if u, ok := claims.(*middleware.CustomClaims); ok {
			user, err := model.GetUser(u.ID)
			return user, err
		}
	}
	return model.User{}, errors.New("无法获取用户信息")
}

// ErrorResponse 返回错误消息
func ErrorResponse(err error) serializer.Response {
	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, e := range ve {
			field := conf.T(fmt.Sprintf("Field.%s", e.Field))
			tag := conf.T(fmt.Sprintf("Tag.Valid.%s", e.Tag))
			return serializer.ParamErr(
				fmt.Sprintf("%s%s", field, tag),
				err,
			)
		}
	}
	if _, ok := err.(*json.UnmarshalTypeError); ok {
		return serializer.ParamErr("JSON类型不匹配", err)
	}

	return serializer.ParamErr("参数错误", err)
}

//刷新token
func TokenRefresh(c *gin.Context) (string, error) {
	j := middleware.NewJWT()
	if token, _ := c.Get("token"); token != nil {
		if tokenStr, ok := token.(string); ok {
			return j.RefreshToken(tokenStr)
		}
	}
	return "", errors.New("刷新失败")
}
