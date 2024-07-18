package user

import (
	"crypto/sha256"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/status"
	"github.com/lanthora/cacao/storage"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name     string `gorm:"uniqueIndex"`
	Password string
	Token    string
	Role     string
	IP       string
}

func init() {
	db := storage.Get()
	err := db.AutoMigrate(User{})
	if err != nil {
		logger.Fatal("auto migrate users failed: %v", err)
	}
}

func isAlphanumeric(s string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9]+$", s)
	return match
}

func sha256sum(data []byte) string {
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash[:])
}

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
		user := &User{}
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

func Register(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}
	if len(request.Username) < 3 || !isAlphanumeric(request.Username) {
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
		db.Model(&User{}).Where(&User{IP: c.ClientIP(), Role: "normal"}).Where("created_at > ?", time.Now().Add(-24*time.Hour)).Count(&count)
		return count > 0
	}() {
		status.UpdateCode(c, status.RegisterTooFrequently)
		return
	}

	if func() bool {
		count := int64(0)
		db.Model(&User{}).Where(&User{Name: request.Username}).Count(&count)
		return count > 0
	}() {
		status.UpdateCode(c, status.UsernameAlreadyTaken)
		return
	}

	role := func() string {
		count := int64(0)
		db.Model(&User{}).Count(&count)
		if count == 0 {
			return "admin"
		}
		return "normal"
	}()

	user := User{
		Name:     request.Username,
		Password: sha256sum([]byte(request.Password)),
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
}

func Login(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	user := User{
		Name:     request.Username,
		Password: sha256sum([]byte(request.Password)),
	}

	db := storage.Get()
	if result := db.Where(user).Take(&user); result.Error != nil {
		status.UpdateCode(c, status.UsernameOrPasswordIncorrect)
		return
	}

	user.Token = uuid.NewString()
	db.Save(user)

	c.SetCookie("id", strconv.FormatUint(uint64(user.ID), 10), 86400, "/", "", false, true)
	c.SetCookie("token", user.Token, 86400, "/", "", false, true)

	status.UpdateSuccess(c, gin.H{
		"name": user.Name,
		"role": user.Role,
	})
}

func Logout(c *gin.Context) {
	user := c.MustGet("user").(*User)
	user.Token = uuid.NewString()

	db := storage.Get()
	db.Save(user)

	c.SetCookie("id", "", -1, "/", "", false, true)
	c.SetCookie("token", "", -1, "/", "", false, true)

	status.UpdateSuccess(c, nil)
}
