package natrule

type NatRule struct {
	Match  Match
	Action Action
}

type Match struct {
	SrcIP   string `json:"srcIP"`
	DstIP   string `json:"dstIP"`
	Protocl string `json:"protocol"`
}

type Action struct {
	SrcIP string `json:"srcIP"`
	DstIP string `json:"dstIP"`
}
