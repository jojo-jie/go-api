package serializer

import (
	"go-api/ent"
)

// User 用户序列化器
type User struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Status    string `json:"status"`
	Avatar    string `json:"avatar"`
	CreatedAt int64  `json:"created_at"`
}

//User 用户序列化器
type UserToken struct {
	ID        uint   `json:"id"`
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	Status    string `json:"status"`
	Avatar    string `json:"avatar"`
	CreatedAt int64  `json:"created_at"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type UserList struct {
	list  []UserToken
	total int `json:"total"`
}

// BuildUser 序列化用户
func BuildUser(user *ent.User) User {
	return User{
		ID:        uint(user.ID),
		Username:  user.Username,
		Nickname:  user.Nickname,
		Status:    user.Status,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt.Unix(),
	}
}

// BuildUserToken 序列化用户带token信息
func BuildUserToken(user *ent.User, token string, expiresAt int64) UserToken {
	return UserToken{
		ID:        uint(user.ID),
		Username:  user.Username,
		Nickname:  user.Nickname,
		Status:    user.Status,
		Avatar:    user.Avatar,
		CreatedAt: user.CreatedAt.Unix(),
		Token:     token,
		ExpiresAt: expiresAt,
	}
}

// BuildUser
func BuildToken(user *ent.User, token string, expiresAt int64) Response {
	return Response{
		Data: BuildUserToken(user, token, expiresAt),
	}
}

// BuildUserResponse 序列化用户响应
func BuildUserResponse(user *ent.User) Response {
	return Response{
		Data: BuildUser(user),
	}
}
