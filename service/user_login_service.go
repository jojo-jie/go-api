package service

import (
	"github.com/dgrijalva/jwt-go"
	"go-api/cache"
	"go-api/middleware"
	"go-api/model"
	"go-api/serializer"
	"go-api/util"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UserLoginService 管理用户登录的服务
type UserLoginService struct {
	Username string `form:"username" json:"username" binding:"required,min=5,max=30"`
	Password string `form:"password" json:"password" binding:"required,min=6,max=40"`
}

// setToken 设置token
func (service *UserLoginService) setToken(user model.User, ttl int64) (string, int64, error) {
	j := middleware.NewJWT()
	notBefore := time.Now().Unix() - 1000
	expiresAt := time.Now().Unix() + ttl
	claims := middleware.CustomClaims{
		ID:    user.ID,
		Name:  user.Nickname,
		Phone: "",
		StandardClaims: jwt.StandardClaims{
			NotBefore: notBefore,
			ExpiresAt: expiresAt,
		},
	}
	token, err := j.CreateToken(claims)
	return token, expiresAt, err
}

// Login 用户登录函数
func (service *UserLoginService) Login(c *gin.Context) serializer.Response {
	var user model.User

	if err := model.DB.Where("username = ?", service.Username).First(&user).Error; err != nil {
		return serializer.ParamErr("账号或密码错误", nil)
	}

	if user.CheckPassword(service.Password) == false {
		return serializer.ParamErr("账号或密码错误", nil)
	}

	// 设置token
	ttl, err := strconv.Atoi(os.Getenv("TOKEN_TTL"))
	if err != nil {
		return serializer.Err(serializer.CodeTokenError, "token 获取失败", err)
	}
	token, expiresAt, err := service.setToken(user, int64(ttl))
	if err != nil {
		return serializer.Err(serializer.CodeTokenError, "token 获取失败", err)
	}

	tokenMD5 := util.StringToMD5(token)
	key := strconv.Itoa(int(user.ID))
	if err := cache.RedisClient.Set("user:"+key, tokenMD5, time.Duration(ttl)*time.Second).Err(); err != nil {
		panic(err)
	}
	return serializer.BuildToken(user, token, expiresAt)
}
