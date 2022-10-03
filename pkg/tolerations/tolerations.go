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

package tolerations

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
)

func PrintPretty(tolerations []v1.Toleration) string {
	var result string

	tolCount := len(tolerations)
	if tolCount == 0 {
		result += "none"
	}
	for i, t := range tolerations {
		res := fmt.Sprintf("%v(", t.Operator)
		if t.Key != "" {
			res += fmt.Sprintf("%v", t.Key)
		}
		if t.Value != "" {
			res += fmt.Sprintf("=%v", t.Value)
		}
		if t.Effect != "" {
			res += fmt.Sprintf(":%v", t.Effect)
		}
		if i < tolCount-1 {
			res += "),\n"
		} else {
			res += ")"
		}
		result += res
	}
	return result
}

func Parse(t string) (v1.Toleration, error) {

	var (
		toleration   v1.Toleration
		key, value   string
		inner, outer string
	)

	// equal is the default
	toleration.Operator = v1.TolerationOpEqual

	i := strings.Index(t, "(")
	if i >= 0 {
		j := strings.Index(t, ")")
		if j >= 0 {
			inner = t[i+1 : j]
			outer = t[0 : i-1]

			if strings.EqualFold(outer, string(v1.TolerationOpExists)) {
				toleration.Operator = v1.TolerationOpExists
			}
		}
	}

	if inner == "" {
		inner = t
	}

	split := strings.Split(inner, ":")
	switch len(split) {
	case 1:
		keyValue := strings.Split(split[0], "=")
		if len(keyValue) > 2 {
			return toleration, errors.Errorf("invalid toleration: %v", t)
		}
		key = keyValue[0]
		if len(keyValue) == 2 {
			value = keyValue[1]
		}
	case 2:
		toleration.Effect = v1.TaintEffect(split[1])
		if err := validateTaintEffect(toleration.Effect); err != nil {
			return toleration, err
		}

		keyValue := strings.Split(split[0], "=")
		if len(keyValue) > 2 {
			return toleration, errors.Errorf("invalid toleration: %v", t)
		}

		key = keyValue[0]
		if len(keyValue) == 2 {
			value = keyValue[1]
		}
	default:
		return toleration, errors.Errorf("invalid toleration: %v", t)
	}

	toleration.Key = key
	toleration.Value = value

	return toleration, nil
}

func validateTaintEffect(effect v1.TaintEffect) error {
	if effect != v1.TaintEffectNoSchedule && effect != v1.TaintEffectPreferNoSchedule && effect != v1.TaintEffectNoExecute {
		return fmt.Errorf("invalid taint effect: %v, unsupported taint effect", effect)
	}

	return nil
}
