package iptables

import (
	"bytes"
	"time"

	k8sIptables "k8s.io/kubernetes/pkg/util/iptables"
	k8sexec "k8s.io/utils/exec"
)

type TableName string
type ChainName string
type Policy string
type L4Protocol string
type L3Protocol string

// RestoreCountersFlag is an option flag for Restore
type RestoreCountersFlag bool

// RestoreCounters a boolean true constant for the option flag RestoreCountersFlag
const RestoreCounters RestoreCountersFlag = true

// NoRestoreCounters a boolean false constant for the option flag RestoreCountersFlag
const NoRestoreCounters RestoreCountersFlag = false

// FlushFlag an option flag for Flush
type FlushFlag bool

// FlushTables a boolean true constant for option flag FlushFlag
const FlushTables FlushFlag = true

// NoFlushTables a boolean false constant for option flag FlushFlag
const NoFlushTables FlushFlag = false

type RulePosition string

const (
	// Prepend is the insert flag for iptable
	Prepend RulePosition = "-I"
	// Append is the append flag for iptable
	Append RulePosition = "-A"
)

const (
	TableNAT    TableName = "nat"
	TableFilter TableName = "filter"
	TableMangle TableName = "mangle"
	TableRaw    TableName = "raw"

	ChainInput       ChainName = "INPUT"
	ChainOutput      ChainName = "OUTPUT"
	ChainForward     ChainName = "FORWARD"
	ChainPrerouting  ChainName = "PREROUTING"
	ChainPostrouting ChainName = "POSTROUTING"

	PolicyAccept Policy = "ACCEPT"
	PolicyReturn Policy = "RETURN"
	PolicyReject Policy = "REJECT"
	PolicyDrop   Policy = "DROP"

	ProtocolIPv4 L3Protocol = "IPv4"
	ProtocolIPv6 L3Protocol = "IPv6"

	ProtocolTCP L4Protocol = "tcp"
	ProtocolUDP L4Protocol = "udp"
)

type Interface interface {
	// EnsureChain checks if the specified chain exists and, if not, creates it.  If the chain existed, return true.
	EnsureChain(table TableName, chain ChainName) (bool, error)
	// FlushChain clears the specified chain.  If the chain did not exist, return error.
	FlushChain(table TableName, chain ChainName) error
	// DeleteChain deletes the specified chain.  If the chain did not exist, return error.
	DeleteChain(table TableName, chain ChainName) error
	// EnsureRule checks if the specified rule is present and, if not, creates it.  If the rule existed, return true.
	EnsureRule(position RulePosition, table TableName, chain ChainName, args ...string) (bool, error)
	// DeleteRule checks if the specified rule is present and, if so, deletes it.
	DeleteRule(table TableName, chain ChainName, args ...string) error
	// IsIPv6 returns true if this is managing ipv6 tables.
	IsIPv6() bool
	// Protocol returns the IP family this instance is managing,
	L3Protocol() L3Protocol
	// SaveInto calls `iptables-save` for table and stores result in a given buffer.
	SaveInto(table TableName, buffer *bytes.Buffer) error
	// Restore runs `iptables-restore` passing data through []byte.
	// table is the Table to restore
	// data should be formatted like the output of SaveInto()
	// flush sets the presence of the "--noflush" flag. see: FlushFlag
	// counters sets the "--counters" flag. see: RestoreCountersFlag
	Restore(table TableName, data []byte, flush FlushFlag, counters RestoreCountersFlag) error
	// RestoreAll is the same as Restore except that no table is specified.
	RestoreAll(data []byte, flush FlushFlag, counters RestoreCountersFlag) error
	// Monitor detects when the given iptables tables have been flushed by an external
	// tool (e.g. a firewall reload) by creating canary chains and polling to see if
	// they have been deleted. (Specifically, it polls tables[0] every interval until
	// the canary has been deleted from there, then waits a short additional time for
	// the canaries to be deleted from the remaining tables as well. You can optimize
	// the polling by listing a relatively empty table in tables[0]). When a flush is
	// detected, this calls the reloadFunc so the caller can reload their own iptables
	// rules. If it is unable to create the canary chains (either initially or after
	// a reload) it will log an error and stop monitoring.
	// (This function should be called from a goroutine.)
	Monitor(canary ChainName, tables []TableName, reloadFunc func(), interval time.Duration, stopCh <-chan struct{})
	// HasRandomFully reveals whether `-j MASQUERADE` takes the
	// `--random-fully` option.  This is helpful to work around a
	// Linux kernel bug that sometimes causes multiple flows to get
	// mapped to the same IP:PORT and consequently some suffer packet
	// drops.
	HasRandomFully() bool
}

type internalRunner struct {
	runner k8sIptables.Interface
}

type IPV4Interface interface {
	Interface
}

type IPV6Interface interface {
	Interface
}

func NewIPV4() IPV4Interface {
	return &internalRunner{
		runner: k8sIptables.New(k8sexec.New(), k8sIptables.ProtocolIPv4),
	}
}

func NewIPV6() IPV6Interface {
	return &internalRunner{
		runner: k8sIptables.New(k8sexec.New(), k8sIptables.ProtocolIPv6),
	}
}

func (r *internalRunner) EnsureChain(table TableName, chain ChainName) (bool, error) {
	return r.runner.EnsureChain(k8sIptables.Table(table), k8sIptables.Chain(chain))
}

func (r *internalRunner) FlushChain(table TableName, chain ChainName) error {
	return r.runner.FlushChain(k8sIptables.Table(table), k8sIptables.Chain(chain))
}

func (r *internalRunner) DeleteChain(table TableName, chain ChainName) error {
	return r.runner.DeleteChain(k8sIptables.Table(table), k8sIptables.Chain(chain))
}

func (r *internalRunner) EnsureRule(position RulePosition, table TableName, chain ChainName, args ...string) (bool, error) {
	return r.runner.EnsureRule(k8sIptables.RulePosition(position), k8sIptables.Table(table), k8sIptables.Chain(chain), args...)
}

func (r *internalRunner) DeleteRule(table TableName, chain ChainName, args ...string) error {
	return r.runner.DeleteRule(k8sIptables.Table(table), k8sIptables.Chain(chain), args...)
}

func (r *internalRunner) IsIPv6() bool {
	return r.runner.IsIPv6()
}

func (r *internalRunner) L3Protocol() L3Protocol {
	proto := r.runner.Protocol()
	if proto == k8sIptables.ProtocolIPv4 {
		return ProtocolIPv4
	} else {
		return ProtocolIPv6
	}
}

func (r *internalRunner) SaveInto(table TableName, buffer *bytes.Buffer) error {
	return r.runner.SaveInto(k8sIptables.Table(table), buffer)
}

func (r *internalRunner) Restore(table TableName, data []byte, flush FlushFlag, counters RestoreCountersFlag) error {
	return r.runner.Restore(k8sIptables.Table(table), data, k8sIptables.FlushFlag(flush), k8sIptables.RestoreCountersFlag(counters))
}

func (r *internalRunner) RestoreAll(data []byte, flush FlushFlag, counters RestoreCountersFlag) error {
	return r.runner.RestoreAll(data, k8sIptables.FlushFlag(flush), k8sIptables.RestoreCountersFlag(counters))
}

func (r *internalRunner) Monitor(canary ChainName, tables []TableName, reloadFunc func(), interval time.Duration, stopCh <-chan struct{}) {
	var k8sIptablesTables = make([]k8sIptables.Table, len(tables))
	for _, val := range tables {
		k8sIptablesTables = append(k8sIptablesTables, k8sIptables.Table(val))
	}
	r.runner.Monitor(k8sIptables.Chain(canary), k8sIptablesTables, reloadFunc, interval, stopCh)
}

func (r *internalRunner) HasRandomFully() bool {
	return r.runner.HasRandomFully()
}
