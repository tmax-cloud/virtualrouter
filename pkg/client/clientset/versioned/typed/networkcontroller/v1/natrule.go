/*
Copyright The Kubernetes Authors.

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

// Code generated by client-gen. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	v1 "github.com/tmax-cloud/virtualrouter/pkg/apis/networkcontroller/v1"
	scheme "github.com/tmax-cloud/virtualrouter/pkg/client/clientset/versioned/scheme"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// NATRulesGetter has a method to return a NATRuleInterface.
// A group's client should implement this interface.
type NATRulesGetter interface {
	NATRules(namespace string) NATRuleInterface
}

// NATRuleInterface has methods to work with NATRule resources.
type NATRuleInterface interface {
	Create(ctx context.Context, nATRule *v1.NATRule, opts metav1.CreateOptions) (*v1.NATRule, error)
	Update(ctx context.Context, nATRule *v1.NATRule, opts metav1.UpdateOptions) (*v1.NATRule, error)
	UpdateStatus(ctx context.Context, nATRule *v1.NATRule, opts metav1.UpdateOptions) (*v1.NATRule, error)
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Get(ctx context.Context, name string, opts metav1.GetOptions) (*v1.NATRule, error)
	List(ctx context.Context, opts metav1.ListOptions) (*v1.NATRuleList, error)
	Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.NATRule, err error)
	NATRuleExpansion
}

// nATRules implements NATRuleInterface
type nATRules struct {
	client rest.Interface
	ns     string
}

// newNATRules returns a NATRules
func newNATRules(c *TmaxV1Client, namespace string) *nATRules {
	return &nATRules{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the nATRule, and returns the corresponding nATRule object, and an error if there is any.
func (c *nATRules) Get(ctx context.Context, name string, options metav1.GetOptions) (result *v1.NATRule, err error) {
	result = &v1.NATRule{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("natrules").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of NATRules that match those selectors.
func (c *nATRules) List(ctx context.Context, opts metav1.ListOptions) (result *v1.NATRuleList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1.NATRuleList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("natrules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested nATRules.
func (c *nATRules) Watch(ctx context.Context, opts metav1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("natrules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a nATRule and creates it.  Returns the server's representation of the nATRule, and an error, if there is any.
func (c *nATRules) Create(ctx context.Context, nATRule *v1.NATRule, opts metav1.CreateOptions) (result *v1.NATRule, err error) {
	result = &v1.NATRule{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("natrules").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nATRule).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a nATRule and updates it. Returns the server's representation of the nATRule, and an error, if there is any.
func (c *nATRules) Update(ctx context.Context, nATRule *v1.NATRule, opts metav1.UpdateOptions) (result *v1.NATRule, err error) {
	result = &v1.NATRule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("natrules").
		Name(nATRule.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nATRule).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *nATRules) UpdateStatus(ctx context.Context, nATRule *v1.NATRule, opts metav1.UpdateOptions) (result *v1.NATRule, err error) {
	result = &v1.NATRule{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("natrules").
		Name(nATRule.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(nATRule).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the nATRule and deletes it. Returns an error if one occurs.
func (c *nATRules) Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("natrules").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *nATRules) DeleteCollection(ctx context.Context, opts metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Namespace(c.ns).
		Resource("natrules").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched nATRule.
func (c *nATRules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts metav1.PatchOptions, subresources ...string) (result *v1.NATRule, err error) {
	result = &v1.NATRule{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("natrules").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
