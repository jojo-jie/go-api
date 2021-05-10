package routers

import (
	"blog/global"
	"blog/internal/middleware"
	"blog/internal/routers/api"
	v1 "blog/internal/routers/api/v1"
	"github.com/gin-gonic/gin"
	"net/http"
)

func NewRouter() *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.Translations())
	upload:=api.NewUpload()
	r.POST("/upload/file", upload.UploadFile)
	//文件服务只有提供静态资源的访问，才能在外部请求本项目HTTP Server时同时提供静态资源的访问
	r.StaticFS("/static", http.Dir(global.AppSetting.UploadSavePath))
	r.StaticFS("/doc", http.Dir(global.AppSetting.UploadDocSavePath))
	apiv1 := r.Group("api/v1/")
	{
		tag := v1.NewTag()
		apiv1.POST("tags", tag.Create)
		apiv1.DELETE("tags/:id", tag.Delete)
		apiv1.PUT("tags/:id", tag.Update)
		apiv1.PATCH("tags/:id/state", tag.Update)
		apiv1.GET("tags", tag.List)

		article := v1.NewArticle()
		apiv1.POST("articles", article.Create)
		apiv1.DELETE("articles/:id", article.Delete)
		apiv1.PUT("articles/:id", article.Update)
		apiv1.PATCH("articles/:id/state", article.Update)
		apiv1.GET("articles/:id", article.Get)
		apiv1.GET("articles", article.List)
	}
	return r
}
