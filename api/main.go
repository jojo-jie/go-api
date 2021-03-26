package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-api/cache"
	"go-api/conf"
	"go-api/ent"
	"go-api/middleware"
	"go-api/model"
	"go-api/serializer"
	"golang.org/x/sync/singleflight"
	"gopkg.in/go-playground/validator.v8"
	"strconv"
	"strings"
	"time"
)

// @Summary 接口调试
// @Accept json
// @Tags Ping
// @Produce  json
// @Router /api/v1/ping [get]
// @Success 200 {object} serializer.Response
func Ping(c *gin.Context) {
	c.JSON(200, serializer.Response{
		Code: 0,
		Msg:  "Pong",
	})
}

// CurrentUser 获取当前用户
func CurrentUser(c *gin.Context) (*ent.User, error) {
	if claims, _ := c.Get("claims"); claims != nil {
		if u, ok := claims.(*middleware.CustomClaims); ok {
			var buf strings.Builder
			buf.WriteString("member:")
			buf.WriteString(strconv.Itoa(int(u.ID)))
			key := buf.String()
			var m *ent.User
			if isExist(key) {
				mj, _ := cache.RedisClient.Get(key).Result()
				var d ent.User
				json.Unmarshal([]byte(mj), &d)
				m = &d
			} else {
				sg:=new(singleflight.Group)
				ch:=make(chan *ent.User)
				for i:=0;i<1000;i++ {
					go func() {
						ch <- getUser(key, c, int(u.ID), sg)
					}()
				}
				m=<-ch
			}
			return m, nil
		}
	}
	return nil, errors.New("无法获取用户信息")
}

func isExist(key string) bool {
	if n,err:=cache.RedisClient.Exists(key).Result();err==nil && n == 0 {
		return false
	}
	return true
}

func getUser(key string, c context.Context, id int, sg *singleflight.Group) *ent.User {
	v,_,_:=sg.Do("getme", func() (interface{}, error) {
		m, _ := model.Client.User.Get(c, id)
		mJson, _ := json.Marshal(m)
		cache.RedisClient.Set(key, string(mJson), 3600*time.Second)
		return m,nil
	})
	return v.(*ent.User)
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
