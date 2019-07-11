package main

import (
	"encoding/binary"
	"fmt"
	"net"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type LSDBv2 struct {
	Router []RouterLinkStatev2 `json:"router"`
}

type RouterLinkStatev2 struct {
	LinkStateID net.IP        `json:"linkStateID"`
	ADVRouter   net.IP        `json:"advRouter"`
	Links       []interface{} `json:"links"`
}

type Stubv2 struct {
	Network net.IP
	Mask    net.IPMask
}
type Transitv2 struct {
	DR        net.IP
	Interface net.IP
}
type P2Pv2 struct {
	Neighbor  net.IP
	Interface net.IP
}

type LSDBv3 struct {
	Router []RouterLinkStatev3 `json:"router"`
}

type RouterLinkStatev3 struct {
	LinkStateID net.IP        `json:"linkStateID"`
	ADVRouter   net.IP        `json:"advRouter"`
	Links       []interface{} `json:"links"`
}

type Transitv3 struct {
	InterfaceID         net.IP
	NeighborInterfaceID net.IP
	NeighborRouterID    net.IP
}

type P2Pv3 struct {
	InterfaceID         net.IP
	NeighborInterfaceID net.IP
	NeighborRouterID    net.IP
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

func startPcap(pcapIf string) error {
	handle, err := pcap.OpenLive(pcapIf, 0xffff, true, pcap.BlockForever)
	if err != nil {
		return err
	}

	go func() {
		defer handle.Close()

		packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
		for packet := range packetSource.Packets() {
			if ospfLayer := packet.Layer(layers.LayerTypeOSPF); ospfLayer != nil {
				switch ospf := ospfLayer.(type) {
				case *layers.OSPFv2:
					switch ospf.Type {
					// case layers.OSPFHello:
					// 	hello := ospf.Content.(layers.HelloPkgV2)
					// 	fmt.Println(hello)
					// 	break
					// case layers.OSPFDatabaseDescription:
					// 	dbd := ospf.Content.([]layers.LSAheader)
					// 	fmt.Println(dbd)
					// 	break
					// case layers.OSPFLinkStateRequest:
					// 	lsr := ospf.Content.([]layers.LSReq)
					// 	fmt.Println(lsr)
					// 	break
					case layers.OSPFLinkStateUpdate:
						lsu := ospf.Content.(layers.LSUpdate)
						for _, lsa := range lsu.LSAs {
							switch lsa.LSType {
							case layers.RouterLSAtypeV2:
								routerLSA := lsa.Content.(layers.RouterLSAV2)
								fmt.Printf("Version: %d\n", ospf.Version)
								fmt.Printf("RouterID: %d\n", ospf.RouterID)
								fmt.Printf("AreaID: %d\n", ospf.AreaID)
								fmt.Printf("Type: %s\n", ospf.Type)
								fmt.Printf("LSSeqNumber: %x\n", lsa.LSSeqNumber)
								fmt.Printf("Links: %d\n", routerLSA.Links)
								fmt.Printf("LSAType: RouterLSA\n")
								for _, router := range routerLSA.Routers {
									fmt.Printf("\n")
									fmt.Printf("  Type: %d\n", router.Type)
									fmt.Printf("  LinkID: %x\n", router.LinkID)
									fmt.Printf("  LinkData: %x\n", router.LinkData)
									fmt.Printf("  Metric: %d\n", router.Metric)
									switch router.Type {
									case 1:
										p2p := P2Pv2{}
										p2p.Neighbor = int2ip(router.LinkID)
										p2p.Interface = int2ip(router.LinkData)
										fmt.Println(p2p)
										break
									case 2:
										transit := Transitv2{}
										transit.DR = int2ip(router.LinkID)
										transit.Interface = int2ip(router.LinkData)
										fmt.Println(transit)
										break
									case 3:
										stub := Stubv2{}
										stub.Network = int2ip(router.LinkID)
										stub.Mask = int2mask(router.LinkData)
										fmt.Println(stub)
										break
									}
								}
								fmt.Printf("\n\n")
								break
							case layers.ASExternalLSAtypeV2:
								// externalLSA := lsa.Content.(layers.ASExternalLSAV2)
								// fmt.Printf("Version: %d\n", ospf.Version)
								// fmt.Printf("RouterID: %d\n", ospf.RouterID)
								// fmt.Printf("AreaID: %d\n", ospf.AreaID)
								// fmt.Printf("Type: %s\n", ospf.Type)
								// fmt.Printf("LSAType: ASExternalLSA\n")
								// fmt.Printf("LinkStateID: %x\n", lsa.LinkStateID)
								// fmt.Printf("  Metric: %d\n", externalLSA.Metric)
								// fmt.Printf("  Mask: %x\n", externalLSA.NetworkMask)
								// fmt.Printf("  ForwardingAddress: %d\n", externalLSA.ForwardingAddress)
								// fmt.Printf("  ExternalRouteTag: %d\n", externalLSA.ExternalRouteTag)
								// fmt.Printf("\n\n")
								// break
							default:
								break
							}
						}
						break
						// case layers.OSPFLinkStateAcknowledgment:
						// 	lsack := ospf.Content.([]layers.LSAheader)
						// 	fmt.Println(lsack)
						// 	break
					}
					break

				case *layers.OSPFv3:
					switch ospf.Type {
					// case layers.OSPFHello:
					// 	hello := ospf.Content.(layers.HelloPkg)
					// 	fmt.Println(hello)
					// 	break
					// case layers.OSPFDatabaseDescription:
					// 	dbd := ospf.Content.([]layers.LSAheader)
					// 	fmt.Println(dbd)
					// 	break
					// case layers.OSPFLinkStateRequest:
					// 	lsr := ospf.Content.([]layers.LSReq)
					// 	fmt.Println(lsr)
					// 	break
					case layers.OSPFLinkStateUpdate:
						lsu := ospf.Content.(layers.LSUpdate)
						for _, lsa := range lsu.LSAs {
							switch lsa.LSType {
							case layers.RouterLSAtype:
								routerLSA := lsa.Content.(layers.RouterLSA)
								fmt.Printf("Version: %d\n", ospf.Version)
								fmt.Printf("RouterID: %d\n", ospf.RouterID)
								fmt.Printf("AreaID: %d\n", ospf.AreaID)
								fmt.Printf("Type: %s\n", ospf.Type)
								fmt.Printf("LSAType: RouterLSA\n")
								for _, router := range routerLSA.Routers {
									fmt.Printf("\n")
									fmt.Printf("  Type: %d\n", router.Type)
									fmt.Printf("  Metric: %d\n", router.Metric)
									fmt.Printf("  InterfaceID: %x\n", router.InterfaceID)
									fmt.Printf("  NeighborInterfaceID: %x\n", router.NeighborInterfaceID)
									fmt.Printf("  NeighborRouterID: %x\n", router.NeighborRouterID)
									switch router.Type {
									case 1:
										p2p := P2Pv3{}
										p2p.InterfaceID = int2ip(router.InterfaceID)
										p2p.NeighborInterfaceID = int2ip(router.NeighborInterfaceID)
										p2p.NeighborRouterID = int2ip(router.NeighborRouterID)
										fmt.Println(p2p)
										break
									case 2:
										transit := Transitv3{}
										transit.InterfaceID = int2ip(router.InterfaceID)
										transit.NeighborInterfaceID = int2ip(router.NeighborInterfaceID)
										transit.NeighborRouterID = int2ip(router.NeighborRouterID)
										fmt.Println(transit)
										break
									}
								}
								fmt.Printf("\n\n")
								break
							// case layers.ASExternalLSAtype:
							// 	externalLSA := lsa.Content.(layers.ASExternalLSA)
							// 	fmt.Printf("Version: %d\n", ospf.Version)
							// 	fmt.Printf("RouterID: %d\n", ospf.RouterID)
							// 	fmt.Printf("AreaID: %d\n", ospf.AreaID)
							// 	fmt.Printf("Type: %s\n", ospf.Type)
							// 	fmt.Printf("LSAType: ASExternalLSA\n")
							// 	fmt.Printf("LinkStateID: %x\n", lsa.LinkStateID)
							// 	fmt.Printf("  Metric: %d\n", externalLSA.Metric)
							// 	fmt.Printf("  Metric: %x\n", externalLSA.AddressPrefix)
							// 	fmt.Printf("  ForwardingAddress: %d\n", externalLSA.ForwardingAddress)
							// 	fmt.Printf("\n\n")
							// 	break
							default:
								break
							}
						}
						break
						// case layers.OSPFLinkStateAcknowledgment:
						// 	lsack := ospf.Content.([]layers.LSAheader)
						// 	fmt.Println(lsack)
						// 	break
					}
					break
				default:
					break
				}
			}
		}
	}()
	return nil
}
