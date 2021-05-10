package iptables_test

import (
	"fmt"
	"testing"

	_ "github.com/cho4036/virtualrouter/pkg/util/iptables"
)

func TestMakeArgs(t *testing.T) {
	fullArgs := append([]string{"ddd"}, "")
	fmt.Println(fullArgs)
	fmt.Println(len(fullArgs))
}
