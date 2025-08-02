package server

import (
	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the API routes for the multitenant server
func SetupRoutes(r *gin.Engine) {
	v1 := r.Group("api/v1")
	{
		v1.POST("/sessions", createSession)
		v1.GET("/sessions", listSessions)
		v1.GET("/sessions/:sid", getSession)
		v1.GET("/sessions/:sid/debug", getDebug)
		v1.DELETE("/sessions/:sid", deleteSession)

		sess := v1.Group("/sessions/:sid")
		{
			sess.POST("/observe", observe)
			sess.POST("/inspect", inspect)
			sess.POST("/uncover", uncover)
			sess.POST("/unlock", unlock)
			sess.POST("/search", search)
			sess.POST("/take", take)
			sess.POST("/inventory", inventory)
			sess.POST("/heal", heal)
			sess.POST("/traverse", traverse)
			sess.POST("/battle", battle)
			sess.POST("/combine", combine)
			sess.POST("/use", use)
		}
	}
}
