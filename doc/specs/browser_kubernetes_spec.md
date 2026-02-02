# Kubernetes Browser Feature Specification

## Overview

This specification describes the implementation of Kubernetes cluster management features within the OVHcloud CLI browser TUI (Terminal User Interface). The browser already supports listing and viewing cloud resources; this spec extends it to provide full Kubernetes cluster list view, detail view, and a creation wizard.

## Scope

### In Scope
1. **List Kubernetes Clusters**: Display all K8s clusters in a table view with key information
2. **View Cluster Details**: Detailed view of a selected cluster with status, configuration, and actions
3. **Create Cluster Wizard**: Multi-step wizard to create a new Kubernetes cluster
4. **Quick Actions**: Actions from the detail view (kubeconfig, upgrade, restart, delete)

### Out of Scope (Future Work)
- Node pool management wizard (add/edit/delete node pools)
- OIDC configuration wizard
- IP restrictions management
- Real-time cluster status updates via websocket

---

## 1. List View (TableView)

### API Endpoint
```
GET /v1/cloud/project/{projectId}/kube
GET /v1/cloud/project/{projectId}/kube/{kubeId}  (for each cluster details)
```

### Table Columns
| Column       | API Field       | Width  | Description                              |
|--------------|-----------------|--------|------------------------------------------|
| Name         | `name`          | 25     | Cluster name                             |
| Status       | `status`        | 12     | Cluster status with icon (🟢/🟡/🔴)       |
| Region       | `region`        | 10     | OVHcloud region (e.g., GRA5, BHS5)       |
| Version      | `version`       | 10     | Kubernetes version (e.g., 1.32)          |
| Nodes        | `nodesCount`    | 6      | Total number of nodes                    |
| Update Policy| `updatePolicy`  | 15     | ALWAYS_UPDATE, MINIMAL_DOWNTIME, etc.    |

### Status Icons
- 🟢 `READY` - Cluster is operational
- 🟡 `INSTALLING`, `UPDATING`, `RESTARTING`, `RESETTING` - In progress
- 🔴 `ERROR`, `DELETING`, `SUSPENDED` - Error or unavailable

### Keyboard Navigation
| Key          | Action                                    |
|--------------|-------------------------------------------|
| `↑/↓` or `j/k` | Navigate through clusters              |
| `Enter`      | Open cluster detail view                  |
| `n` or `c`   | Open create cluster wizard                |
| `/`          | Enter filter mode                         |
| `Escape`     | Clear filter or exit                      |
| `q`          | Back to product selection                 |
| `?`          | Toggle help panel                         |

### Empty State
When no clusters exist, display:
```
┌─ Kubernetes clusters ─────────────────────────────────┐
│                                                        │
│   ☸️  No Kubernetes clusters found                     │
│                                                        │
│   Press 'n' to create your first cluster              │
│   or run: ovhcloud cloud kube create                  │
│           --cloud-project {projectId}                 │
│                                                        │
└────────────────────────────────────────────────────────┘
```

---

## 2. Detail View (DetailView)

### Layout
```
┌─ Actions (←/→ pour naviguer, Enter pour exécuter) ────────────────────────────────┐
│ [Kubeconfig] [Upgrade] [Restart] [Edit] [Delete]                                   │
└────────────────────────────────────────────────────────────────────────────────────┘

┌─ Cluster: my-cluster ──────────────┐  ┌─ Configuration ────────────────────────────┐
│ Status         🟢 READY            │  │ Update Policy    ALWAYS_UPDATE             │
│ ID             abc123-def456...    │  │ Plan             free                      │
│ Region         GRA5                │  │ Kube-proxy Mode  iptables                  │
│ Version        1.32                │  │ Private Network  my-network (optional)     │
│ Nodes          3                   │  │ Nodes Subnet     subnet-123 (optional)     │
│ Created        2026-01-15 10:30    │  │ LB Subnet        subnet-456 (optional)     │
└────────────────────────────────────┘  └────────────────────────────────────────────┘

┌─ Node Pools ──────────────────────────────────────────────────────────────────────┐
│ NAME              FLAVOR      NODES   STATUS    AUTOSCALE                          │
│ default-pool      b3-8        3/3     READY     ✗                                  │
│ worker-pool       b3-16       0-5/2   READY     ✓ (min:0, max:5)                   │
└────────────────────────────────────────────────────────────────────────────────────┘
```

### Actions
| Action       | Description                                          | Requires Confirmation |
|--------------|------------------------------------------------------|----------------------|
| Kubeconfig   | Generate and display kubeconfig                      | No                   |
| Upgrade      | Upgrade cluster to next version                      | Yes                  |
| Restart      | Restart control plane                                | Yes                  |
| Edit         | Edit cluster name/update policy                      | No                   |
| Delete       | Delete the cluster                                   | Yes (type name)      |

### Kubeconfig Action
When "Kubeconfig" is selected:
1. Call `POST /v1/cloud/project/{projectId}/kube/{kubeId}/kubeconfig`
2. Display options:
   - Copy to clipboard
   - Save to file (~/.kube/config or custom path)
   - Merge with existing kubeconfig
3. Show success notification

### Additional Data to Fetch
```
GET /v1/cloud/project/{projectId}/kube/{kubeId}/nodepool  (list node pools)
```

---

## 3. Create Cluster Wizard (WizardView)

### Wizard Steps

```
Step 1: Region     →  Step 2: Version    →  Step 3: Network (optional)  →
Step 4: Name       →  Step 5: Options    →  Step 6: Confirm & Create
```

### Wizard Step Constants (to add)
```go
const (
    KubeWizardStepRegion WizardStep = iota + 100  // Offset to not conflict with instance wizard
    KubeWizardStepVersion
    KubeWizardStepNetwork
    KubeWizardStepName
    KubeWizardStepOptions
    KubeWizardStepConfirm
)
```

### Step 1: Select Region

**API Call**: 
```
GET /v1/cloud/project/{projectId}/capabilities/kube/regions
```
Falls back to:
```
GET /v1/cloud/project/{projectId}/kube/regions
```

**Display**:
```
┌─ Create Kubernetes Cluster ───────────────────────────────────────────────────────┐
│ Step 1/6: Select Region                                                            │
│                                                                                    │
│ Choose the region where your cluster will be deployed:                            │
│                                                                                    │
│   > GRA5 (Gravelines, France)                                                     │
│     GRA7 (Gravelines, France)                                                     │
│     BHS5 (Beauharnois, Canada)                                                    │
│     SBG5 (Strasbourg, France)                                                     │
│     WAW1 (Warsaw, Poland)                                                         │
│     DE1  (Frankfurt, Germany)                                                     │
│                                                                                    │
│ [Filter: _________]                                                               │
│                                                                                    │
│ ↑/↓: Navigate  Enter: Select  /: Filter  Escape: Cancel                          │
└────────────────────────────────────────────────────────────────────────────────────┘
```

### Step 2: Select Kubernetes Version

**API Call**:
```
GET /v1/cloud/project/{projectId}/capabilities/kube/versions
```
Falls back to:
```
GET /v1/cloud/project/{projectId}/kube/versions
```

**Display**:
```
┌─ Create Kubernetes Cluster ───────────────────────────────────────────────────────┐
│ Step 2/6: Select Kubernetes Version                                                │
│                                                                                    │
│ Region: GRA5                                                                       │
│                                                                                    │
│ Choose the Kubernetes version:                                                     │
│                                                                                    │
│   > 1.32 (latest)                                                                 │
│     1.31                                                                          │
│     1.30                                                                          │
│     1.29                                                                          │
│                                                                                    │
│ ↑/↓: Navigate  Enter: Select  Backspace: Previous step  Escape: Cancel           │
└────────────────────────────────────────────────────────────────────────────────────┘
```

### Step 3: Network Configuration (Optional)

**API Calls**:
```
GET /v1/cloud/project/{projectId}/network/private  (list private networks)
GET /v1/cloud/project/{projectId}/network/private/{networkId}/subnet  (list subnets)
```

**Display**:
```
┌─ Create Kubernetes Cluster ───────────────────────────────────────────────────────┐
│ Step 3/6: Network Configuration (Optional)                                         │
│                                                                                    │
│ Region: GRA5  |  Version: 1.32                                                     │
│                                                                                    │
│ Configure private network for your cluster:                                        │
│                                                                                    │
│   > No private network (public only)                                              │
│     ─────────────────────────────────────                                         │
│     my-private-network (10.0.0.0/24)                                              │
│     production-network (172.16.0.0/16)                                            │
│     ─────────────────────────────────────                                         │
│     + Create new private network                                                  │
│                                                                                    │
│ ℹ️  Private network enables secure communication between cluster nodes             │
│    and other OVHcloud services (databases, instances, etc.)                       │
│                                                                                    │
│ ↑/↓: Navigate  Enter: Select  Backspace: Previous  Escape: Cancel                │
└────────────────────────────────────────────────────────────────────────────────────┘
```

**If Private Network Selected - Subnet Selection**:
```
┌─ Create Kubernetes Cluster ───────────────────────────────────────────────────────┐
│ Step 3/6: Network Configuration                                                    │
│                                                                                    │
│ Network: my-private-network                                                        │
│                                                                                    │
│ Select subnet for cluster nodes:                                                   │
│                                                                                    │
│   > subnet-default (10.0.0.0/24) - DHCP enabled                                   │
│     subnet-workers (10.0.1.0/24) - DHCP enabled                                   │
│                                                                                    │
│ (Optional) Select subnet for load balancers:                                       │
│                                                                                    │
│   > Same as nodes subnet                                                          │
│     subnet-lb (10.0.2.0/24) - DHCP enabled                                        │
│                                                                                    │
│ ↑/↓: Navigate  Enter: Select  Backspace: Previous  Escape: Cancel                │
└────────────────────────────────────────────────────────────────────────────────────┘
```

### Step 4: Cluster Name

**Display**:
```
┌─ Create Kubernetes Cluster ───────────────────────────────────────────────────────┐
│ Step 4/6: Cluster Name                                                             │
│                                                                                    │
│ Region: GRA5  |  Version: 1.32  |  Network: my-private-network                     │
│                                                                                    │
│ Enter a name for your cluster:                                                     │
│                                                                                    │
│   Cluster name: my-production-cluster_                                            │
│                                                                                    │
│ ℹ️  Name must be alphanumeric with dashes/underscores (max 63 chars)               │
│                                                                                    │
│ Enter: Continue  Backspace: Previous  Escape: Cancel                              │
└────────────────────────────────────────────────────────────────────────────────────┘
```

**Validation**:
- Required field
- Alphanumeric + dashes/underscores
- Max 63 characters
- Must start with a letter

### Step 5: Advanced Options (Optional)

**Display**:
```
┌─ Create Kubernetes Cluster ───────────────────────────────────────────────────────┐
│ Step 5/6: Advanced Options                                                         │
│                                                                                    │
│ Configure additional settings (all optional):                                      │
│                                                                                    │
│   Plan:           [free     ▼]  (free / standard)                                 │
│   Update Policy:  [ALWAYS_UPDATE ▼]                                               │
│   Kube-proxy:     [iptables ▼]  (iptables / ipvs)                                 │
│                                                                                    │
│ Private Network Routing (if private network selected):                             │
│   [ ] Use private network routing as default                                      │
│   Gateway IP:     [____________]  (e.g., 10.0.0.1)                                │
│                                                                                    │
│ Tab: Next field  Enter: Continue  Backspace: Previous  Escape: Cancel            │
└────────────────────────────────────────────────────────────────────────────────────┘
```

**Options**:
| Field              | Options                                          | Default        |
|--------------------|--------------------------------------------------|----------------|
| Plan               | `free`, `standard`                               | `free`         |
| Update Policy      | `ALWAYS_UPDATE`, `MINIMAL_DOWNTIME`, `NEVER_UPDATE` | `ALWAYS_UPDATE` |
| Kube-proxy Mode    | `iptables`, `ipvs`                               | `iptables`     |
| Private Routing    | checkbox                                          | unchecked      |
| Gateway IP         | text input (optional)                            | empty          |

### Step 6: Confirmation & Create

**Display**:
```
┌─ Create Kubernetes Cluster ───────────────────────────────────────────────────────┐
│ Step 6/6: Review and Create                                                        │
│                                                                                    │
│ Please review your cluster configuration:                                          │
│                                                                                    │
│   ┌─ Cluster Summary ─────────────────────────────────────────────────────────┐   │
│   │ Name:              my-production-cluster                                   │   │
│   │ Region:            GRA5                                                    │   │
│   │ Kubernetes:        1.32                                                    │   │
│   │ Plan:              free                                                    │   │
│   │ Update Policy:     ALWAYS_UPDATE                                           │   │
│   │ Kube-proxy Mode:   iptables                                                │   │
│   │ Private Network:   my-private-network                                      │   │
│   │ Nodes Subnet:      subnet-default (10.0.0.0/24)                           │   │
│   │ LB Subnet:         Same as nodes                                          │   │
│   └────────────────────────────────────────────────────────────────────────────┘   │
│                                                                                    │
│   ⚠️  The cluster will be created without any node pools.                          │
│      You can add node pools after creation via the CLI or browser.                │
│                                                                                    │
│           [Cancel]                    [Create Cluster]                            │
│                                                                                    │
│ ←/→: Select button  Enter: Confirm  Backspace: Previous  Escape: Cancel          │
└────────────────────────────────────────────────────────────────────────────────────┘
```

### Creation API Call

```
POST /v1/cloud/project/{projectId}/kube
```

**Request Body**:
```json
{
  "name": "my-production-cluster",
  "region": "GRA5",
  "version": "1.32",
  "plan": "free",
  "updatePolicy": "ALWAYS_UPDATE",
  "kubeProxyMode": "iptables",
  "privateNetworkId": "pn-12345-67890",
  "nodesSubnetId": "subnet-abc123",
  "loadBalancersSubnetId": "subnet-abc123",
  "privateNetworkConfiguration": {
    "defaultVrackGateway": "10.0.0.1",
    "privateNetworkRoutingAsDefault": true
  }
}
```

### Post-Creation Flow

1. Show loading spinner: "Creating cluster... This may take a few minutes."
2. On success:
   - Display success notification
   - Show option: "Would you like to add a node pool now?"
   - Return to cluster detail view (showing status: INSTALLING)
3. On error:
   - Display error message
   - Offer to retry or go back to edit

---

## 4. Data Structures

### KubeWizardData (to add to WizardData struct)
```go
// Kubernetes wizard fields
kubeRegions             []map[string]interface{} // Available regions for K8s
kubeVersions            []string                 // Available K8s versions
kubeNetworks            []map[string]interface{} // Private networks
kubeSubnets             []map[string]interface{} // Subnets for selected network
kubeLBSubnets           []map[string]interface{} // Subnets for load balancers
selectedKubeRegion      string                   // Selected region
selectedKubeVersion     string                   // Selected K8s version
selectedKubeNetwork     string                   // Selected private network ID
selectedKubeNetworkName string                   // Selected private network name
selectedNodesSubnet     string                   // Selected nodes subnet ID
selectedNodesSubnetCIDR string                   // Selected nodes subnet CIDR
selectedLBSubnet        string                   // Selected LB subnet ID (empty = same as nodes)
selectedLBSubnetCIDR    string                   // Selected LB subnet CIDR
kubeName                string                   // Cluster name
kubePlan                string                   // "free" or "standard"
kubeUpdatePolicy        string                   // Update policy
kubeProxyMode           string                   // "iptables" or "ipvs"
kubePrivateRouting      bool                     // Use private routing as default
kubeGatewayIP           string                   // vRack gateway IP
kubeOptionsFieldIndex   int                      // Current field in options step
kubeConfirmButtonIndex  int                      // 0 = Cancel, 1 = Create
```

### Messages (to add)
```go
// Kubernetes wizard messages
type kubeRegionsLoadedMsg struct {
    regions []map[string]interface{}
    err     error
}

type kubeVersionsLoadedMsg struct {
    versions []string
    err      error
}

type kubeNetworksLoadedMsg struct {
    networks []map[string]interface{}
    err      error
}

type kubeSubnetsLoadedMsg struct {
    subnets []map[string]interface{}
    err     error
}

type kubeClusterCreatedMsg struct {
    cluster map[string]interface{}
    err     error
}

type kubeNodePoolsLoadedMsg struct {
    nodePools []map[string]interface{}
    err       error
}

type kubeKubeconfigGeneratedMsg struct {
    kubeconfig string
    err        error
}
```

---

## 5. API Functions (to add to api.go)

```go
// fetchKubeRegions fetches available regions for Kubernetes
func (m Model) fetchKubeRegions() tea.Cmd

// fetchKubeVersions fetches available Kubernetes versions
func (m Model) fetchKubeVersions() tea.Cmd

// fetchKubeNetworks fetches private networks for K8s cluster
func (m Model) fetchKubeNetworks() tea.Cmd

// fetchKubeSubnets fetches subnets for a private network
func (m Model) fetchKubeSubnets(networkId string) tea.Cmd

// createKubeCluster creates a new Kubernetes cluster
func (m Model) createKubeCluster() tea.Cmd

// fetchKubeNodePools fetches node pools for a cluster
func (m Model) fetchKubeNodePools(kubeId string) tea.Cmd

// generateKubeconfig generates kubeconfig for a cluster
func (m Model) generateKubeconfig(kubeId string) tea.Cmd

// upgradeKubeCluster upgrades a cluster
func (m Model) upgradeKubeCluster(kubeId string, strategy string) tea.Cmd

// restartKubeCluster restarts control plane
func (m Model) restartKubeCluster(kubeId string, force bool) tea.Cmd

// deleteKubeCluster deletes a cluster
func (m Model) deleteKubeCluster(kubeId string) tea.Cmd
```

---

## 6. Rendering Functions (to add to manager.go)

```go
// renderKubeWizard renders the current step of the K8s creation wizard
func (m Model) renderKubeWizard(width, height int) string

// renderKubeWizardRegion renders region selection step
func (m Model) renderKubeWizardRegion(width int) string

// renderKubeWizardVersion renders version selection step
func (m Model) renderKubeWizardVersion(width int) string

// renderKubeWizardNetwork renders network configuration step
func (m Model) renderKubeWizardNetwork(width int) string

// renderKubeWizardSubnet renders subnet selection (sub-step)
func (m Model) renderKubeWizardSubnet(width int) string

// renderKubeWizardName renders name input step
func (m Model) renderKubeWizardName(width int) string

// renderKubeWizardOptions renders advanced options step
func (m Model) renderKubeWizardOptions(width int) string

// renderKubeWizardConfirm renders confirmation step
func (m Model) renderKubeWizardConfirm(width int) string
```

---

## 7. Update Handlers (to add to manager.go Update function)

Handle keyboard input for Kubernetes wizard steps:

```go
case KubeWizardStepRegion:
    // Handle region selection (up/down/enter/filter)
    
case KubeWizardStepVersion:
    // Handle version selection
    
case KubeWizardStepNetwork:
    // Handle network selection or skip
    
case KubeWizardStepName:
    // Handle text input for name
    
case KubeWizardStepOptions:
    // Handle options form navigation
    
case KubeWizardStepConfirm:
    // Handle confirmation buttons
```

---

## 8. Implementation Order

### Phase 1: List & Detail View Enhancement
1. Enhance `fetchKubernetesData()` to include more details
2. Update table columns for Kubernetes list
3. Enhance `renderKubernetesDetail()` with node pools
4. Add node pools fetching in detail view

### Phase 2: Detail View Actions
1. Implement kubeconfig generation action
2. Implement upgrade action with confirmation
3. Implement restart action with confirmation
4. Implement delete action with name confirmation

### Phase 3: Creation Wizard
1. Add wizard step constants and data structures
2. Implement region fetching and selection (Step 1)
3. Implement version fetching and selection (Step 2)
4. Implement network configuration (Step 3)
5. Implement name input (Step 4)
6. Implement options form (Step 5)
7. Implement confirmation and creation (Step 6)
8. Add post-creation flow

### Phase 4: Polish & Testing
1. Add loading states for all API calls
2. Add error handling and recovery
3. Add keyboard shortcut help
4. Test all wizard paths
5. Test edge cases (no networks, API errors, etc.)

---

## 9. Files to Modify

| File | Changes |
|------|---------|
| `internal/services/browser/manager.go` | Add wizard steps, data structures, render functions, update handlers |
| `internal/services/browser/api.go` | Add K8s-specific API functions |

---

## 10. Testing Scenarios

1. **List empty state**: No clusters → shows create prompt
2. **List with clusters**: Multiple clusters → correct display
3. **Filter clusters**: Filter by name works
4. **Detail view**: All fields display correctly
5. **Kubeconfig action**: Generates and offers save options
6. **Delete with confirmation**: Type name to confirm
7. **Wizard - happy path**: Complete all steps → cluster created
8. **Wizard - no networks**: Skip network step → cluster created
9. **Wizard - with network**: Select network and subnet → cluster created
10. **Wizard - cancel**: Cancel at any step → returns to list
11. **Wizard - back navigation**: Go back through steps
12. **Wizard - API error**: Handle errors gracefully
13. **Wizard - validation**: Invalid name shows error

---

## Appendix A: API Reference

### Kubernetes Endpoints Used

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/cloud/project/{projectId}/kube` | List cluster IDs |
| GET | `/v1/cloud/project/{projectId}/kube/{kubeId}` | Get cluster details |
| POST | `/v1/cloud/project/{projectId}/kube` | Create cluster |
| PUT | `/v1/cloud/project/{projectId}/kube/{kubeId}` | Update cluster |
| DELETE | `/v1/cloud/project/{projectId}/kube/{kubeId}` | Delete cluster |
| GET | `/v1/cloud/project/{projectId}/kube/regions` | Get available regions |
| GET | `/v1/cloud/project/{projectId}/kube/versions` | Get available versions |
| GET | `/v1/cloud/project/{projectId}/kube/{kubeId}/nodepool` | List node pools |
| POST | `/v1/cloud/project/{projectId}/kube/{kubeId}/kubeconfig` | Generate kubeconfig |
| POST | `/v1/cloud/project/{projectId}/kube/{kubeId}/restart` | Restart control plane |
| POST | `/v1/cloud/project/{projectId}/kube/{kubeId}/update` | Upgrade cluster |

### Network Endpoints Used

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/v1/cloud/project/{projectId}/network/private` | List private networks |
| GET | `/v1/cloud/project/{projectId}/network/private/{networkId}/subnet` | List subnets |

---

## Appendix B: Keyboard Shortcuts Summary

### Global (All Views)
| Key | Action |
|-----|--------|
| `q` | Quit / Go back |
| `?` | Toggle help |
| `D` | Toggle debug panel |
| `Tab` | Next navigation item |

### List View
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate items |
| `Enter` | Open detail |
| `n` or `c` | Create new |
| `/` | Filter |
| `r` | Refresh |

### Detail View
| Key | Action |
|-----|--------|
| `←/→` or `h/l` | Navigate actions |
| `Enter` | Execute action |
| `Escape` | Cancel confirmation |
| `Backspace` | Back to list |

### Wizard
| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate options |
| `Enter` | Select / Continue |
| `Backspace` | Previous step |
| `Escape` | Cancel wizard |
| `/` | Filter (in lists) |
| `Tab` | Next field (in forms) |
