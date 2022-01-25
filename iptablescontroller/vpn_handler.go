package iptablescontroller

import (
	"github.com/tmax-cloud/virtualrouter/executor/iptables"
	networkv1alpha1 "github.com/tmax-cloud/virtualrouter/pkg/apis/network/v1alpha1"
	"k8s.io/klog/v2"
)

func (i *Iptablescontroller) UpdateVPNPolicyRule(vpn networkv1alpha1.VPN) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	key := vpn.GetNamespace() + "/" + vpn.GetName()

	oldVPN, exist := i.vpnMap[key]
	if exist {
		if isVPNConnectionChanged(oldVPN, vpn) == false {
			klog.InfoS("No iptables change for vpn", "VPN", key)
			i.vpnMap[key] = vpn
			return nil
		}

		// Delete old rule
		i.deleteVPNPolicyRule(oldVPN)
	}

	// Add (new) rule
	err := i.ensureVPNPolicyRule(vpn)
	if err != nil {
		return err
	}

	i.vpnMap[key] = vpn

	return nil
}

func (i *Iptablescontroller) DeleteVPNPolicyRule(key string) {
	i.mu.Lock()
	defer i.mu.Unlock()

	vpn, exist := i.vpnMap[key]
	if !exist {
		klog.InfoS("No VPN entry exists", "key", key)
		return
	}
	i.deleteVPNPolicyRule(vpn)

	delete(i.vpnMap, key)

	return
}

func isVPNConnectionChanged(old, new networkv1alpha1.VPN) bool {
	if len(old.Spec.Connections) != len(new.Spec.Connections) {
		return true
	}

	for i, oldConn := range old.Spec.Connections {
		if oldConn.LeftSubnet != new.Spec.Connections[i].LeftSubnet {
			return true
		}
		if oldConn.RightSubnet != new.Spec.Connections[i].RightSubnet {
			return true
		}
	}

	return false
}

func (i *Iptablescontroller) ensureVPNPolicyRule(vpn networkv1alpha1.VPN) error {
	prevArgs := [][]string{}
	for _, conn := range vpn.Spec.Connections {
		args := []string{}
		args = append(args,
			"-s", conn.LeftSubnet, "-d", conn.RightSubnet,
			"-m", "comment", "--comment", vpn.GetName()+"/"+conn.Name,
			"-m", "policy", "--dir", "out", "--pol", "ipsec",
			"-j", "ACCEPT")

		_, err := i.iptables.EnsureRule(iptables.Append, iptables.TableNAT, natPostroutingVPNChain, args...)
		if err != nil {
			klog.ErrorS(err, "Failed to ensure VPN policy rule", "vpn", vpn.GetName(), "conn", conn.Name,
				"leftSubnet", conn.LeftSubnet, "rightSubnet", conn.RightSubnet)

			for _, prevArg := range prevArgs {
				err := i.iptables.DeleteRule(iptables.TableNAT, natPostroutingVPNChain, prevArg...)
				if err != nil {
					klog.ErrorS(err, "Failed cleanup previous VPN policy rule",
						"leftSubnet", prevArg[1], "rightSubnet", prevArg[3])
				}
			}

			return err
		}

		klog.InfoS("Ensured VPN policy rule", "vpn", vpn.GetName(), "conn", conn.Name,
			"leftSubnet", conn.LeftSubnet, "rightSubnet", conn.RightSubnet)

		prevArgs = append(prevArgs, args)
	}

	return nil
}

func (i *Iptablescontroller) deleteVPNPolicyRule(vpn networkv1alpha1.VPN) {
	for _, conn := range vpn.Spec.Connections {
		args := []string{}
		args = append(args,
			"-s", conn.LeftSubnet, "-d", conn.RightSubnet,
			"-m", "comment", "--comment", vpn.GetName()+"/"+conn.Name,
			"-m", "policy", "--dir", "out", "--pol", "ipsec",
			"-j", "ACCEPT")

		err := i.iptables.DeleteRule(iptables.TableNAT, natPostroutingVPNChain, args...)
		if err != nil {
			klog.ErrorS(err, "Failed to delete VPN policy rule", "vpn", vpn.GetName(), "conn", conn.Name,
				"leftSubnet", conn.LeftSubnet, "rightSubnet", conn.RightSubnet)
		}

		klog.InfoS("Deleted VPN policy rule", "vpn", vpn.GetName(), "conn", conn.Name,
			"leftSubnet", conn.LeftSubnet, "rightSubnet", conn.RightSubnet)
	}

	return
}
