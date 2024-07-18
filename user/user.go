package user

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lanthora/cacao/logger"
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

func Register(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&request); err != nil {
		logger.Info("register bind failed: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"status": 1,
			"msg":    "bind json failed",
		})
		return
	}
	if len(request.Username) < 3 || !isAlphanumeric(request.Username) {
		c.JSON(http.StatusOK, gin.H{
			"status": 2,
			"msg":    "username format invalid",
		})
		return
	}
	if len(request.Password) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"status": 3,
			"msg":    "password format invalid",
		})
		return
	}

	db := storage.Get()
	if func() bool {
		count := int64(0)
		db.Model(&User{}).Where(&User{IP: c.ClientIP(), Role: "normal"}).Where("created_at > ?", time.Now().Add(-24*time.Hour)).Count(&count)
		return count > 0
	}() {
		c.JSON(http.StatusOK, gin.H{
			"status": 4,
			"msg":    fmt.Sprintf("register too frequently: %v", c.ClientIP()),
		})
		return
	}

	if func() bool {
		count := int64(0)
		db.Model(&User{}).Where(&User{Name: request.Username}).Count(&count)
		return count > 0
	}() {
		c.JSON(http.StatusOK, gin.H{
			"status": 5,
			"msg":    "username already taken",
		})
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
		c.JSON(http.StatusOK, gin.H{
			"status": 255,
			"msg":    result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": 0,
		"msg":    "success",
		"data": gin.H{
			"id":    user.ID,
			"token": user.Token,
		},
	})
}
