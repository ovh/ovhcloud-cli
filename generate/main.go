package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"text/template"
)

type cfgEntry struct {
	productName      string
	columnsToDisplay []string
}

var config = map[string]*cfgEntry{
	"/allDom": {
		productName:      "allDom",
		columnsToDisplay: []string{"name", "type", "offer"},
	},
	"/cdn/dedicated": {
		productName:      "cdnDedicated",
		columnsToDisplay: []string{"service", "offer", "anycast"},
	},
	// "/cloud/project": {
	// 	productName:      "cloudProject",
	// 	columnsToDisplay: []string{"project_id", "projectName", "status", "description"},
	// },
	"/dbaas/logs": {
		productName:      "ldp",
		columnsToDisplay: []string{"serviceName", "displayName", "isClusterOwner", "state", "username"},
	},
	"/dedicated/ceph": {
		productName:      "dedicatedCeph",
		columnsToDisplay: []string{"serviceName", "region", "state", "status"},
	},
	"/dedicated/cluster": {
		productName:      "dedicatedCluster",
		columnsToDisplay: []string{"id", "region", "model", "status"},
	},
	"/dedicated/nasha": {
		productName:      "dedicatedNasHA",
		columnsToDisplay: []string{"serviceName", "customName", "datacenter"},
	},
	"/dedicated/server": {
		productName:      "baremetal",
		columnsToDisplay: []string{"name", "region", "os", "powerState", "state"},
	},
	"/dedicatedCloud": {
		productName:      "dedicatedCloud",
		columnsToDisplay: []string{"serviceName", "location", "state", "description"},
	},
	"/domain": {
		productName:      "domainName",
		columnsToDisplay: []string{"domain", "state", "whoisOwner", "expirationDate", "renewalDate"},
	},
	"/domain/zone": {
		productName:      "domainZone",
		columnsToDisplay: []string{"name", "dnssecSupported", "hasDnsAnycast"},
	},
	"/email/domain": {
		productName:      "emailDomain",
		columnsToDisplay: []string{"domain", "status", "offer"},
	},
	// "/email/exchange"
	"/email/mxplan": {
		productName:      "emailMXPlan",
		columnsToDisplay: []string{"domain", "displayName", "state", "offer"},
	},
	"/email/pro": {
		productName:      "emailPro",
		columnsToDisplay: []string{"domain", "displayName", "state", "offer"},
	},
	"/hosting/privateDatabase": {
		productName:      "hostingPrivateDatabase",
		columnsToDisplay: []string{"serviceName", "displayName", "type", "version", "state"},
	},
	"/hosting/web": {
		productName:      "webHosting",
		columnsToDisplay: []string{"serviceName", "displayName", "datacenter", "state"},
	},
	"/ip": {
		productName:      "ip",
		columnsToDisplay: []string{"ip", "rir", "routedTo", "country", "description"},
	},
	"/ipLoadbalancing": {
		productName:      "ipLoadbalancing",
		columnsToDisplay: []string{"serviceName", "displayName", "zone", "state"},
	},
	// "/license/cloudLinux"
	// "/license/cpanel"
	// "/license/directadmin"
	// "/license/hycu"
	// "/license/office"
	// "/license/officePrepaid"
	// "/license/plesk"
	// "/license/redhat"
	// "/license/sqlserver"
	// "/license/virtuozzo"
	// "/license/windows"
	// "/license/worklight"
	// "/me"
	// "/msServices"
	"/nutanix": {
		productName:      "nutanix",
		columnsToDisplay: []string{"serviceName", "status"},
	},
	"/overTheBox": {
		productName:      "overTheBox",
		columnsToDisplay: []string{"serviceName", "offer", "status", "bandwidth"},
	},
	"/ovhCloudConnect": {
		productName:      "ovhCloudConnect",
		columnsToDisplay: []string{"uuid", "provider", "status", "description"},
	},
	"/pack/xdsl": {
		productName:      "packXDSL",
		columnsToDisplay: []string{"packName", "description"},
	},
	// "/products"
	// "/saas/csp2"
	// "/service"
	// "/services"
	"/sms": {
		productName:      "sms",
		columnsToDisplay: []string{"name", "status"},
	},
	"/ssl": {
		productName:      "ssl",
		columnsToDisplay: []string{"serviceName", "type", "authority", "status"},
	},
	"/sslGateway": {
		productName:      "sslGateway",
		columnsToDisplay: []string{"serviceName", "displayName", "state", "zones"},
	},
	"/storage/netapp": {
		productName:      "storageNetApp",
		columnsToDisplay: []string{"id", "name", "region", "status"},
	},
	"/support/tickets": {
		productName:      "supportTickets",
		columnsToDisplay: []string{"ticketId", "serviceName", "type", "category", "state"},
	},
	"/telephony": {
		productName:      "telephony",
		columnsToDisplay: []string{"billingAccount", "description", "status"},
	},
	"/veeam/veeamEnterprise": {
		productName:      "veeamEnterprise",
		columnsToDisplay: []string{"serviceName", "activationStatus", "ip", "sourceIp"},
	},
	"/veeamCloudConnect": {
		productName:      "veeamCloudConnect",
		columnsToDisplay: []string{"serviceName", "productOffer", "location", "vmCount"},
	},
	"/vps": {
		productName:      "vps",
		columnsToDisplay: []string{"name", "displayName", "state", "zone"},
	},
	"/vrack": {
		productName: "vrack",
		// TODO: service name not returned in response body, to fix
		columnsToDisplay: []string{"name", "description"},
	},
	"/xdsl": {
		productName:      "xdsl",
		columnsToDisplay: []string{"accessName", "accessType", "provider", "role", "status"},
	},
	// "/v2/domain/name"
	// "/v2/iam"
	"/v2/location": {
		productName:      "location",
		columnsToDisplay: []string{"name", "type", "specificType", "location"},
	},
	// "/v2/networkDefense"
	"/v2/okms/resource": {
		productName:      "okms",
		columnsToDisplay: []string{"id", "region"},
	},
	// "/v2/publicCloud"
	"/v2/vmwareCloudDirector/organization": {
		productName:      "vmwareCloudDirectorOrganization",
		columnsToDisplay: []string{"id", "currentState.fullName", "currentState.region", "resourceStatus"},
	},
	"/v2/vmwareCloudDirector/backup": {
		productName:      "vmwareCloudDirectorBackup",
		columnsToDisplay: []string{"id", "iam.displayName", "currentState.azName", "resourceStatus"},
	},
	"/v2/vrackServices/resource": {
		productName:      "vrackServices",
		columnsToDisplay: []string{"id", "currentState.region", "currentState. productStatus", "resourceStatus"},
	},
	// "/v2/webhosting"
	// "/v2/zimbra"
}

var templ = `package cmd

import (
	"github.com/spf13/cobra"
)

var (
	{{.ProductNameLower}}ColumnsToDisplay = []string{ {{.ColumnsStr}} }
)

func list{{.ProductName}}(_ *cobra.Command, _ []string) {
	manageListRequest("{{.Path}}", {{.ProductNameLower}}ColumnsToDisplay, genericFilters)
}

func get{{.ProductName}}(_ *cobra.Command, args []string) {
	manageObjectRequest("{{.Path}}", args[0], {{.ProductNameLower}}ColumnsToDisplay[0])
}

func init() {
	{{.ProductNameLower}}Cmd := &cobra.Command{
		Use:   "{{.ProductNameLower}}",
		Short: "Retrieve information and manage your {{.ProductName}} services",
	}

	// Command to list {{.ProductName}} services
	{{.ProductNameLower}}ListCmd := &cobra.Command{
		Use:   "list",
		Short: "List your {{.ProductName}} services",
		Run:   list{{.ProductName}},
	}
	{{.ProductNameLower}}ListCmd.PersistentFlags().StringArrayVar(
		&genericFilters,
		"filter",
		nil,
		"Filter results by any property using github.com/PaesslerAG/gval syntax'",
	)
	{{.ProductNameLower}}Cmd.AddCommand({{.ProductNameLower}}ListCmd)

	// Command to get a single {{.ProductName}}
	{{.ProductNameLower}}Cmd.AddCommand(&cobra.Command{
		Use:        "get",
		Short:      "Retrieve information of a specific {{.ProductName}}",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"service_name"},
		Run:        get{{.ProductName}},
	})

	rootCmd.AddCommand({{.ProductNameLower}}Cmd)
}
`

func main() {
	t := template.Must(template.New("section").Parse(templ))

	for path, cfg := range config {

		formattedCols := make([]string, 0, len(cfg.columnsToDisplay))
		for _, col := range cfg.columnsToDisplay {
			formattedCols = append(formattedCols, fmt.Sprintf("%q", col))
		}

		f, err := os.Create("../cmd/" + strings.ToLower(cfg.productName) + ".go")
		if err != nil {
			log.Fatalf("failed to open output file: %s", err)
		}
		defer f.Close()

		err = t.Execute(f, map[string]any{
			"Path":             path,
			"ProductName":      strings.Title(cfg.productName),
			"ProductNameLower": strings.ToLower(cfg.productName),
			"ColumnsStr":       strings.Join(formattedCols, ","),
		})
		if err != nil {
			log.Fatalf("execution failed: %s", err)
		}
	}
}
