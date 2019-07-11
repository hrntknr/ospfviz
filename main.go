package main

import (
	"flag"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	var (
		pcapIf = flag.String("i", "eth0", "inetrface")
	)
	flag.Parse()

	err := startPcap(*pcapIf)
	if err != nil {
		panic(err)
	}

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.LoadHTMLGlob("./templates/*.html")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "index.html", gin.H{})
	})

	r.StaticFile("/build.js", "./static/build.js")

	r.GET("/api/ospf", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "not implement exception",
		})
	})

	r.Run()
}
