package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/boj/redistore"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/uptrace/bun"
	"net/http"
	"net/url"
	"time"
)

const ErrUserNotFound = "sql: no rows in result set"

type ThirdServer struct {
	DB      *bun.DB
	Authn   *webauthn.WebAuthn
	Session *redistore.RediStore
}

var ts ThirdServer

func NewThirdServer(tt ThirdServer) {
	ts = tt
}

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`
	ID            int                   `bun:"id,pk,autoincrement"`
	AuthnID       string                `bun:"authn_id,notnull"`
	Username      string                `bun:"username,notnull"`
	DisplayName   string                `bun:"display_name,notnull"`
	Credentials   []webauthn.Credential `bun:"credentials"`
	CreatedTime   time.Time             `bun:",nullzero,notnull,default:current_timestamp"`
	UpdatedTime   time.Time             `bun:",nullzero,notnull,default:current_timestamp"`
}

func (u *User) WebAuthnID() []byte {
	return []byte(u.AuthnID)
}

func (u *User) WebAuthnName() string {
	//TODO implement me
	return u.Username
}

func (u *User) WebAuthnDisplayName() string {
	//TODO implement me
	return u.DisplayName
}

func (u *User) WebAuthnCredentials() []webauthn.Credential {
	return u.Credentials
}

func (u *User) AddCredential(cred webauthn.Credential) {
	u.Credentials = append(u.Credentials, cred)
}

func (u *User) CredentialExcludeList() []protocol.CredentialDescriptor {
	credentials := u.WebAuthnCredentials()
	var credentialExcludeList []protocol.CredentialDescriptor
	for _, cred := range credentials {
		descriptor := protocol.CredentialDescriptor{
			Type:         protocol.PublicKeyCredentialType,
			CredentialID: cred.ID,
		}
		credentialExcludeList = append(credentialExcludeList, descriptor)
	}

	return credentialExcludeList
}

func HandleBeginRegistration(writer http.ResponseWriter, request *http.Request) {
	u := &User{}
	err := json.NewDecoder(request.Body).Decode(u)
	if err != nil {
		Error(writer, err.Error())
		return
	}

	u.AuthnID = getAuthnID(u.Username)
	_, err = getUser(request.Context(), u.AuthnID)
	if err != nil && ErrUserNotFound != err.Error() {
		Error(writer, err.Error())
		return
	}
	if err == nil {
		Error(writer, "user already exists")
		return
	}

	u.DisplayName = u.Username

	registerOptions := func(credCreationOpts *protocol.PublicKeyCredentialCreationOptions) {
		credCreationOpts.CredentialExcludeList = u.CredentialExcludeList()
	}

	options, sessionData, err := ts.Authn.BeginRegistration(u, registerOptions)
	if err != nil {
		Error(writer, err.Error())
		return
	}
	sessionValue, err := sessionDataToStr(sessionData)
	if err != nil {
		Error(writer, err.Error())
		return
	}
	_, err = ts.DB.NewInsert().Model(u).Exec(request.Context())
	if err != nil {
		Error(writer, err.Error())
		return
	}

	http.SetCookie(writer, &http.Cookie{
		Name:     "session",
		Value:    sessionValue,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
	})
	Success(writer, "ok", options)
}

func HandleFinishRegistration(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("session")
	if err != nil {
		return
	}
	cookieObj, _ := url.QueryUnescape(cookie.Value)
	sessionData := &webauthn.SessionData{}
	err = json.Unmarshal([]byte(cookieObj), sessionData)
	if err != nil {
		Error(writer, err.Error())
		return
	}

	user, err := getUser(request.Context(), string(sessionData.UserID))
	if err != nil {
		Error(writer, err.Error())
		return
	}

	credential, err := ts.Authn.FinishRegistration(user, *sessionData, request)
	if err != nil {
		Error(writer, err.Error())
		return
	}
	user.AddCredential(*credential)
	_, err = ts.DB.NewUpdate().Model(user).WherePK().Exec(request.Context())
	if err != nil {
		Error(writer, err.Error())
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})
	Success(writer, "ok", nil)
	return
}

func HandleBeginLogin(writer http.ResponseWriter, request *http.Request) {
	u := &User{}
	if err := json.NewDecoder(request.Body).Decode(u); err != nil {
		Error(writer, "username resolution failed. Procedure")
		return
	}

	user, err := getUser(request.Context(), getAuthnID(u.Username))
	if err != nil {
		Error(writer, err.Error())
		return
	}

	options, sessionData, err := ts.Authn.BeginLogin(user)
	if err != nil {
		Error(writer, err.Error())
		return
	}
	sessionValue, err := sessionDataToStr(sessionData)
	if err != nil {
		Error(writer, err.Error())
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     "session",
		Value:    sessionValue,
		Path:     "/",
		MaxAge:   300,
		HttpOnly: true,
	})
	Success(writer, "ok", options)
	return
}

func HandleFinishLogin(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("session")
	if err != nil {
		return
	}

	cookieObj, _ := url.QueryUnescape(cookie.Value)
	sessionData := &webauthn.SessionData{}
	err = json.Unmarshal([]byte(cookieObj), sessionData)
	user, err := getUser(request.Context(), string(sessionData.UserID))
	if err != nil {
		Error(writer, err.Error())
		return
	}

	_, err = ts.Authn.FinishLogin(user, *sessionData, request)
	if err != nil {
		Error(writer, err.Error())
		return
	}
	http.SetCookie(writer, &http.Cookie{
		Name:     "session",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
	})

	ts.Session.SetMaxAge(86400)
	session, err := ts.Session.Get(request, "skey")
	if err != nil {
		Error(writer, err.Error())
		return
	}
	session.Values["login_user_id"] = user.Username
	if err = ts.Session.Save(request, writer, session); err != nil {
		Error(writer, err.Error())
		return
	}
	Success(writer, "ok", nil)
	return
}

func getAuthnID(username string) string {
	h := md5.New()
	h.Write([]byte(username))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func sessionDataToStr(sessionData *webauthn.SessionData) (string, error) {
	marshal, err := json.Marshal(sessionData)
	if err != nil {
		return "", err
	}
	return url.QueryEscape(string(marshal)), nil
}

func getUser(ctx context.Context, authnID string) (*User, error) {
	user := &User{}
	err := ts.DB.NewSelect().Model(user).Where("authn_id = ?", authnID).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func HandleUsers(writer http.ResponseWriter, request *http.Request) {
	session, err := ts.Session.Get(request, "skey")
	if err != nil {
		Error(writer, err.Error())
		return
	}

	Success(writer, "ok", map[string]interface{}{
		"name": session.Values["login_user_id"],
	})
	return
}
