package main

import (
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
	LSDBv2 []LSDBItem
	LSDBv3 []LSDBItem
}

func NewOSPF() (*OSPF, error) {
	return &OSPF{
		LSDBv2: []LSDBItem{},
	}, nil
}

type LSDBItem struct {
	LSType      uint16      `json:"lsType"`
	ADVRouter   uint32      `json:"advRouter"`
	LinkStateID uint32      `json:"linkStateID"`
	LSSeqNumber uint32      `json:"lsSeqNumber"`
	LSAge       uint16      `json:"lsAge"`
	LSChecksum  uint16      `json:"lsChecksum"`
	Content     interface{} `json:"links"`
}

type RouterLSAv2 struct {
	LinkType uint8  `json:"type"`
	Link     Linkv2 `json:"link"`
}

type Linkv2 struct {
	LinkID    uint32 `json:"linkID"`
	LinikData uint32 `json:"linkData"`
}

type RouterLSAv3 struct {
	LinkType uint8  `json:"type"`
	Link     Linkv3 `json:"link"`
}
type Linkv3 struct {
	InterfaceID         uint32 `json:"interfaceID"`
	NeighborInterfaceID uint32 `json:"neighborInterfaceID"`
	NeighborADVRouter   uint32 `json:"neighborADVRouter"`
}

func (ospf *OSPF) updateLSDBv2(lsa *layers.LSA) {
	appendFlag := true
	for i, lsdbItem := range ospf.LSDBv2 {
		if lsdbItem.LSType == lsa.LSType && lsdbItem.ADVRouter == lsa.AdvRouter && lsdbItem.LinkStateID == lsa.LinkStateID {
			if lsa.LSSeqNumber >= lsdbItem.LSSeqNumber {
				ospf.updateLSDBIndexv2(lsa, i)
			} else if lsa.LSSeqNumber == lsdbItem.LSSeqNumber {
				if lsa.LSChecksum > lsdbItem.LSChecksum {
					ospf.updateLSDBIndexv2(lsa, i)
				} else if lsa.LSAge == MAX_AGE {
					ospf.updateLSDBIndexv2(lsa, i)
				} else if lsa.LSAge-MAX_AGE_DIFF > lsdbItem.LSAge {
					ospf.updateLSDBIndexv2(lsa, i)
				} else {
					appendFlag = false
				}
			}
			break
		}
	}
	if appendFlag {
		ospf.appendLSDBv2(lsa)
	}
}

func (ospf *OSPF) updateLSDBIndexv2(lsa *layers.LSA, index int) {
	ospf.LSDBv2 = append(ospf.LSDBv2[:index], ospf.LSDBv2[index+1:]...)
}

func (ospf *OSPF) updateLSDBv3(lsa *layers.LSA) {
	appendFlag := true
	for i, lsdbItem := range ospf.LSDBv3 {
		if lsdbItem.LSType == lsa.LSType && lsdbItem.ADVRouter == lsa.AdvRouter && lsdbItem.LinkStateID == lsa.LinkStateID {
			if lsa.LSSeqNumber > lsdbItem.LSSeqNumber {
				ospf.updateLSDBIndexv3(lsa, i)
			} else if lsa.LSSeqNumber == lsdbItem.LSSeqNumber {
				if lsa.LSChecksum > lsdbItem.LSChecksum {
					ospf.updateLSDBIndexv3(lsa, i)
				} else if lsa.LSAge == MAX_AGE {
					ospf.updateLSDBIndexv3(lsa, i)
				} else if lsa.LSAge-MAX_AGE_DIFF > lsdbItem.LSAge {
					ospf.updateLSDBIndexv3(lsa, i)
				} else {
					appendFlag = false
				}
			}
			break
		}
	}
	if appendFlag {
		ospf.appendLSDBv3(lsa)
	}
}

func (ospf *OSPF) updateLSDBIndexv3(lsa *layers.LSA, index int) {
	ospf.LSDBv3 = append(ospf.LSDBv3[:index], ospf.LSDBv3[index+1:]...)
}

func (ospf *OSPF) appendLSDBv2(lsa *layers.LSA) {
	switch lsa.LSType {
	case layers.RouterLSAtypeV2:
		routerLSA := lsa.Content.(layers.RouterLSAV2)
		content := []RouterLSAv2{}
		for _, router := range routerLSA.Routers {
			content = append(content, RouterLSAv2{
				LinkType: router.Type,
				Link: Linkv2{
					LinkID:    router.LinkID,
					LinikData: router.LinkData,
				},
			})
		}
		ospf.LSDBv2 = append(ospf.LSDBv2, LSDBItem{
			LSType:      lsa.LSType,
			ADVRouter:   lsa.AdvRouter,
			LinkStateID: lsa.LinkStateID,
			LSSeqNumber: lsa.LSSeqNumber,
			LSAge:       lsa.LSAge,
			LSChecksum:  lsa.LSChecksum,
			Content:     content,
		})
		break
	}
}

func (ospf *OSPF) appendLSDBv3(lsa *layers.LSA) {
	switch lsa.LSType {
	case layers.RouterLSAtype:
		routerLSA := lsa.Content.(layers.RouterLSA)
		content := []RouterLSAv3{}
		for _, router := range routerLSA.Routers {
			content = append(content, RouterLSAv3{
				LinkType: router.Type,
				Link: Linkv3{
					NeighborADVRouter:   router.NeighborRouterID,
					NeighborInterfaceID: router.NeighborInterfaceID,
					InterfaceID:         router.InterfaceID,
				},
			})
		}
		ospf.LSDBv3 = append(ospf.LSDBv3, LSDBItem{
			LSType:      lsa.LSType,
			ADVRouter:   lsa.AdvRouter,
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

				case *layers.OSPFv3:
					switch ospfPacket.Type {
					case layers.OSPFLinkStateUpdate:
						lsu := ospfPacket.Content.(layers.LSUpdate)
						for _, lsa := range lsu.LSAs {
							ospf.updateLSDBv3(&lsa)
						}
						break
					}
					break
				}
			}
		}
	}()
	go func() {
		for {
			time.Sleep(time.Second * 1)
			var offset int
			offset = 0
			for i := range ospf.LSDBv2 {
				ospf.LSDBv2[i-offset].LSAge++
				if ospf.LSDBv2[i-offset].LSAge >= MAX_AGE {
					ospf.LSDBv2 = append(ospf.LSDBv2[:i-offset], ospf.LSDBv2[i-offset+1:]...)
					offset++
				}
			}
			offset = 0
			for i := range ospf.LSDBv3 {
				ospf.LSDBv3[i-offset].LSAge++
				if ospf.LSDBv3[i-offset].LSAge >= MAX_AGE {
					ospf.LSDBv3 = append(ospf.LSDBv3[:i-offset], ospf.LSDBv3[i-offset+1:]...)
					offset++
				}
			}
		}
	}()
	return nil
}
