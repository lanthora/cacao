package main

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/api"
	"github.com/lanthora/cacao/argp"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/frontend"
	"github.com/lanthora/cacao/logger"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

func main() {
	apikey := argp.Get("apikey", ".")
	apicrt := argp.Get("apicrt", ".")
	tlsEnable := apikey != "." && apicrt != "." && fileExists(apikey) && fileExists(apicrt)

	addr := argp.Get("listen", ":80")
	logger.Info("listen=[%v],https=[%v],apikey=[%v],apicrt=[%v]", addr, tlsEnable, apikey, apicrt)

	r := gin.New()
	r.Use(candy.WebsocketMiddleware(), api.LoginMiddleware(), api.AdminMiddleware())

	admin := r.Group("/api/admin")
	admin.POST("/showUsers", api.AdminShowUsers)
	admin.POST("/addUser", api.AdminAddUser)
	admin.POST("/deleteUser", api.AdminDeleteUser)
	admin.POST("/updateUserPassword", api.AdminUpdateUserPassword)
	admin.POST("/getOpenRegisterConfig", api.AdminGetOpenRegisterConfig)
	admin.POST("/setOpenRegisterConfig", api.AdminSetOpenRegisterConfig)
	admin.POST("/getRegisterIntervalConfig", api.AdminGetRegisterIntervalConfig)
	admin.POST("/setRegisterIntervalConfig", api.AdminSetRegisterIntervalConfig)
	admin.POST("/getAutoCleanUserConfig", api.AdminGetAutoCleanUserConfig)
	admin.POST("/setAutoCleanUserConfig", api.AdminSetAutoCleanUserConfig)
	admin.POST("/getInactiveUserThresholdConfig", api.AdminGetInactiveUserThresholdConfig)
	admin.POST("/setInactiveUserThresholdConfig", api.AdminSetInactiveUserThresholdConfig)
	admin.POST("/cleanInactiveUser", api.AdminCleanInactiveUser)

	user := r.Group("/api/user")
	user.POST("/info", api.UserInfo)
	user.POST("/statistics", api.UserStatistics)
	user.POST("/register", api.UserRegister)
	user.POST("/login", api.UserLogin)
	user.POST("/changePassword", api.ChangePassword)
	user.POST("/logout", api.UserLogout)

	net := r.Group("/api/net")
	net.POST("/show", api.NetShow)
	net.POST("/insert", api.NetInsert)
	net.POST("/edit", api.NetEdit)
	net.POST("/delete", api.NetDelete)

	device := r.Group("/api/device")
	device.POST("/show", api.DeviceShow)
	device.POST("/delete", api.DeviceDelete)

	route := r.Group("/api/route")
	route.POST("/show", api.RouteShow)
	route.POST("/insert", api.RouteInsert)
	route.POST("/delete", api.RouteDelete)

	r.NoRoute(frontend.Static)

	var err error
	if tlsEnable {
		err = r.RunTLS(addr, apicrt, apikey)
	} else {
		err = r.Run(addr)
	}
	if err != nil {
		logger.Fatal("service run failed: %v", err)
	}
}
