package iptablescontroller

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/tmax-cloud/virtualrouter/executor/iptables"
	healthchecker "github.com/tmax-cloud/virtualrouter/healthchecker"
	v1 "github.com/tmax-cloud/virtualrouter/pkg/apis/networkcontroller/v1"
	"k8s.io/klog/v2"
)

var (
	// key: loadbalanverIP:loadbalancerPort
	registeredLBRule map[string]v1.LBRules

	// key: CR's name,
	lbRuleSetMap map[string]*lbRuleSet
	// key: loadbalanverIP:loadbalancerPort
	lbRuleMap map[string]*lbRule
	// key: endpointIP:endpointPort
	endpointMap map[string]*endpoint
)

func init() {
	lbRuleSetMap = make(map[string]*lbRuleSet)
	lbRuleMap = make(map[string]*lbRule)
	endpointMap = make(map[string]*endpoint)
	// cleanUpRule = make(map[string]bool)
	// healthcheckerList = make(map[string]*healthCheckerInstance)
}

type lbRuleSet struct {
	key        string
	lbRuleKeys []string
}

type lbRule struct {
	mu                 sync.Mutex
	key                string
	lbRuleSetKey       string
	lbVirtualIP        string
	lbVirtualPort      int
	protocol           string
	endpointKeys       endpointKeysList
	healthCheckManager *healthCheckerInstance
	ctx                context.Context
	cancle             context.CancelFunc
	endpointAdded      chan string
	endpointDeleted    chan string
}
type endpointKeysList []string

func (e endpointKeysList) Len() int {
	return len(e)
}

func (e endpointKeysList) Less(i, j int) bool {
	return endpointMap[e[i]].weight < endpointMap[e[j]].weight
}

func (e endpointKeysList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}

type endpoint struct {
	key             string
	lbRuleKey       string
	endpointIP      string
	endpointPort    int
	weight          int
	status          bool
	method          string
	healthCheckIP   string
	healthCheckPort int
}

// type rule struct {
// 	vip          string
// 	vPort        int
// 	endpointIP   string
// 	endpointPort int
// 	weight       int
// 	status       bool
// }

func (n *Iptablescontroller) OnLoadbalanceAdd(loadbalancerrule *v1.LoadBalancerRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("OnLoadbalanceAdd called")
	// key := rule.GetNamespace() + rule.GetName()
	if validationCheck(loadbalancerrule) == INVALIDE {
		return fmt.Errorf("LB Rule is invalid")
	}
	lbRuleSetKey := loadbalancerrule.GetNamespace() + loadbalancerrule.GetName()

	if _, exist := lbRuleSetMap[lbRuleSetKey]; exist {
		klog.Info("Duplicate resource")
		return fmt.Errorf("duplicated name of resource: %s", lbRuleSetKey)
	}

	// add lbRuleset
	lbRuleSetInstance := &lbRuleSet{
		key:        lbRuleSetKey,
		lbRuleKeys: make([]string, 0),
	}
	lbRuleSetMap[lbRuleSetKey] = lbRuleSetInstance
	for _, lbrule := range loadbalancerrule.Spec.Rules {
		klog.Infoln("\nUn LB rule added\n", lbrule)
		// lbRuleKey := generatelbkey(&lbrule)
		// lbRuleSetInstance.lbRuleKeys = append(lbRuleSetInstance.lbRuleKeys, lbRuleKey)
		n.lbRuleAddEventHandler(&lbrule, lbRuleSetKey)
	}
	// lbRuleSetMap[lbRuleSetKey] = lbRuleSetInstance

	return nil
}

func (n *Iptablescontroller) lbRuleAddEventHandler(lbrule *v1.LBRules, lbRuleSetKey string) {
	klog.Info("lbRuleAddEventHandler called")

	ctx, cancle := context.WithCancel(n.ctx)
	lbRuleKey := generatelbkey(lbrule)
	klog.Info(lbRuleKey)
	lbRuleInstance := &lbRule{
		key:           lbRuleKey,
		lbVirtualIP:   lbrule.LoadBalancerIP,
		lbVirtualPort: lbrule.LoadBalancerPort,
		protocol:      lbrule.Protocol,
		endpointKeys:  make([]string, 0),
		healthCheckManager: &healthCheckerInstance{
			runner: healthchecker.NewHealthChecker(&healthchecker.Config{
				Interval: healthchecker.DEFAULT_INTERVAL,
				Timeout:  healthchecker.DEFAULT_TIMEOUT,
			}),
			stopCh: make(chan struct{}),
		},
		lbRuleSetKey: lbRuleSetKey,
		ctx:          ctx,
		cancle:       cancle,
		// naive buffer size, no reason for value(20)
		endpointAdded:   make(chan string, 20),
		endpointDeleted: make(chan string, 20),
	}
	lbRuleMap[lbRuleKey] = lbRuleInstance
	lbRuleSetMap[lbRuleSetKey].lbRuleKeys = append(lbRuleSetMap[lbRuleSetKey].lbRuleKeys, lbRuleKey)

	if err := n.ensureLBChain(lbRuleKey, string(iptables.TableNAT), lbRuleKey, string(natPreroutingStaticNATChain)); err != nil {
		klog.Error(err)
	}
	if err := setRouteForProxyARP(lbRuleInstance.lbVirtualIP); err != nil {
		klog.ErrorS(err, "setRouteForProxyARP")
	}

	go n.runLBRuleSyncer(lbRuleKey)

	klog.Info(lbrule.Backends)
	for _, backend := range lbrule.Backends {
		// endpointKey := generateEndpointkey(&backend)
		// lbRuleInstance.endpointKeys = append(lbRuleInstance.endpointKeys, endpointKey)
		n.endpointAddEventHandler(&backend, lbRuleKey)
	}
}

func (n *Iptablescontroller) lbRuleDeleteEventHandler(lbRuleKey string) error {
	klog.Info("lbRuleDeleteEventHandler called")
	lbRuleInstance := lbRuleMap[lbRuleKey]
	lbRuleInstance.cancle()

	for _, endpointKey := range lbRuleInstance.endpointKeys {
		if err := n.endpointDeleteEventHandler(endpointKey); err != nil {
			return err
		}
	}

	if err := n.flushLBChain(lbRuleKey, string(iptables.TableNAT), lbRuleKey); err != nil { //
		klog.Error(err)
		return err
	}

	if err := n.deleteLBChain(lbRuleKey, string(iptables.TableNAT), lbRuleKey, string(natPreroutingStaticNATChain)); err != nil {
		klog.Error(err)
		return err
	}

	if err := delRouteForProxyARP(lbRuleInstance.lbVirtualIP); err != nil {
		return err
	}

	lbRuleSetInstance := lbRuleSetMap[lbRuleInstance.lbRuleSetKey]
	var lbRuleKeyIndex int
	for i, key := range lbRuleSetInstance.lbRuleKeys {
		if key == lbRuleKey {
			lbRuleKeyIndex = i
			break
		}
	}
	lbRuleSetInstance.lbRuleKeys = append(lbRuleSetInstance.lbRuleKeys[:lbRuleKeyIndex], lbRuleSetInstance.lbRuleKeys[lbRuleKeyIndex+1:]...)
	// delete(lbRuleMap, lbRuleKey)

	return nil
}

func (n *Iptablescontroller) endpointAddEventHandler(backend *v1.Backend, lbRuleKey string) {
	endpointKey := generateEndpointkey(lbRuleKey, backend)
	klog.Info(endpointKey)
	var status bool

	endpointInstance := &endpoint{
		key:             endpointKey,
		endpointIP:      backend.BackendIP,
		endpointPort:    backend.BackendPort,
		weight:          backend.Weight,
		status:          status,
		method:          backend.HealthCheckMethod,
		lbRuleKey:       lbRuleKey,
		healthCheckIP:   backend.HealthCheckIP,
		healthCheckPort: backend.HealthCheckPort,
	}
	endpointMap[endpointKey] = endpointInstance
	lbRuleMap[lbRuleKey].mu.Lock()
	defer lbRuleMap[lbRuleKey].mu.Unlock()
	lbRuleMap[lbRuleKey].endpointKeys = append(lbRuleMap[lbRuleKey].endpointKeys, endpointKey)
	sort.Sort(lbRuleMap[lbRuleKey].endpointKeys)
	lbRuleMap[lbRuleKey].endpointAdded <- endpointKey
}

func (n *Iptablescontroller) endpointDeleteEventHandler(endpointKey string) error {
	lbRuleKey := endpointMap[endpointKey].lbRuleKey
	lbRuleMap[lbRuleKey].mu.Lock()
	defer lbRuleMap[lbRuleKey].mu.Unlock()
	lbRuleInstance := lbRuleMap[lbRuleKey]

	var endpointKeyIndex int
	for i, key := range lbRuleInstance.endpointKeys {
		if key == endpointKey {
			endpointKeyIndex = i
			break
		}
	}
	// lbRuleInstance.endpointKeys[endpointKeyIndex] = lbRuleInstance.endpointKeys[len(lbRuleInstance.endpointKeys)-1]
	// lbRuleInstance.endpointKeys = lbRuleInstance.endpointKeys[:len(lbRuleInstance.endpointKeys)-1]
	lbRuleInstance.endpointKeys = append(lbRuleInstance.endpointKeys[:endpointKeyIndex], lbRuleInstance.endpointKeys[endpointKeyIndex+1:]...)
	targetKey := healthchecker.GenerateEndpointKey(&healthchecker.Endpoint{
		IPaddr:      net.ParseIP(endpointMap[endpointKey].healthCheckIP),
		Port:        endpointMap[endpointKey].healthCheckPort,
		HealthCheck: healthchecker.METHOD(endpointMap[endpointKey].method),
	})
	lbRuleInstance.endpointDeleted <- targetKey
	delete(endpointMap, endpointKey)

	return nil
}

func (n *Iptablescontroller) runLBRuleSyncer(lbRuleKey string) {
	// lbRuleInstance := lbRuleMap[lbRuleKey]
	// changedRules := make([]*rule, 0)
	// changedEndpoint := make([]string, 0)

	statusSyncInterval := time.NewTicker(1 * time.Second)
	go lbRuleMap[lbRuleKey].healthCheckManager.runner.Run(lbRuleMap[lbRuleKey].healthCheckManager.stopCh)
	klog.Info("deployRules called")
	n.deployRules(lbRuleMap[lbRuleKey])

	for {
		ruleChanged := false
		select {
		case <-statusSyncInterval.C:
			for _, endpointKey := range lbRuleMap[lbRuleKey].endpointKeys {
				var endpointInstance *endpoint
				if endpoint, exist := endpointMap[endpointKey]; exist {
					endpointInstance = endpoint
				}
				healthCheckTargetKey := healthchecker.GenerateEndpointKey(
					&healthchecker.Endpoint{
						IPaddr:      net.ParseIP(endpointInstance.healthCheckIP),
						Port:        endpointInstance.healthCheckPort,
						HealthCheck: healthchecker.METHOD(endpointInstance.method),
					})
				// klog.Info(endpointKey)
				// klog.Info(endpointInstance.status)
				// klog.Info(healthCheckTargetKey)
				// klog.Info(lbRuleMap[lbRuleKey].healthCheckManager.runner.GetTargetStatus(healthCheckTargetKey))
				if endpointInstance.status != lbRuleMap[lbRuleKey].healthCheckManager.runner.GetTargetStatus(healthCheckTargetKey) {
					ruleChanged = true
					endpointInstance.status = lbRuleMap[lbRuleKey].healthCheckManager.runner.GetTargetStatus(healthCheckTargetKey)
				}

				// changedEndpoint = append(changedEndpoint, endpointKey)
				// 	changedRules = append(changedRules, &rule{
				// 		vip:          lbRuleInstance.lbVirtualIP,
				// 		vPort:        lbRuleInstance.lbVirtualPort,
				// 		endpointIP:   endpointMap[endpointKey].endpointIP,
				// 		endpointPort: endpointMap[endpointKey].endpointPort,
				// 		weight:       endpointMap[endpointKey].weight,
				// 		status:       endpointMap[endpointKey].status,
				// 	})
				// }
			}
			if ruleChanged {
				klog.Info("deployRules called")
				n.deployRules(lbRuleMap[lbRuleKey])
			}
		case endpointKey := <-lbRuleMap[lbRuleKey].endpointAdded:
			klog.Info("endpoint added")
			klog.Info(endpointKey)
			targetKey := healthchecker.GenerateEndpointKey(&healthchecker.Endpoint{
				IPaddr:      net.ParseIP(endpointMap[endpointKey].healthCheckIP),
				Port:        endpointMap[endpointKey].healthCheckPort,
				HealthCheck: healthchecker.METHOD(endpointMap[endpointKey].method),
			})
			lbRuleMap[lbRuleKey].healthCheckManager.runner.EnsureTarget(targetKey, net.ParseIP(endpointMap[endpointKey].healthCheckIP), endpointMap[endpointKey].healthCheckPort, healthchecker.METHOD(endpointMap[endpointKey].method))
			klog.Info("deployRules called")
			n.deployRules(lbRuleMap[lbRuleKey])
		case targetKey := <-lbRuleMap[lbRuleKey].endpointDeleted:
			klog.Info("endpoint deleted")
			klog.Info(targetKey)
			lbRuleMap[lbRuleKey].healthCheckManager.runner.ClearTargetWithKey(targetKey)
			klog.Info("deployRules called")
			n.deployRules(lbRuleMap[lbRuleKey])
		case <-lbRuleMap[lbRuleKey].ctx.Done():
			klog.Info(lbRuleKey)
			lbRuleMap[lbRuleKey].healthCheckManager.stopCh <- struct{}{}
			klog.Infoln("Done!")
			delete(lbRuleMap, lbRuleKey)
			return
		}
	}
}

// func (n *Iptablescontroller) appendIPtablesRule

// ToDo: enhance flush method -> specific change method, have to use line number
func (n *Iptablescontroller) deployRules(lbRuleInstance *lbRule) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.iptablesdata.Reset()
	n.flushLBChain(lbRuleInstance.key, string(iptables.TableNAT), lbRuleInstance.key)
	n.iptables.SaveInto(iptables.TableNAT, n.iptablesdata)
	lines := strings.Split(n.iptablesdata.String(), "\n")

	lines = lines[1 : len(lines)-3] // remove tails with COMMIT

	rules := transRule(lbRuleInstance)
	chainName := lbRuleInstance.key
	for _, rule := range rules {
		n.appendRule(rule, chainName, &lines, rule.Args...)
	}

	lines = append(lines, "COMMIT")
	n.iptablesdata.Reset()
	writeLine(n.iptablesdata, lines...)
	klog.Infof("Deploying rules : %s", n.iptablesdata.String())

	if err := n.iptables.Restore("nat", n.iptablesdata.Bytes(), true, true); err != nil {
		klog.Error(err)
	}
}

func (n *Iptablescontroller) OnLoadbalanceDelete(loadbalancerrule *v1.LoadBalancerRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	klog.Info("OnLoadbalanceDelete called")

	lbRuleSetKey := loadbalancerrule.GetNamespace() + loadbalancerrule.GetName()
	if _, exist := lbRuleSetMap[lbRuleSetKey]; !exist {
		klog.Errorf("duplicated Deletion key name of resource: %s", lbRuleSetKey)
		return nil
	}

	for _, lbRuleKey := range lbRuleSetMap[lbRuleSetKey].lbRuleKeys {
		if err := n.lbRuleDeleteEventHandler(lbRuleKey); err != nil {
			return err
		}
	}
	delete(lbRuleSetMap, lbRuleSetKey)
	return nil
}

func (n *Iptablescontroller) OnLoadbalanceUpdate(loadbalancerrule *v1.LoadBalancerRule) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	klog.Info("OnLoadbalanceUpdate called")
	if validationCheck(loadbalancerrule) == INVALIDE {
		return fmt.Errorf("LB Rule is invalid")
	}

	lbRuleSetKey := loadbalancerrule.GetNamespace() + loadbalancerrule.GetName()

	if _, exist := lbRuleSetMap[lbRuleSetKey]; !exist {
		return fmt.Errorf("empty ruleset of which name of resource: %s", lbRuleSetKey)
	}

	deleteKeys := make(map[string]bool)
	for _, lbRuleKey := range lbRuleSetMap[lbRuleSetKey].lbRuleKeys {
		deleteKeys[lbRuleKey] = true
	}
	// addKeys := make(map[string]bool)
	for _, lbrule := range loadbalancerrule.Spec.Rules {
		lbRuleKey := generatelbkey(&lbrule)
		if _, exist := deleteKeys[lbRuleKey]; exist {
			deleteKeys[lbRuleKey] = false
			klog.Info("lbRuleUpdateEventHandler called")
			n.lbRuleUpdateEventHandler(&lbrule, lbRuleSetKey)
			continue
		}
		klog.Info("lbRuleAddEventHandler called")
		n.lbRuleAddEventHandler(&lbrule, lbRuleSetKey)
	}

	for lbRuleKey, deleteFlag := range deleteKeys {
		if deleteFlag {
			klog.Info("lbRuleDeleteEventHandler called")
			klog.Info(lbRuleKey)
			if err := n.lbRuleDeleteEventHandler(lbRuleKey); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Iptablescontroller) lbRuleUpdateEventHandler(lbrule *v1.LBRules, lbRuleSetKey string) error {
	klog.Info("lbRuleUpdateEventHandler called")
	lbRuleKey := generatelbkey(lbrule)

	if _, exist := lbRuleMap[lbRuleKey]; !exist {
		return fmt.Errorf("empty rule of which name of resource: %s", lbRuleKey)
	}

	deleteKeys := make(map[string]bool)
	for _, endpointKey := range lbRuleMap[lbRuleKey].endpointKeys {
		deleteKeys[endpointKey] = true
	}
	// addKeys := make(map[string]bool)
	for _, backend := range lbrule.Backends {
		endpointKey := generateEndpointkey(lbRuleKey, &backend)
		if _, exist := deleteKeys[endpointKey]; exist {
			deleteKeys[endpointKey] = false
			continue
		}
		klog.Info("endpointAddEventHandler called")
		n.endpointAddEventHandler(&backend, lbRuleKey)
	}

	for endpointKey, deleteFlag := range deleteKeys {
		if deleteFlag {
			n.endpointDeleteEventHandler(endpointKey)
		}
	}
	return nil
}

type LBRULEVALIDATION bool
type COMMAND int

const (
	INVALIDE        LBRULEVALIDATION = false
	VALID           LBRULEVALIDATION = true
	ADDCOMMAND      COMMAND          = 0
	DELETECOMMAND   COMMAND          = 1
	MODIFIEDCOMMAND COMMAND          = 2
)

type healthCheckerInstance struct {
	runner healthchecker.HealthChecker
	stopCh chan struct{}
}

func transRule(lbRuleInstance *lbRule) []*v1.Rules {
	var rules []*v1.Rules
	var totalWeight int
	totalWeight = 0

	for _, endpointKey := range lbRuleInstance.endpointKeys {
		if !endpointMap[endpointKey].status { //if the endpoint is not alive
			continue
		}
		totalWeight = totalWeight + endpointMap[endpointKey].weight
	}

	for _, endpointKey := range lbRuleInstance.endpointKeys {
		// klog.Info(endpointMap[endpointKey])
		if !endpointMap[endpointKey].status {
			continue
		}
		// weight := float64(endpointMap[endpointKey].weight) * 0.01
		weight := float64(endpointMap[endpointKey].weight) / float64(totalWeight)
		klog.Infoln("weight:", weight, "endpointWeigh:", endpointMap[endpointKey].weight, "total:", totalWeight)
		totalWeight = totalWeight - endpointMap[endpointKey].weight

		var rule *v1.Rules
		if lbRuleInstance.lbVirtualPort == 0 || endpointMap[endpointKey].endpointPort == 0 {
			rule = &v1.Rules{
				Match: v1.Match{
					DstIP: lbRuleInstance.lbVirtualIP,
				},
				Action: v1.Action{
					DstIP: endpointMap[endpointKey].endpointIP,
				},
				Args: []string{
					// "-m statistic --mode random --probability " + strconv.FormatFloat(weight, 'e', -1, 64),
					"-m statistic --mode random --probability " + fmt.Sprintf("%.2f", weight),
				},
			}
		} else {
			rule = &v1.Rules{
				Match: v1.Match{
					DstIP:    lbRuleInstance.lbVirtualIP,
					DstPort:  lbRuleInstance.lbVirtualPort,
					Protocol: lbRuleInstance.protocol,
				},
				Action: v1.Action{
					DstIP:   endpointMap[endpointKey].endpointIP,
					DstPort: endpointMap[endpointKey].endpointPort,
				},
				Args: []string{
					// "-m statistic --mode random --probability " + strconv.FormatFloat(weight, 'e', -1, 64),
					"-m statistic --mode random --probability " + fmt.Sprintf("%.2f", weight),
				},
			}
		}

		rules = append(rules, rule)
	}

	return rules
}

func validationCheck(lbRule *v1.LoadBalancerRule) LBRULEVALIDATION {
	// validate lbRule instance
	if validLBRuleFormat(&lbRule.Spec.Rules) == INVALIDE {
		return INVALIDE
	}

	// validate existing key(LoadBalancerIP:LoadBalancerPort)
	if validDuplicatekey(&lbRule.Spec.Rules) == INVALIDE {
		return INVALIDE
	}

	return VALID
}

func validLBRuleFormat(lbRules *[]v1.LBRules) LBRULEVALIDATION {
	m := make(map[string]bool)
	for _, rule := range *lbRules {
		key := generatelbkey(&rule)
		if _, exist := m[key]; exist {
			klog.Info("The key exists")
			return INVALIDE
		} else {
			// insert but do nothing
			m[key] = true
		}
		if rule.LoadBalancerPort > 65535 {
			klog.Info("Too large vport number:", rule.LoadBalancerPort)
			return INVALIDE
		}
		for _, endpoint := range rule.Backends {
			if endpoint.BackendPort > 65535 {
				klog.Info("Too large Backend port number:", endpoint.BackendPort)
				return INVALIDE
			}
			if endpoint.HealthCheckPort > 65535 {
				klog.Info("Too large Healthcheck port number:", endpoint.HealthCheckPort)
				return INVALIDE
			}
			if (endpoint.BackendPort != 0 && rule.LoadBalancerPort == 0) ||
				(rule.LoadBalancerPort != 0 && endpoint.BackendPort == 0) {
				klog.Info("Abnormal port number composition:", endpoint.BackendPort, rule.LoadBalancerPort, rule.LoadBalancerPort, endpoint.BackendPort)
				return INVALIDE
			}
			if rule.LoadBalancerPort != 0 && rule.Protocol == "" {
				klog.Info("Cannot enforce the LB rule: In case the LB port is specified, protocol information should  be specified (iptables constraint)")
				return INVALIDE
			}
			if endpoint.HealthCheckMethod == string(healthchecker.L4TCPHEALTHCHECK) &&
				endpoint.HealthCheckPort == 0 {
				return INVALIDE
			}
		}
	}
	return VALID
}

func validDuplicatekey(lbRules *[]v1.LBRules) LBRULEVALIDATION {
	for _, rule := range *lbRules {
		key := generatelbkey(&rule)
		if _, exist := registeredLBRule[key]; exist {
			return INVALIDE
		}
	}
	return VALID
}

// key: loadbalanverIP:loadbalancerPort
func generatelbkey(lbrule *v1.LBRules) string {
	return lbrule.LoadBalancerIP + ":" + strconv.Itoa(lbrule.LoadBalancerPort)
}

// key: backend.BackendIP:backend.BackendPort
// func generateEndpointkey(backend *v1.Backend) string {
// 	return backend.BackendIP + ":" + strconv.Itoa(backend.BackendPort)
// }

func generateEndpointkey(lbRuleKey string, backend *v1.Backend) string {
	return lbRuleKey + backend.BackendIP + strconv.Itoa(backend.BackendPort) + backend.HealthCheckIP +
		strconv.Itoa(backend.HealthCheckPort) + strconv.Itoa(backend.Weight) + backend.HealthCheckMethod
}
