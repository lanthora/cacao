package api

import (
	"crypto/sha256"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/model"
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
			setErrorCode(c, NotLoggedIn)
			c.Abort()
			return
		}
		id, err := strconv.ParseUint(idstr, 10, 64)
		if err != nil {
			setErrorCode(c, NotLoggedIn)
			c.Abort()
			return
		}
		user := &model.User{}
		user.ID = uint(id)

		found, err := model.Authenticate(user.ID, token, c.Request.Context())
		if databaseError(c, err, NotLoggedIn) {
			c.Abort()
			return
		}
		*user = found
		if user.Token != token {
			setErrorCode(c, NotLoggedIn)
			c.Abort()
			return
		}
		c.Set("user", user)
		c.Next()
	}
}

func UserInfo(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	setResponseData(c, gin.H{
		"name":    user.Name,
		"role":    user.Role,
		"regtime": user.CreatedAt.Format(time.DateTime),
	})
}

func UserStatistics(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	stats, err := model.GetUserStats(user.ID, c.Request.Context())
	if databaseError(c, err, Unexpected) {
		return
	}
	setResponseData(c, gin.H{"netnum": stats.NetNum, "devnum": stats.DevNum, "rxsum": stats.RX, "txsum": stats.TX})
}

func UserRegister(c *gin.Context) {

	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}
	if request.Username == "@" {
		setErrorCode(c, InvalidUsername)
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
	user, network, err := model.RegisterUser(model.User{
		Name: request.Username, Password: hashUserPassword(request.Username, request.Password), Token: uuid.NewString(), IP: c.ClientIP(),
	}, model.Net{Name: "@", Password: randomString(8), DHCP: "192.168.202.0/24", Broadcast: true}, false, c.Request.Context())
	if databaseError(c, err, Unexpected) {
		return
	}
	if network != nil {
		candy.InsertNet(network)
	}
	c.SetCookie("id", strconv.FormatUint(uint64(user.ID), 10), 86400, "/", "", false, true)
	c.SetCookie("token", user.Token, 86400, "/", "", false, true)
	setResponseData(c, gin.H{"name": user.Name, "role": user.Role})
}

func UserLogin(c *gin.Context) {
	var request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}

	user, err := model.FindUser(request.Username, c.Request.Context())
	if databaseError(c, err, IncorrectUsernameOrPassword) {
		return
	}
	if user.Password != hashUserPassword(request.Username, request.Password) {
		setErrorCode(c, IncorrectUsernameOrPassword)
		return
	}
	ip := user.IP
	if ip == "" {
		ip = c.ClientIP()
	}
	if databaseError(c, user.UpdateCredentials(user.Password, uuid.NewString(), ip, c.Request.Context()), Unexpected) {
		return
	}

	c.SetCookie("id", strconv.FormatUint(uint64(user.ID), 10), 86400, "/", "", false, true)
	c.SetCookie("token", user.Token, 86400, "/", "", false, true)

	setResponseData(c, gin.H{
		"name": user.Name,
		"role": user.Role,
	})
}

func UserLogout(c *gin.Context) {
	user := c.MustGet("user").(*model.User)
	if databaseError(c, user.UpdateCredentials(user.Password, uuid.NewString(), user.IP, c.Request.Context()), Unexpected) {
		return
	}

	c.SetCookie("id", "", -1, "/", "", false, true)
	c.SetCookie("token", "", -1, "/", "", false, true)

	setResponseData(c, nil)
}

func ChangePassword(c *gin.Context) {
	var request struct {
		OldPassword string `json:"old"`
		NewPassword string `json:"new"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		setErrorCode(c, InvalidRequest)
		return
	}

	user := c.MustGet("user").(*model.User)

	if user.Password != hashUserPassword(user.Name, request.OldPassword) {
		setErrorCode(c, IncorrectUsernameOrPassword)
		return
	}
	if len(request.NewPassword) == 0 {
		setErrorCode(c, InvalidPassword)
		return
	}

	if databaseError(c, user.UpdateCredentials(hashUserPassword(user.Name, request.NewPassword), uuid.NewString(), user.IP, c.Request.Context()), Unexpected) {
		return
	}

	c.SetCookie("id", strconv.FormatUint(uint64(user.ID), 10), 86400, "/", "", false, true)
	c.SetCookie("token", user.Token, 86400, "/", "", false, true)

	setResponseData(c, nil)
}

func hashUserPassword(username, password string) string {
	hash := sha256.Sum256([]byte(username + ":" + password))
	return fmt.Sprintf("%x", hash[:])
}

func randomString(n int) string {
	letters := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[r.Intn(len(letters))]
	}
	return string(b)
}
