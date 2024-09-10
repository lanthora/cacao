package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
	"github.com/lanthora/cacao/storage"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.String()
		if strings.HasPrefix(path, "/api/") {
			if user, ok := c.Get("user"); ok {
				user := user.(*model.User)
				if strings.HasPrefix(path, "/api/admin/") {
					if user.Role == "admin" {
						c.Next()
					} else {
						setErrorCode(c, PermissionDenied)
						c.Abort()
					}
				} else if path == "/api/user/info" || path == "/api/user/logout" {
					c.Next()
				} else if user.Role == "normal" {
					c.Next()
				} else {
					setErrorCode(c, PermissionDenied)
					c.Abort()
				}
			}
		}
	}
}

func AdminShowUsers(c *gin.Context) {
	users := model.GetUsers()

	type userinfo struct {
		UserID         uint   `json:"userid"`
		Username       string `json:"username"`
		Role           string `json:"role"`
		RegTime        string `json:"regtime"`
		LastActiveTime string `json:"lastActiveTime"`
		NetNum         uint   `json:"netnum"`
		DevNum         uint   `json:"devnum"`
		RxSum          uint64 `json:"rxsum"`
		TxSum          uint64 `json:"txsum"`
	}

	response := make([]userinfo, 0)
	for _, u := range users {
		response = append(response, userinfo{
			UserID:         u.ID,
			Username:       u.Name,
			Role:           u.Role,
			RegTime:        u.CreatedAt.Format(time.DateTime),
			LastActiveTime: model.GetLastActiveTimeByUserID(u.ID).Format(time.DateTime),
			NetNum:         uint(len(model.GetNetsByUserID(u.ID))),
			DevNum:         uint(len(model.GetDevicesByUserID(u.ID))),
			RxSum:          model.GetRxSumByUserID(u.ID),
			TxSum:          model.GetTxSumByUserID(u.ID),
		})
	}

	setResponseData(c, gin.H{
		"users": response,
	})
}

func AdminAddUser(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}
	if !candy.IsValidUsername(request.Username) {
		setErrorCode(c, InvalidUsername)
		return
	}
	if len(request.Password) == 0 {
		setErrorCode(c, InvalidPassword)
		return
	}

	db := storage.Get()
	if func() bool {
		count := int64(0)
		db.Model(&model.User{}).Where(&model.User{Name: request.Username}).Count(&count)
		return count > 0
	}() {
		setErrorCode(c, UsernameAlreadyTaken)
		return
	}

	user := model.User{
		Name:     request.Username,
		Password: hashUserPassword(request.Username, request.Password),
		Token:    uuid.NewString(),
		Role:     "normal",
	}

	if result := db.Create(&user); result.Error != nil {
		setUnexpectedMessage(c, result.Error.Error())
		return
	}

	setResponseData(c, gin.H{
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

	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}

	user := c.MustGet("user").(*model.User)
	if request.UserID == user.ID {
		setErrorCode(c, CannotDeleteAdmin)
		return
	}

	nets := model.GetNetsByUserID(request.UserID)
	for _, n := range nets {
		candy.DeleteNet(n.ID)
		model.DeleteDevicesByNetID(n.ID)
		model.DeleteNetByNetID(n.ID)
	}
	model.DeleteUserByUserID(request.UserID)
	setResponseData(c, nil)
}

func AdminUpdateUserPassword(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}
	if !candy.IsValidUsername(request.Username) {
		setErrorCode(c, InvalidUsername)
		return
	}
	if len(request.Password) == 0 {
		setErrorCode(c, InvalidPassword)
		return
	}

	user := model.User{Name: request.Username}
	db := storage.Get()
	if result := db.Model(&model.User{}).Where(user).Take(&user); result.Error != nil {
		setErrorCode(c, Unexpected)
		return
	}
	user.Password = hashUserPassword(user.Name, request.Password)
	user.Save()
	setResponseData(c, nil)
}

func AdminGetOpenRegisterConfig(c *gin.Context) {
	openreg := model.GetConfig("openreg", "true") == "true"
	setResponseData(c, gin.H{
		"openreg": openreg,
	})
}

func AdminSetOpenRegisterConfig(c *gin.Context) {
	var request struct {
		OpenReg bool `json:"openreg"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}
	if request.OpenReg {
		model.SetConfig("openreg", "true")
	} else {
		model.SetConfig("openreg", "false")
	}
	setResponseData(c, nil)
}

func AdminGetRegisterIntervalConfig(c *gin.Context) {
	intervalStr := model.GetConfig("reginterval", "1440")
	interval, err := strconv.Atoi(intervalStr)
	if err != nil {
		interval = 1440
	}
	setResponseData(c, gin.H{
		"reginterval": interval,
	})
}

func AdminSetRegisterIntervalConfig(c *gin.Context) {
	var request struct {
		RegInterval uint `json:"reginterval"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}

	model.SetConfig("reginterval", strconv.FormatUint(uint64(request.RegInterval), 10))
	setResponseData(c, nil)
}

func AdminGetAutoCleanUserConfig(c *gin.Context) {
	autoCleanUser := model.GetConfig("autoCleanUser", "false") == "true"
	setResponseData(c, gin.H{
		"autoCleanUser": autoCleanUser,
	})
}

func AdminSetAutoCleanUserConfig(c *gin.Context) {
	var request struct {
		AutoCleanInactiveUser bool `json:"autoCleanUser"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}
	if request.AutoCleanInactiveUser {
		model.SetConfig("autoCleanUser", "true")
	} else {
		model.SetConfig("autoCleanUser", "false")
	}
	setResponseData(c, nil)
}

func AdminGetInactiveUserThresholdConfig(c *gin.Context) {
	thresholdStr := model.GetConfig("inactiveUserThreshold", "7")
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil {
		threshold = 7
	}
	setResponseData(c, gin.H{
		"inactiveUserThreshold": threshold,
	})
}

func AdminSetInactiveUserThresholdConfig(c *gin.Context) {
	var request struct {
		InactiveUserThreshold uint `json:"inactiveUserThreshold"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}
	if request.InactiveUserThreshold <= 0 {
		setErrorCode(c, InvalidInactiveUserThreshold)
		return
	}

	model.SetConfig("inactiveUserThreshold", strconv.FormatUint(uint64(request.InactiveUserThreshold), 10))
	setResponseData(c, nil)
}

func AdminCleanInactiveUser(c *gin.Context) {
	candy.CleanInactiveUser()
	setResponseData(c, nil)
}
