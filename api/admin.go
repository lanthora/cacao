package api

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/status"
	"github.com/lanthora/cacao/storage"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.String(), "/api/admin/") {
			if user := c.MustGet("user").(*model.User); user.Role != "admin" {
				status.UpdateCode(c, status.AdminAccessRequired)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

func AdminAddUser(c *gin.Context) {
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
		db.Model(&model.User{}).Where(&model.User{Name: request.Username}).Count(&count)
		return count > 0
	}() {
		status.UpdateCode(c, status.UsernameAlreadyTaken)
		return
	}

	user := model.User{
		Name:     request.Username,
		Password: candy.Sha256sum([]byte(request.Password)),
		Token:    uuid.NewString(),
		Role:     "normal",
		IP:       c.ClientIP(),
	}

	if result := db.Create(&user); result.Error != nil {
		status.UpdateUnexpected(c, result.Error.Error())
		return
	}

	status.UpdateSuccess(c, gin.H{
		"name": user.Name,
		"role": user.Role,
	})

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

func AdminOpenRegister(c *gin.Context) {
	var request struct {
		OpenReg bool `json:"openreg"`
	}
	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}
	if request.OpenReg {
		model.SetConfig("openreg", "true")
	} else {
		model.SetConfig("openreg", "false")
	}
	status.UpdateSuccess(c, nil)
}
