package iptables

import (
	"bytes"

	v1 "github.com/cho4036/virtualrouter/pkg/apis/networkcontroller/v1"
)

func NF_NAT_ADD(m v1.Match, a v1.Action, chain string, buffer *bytes.Buffer) {
	if a.DstIP != "" {
		buffer.WriteString("-A" + " " + chain)
		match2string(m, buffer)
		dnatAction2string(a, buffer)
	}
	if a.SrcIP != "" {
		buffer.WriteString("-A" + " " + chain)
		match2string(m, buffer)
		snatAction2string(a, buffer)
	}
	// writeLine(buffer, "COMMIT")
}

func NF_NAT_DEL(m v1.Match, a v1.Action, chain string, buffer *bytes.Buffer) {
	if a.DstIP != "" {
		buffer.WriteString("-D" + " " + chain)
		match2string(m, buffer)
		dnatAction2string(a, buffer)
	}
	if a.SrcIP != "" {
		buffer.WriteString("-D" + " " + chain)
		match2string(m, buffer)
		snatAction2string(a, buffer)
	}
	// writeLine(buffer, "COMMIT")
}

func match2string(m v1.Match, buffer *bytes.Buffer) {
	if m.DstIP != "" {
		buffer.WriteString(" -d " + m.DstIP)
	}
	if m.SrcIP != "" {
		buffer.WriteString(" -s " + m.SrcIP)
	}
	if m.Protocol != "" {
		buffer.WriteString(" -p " + m.Protocol)
	}
}

func dnatAction2string(a v1.Action, buffer *bytes.Buffer) {
	if a.DstIP == "0.0.0.0" || a.DstIP == "0.0.0.0/0" {
		writeLine(buffer, " -j", "MASQUERADE", "--random-fully")
		return
	}
	writeLine(buffer, " -j", "DNAT", "--to-destination", a.DstIP)
}

func snatAction2string(a v1.Action, buffer *bytes.Buffer) {
	writeLine(buffer, " -j", "SNAT", "--to-destination", a.SrcIP)
}

func writeLine(buf *bytes.Buffer, words ...string) {
	// We avoid strings.Join for performance reasons.
	for i := range words {
		buf.WriteString(words[i])
		if i < len(words)-1 {
			buf.WriteByte(' ')
		} else {
			buf.WriteByte('\n')
		}
	}
}
