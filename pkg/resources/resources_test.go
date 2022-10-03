/*
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

package resources

import (
	"context"
	"reflect"
	"testing"

	v1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	fake "k8s.io/client-go/dynamic/fake"
)

func TestParse(t *testing.T) {
	tests := []struct {
		Description string
		ApiVersion  string
		Kind        string
		ExpectedGVR schema.GroupVersionResource
	}{
		{
			Description: "valid input apps/v1",
			ApiVersion:  "apps/v1",
			Kind:        "deployments",
			ExpectedGVR: _groupVersionResource("apps", "v1", "deployments"),
		},
		{
			Description: "valid input core v1",
			ApiVersion:  "v1",
			Kind:        "deployments",
			ExpectedGVR: _groupVersionResource("", "v1", "deployments"),
		},
	}

	for _, test := range tests {
		t.Log(test.Description)
		gvr := Parse(test.ApiVersion, test.Kind)
		assert.Equal(t, test.ExpectedGVR, gvr)
	}
}

func TestListResourceTolerations(t *testing.T) {
	tests := []struct {
		Description         string
		Deployments         []*unstructured.Unstructured
		ExpectedResourceMap map[ResourceReference][]v1.Toleration
	}{
		{
			Description: "deployment with tolerations",
			Deployments: []*unstructured.Unstructured{
				_unstructuredDeployment(
					"kube-system",
					"coredns",
					_toleration("Equal", "key", "value", "NoSchedule"),
					_toleration("Exists", "key2", "value2", "NoSchedule"),
				),
				_unstructuredDeployment(
					"kube-system",
					"nginx",
					_toleration("Equal", "key3", "value", "NoSchedule"),
					_toleration("Exists", "key4", "", "NoSchedule"),
				),
			},
			ExpectedResourceMap: map[ResourceReference][]v1.Toleration{
				_resourceReference("kube-system", "coredns", "Deployment"): {
					_toleration("Equal", "key", "value", "NoSchedule"),
					_toleration("Exists", "key2", "value2", "NoSchedule"),
				},
				_resourceReference("kube-system", "nginx", "Deployment"): {
					_toleration("Equal", "key3", "value", "NoSchedule"),
					_toleration("Exists", "key4", "", "NoSchedule"),
				},
			},
		},
		{
			Description: "deployment with no tolerations",
			Deployments: []*unstructured.Unstructured{
				_unstructuredDeployment(
					"kube-system",
					"coredns",
				),
				_unstructuredDeployment(
					"kube-system",
					"nginx",
					_toleration("", "key3", "value", "NoSchedule"),
					_toleration("Exists", "key4", "", "NoSchedule"),
				),
			},
			ExpectedResourceMap: map[ResourceReference][]v1.Toleration{
				_resourceReference("kube-system", "coredns", "Deployment"): {},
				_resourceReference("kube-system", "nginx", "Deployment"): {
					_toleration("Equal", "key3", "value", "NoSchedule"),
					_toleration("Exists", "key4", "", "NoSchedule"),
				},
			},
		},
	}

	gvr := _groupVersionResource("apps", "v1", "deployments")
	for _, test := range tests {
		t.Log(test.Description)
		client := _fakeClient()

		for _, deployment := range test.Deployments {
			ns := deployment.GetNamespace()
			_, err := client.Resource(gvr).Namespace(ns).Create(context.Background(), deployment, metav1.CreateOptions{})
			assert.NoError(t, err)
		}

		resourceMap, err := ListResourceTolerations(client, gvr, "kube-system")
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(test.ExpectedResourceMap, resourceMap))
	}
}

func TestListNodeTaints(t *testing.T) {
	tests := []struct {
		Description         string
		Nodes               []*unstructured.Unstructured
		ExpectedResourceMap map[ResourceReference][]v1.Taint
	}{
		{
			Description: "nodes with taints",
			Nodes: []*unstructured.Unstructured{
				_unstructuredNode("ip-1-2-3-4.ec2.internal", _taint("key", "value", "NoSchedule")),
				_unstructuredNode("ip-1-2-3-5.ec2.internal", _taint("key1", "value1", "NoSchedule")),
			},
			ExpectedResourceMap: map[ResourceReference][]v1.Taint{
				_resourceReference("", "ip-1-2-3-4.ec2.internal", "Node"): {
					_taint("key", "value", "NoSchedule"),
				},
				_resourceReference("", "ip-1-2-3-5.ec2.internal", "Node"): {
					_taint("key1", "value1", "NoSchedule"),
				},
			},
		},
		{
			Description: "nodes without taints",
			Nodes: []*unstructured.Unstructured{
				_unstructuredNode("ip-1-2-3-4.ec2.internal"),
				_unstructuredNode("ip-1-2-3-5.ec2.internal", _taint("key1", "value1", "NoSchedule")),
			},
			ExpectedResourceMap: map[ResourceReference][]v1.Taint{
				_resourceReference("", "ip-1-2-3-4.ec2.internal", "Node"): {},
				_resourceReference("", "ip-1-2-3-5.ec2.internal", "Node"): {
					_taint("key1", "value1", "NoSchedule"),
				},
			},
		},
	}

	gvr := _groupVersionResource("", "v1", "nodes")
	for _, test := range tests {
		t.Log(test.Description)
		client := _fakeClient()

		for _, node := range test.Nodes {
			_, err := client.Resource(gvr).Create(context.Background(), node, metav1.CreateOptions{})
			assert.NoError(t, err)
		}

		resourceMap, err := ListNodeTaints(client)
		assert.NoError(t, err)
		assert.True(t, reflect.DeepEqual(test.ExpectedResourceMap, resourceMap))
	}
}

func _toleration(operator, key, value, effect string) v1.Toleration {
	return v1.Toleration{
		Operator: v1.TolerationOperator(operator),
		Key:      key,
		Value:    value,
		Effect:   v1.TaintEffect(effect),
	}
}

func _taint(key, value, effect string) v1.Taint {
	return v1.Taint{
		Key:    key,
		Value:  value,
		Effect: v1.TaintEffect(effect),
	}
}

func _resourceReference(namespace, name, kind string) ResourceReference {
	return ResourceReference{
		Namespace: namespace,
		Name:      name,
		Kind:      kind,
	}
}

func _unstructuredNode(name string, taints ...v1.Taint) *unstructured.Unstructured {
	base := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Node",
			"metadata": map[string]interface{}{
				"name": name,
			},
		},
	}

	unstructuredTaints := make([]interface{}, 0)
	for _, t := range taints {
		ts, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&t)
		unstructuredTaints = append(unstructuredTaints, ts)
	}

	unstructured.SetNestedField(base.Object, unstructuredTaints, TaintPath...)
	return base
}

func _unstructuredDeployment(namespace, name string, tolerations ...v1.Toleration) *unstructured.Unstructured {
	base := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"namespace": namespace,
				"name":      name,
			},
		},
	}

	unstructuredTolerations := make([]interface{}, 0)
	for _, t := range tolerations {
		ts, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(&t)
		unstructuredTolerations = append(unstructuredTolerations, ts)
	}

	unstructured.SetNestedField(base.Object, unstructuredTolerations, TolerationPath...)
	return base
}

func _fakeClient() dynamic.Interface {
	return fake.NewSimpleDynamicClientWithCustomListKinds(runtime.NewScheme(), map[schema.GroupVersionResource]string{
		{Version: "v1", Resource: "nodes"}:                      "NodeList",
		{Group: "apps", Version: "v1", Resource: "deployments"}: "DeploymentList",
		{Group: "apps", Version: "v1", Resource: "daemonsets"}:  "DaemonsetList",
	})
}

func _groupVersionResource(group, version, resource string) schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
}
