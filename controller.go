/*
Copyright 2017 The Kubernetes Authors.

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

package main

import (
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/klog/v2"

	"github.com/tmax-cloud/virtualrouter/iptablescontroller"
	v1 "github.com/tmax-cloud/virtualrouter/pkg/apis/networkcontroller/v1"
	clientset "github.com/tmax-cloud/virtualrouter/pkg/client/clientset/versioned"
	samplescheme "github.com/tmax-cloud/virtualrouter/pkg/client/clientset/versioned/scheme"
	informers "github.com/tmax-cloud/virtualrouter/pkg/client/informers/externalversions/networkcontroller/v1"
	listers "github.com/tmax-cloud/virtualrouter/pkg/client/listers/networkcontroller/v1"
)

const controllerAgentName = "virtual-router"

const (
	// SuccessSynced is used as part of the Event 'reason' when a Foo is synced
	SuccessSynced = "Synced"
	// ErrResourceExists is used as part of the Event 'reason' when a Foo fails
	// to sync due to a Deployment of the same name already existing.
	ErrResourceExists = "ErrResourceExists"

	// MessageResourceExists is the message used for Events when a resource
	// fails to sync due to a Deployment already existing
	MessageResourceExists = "Resource %q already exists and is not managed by Foo"
	// MessageResourceSynced is the message used for an Event fired when a Foo
	// is synced successfully
	MessageResourceSynced = "Foo synced successfully"
)

const (
	SyncStateSuccess = iota
	SyncStateError
	SyncStateReprocess
)

type natRuleKey string
type natRuleChangeKey string
type natRuleDeleteKey string

type firewallRuleKey string
type firewallRuleChangeKey string
type firewallRuleDeleteKey string

type loadbalanceRuleKey string
type loadbalanceRuleChangeKey string
type loadbalanceRuleDeleteKey string

type SyncState int

// Controller is the controller implementation for Foo resources
type Controller struct {
	// kubeclientset is a standard kubernetes clientset
	kubeclientset kubernetes.Interface
	// sampleclientset is a clientset for our own API group
	sampleclientset clientset.Interface

	natRulesLister          listers.NATRuleLister
	firewallRulesLister     listers.FireWallRuleLister
	loadbalancerRulesLister listers.LoadBalancerRuleLister

	syncFuncs []cache.InformerSynced

	// workqueue is a rate limited work queue. This is used to queue work to be
	// processed instead of performing it as soon as a change happens. This
	// means we can ensure we only process a fixed amount of resources at a
	// time, and makes it easy to ensure we are never processing the same item
	// simultaneously in two different workers.
	workqueue workqueue.RateLimitingInterface

	// recorder is an event recorder for recording Event resources to the
	// Kubernetes API.
	recorder record.EventRecorder

	iptablescontroller *iptablescontroller.Iptablescontroller
}

// NewController returns a new sample controller
func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	natRuleInformer informers.NATRuleInformer,
	firewallruleInformer informers.FireWallRuleInformer,
	loadbalancerruleInformer informers.LoadBalancerRuleInformer,
	iptablescontroller *iptablescontroller.Iptablescontroller) *Controller {

	// Create event broadcaster
	// Add virtual-router types to the default Kubernetes Scheme so Events can be
	// logged for virtual-router types.
	utilruntime.Must(samplescheme.AddToScheme(scheme.Scheme))
	klog.V(4).Info("Creating event broadcaster")
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartStructuredLogging(0)
	eventBroadcaster.StartRecordingToSink(&typedcorev1.EventSinkImpl{Interface: kubeclientset.CoreV1().Events("")})
	recorder := eventBroadcaster.NewRecorder(scheme.Scheme, corev1.EventSource{Component: controllerAgentName})

	controller := &Controller{
		kubeclientset:           kubeclientset,
		sampleclientset:         sampleclientset,
		natRulesLister:          natRuleInformer.Lister(),
		firewallRulesLister:     firewallruleInformer.Lister(),
		loadbalancerRulesLister: loadbalancerruleInformer.Lister(),
		syncFuncs:               make([]cache.InformerSynced, 0),
		workqueue:               workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter()),
		recorder:                recorder,
	}
	controller.syncFuncs = append(controller.syncFuncs, natRuleInformer.Informer().HasSynced)
	controller.syncFuncs = append(controller.syncFuncs, firewallruleInformer.Informer().HasSynced)
	controller.syncFuncs = append(controller.syncFuncs, loadbalancerruleInformer.Informer().HasSynced)

	klog.Info("Setting up event handlers")
	// Set up an event handler for when Foo resources change
	natRuleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(natRuleKey(key))
			}
		},

		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(natRuleChangeKey(key))
			}
		},

		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(natRuleDeleteKey(key))
			}
		},
	})
	firewallruleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(firewallRuleKey(key))
			}
		},

		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(firewallRuleChangeKey(key))
			}
		},

		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(firewallRuleDeleteKey(key))
			}
		},
	})
	loadbalancerruleInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(loadbalanceRuleKey(key))
			}
		},

		UpdateFunc: func(old, new interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(new)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(loadbalanceRuleChangeKey(key))
			}
		},

		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err != nil {
				utilruntime.HandleError(err)
			} else {
				controller.workqueue.Add(loadbalanceRuleDeleteKey(key))
			}
		},
	})

	controller.iptablescontroller = iptablescontroller
	return controller
}

// Run will set up the event handlers for types we are interested in, as well
// as syncing informer caches and starting workers. It will block until stopCh
// is closed, at which point it will shutdown the workqueue and wait for
// workers to finish processing their current work items.
func (c *Controller) Run(threadiness int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workqueue.ShutDown()

	// Start the informer factories to begin populating the informer caches
	klog.Info("Starting Foo controller")

	// Wait for the caches to be synced before starting workers
	klog.Info("Waiting for informer caches to sync")
	if ok := cache.WaitForCacheSync(stopCh, c.syncFuncs...); !ok {
		return fmt.Errorf("failed to wait for caches to sync")
	}

	klog.Info("Starting workers")
	// Launch two workers to process Foo resources
	for i := 0; i < threadiness; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}

	klog.Info("Started workers")
	<-stopCh
	klog.Info("Shutting down workers")

	return nil
}

// runWorker is a long-running function that will continually call the
// processNextWorkItem function in order to read and process a message on the
// workqueue.
func (c *Controller) runWorker() {
	for c.processNextWorkItem() {
	}
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the syncHandler.
func (c *Controller) processNextWorkItem() bool {
	obj, shutdown := c.workqueue.Get()

	if shutdown {
		return false
	}

	// We wrap this block in a func so we can defer c.workqueue.Done.
	err := func(obj interface{}) error {
		// We call Done here so the workqueue knows we have finished
		// processing this item. We also must remember to call Forget if we
		// do not want this work item being re-queued. For example, we do
		// not call Forget if a transient error occurs, instead the item is
		// put back on the workqueue and attempted again after a back-off
		// period.
		defer c.workqueue.Done(obj)

		// var key string
		// var ok bool
		// if key, ok = obj.(string); !ok {

		// 	c.workqueue.Forget(obj)
		// 	utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v", obj))
		// 	return nil
		// }

		status := c.syncHandler(obj)
		switch status {
		case SyncStateSuccess:
			c.workqueue.Forget(obj)
		case SyncStateError:
			c.workqueue.AddRateLimited(obj)
		case SyncStateReprocess:
			c.workqueue.AddRateLimited(obj)
		}
		return nil
	}(obj)

	if err != nil {
		utilruntime.HandleError(err)
		return true
	}

	return true
}

// syncHandler compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (c *Controller) syncHandler(key interface{}) SyncState {
	klog.Info("syncHanlder called")
	switch k := key.(type) {
	case natRuleKey:
		if err := c.addNatRule(string(k)); err != nil {
			return SyncStateError
		}
	case natRuleChangeKey:
		if err := c.updateNatRule(string(k)); err != nil {
			return SyncStateError
		}
	case natRuleDeleteKey:
		if err := c.deleteNatRule(string(k)); err != nil {
			return SyncStateError
		}

	case firewallRuleKey:
		if err := c.addFirewallRule(string(k)); err != nil {
			return SyncStateError
		}
	case firewallRuleChangeKey:
		if err := c.updateFirewallRule(string(k)); err != nil {
			return SyncStateError
		}
	case firewallRuleDeleteKey:
		if err := c.deleteFirewallRule(string(k)); err != nil {
			return SyncStateError
		}

	case loadbalanceRuleKey:
		if err := c.addLoadbalanceRule(string(k)); err != nil {
			return SyncStateError
		}
	case loadbalanceRuleChangeKey:
		if err := c.updateLoadbalanceRule(string(k)); err != nil {
			return SyncStateError
		}
	case loadbalanceRuleDeleteKey:
		if err := c.deleteLoadbalanceRule(string(k)); err != nil {
			return SyncStateError
		}
	}

	// Convert the namespace/name string into a distinct namespace and name
	//namespace, name, err := cache.SplitMetaNamespaceKey(key)

	//c.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
	return SyncStateSuccess
}

func (c *Controller) addNatRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	natRule, err := c.natRulesLister.NATRules(namespace).Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	c.iptablescontroller.OnNATAdd(natRule)
	return nil
}

func (c *Controller) updateNatRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	natRule, err := c.natRulesLister.NATRules(namespace).Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	c.iptablescontroller.OnNATUpdate(natRule)
	return nil
}

func (c *Controller) deleteNatRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	emptyObj := &v1.NATRule{}
	emptyObj.SetName(name)
	emptyObj.SetNamespace(namespace)

	c.iptablescontroller.OnNATDelete(emptyObj)
	return nil
}

func (c *Controller) addFirewallRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	firewallRule, err := c.firewallRulesLister.FireWallRules(namespace).Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	c.iptablescontroller.OnFirewallAdd(firewallRule)
	return nil
}

func (c *Controller) updateFirewallRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	firewallRule, err := c.firewallRulesLister.FireWallRules(namespace).Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	c.iptablescontroller.OnFirewallUpdate(firewallRule)
	return nil
}

func (c *Controller) deleteFirewallRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	emptyObj := &v1.FireWallRule{}
	emptyObj.SetName(name)
	emptyObj.SetNamespace(namespace)

	c.iptablescontroller.OnFirewallDelete(emptyObj)
	return nil
}

func (c *Controller) addLoadbalanceRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	loadbalanceRule, err := c.loadbalancerRulesLister.LoadBalancerRules(namespace).Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	c.iptablescontroller.OnLoadbalanceAdd(loadbalanceRule)
	return nil
}

func (c *Controller) updateLoadbalanceRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	loadbalanceRule, err := c.loadbalancerRulesLister.LoadBalancerRules(namespace).Get(name)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}

	c.iptablescontroller.OnLoadbalanceUpdate(loadbalanceRule)
	return nil
}

func (c *Controller) deleteLoadbalanceRule(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	emptyObj := &v1.LoadBalancerRule{}
	emptyObj.SetName(name)
	emptyObj.SetNamespace(namespace)

	c.iptablescontroller.OnLoadbalanceDelete(emptyObj)
	return nil
}
