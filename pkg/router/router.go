package router

import (
	"github.com/gin-gonic/gin"
	"github.com/mangohow/cloud-ide/pkg/logger"
)

func NewGinRouter(mode string, middlewares ...gin.HandlerFunc) *gin.Engine {
	var router *gin.Engine
	if mode == "dev" {
		router = gin.Default()
	} else {
		gin.SetMode(gin.ReleaseMode)
		router = gin.New()
	}
	if len(middlewares) > 0 {
		router.Use(gin.RecoveryWithWriter(logger.Output()))
		router.Use(middlewares...)
	}

	return router
}
