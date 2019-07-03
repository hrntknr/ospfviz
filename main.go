package main

import (
	"encoding/json"
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
	routers, err := configParser(fd)
	if err != nil {
		return err
	}

	body, err := json.Marshal(routers)
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", body)

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
