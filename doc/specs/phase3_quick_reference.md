# Phase 3 Quick Reference - Kubernetes Creation Wizard

## At a Glance

**What**: 6-step interactive wizard for creating Kubernetes clusters  
**Where**: Triggered via 'c' key on Kubernetes list view  
**Status**: ✅ Complete and Tested  
**Lines**: ~2,400 across api.go and manager.go

---

## Wizard Steps

### 1️⃣ Region Selection
- **Display**: List of available regions with codes
- **Keys**: ↑↓ Navigate, Enter Select, Backspace Cancel
- **Next**: Loads Kubernetes versions for selected region

### 2️⃣ Version Selection
- **Display**: K8s versions (newest first)
- **Keys**: ↑↓ Navigate, Enter Select, Backspace Back
- **Next**: Shows network selection

### 3️⃣ Network Selection
- **Display**: Public only OR select private network
- **Keys**: ↑↓ Navigate, Enter Select, Backspace Back
- **Next**: If private selected → subnet config, else → name input

### 4️⃣ Subnet Selection (conditional)
- **Display**: Two-phase selection (nodes, then LB)
- **Keys**: ↑↓ Navigate, Tab Switch, Enter Select, Backspace Back
- **Next**: Cluster name input

### 5️⃣ Cluster Name Input
- **Display**: Text input box with validation (3-32 chars)
- **Keys**: Type alphanumeric/hyphens, Backspace clear, Enter confirm
- **Validation**: Must be 3-32 alphanumeric or hyphen chars
- **Next**: Advanced options

### 6️⃣ Advanced Options
- **Display**: Plan, Update Policy, Proxy Mode, Private Routing, Gateway IP
- **Fields**:
  - Plan: free ↔ standard
  - Update Policy: always-update ↔ never-update
  - Proxy Mode: iptables ↔ ipvs
  - Private Routing: toggle (shows Gateway IP field when enabled)
  - Gateway IP: IPv4 address input
- **Keys**: ↑↓ Navigate, ←→ Toggle/Edit, Enter Confirm
- **Next**: Confirmation summary

### 7️⃣ Confirmation
- **Display**: Summary of all selections
- **Options**: [Create Cluster] or [Cancel]
- **Keys**: ←→ Select button, Enter confirm, Backspace Back
- **Action**: Creates cluster with selected configuration

---

## Data Flow

```
User triggers 'c' key on Kubernetes view
  ↓
creationWizardMsg(ProductKubernetes)
  ↓
initWizard(KubeWizardStepRegion) + fetchKubeRegions()
  ↓
kubeRegionsLoadedMsg → handleKubeRegionsLoaded()
  ↓
User selects region + Enter
  ↓
fetchKubeVersions(region)
  ↓
kubeVersionsLoadedMsg → handleKubeVersionsLoaded()
  ↓
[... repeat for networks, subnets ...]
  ↓
Name input + Enter → Advanced options
  ↓
Options configured + Enter → Confirmation
  ↓
Confirmation + Enter → createKubeCluster(config)
  ↓
kubeClusterCreatedMsg → handleKubeClusterCreated()
  ↓
Success notification + refresh Kubernetes list
```

---

## Key Code Locations

### WizardData Fields (manager.go, lines ~165-195)
```go
kubeRegions, kubeVersions, kubeNetworks, kubeSubnets
selectedKubeRegion, selectedKubeVersion, selectedKubeNetwork
kubeName, kubeNameInput
kubePlan, kubeUpdatePolicy, kubeProxyMode
kubePrivateRouting, kubeGatewayIP
kubeOptionsFieldIndex, kubeConfirmButtonIndex
```

### Rendering (manager.go, lines ~1800-2110)
```go
renderKubeWizardRegionStep()      // Region list display
renderKubeWizardVersionStep()     // Version list display
renderKubeWizardNetworkStep()     // Network selection
renderKubeWizardSubnetStep()      // Subnet configuration
renderKubeWizardNameStep()        // Name input
renderKubeWizardOptionsStep()     // Options form
renderKubeWizardConfirmStep()     // Confirmation summary
```

### Key Handlers (manager.go, lines ~4120-4410)
```go
handleKubeWizardRegionKeys()      // Region navigation & selection
handleKubeWizardVersionKeys()     // Version navigation & selection
handleKubeWizardNetworkKeys()     // Network selection
handleKubeWizardSubnetKeys()      // Subnet selection with Tab switching
handleKubeWizardNameKeys()        // Name input with validation
handleKubeWizardOptionsKeys()     // Options navigation & toggling
handleKubeWizardConfirmKeys()     // Confirmation & cluster creation
```

### API Functions (api.go, lines ~310-430)
```go
fetchKubeRegions()                // GET regions
fetchKubeVersions(region)         // GET versions
fetchKubeNetworks()               // GET networks
fetchKubeSubnets(networkID)       // GET subnets
createKubeCluster(config)         // POST create cluster
```

### Message Handlers (api.go, lines ~434-485)
```go
handleKubeRegionsLoaded()         // Process regions response
handleKubeVersionsLoaded()        // Process versions response
handleKubeNetworksLoaded()        // Process networks response
handleKubeSubnetsLoaded()         // Process subnets response
handleKubeClusterCreated()        // Process creation response
```

---

## Important Constants

```go
// Wizard steps (manager.go)
const (
  KubeWizardStepRegion   WizardStep = iota + 100  // 100
  KubeWizardStepVersion                           // 101
  KubeWizardStepNetwork                           // 102
  KubeWizardStepSubnet                            // 103
  KubeWizardStepName                              // 104
  KubeWizardStepOptions                           // 105
  KubeWizardStepConfirm                           // 106
)

// Use +100 offset to distinguish from instance wizard (steps 0-8)
// Allows easy type detection: if step >= 100 → Kubernetes
```

---

## Integration Points

### 1. Message Router (Update function, ~line 615)
```go
case kubeRegionsLoadedMsg:
  return m.handleKubeRegionsLoaded(msg)
case kubeVersionsLoadedMsg:
  return m.handleKubeVersionsLoaded(msg)
case kubeNetworksLoadedMsg:
  return m.handleKubeNetworksLoaded(msg)
case kubeSubnetsLoadedMsg:
  return m.handleKubeSubnetsLoaded(msg)
case kubeClusterCreatedMsg:
  return m.handleKubeClusterCreated(msg)
```

### 2. Wizard Initializer (creationWizardMsg handler, ~line 545)
```go
if msg.product == ProductKubernetes {
  m.mode = WizardView
  m.wizard = WizardData{
    step: KubeWizardStepRegion,
    isLoading: true,
    loadingMessage: "Loading Kubernetes regions...",
  }
  return m, m.fetchKubeRegions()
}
```

### 3. Key Router (handleWizardKeyPress, ~line 3100)
```go
case KubeWizardStepRegion:
  return m.handleKubeWizardRegionKeys(key, msg)
case KubeWizardStepVersion:
  return m.handleKubeWizardVersionKeys(key, msg)
// ... etc for all 7 steps
```

### 4. View Router (renderWizardView, ~line 1070)
```go
switch m.wizard.step {
// ... instance steps ...
case KubeWizardStepRegion:
  content.WriteString(m.renderKubeWizardRegionStep(width))
case KubeWizardStepVersion:
  content.WriteString(m.renderKubeWizardVersionStep(width))
// ... etc for all 7 steps
}
```

---

## Common Patterns

### List Navigation (e.g., Region Selection)
```go
func (m Model) handleKubeWizardRegionKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
  switch key {
  case "up":
    if m.wizard.selectedIndex > 0 {
      m.wizard.selectedIndex--
    }
  case "down":
    if m.wizard.selectedIndex < len(m.wizard.kubeRegions)-1 {
      m.wizard.selectedIndex++
    }
  case "enter":
    // Process selection
    region := m.wizard.kubeRegions[m.wizard.selectedIndex]
    m.wizard.selectedKubeRegion = region["code"].(string)
    m.wizard.step = KubeWizardStepVersion
    // Load next step data
    return m, m.fetchKubeVersions(m.wizard.selectedKubeRegion)
  }
}
```

### Text Input (Cluster Name)
```go
func (m Model) handleKubeWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
  switch msg.Type {
  case tea.KeyRunes:
    for _, r := range msg.Runes {
      if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
        if len(m.wizard.kubeNameInput) < 32 {
          m.wizard.kubeNameInput += string(r)
        }
      }
    }
  case tea.KeyBackspace:
    if len(m.wizard.kubeNameInput) > 0 {
      m.wizard.kubeNameInput = m.wizard.kubeNameInput[:len(m.wizard.kubeNameInput)-1]
    }
  case tea.KeyEnter:
    if len(m.wizard.kubeNameInput) >= 3 {
      m.wizard.kubeName = m.wizard.kubeNameInput
      m.wizard.step = KubeWizardStepOptions
    }
  }
}
```

### Option Toggling
```go
case "left", "right":
  switch m.wizard.kubeOptionsFieldIndex {
  case 0: // Plan
    if m.wizard.kubePlan == "free" {
      m.wizard.kubePlan = "standard"
    } else {
      m.wizard.kubePlan = "free"
    }
  case 1: // Update policy
    // Toggle between two options
  }
```

---

## Error Handling

### Failed API Call
```go
func (m Model) handleKubeRegionsLoaded(msg kubeRegionsLoadedMsg) (tea.Model, tea.Cmd) {
  if msg.err != nil {
    m.wizard.errorMsg = fmt.Sprintf("Failed to load regions: %s", msg.err)
    return m, nil
  }
  // Process success
}
```

### Creation Failure
```go
func (m Model) handleKubeClusterCreated(msg kubeClusterCreatedMsg) (tea.Model, tea.Cmd) {
  if msg.err != nil {
    m.wizard.errorMsg = fmt.Sprintf("Failed to create Kubernetes cluster: %s", msg.err)
    m.notification = fmt.Sprintf("❌ Cluster creation failed: %s", msg.err)
    return m, nil
  }
  // Process success - show notification and refresh list
}
```

---

## Testing Checklist

- [ ] Region list loads and displays
- [ ] Navigate regions with ↑↓, select with Enter
- [ ] Version list loads for selected region
- [ ] Network list shows "No private network" option
- [ ] Selecting private network loads subnets
- [ ] Subnet selection works with Tab switching
- [ ] Name input validates 3-32 chars
- [ ] Options can be toggled with ←→
- [ ] Private routing toggle shows/hides gateway IP
- [ ] Confirmation shows all values correctly
- [ ] Cluster creation succeeds
- [ ] Success notification appears
- [ ] Kubernetes list refreshes with new cluster
- [ ] Escape cancels wizard properly
- [ ] Backspace navigates back through steps

---

## Troubleshooting

### Regions/Versions/Networks not loading
- Check API endpoints in fetchKubeRegions(), fetchKubeVersions(), etc.
- Verify kubeProject is set correctly
- Check network connectivity
- Look for error message in errorMsg field

### Wizard doesn't advance
- Check that Enter key is pressed (not just navigation)
- Verify form validation passes (e.g., name must be 3-32 chars)
- Check console for any async operation errors

### Cluster not created
- Check createKubeCluster() payload format
- Verify POST endpoint `/v1/cloud/project/{id}/kube` is correct
- Check for required fields in cluster creation payload

---

## Performance Notes

- Region/version/network lists are paginated by API
- Only active subnet list is loaded (via fetchKubeSubnets)
- Messages are processed asynchronously to prevent UI blocking
- List rendering is optimized for large number of items

---

## Future Enhancements

1. **Node Pool Configuration** - Add step for initial node pool setup
2. **Cost Estimation** - Show estimated monthly cost based on plan
3. **Networking Presets** - Quick templates for common configurations
4. **VPC Integration** - Suggest networks based on existing infrastructure
5. **Template Export** - Generate Terraform or Helm configs

---

## References

- **Specification**: doc/specs/browser_kubernetes_spec.md
- **Implementation**: doc/specs/phase3_implementation.md
- **Development Report**: doc/specs/development_report.md
