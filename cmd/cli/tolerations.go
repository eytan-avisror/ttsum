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
	"github.com/eytan-avisror/ttsum/pkg/tolerations"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	v1 "k8s.io/api/core/v1"
)

var (
	kubeconfigPath string
	namespace      string
	match          string
	noMatch        string
)

var tolerationsCmd = &cobra.Command{
	Use:   "tolerations [apiVersion kind] --namespace <namespace>",
	Short: "tolerations summarizes tolerations for a resource",
	Long:  "For example; $ ttsum tolerations apps/v1 deployment --namespace kube-system",
	Run:   RunTolerationsCommand,
}

func RunTolerationsCommand(cmd *cobra.Command, args []string) {
	if len(args) != 2 {
		log.Fatal("must provide group/resource e.g. ttsum tolerations apps/v1 deployments")
	}

	if match != "" && noMatch != "" {
		log.Fatal("--match and --no-match are mutually exclusive arguments")
	}

	gvr := resources.Parse(args[0], args[1])

	k8s, err := getKubernetesClient(kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	resourceTolerations, err := resources.ListResourceTolerations(k8s, gvr, namespace)
	if err != nil {
		log.Fatal(err)
	}

	if match != "" {
		expr, err := tolerations.Parse(match)
		if err != nil {
			log.Fatal(err)
		}

		resourceTolerations = resources.FilterTolerations(resourceTolerations, expr, true)
	} else if noMatch != "" {
		expr, err := tolerations.Parse(noMatch)
		if err != nil {
			log.Fatal(err)
		}

		resourceTolerations = resources.FilterTolerations(resourceTolerations, expr, false)
	}

	results := make([]TolerationsResult, 0)
	for resource, rawTolerations := range resourceTolerations {
		results = append(results, TolerationsResult{
			ResourceReference: resources.ResourceReference{
				Name:      resource.Name,
				Namespace: resource.Namespace,
			},
			Tolerations: rawTolerations,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Namespace < results[j].Namespace
	})

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"NAMESPACE", "NAME", "TOLERATIONS"})
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
		data = append(data, []string{result.Namespace, result.Name, tolerations.PrintPretty(result.Tolerations)})
	}

	table.AppendBulk(data)
	table.Render()
}

type TolerationsResult struct {
	resources.ResourceReference
	Tolerations []v1.Toleration
}

func init() {
	rootCmd.AddCommand(tolerationsCmd)
	tolerationsCmd.Flags().StringVar(&kubeconfigPath, "kubeconfig", "", "Path to kubeconfig")
	tolerationsCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Target a specific namespaces, defaults to all namespaces")
	tolerationsCmd.Flags().StringVar(&match, "match", "", "Show resources with toleration match, must be in format Operator(key=value:effect)")
	tolerationsCmd.Flags().StringVar(&noMatch, "no-match", "", "Show resources without toleration match, must be in format Operator(key=value:effect)")
}
