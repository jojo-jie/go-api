package api

import (
	"go-api/cache"
	"go-api/middleware"
	"go-api/serializer"
	"go-api/service"
	"go-api/util"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserRegister 用户注册接口
func UserRegister(c *gin.Context) {
	var registerService service.UserRegisterService
	if err := c.ShouldBind(&registerService); err == nil {
		res := registerService.Register()
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserLogin 用户登录接口
// @Summary 用户登录接口
// @Tags User
// @Param username formData string true "username"
// @Param password formData string true "password"
// @Router /api/v1/user/login [post]
// @Success 200 {object} serializer.BuildUser
func UserLogin(c *gin.Context) {
	var loginService service.UserLoginService
	defer func() {
		if err := recover(); err != nil {
			c.JSON(400, serializer.Err(40014, "发生错误", nil))
		}
	}()
	if err := c.ShouldBind(&loginService); err == nil {
		res := loginService.Login(c)
		c.JSON(200, res)
	} else {
		c.JSON(200, ErrorResponse(err))
	}
}

// UserMe 用户详情
func UserMe(c *gin.Context) {
	user, err := CurrentUser(c)
	if err != nil {
		c.JSON(200, serializer.Err(40004, "无法获取用户信息", err))
	}
	res := serializer.BuildUserResponse(user)
	c.JSON(200, res)
}

// UserLogout 用户登出
func UserLogout(c *gin.Context) {
	if claims, _ := c.Get("claims"); claims != nil {
		if u, ok := claims.(*middleware.CustomClaims); ok {
			key := strconv.Itoa(int(u.ID))
			cache.RedisClient.Del("user:" + key).Err()
			c.JSON(200, serializer.Response{
				Code: 0,
				Msg:  "你丫的GG了",
			})
		}
	} else {
		c.JSON(200, serializer.Err(40010, "发生错误", nil))
	}

	//session 销毁
	/*s := sessions.Default(c)
	s.Clear()
	s.Save()
	c.JSON(200, serializer.Response{
		Code: 0,
		Msg:  "你丫的GG了",
	})*/
}

func UserTokenRefresh(c *gin.Context) {
	token, err := TokenRefresh(c)
	if err != nil {
		c.JSON(200, serializer.Err(40010, "发生错误", err))
	} else {
		if claims, _ := c.Get("claims"); claims != nil {
			if u, ok := claims.(*middleware.CustomClaims); ok {
				tokenMD5 := util.StringToMD5(token)
				key := strconv.Itoa(int(u.ID))
				ttl, _ := strconv.Atoi(os.Getenv("TOKEN_TTL"))
				cache.RedisClient.Set("user:"+key, tokenMD5, time.Duration(ttl)*time.Second).Err()
			}
		}

		c.JSON(200, serializer.Response{
			Code: 0,
			Msg:  "居然刷上了",
			Data: token,
		})
	}
}

//route 或 method 不存在 统一错误信息
func HandleNotFound(c *gin.Context) {
	c.JSON(404, serializer.Err(400004, "请求不存在", nil))
}
