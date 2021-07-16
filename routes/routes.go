package routes

import (
	"monica-adaptor/controllers/metrics"
	"monica-adaptor/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

// InitRoute 路由配置
func InitRoute(app *gin.Engine) {
	app.Use(logger.GinLogger(), logger.GinRecovery(true))

	app.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	app.Any("/health", func(ctx *gin.Context) { //healthCheck
		ctx.String(http.StatusOK, "SUCCESS")
	})

	// 注册controller/metrics的route
	metrics.InitRoute(app)
}
