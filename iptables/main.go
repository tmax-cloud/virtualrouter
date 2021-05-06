package main

import (
	"fmt"

	"github.com/coreos/go-iptables/iptables"
)

func main() {
	ipt, err := iptables.New()

	if err != nil {
		fmt.Printf("list chains of initial faied %v", err)
	}

	list, err := ipt.List("nat", "POSTROUTING")

	for _, val := range list {
		fmt.Println(val)
	}

}
