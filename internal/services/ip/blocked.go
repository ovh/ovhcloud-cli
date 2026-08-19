// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package ip

import (
	_ "embed"
	"encoding/json"
	"fmt"
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

//go:embed templates/phishing.tmpl
var phishingTemplate string

// Three separate mechanisms can block an address, each with its own list, its
// own detail route and its own unblock route. The operator's question during
// an incident is "is one of my addresses blocked, and can I release it", not
// "which of the three lists should I query first". So the three are read
// together, and the mechanism that holds an address becomes the argument the
// release needs.

const (
	antihackMechanism = "antihack"
	arpMechanism      = "arp"
	spamMechanism     = "spam"
)

// blockMechanisms is the order the three are read and printed in. It is fixed
// rather than alphabetical by accident: it happens to match, and a table whose
// row order changes between two runs cannot be compared with the previous one.
var blockMechanisms = []string{antihackMechanism, arpMechanism, spamMechanism}

var (
	// UnblockReason names the mechanism to release an address from, when the
	// CLI should not work it out itself.
	UnblockReason string

	// SpamStatsFrom and SpamStatsTo bound the window of the spam report.
	SpamStatsFrom string
	SpamStatsTo   string
)

// blockedAddress is one address held by one mechanism.
type blockedAddress struct {
	IP        string
	Mechanism string
	State     string
	Since     string

	// Seconds is the API's `time` field, and it does not mean the same thing
	// twice. On antihack and arp it is "time remaining before you can request
	// your IP to be unblocked"; on spam it is "time while the IP will be
	// blocked". One is a cooldown the operator waits out before acting, the
	// other is the sentence itself. They are kept in the same field because
	// the API puts them there, and turned into a sentence per mechanism by
	// blockNote — a single column holding both numbers under one header would
	// be wrong in a way nobody would catch.
	Seconds int64

	// Raw carries every field the API answered, so -o json is not poorer than
	// the API it wraps. The antihack and arp objects hold a `logs` field whose
	// content is the reason for the block; it is far too long for a table and
	// far too useful to drop.
	Raw map[string]any
}

// blockedAddresses reads the three block lists of an IP block.
//
// A mechanism that answers nothing is not an error: measured on 537 blocks of
// a real account, all three answered 200 with an empty list on every single
// one. "Nothing is blocked" is the answer this command exists to give.
func blockedAddresses(ipBlock string) ([]blockedAddress, error) {
	var found []blockedAddress

	for _, mechanism := range blockMechanisms {
		addresses, err := blockedBy(ipBlock, mechanism)
		if err != nil {
			return nil, err
		}
		found = append(found, addresses...)
	}

	sort.Slice(found, func(i, j int) bool {
		if found[i].IP != found[j].IP {
			return found[i].IP < found[j].IP
		}
		return found[i].Mechanism < found[j].Mechanism
	})

	return found, nil
}

// blockedBy reads one mechanism's list, then each entry's detail.
func blockedBy(ipBlock, mechanism string) ([]blockedAddress, error) {
	var addresses []string

	listPath := fmt.Sprintf("/v1/ip/%s/%s", url.PathEscape(ipBlock), mechanism)
	if err := httpLib.Client.Get(listPath, &addresses); err != nil {
		return nil, fmt.Errorf("failed to read the %s blocks of %s: %w", mechanism, ipBlock, err)
	}

	blocked := make([]blockedAddress, 0, len(addresses))
	for _, address := range addresses {
		var detail map[string]any
		detailPath := fmt.Sprintf("%s/%s", listPath, url.PathEscape(address))
		if err := httpLib.Client.Get(detailPath, &detail); err != nil {
			return nil, fmt.Errorf("failed to read the %s block of %s: %w", mechanism, address, err)
		}

		blocked = append(blocked, blockedAddress{
			IP:        address,
			Mechanism: mechanism,
			State:     stringField(detail, "state"),
			Since:     firstStringField(detail, "blockedSince", "date"),
			Seconds:   intField(detail, "time"),
			Raw:       detail,
		})
	}

	return blocked, nil
}

// ListBlocked shows every address of an IP block held by one of the three
// blocking mechanisms.
func ListBlocked(_ *cobra.Command, args []string) {
	ipBlock := args[0]

	blocked, err := blockedAddresses(ipBlock)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	if len(blocked) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "blocked": []any{}},
			"✅ No address of %s is blocked by anti-hack, ARP or anti-spam.", ipBlock)
		return
	}

	rows := make([]map[string]any, 0, len(blocked))
	for _, address := range blocked {
		row := map[string]any{}
		for key, value := range address.Raw {
			row[key] = value
		}
		row["ip"] = address.IP
		row["mechanism"] = address.Mechanism
		row["since"] = address.Since
		row["note"] = blockNote(address)
		rows = append(rows, row)
	}

	renderFiltered(rows, []string{"ip", "mechanism", "state", "since", "note"})
}

// blockNote turns the API's `time` into a sentence that is true for the
// mechanism that produced it. See blockedAddress.Seconds for why one number
// cannot be printed under one header.
func blockNote(address blockedAddress) string {
	switch address.Mechanism {
	case spamMechanism:
		if address.Seconds <= 0 {
			return ""
		}
		return fmt.Sprintf("blocked for %s", formatDelay(address.Seconds))
	default:
		if address.Seconds <= 0 {
			return "can be unblocked now"
		}
		return fmt.Sprintf("can be unblocked in %s", formatDelay(address.Seconds))
	}
}

// UnblockIp releases an address from the mechanism holding it.
func UnblockIp(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	mechanism, cooldown, err := mechanismToRelease(ipBlock, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	// The wait is checked here rather than left to the API because the API
	// answers it with "The requested object (ipBlocked = x) does not exist",
	// which reads as "this address is unknown" when it means "not yet".
	if cooldown > 0 {
		display.OutputError(&flags.OutputFormatConfig,
			"%s is blocked by %s and cannot be released for another %s.",
			target, mechanism, formatDelay(cooldown))
		return
	}

	if !common.ConfirmAction(common.Disruptive, target, fmt.Sprintf(
		"Releasing %s from the %s block lets its traffic through again.", target, mechanism)) {
		display.OutputError(&flags.OutputFormatConfig, "unblock of %s cancelled", target)
		return
	}

	endpoint := fmt.Sprintf("/v1/ip/%s/%s/%s/unblock",
		url.PathEscape(ipBlock), mechanism, url.PathEscape(target))
	if common.ReportDryRun(common.Call{Method: "POST", Endpoint: endpoint}) {
		return
	}

	if err := httpLib.Client.Post(endpoint, nil, nil); err != nil {
		display.OutputError(&flags.OutputFormatConfig, "failed to unblock %s: %s", target, err)
		return
	}

	display.OutputInfo(&flags.OutputFormatConfig,
		map[string]any{"ip": target, "mechanism": mechanism},
		"✅ %s was released from the %s block. Follow it with: ovhcloud ip blocked %s",
		target, mechanism, ipBlock)
}

// mechanismToRelease works out which of the three mechanisms holds an address,
// and how long is left before it accepts a release.
//
// --reason short-circuits the lookup: a script that already knows the mechanism
// should not pay three requests to be told what it passed in. Without it the
// three lists are read, because an operator reading "my address is blocked" has
// no way to know which of the three did it, and guessing wrong costs a 404 that
// says the address does not exist.
func mechanismToRelease(ipBlock, target string) (string, int64, error) {
	if UnblockReason != "" {
		if !slicesContainString(blockMechanisms, UnblockReason) {
			return "", 0, fmt.Errorf("unknown reason %q: an address is blocked by %s",
				UnblockReason, strings.Join(blockMechanisms, ", "))
		}

		return UnblockReason, cooldownOf(ipBlock, UnblockReason, target), nil
	}

	blocked, err := blockedAddresses(ipBlock)
	if err != nil {
		return "", 0, err
	}

	return chooseMechanism(ipBlock, target, blocked)
}

// chooseMechanism picks the one mechanism holding an address among everything
// the block reported.
//
// It is separate from the reading so the three answers it can give — none, one,
// several — can be exercised without an account that has an address blocked
// three ways. Nothing on the account measured for this lot was blocked at all.
func chooseMechanism(ipBlock, target string, blocked []blockedAddress) (string, int64, error) {
	var holding []blockedAddress
	for _, address := range blocked {
		if strings.EqualFold(address.IP, target) {
			holding = append(holding, address)
		}
	}

	switch len(holding) {
	case 0:
		return "", 0, fmt.Errorf("%s is not blocked by anti-hack, ARP or anti-spam, so there is nothing to release.\n   List what is blocked with: ovhcloud ip blocked %s",
			target, ipBlock)
	case 1:
		return holding[0].Mechanism, holding[0].Seconds, nil
	default:
		mechanisms := make([]string, 0, len(holding))
		for _, address := range holding {
			mechanisms = append(mechanisms, address.Mechanism)
		}
		return "", 0, fmt.Errorf("%s is blocked by %s at once, and each is released separately.\n   Pick one with: --reason %s",
			target, strings.Join(mechanisms, " and "), mechanisms[0])
	}
}

// cooldownOf reads how long is left before a mechanism releases an address.
//
// A read failure answers zero rather than an error on purpose: --reason is the
// path a script takes, and turning a transient read into a refusal would stop a
// release the API would have accepted. The API still refuses too early requests
// on its own; this only improves the message when the read succeeds.
func cooldownOf(ipBlock, mechanism, target string) int64 {
	var detail map[string]any

	path := fmt.Sprintf("/v1/ip/%s/%s/%s",
		url.PathEscape(ipBlock), mechanism, url.PathEscape(target))
	if err := httpLib.Client.Get(path, &detail); err != nil {
		return 0
	}

	if mechanism == spamMechanism {
		// On spam the field is the length of the sentence, not a cooldown, so
		// it must not gate the release.
		return 0
	}

	return intField(detail, "time")
}

// SpamStats reports what an address sent while the anti-spam system held it.
func SpamStats(_ *cobra.Command, args []string) {
	ipBlock, target := args[0], args[1]

	from, to, source, err := spamWindow(ipBlock, target)
	if err != nil {
		display.OutputError(&flags.OutputFormatConfig, "%s", err)
		return
	}

	path := fmt.Sprintf("/v1/ip/%s/spam/%s/stats?from=%s&to=%s",
		url.PathEscape(ipBlock), url.PathEscape(target),
		url.QueryEscape(from), url.QueryEscape(to))

	var stats []map[string]any
	if err := httpLib.Client.Get(path, &stats); err != nil {
		display.OutputError(&flags.OutputFormatConfig,
			"failed to read the spam statistics of %s: %s", target, err)
		return
	}

	if len(stats) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": target, "from": from, "to": to, "stats": []any{}},
			"No spam was recorded for %s between %s and %s (%s).", target, from, to, source)
		return
	}

	rows := make([]map[string]any, 0, len(stats))
	for _, entry := range stats {
		row := map[string]any{}
		for key, value := range entry {
			row[key] = value
		}
		row["blockedAt"] = formatTimestamp(intField(entry, "timestamp"))
		if detected, ok := entry["detectedSpams"].([]any); ok {
			row["samples"] = len(detected)
		} else {
			row["samples"] = 0
		}
		rows = append(rows, row)
	}

	renderFiltered(rows, []string{"blockedAt", "total", "numberOfSpams", "averageSpamscore", "samples"})
}

// spamWindow decides the period the report covers.
//
// The API makes `from` and `to` mandatory and gives no default, so a command
// that just forwards them is unusable without two flags whose right values the
// operator cannot know. The window that answers the question is the one around
// the block, and the sibling route already holds its date — so the CLI reads it
// and says where the bound came from, the same way `baremetal traffic` resolves
// a server into the MAC addresses its own API demands.
func spamWindow(ipBlock, target string) (string, string, string, error) {
	to := SpamStatsTo
	if to == "" {
		to = time.Now().UTC().Format(time.RFC3339)
	}

	if SpamStatsFrom != "" {
		return SpamStatsFrom, to, "window given on the command line", nil
	}

	var detail map[string]any
	path := fmt.Sprintf("/v1/ip/%s/spam/%s", url.PathEscape(ipBlock), url.PathEscape(target))
	if err := httpLib.Client.Get(path, &detail); err != nil {
		return "", "", "", fmt.Errorf("%s is not flagged by the anti-spam system, so it has no statistics.\n   List the flagged addresses with: ovhcloud ip blocked %s\n   Or give an explicit window with --from and --to", target, ipBlock)
	}

	if blockedAt := stringField(detail, "date"); blockedAt != "" {
		return blockedAt, to, "since the block on " + blockedAt, nil
	}

	return time.Now().UTC().AddDate(0, 0, -7).Format(time.RFC3339), to, "last 7 days", nil
}

// ListPhishing lists the phishing entries reported on an IP block.
func ListPhishing(_ *cobra.Command, args []string) {
	ipBlock := args[0]
	basePath := fmt.Sprintf("/v1/ip/%s/phishing", url.PathEscape(ipBlock))

	var ids []int64
	if err := httpLib.Client.Get(basePath, &ids); err != nil {
		// This route answers 500 for every address hosted outside Europe —
		// measured on 52 of 537 blocks, while the four sibling families answer
		// 200 on those very same blocks. It is reported as the failure it is:
		// printing an empty table here would say "no phishing reported" about
		// a block nobody managed to read.
		display.OutputError(&flags.OutputFormatConfig,
			"failed to read the phishing entries of %s: %s", ipBlock, err)
		return
	}

	if len(ids) == 0 {
		display.OutputInfo(&flags.OutputFormatConfig,
			map[string]any{"ip": ipBlock, "phishing": []any{}},
			"✅ No phishing URL is reported on %s.", ipBlock)
		return
	}

	rows := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		var entry map[string]any
		if err := httpLib.Client.Get(fmt.Sprintf("%s/%d", basePath, id), &entry); err != nil {
			display.OutputError(&flags.OutputFormatConfig,
				"failed to read the phishing entry %d: %s", id, err)
			return
		}
		rows = append(rows, entry)
	}

	renderFiltered(rows, []string{"id", "ipOnAntiphishing", "state", "creationDate", "urlPhishing"})
}

// GetPhishing shows one phishing entry.
func GetPhishing(_ *cobra.Command, args []string) {
	common.ManageObjectRequest(
		fmt.Sprintf("/v1/ip/%s/phishing", url.PathEscape(args[0])), args[1], phishingTemplate)
}

// formatDelay says a number of seconds the way somebody waiting reads it.
func formatDelay(seconds int64) string {
	if seconds <= 0 {
		return "0s"
	}

	duration := time.Duration(seconds) * time.Second
	switch {
	case duration >= 24*time.Hour:
		days := seconds / 86400
		hours := (seconds % 86400) / 3600
		if hours == 0 {
			return fmt.Sprintf("%dd", days)
		}
		return fmt.Sprintf("%dd %dh", days, hours)
	case duration >= time.Hour:
		return fmt.Sprintf("%dh %dm", seconds/3600, (seconds%3600)/60)
	case duration >= time.Minute:
		return fmt.Sprintf("%dm", seconds/60)
	default:
		return fmt.Sprintf("%ds", seconds)
	}
}

// formatTimestamp renders a Unix timestamp, and an empty string when there is
// none. Zero is not a date here: it is the field being absent.
func formatTimestamp(value int64) string {
	if value <= 0 {
		return ""
	}
	return time.Unix(value, 0).UTC().Format(time.RFC3339)
}

func stringField(object map[string]any, key string) string {
	if value, ok := object[key].(string); ok {
		return value
	}
	return ""
}

func firstStringField(object map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := stringField(object, key); value != "" {
			return value
		}
	}
	return ""
}

// intField reads a whole number whatever shape the decoder left it in.
//
// go-ovh decodes with UseNumber, so a number reaches here as json.Number and
// never as float64. A type switch that only knew float64 turned the vRack task
// handling into dead code once already; this one is written from that.
func intField(object map[string]any, key string) int64 {
	switch value := object[key].(type) {
	case int64:
		return value
	case int:
		return int64(value)
	case float64:
		return int64(value)
	case json.Number:
		parsed, err := value.Int64()
		if err != nil {
			return 0
		}
		return parsed
	}
	return 0
}

func slicesContainString(haystack []string, needle string) bool {
	for _, candidate := range haystack {
		if strings.EqualFold(candidate, needle) {
			return true
		}
	}
	return false
}
