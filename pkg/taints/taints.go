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

package taints

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

func PrintPretty(taints []v1.Taint) string {
	var result string
	taintCount := len(taints)
	if taintCount == 0 {
		result += "none"
	}
	for i, t := range taints {
		var res string
		if t.Key != "" {
			res += fmt.Sprintf("%v", t.Key)
		}
		if t.Value != "" {
			res += fmt.Sprintf("=%v", t.Value)
		}
		if t.Effect != "" {
			res += fmt.Sprintf(":%v", t.Effect)
		}
		if i < taintCount-1 {
			res += ",\n"
		}
		result += res
	}
	return result
}

func Parse(t string) (v1.Taint, error) {

	var (
		taint      v1.Taint
		key, value string
	)

	split := strings.Split(t, ":")
	switch len(split) {
	case 1:
		keyValue := strings.Split(split[0], "=")
		if len(keyValue) > 2 {
			return taint, errors.Errorf("invalid taint: %v", t)
		}
		key = keyValue[0]
		if len(keyValue) == 2 {
			value = keyValue[1]
		}
	case 2:
		taint.Effect = v1.TaintEffect(split[1])
		if err := validateTaintEffect(taint.Effect); err != nil {
			return taint, err
		}

		keyValue := strings.Split(split[0], "=")
		if len(keyValue) > 2 {
			return taint, errors.Errorf("invalid taint: %v", t)
		}

		key = keyValue[0]
		if len(keyValue) == 2 {
			value = keyValue[1]
		}
	default:
		return taint, errors.Errorf("invalid taint: %v", t)
	}

	taint.Key = key
	taint.Value = value

	return taint, nil
}

func validateTaintEffect(effect v1.TaintEffect) error {
	if effect != v1.TaintEffectNoSchedule && effect != v1.TaintEffectPreferNoSchedule && effect != v1.TaintEffectNoExecute {
		return fmt.Errorf("invalid taint effect: %v, unsupported taint effect", effect)
	}

	return nil
}
