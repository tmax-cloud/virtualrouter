/*
Copyright 2014 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package iptables

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	utilexec "github.com/cho4036/virtualrouter/pkg/util/exec"
	"k8s.io/apimachinery/pkg/util/sets"
	utilversion "k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog/v2"
)

type RulePosition string

const (
	Prepend RulePosition = "-I"
	Append  RulePosition = "-A"
)

type Table string

const (
	TableNAT    Table = "nat"
	TableFilter Table = "filter"
	TableMangle Table = "mangle"
)

type Chain string

const (
	ChainPrerouting  Chain = "PREROUTING"
	ChainPostrouting Chain = "POSTROUTING"
	ChainInput       Chain = "INPUT"
	ChainForward     Chain = "FORWARD"
	ChainOUTPUT      Chain = "OUTPUT"
)

type IptablesCommand string

const (
	cmdIptablesSave   IptablesCommand = "iptables-save"
	cmdIptablesRestor IptablesCommand = "iptables-restore"
	cmdIptables       IptablesCommand = "iptables"

	cmdIp6tablesRestore IptablesCommand = "ip6tables-restore"
	cmdIp6tablesSave    IptablesCommand = "ip6tables-save"
	cmdIp6tables        IptablesCommand = "ip6tables"
)

type RestoreCountersFlag bool

const RestoreCounters RestoreCountersFlag = true
const NoRestoreCounters RestoreCountersFlag = false

type FlushFlag bool

const FlushTables FlushFlag = true
const NoFlushTables FlushFlag = false

// MinCheckVersion minimum version to be checked
// Versions of iptables less than this do not support the -C / --check flag
// (test whether a rule exists).
var MinCheckVersion = utilversion.MustParseGeneric("1.4.11")

// RandomFullyMinVersion is the minimum version from which the --random-fully flag is supported,
// used for port mapping to be fully randomized
var RandomFullyMinVersion = utilversion.MustParseGeneric("1.6.2")

// WaitMinVersion a minimum iptables versions supporting the -w and -w<seconds> flags
var WaitMinVersion = utilversion.MustParseGeneric("1.4.20")

// WaitIntervalMinVersion a minimum iptables versions supporting the wait interval useconds
var WaitIntervalMinVersion = utilversion.MustParseGeneric("1.6.1")

// WaitSecondsMinVersion a minimum iptables versions supporting the wait seconds
var WaitSecondsMinVersion = utilversion.MustParseGeneric("1.4.22")

// WaitRestoreMinVersion a minimum iptables versions supporting the wait restore seconds
var WaitRestoreMinVersion = utilversion.MustParseGeneric("1.6.2")

// WaitString a constant for specifying the wait flag
const WaitString = "-w"

// WaitSecondsValue a constant for specifying the default wait seconds
const WaitSecondsValue = "5"

// WaitIntervalString a constant for specifying the wait interval flag
const WaitIntervalString = "-W"

// WaitIntervalUsecondsValue a constant for specifying the default wait interval useconds
const WaitIntervalUsecondsValue = "100000"

const LockfilePath16x = "/run/xtables.lock"

type Protocol string

const (
	ProtocolIPv4 Protocol = "IPv4"

	ProtocolIPv6 Protocol = "IPv6"
)

type operation string

const (
	opCreateChain operation = "-N"
	opFlushChain  operation = "-F"
	opDeleteChain operation = "-X"
	opListChain   operation = "-S"
	opAppendRule  operation = "-A"
	opCheckRule   operation = "-C"
	opDeleteRule  operation = "-D"
)

type Interface interface {
	EnsureChian(table Table, chain Chain) (bool, error)
	FlushChain(table Table, chain Chain) error
	DeleteChain(table Table, chain Chain) error

	EnsureRule(position RulePosition, table Table, chain Chain, args ...string) (bool, error)
	DeleteRule(table Table, chain Chain, args ...string) error

	SaveInto(table Table, buffer *bytes.Buffer) error
	Restore(table Table, data []byte, flush FlushFlag, counters RestoreCountersFlag) error
	RestoreAll(data []byte, flush FlushFlag, counters RestoreCountersFlag) error
}

type runner struct {
	mu              sync.Mutex
	exec            utilexec.Interface
	hasCheck        bool
	waitFlag        []string
	restoreWaitFlag []string
	protocol        Protocol
	lockfilePath    string
}

func newIptablesRunner(exec utilexec.Interface, protocol Protocol, lockfilePath string) Interface {
	version, err := getIPTablesVersion(exec, protocol)
	if err != nil {
		klog.Warningf("Error checking iptables version, assuming version at least %s: %v", MinCheckVersion, err)
		version = MinCheckVersion
	}

	if lockfilePath == "" {
		lockfilePath = LockfilePath16x
	}

	runner := &runner{
		exec:            exec,
		protocol:        protocol,
		hasCheck:        version.AtLeast(MinCheckVersion),
		waitFlag:        getIptablesWaitFlag(version),
		restoreWaitFlag: getIptablesRestoreWaitFlag(version, exec, protocol),
		lockfilePath:    lockfilePath,
	}
	return runner
}

func New(exec utilexec.Interface, protocol Protocol) Interface {
	return newIptablesRunner(exec, protocol, "")
}

func (r *runner) EnsureChian(table Table, chain Chain) (bool, error) {
	fullArgs := makeFullArgs(table, chain)

	r.mu.Lock()
	defer r.mu.Unlock()

	out, err := r.run(opCreateChain, fullArgs)
	if err != nil {
		if ee, ok := err.(utilexec.ExitError); ok {
			if ee.Exited() && ee.ExitStatus() == 1 {
				return true, nil
			}
		}
		return false, fmt.Errorf("error creating chain %q: %v: %s", chain, err, out)
	}
	return false, nil
}

func (r *runner) FlushChain(table Table, chain Chain) error {
	fullArgs := makeFullArgs(table, chain)

	r.mu.Lock()
	defer r.mu.Unlock()

	out, err := r.run(opFlushChain, fullArgs)
	if err != nil {
		return fmt.Errorf("error flushing chain %q: %v :%s", chain, err, out)
	}
	return nil

}
func (r *runner) DeleteChain(table Table, chain Chain) error {
	fullArgs := makeFullArgs(table, chain)

	r.mu.Lock()
	defer r.mu.Unlock()

	out, err := r.run(opFlushChain, fullArgs)
	if err != nil {
		return fmt.Errorf("error deleting chain %q: %v :%s", chain, err, out)
	}
	return nil

}

func (r *runner) EnsureRule(position RulePosition, table Table, chain Chain, args ...string) (bool, error) {
	fullArgs := makeFullArgs(table, chain, args...)

	r.mu.Lock()
	defer r.mu.Unlock()

	exists, err := r.checkRule(table, chain, args...)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}

	// if rule doesn't exist already
	out, err := r.run(operation(position), fullArgs)
	if err != nil {
		return false, fmt.Errorf("error appending rule %v :%s", err, out)
	}
	return false, nil
}
func (r *runner) DeleteRule(table Table, chain Chain, args ...string) error {
	fullArgs := makeFullArgs(table, chain, args...)

	r.mu.Lock()
	defer r.mu.Unlock()

	exists, err := r.checkRule(table, chain, args...)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	out, err := r.run(opAppendRule, fullArgs)
	if err != nil {
		return fmt.Errorf("error deleting rule %v :%s", err, out)
	}
	return nil
}

func (r *runner) SaveInto(table Table, buffer *bytes.Buffer) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	iptablesSaveCmd := iptablesSaveCommand(r.protocol)
	args := makeFullArgs(table, "")
	klog.Info("running %s %v", iptablesSaveCmd, args)
	cmd := r.exec.Command(iptablesSaveCmd, args...)
	cmd.SetStdout(buffer)
	stderrBuffer := bytes.NewBuffer(nil)
	cmd.SetStderr(stderrBuffer)

	err := cmd.Run()
	if err != nil {
		stderrBuffer.WriteTo(buffer)
	}
	return err
}
func (r *runner) Restore(table Table, data []byte, flush FlushFlag, counters RestoreCountersFlag) error {
	args := []string{"-T", string(table)}
	return r.restoreInternal(args, data, flush, counters)
}
func (r *runner) RestoreAll(data []byte, flush FlushFlag, counters RestoreCountersFlag) error {
	args := make([]string, 0)
	return r.restoreInternal(args, data, flush, counters)
}

type iptablesLocker interface {
	Close() error
}

// restoreInternal is the shared part of Restore/RestoreAll
func (runner *runner) restoreInternal(args []string, data []byte, flush FlushFlag, counters RestoreCountersFlag) error {
	runner.mu.Lock()
	defer runner.mu.Unlock()

	if !flush {
		args = append(args, "--noflush")
	}
	if counters {
		args = append(args, "--counters")
	}

	// Grab the iptables lock to prevent iptables-restore and iptables
	// from stepping on each other.  iptables-restore 1.6.2 will have
	// a --wait option like iptables itself, but that's not widely deployed.
	if len(runner.restoreWaitFlag) == 0 {
		locker, err := grabIptablesLocks(runner.lockfilePath)
		if err != nil {
			return err
		}

		defer func(locker iptablesLocker) {
			if err := locker.Close(); err != nil {
				klog.Errorf("Failed to close iptables locks: %v", err)
			}
		}(locker)
	}

	// run the command and return the output or an error including the output and error
	fullArgs := append(runner.restoreWaitFlag, args...)
	iptablesRestoreCmd := iptablesRestoreCommand(runner.protocol)
	klog.V(4).Infof("running %s %v", iptablesRestoreCmd, fullArgs)
	cmd := runner.exec.Command(iptablesRestoreCmd, fullArgs...)
	cmd.SetStdin(bytes.NewBuffer(data))
	b, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%v (%s)", err, b)
	}
	return nil
}

func makeFullArgs(table Table, chain Chain, args ...string) []string {
	if chain == "" {
		return append([]string{"-t", string(table)}, args...)
	}
	return append([]string{string(chain), "-t", string(table)}, args...)
}

func iptablesSaveCommand(protocol Protocol) string {
	if protocol == ProtocolIPv6 {
		return string(cmdIp6tablesSave)
	}
	return string(cmdIptablesSave)
}

func iptablesRestoreCommand(protocol Protocol) string {
	if protocol == ProtocolIPv6 {
		return string(cmdIp6tablesRestore)
	}
	return string(cmdIptablesRestor)

}

func iptablesCommand(protocol Protocol) string {
	if protocol == ProtocolIPv6 {
		return string(cmdIp6tables)
	}
	return string(cmdIptables)
}

const iptablesVersionPattern = `v([0-9]+(\.[0-9]+)+)`

// getIPTablesVersion runs "iptables --version" and parses the returned version
func getIPTablesVersion(exec utilexec.Interface, protocol Protocol) (*utilversion.Version, error) {
	// this doesn't access mutable state so we don't need to use the interface / runner
	iptablesCmd := iptablesCommand(protocol)
	bytes, err := exec.Command(iptablesCmd, "--version").CombinedOutput()
	if err != nil {
		return nil, err
	}
	versionMatcher := regexp.MustCompile(iptablesVersionPattern)
	match := versionMatcher.FindStringSubmatch(string(bytes))
	if match == nil {
		return nil, fmt.Errorf("no iptables version found in string: %s", bytes)
	}
	version, err := utilversion.ParseGeneric(match[1])
	if err != nil {
		return nil, fmt.Errorf("iptables version %q is not a valid version string: %v", match[1], err)
	}

	return version, nil
}

func (r *runner) run(op operation, args []string) ([]byte, error) {
	return r.runContext(context.TODO(), op, args)
}

func (r *runner) runContext(ctx context.Context, op operation, args []string) ([]byte, error) {
	iptablesCmd := iptablesCommand(r.protocol)
	fullArgs := append(r.waitFlag, string(op))
	fullArgs = append(fullArgs, args...)
	klog.V(5).Infof("running iptables: %s %v", iptablesCmd, fullArgs)
	if ctx == nil {
		return r.exec.Command(iptablesCmd, fullArgs...).CombinedOutput()
	}
	return r.exec.CommandContext(ctx, iptablesCmd, fullArgs...).CombinedOutput()
	// Don't log err here - callers might not think it is an error.
}

func (r *runner) checkRule(table Table, chain Chain, args ...string) (bool, error) {
	if r.hasCheck {
		return r.checkRuleUsingCheck(makeFullArgs(table, chain, args...))
	}
	return r.checkRuleWithoutCheck(table, chain, args...)
}

func (r *runner) checkRuleUsingCheck(args []string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	out, err := r.runContext(ctx, opCheckRule, args)
	if ctx.Err() == context.DeadlineExceeded {
		return false, fmt.Errorf("timed out while checking rules")
	}
	if err == nil {
		return true, nil
	}
	if ee, ok := err.(utilexec.ExitError); ok {
		// iptables uses exit(1) to indicate a failure of the operation,
		// as compared to a malformed commandline, for example.
		if ee.Exited() && ee.ExitStatus() == 1 {
			return false, nil
		}
	}
	return false, fmt.Errorf("error checking rule: %v: %s", err, out)
}

// Executes the rule check without using the "-C" flag, instead parsing iptables-save.
// Present for compatibility with <1.4.11 versions of iptables.  This is full
// of hack and half-measures.  We should nix this ASAP.
func (r *runner) checkRuleWithoutCheck(table Table, chain Chain, args ...string) (bool, error) {
	iptablesSaveCmd := iptablesSaveCommand(r.protocol)
	klog.V(1).Infof("running %s -t %s", iptablesSaveCmd, string(table))
	out, err := r.exec.Command(iptablesSaveCmd, "-t", string(table)).CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("error checking rule: %v", err)
	}

	// Sadly, iptables has inconsistent quoting rules for comments. Just remove all quotes.
	// Also, quoted multi-word comments (which are counted as a single arg)
	// will be unpacked into multiple args,
	// in order to compare against iptables-save output (which will be split at whitespace boundary)
	// e.g. a single arg('"this must be before the NodePort rules"') will be unquoted and unpacked into 7 args.
	var argsCopy []string
	for i := range args {
		tmpField := strings.Trim(args[i], "\"")
		tmpField = trimhex(tmpField)
		argsCopy = append(argsCopy, strings.Fields(tmpField)...)
	}
	argset := sets.NewString(argsCopy...)

	for _, line := range strings.Split(string(out), "\n") {
		var fields = strings.Fields(line)

		// Check that this is a rule for the correct chain, and that it has
		// the correct number of argument (+2 for "-A <chain name>")
		if !strings.HasPrefix(line, fmt.Sprintf("-A %s", string(chain))) || len(fields) != len(argsCopy)+2 {
			continue
		}

		// Sadly, iptables has inconsistent quoting rules for comments.
		// Just remove all quotes.
		for i := range fields {
			fields[i] = strings.Trim(fields[i], "\"")
			fields[i] = trimhex(fields[i])
		}

		// TODO: This misses reorderings e.g. "-x foo ! -y bar" will match "! -x foo -y bar"
		if sets.NewString(fields...).IsSuperset(argset) {
			return true, nil
		}
		klog.V(5).Infof("DBG: fields is not a superset of args: fields=%v  args=%v", fields, args)
	}

	return false, nil
}

var hexnumRE = regexp.MustCompile("0x0+([0-9])")

func trimhex(s string) string {
	return hexnumRE.ReplaceAllString(s, "0x$1")
}

func getIptablesWaitFlag(version *utilversion.Version) []string {
	switch {
	case version.AtLeast(WaitIntervalMinVersion):
		return []string{WaitString, WaitSecondsValue, WaitIntervalString, WaitIntervalUsecondsValue}
	case version.AtLeast(WaitSecondsMinVersion):
		return []string{WaitString, WaitSecondsValue}
	case version.AtLeast(WaitMinVersion):
		return []string{WaitString}
	default:
		return nil
	}
}

// Checks if iptables-restore has a "wait" flag
func getIptablesRestoreWaitFlag(version *utilversion.Version, exec utilexec.Interface, protocol Protocol) []string {
	if version.AtLeast(WaitRestoreMinVersion) {
		return []string{WaitString, WaitSecondsValue, WaitIntervalString, WaitIntervalUsecondsValue}
	}

	// Older versions may have backported features; if iptables-restore supports
	// --version, assume it also supports --wait
	vstring, err := getIptablesRestoreVersionString(exec, protocol)
	if err != nil || vstring == "" {
		klog.V(3).Infof("couldn't get iptables-restore version; assuming it doesn't support --wait")
		return nil
	}
	if _, err := utilversion.ParseGeneric(vstring); err != nil {
		klog.V(3).Infof("couldn't parse iptables-restore version; assuming it doesn't support --wait")
		return nil
	}
	return []string{WaitString}
}

// getIPTablesRestoreVersionString runs "iptables-restore --version" to get the version string
// in the form "X.X.X"
func getIptablesRestoreVersionString(exec utilexec.Interface, protocol Protocol) (string, error) {
	// this doesn't access mutable state so we don't need to use the interface / runner

	// iptables-restore hasn't always had --version, and worse complains
	// about unrecognized commands but doesn't exit when it gets them.
	// Work around that by setting stdin to nothing so it exits immediately.
	iptablesRestoreCmd := iptablesRestoreCommand(protocol)
	cmd := exec.Command(iptablesRestoreCmd, "--version")
	cmd.SetStdin(bytes.NewReader([]byte{}))
	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	versionMatcher := regexp.MustCompile(iptablesVersionPattern)
	match := versionMatcher.FindStringSubmatch(string(bytes))
	if match == nil {
		return "", fmt.Errorf("no iptables version found in string: %s", bytes)
	}
	return match[1], nil
}
