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

type LinkType uint8

const (
	StubNetwork    LinkType = 0
	TransitNetwork LinkType = 1
	P2PNetwork     LinkType = 2
)

type Router struct {
	RouterID string   `json:"routerID"`
	HostName []string `json:"hostname"`
	Links    []Link   `json:"links"`
}

type Link struct {
	Type    LinkType     `json:"type"`
	Stub    *StubInfo    `json:"stub,omitempty"`
	Transit *TransitInfo `json:"transit,omitempty"`
	P2P     *P2PInfo     `json:"p2p,omitempty"`
}
type StubInfo struct {
	Network string `json:"network"`
	Mask    string `json:"mask"`
	Cost    int    `json:"cost"`
}
type TransitInfo struct {
	DR        string `json:"dr"`
	Interface string `json:"interface"`
	Cost      int    `json:"cost"`
}
type P2PInfo struct {
	Neighbor  string `json:"neighbor"`
	Interface string `json:"interface"`
	Cost      int    `json:"cost"`
}
