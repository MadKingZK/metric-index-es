package metrics

import "github.com/gin-gonic/gin"

// Init 路由配置
func Init(app *gin.Engine) {
	group := app.Group("/api/v1/metrics")
	group.POST("/write", Write)
	group.GET("/write/stats", Stats)

	// 废弃接口
	//group.POST("/", Receiver)
}
