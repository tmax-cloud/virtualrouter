package natrulecontroller

import (
	"bytes"
	"strings"
	"sync"
	"time"

	"github.com/cho4036/virtualrouter/executor/iptables"
	v1 "github.com/cho4036/virtualrouter/pkg/apis/networkcontroller/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
)

var (
	commitBytes = []byte("COMMIT")
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
}

func New(iptables iptables.Interface, minSyncPeriod time.Duration) *NatRuleController {
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
	}

	if err := c.initialize(); err != nil {
		klog.Error("IPTables Initalize is failed")
	}

	return c
}

func (n *NatRuleController) initialize() error {
	n.iptables.EnsureChain("nat", "test")
	return nil
}

func (n *NatRuleController) Run(stop <-chan struct{}) {
	wait.Until(n.syncLoop, n.syncPeriod, stop)
	<-stop
	klog.Info("natrulecontroller run() is done")
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
	oldRules, ok := n.natruleMap[key]
	if ok {
		//n.natruleSynced = false
		klog.Warningf("Duplicated key(%s) detected During OnAdd Event. Going to overwrite rule : %+v to %+v", key, oldRules, natrule)
		n.removeRule(&oldRules, &lines)
	}

	n.appendRule(natrule, &lines)

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
	_, ok := n.natruleMap[key]
	if !ok {
		// n.natruleSynced = false
		klog.Warningf("Deleting empty value on key(%s) detected During OnDelete Event", key)
	}

	n.removeRule(natrule, &lines)
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
	for k, v := range n.natruleMap {
		klog.Infof("key: %s, value: %+v", k, v)
	}

	val, ok := n.natruleMap[key]
	if !ok {
		n.natruleSynced = false
		klog.Warningf("Updating empty value on key(%s) detected During OnUpdate Event", key)
	} else {
		n.removeRule(&val, &lines)
	}

	n.appendRule(natrule, &lines)

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

func (n *NatRuleController) appendRule(natrule *v1.NATRule, rules *[]string) {
	buf := new(bytes.Buffer)
	for _, rule := range natrule.Spec.Rules {
		iptables.NF_NAT_ADD(rule.Match, rule.Action, "test", buf)
		*rules = append(*rules, buf.String())
	}
}

func (n *NatRuleController) removeRule(natrule *v1.NATRule, rules *[]string) {
	buf := new(bytes.Buffer)
	for _, rule := range natrule.Spec.Rules {
		iptables.NF_NAT_DEL(rule.Match, rule.Action, "test", buf)
		*rules = append(*rules, buf.String())
	}
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
