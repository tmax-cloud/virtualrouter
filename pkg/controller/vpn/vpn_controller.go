package vpn

import (
	"fmt"
	"strconv"
	"time"

	"github.com/tmax-cloud/virtualrouter/iptablescontroller"
	"github.com/tmax-cloud/virtualrouter/pkg/apis/network/v1alpha1"
	clientset "github.com/tmax-cloud/virtualrouter/pkg/generated/clientset/versioned"
	"github.com/tmax-cloud/virtualrouter/pkg/generated/clientset/versioned/scheme"
	informers "github.com/tmax-cloud/virtualrouter/pkg/generated/informers/externalversions/network/v1alpha1"
	listers "github.com/tmax-cloud/virtualrouter/pkg/generated/listers/network/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	v1core "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	klog "k8s.io/klog/v2"
)

const (
	maxRetries     int = 5
	controllerName     = "VPNController"
)

type VPNController struct {
	kubeClientset   kubernetes.Interface
	customClientset clientset.Interface

	vpnLister listers.VPNLister
	vpnSynced cache.InformerSynced

	queue    workqueue.RateLimitingInterface
	recorder record.EventRecorder

	connectionCache    map[string][]string
	secretCache        map[string][]string
	iptablesController *iptablescontroller.Iptablescontroller
	forwardMark        uint32
}

func NewVPNController(
	kubeClientset kubernetes.Interface,
	customClientset clientset.Interface,
	vpnInformer informers.VPNInformer,
	iptablesController *iptablescontroller.Iptablescontroller,
	forwardMark uint32) *VPNController {

	broadcaster := record.NewBroadcaster()
	broadcaster.StartStructuredLogging(0)
	broadcaster.StartRecordingToSink(&v1core.EventSinkImpl{Interface: kubeClientset.CoreV1().Events("")})
	recorder := broadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerName})

	v := &VPNController{
		kubeClientset:      kubeClientset,
		customClientset:    customClientset,
		queue:              workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "vpn"),
		recorder:           recorder,
		iptablesController: iptablesController,
		forwardMark:        forwardMark,
	}

	vpnInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: v.onVPNUpdate,
		UpdateFunc: func(old, cur interface{}) {
			v.onVPNUpdate(cur)
		},
		DeleteFunc: v.onVPNDelete,
	})
	v.vpnLister = vpnInformer.Lister()
	v.vpnSynced = vpnInformer.Informer().HasSynced

	v.connectionCache = make(map[string][]string)
	v.secretCache = make(map[string][]string)

	return v
}

func (v *VPNController) onVPNUpdate(obj interface{}) {
	var key string
	var err error

	if key, err = cache.MetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	v.queue.Add(key)
}

func (v *VPNController) onVPNDelete(obj interface{}) {
	var key string
	var err error

	if key, err = cache.DeletionHandlingMetaNamespaceKeyFunc(obj); err != nil {
		utilruntime.HandleError(fmt.Errorf("couldn't get key for object %+v: %v", obj, err))
		return
	}

	v.queue.Add(key)
}

func (v *VPNController) Run(workers int, stopCh <-chan struct{}) {
	defer utilruntime.HandleCrash()
	defer v.queue.ShutDown()

	klog.Info("Starting vpn controller...")
	defer klog.Info("Shutting down vpn controller...")

	if !cache.WaitForCacheSync(stopCh, v.vpnSynced) {
		utilruntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	// No consideration for multiple workers
	for i := 0; i < 1; i++ {
		go wait.Until(v.worker, time.Second, stopCh)
	}

	<-stopCh
}

func (v *VPNController) worker() {
	for v.processNextWorkItem() {
	}
}

func (v *VPNController) processNextWorkItem() bool {
	vKey, quit := v.queue.Get()
	if quit {
		return false
	}
	defer v.queue.Done(vKey)

	err := v.syncHandler(vKey.(string))
	v.handleErr(err, vKey)

	return true
}

func (v *VPNController) handleErr(err error, key interface{}) {
	if err == nil {
		v.queue.Forget(key)
		return
	}

	if v.queue.NumRequeues(key) < maxRetries {
		klog.Infof("Error syncing vpn %v: %v", key, err)

		v.queue.AddRateLimited(key)
		return
	}

	klog.Infof("Dropping vpn %q out of the queue: %v", key, err)
	v.queue.Forget(key)
	utilruntime.HandleError(err)
}

func (v *VPNController) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return err
	}

	vpn, err := v.vpnLister.VPNs(namespace).Get(name)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}

		klog.InfoS("VPN resource deleted", "key", key)
		v.unloadAll(key)

		// delete iptables policy accept rule
		v.iptablesController.DeleteVPNPolicyRule(key)

		return nil
	}

	klog.InfoS("VPN resource updated", "key", key)

	// Parse VPN CR to create messages and send it
	conns, secrets := v.createVICIMessage(vpn)

	// add iptables policy accept rule
	err = v.iptablesController.UpdateVPNPolicyRule(*vpn)
	if err != nil {
		return err
	}

	err = v.loadSecrets(secrets, key)
	if err != nil {
		return err
	}

	err = v.loadConns(conns, key)
	if err != nil {
		return err
	}

	return nil
}

func (v *VPNController) createVICIMessage(vpn *v1alpha1.VPN) ([]Connection, []Secret) {
	conns := []Connection{}
	secrets := []Secret{}
	for _, connection := range vpn.Spec.Connections {
		espProposals := []string{}
		if connection.ESP != "" {
			espProposals = append(espProposals, connection.ESP)
		} else {
			espProposals = append(espProposals, "aes256-sha256-modp2048")
		}

		children := make(map[string]*ChildSA)
		children["net"] = &ChildSA{
			LocalTS:      []string{connection.LeftSubnet},
			RemoteTS:     []string{connection.RightSubnet},
			ESPProposals: espProposals,
			StartAction:  "start",
			SetMarkOut:   strconv.FormatUint(uint64(v.forwardMark), 10),
		}

		var version int
		switch connection.Keyexchange {
		case "ikev1":
			version = 1
		case "ikev2":
			version = 2
		default:
			version = 0
		}

		proposals := []string{}
		if connection.IKE != "" {
			proposals = append(proposals, connection.IKE)
		} else {
			proposals = append(proposals, "aes256-sha256-modp2048")
		}

		leftID := connection.Left
		if connection.LeftID != "" {
			leftID = connection.LeftID
		}

		rightID := connection.Right
		if connection.RightID != "" {
			rightID = connection.RightID
		}

		conn := Connection{
			Name:        connection.Name,
			LocalAddrs:  []string{connection.Left},
			RemoteAddrs: []string{connection.Right},
			Local: &LocalAuth{
				Auth: "psk",
				ID:   leftID,
			},
			Remote: &RemoteAuth{
				Auth: "psk",
				ID:   rightID,
			},
			Children:    children,
			Version:     version,
			Proposals:   proposals,
			Keyingtries: 0,
		}

		klog.InfoS("Creating connection message", "name", connection.Name, "left", connection.Left, "right", connection.Right,
			"leftSubnet", connection.LeftSubnet, "rightSubnet", connection.RightSubnet)
		conns = append(conns, conn)

		secret := Secret{
			ID:     "ike-" + connection.Name,
			Type:   "IKE",
			Data:   connection.PSK,
			Owners: []string{leftID, rightID},
		}

		klog.InfoS("Creating secret message", "name", secret.ID, "Owners", secret.Owners)
		secrets = append(secrets, secret)
	}

	return conns, secrets
}

func (v *VPNController) loadConns(conns []Connection, key string) error {
	connNames := []string{}
	for _, conn := range conns {
		klog.InfoS("Loading connection", "key", key, "conn", conn.Name)
		err := loadConn(conn)
		if err != nil {
			klog.ErrorS(err, "Failed to load connection", "key", key, "conn", conn.Name)
			return err
		}
		connNames = append(connNames, conn.Name)
	}

	v.connectionCache[key] = connNames

	return nil
}

func (v *VPNController) loadSecrets(secrets []Secret, key string) error {
	secretIDs := []string{}
	for _, secret := range secrets {
		klog.InfoS("Loading secret", "key", key, "secret", secret.ID)
		err := loadShared(secret)
		if err != nil {
			klog.ErrorS(err, "Failed to loading secret", "key", key, "secret", secret.ID)
			return err
		}
		secretIDs = append(secretIDs, secret.ID)
	}

	v.secretCache[key] = secretIDs

	return nil
}

func (v *VPNController) unloadAll(key string) {
	connKeys := v.connectionCache[key]
	secretKeys := v.secretCache[key]

	for _, connKey := range connKeys {
		klog.InfoS("Unloading connection", "key", key, "conn", connKey)
		err := unloadConn(connKey)
		if err != nil {
			klog.ErrorS(err, "Failed to unload connection", "conn", connKey)
		}
	}
	delete(v.connectionCache, key)

	for _, secretKey := range secretKeys {
		klog.InfoS("Unloading secret", "key", key, "secret", secretKey)
		err := unloadShared(secretKey)
		if err != nil {
			klog.ErrorS(err, "Failed to unload secret", "secret", secretKey)
		}
	}
	delete(v.secretCache, key)
}
