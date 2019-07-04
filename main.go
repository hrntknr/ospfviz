package main

import (
	"net/http"
	"os"
	"regexp"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.LoadHTMLGlob("./templates/*.html")

	r.GET("/", func(ctx *gin.Context) {
		ctx.HTML(200, "index.html", gin.H{})
	})

	r.StaticFile("/build.js", "./static/build.js")

	r.GET("/api/ospf", func(c *gin.Context) {
		fd, err := os.Open("router_database")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}
		defer fd.Close()
		routers, err := configParser(fd)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}
		c.JSON(200, routers)
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
	RouterID string `json:"routerID" vyos:"Advertising Router"`
	Links    []Link `json:"links"`
}

type Link struct {
	Type    LinkType     `json:"type"`
	Stub    *StubInfo    `json:"stub,omitempty"`
	Transit *TransitInfo `json:"transit,omitempty"`
	P2P     *P2PInfo     `json:"p2p,omitempty"`
}
type StubInfo struct {
	Network string `json:"network" vyos:"(Link ID) Net"`
	Mask    string `json:"mask"    vyos:"(Link Data) Network Mask"`
	Cost    int    `json:"cost"    vyos:"TOS 0 Metric"`
}
type TransitInfo struct {
	DR        string `json:"dr"        vyos:"(Link ID) Designated Router address"`
	Interface string `json:"interface" vyos:"(Link Data) Router Interface address"`
	Cost      int    `json:"cost"      vyos:"TOS 0 Metric"`
}
type P2PInfo struct {
	Neighbor  string `json:"neighbor"  vyos:"(Link ID) Neighboring Router ID"`
	Interface string `json:"interface" vyos:"(Link Data) Router Interface address"`
	Cost      int    `json:"cost"      vyos:"TOS 0 Metric"`
}

var indentMatch = regexp.MustCompile(`^(\s)*`)
var attrMatch = regexp.MustCompile(`^(.+)(:\s(.+))+`)
