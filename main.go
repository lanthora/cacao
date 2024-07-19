package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/api"
	"github.com/lanthora/cacao/argp"
	"github.com/lanthora/cacao/candy"
	"github.com/lanthora/cacao/logger"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	addr := argp.Get("listen", ":80")
	logger.Info("listen=[%v]", addr)

	r := gin.New()
	r.Use(candy.WebsocketMiddleware(), api.LoginMiddleware())

	user := r.Group("/api/user")
	user.POST("/register", api.UserRegister)
	user.POST("/login", api.UserLogin)
	user.POST("/logout", api.UserLogout)

	net := r.Group("/api/net")
	net.POST("/show", api.NetShow)
	net.POST("/insert", api.NetInsert)
	net.POST("/edit", api.NetEdit)
	net.POST("/delete", api.NetDelete)

	if err := r.Run(addr); err != nil {
		logger.Fatal("service run failed: %v", err)
	}
}
