package middleware

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"github.com/gin-gonic/gin"
	"go-api/cache"
	"go-api/serializer"
	"go-api/util"
	"os"
	"strconv"
	"strings"
	"time"
)

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("token")
		if token == "" {
			c.JSON(200, serializer.Err(serializer.CodeTokenError, "缺少token", nil))
			c.Abort()
			return
		}
		//fmt.Fprintln(gin.DefaultWriter, token)
		j := NewJWT()
		// parseToken 解析token包含的信息
		claims, err := j.ParseToken(token)
		if err != nil {
			if err == TokenExpired {
				c.JSON(200, serializer.Err(serializer.CodeTokenError, "已过期", err))
				c.Abort()
				return
			}
			c.JSON(200, serializer.Err(serializer.CodeTokenError, "", err))
			c.Abort()
			return
		}

		tokenMD5 := util.StringToMD5(token)
		key := strconv.Itoa(int(claims.ID))
		if strings.Compare(cache.RedisClient.Get("user:"+key).Val(), tokenMD5) != 0 {
			c.JSON(400, serializer.Err(serializer.CodeParamErr, "token失效, 重新获取", nil))
			c.Abort()
			return
		}

		// 继续交由下一个路由处理，并将解析出的信息传递下去
		c.Set("claims", claims)
		c.Set("token", token)
		c.Next()
	}
}

type JWT struct {
	SigningKey []byte
}

var (
	TokenExpired     error  = errors.New("token 过期了")
	TokenNotValidYet error  = errors.New("token 尚未激活")
	TokenMalformed   error  = errors.New("非法 token")
	TokenInvalid     error  = errors.New("无法处理此 token")
	SignKey          string = "newtrekWang"
)

// 载荷可以自定义信息
type CustomClaims struct {
	ID    uint   `json:"userId"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
	jwt.StandardClaims
}

// 创建jwt 实例
func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

//获取SignKey
func GetSignKey() string {
	return SignKey
}

//设置SignKey
func SetSignKey(key string) string {
	SignKey = key
	return SignKey
}

//创建token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

//解析token
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, TokenMalformed
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				return nil, TokenExpired
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				return nil, TokenNotValidYet
			} else {
				return nil, TokenInvalid
			}
		}
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, TokenInvalid
}

//更新token
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		var ttl int
		if ttl, err = strconv.Atoi(os.Getenv("TOKEN_TTL")); err != nil {
			return "", err
		}
		claims.StandardClaims.ExpiresAt = time.Now().Add(time.Duration(ttl) * time.Second).Unix()
		return j.CreateToken(*claims)
	}
	return "", TokenInvalid
}
