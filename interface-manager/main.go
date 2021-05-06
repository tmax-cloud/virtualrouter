package main

import (
	"context"
	"fmt"
	"net"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

type dockerType string

func main() {

	var err error

	rootNetlinkHandle := getRootNetlinkHandle()
	target, err := getDockerContainerIDByAncestor("nginx")

	TargetNetlinkHandle := getTargetNetlinkHandle(getNsHandle(dockerType(target)))
	_ = TargetNetlinkHandle
	//checking bridge
	var exist bool
	la := netlink.NewLinkAttrs()
	la.Name = "br0"

	_, err = rootNetlinkHandle.LinkByName(la.Name)
	if err != nil {
		if err.Error() == "Link not found" {
			fmt.Println("notfound")
			exist = false
		} else {
			fmt.Errorf("Error to getting Links")
		}
	} else {
		exist = true
	}

	mybridge := &netlink.Bridge{LinkAttrs: la, VlanFiltering: &[]bool{true}[0]}
	//// making bridge
	if !exist {
		err = rootNetlinkHandle.LinkAdd(mybridge)
		if err != nil {
			fmt.Printf("could not add %s: %v\n", la.Name, err)
		}
	}

	la2 := netlink.NewLinkAttrs()
	la2.Name = "externalBridge"

	exist = false
	_, err = rootNetlinkHandle.LinkByName(la2.Name)
	if err != nil {
		if err.Error() == "Link not found" {
			fmt.Println("notfound")
			exist = false
		} else {
			fmt.Errorf("Error to getting Links")
		}
	} else {
		exist = true
	}

	mybridge2 := &netlink.Bridge{LinkAttrs: la2, VlanFiltering: &[]bool{true}[0]}
	//// making bridge
	if !exist {
		err = rootNetlinkHandle.LinkAdd(mybridge2)
		if err != nil {
			fmt.Printf("could not add %s: %v\n", la2.Name, err)
		}
	}
	rootNetlinkHandle.LinkSetUp(mybridge2)
	if err != nil {
		fmt.Println(err)
	}

	// checking eno1

	extInterface, err := rootNetlinkHandle.LinkByName("eno1")
	if err != nil {
		if err.Error() == "Link not found" {
			exist = false
		} else {
			fmt.Errorf("Error to getting Links")
		}
	} else {
		exist = true
	}

	// make master eno1 to bridge

	if exist {
		err = rootNetlinkHandle.LinkSetMaster(extInterface, mybridge)
		if err != nil {
			fmt.Printf("could not add %s: %v\n", la.Name, err)
		}
	}

	err = rootNetlinkHandle.LinkSetUp(extInterface)
	if err != nil {
		fmt.Println(err)
	}
	rootNetlinkHandle.LinkSetUp(mybridge)
	if err != nil {
		fmt.Println(err)
	}

	// checking enx00e04c321315
	exist = false
	extInterface2, err := rootNetlinkHandle.LinkByName("enx00e04c321315")
	if err != nil {
		if err.Error() == "Link not found" {
			exist = false
		} else {
			fmt.Errorf("Error to getting Links")
		}
	} else {
		exist = true
	}

	// make master enx00e04c321315 to bridge

	if exist {
		err = rootNetlinkHandle.LinkSetMaster(extInterface2, mybridge2)
		if err != nil {
			fmt.Printf("could not add %s: %v\n", la.Name, err)
		}
	}

	err = rootNetlinkHandle.LinkSetUp(extInterface2)
	if err != nil {
		fmt.Println(err)
	}

	addr2, err := netlink.ParseAddr("192.168.7.191/24")
	fmt.Println(addr2)

	err = rootNetlinkHandle.AddrDel(extInterface2, addr2)

	exist = false
	// make host EXT veth
	veth3, err := rootNetlinkHandle.LinkByName("ExtVeth0")
	if err != nil {
		if err.Error() == "Link not found" {
			exist = false
		} else {
			fmt.Errorf("Error to getting Links")
		}
	} else {
		exist = true
		rootNetlinkHandle.LinkDel(veth3)
	}

	vethAttr := netlink.NewLinkAttrs()
	vethAttr.Name = "ExtVeth0"
	veth3 = &netlink.Veth{
		LinkAttrs: vethAttr,
		PeerName:  "ExtVeth1",
	}
	err = rootNetlinkHandle.LinkAdd(veth3)
	if err != nil {
		fmt.Println(err)
	}
	peer3, err := rootNetlinkHandle.LinkByName("ExtVeth1")
	rootNetlinkHandle.LinkSetUp(veth3)
	rootNetlinkHandle.LinkSetUp(peer3)
	rootNetlinkHandle.LinkSetMaster(veth3, mybridge2)
	addr, err := netlink.ParseAddr("192.168.7.191/24")
	rootNetlinkHandle.AddrAdd(peer3, addr)
	addr, err = netlink.ParseAddr("0.0.0.0/0")
	dst := &net.IPNet{
		IP:   net.IPv4(0, 0, 0, 0),
		Mask: net.CIDRMask(0, 32),
	}
	gw := net.IPv4(192, 168, 7, 1)
	//route := netlink.Route{LinkIndex: peer3.Attrs().Index, Dst: dst, Gw: gw}
	route := netlink.Route{Dst: dst, Gw: gw}
	rootNetlinkHandle.RouteAdd(&route)

	// make host veth
	veth, err := rootNetlinkHandle.LinkByName("hostVeth0")
	if err != nil {
		if err.Error() == "Link not found" {
			exist = false
		} else {
			fmt.Errorf("Error to getting Links")
		}
	} else {
		exist = true
		rootNetlinkHandle.LinkDel(veth)
	}

	vethAttr = netlink.NewLinkAttrs()
	vethAttr.Name = "hostVeth0"
	veth = &netlink.Veth{
		LinkAttrs: vethAttr,
		PeerName:  "hostVeth1",
	}
	err = rootNetlinkHandle.LinkAdd(veth)
	if err != nil {
		fmt.Println(err)
	}
	peer, err := rootNetlinkHandle.LinkByName("hostVeth1")
	rootNetlinkHandle.LinkSetUp(veth)
	rootNetlinkHandle.LinkSetUp(peer)
	rootNetlinkHandle.LinkSetMaster(veth, mybridge)
	addr, err = netlink.ParseAddr("10.0.0.1/24")
	rootNetlinkHandle.AddrAdd(peer, addr)

	/// make peer interface

	// checking peer interface name
	veth, err = rootNetlinkHandle.LinkByName("testVeth0")
	if err != nil {
		if err.Error() == "Link not found" {
			exist = false
		} else {
			fmt.Errorf("Error to getting Links")
		}
	} else {
		exist = true
		rootNetlinkHandle.LinkDel(veth)
	}

	vethAttr = netlink.NewLinkAttrs()
	vethAttr.Name = "testVeth0"
	veth = &netlink.Veth{
		LinkAttrs: vethAttr,
		PeerName:  "testVeth1",
	}
	err = rootNetlinkHandle.LinkAdd(veth)
	if err != nil {
		fmt.Println(err)
	}
	rootNetlinkHandle.LinkSetUp(veth)

	peerVeth, err := rootNetlinkHandle.LinkByName("testVeth1")
	err = rootNetlinkHandle.LinkSetNsFd(peerVeth, int(getNsHandle(dockerType(target))))
	if err != nil {
		fmt.Println(err)
	}

	/// make master for testVeth0

	err = rootNetlinkHandle.LinkSetMaster(veth, mybridge)
	if err != nil {
		fmt.Printf("could not add %s: %v\n", la.Name, err)
	}

	addr, err = netlink.ParseAddr("10.0.0.10/24")
	fmt.Println(addr)

	err = TargetNetlinkHandle.AddrAdd(peerVeth, addr)
	TargetNetlinkHandle.LinkSetUp(peerVeth)

	rootNetlinkHandle.BridgeVlanAdd(extInterface, 2, true, false, false, true)
	rootNetlinkHandle.BridgeVlanAdd(veth, 2, true, true, false, true)

	exist = false
	// checking peer interface name
	veth2, err := rootNetlinkHandle.LinkByName("extVeth0")
	if err != nil {
		if err.Error() == "Link not found" {
			exist = false
		} else {
			fmt.Errorf("Error to getting Links")
		}
	} else {
		exist = true
		rootNetlinkHandle.LinkDel(veth2)
	}

	vethAttr = netlink.NewLinkAttrs()
	vethAttr.Name = "extVeth0"
	veth2 = &netlink.Veth{
		LinkAttrs: vethAttr,
		PeerName:  "extVeth1",
	}
	err = rootNetlinkHandle.LinkAdd(veth2)
	if err != nil {
		fmt.Println(err)
	}
	rootNetlinkHandle.LinkSetUp(veth2)

	peerVeth2, err := rootNetlinkHandle.LinkByName("extVeth1")
	err = rootNetlinkHandle.LinkSetNsFd(peerVeth2, int(getNsHandle(dockerType(target))))
	if err != nil {
		fmt.Println(err)
	}

	/// make master for testVeth0

	err = rootNetlinkHandle.LinkSetMaster(veth2, mybridge2)
	if err != nil {
		fmt.Printf("could not add %s: %v\n", la2.Name, err)
	}

	addr, err = netlink.ParseAddr("192.168.7.192/24")
	fmt.Println(addr)

	err = TargetNetlinkHandle.AddrAdd(peerVeth2, addr)
	TargetNetlinkHandle.LinkSetUp(peerVeth2)

	dst = &net.IPNet{
		IP:   net.IPv4(0, 0, 0, 0),
		Mask: net.CIDRMask(0, 32),
	}
	gw = net.IPv4(192, 168, 7, 1)
	//route := netlink.Route{LinkIndex: peer3.Attrs().Index, Dst: dst, Gw: gw}
	route = netlink.Route{Dst: dst, Gw: gw}
	TargetNetlinkHandle.RouteReplace(&route)

	/*
			&netlink.Addr{
			&net.IPNet{
				IP: net.ParseIP("10.0.0.10"),
				Mask: net.IPv4Mask(255,255,255,0),
			},
		})
	*/

	/*
		mybridge := &netlink.Bridge{LinkAttrs: la}

			err = rootNetlinkHandle.LinkAdd(mybridge)
			if err != nil {
				fmt.Printf("could not add %s: %v\n", la.Name, err)
			}

			///

			vethAttr := netlink.NewLinkAttrs()
			vethAttr.Name = "testVeth0"
			veth := &netlink.Veth{
				LinkAttrs: vethAttr,
				PeerName:  "testVeth1",
			}
			err = rootNetlinkHandle.LinkAdd(veth)
			rootNetlinkHandle.LinkAdd()
			// netlink handle

			/*
				_ = newNS
				list, err := newNS.LinkList()
				for _, val := range list {
					fmt.Println(val)
					t, err := newNS.AddrList(val, netlink.FAMILY_V4)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println(t)
				}
	*/
	/// add bridge
	/*
		la := netlink.NewLinkAttrs()
		la.Name = "test"
		mybridge := &netlink.Bridge{LinkAttrs: la}
		err = rootNetlinkHandle.LinkAdd(mybridge)
		if err != nil {
			fmt.Printf("could not add %s: %v\n", la.Name, err)
		}
	*/
	/*
		if err != nil {
			fmt.Printf("could not add : %v\n", err)
		}

		//a, err := netlink.VethPeerIndex(veth)
		peerVeth, err := rootNetlinkHandle.LinkByName("testVeth1")
		err = rootNetlinkHandle.LinkSetNsFd(peerVeth, int(nsHandler))
		fmt.Println(int(nsHandler))
		if err != nil {
			fmt.Println(err)
		}
	*/
	// err = rootNS.LinkSetNsFd(vethToNewNS, containerNsId)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// 	la := netlink.NewLinkAttrs()
	// 	la.Name = "test"
	// 	mybridge := &netlink.Bridge{LinkAttrs: la}
	// 	err := netlink.LinkAdd(mybridge)
	// 	if err != nil {
	// 		fmt.Printf("could not add %s: %v\n", la.Name, err)
	// 	}
	// 	eth1, _ := netlink.LinkByName("eno1")
	// 	netlink.LinkSetMaster(eth1, mybridge)
}

func getRootNetlinkHandle() *netlink.Handle {
	handle, err := netlink.NewHandle()
	if err != nil {
		fmt.Errorf("Error occured while geting RootNSHandle %v\n", err)
	}
	return handle
}

func getTargetNetlinkHandle(ns netns.NsHandle) *netlink.Handle {
	if ns == 0 {
		fmt.Errorf("Getting wrong ns number(0)")
		return nil
	}
	handle, err := netlink.NewHandleAt(ns)
	if err != nil {
		fmt.Errorf("Error occured while geting getTargetNSHandle %v\n", err)
	}
	return handle
}

func getNsHandle(arg interface{}) netns.NsHandle {
	switch arg.(type) {

	case dockerType:
		handle, err := netns.GetFromDocker(string(arg.(dockerType)))
		if err != nil {
			fmt.Errorf("%v\n", err)
			return 0
		}
		return handle
	case string:
		if arg.(string) == "root" {
			handle, err := netns.Get()
			if err != nil {
				fmt.Errorf("%v\n", err)
				return 0
			}
			return handle
		} else {
			handle, err := netns.GetFromName(arg.(string))
			if err != nil {
				fmt.Errorf("%v\n", err)
				return 0
			}
			return handle
		}
	}
	fmt.Errorf("Wrong argument")
	return 0
}

func getDockerContainerIDByAncestor(name string) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}

	args := filters.NewArgs()
	args.Add("ancestor", name)
	containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{
		Filters: args,
	})
	if err != nil {
		return "", err
	}

	var containerID string
	for _, container := range containers {
		containerID = container.ID
		fmt.Printf("%s %s\n", container.ID[:10], container.Image)
	}
	return containerID, nil

}

func getDockerContinerPidByID(id string) (int, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return 0, err
	}
	insp, err := cli.ContainerInspect(context.Background(), id)
	if err != nil {
		return 0, err
	}

	return insp.State.Pid, nil
}

/*
func setBridge(name string) error {
	la := netlink.NewLinkAttrs()
	la.Name = "test"
	mybridge := &netlink.Bridge{LinkAttrs: la}
	err = rootNetlinkHandle.LinkAdd(mybridge)
	if err != nil {
		fmt.Printf("could not add %s: %v\n", la.Name, err)
	}
}
*/
