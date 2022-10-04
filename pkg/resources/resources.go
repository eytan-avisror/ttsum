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
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	TolerationPath = []string{"spec", "template", "spec", "tolerations"}
	TaintPath      = []string{"spec", "taints"}
	NodeGVR        = schema.GroupVersionResource{
		Resource: "nodes",
		Version:  "v1",
	}
)

func Parse(apiVersion, kind string) schema.GroupVersionResource {
	var gvr schema.GroupVersionResource

	gvr.Group = ""
	gvr.Resource = kind

	spl := strings.Split(apiVersion, "/")
	if len(spl) == 1 {
		gvr.Version = spl[0]
	} else if len(spl) == 2 {
		gvr.Group = spl[0]
		gvr.Version = spl[1]
	}
	return gvr
}

type ResourceReference struct {
	Namespace string
	Name      string
	Kind      string
}

func ListResourceTolerations(client dynamic.Interface, gvr schema.GroupVersionResource, namespace string) (map[ResourceReference][]v1.Toleration, error) {
	var tolerations = make(map[ResourceReference][]v1.Toleration)

	r, err := client.Resource(gvr).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return tolerations, err
	}

	for _, resource := range r.Items {
		ref := ResourceReference{
			Namespace: resource.GetNamespace(),
			Name:      resource.GetName(),
			Kind:      resource.GetKind(),
		}
		tolerations[ref] = make([]v1.Toleration, 0)
		res, ok, err := unstructured.NestedSlice(resource.Object, TolerationPath...)
		if !ok {
			continue
		}
		if err != nil {
			return tolerations, err
		}

		for _, obj := range res {
			var toleration v1.Toleration
			convert := obj.(map[string]interface{})
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(convert, &toleration)
			if err != nil {
				return tolerations, err
			}

			if toleration.Operator == "" {
				toleration.Operator = v1.TolerationOpEqual
			}
			tolerations[ref] = append(tolerations[ref], toleration)
		}
	}
	return tolerations, nil
}

func ListNodeTaints(client dynamic.Interface) (map[ResourceReference][]v1.Taint, error) {
	var taints = make(map[ResourceReference][]v1.Taint)

	r, err := client.Resource(NodeGVR).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return taints, err
	}

	for _, resource := range r.Items {
		ref := ResourceReference{
			Name: resource.GetName(),
			Kind: resource.GetKind(),
		}
		taints[ref] = make([]v1.Taint, 0)
		res, ok, err := unstructured.NestedSlice(resource.Object, TaintPath...)
		if !ok {
			continue
		}
		if err != nil {
			return taints, err
		}

		for _, obj := range res {
			var taint v1.Taint
			convert := obj.(map[string]interface{})
			err := runtime.DefaultUnstructuredConverter.FromUnstructured(convert, &taint)
			if err != nil {
				return taints, err
			}
			taints[ref] = append(taints[ref], taint)
		}
	}
	return taints, nil
}

func FilterTolerations(objs map[ResourceReference][]v1.Toleration, matchToleration v1.Toleration, condition bool) map[ResourceReference][]v1.Toleration {
	filteredMap := make(map[ResourceReference][]v1.Toleration)

	if matchToleration.Operator == "" {
		matchToleration.Operator = v1.TolerationOpEqual
	}

	for res, tols := range objs {
		var hit bool

		for _, tol := range tols {

			hit = reflect.DeepEqual(matchToleration, tol)

			if condition && hit {
				filteredMap[res] = tols
				break
			}
		}
		if !condition && !hit {
			filteredMap[res] = tols
		}
	}
	return filteredMap
}

func FilterTaints(objs map[ResourceReference][]v1.Taint, matchTaint v1.Taint, condition bool) map[ResourceReference][]v1.Taint {
	filteredMap := make(map[ResourceReference][]v1.Taint)

	for res, taints := range objs {
		var hit bool

		for _, taint := range taints {

			hit = reflect.DeepEqual(matchTaint, taint)

			if condition && hit {
				filteredMap[res] = taints
				break
			}
		}
		if !condition && !hit {
			filteredMap[res] = taints
		}
	}
	return filteredMap
}
