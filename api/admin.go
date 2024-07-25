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

func AdminShowUsers(c *gin.Context) {
	candy.Flush()
	users := model.GetUsers()

	type userinfo struct {
		UserID   uint   `json:"userid"`
		Username string `json:"username"`
		Role     string `json:"role"`
		RegTime  string `json:"regtime"`
		NetNum   uint   `json:"netnum"`
		DevNum   uint   `json:"devnum"`
		RxSum    uint64 `json:"rxsum"`
		TxSum    uint64 `json:"txsum"`
	}

	response := make([]userinfo, 0)
	for _, u := range users {
		response = append(response, userinfo{
			UserID:   u.ID,
			Username: u.Name,
			Role:     u.Role,
			RegTime:  u.CreatedAt.Format(time.DateTime),
			NetNum:   uint(len(model.GetNetsByUserID(u.ID))),
			DevNum:   uint(len(model.GetDevicesByUserID(u.ID))),
			RxSum:    model.GetRxSumByUserID(u.ID),
			TxSum:    model.GetTxSumByUserID(u.ID),
		})
	}

	status.UpdateSuccess(c, gin.H{
		"users": response,
	})
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
	if isInvalidUsername(request.Username) {
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
		Password: hashUserPassword(request.Username, request.Password),
		Token:    uuid.NewString(),
		Role:     "normal",
	}

	if result := db.Create(&user); result.Error != nil {
		status.UpdateUnexpected(c, result.Error.Error())
		return
	}

	status.UpdateSuccess(c, gin.H{
		"name": user.Name,
		"role": user.Role,
	})

	netModel := &model.Net{
		UserID:    user.ID,
		Name:      "@",
		Password:  request.Password,
		DHCP:      "192.168.202.0/24",
		Broadcast: true,
	}
	netModel.Create()
	candy.InsertNet(netModel)
}

func AdminDeleteUser(c *gin.Context) {
	var request struct {
		UserID uint `json:"userid"`
	}

	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	user := c.MustGet("user").(*model.User)
	if request.UserID == user.ID {
		status.UpdateCode(c, status.CannotDeleteAdmin)
		return
	}

	nets := model.GetNetsByUserID(request.UserID)
	for _, n := range nets {
		candy.DeleteNet(n.ID)
		model.DeleteDevicesByNetID(n.ID)
		model.DeleteNetByNetID(n.ID)
	}
	model.DeleteUserByUserID(request.UserID)
	status.UpdateSuccess(c, nil)
}

func AdminUpdateUserPassword(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}
	if isInvalidUsername(request.Username) {
		status.UpdateCode(c, status.InvalidUsername)
		return
	}
	if len(request.Password) == 0 {
		status.UpdateCode(c, status.InvalidPassword)
		return
	}

	user := model.User{Name: request.Username}
	db := storage.Get()
	if result := db.Model(&model.User{}).Where(user).Take(&user); result.Error != nil {
		status.UpdateCode(c, status.UnexpectedError)
		return
	}
	user.Password = hashUserPassword(user.Name, request.Password)
	user.Save()
	status.UpdateSuccess(c, nil)
}

func AdminGetOpenRegisterConfig(c *gin.Context) {
	openreg := model.GetConfig("openreg", "true") == "true"
	status.UpdateSuccess(c, gin.H{
		"openreg": openreg,
	})
}

func AdminSetOpenRegisterConfig(c *gin.Context) {
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

func AdminGetRegisterIntervalConfig(c *gin.Context) {
	intervalStr := model.GetConfig("reginterval", "1440")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 1440
	}
	status.UpdateSuccess(c, gin.H{
		"reginterval": interval,
	})
}

func AdminSetRegisterIntervalConfig(c *gin.Context) {
	var request struct {
		RegInterval uint `json:"reginterval"`
	}
	if err := c.BindJSON(&request); err != nil {
		status.UpdateCode(c, status.InvalidRequest)
		return
	}

	model.SetConfig("reginterval", strconv.FormatUint(uint64(request.RegInterval), 10))
	status.UpdateSuccess(c, nil)
}
