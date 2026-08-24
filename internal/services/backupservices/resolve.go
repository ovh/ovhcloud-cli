// SPDX-FileCopyrightText: 2026 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

package backupservices

import (
	"fmt"
	"net/url"
	"strings"

	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
)

// The Veeam surface is a hierarchy of three UUIDs — tenant, then VSPC tenant,
// then agent — and none of them is a thing anybody knows by heart. An account
// has one tenant and one VSPC tenant to begin with, and typing two identifiers
// to reach the third is the kind of friction that sends people back to the web
// interface.
//
// So the levels are resolved when they are not ambiguous, exactly as
// `vrack attach` resolves a server to its interface and `baremetal logs`
// resolves a stream title. One is taken, several are refused with their names,
// none is an answer rather than an error.

const (
	// TenantsPath is the collection every other path hangs from.
	TenantsPath = "/v2/backupServices/tenant"
)

var (
	// Tenant and Vspc override the resolution when an account has more than one.
	Tenant string
	Vspc   string
)

// resource is the shape this whole API answers with: an identity, what was
// asked for, what is true now, and what is happening to it.
type resource struct {
	ID             string         `json:"id"`
	ResourceStatus string         `json:"resourceStatus"`
	TargetSpec     map[string]any `json:"targetSpec"`
	CurrentState   map[string]any `json:"currentState"`
	CurrentTasks   []currentTask  `json:"currentTasks"`
	CreatedAt      string         `json:"createdAt"`
	UpdatedAt      string         `json:"updatedAt"`
}

type currentTask struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Link   string `json:"link"`
	Status string `json:"status"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// name reads the human name of a resource, which lives in the target spec and
// is echoed in the current state.
func (r resource) name() string {
	if name, ok := r.CurrentState["name"].(string); ok && name != "" {
		return name
	}
	if name, ok := r.TargetSpec["name"].(string); ok && name != "" {
		return name
	}

	return r.ID
}

// ResolveTenant settles which backup tenant a command works on.
func ResolveTenant() (string, error) {
	if Tenant != "" {
		return Tenant, nil
	}

	var tenants []resource
	if err := httpLib.Client.Get(TenantsPath, &tenants); err != nil {
		return "", fmt.Errorf("failed to list the backup tenants: %w", err)
	}

	switch len(tenants) {
	case 1:
		return tenants[0].ID, nil

	case 0:
		return "", fmt.Errorf("this account has no backup tenant, so there is nothing to back up to.\n" +
			"   Order one before using these commands")

	default:
		return "", fmt.Errorf("this account has %d backup tenants; name one with --tenant:\n%s",
			len(tenants), listResources(tenants))
	}
}

// ResolveVspc settles which VSPC tenant a command works on, inside a tenant.
//
// The VSPC tenant is the Veeam Service Provider Console side of the product: it
// is what holds the agents, the licences and the deployment script, while the
// vaults beside it hold the storage.
func ResolveVspc(tenant string) (string, error) {
	if Vspc != "" {
		return Vspc, nil
	}

	vspcs, err := listVspc(tenant)
	if err != nil {
		return "", err
	}

	switch len(vspcs) {
	case 1:
		return vspcs[0].ID, nil

	case 0:
		return "", fmt.Errorf("backup tenant %s has no VSPC tenant, so it holds no agent", tenant)

	default:
		return "", fmt.Errorf("backup tenant %s has %d VSPC tenants; name one with --vspc:\n%s",
			tenant, len(vspcs), listResources(vspcs))
	}
}

// ResolveBoth is the pair every agent command needs.
func ResolveBoth() (tenant, vspc string, err error) {
	if tenant, err = ResolveTenant(); err != nil {
		return "", "", err
	}
	if vspc, err = ResolveVspc(tenant); err != nil {
		return "", "", err
	}

	return tenant, vspc, nil
}

func listVspc(tenant string) ([]resource, error) {
	var vspcs []resource

	path := fmt.Sprintf("%s/%s/vspc", TenantsPath, url.PathEscape(tenant))
	if err := httpLib.Client.Get(path, &vspcs); err != nil {
		return nil, fmt.Errorf("failed to list the VSPC tenants of %s: %w", tenant, err)
	}

	return vspcs, nil
}

// listResources renders a refusal that can be acted on: an identifier to paste
// and a name to recognise it by.
func listResources(resources []resource) string {
	lines := make([]string, 0, len(resources))
	for _, r := range resources {
		lines = append(lines, fmt.Sprintf("     %s  (%s)", r.ID, r.name()))
	}

	return strings.Join(lines, "\n")
}

// VspcPath is the prefix of everything that hangs off a VSPC tenant.
func VspcPath(tenant, vspc string) string {
	return fmt.Sprintf("%s/%s/vspc/%s", TenantsPath, url.PathEscape(tenant), url.PathEscape(vspc))
}

// VaultPath is the prefix of everything that hangs off a vault.
func VaultPath(tenant, vault string) string {
	return fmt.Sprintf("%s/%s/vault/%s", TenantsPath, url.PathEscape(tenant), url.PathEscape(vault))
}
