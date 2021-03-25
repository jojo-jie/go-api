package service

import (
	"context"
	"go-api/cache"
	"go-api/ent"
	"go-api/ent/user"
	"go-api/middleware"
	"go-api/model"
	"go-api/serializer"
	"go-api/util"
	"os"
	"strconv"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"
)

// UserLoginService 管理用户登录的服务
type UserLoginService struct {
	Username string `form:"username" json:"username" binding:"required,min=5,max=30"`
	Password string `form:"password" json:"password" binding:"required,min=6,max=40"`
}

// setToken 设置token
func (service *UserLoginService) setToken(user *ent.User, ttl int64) (string, int64, error) {
	j := middleware.NewJWT()
	notBefore := time.Now().Unix() - 1000
	expiresAt := time.Now().Unix() + ttl
	claims := middleware.CustomClaims{
		ID:    uint(user.ID),
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
	member, err := model.Client.User.Query().Where(user.Username(service.Username)).First(context.TODO())
	if err != nil {
		return serializer.ParamErr("账号或d密码错误", nil)
	}

	if checkPassword(member, service) {
		return serializer.ParamErr("账号或密码错误", nil)
	}

	// 设置token
	ttl, err := strconv.Atoi(os.Getenv("TOKEN_TTL"))
	if err != nil {
		return serializer.Err(serializer.CodeTokenError, "token 获取失败", err)
	}
	token, expiresAt, err := service.setToken(member, int64(ttl))
	if err != nil {
		return serializer.Err(serializer.CodeTokenError, "token 获取失败", err)
	}

	tokenMD5 := util.StringToMD5(token)
	key := strconv.Itoa(member.ID)
	if err := cache.RedisClient.Set("user:"+key, tokenMD5, time.Duration(ttl)*time.Second).Err(); err != nil {
		panic(err)
	}
	return serializer.BuildToken(member, token, expiresAt)
}

func checkPassword(u *ent.User, service *UserLoginService) bool {
	return bcrypt.CompareHashAndPassword([]byte(u.PasswordDigest), []byte(service.Password)) == nil
}
