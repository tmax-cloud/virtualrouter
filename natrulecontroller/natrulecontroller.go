package natrulecontroller

import (
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cho4036/virtualrouter/executor/iptables"
	v1 "github.com/cho4036/virtualrouter/pkg/apis/networkcontroller/v1"
	"github.com/vishvananda/netlink"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

const (
	mangleCountForwardChain      iptables.ChainName = "count_forward"
	mangleCountLanChain          iptables.ChainName = "count_lan"
	mangleCountWanChain          iptables.ChainName = "count_wan"
	manglePreroutingMarkLanChain iptables.ChainName = "prerouting_mark_lan"
	manglePreroutingMarkWanChain iptables.ChainName = "prerouting_mark_wan"

	natPreroutingLoadBalanceChain  iptables.ChainName = "prerouting_loadbalancing"
	natPostroutingLoadBalanceChain iptables.ChainName = "postrouting_loadbalancing"
	natPreroutingStaticNATChain    iptables.ChainName = "prerouting_staticnat"
	natPostroutingStaticNATChain   iptables.ChainName = "postrouting_staticnat"
	natPostroutingSNATChain        iptables.ChainName = "postrouting_snat"

	filterCountLanChain       iptables.ChainName = "count_lan"
	filterCountWanChain       iptables.ChainName = "count_wan"
	filterForwardFromLanChain iptables.ChainName = "forward_lan_to_wan"
	filterForwardFromWanChain iptables.ChainName = "forward_wan_to_lan"
)

type iptablesJumpChain struct {
	table     iptables.TableName
	dstChain  iptables.ChainName
	srcChain  iptables.ChainName
	comment   string
	extraArgs []string
}

var (
	// commitBytes              = []byte("COMMIT")
	iptablesMangleJumpChains = []iptablesJumpChain{
		{iptables.TableMangle, mangleCountForwardChain, iptables.ChainForward, "count forward packets", nil},
	}
	iptablesNatJumpChains = []iptablesJumpChain{
		{iptables.TableNAT, natPreroutingStaticNATChain, iptables.ChainPrerouting, "to staticnat chain", nil},
		{iptables.TableNAT, natPostroutingStaticNATChain, iptables.ChainPostrouting, "to staticnat chain", nil},
		{iptables.TableNAT, natPostroutingSNATChain, iptables.ChainPostrouting, "to snat chain", nil},
	}
	iptablesFilterJumpChains = []iptablesJumpChain{}
)

type NatRuleController struct {
	// natrulesChange *tracker.Tracker

	mu         sync.Mutex
	natruleMap map[string]v1.NATRule

	natruleSynced bool
	syncPeriod    time.Duration

	iptables iptables.Interface

	iptablesdata *bytes.Buffer
	filterChains *bytes.Buffer
	filterRules  *bytes.Buffer
	natChains    *bytes.Buffer
	natRules     *bytes.Buffer
	//ToDo: make this variable
	fwmark uint32
	intif  string
	extif  string
}

func New(intif string, extif string, fwmark uint32, iptables iptables.Interface, minSyncPeriod time.Duration) *NatRuleController {
	c := &NatRuleController{
		mu:            sync.Mutex{},
		natruleMap:    make(map[string]v1.NATRule),
		natruleSynced: false,
		syncPeriod:    minSyncPeriod,
		iptables:      iptables,
		iptablesdata:  bytes.NewBuffer(nil),
		filterChains:  bytes.NewBuffer(nil),
		filterRules:   bytes.NewBuffer(nil),
		natChains:     bytes.NewBuffer(nil),
		natRules:      bytes.NewBuffer(nil),
		fwmark:        fwmark,
		intif:         intif,
		extif:         extif,
	}

	if err := c.initialize(); err != nil {
		klog.Error("IPTables Initalize is failed")
	}

	return c
}

func (n *NatRuleController) initialize() error {
	initMangleJumpChains := []iptablesJumpChain{
		{iptables.TableMangle, mangleCountLanChain, iptables.ChainPrerouting, "count lan packets", []string{"-i", n.intif}},
		{iptables.TableMangle, mangleCountWanChain, iptables.ChainPrerouting, "count wan packets", []string{"-i", n.extif}},
		{iptables.TableMangle, manglePreroutingMarkLanChain, iptables.ChainPrerouting, "mark new conn for lan", []string{"-i", n.intif, "-m", "conntrack", "--ctstate", "NEW"}},
		{iptables.TableMangle, manglePreroutingMarkWanChain, iptables.ChainPrerouting, "mark new conn for wan", []string{"-i", n.extif, "-m", "conntrack", "--ctstate", "NEW"}},
	}
	iptablesMangleJumpChains = append(iptablesMangleJumpChains, initMangleJumpChains...)
	for _, jump := range iptablesMangleJumpChains {
		if _, err := n.iptables.EnsureChain(jump.table, jump.dstChain); err != nil {
			klog.ErrorS(err, "Failed to ensure chain exists", "table", jump.table, "chain", jump.dstChain)
			return err
		}
		args := append(jump.extraArgs,
			"-m", "comment", "--comment", jump.comment,
			"-j", string(jump.dstChain),
		)
		if _, err := n.iptables.EnsureRule(iptables.Prepend, jump.table, jump.srcChain, args...); err != nil {
			klog.ErrorS(err, "Failed to ensure chain jumps", "table", jump.table, "srcChain", jump.srcChain, "dstChain", jump.dstChain)
			return err
		}
	}

	initNatJumpChains := []iptablesJumpChain{
		{iptables.TableNAT, natPreroutingLoadBalanceChain, iptables.ChainPrerouting, "to load balancing chain", []string{"-m", "mark", "--mark", strconv.FormatUint(uint64(n.fwmark), 10)}},
		{iptables.TableNAT, natPostroutingLoadBalanceChain, iptables.ChainPostrouting, "to load balancing chain", []string{"-m", "mark", "--mark", strconv.FormatUint(uint64(n.fwmark), 10)}},
	}
	iptablesNatJumpChains = append(iptablesNatJumpChains, initNatJumpChains...)
	for _, jump := range iptablesNatJumpChains {
		if _, err := n.iptables.EnsureChain(jump.table, jump.dstChain); err != nil {
			klog.ErrorS(err, "Failed to ensure chain exists", "table", jump.table, "chain", jump.dstChain)
			return err
		}
		args := append(jump.extraArgs,
			"-m", "comment", "--comment", jump.comment,
			"-j", string(jump.dstChain),
		)
		if jump.dstChain == natPostroutingSNATChain {
			if _, err := n.iptables.EnsureRule(iptables.Append, jump.table, jump.srcChain, args...); err != nil {
				klog.ErrorS(err, "Failed to ensure chain jumps", "table", jump.table, "srcChain", jump.srcChain, "dstChain", jump.dstChain)
				return err
			}
		} else {
			if _, err := n.iptables.EnsureRule(iptables.Prepend, jump.table, jump.srcChain, args...); err != nil {
				klog.ErrorS(err, "Failed to ensure chain jumps", "table", jump.table, "srcChain", jump.srcChain, "dstChain", jump.dstChain)
				return err
			}
		}
	}

	initFilterJumpChains := []iptablesJumpChain{
		{iptables.TableFilter, filterCountLanChain, iptables.ChainForward, "count lan packets", []string{"-i", n.intif}},
		{iptables.TableFilter, filterCountWanChain, iptables.ChainForward, "count wan packets", []string{"-i", n.extif}},
		{iptables.TableFilter, filterForwardFromLanChain, iptables.ChainForward, "from lan to wan filter chain", []string{"-i", n.intif}},
		{iptables.TableFilter, filterForwardFromWanChain, iptables.ChainForward, "from wan to lan filter chain", []string{"-i", n.extif}},
	}
	iptablesFilterJumpChains = append(iptablesFilterJumpChains, initFilterJumpChains...)

	for _, jump := range initFilterJumpChains {
		if _, err := n.iptables.EnsureChain(jump.table, jump.dstChain); err != nil {
			klog.ErrorS(err, "Failed to ensure chain exists", "table", jump.table, "chain", jump.dstChain)
			return err
		}
		args := append(jump.extraArgs,
			"-m", "comment", "--comment", jump.comment,
			"-j", string(jump.dstChain),
		)
		if _, err := n.iptables.EnsureRule(iptables.Prepend, jump.table, jump.srcChain, args...); err != nil {
			klog.ErrorS(err, "Failed to ensure chain jumps", "table", jump.table, "srcChain", jump.srcChain, "dstChain", jump.dstChain)
			return err
		}
	}

	//ToDo: Convert using EnsureRule Function
	cmd := exec.Command("iptables", "-t", "filter", "-P", "FORWARD", "DROP")
	if err := cmd.Run(); err != nil {
		klog.ErrorS(err, "Failed to apply iptables rule", "command", "iptables -t filter -P DROP")
		return fmt.Errorf("failed to set policy of filter table to DROP")
	}

	if err := n.initMark(); err != nil {
		klog.ErrorS(err, "Failed to initMark")
		return err
	}

	return nil
}

func (n *NatRuleController) Run(stop <-chan struct{}) {
	wait.Until(n.syncLoop, n.syncPeriod, stop)
	<-stop
	klog.Info("natrulecontroller run() is done")
}

func (n *NatRuleController) initMark() error {
	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableMangle, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	// buf := new(bytes.Buffer)
	// buf.WriteString("-A" + " " + "PREROUTING")
	// iptables.NF_NAT_ADD(rule.Match, rule.Action, "test", buf)
	initRules := []string{
		"-A PREROUTING -i" + " " + n.intif + " " + "-m conntrack --ctstate RELATED,ESTABLISHED -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		"-A PREROUTING -i" + " " + n.extif + " " + "-m conntrack --ctstate RELATED,ESTABLISHED -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		"-A OUTPUT -m conntrack --ctstate RELATED,ESTABLISHED -j CONNMARK --restore-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		"-A count_forward -j RETURN",
		"-A count_lan -j RETURN",
		"-A count_wan -j RETURN",
		"-A prerouting_mark_lan -j MARK --set-xmark 0xc8/0xffffffff",
		"-A prerouting_mark_lan -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		"-A prerouting_mark_lan -j RETURN",
		"-A prerouting_mark_wan -j MARK --set-xmark 0xc8/0xffffffff",
		"-A prerouting_mark_wan -j CONNMARK --save-mark --nfmask 0xffffffff --ctmask 0xffffffff",
		"-A prerouting_mark_wan -j RETURN",
	}
	lines = append(lines, initRules...)

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore(iptables.TableMangle, n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}
	return nil
}

func (n *NatRuleController) OnAdd(natrule *v1.NATRule) error {
	n.mu.Lock()
	klog.Info("onAdd Called")
	defer n.mu.Unlock()

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := getNamespaceName(natrule)
	var chainName string
	oldRules, ok := n.natruleMap[key]
	if ok {
		//n.natruleSynced = false
		klog.Warningf("Duplicated key(%s) detected During OnAdd Event. Going to overwrite rule : %+v to %+v", key, oldRules, natrule)
		for _, rule := range oldRules.Spec.Rules {
			if rule.Action.SrcIP == "0.0.0.0" {
				chainName = string(natPostroutingSNATChain)
				n.removeRule(&rule, chainName, &lines)
			} else if rule.Action.SrcIP != "" {
				chainName = string(natPostroutingStaticNATChain)
				n.removeRule(&rule, chainName, &lines)
				if err := delRouteForProxyARP(rule.Action.SrcIP); err != nil {
					klog.ErrorS(err, "delRouteForProxyARP")
					return err
				}
			} else if rule.Action.DstIP != "" {
				chainName = string(natPreroutingStaticNATChain)
				n.removeRule(&rule, chainName, &lines)
			} else {
				klog.Errorln("Wrong Format of rules")
			}
		}
	}

	for _, rule := range natrule.Spec.Rules {
		if rule.Action.SrcIP == "0.0.0.0" {
			chainName = string(natPostroutingSNATChain)
			n.appendRule(&rule, chainName, &lines)
		} else if rule.Action.SrcIP != "" {
			chainName = string(natPostroutingStaticNATChain)
			n.appendRule(&rule, chainName, &lines)
			if err := setRouteForProxyARP(rule.Action.SrcIP); err != nil {
				klog.ErrorS(err, "delRouteForProxyARP")
				return err
			}
		} else if rule.Action.DstIP != "" {
			chainName = string(natPreroutingStaticNATChain)
			n.appendRule(&rule, chainName, &lines)
		} else {
			klog.Errorln("Wrong Format of rules")
		}
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	n.natruleMap[key] = *natrule
	return nil
}

func (n *NatRuleController) OnDelete(natrule *v1.NATRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("onDelete Called")

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := getNamespaceName(natrule)
	val, ok := n.natruleMap[key]
	if !ok {
		// n.natruleSynced = false
		klog.Warningf("Deleting empty value on key(%s) detected During OnDelete Event", key)
	}

	var chainName string
	for _, rule := range val.Spec.Rules {
		if rule.Action.SrcIP == "0.0.0.0" {
			chainName = string(natPostroutingSNATChain)
			n.removeRule(&rule, chainName, &lines)
		} else if rule.Action.SrcIP != "" {
			chainName = string(natPostroutingStaticNATChain)
			n.removeRule(&rule, chainName, &lines)
			if err := delRouteForProxyARP(rule.Action.SrcIP); err != nil {
				klog.ErrorS(err, "delRouteForProxyARP")
				return err
			}
		} else if rule.Action.DstIP != "" {
			chainName = string(natPreroutingStaticNATChain)
			n.removeRule(&rule, chainName, &lines)
		} else {
			klog.Errorln("Wrong Format of rules")
		}
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}
	delete(n.natruleMap, key)
	return nil
}

func (n *NatRuleController) OnUpdate(natrule *v1.NATRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("onUpdate Called")

	n.iptablesdata.Reset()

	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	key := getNamespaceName(natrule)
	// for k, v := range n.natruleMap {
	// 	klog.Infof("key: %s, value: %+v", k, v)
	// }

	var chainName string
	val, ok := n.natruleMap[key]
	if !ok {
		n.natruleSynced = false
		klog.Warningf("Updating empty value on key(%s) detected During OnUpdate Event", key)
	} else {
		for _, rule := range val.Spec.Rules {
			if rule.Action.SrcIP == "0.0.0.0" {
				chainName = string(natPostroutingSNATChain)
				n.removeRule(&rule, chainName, &lines)
			} else if rule.Action.SrcIP != "" {
				chainName = string(natPostroutingStaticNATChain)
				n.removeRule(&rule, chainName, &lines)
				if err := delRouteForProxyARP(rule.Action.SrcIP); err != nil {
					klog.ErrorS(err, "delRouteForProxyARP")
					return err
				}
			} else if rule.Action.DstIP != "" {
				chainName = string(natPreroutingStaticNATChain)
				n.removeRule(&rule, chainName, &lines)
			} else {
				klog.Errorln("Wrong Format of rules")
			}
		}
	}

	for _, rule := range natrule.Spec.Rules {
		if rule.Action.SrcIP == "0.0.0.0" {
			chainName = string(natPostroutingSNATChain)
			n.appendRule(&rule, chainName, &lines)
		} else if rule.Action.SrcIP != "" {
			chainName = string(natPostroutingStaticNATChain)
			n.appendRule(&rule, chainName, &lines)
			if err := setRouteForProxyARP(rule.Action.SrcIP); err != nil {
				klog.ErrorS(err, "delRouteForProxyARP")
				return err
			}
		} else if rule.Action.DstIP != "" {
			chainName = string(natPreroutingStaticNATChain)
			n.appendRule(&rule, chainName, &lines)
		} else {
			klog.Errorln("Wrong Format of rules")
		}
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)

	klog.Infof("Deploying rules : %s", n.iptablesdata.String())
	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
		return err
	}

	n.natruleMap[key] = *natrule
	return nil
}

func (n *NatRuleController) sync() {
	n.mu.Lock()
	defer n.mu.Unlock()

	// n.natRules.Reset()

	// for _, val := range n.natruleMap {
	// 	n.appendRule(&val)
	// }
	// n.natRules.WriteString("COMMIT")
	// n.natRules.WriteByte('\n')

	// if err := n.iptables.Restore("nat", n.natRules.Bytes(), true, true); err != nil {
	// 	klog.Errorf("Periodic Sync is failed : %+v", err)
	// }
}

func (n *NatRuleController) syncLoop() {
	for {
		n.sync()
	}
}

func getNamespaceName(natrule *v1.NATRule) string {
	return natrule.GetNamespace() + natrule.GetName()
}

func (n *NatRuleController) appendRule(rule *v1.Rules, chainName string, rules *[]string) {
	buf := new(bytes.Buffer)
	iptables.NF_NAT_ADD(rule.Match, rule.Action, chainName, buf)
	*rules = append(*rules, buf.String())
}

func (n *NatRuleController) removeRule(rule *v1.Rules, chainName string, rules *[]string) {
	buf := new(bytes.Buffer)
	iptables.NF_NAT_DEL(rule.Match, rule.Action, chainName, buf)
	*rules = append(*rules, buf.String())
}

func writeLine(buf *bytes.Buffer, words ...string) {
	// We avoid strings.Join for performance reasons.
	for i := range words {
		buf.WriteString(words[i])
		if i < len(words)-1 {
			buf.WriteByte('\n')
		} else {
			buf.WriteByte('\n')
		}
	}
}

func getRootNetlinkHandle() (*netlink.Handle, error) {
	handle, err := netlink.NewHandle()
	if err != nil {
		klog.ErrorS(err, "Error occured while geting RootNSHandle")
		return nil, err
	}
	return handle, nil
}

func setRouteForProxyARP(ip string) error {
	if !strings.Contains(ip, "/") {
		ip = ip + "/32"
	}
	var netlinkHandle *netlink.Handle
	// var newinterfaceName string

	if handle, err := getRootNetlinkHandle(); err != nil {
		klog.ErrorS(err, "GetTargetNetlinkHandle")
		return err
	} else {
		netlinkHandle = handle
	}

	var targetIP net.IP
	var targetIPNet *net.IPNet
	if ip, ipnet, err := net.ParseCIDR(ip); err != nil {
		klog.ErrorS(err, "Failed ParseCIDR", "IP", ip)
		return err
	} else {
		targetIP = ip
		targetIPNet = ipnet
	}

	var internalIntf netlink.Link
	if link, err := netlinkHandle.LinkByName("ethint"); err != nil {
		klog.ErrorS(err, "Failed LinkByName", "interfaceName", "ethint")
		return err
	} else {
		internalIntf = link
	}

	if ruleList, err := netlinkHandle.RouteList(internalIntf, 0); err != nil {
		klog.Error(err)
	} else {
		for _, v := range ruleList {
			if v.Dst.IP.Equal(targetIP) {
				return nil
			}
		}
	}

	if err := netlinkHandle.RouteAdd(&netlink.Route{
		Dst:       targetIPNet,
		LinkIndex: internalIntf.Attrs().Index,
	}); err != nil {
		klog.Error(err)
	}

	return nil
}

func delRouteForProxyARP(ip string) error {
	if !strings.Contains(ip, "/") {
		ip = ip + "/32"
	}
	var netlinkHandle *netlink.Handle
	// var newinterfaceName string

	if handle, err := getRootNetlinkHandle(); err != nil {
		klog.ErrorS(err, "GetTargetNetlinkHandle")
		return err
	} else {
		netlinkHandle = handle
	}

	var targetIP net.IP
	var targetIPNet *net.IPNet
	if ip, ipnet, err := net.ParseCIDR(ip); err != nil {
		klog.ErrorS(err, "Failed ParseCIDR", "IP", ip)
		return err
	} else {
		targetIP = ip
		targetIPNet = ipnet
	}

	var internalIntf netlink.Link
	if link, err := netlinkHandle.LinkByName("ethint"); err != nil {
		klog.ErrorS(err, "Failed LinkByName", "interfaceName", "ethint")
		return err
	} else {
		internalIntf = link
	}

	if ruleList, err := netlinkHandle.RouteList(internalIntf, 0); err != nil {
		klog.Error(err)
	} else {
		for _, v := range ruleList {
			if v.Dst.IP.Equal(targetIP) {
				return nil
			}
		}
	}

	if err := netlinkHandle.RouteDel(&netlink.Route{
		Dst:       targetIPNet,
		LinkIndex: internalIntf.Attrs().Index,
	}); err != nil {
		klog.Error(err)
	}

	return nil
}
