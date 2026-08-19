// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	_ "embed"
	"fmt"
	"log"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/ovh/ovhcloud-cli/internal/display"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/ovh/ovhcloud-cli/internal/services/common"
	"github.com/spf13/cobra"
)

// An additional IP is bought to be moved. That is what makes it additional: it
// follows a service rather than a machine, so a failover survives the server
// under it. The CLI could list those IPs and show what each was routed to, and
// could not move a single one.
//
// The move is guarded because it is the traffic of the service it currently
// serves that stops. It is validated first because the API takes a service
// name and the account has eighty of them: a typo there is a 400 several
// seconds later, when the CLI already holds the list of the ones that work.

//go:embed templates/destinations.tmpl
var destinationsTemplate string

var (
	// MoveNexthop selects a next hop when the destination offers several.
	MoveNexthop string

	// IPWait follows the move until the IP is actually routed.
	IPWait bool
)

// movePollInterval and movePollAttempts bound how long --wait follows a move.
//
// They are variables rather than constants so a test can exercise the timeout
// in milliseconds instead of in ten minutes.
var (
	movePollInterval = 5 * time.Second
	movePollAttempts = 120
)

// destination is one place an IP can go.
type destination struct {
	Family  string   `json:"family"`
	Service string   `json:"service"`
	Nexthop []string `json:"nexthop"`
}

// destinationsOf reads where an IP may be routed.
//
// The API answers an object keyed by service family — vps, dedicatedServer,
// cloudProject and four more — each holding its own list. Families that accept
// nothing come back empty rather than absent, so they are dropped here: a table
// with four empty sections answers the question worse than one with three full
// ones.
func destinationsOf(ip string) ([]destination, error) {
	var families map[string][]struct {
		Service string   `json:"service"`
		Nexthop []string `json:"nexthop"`
	}

	path := fmt.Sprintf("/v1/ip/%s/move", url.PathEscape(ip))
	if err := httpLib.Client.Get(path, &families); err != nil {
		return nil, fmt.Errorf("failed to read where %s can be moved: %w", ip, err)
	}

	var destinations []destination
	for family, services := range families {
		for _, service := range services {
			destinations = append(destinations, destination{
				Family:  family,
				Service: service.Service,
				Nexthop: service.Nexthop,
			})
		}
	}

	sortDestinations(destinations)

	return destinations, nil
}

// sortDestinations groups the answer by family, then by service.
//
// The API returns a map, so its order is whatever Go's iteration gives that
// run: without this the same command printed a hundred rows in a different
// order every time, which makes two readings impossible to compare.
func sortDestinations(destinations []destination) {
	sort.Slice(destinations, func(i, j int) bool {
		if destinations[i].Family != destinations[j].Family {
			return destinations[i].Family < destinations[j].Family
		}
		return destinations[i].Service < destinations[j].Service
	})
}

// ListIpDestinations shows where an IP can be moved.
func ListIpDestinations(_ *cobra.Command, args []string) {
	ip := args[0]

	destinations, err := destinationsOf(ip)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(destinations) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ip, "destinations": 0},
			"%s cannot be moved to any service.", ip)
		return
	}

	rows := make([]map[string]any, 0, len(destinations))
	families := map[string]int{}
	for _, d := range destinations {
		families[d.Family]++
		rows = append(rows, map[string]any{
			"family":  d.Family,
			"service": d.Service,
			"nexthop": strings.Join(d.Nexthop, ", "),
		})
	}

	summary := make([]string, 0, len(families))
	for _, family := range sortedKeys(families) {
		summary = append(summary, fmt.Sprintf("%d %s", families[family], family))
	}

	display.OutputObject(map[string]any{
		"summary":      strings.Join(summary, " · "),
		"destinations": rows,
	}, ip, destinationsTemplate, &flags.OutputFormatConfig)
}

// MoveIp routes an IP to another service.
func MoveIp(_ *cobra.Command, args []string) {
	ip, target := args[0], args[1]

	// The destination is checked here rather than by the API. The CLI is about
	// to ask for this list anyway to know the next hops, and a service name
	// that does not accept this IP comes back as a 400 several seconds later —
	// with no hint of which names would have worked.
	destinations, err := destinationsOf(ip)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	chosen, ok := pickDestination(destinations, target)
	if !ok {
		display.OutputError(&flags.OutputFormatConfig, "%s", unknownDestination(ip, target, destinations))
		return
	}

	from := routedTo(ip)
	if !common.ConfirmAction(common.Disruptive, ip, moveWarning(ip, from, chosen)) {
		display.OutputError(&flags.OutputFormatConfig, "move of %s cancelled", ip)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/move", url.PathEscape(ip))
	// The destination travels in the endpoint string here. #242 adds a Detail
	// field for exactly this and is not in this branch's ancestry; when both
	// land, this belongs there so that -o json keeps endpoint a path.
	if common.ReportDryRun(common.Call{
		Method:   "POST",
		Endpoint: endpoint + "  (to " + chosen.Service + ")",
	}) {
		return
	}

	payload := map[string]any{"to": chosen.Service}
	if MoveNexthop != "" {
		payload["nexthop"] = MoveNexthop
	}

	var task map[string]any
	if err := httpLib.Client.Post(endpoint, payload, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to move %s to %s: %s", ip, chosen.Service, err)
		return
	}

	log.Printf("⚡️ %s is moving to %s…", ip, chosen.Service)

	if !IPWait {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ip, "to": chosen.Service, "task": task["taskId"]},
			"⚡️ %s is being moved to %s. Follow it with: ovhcloud ip tasks %s", ip, chosen.Service, ip)
		return
	}

	if err := waitForRouting(ip, chosen.Service); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": ip, "routedTo": chosen.Service},
		"✅ %s is now routed to %s.", ip, chosen.Service)
}

// ParkIp detaches an IP from whatever it currently serves.
func ParkIp(_ *cobra.Command, args []string) {
	ip := args[0]

	from := routedTo(ip)
	if from == "" {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ip, "routedTo": nil},
			"%s is not routed to any service, so there is nothing to park.", ip)
		return
	}

	if !common.ConfirmAction(common.Disruptive, ip, fmt.Sprintf(
		"Parking %s stops the traffic it carries for %s, and routes it nowhere.", ip, from)) {
		display.OutputError(&flags.OutputFormatConfig, "parking of %s cancelled", ip)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/park", url.PathEscape(ip))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to park %s: %s", ip, err)
		return
	}

	log.Printf("⚡️ %s is being parked…", ip)

	if !IPWait {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ip}, "⚡️ %s is being parked.", ip)
		return
	}

	if err := waitForRouting(ip, ""); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": ip, "routedTo": nil}, "✅ %s is parked.", ip)
}

// ListIpTasks lists the tasks of an IP.
func ListIpTasks(_ *cobra.Command, args []string) {
	common.ManageListRequest(fmt.Sprintf("/v1/ip/%s/task", url.PathEscape(args[0])), "",
		[]string{"taskId", "function", "status", "startDate", "doneDate", "comment"},
		flags.GenericFilters)
}

// waitForRouting follows a move by reading where the IP is routed, not by
// reading the task.
//
// The task says the work finished; only the IP says where it landed. A vRack
// attach measured earlier in this CLI created two tasks for one request, so a
// wait that trusts the task it was handed can report success while the other
// half is still running. Reading the state costs one call and cannot be wrong
// about the thing it is asked.
//
// An empty want means parked.
func waitForRouting(ip, want string) error {
	for attempt := 0; attempt < movePollAttempts; attempt++ {
		if current := routedTo(ip); current == want {
			return nil
		}

		time.Sleep(movePollInterval)
	}

	where := "parked"
	if want != "" {
		where = "routed to " + want
	}

	return fmt.Errorf("stopped waiting after %s; %s is not %s yet, follow it with: ovhcloud ip tasks %s",
		time.Duration(movePollAttempts)*movePollInterval, ip, where, ip)
}

// routedTo answers the service an IP currently serves, and an empty string when
// it serves none.
//
// A failure to read is reported as "no service" on purpose: this feeds a
// confirmation prompt and a wait, and neither is improved by turning a
// transient read error into a failed command. The prompt says "not routed"
// rather than naming a service it could not confirm.
func routedTo(ip string) string {
	var block struct {
		RoutedTo struct {
			ServiceName string `json:"serviceName"`
		} `json:"routedTo"`
	}

	if err := httpLib.Client.Get(fmt.Sprintf("/v1/ip/%s", url.PathEscape(ip)), &block); err != nil {
		return ""
	}

	return block.RoutedTo.ServiceName
}

// pickDestination finds the requested service among the ones that accept this IP.
func pickDestination(destinations []destination, target string) (destination, bool) {
	for _, candidate := range destinations {
		if strings.EqualFold(candidate.Service, target) {
			return candidate, true
		}
	}

	return destination{}, false
}

// unknownDestination refuses a service this IP cannot reach, and names the ones
// it can.
//
// The full list is not printed: an account measured while writing this offered
// a hundred destinations for one address. The families are named with their
// counts, and the command that prints them is given.
func unknownDestination(ip, target string, destinations []destination) error {
	families := map[string]int{}
	for _, d := range destinations {
		families[d.Family]++
	}

	summary := make([]string, 0, len(families))
	for _, family := range sortedKeys(families) {
		summary = append(summary, fmt.Sprintf("%d %s", families[family], family))
	}

	return fmt.Errorf("%s does not accept %s.\n   It can be moved to %s — list them with: ovhcloud ip destinations %s",
		target, ip, strings.Join(summary, ", "), ip)
}

// moveWarning is the sentence somebody reads before the traffic stops. It is
// built here rather than inline so a test can read the wording that stands
// between an operator and a service going dark.
func moveWarning(ip, from string, to destination) string {
	if from == "" {
		return fmt.Sprintf("%s is not routed to any service; moving it will route it to %s (%s).",
			ip, to.Service, to.Family)
	}

	return fmt.Sprintf("Moving %s to %s (%s) stops the traffic it carries for %s.",
		ip, to.Service, to.Family, from)
}

func sortedKeys(counts map[string]int) []string {
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
