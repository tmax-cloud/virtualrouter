package main

import "github.com/vishvananda/netlink"

type handler interface {
	getBridge()
	setBridge()
	setLink(_type string, name string)
}

type Handler struct {
	netlink.Handle
}
