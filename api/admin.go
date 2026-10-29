package api

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
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
	users, err := model.GetUsers(c.Request.Context())
	if databaseError(c, err, Unexpected) {
		return
	}

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
		stats, err := model.GetUserStats(u.ID, c.Request.Context())
		if databaseError(c, err, Unexpected) {
			return
		}
		response = append(response, userinfo{
			UserID:         u.ID,
			Username:       u.Name,
			Role:           u.Role,
			RegTime:        u.CreatedAt.Format(time.DateTime),
			LastActiveTime: u.UpdatedAt.Format(time.DateTime),
			NetNum:         uint(stats.NetNum),
			DevNum:         uint(stats.DevNum),
			RxSum:          stats.RX,
			TxSum:          stats.TX,
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

	candy.ConfigMu.Lock()
	defer candy.ConfigMu.Unlock()
	user, network, err := model.RegisterUser(model.User{Name: request.Username, Password: hashUserPassword(request.Username, request.Password), Token: uuid.NewString()}, model.Net{Name: "@", Password: randomString(8), DHCP: "192.168.202.0/24", Broadcast: true}, true, c.Request.Context())
	if databaseError(c, err, Unexpected) {
		return
	}
	candy.InsertNet(network)
	setResponseData(c, gin.H{"name": user.Name, "role": user.Role})
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

	candy.ConfigMu.Lock()
	defer candy.ConfigMu.Unlock()
	ids, err := model.DeleteUserTree(request.UserID, nil, c.Request.Context())
	if databaseError(c, err, UserNotExists) {
		return
	}
	for _, id := range ids {
		candy.DeleteNet(id)
	}
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

	user, err := model.FindUser(request.Username, c.Request.Context())
	if databaseError(c, err, UserNotExists) {
		return
	}
	if databaseError(c, user.UpdateCredentials(hashUserPassword(user.Name, request.Password), uuid.NewString(), user.IP, c.Request.Context()), Unexpected) {
		return
	}
	setResponseData(c, nil)
}

func AdminGetOpenRegisterConfig(c *gin.Context) {
	value, ok := config(c, "openreg", "true")
	if !ok {
		return
	}
	openreg := value == "true"
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
	saveConfig(c, "openreg", strconv.FormatBool(request.OpenReg))
}

func AdminGetRegisterIntervalConfig(c *gin.Context) {
	intervalStr, ok := config(c, "reginterval", "1440")
	if !ok {
		return
	}
	interval, err := strconv.Atoi(intervalStr)
	if err != nil || interval < 0 {
		databaseError(c, fmt.Errorf("invalid persisted registration interval %q", intervalStr), Unexpected)
		return
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

	saveConfig(c, "reginterval", strconv.FormatUint(uint64(request.RegInterval), 10))
}

func AdminGetAutoCleanUserConfig(c *gin.Context) {
	value, ok := config(c, "autoCleanUser", "false")
	if !ok {
		return
	}
	autoCleanUser := value == "true"
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
	saveConfig(c, "autoCleanUser", strconv.FormatBool(request.AutoCleanInactiveUser))
}

func AdminGetInactiveUserThresholdConfig(c *gin.Context) {
	thresholdStr, ok := config(c, "inactiveUserThreshold", "7")
	if !ok {
		return
	}
	threshold, err := strconv.Atoi(thresholdStr)
	if err != nil || threshold <= 0 {
		databaseError(c, fmt.Errorf("invalid persisted inactive user threshold %q", thresholdStr), Unexpected)
		return
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

	saveConfig(c, "inactiveUserThreshold", strconv.FormatUint(uint64(request.InactiveUserThreshold), 10))
}

func AdminCleanInactiveUser(c *gin.Context) {
	if !databaseError(c, candy.CleanInactiveUser(c.Request.Context()), Unexpected) {
		setResponseData(c, nil)
	}
}
