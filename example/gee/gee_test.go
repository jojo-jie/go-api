package gee

import (
	"log"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"
)

func onlyForV2() HandlerFunc {
	return func(c *Context) {
		t := time.Now()
		c.Fail(500, "Internal Server Error")
		log.Printf("[%d] %s in %v for group v2", c.StatusCode, c.Req.RequestURI, time.Since(t))
	}
}

func TestGeeServer(t *testing.T) {
	r := New()
	r.Use(Logger())
	r.GET("/index", func(c *Context) {
		c.HTML(http.StatusOK, "<h1>Hello Gee</h1>", nil)
	})

	v1 := r.Group("/v1")
	{
		v1.GET("/hello", func(c *Context) {
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
	v2.Use(onlyForV2())
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

func TestCond(t *testing.T) {
	var mu sync.Mutex
	c := sync.NewCond(&mu)
	var count int
	for i := 0; i < 10; i++ {
		go func() {
			c.L.Lock()
			count++
			c.L.Unlock()
			c.Broadcast()
		}()
	}

	c.L.Lock()
	for count != 10 {
		t.Log("主 goroutine 等待")
		c.Wait()
		t.Log("主 goroutine 被唤醒")
	}
	c.L.Unlock()
	t.Logf("goroutine num : %d count: %d	", runtime.NumGoroutine(), count)
}
