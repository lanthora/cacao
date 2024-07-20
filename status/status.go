package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	statusMessage = make(map[int]string)
	statusMessage[Success] = "success"
	statusMessage[Unexpected] = "unexpected"
	statusMessage[NotLoggedIn] = "not logged in"
	statusMessage[InvalidRequest] = "invalid request "
	statusMessage[InvalidUsername] = "invalid username"
	statusMessage[InvalidPassword] = "invalid password"
	statusMessage[RegisterTooFrequently] = "register too frequently"
	statusMessage[UsernameAlreadyTaken] = "username already taken"
	statusMessage[UsernameOrPasswordIncorrect] = "username or password is incorret"
	statusMessage[NetworkAlreadyExists] = "network already exists"
	statusMessage[NetworkNotExists] = "network not exists"
	statusMessage[AdminAccessRequired] = "admin access required"
	statusMessage[RegistrationDisabled] = "registration disabled"
}

const (
	Success int = iota
	Unexpected
	NotLoggedIn
	InvalidRequest
	InvalidUsername
	InvalidPassword
	RegisterTooFrequently
	UsernameAlreadyTaken
	UsernameOrPasswordIncorrect
	NetworkAlreadyExists
	NetworkNotExists
	AdminAccessRequired
	RegistrationDisabled
)

var statusMessage map[int]string

func UpdateSuccess(c *gin.Context, data gin.H) {
	c.JSON(http.StatusOK, gin.H{
		"status": Success,
		"msg":    statusMessage[Success],
		"data":   data,
	})
}

func UpdateUnexpected(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{
		"status": Unexpected,
		"msg":    msg,
	})
}

func UpdateCode(c *gin.Context, code int) {
	c.JSON(http.StatusOK, gin.H{
		"status": code,
		"msg":    statusMessage[code],
	})
}
