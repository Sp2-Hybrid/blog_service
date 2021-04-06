package routers

import (
	v1 "blog_service/internal/routers/api/v1"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	tag := v1.NewTag()
	article := v1.NewArticle()
	apiv1 := router.Group("/api/v1")
	{
		apiv1.POST("/tags", tag.Create)
		apiv1.GET("/tags", tag.List)
		apiv1.PUT("/tags/:id", tag.Update)
		apiv1.PATCH("/tags/:id/state", tag.Update)
		apiv1.DELETE("/tags/:id", tag.Delete)

		apiv1.POST("/articles", article.Create)
		apiv1.DELETE("/articles/:id", article.Delete)
		apiv1.GET("/articles/:id", article.Get)
		apiv1.GET("/articles", article.List)
		apiv1.PUT("/articles/:id", article.Update)
		apiv1.PATCH("/articles/:id/state", article.Update)
	}
	return router
}
