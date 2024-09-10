package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	statusMessage = make(map[int]string)
	statusMessage[Success] = "success"
	statusMessage[Unexpected] = "unexpected"
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
	statusMessage[InvalidInactiveUserThreshold] = "invalid inactive user threshold"
}

const (
	Success int = iota
	Unexpected
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
	InvalidInactiveUserThreshold
)

var statusMessage map[int]string

func setResponse(c *gin.Context, code int, msg string, data gin.H) {
	c.JSON(http.StatusOK, gin.H{
		"status": code,
		"msg":    msg,
		"data":   data,
	})
}

func setErrorCode(c *gin.Context, code int) {
	setResponse(c, code, statusMessage[code], nil)
}

func setUnexpectedMessage(c *gin.Context, msg string) {
	setResponse(c, Unexpected, msg, nil)
}

func setResponseData(c *gin.Context, data gin.H) {
	setResponse(c, Success, statusMessage[Success], data)
}
