// This is a generated file. Do not edit directly.

module github.com/tmax-cloud/virtualrouter

go 1.16

require (
	github.com/containerd/containerd v1.5.0 // indirect
	github.com/docker/docker v20.10.6+incompatible
	github.com/go-ping/ping v0.0.0-20211014180314-6e2b003bffdd
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/moby/term v0.0.0-20201216013528-df9cb8a40635 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/vishvananda/netlink v1.1.1-0.20201029203352-d40f9887b852
	github.com/vishvananda/netns v0.0.0-20210104183010-2eb08e3e575f
	golang.org/x/net v0.0.0-20210316092652-d523dce5a7f4
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
	k8s.io/code-generator v0.19.0
	k8s.io/klog/v2 v2.8.0
	k8s.io/kubernetes v1.13.0
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920

)

replace (
	github.com/kubernetes/kubernetes => github.com/kubernetes/kubernetes v0.19.0
	k8s.io/api => k8s.io/api v0.0.0-20210329192645-60680b5087d3
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20210329192041-0c7db653e2b6
	k8s.io/apiserver => k8s.io/apiserver v0.19.0
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20210329193902-8b9f5901612d
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.0
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.0
	k8s.io/code-generator => k8s.io/code-generator v0.19.5
	k8s.io/component-base => k8s.io/component-base v0.19.0
	k8s.io/cri-api => k8s.io/cri-api v0.19.0
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.0
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.0
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.0
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.0
	k8s.io/kubectl => k8s.io/kubectl v0.19.0
	k8s.io/kubelet => k8s.io/kubelet v0.19.0
	k8s.io/kubernetes => k8s.io/kubernetes v1.19.0

	// k8s.io/component-helpers v0.0.0
	// k8s.io/controller-manager v0.0.0
	// k8s.io/cri-api v0.0.0
	// k8s.io/csi-translation-lib v0.0.0
	// k8s.io/kube-aggregator v0.0.0
	// k8s.io/kube-controller-manager v0.0.0
	// k8s.io/kube-proxy v0.0.0
	// k8s.io/kube-scheduler v0.0.0
	// k8s.io/kubectl v0.0.0
	// k8s.io/kubelet v0.0.0
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.0
	k8s.io/metrics => k8s.io/metrics v0.19.0
	// k8s.io/mount-utils v0.0.0
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.0
)
