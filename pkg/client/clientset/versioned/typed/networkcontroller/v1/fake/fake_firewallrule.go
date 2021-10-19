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

package fake

import (
	"context"

	networkcontrollerv1 "github.com/cho4036/virtualrouter/pkg/apis/networkcontroller/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeFireWallRules implements FireWallRuleInterface
type FakeFireWallRules struct {
	Fake *FakeTmaxV1
	ns   string
}

var firewallrulesResource = schema.GroupVersionResource{Group: "tmax.hypercloud.com", Version: "v1", Resource: "firewallrules"}

var firewallrulesKind = schema.GroupVersionKind{Group: "tmax.hypercloud.com", Version: "v1", Kind: "FireWallRule"}

// Get takes name of the fireWallRule, and returns the corresponding fireWallRule object, and an error if there is any.
func (c *FakeFireWallRules) Get(ctx context.Context, name string, options v1.GetOptions) (result *networkcontrollerv1.FireWallRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(firewallrulesResource, c.ns, name), &networkcontrollerv1.FireWallRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*networkcontrollerv1.FireWallRule), err
}

// List takes label and field selectors, and returns the list of FireWallRules that match those selectors.
func (c *FakeFireWallRules) List(ctx context.Context, opts v1.ListOptions) (result *networkcontrollerv1.FireWallRuleList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(firewallrulesResource, firewallrulesKind, c.ns, opts), &networkcontrollerv1.FireWallRuleList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &networkcontrollerv1.FireWallRuleList{ListMeta: obj.(*networkcontrollerv1.FireWallRuleList).ListMeta}
	for _, item := range obj.(*networkcontrollerv1.FireWallRuleList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested fireWallRules.
func (c *FakeFireWallRules) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(firewallrulesResource, c.ns, opts))

}

// Create takes the representation of a fireWallRule and creates it.  Returns the server's representation of the fireWallRule, and an error, if there is any.
func (c *FakeFireWallRules) Create(ctx context.Context, fireWallRule *networkcontrollerv1.FireWallRule, opts v1.CreateOptions) (result *networkcontrollerv1.FireWallRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(firewallrulesResource, c.ns, fireWallRule), &networkcontrollerv1.FireWallRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*networkcontrollerv1.FireWallRule), err
}

// Update takes the representation of a fireWallRule and updates it. Returns the server's representation of the fireWallRule, and an error, if there is any.
func (c *FakeFireWallRules) Update(ctx context.Context, fireWallRule *networkcontrollerv1.FireWallRule, opts v1.UpdateOptions) (result *networkcontrollerv1.FireWallRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(firewallrulesResource, c.ns, fireWallRule), &networkcontrollerv1.FireWallRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*networkcontrollerv1.FireWallRule), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeFireWallRules) UpdateStatus(ctx context.Context, fireWallRule *networkcontrollerv1.FireWallRule, opts v1.UpdateOptions) (*networkcontrollerv1.FireWallRule, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(firewallrulesResource, "status", c.ns, fireWallRule), &networkcontrollerv1.FireWallRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*networkcontrollerv1.FireWallRule), err
}

// Delete takes name of the fireWallRule and deletes it. Returns an error if one occurs.
func (c *FakeFireWallRules) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(firewallrulesResource, c.ns, name), &networkcontrollerv1.FireWallRule{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeFireWallRules) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(firewallrulesResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &networkcontrollerv1.FireWallRuleList{})
	return err
}

// Patch applies the patch and returns the patched fireWallRule.
func (c *FakeFireWallRules) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *networkcontrollerv1.FireWallRule, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(firewallrulesResource, c.ns, name, pt, data, subresources...), &networkcontrollerv1.FireWallRule{})

	if obj == nil {
		return nil, err
	}
	return obj.(*networkcontrollerv1.FireWallRule), err
}
