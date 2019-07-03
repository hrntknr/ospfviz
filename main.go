package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
)

func main() {
	err := _main()
	if err != nil {
		fmt.Printf("%s", err)
	}
}

func _main() error {
	var (
		config = flag.String("config", "router_database", "router database file.(show ip ospf database router)")
	)
	flag.Parse()
	fd, err := os.Open(*config)
	if err != nil {
		return err
	}
	defer fd.Close()
	err = configParser(fd)
	if err != nil {
		return err
	}

	return nil
}

type LinkType uint8

const (
	StubNetwork    LinkType = 0
	TransitNetwork LinkType = 1
	P2PNetwork     LinkType = 2
)

type Router struct {
	RouterID string `vyos:"Advertising Router"`
	Links    []Link
}

type Link struct {
	Type    LinkType
	Stub    StubInfo
	Transit TransitInfo
	P2P     P2PInfo
}
type StubInfo struct {
	Network string `vyos:"(Link ID) Net"`
	Mask    string `vyos:"(Link Data) Network Mask"`
	Cost    int    `vyos:"TOS 0 Metric"`
}
type TransitInfo struct {
	DR        string `vyos:"(Link ID) Designated Router address"`
	Interface string `vyos:"(Link Data) Router Interface address"`
	Cost      int    `vyos:"TOS 0 Metric"`
}
type P2PInfo struct {
	Neighbor  string `vyos:"(Link ID) Neighboring Router ID"`
	Interface string `vyos:"(Link Data) Router Interface address"`
	Cost      int    `vyos:"TOS 0 Metric"`
}

var indentMatch = regexp.MustCompile(`^(\s)*`)
var attrMatch = regexp.MustCompile(`^(.+)(:\s(.+))+`)
