// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"slices"
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

	from, err := routedTo(ip)
	if err != nil {
		// The move itself does not depend on this read, so it is not a reason
		// to refuse. But the prompt must not imply the IP is free when what
		// actually happened is that nobody knows.
		log.Printf("%s", err)
	}
	if !common.ConfirmAction(common.Disruptive, ip, moveWarning(ip, from, chosen, err != nil)) {
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

	if err := waitForRouting(ip, chosen.Service, taskIDOf(task)); err != nil {
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

	from, err := routedTo(ip)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}
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

	// POST /ip/{ip}/park answers with an ip.IpTask, which this used to discard.
	// Keeping it is what lets the wait below notice a park that failed instead
	// of polling the state for ten minutes.
	var task map[string]any
	if err := httpLib.Client.Post(endpoint, nil, &task); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to park %s: %s", ip, err)
		return
	}

	log.Printf("⚡️ %s is being parked…", ip)

	if !IPWait {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ip, "task": task["taskId"]}, "⚡️ %s is being parked.", ip)
		return
	}

	if err := waitForRouting(ip, "", taskIDOf(task)); err != nil {
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
func waitForRouting(ip, want string, taskID int64) error {
	for attempt := 0; attempt < movePollAttempts; attempt++ {
		// A read that failed is not a read that came back empty. Parking waits
		// for exactly the empty string, so counting a failure as one would end
		// the wait on the answer it was looking for and report the park done.
		current, err := routedTo(ip)
		if err == nil && current == want {
			return nil
		}
		if err != nil {
			log.Printf("%s", err)
		}

		// The task is asked second and only about failure, which is the one
		// question the state cannot answer. A task that ends in cancelled,
		// customerError or ovhError leaves the IP where it was, so the state
		// reads "not there yet" forever: the wait ran its full ten minutes and
		// then said "not routed yet, follow it with ovhcloud ip tasks", which
		// reads as still in progress. Ten minutes and a misleading sentence for
		// something the API knew in seconds.
		//
		// It stays a complement and never a substitute — a vRack attach measured
		// earlier in this CLI created two tasks for one request, so a wait that
		// concluded from the task alone could report success while the other half
		// was still running. The state is read first every round, and read once
		// more below before a failure is announced.
		if failed, status, comment := taskFailed(ip, taskID); failed {
			if current, err := routedTo(ip); err == nil && current == want {
				return nil
			}
			return fmt.Errorf("the %s of %s failed: task %d ended %s%s\n   See it with: ovhcloud ip tasks %s --id %d",
				operationName(want), ip, taskID, status, taskComment(comment), ip, taskID)
		}

		time.Sleep(movePollInterval)
	}

	return fmt.Errorf("stopped waiting after %s; %s is not %s yet, follow it with: ovhcloud ip tasks %s",
		time.Duration(movePollAttempts)*movePollInterval, ip, destinationLabel(want), ip)
}

// terminalTaskFailures are the three ip.TaskStatusEnum values that mean the work
// stopped and will not resume.
//
// Retyped here because the schema lists the seven statuses without saying which
// of them are failures — that classification does not exist in the document. A
// guard test holds these three against the enum, so a rename shows up as a red
// test instead of quietly turning this branch off.
var terminalTaskFailures = []string{"cancelled", "customerError", "ovhError"}

// taskFailed asks whether the task handed to us has stopped for good.
//
// A task that cannot be read is not a task that failed: this is a complement to
// the state, and turning a transient 500 on the task route into "your move
// failed" would be worse than the ten-minute wait it replaces.
func taskFailed(ip string, taskID int64) (bool, string, string) {
	if taskID == 0 {
		return false, "", ""
	}

	var task struct {
		Status  string `json:"status"`
		Comment string `json:"comment"`
	}
	endpoint := fmt.Sprintf("/v1/ip/%s/task/%d", url.PathEscape(ip), taskID)
	if err := httpLib.Client.Get(endpoint, &task); err != nil {
		log.Printf("failed to read task %d of %s: %s", taskID, ip, err)
		return false, "", ""
	}

	return slices.Contains(terminalTaskFailures, task.Status), task.Status, task.Comment
}

func taskComment(comment string) string {
	if comment == "" {
		return ""
	}
	return " — " + comment
}

// operationName and destinationLabel keep the two messages above readable in
// both directions: this loop serves a move and a park, and each has to name what
// it was doing.
func operationName(want string) string {
	if want == "" {
		return "parking"
	}
	return "move"
}

func destinationLabel(want string) string {
	if want == "" {
		return "parked"
	}
	return "routed to " + want
}

// taskIDOf reads the identifier out of the task an operation returns.
//
// go-ovh decodes with UseNumber, so a JSON integer arrives as json.Number and
// not as float64 — a type switch that only handles float64 is dead code, which
// is a mistake this repository has already made twice.
func taskIDOf(task map[string]any) int64 {
	switch id := task["taskId"].(type) {
	case json.Number:
		if n, err := id.Int64(); err == nil {
			return n
		}
	case float64:
		return int64(id)
	case int64:
		return id
	}
	return 0
}

// routedTo returns the service this IP currently serves, and an empty string
// when it serves none.
//
// The error is returned rather than folded into that empty string, because the
// two mean opposite things and every caller here acts on the difference: "this
// IP is free" is a green light, "I could not find out" is not. Folding them
// made `park` answer "nothing to park" on a failed read, and made the wait
// loop below — whose target for a park IS the empty string — call the park a
// success because the check had failed.
func routedTo(ip string) (string, error) {
	var block struct {
		RoutedTo struct {
			ServiceName string `json:"serviceName"`
		} `json:"routedTo"`
	}

	if err := httpLib.Client.Get(fmt.Sprintf("/v1/ip/%s", url.PathEscape(ip)), &block); err != nil {
		return "", fmt.Errorf("failed to read what %s is routed to: %w", ip, err)
	}

	return block.RoutedTo.ServiceName, nil
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
func moveWarning(ip, from string, to destination, unknown bool) string {
	if unknown {
		return fmt.Sprintf("Could not read what %s currently serves; moving it to %s (%s) will stop whatever traffic it carries.",
			ip, to.Service, to.Family)
	}

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
