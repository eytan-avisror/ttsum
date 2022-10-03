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

package cli

import (
	"log"
	"os"
	"sort"

	"github.com/eytan-avisror/ttsum/pkg/resources"
	"github.com/eytan-avisror/ttsum/pkg/taints"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

var taintCmd = &cobra.Command{
	Use:   "taints --match [toleration]",
	Short: "taints summarizes taints for nodes, and whether they will accept a toleration",
	Long:  "For example; $ ttsum taints",
	Run:   RunTaintsCommand,
}

func RunTaintsCommand(cmd *cobra.Command, args []string) {
	if match != "" && noMatch != "" {
		log.Fatal("--match and --no-match are mutually exclusive arguments")
	}

	k8s, err := getKubernetesClient(kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	resourceTaints, err := resources.ListNodeTaints(k8s)
	if err != nil {
		log.Fatal(err)
	}

	if match != "" {
		expr, err := taints.Parse(match)
		if err != nil {
			log.Fatal(err)
		}

		resourceTaints = resources.FilterTaints(resourceTaints, expr, true)
	} else if noMatch != "" {
		expr, err := taints.Parse(noMatch)
		if err != nil {
			log.Fatal(err)
		}

		resourceTaints = resources.FilterTaints(resourceTaints, expr, false)
	}

	results := make([]TaintsResult, 0)
	for resource, rawTaints := range resourceTaints {
		results = append(results, TaintsResult{
			ResourceReference: resources.ResourceReference{
				Name:      resource.Name,
				Namespace: resource.Namespace,
			},
			Taints: rawTaints,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAME", "TAINTS"})
	table.SetAutoWrapText(false)
	table.SetAutoFormatHeaders(true)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetTablePadding("\t")
	table.SetNoWhiteSpace(true)
	data := make([][]string, 0)

	for _, result := range results {
		data = append(data, []string{result.Name, taints.PrintPretty(result.Taints)})
	}

	table.AppendBulk(data)
	table.Render()
}

type TaintsResult struct {
	resources.ResourceReference
	Taints []v1.Taint
}

func init() {
	rootCmd.AddCommand(taintCmd)
	taintCmd.Flags().StringVar(&match, "match", "", "Show resources with toleration match, must be in format Operator(key=value:effect)")
	taintCmd.Flags().StringVar(&noMatch, "no-match", "", "Show resources without toleration match, must be in format Operator(key=value:effect)")
}
