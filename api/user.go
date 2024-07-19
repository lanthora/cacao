package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/status"
	"github.com/lanthora/cacao/storage"
)

func LoginMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.String()
		if !strings.HasPrefix(path, "/api/") {
			c.Next()
			return
		}
		if path == "/api/user/register" || path == "/api/user/login" {
			c.Next()
			return
		}
		idstr, errid := c.Cookie("id")
		token, errtoken := c.Cookie("token")
		if errid != nil || errtoken != nil || len(idstr) == 0 || len(token) == 0 {
			status.UpdateCode(c, status.NotLoggedIn)
			c.Abort()
			return
		}
		id, err := strconv.ParseUint(idstr, 10, 64)
		if err != nil {
			status.UpdateCode(c, status.NotLoggedIn)
			c.Abort()
			return
		}
		user := &model.User{}
		user.ID = uint(id)

		db := storage.Get()
		result := db.Where(user).Take(user)
		if result.Error != nil || user.Token != token {
			status.UpdateCode(c, status.NotLoggedIn)
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func UserRegister(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}
	if len(request.Username) < 3 || !candy.IsAlphanumeric(request.Username) {
		status.UpdateCode(c, status.InvalidUsername)
		return
	}
	if len(request.Password) == 0 {
		status.UpdateCode(c, status.InvalidPassword)
		return
	}

	db := storage.Get()
	if func() bool {
		count := int64(0)
		db.Model(&model.User{}).Where(&model.User{IP: c.ClientIP(), Role: "normal"}).Where("created_at > ?", time.Now().Add(-24*time.Hour)).Count(&count)
		return count > 0
	}() {
		status.UpdateCode(c, status.RegisterTooFrequently)
		return
	}

	if func() bool {
		count := int64(0)
		db.Model(&model.User{}).Where(&model.User{Name: request.Username}).Count(&count)
		return count > 0
	}() {
		status.UpdateCode(c, status.UsernameAlreadyTaken)
		return
	}

	role := func() string {
		count := int64(0)
		db.Model(&model.User{}).Count(&count)
		if count == 0 {
			return "admin"
		}
		return "normal"
	}()

	user := model.User{
		Name:     request.Username,
		Password: candy.Sha256sum([]byte(request.Password)),
		Token:    uuid.NewString(),
		Role:     role,
		IP:       c.ClientIP(),
	}

	if result := db.Create(&user); result.Error != nil {
		status.UpdateUnexpected(c, result.Error.Error())
		return
	}

	c.SetCookie("id", strconv.FormatUint(uint64(user.ID), 10), 86400, "/", "", false, true)
	c.SetCookie("token", user.Token, 86400, "/", "", false, true)

	status.UpdateSuccess(c, gin.H{
		"name": user.Name,
		"role": user.Role,
	})

	if role == "normal" {
		modelNet := &model.Net{
			UserID:    user.ID,
			Name:      "@",
			Password:  request.Password,
			DHCP:      "192.168.202.0/24",
			Broadcast: true,
		}
		modelNet.Create()
		candy.InsertNet(modelNet)
	}
}

func UserLogin(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	user := model.User{
		Name:     request.Username,
		Password: candy.Sha256sum([]byte(request.Password)),
	}

	db := storage.Get()

	if result := db.Where(user).Take(&user); result.Error != nil {
		status.UpdateCode(c, status.UsernameOrPasswordIncorrect)
		return
	}

	user.Token = uuid.NewString()
	user.Save()

	c.SetCookie("id", strconv.FormatUint(uint64(user.ID), 10), 86400, "/", "", false, true)
	c.SetCookie("token", user.Token, 86400, "/", "", false, true)

	status.UpdateSuccess(c, gin.H{
		"name": user.Name,
		"role": user.Role,
	})
}

func UserLogout(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	user.Token = uuid.NewString()
	user.Save()

	c.SetCookie("id", "", -1, "/", "", false, true)
	c.SetCookie("token", "", -1, "/", "", false, true)

	status.UpdateSuccess(c, nil)
}
