package main

import (
	"fmt"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

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
								fmt.Printf("LSAType: RouterLSA\n")
								for _, router := range routerLSA.Routers {
									fmt.Printf("\n")
									fmt.Printf("  Type: %d\n", router.Type)
									fmt.Printf("  LinkID: %x\n", router.LinkID)
									fmt.Printf("  LinkData: %x\n", router.LinkData)
									fmt.Printf("  Metric: %d\n", router.Metric)
								}
								fmt.Printf("\n\n")
								break
							case layers.ASExternalLSAtypeV2:
								externalLSA := lsa.Content.(layers.ASExternalLSAV2)
								fmt.Printf("Version: %d\n", ospf.Version)
								fmt.Printf("RouterID: %d\n", ospf.RouterID)
								fmt.Printf("AreaID: %d\n", ospf.AreaID)
								fmt.Printf("Type: %s\n", ospf.Type)
								fmt.Printf("LSAType: ASExternalLSA\n")
								fmt.Printf("LinkStateID: %x\n", lsa.LinkStateID)
								fmt.Printf("  Metric: %d\n", externalLSA.Metric)
								fmt.Printf("  Mask: %x\n", externalLSA.NetworkMask)
								fmt.Printf("  ForwardingAddress: %d\n", externalLSA.ForwardingAddress)
								fmt.Printf("  ExternalRouteTag: %d\n", externalLSA.ExternalRouteTag)
								fmt.Printf("\n\n")
								break
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
								}
								fmt.Printf("\n\n")
								break
							case layers.ASExternalLSAtype:
								externalLSA := lsa.Content.(layers.ASExternalLSA)
								fmt.Printf("Version: %d\n", ospf.Version)
								fmt.Printf("RouterID: %d\n", ospf.RouterID)
								fmt.Printf("AreaID: %d\n", ospf.AreaID)
								fmt.Printf("Type: %s\n", ospf.Type)
								fmt.Printf("LSAType: ASExternalLSA\n")
								fmt.Printf("LinkStateID: %x\n", lsa.LinkStateID)
								fmt.Printf("  Metric: %d\n", externalLSA.Metric)
								fmt.Printf("  Metric: %x\n", externalLSA.AddressPrefix)
								fmt.Printf("  ForwardingAddress: %d\n", externalLSA.ForwardingAddress)
								// fmt.Printf("  ExternalRouteTag: %d\n", externalLSA.ExternalRouteTag)
								fmt.Printf("\n\n")
								break
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
