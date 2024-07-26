package status

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	statusMessage = make(map[int]string)
	statusMessage[Success] = "success"
	statusMessage[UnexpectedError] = "unexpected error"
	statusMessage[NotLoggedIn] = "not logged in"
	statusMessage[InvalidRequest] = "invalid request"
	statusMessage[InvalidUsername] = "invalid username"
	statusMessage[InvalidPassword] = "invalid password"
	statusMessage[RegisterTooOften] = "register too often"
	statusMessage[UsernameAlreadyTaken] = "username already taken"
	statusMessage[IncorrectUsernameOrPassword] = "incorrect username or password"
	statusMessage[NetworkAlreadyExists] = "network already exists"
	statusMessage[NetworkNotExists] = "network not exist"
	statusMessage[PermissionDenied] = "permission denied"
	statusMessage[RegistrationDisabled] = "registration disabled"
	statusMessage[InvalidNetworkName] = "invalid network name"
	statusMessage[InvalidDhcp] = "invalid dhcp"
	statusMessage[UserNotExists] = "user not exists"
	statusMessage[CannotDeleteAdmin] = "cannot delete admin"
	statusMessage[InvalidIPAddress] = "invalid ip address"
	statusMessage[RouteNotExists] = "route not exists"
	statusMessage[DeviceNotExists] = "device not exists"
	statusMessage[CannotDeleteOnlineDevice] = "cannot delete online device"
}

const (
	Success int = iota
	UnexpectedError
	NotLoggedIn
	InvalidRequest
	InvalidUsername
	InvalidPassword
	RegisterTooOften
	UsernameAlreadyTaken
	IncorrectUsernameOrPassword
	NetworkAlreadyExists
	NetworkNotExists
	PermissionDenied
	RegistrationDisabled
	InvalidNetworkName
	InvalidDhcp
	UserNotExists
	CannotDeleteAdmin
	InvalidIPAddress
	RouteNotExists
	DeviceNotExists
	CannotDeleteOnlineDevice
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
		"status": UnexpectedError,
		"msg":    msg,
	})
}

func UpdateCode(c *gin.Context, code int) {
	c.JSON(http.StatusOK, gin.H{
		"status": code,
		"msg":    statusMessage[code],
	})
}
