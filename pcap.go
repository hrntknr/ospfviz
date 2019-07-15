package main

import (
	"encoding/binary"
	"net"
	"time"

	"github.com/hrntknr/gopacket"
	"github.com/hrntknr/gopacket/layers"
	"github.com/hrntknr/gopacket/pcap"
)

var (
	MAX_AGE      uint16 = 3600
	MAX_AGE_DIFF uint16 = 900
)

type OSPF struct {
	LSDB []LSDBv2Item
}

func NewOSPF() (*OSPF, error) {
	return &OSPF{
		LSDB: []LSDBv2Item{},
	}, nil
}

type LSDBv2Item struct {
	LSType      uint16      `json:"lsType"`
	ADVRouter   uint32      `json:"advRouter"`
	RouterID    string      `json:"routerID"`
	LinkStateID uint32      `json:"linkStateID"`
	LSSeqNumber uint32      `json:"lsSeqNumber"`
	LSAge       uint16      `json:"lsAge"`
	LSChecksum  uint16      `json:"lsChecksum"`
	Content     interface{} `json:"links"`
}

type RouterLSAv2 struct {
	LinkType uint8       `json:"type"`
	Link     interface{} `json:"link"`
}

type Stubv2 struct {
	Network string `json:"network"`
	Mask    string `json:"mask"`
}
type Transitv2 struct {
	DR        string `json:"dr"`
	Interface string `json:"interface"`
}
type P2Pv2 struct {
	Neighbor  string `json:"neighbor"`
	Interface string `json:"interface"`
}

type LSDBv3 struct {
	Router []RouterLinkStatev3 `json:"router"`
}

type RouterLinkStatev3 struct {
	LinkStateID string        `json:"linkStateID"`
	ADVRouter   string        `json:"advRouter"`
	Links       []interface{} `json:"links"`
}

type Transitv3 struct {
	InterfaceID         string
	NeighborInterfaceID string
	NeighborRouterID    string
}

type P2Pv3 struct {
	InterfaceID         string
	NeighborInterfaceID string
	NeighborRouterID    string
}

func int2ip(nn uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip
}

func int2mask(nn uint32) net.IPMask {
	mask := make(net.IPMask, 4)
	binary.BigEndian.PutUint32(mask, nn)
	return mask
}

func (ospf *OSPF) updateLSDBv2(lsa *layers.LSA) {
	for i, lsdbItem := range ospf.LSDB {
		if lsdbItem.LSType == lsa.LSType && lsdbItem.ADVRouter == lsa.AdvRouter && lsdbItem.LinkStateID == lsa.LinkStateID {
			if lsa.LSSeqNumber > lsdbItem.LSSeqNumber {
				ospf.updateLSDBv2Index(lsa, i)
			} else if lsa.LSSeqNumber == lsdbItem.LSSeqNumber {
				if lsa.LSChecksum > lsdbItem.LSChecksum {
					ospf.updateLSDBv2Index(lsa, i)
				} else if lsa.LSAge == MAX_AGE {
					ospf.updateLSDBv2Index(lsa, i)
				} else if lsa.LSAge-MAX_AGE_DIFF > lsdbItem.LSAge {
					ospf.updateLSDBv2Index(lsa, i)
				}
			}
			break
		}
	}
	ospf.appendLSDBv2(lsa)
}

func (ospf *OSPF) updateLSDBv2Index(lsa *layers.LSA, index int) {
	ospf.LSDB = append(ospf.LSDB[:index], ospf.LSDB[index+1:]...)
}

func (ospf *OSPF) appendLSDBv2(lsa *layers.LSA) {
	switch lsa.LSType {
	case layers.RouterLSAtypeV2:
		routerLSA := lsa.Content.(layers.RouterLSAV2)
		content := []RouterLSAv2{}
		for _, router := range routerLSA.Routers {
			switch router.Type {
			case 1:
				content = append(content, RouterLSAv2{
					LinkType: 1,
					Link: P2Pv2{
						Neighbor:  int2ip(router.LinkID).String(),
						Interface: int2ip(router.LinkData).String(),
					},
				})
				break
			case 2:
				content = append(content, RouterLSAv2{
					LinkType: 2,
					Link: Transitv2{
						DR:        int2ip(router.LinkID).String(),
						Interface: int2ip(router.LinkData).String(),
					},
				})
				break
			case 3:
				content = append(content, RouterLSAv2{
					LinkType: 3,
					Link: Stubv2{
						Network: int2ip(router.LinkID).String(),
						Mask:    int2mask(router.LinkData).String(),
					},
				})
				break
			}
		}
		ospf.LSDB = append(ospf.LSDB, LSDBv2Item{
			LSType:      lsa.LSType,
			ADVRouter:   lsa.AdvRouter,
			RouterID:    int2ip(lsa.AdvRouter).String(),
			LinkStateID: lsa.LinkStateID,
			LSSeqNumber: lsa.LSSeqNumber,
			LSAge:       lsa.LSAge,
			LSChecksum:  lsa.LSChecksum,
			Content:     content,
		})
		break
	}
}

func (ospf *OSPF) StartPcap(pcapIf string) error {
	handle, err := pcap.OpenLive(pcapIf, 0xffff, true, pcap.BlockForever)
	if err != nil {
		return err
	}

	go func() {
		defer handle.Close()

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			if ospfLayer := packet.Layer(layers.LayerTypeOSPF); ospfLayer != nil {
				switch ospfPacket := ospfLayer.(type) {
				case *layers.OSPFv2:
					switch ospfPacket.Type {
					case layers.OSPFLinkStateUpdate:
						lsu := ospfPacket.Content.(layers.LSUpdate)
						for _, lsa := range lsu.LSAs {
							ospf.updateLSDBv2(&lsa)
						}
						break
					}
					break

					// case *layers.OSPFv3:
					// 	switch ospfPacket.Type {
					// 	case layers.OSPFLinkStateUpdate:
					// 		lsu := ospfPacket.Content.(layers.LSUpdate)
					// 		for _, lsa := range lsu.LSAs {
					// 			switch lsa.LSType {
					// 			case layers.RouterLSAtype:
					// 				routerLSA := lsa.Content.(layers.RouterLSA)
					// 				for _, router := range routerLSA.Routers {
					// 					switch router.Type {
					// 					case 1:
					// 						p2p := P2Pv3{
					// 							InterfaceID:         int2ip(router.InterfaceID),
					// 							NeighborInterfaceID: int2ip(router.NeighborInterfaceID),
					// 							NeighborRouterID:    int2ip(router.NeighborRouterID),
					// 						}
					// 						break
					// 					case 2:
					// 						transit := Transitv3{
					// 							InterfaceID:         int2ip(router.InterfaceID),
					// 							NeighborInterfaceID: int2ip(router.NeighborInterfaceID),
					// 							NeighborRouterID:    int2ip(router.NeighborRouterID),
					// 						}
					// 						break
					// 					}
					// 				}
					// 				break
					// 			}
					// 		}
					// 		break
					// 	}
					// 	break
				}
			}
		}
	}()
	go func() {
		for {
			time.Sleep(time.Second * 1)
			offset := 0
			for i := range ospf.LSDB {
				ospf.LSDB[i-offset].LSAge++
				if ospf.LSDB[i-offset].LSAge >= MAX_AGE {
					ospf.LSDB = append(ospf.LSDB[:i-offset], ospf.LSDB[i-offset+1:]...)
					offset++
				}
			}
			// body, err := json.Marshal(ospf.LSDB)
			// if err != nil {
			// 	fmt.Printf("%s\n", err)
			// }
			// var buf bytes.Buffer
			// err = json.Indent(&buf, body, "", "  ")
			// if err != nil {
			// 	panic(err)
			// }
			// fmt.Printf("%s\n", buf.String())
		}
	}()
	return nil
}
