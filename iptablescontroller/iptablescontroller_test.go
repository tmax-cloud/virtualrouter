package iptablescontroller

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/tmax-cloud/virtualrouter/executor/iptables"
)

func TestIPtablesTest(t *testing.T) {
	c := iptables.NewIPV4()
	var b *bytes.Buffer
	b = &bytes.Buffer{}
	c.SaveInto(iptables.TableNAT, b)
	fmt.Println(b)
}
