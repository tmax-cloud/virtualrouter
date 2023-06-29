package iptables

import (
	"bytes"
	"strconv"

	v1 "github.com/tmax-cloud/virtualrouter/pkg/apis/networkcontroller/v1"
)

func NF_ADD(m v1.Match, a v1.Action, chain string, buffer *bytes.Buffer, args ...string) {
	if chain != "" {
		buffer.WriteString("-A" + " " + chain)
	}
	if a.Policy != "" {
		// buffer.WriteString("-A" + " " + chain)
		if args != nil {
			args2string(args, buffer)
		}
		match2string(m, buffer)
		policyAction2string(a, buffer)
		return
	}
	if a.DstIP != "" {
		// buffer.WriteString("-A" + " " + chain)
		if args != nil {
			args2string(args, buffer)
		}
		match2string(m, buffer)
		dnatAction2string(a, buffer)
	}
	if a.SrcIP != "" {
		// buffer.WriteString("-A" + " " + chain)
		if args != nil {
			args2string(args, buffer)
		}
		match2string(m, buffer)
		snatAction2string(a, buffer)
	}
	// writeLine(buffer, "COMMIT")
}

func NF_DEL(m v1.Match, a v1.Action, chain string, buffer *bytes.Buffer, args ...string) {
	if a.Policy != "" {
		buffer.WriteString("-D" + " " + chain)
		if args != nil {
			args2string(args, buffer)
		}
		match2string(m, buffer)
		policyAction2string(a, buffer)
		return
	}

	if a.DstIP != "" {
		// buffer.WriteString("-D" + " " + chain)
		buffer.WriteString("-D" + " " + chain)
		if args != nil {
			args2string(args, buffer)
		}
		match2string(m, buffer)
		dnatAction2string(a, buffer)
	}
	if a.SrcIP != "" {
		// buffer.WriteString("-D" + " " + chain)
		buffer.WriteString("-D" + " " + chain)
		if args != nil {
			args2string(args, buffer)
		}
		match2string(m, buffer)
		snatAction2string(a, buffer)
	}
	// writeLine(buffer, "COMMIT")
}

func args2string(args []string, buffer *bytes.Buffer) {
	for i := range args {
		buffer.WriteString(" " + args[i])
	}
}

func match2string(m v1.Match, buffer *bytes.Buffer) {
	if m.SrcIP != "" {
		buffer.WriteString(" -s " + m.SrcIP)
	}
	if m.DstIP != "" {
		buffer.WriteString(" -d " + m.DstIP)
	}
	if m.Protocol == "all" {
		// Do not specify the protocol
	} else if m.Protocol != "" {
		buffer.WriteString(" -p " + m.Protocol)
	}
	if m.DstPort != 0 {
		buffer.WriteString(" --dport " + strconv.Itoa(m.DstPort))
	}
	if m.SrcPort != 0 {
		buffer.WriteString(" --sport " + strconv.Itoa(m.SrcPort))
	}
}

func policyAction2string(a v1.Action, buffer *bytes.Buffer) {
	writeLine(buffer, " -j", a.Policy)
}

func dnatAction2string(a v1.Action, buffer *bytes.Buffer) {
	var destination string = a.DstIP
	if a.DstPort != 0 {
		destination += ":" + strconv.Itoa(a.DstPort)
	}
	writeLine(buffer, " -j", "DNAT", "--to-destination", destination)
}

func snatAction2string(a v1.Action, buffer *bytes.Buffer) {
	if a.SrcIP == "0.0.0.0" || a.SrcIP == "0.0.0.0/0" {
		writeLine(buffer, " -j", "MASQUERADE", "--random-fully")
		return
	}
	var source string = a.SrcIP
	if a.SrcPort != 0 {
		source += ":" + strconv.Itoa(a.SrcPort)
	}
	writeLine(buffer, " -j", "SNAT", "--to-source", source)
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
