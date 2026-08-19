// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	_ "embed"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/ovh/ovhcloud-cli/internal/assets"
	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/openapi"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

//go:embed templates/ripe.tmpl
var ripeTemplate string

//go:embed templates/migration_token.tmpl
var migrationTokenTemplate string

const confirmTerminationPath = "/ip/service/{serviceName}/confirmTermination"

var (
	// RipeNetname and RipeDescription are what `ripe set` writes.
	RipeNetname     string
	RipeDescription string

	// MigrationCustomerId names who may claim the IP.
	MigrationCustomerId string

	// RevealMigrationToken prints the token instead of its fingerprint.
	RevealMigrationToken bool

	// ByoipSlicingSize and ByoipAggregationIp pick one of the configurations
	// the preview routes offer.
	ByoipSlicingSize   int
	ByoipAggregationIp string

	// Contacts of `service change-contact`.
	ContactAdmin   string
	ContactBilling string
	ContactTech    string

	// Survey fields of `service confirm-termination`.
	TerminationReason    string
	TerminationFutureUse string
	TerminationComment   string
)

// The accepted survey values are read from the specification embedded in this
// binary rather than transcribed: `reason` carries fourteen values today, and a
// list copied into Go drifts the day the API gains a fifteenth — silently, into
// a 400 nobody can explain. Behind sync.OnceValues because every invocation of
// the CLI registers the flags and only a completion or a termination needs the
// values.
var (
	terminationReasons = sync.OnceValues(func() ([]string, error) {
		return openapi.GetRequestFieldEnum(assets.IpOpenapiSchema, confirmTerminationPath, "post", "reason")
	})
	terminationFutureUses = sync.OnceValues(func() ([]string, error) {
		return openapi.GetRequestFieldEnum(assets.IpOpenapiSchema, confirmTerminationPath, "post", "futureUse")
	})
)

func CompleteTerminationReason(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return common.CompleteEnum(terminationReasons)
}

func CompleteTerminationFutureUse(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
	return common.CompleteEnum(terminationFutureUses)
}

// ListCampus shows where IPs can live, and which registries each site accepts
// for a bring-your-own-IP announcement.
//
// It is the only route of this domain that answers something about OVHcloud
// rather than about the account, and nothing in the CLI referenced it. It is
// what somebody reads before buying: `bringYourOwnIpSupportedRirForIp` says
// whether their own prefix can be announced there at all.
func ListCampus(_ *cobra.Command, _ []string) {
	var campuses []map[string]any
	if err := httpLib.Client.Get("/v1/ip/campus", &campuses); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to read the IP campuses: %s", err)
		return
	}

	rows := make([]map[string]any, 0, len(campuses))
	for _, campus := range campuses {
		row := map[string]any{}
		for key, value := range campus {
			row[key] = value
		}
		row["datacentersReadable"] = strings.Join(sortedCopy(stringSlice(campus["datacenters"])), ", ")
		row["rirReadable"] = strings.Join(sortedCopy(stringSlice(campus["bringYourOwnIpSupportedRirForIp"])), ", ")
		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return stringField(rows[i], "name") < stringField(rows[j], "name")
	})

	renderFiltered(rows, []string{"name", "description",
		"rirReadable byoip", "datacentersReadable datacenters"})
}

// ListIpServices lists the billed IP services.
//
// `ip list` and `ip service list` are not two views of one thing: the account
// measured for this holds 537 blocks and 80 services. A block is what gets
// routed; a service is what gets renewed, has contacts and can be terminated.
func ListIpServices(_ *cobra.Command, _ []string) {
	common.ManageListRequest("/v1/ip/service", "",
		[]string{"ip", "type", "country", "routedTo.serviceName", "canBeTerminated"},
		flags.GenericFilters)
}

func GetIpService(_ *cobra.Command, args []string) {
	common.ManageObjectRequest("/v1/ip/service", args[0], "")
}

func EditIpService(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/ip/service/{serviceName}",
		fmt.Sprintf("/v1/ip/service/%s", url.PathEscape(args[0])),
		IPServiceSpec,
		assets.IpOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

func GetIpServiceInfo(_ *cobra.Command, args []string) {
	endpoint := fmt.Sprintf("/v1/ip/service/%s/serviceInfos", url.PathEscape(args[0]))

	var object map[string]any
	if err := httpLib.Client.Get(endpoint, &object); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"error fetching service information for %s: %s", args[0], err)
		return
	}

	display.OutputObject(object, args[0], common.ServiceInfoTemplate, &flags.OutputFormatConfig)
}

func EditIpServiceInfo(cmd *cobra.Command, args []string) {
	if err := common.EditResource(
		cmd,
		"/ip/service/{serviceName}/serviceInfos",
		fmt.Sprintf("/v1/ip/service/%s/serviceInfos", url.PathEscape(args[0])),
		common.ServiceInfoRenewPayload(cmd),
		assets.IpOpenapiSchema,
	); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
}

// ChangeIpServiceContact starts a contact change procedure.
func ChangeIpServiceContact(_ *cobra.Command, args []string) {
	service := args[0]

	body := map[string]any{}
	for name, value := range map[string]string{
		"contactAdmin":   ContactAdmin,
		"contactBilling": ContactBilling,
		"contactTech":    ContactTech,
	} {
		if value != "" {
			body[name] = value
		}
	}

	// The API accepts a body with no contact in it and answers an empty list
	// of tasks: a command that appears to have worked and changed nothing.
	if len(body) == 0 {
		display.OutputError(&flags.OutputFormatConfig,
			"no contact given; name at least one of --admin, --billing or --tech")
		return
	}

	if !common.ConfirmAction(common.Disruptive, service, fmt.Sprintf(
		"Changing a contact on %s starts a procedure each new contact has to accept by email.", service)) {
		display.OutputError(&flags.OutputFormatConfig, "contact change on %s cancelled", service)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/service/%s/changeContact", url.PathEscape(service))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var tasks []int64
	if err := httpLib.Client.Post(endpoint, body, &tasks); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the contacts of %s: %s", service, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"service": service, "tasks": tasks},
		"⚡️ Contact change started on %s (%d procedure(s) to accept by email).", service, len(tasks))
}

// TerminateIpService asks for the termination of an IP service. It stops
// nothing: the API emails a token, and the service runs until it is confirmed.
func TerminateIpService(_ *cobra.Command, args []string) {
	service := args[0]
	endpoint := fmt.Sprintf("/v1/ip/service/%s/terminate", url.PathEscape(service))

	if !common.ConfirmAction(common.Disruptive, service, fmt.Sprintf(
		"This asks for the termination of %s. Nothing stops now: a termination token is emailed to the administrative contact, and the IP keeps serving until that token is confirmed.",
		service)) {
		display.OutputError(&flags.OutputFormatConfig, "termination request for %s cancelled", service)
		return
	}

	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var response string
	if err := httpLib.Client.Post(endpoint, nil, &response); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"error requesting termination of %s: %s", service, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"service": service, "response": response},
		"⚡️ Termination of %s requested. A token has been emailed to the administrative contact; confirm with:\n  ovhcloud ip service confirm-termination %s <token>",
		service, service)
}

// ConfirmIpServiceTermination is the irreversible half.
func ConfirmIpServiceTermination(_ *cobra.Command, args []string) {
	service := args[0]

	if err := common.CheckEnumFlag("reason", TerminationReason, terminationReasons); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
	if err := common.CheckEnumFlag("future-use", TerminationFutureUse, terminationFutureUses); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// cobra counts the arguments, it does not look at them: an empty second
	// argument satisfies ExactArgs(2) and would travel all the way to a 400.
	token := strings.TrimSpace(args[1])
	if token == "" {
		display.OutputError(&flags.OutputFormatConfig,
			"no termination token given; it is the token emailed to the administrative contact by `ip service terminate %s`", service)
		return
	}

	if !common.ConfirmAction(common.Destructive, service, fmt.Sprintf(
		"This confirms the termination of %s. The IP is returned to OVHcloud at expiry; there is no undoing it from here.", service)) {
		display.OutputError(&flags.OutputFormatConfig, "termination of %s not confirmed", service)
		return
	}

	body := map[string]any{"token": token}
	if TerminationReason != "" {
		body["reason"] = TerminationReason
	}
	if TerminationFutureUse != "" {
		body["futureUse"] = TerminationFutureUse
	}
	if TerminationComment != "" {
		body["commentary"] = TerminationComment
	}

	endpoint := fmt.Sprintf("/v1/ip/service/%s/confirmTermination", url.PathEscape(service))
	if common.ReportDryRun(common.Call{
		Method:   "POST",
		Endpoint: fmt.Sprintf("%s  (%s)", endpoint, common.DescribeTerminationBody(body)),
	}) {
		return
	}

	if err := httpLib.Client.Post(endpoint, body, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"error confirming termination of %s: %s", service, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, map[string]any{"service": service},
		"✅ Termination of %s confirmed", service)
}

// ListDelegation lists the servers a reverse delegation points at.
func ListDelegation(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	var targets []string
	path := fmt.Sprintf("/v1/ip/%s/delegation", url.PathEscape(ipBlock))
	if err := httpLib.Client.Get(path, &targets); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", delegationError(ipBlock, err))
		return
	}

	if len(targets) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "delegation": []any{}},
			"No reverse delegation is set on %s.", ipBlock)
		return
	}

	rows := make([]map[string]any, 0, len(targets))
	for _, target := range targets {
		rows = append(rows, map[string]any{"target": target})
	}

	renderFiltered(rows, []string{"target"})
}

func GetDelegation(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(
		fmt.Sprintf("/v1/ip/%s/delegation", url.PathEscape(args[0])), args[1], "")
}

// AddDelegation points a reverse delegation at a name server.
func AddDelegation(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	if !common.ConfirmAction(common.Disruptive, target, fmt.Sprintf(
		"Delegating the reverse of %s to %s hands the reverse resolution of the whole subnet to that server.", ipBlock, target)) {
		display.OutputError(&flags.OutputFormatConfig, "delegation of %s cancelled", ipBlock)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/delegation", url.PathEscape(ipBlock))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Post(endpoint, map[string]string{"target": target}, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", delegationError(ipBlock, err))
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": ipBlock, "target": target},
		"✅ The reverse of %s is now delegated to %s.", ipBlock, target)
}

// RemoveDelegation takes a name server out of a reverse delegation.
func RemoveDelegation(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	if !common.ConfirmAction(common.Disruptive, target, fmt.Sprintf(
		"Removing %s from the reverse delegation of %s stops it answering for that subnet.", target, ipBlock)) {
		display.OutputError(&flags.OutputFormatConfig, "removal of %s cancelled", target)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/delegation/%s",
		url.PathEscape(ipBlock), url.PathEscape(target))
	if common.ReportDryRun(common.Call{Method: "DELETE", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Delete(endpoint, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", delegationError(ipBlock, err))
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": ipBlock, "target": target},
		"✅ %s no longer answers for the reverse of %s.", target, ipBlock)
}

// delegationError reports a delegation failure and, when the block is IPv4,
// adds what the route says about itself.
//
// The API summarises this route as "Reverse delegation on IPv6 subnet" and
// answers HTTP 500 for everything else — measured on 492 of 537 blocks, every
// IPv4 mask and every IPv6 /128, while the 45 IPv6 /56 and /64 answered 200.
// The note is attached to the error rather than replacing the call, because
// the scope is the route's documented one and not a rule this CLI should be
// enforcing on the API's behalf.
func delegationError(ipBlock string, err error) error {
	if strings.Contains(ipBlock, ":") {
		return fmt.Errorf("failed to read the reverse delegation of %s: %w", ipBlock, err)
	}

	return fmt.Errorf("failed to read the reverse delegation of %s: %w\n   This route covers reverse delegation on IPv6 subnets, and %s is IPv4.",
		ipBlock, err, ipBlock)
}

// ListLicenses answers, in one command, which licences are attached to an IP.
//
// The API has one route per product and no index, so the question "does this
// address carry a licence" is eight requests. Asking it eight times is the kind
// of thing a CLI exists to stop doing by hand.
func ListLicenses(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	rows := make([]map[string]any, 0, len(licenseProducts))
	for _, product := range licenseProducts {
		var licenses []string
		path := fmt.Sprintf("/v1/ip/%s/license/%s", url.PathEscape(ipBlock), product)
		if err := httpLib.Client.Get(path, &licenses); err != nil {
			display.OutputError(&flags.OutputFormatConfig,
				"failed to read the %s licences of %s: %s", product, ipBlock, err)
			return
		}

		for _, license := range licenses {
			rows = append(rows, map[string]any{"product": product, "license": license})
		}
	}

	if len(rows) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "licenses": []any{}},
			"No licence is attached to %s.", ipBlock)
		return
	}

	renderFiltered(rows, []string{"product", "license"})
}

// licenseProducts is the eight products with a licence route. The list is
// spelled out because the API has no index of them: it is the index.
var licenseProducts = []string{
	"cloudLinux", "cpanel", "directadmin", "plesk",
	"sqlserver", "virtuozzo", "windows", "worklight",
}

// GetRipe shows the RIPE record published for a block.
func GetRipe(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	record, err := ripeRecord(ipBlock)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputObject(record, ipBlock, ripeTemplate, &flags.OutputFormatConfig)
}

// SetRipe changes the RIPE record published for a block.
//
// The record is read first and sent back whole. The API takes the full object,
// so a request carrying only --description would publish an empty netname —
// a field that is visible in the public registry, and that nobody asked to
// clear.
func SetRipe(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	if RipeNetname == "" && RipeDescription == "" {
		display.OutputError(&flags.OutputFormatConfig,
			"nothing to change; give --netname, --description, or both")
		return
	}

	record, err := ripeRecord(ipBlock)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	payload := map[string]any{
		"netname":     stringField(record, "netname"),
		"description": stringField(record, "description"),
	}
	if RipeNetname != "" {
		payload["netname"] = RipeNetname
	}
	if RipeDescription != "" {
		payload["description"] = RipeDescription
	}

	if !common.ConfirmAction(common.Disruptive, ipBlock, fmt.Sprintf(
		"The RIPE record of %s becomes netname %q, description %q. It is published in the public registry.",
		ipBlock, payload["netname"], payload["description"])) {
		display.OutputError(&flags.OutputFormatConfig, "change on %s cancelled", ipBlock)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/ripe", url.PathEscape(ipBlock))
	if common.ReportDryRun(common.Call{Method: "PUT", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Put(endpoint, payload, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the RIPE record of %s: %s", ipBlock, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, payload,
		"✅ The RIPE record of %s was updated.", ipBlock)
}

func ripeRecord(ipBlock string) (map[string]any, error) {
	var record map[string]any

	path := fmt.Sprintf("/v1/ip/%s/ripe", url.PathEscape(ipBlock))
	if err := httpLib.Client.Get(path, &record); err != nil {
		return nil, fmt.Errorf("failed to read the RIPE record of %s: %w", ipBlock, err)
	}

	return record, nil
}

// GetMigrationToken shows the token that lets another account claim this IP.
//
// The token is withheld by default, and --reveal prints it. It is a bearer
// credential: whoever holds it, with the customer identifier printed beside it,
// can take the address. `get` is most often run to find out whether a migration
// is pending and for whom, and that question is answered without putting a
// live credential into a terminal buffer or a pipeline log.
func GetMigrationToken(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	var token map[string]any
	path := fmt.Sprintf("/v1/ip/%s/migrationToken", url.PathEscape(ipBlock))
	if err := httpLib.Client.Get(path, &token); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"no migration token exists for %s: %s\n   Create one with: ovhcloud ip migration-token create %s --customer-id <customer>",
			ipBlock, err, ipBlock)
		return
	}

	display.OutputObject(migrationTokenView(token), ipBlock, migrationTokenTemplate, &flags.OutputFormatConfig)
}

// CreateMigrationToken generates a token for another account to claim this IP.
func CreateMigrationToken(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	if MigrationCustomerId == "" {
		display.OutputError(&flags.OutputFormatConfig,
			"--customer-id is required; it names the account that will be able to claim %s", ipBlock)
		return
	}

	if !common.ConfirmAction(common.Disruptive, ipBlock, fmt.Sprintf(
		"This creates a token letting %s claim %s. Whoever holds the token can move the address to that account.",
		MigrationCustomerId, ipBlock)) {
		display.OutputError(&flags.OutputFormatConfig, "migration token for %s cancelled", ipBlock)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/migrationToken", url.PathEscape(ipBlock))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var token map[string]any
	if err := httpLib.Client.Post(endpoint,
		map[string]string{"customerId": MigrationCustomerId}, &token); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to create a migration token for %s: %s", ipBlock, err)
		return
	}

	display.OutputObject(migrationTokenView(token), ipBlock, migrationTokenTemplate, &flags.OutputFormatConfig)
}

// migrationTokenView replaces the token with its fingerprint unless --reveal.
//
// The substitution happens on the object rather than in the template, so
// -o json and a custom format are covered by the same decision as the table.
// A masking that only applies to the human-readable output protects nothing:
// the pipeline is where the value would have been logged.
func migrationTokenView(token map[string]any) map[string]any {
	view := map[string]any{}
	for key, value := range token {
		view[key] = value
	}

	if !RevealMigrationToken {
		view["token"] = common.Fingerprint(stringField(token, "token"))
		view["hidden"] = true
	}

	return view
}

// ChangeIpOrganisation changes the organisation an IP is registered to.
func ChangeIpOrganisation(_ *cobra.Command, args []string) {
	ipBlock, organisation := args[0], args[1]

	// Destructive rather than Disruptive: this rewrites who holds the address
	// in the regional registry. Nothing stops, and nothing is undone by
	// running the command again with the old value either — the previous
	// holder has to agree to take it back.
	if !common.ConfirmAction(common.Destructive, ipBlock, fmt.Sprintf(
		"This registers %s to the organisation %s. The change is published to the regional registry.",
		ipBlock, organisation)) {
		display.OutputError(&flags.OutputFormatConfig, "organisation change on %s cancelled", ipBlock)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/changeOrg", url.PathEscape(ipBlock))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint,
		map[string]string{"organisation": organisation}, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to change the organisation of %s: %s", ipBlock, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task,
		"⚡️ %s is being registered to %s. Follow it with: ovhcloud ip tasks %s",
		ipBlock, organisation, ipBlock)
}

// ListByoipAggregations shows how a bring-your-own-IP block could be merged
// with its neighbours.
func ListByoipAggregations(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	var previews []map[string]any
	path := fmt.Sprintf("/v1/ip/%s/bringYourOwnIp/aggregate", url.PathEscape(ipBlock))
	if err := httpLib.Client.Get(path, &previews); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", byoipError(ipBlock, err))
		return
	}

	if len(previews) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "aggregations": []any{}},
			"%s cannot be aggregated with any neighbour.", ipBlock)
		return
	}

	for _, preview := range previews {
		preview["childrenReadable"] = strings.Join(stringSlice(preview["childrenIps"]), ", ")
	}

	renderFiltered(previews, []string{"aggregationIp", "childrenReadable children"})
}

// ListByoipSlices shows how a bring-your-own-IP block could be split.
func ListByoipSlices(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	var previews []map[string]any
	path := fmt.Sprintf("/v1/ip/%s/bringYourOwnIp/slice", url.PathEscape(ipBlock))
	if err := httpLib.Client.Get(path, &previews); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", byoipError(ipBlock, err))
		return
	}

	if len(previews) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "slices": []any{}},
			"%s cannot be sliced.", ipBlock)
		return
	}

	for _, preview := range previews {
		preview["childrenReadable"] = strings.Join(stringSlice(preview["childrenIps"]), ", ")
	}

	renderFiltered(previews, []string{"slicingSize", "childrenReadable children"})
}

// AggregateByoip merges a block with its neighbours.
func AggregateByoip(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	if ByoipAggregationIp == "" {
		display.OutputError(&flags.OutputFormatConfig,
			"--into is required; list the possible parents with: ovhcloud ip byoip aggregations %s", ipBlock)
		return
	}

	if !common.ConfirmAction(common.Disruptive, ipBlock, fmt.Sprintf(
		"Aggregating %s into %s replaces it and its neighbours with a single block.", ipBlock, ByoipAggregationIp)) {
		display.OutputError(&flags.OutputFormatConfig, "aggregation of %s cancelled", ipBlock)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/bringYourOwnIp/aggregate", url.PathEscape(ipBlock))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint,
		map[string]string{"aggregationIp": ByoipAggregationIp}, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", byoipError(ipBlock, err))
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task,
		"⚡️ %s is being aggregated into %s. Follow it with: ovhcloud ip tasks %s",
		ipBlock, ByoipAggregationIp, ipBlock)
}

// SliceByoip splits a block into smaller ones.
func SliceByoip(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	if ByoipSlicingSize == 0 {
		display.OutputError(&flags.OutputFormatConfig,
			"--size is required; list the possible sizes with: ovhcloud ip byoip slices %s", ipBlock)
		return
	}

	if !common.ConfirmAction(common.Disruptive, ipBlock, fmt.Sprintf(
		"Slicing %s into /%d blocks replaces it with the smaller ones.", ipBlock, ByoipSlicingSize)) {
		display.OutputError(&flags.OutputFormatConfig, "slicing of %s cancelled", ipBlock)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/bringYourOwnIp/slice", url.PathEscape(ipBlock))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint,
		map[string]int{"slicingSize": ByoipSlicingSize}, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", byoipError(ipBlock, err))
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig, task,
		"⚡️ %s is being sliced into /%d blocks. Follow it with: ovhcloud ip tasks %s",
		ipBlock, ByoipSlicingSize, ipBlock)
}

// byoipError keeps the API's own words.
//
// Every block of the account measured answered "This IP is not part of the
// Bring you own IP product." — a business answer carried by an HTTP 400, and a
// better one than anything this CLI could add. It is passed through rather
// than replaced.
func byoipError(ipBlock string, err error) error {
	return fmt.Errorf("failed to read the bring-your-own-IP configuration of %s: %w", ipBlock, err)
}
