package gee

import (
	"net/http"
	"testing"
)

func TestGeeServer(t *testing.T) {
	r := New()

	r.GET("/index", func(c *Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>")
	})
	v1 := r.Group("/v1")
	{
		v1.GET("/hello", func(c *Context) {
			// expect /hello?name=geektutu
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Query("name"), c.Path)
		})

		v1.POST("/login", func(c *Context) {
			c.JSON(http.StatusOK, H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	}

	v2 := r.Group("/v2")
	{
		v2.GET("/hello/:name", func(c *Context) {
			c.String(http.StatusOK, "hello %s, you're at %s\n", c.Param("name"), c.Path)
		})
		v2.POST("/login", func(c *Context) {
			c.JSON(http.StatusOK, H{
				"username": c.PostForm("username"),
				"password": c.PostForm("password"),
			})
		})
	}

	r.Run(":9000")
}
