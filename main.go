package main

import (
	"github.com/gin-gonic/gin"
	"github.com/lanthora/cacao/argp"
	"github.com/lanthora/cacao/logger"
	"github.com/lanthora/cacao/user"
)

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func main() {
	addr := argp.Get("listen", ":80")
	logger.Info("listen=[%v]", addr)

	r := gin.New()
	r.Use(user.LoginMiddleware())
	r.POST("/api/user/register", user.Register)
	r.POST("/api/user/login", user.Login)
	r.POST("/api/user/logout", user.Logout)

	if err := r.Run(addr); err != nil {
		logger.Fatal("service run failed: %v", err)
	}
}
