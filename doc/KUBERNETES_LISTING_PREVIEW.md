# Kubernetes List View - Visual Preview

## Example Display

```
☁ OVHcloud Manager • Project: my-project

┌─ ☸️ Kubernetes ────────────────────────────────────────────────────┐
│                                                                    │
│ Filter: (press / to edit, Esc to clear)                           │
│                                                                    │
│ NAME                    STATUS        REGION    VERSION NODES UPDATE POLICY   │
│ ─────────────────────────────────────────────────────────────────────────── │
│ > my-cluster-01         🟢 READY       GRA5      1.32    3     ALWAYS_UPDATE   │
│   staging-kube          🟡 UPDATING    BHS5      1.31    5     MINIMAL_DOWNTIME│
│   prod-europe           🟢 READY       SBG5      1.32    10    ALWAYS_UPDATE   │
│   test-cluster          🔴 ERROR       WAW1      1.30    2     NEVER_UPDATE    │
│                                                                    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘

Press ↓ for next, ↑ for previous, Enter for details, 'n' to create, '/' to filter
```

## With Filter Applied

```
☁ OVHcloud Manager • Project: my-project

┌─ ☸️ Kubernetes ────────────────────────────────────────────────────┐
│                                                                    │
│ Filter: prod (press / to edit, Esc to clear)                      │
│                                                                    │
│ NAME                    STATUS        REGION    VERSION NODES UPDATE POLICY   │
│ ─────────────────────────────────────────────────────────────────────────── │
│ > prod-europe           🟢 READY       SBG5      1.32    10    ALWAYS_UPDATE   │
│                                                                    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘

Press ↓ for next, ↑ for previous, Enter for details, 'n' to create, '/' to filter
```

## Empty State

```
☁ OVHcloud Manager • Project: my-project

┌─ ☸️ Kubernetes ────────────────────────────────────────────────────┐
│                                                                    │
│        📭                                                          │
│                                                                    │
│        No Kubernetes clusters found in this project               │
│                                                                    │
│        Press 'c' to create one, or run:                           │
│                                                                    │
│        ovhcloud cloud kube create --cloud-project <project-id>    │
│                                                                    │
└────────────────────────────────────────────────────────────────────┘
```

## Status Icon Legend

| Icon | Status | Description |
|------|--------|-------------|
| 🟢 | READY | Cluster is fully operational |
| 🟡 | INSTALLING | Cluster is being created |
| 🟡 | UPDATING | Cluster is being updated to a new version |
| 🟡 | RESTARTING | Control plane is restarting |
| 🟡 | RESETTING | Cluster is being reset |
| 🔴 | ERROR | Cluster has encountered an error |
| 🔴 | DELETING | Cluster is being deleted |
| 🔴 | SUSPENDED | Cluster is suspended |

## Key Controls

### Navigation
- **↑/↓** or **j/k**: Move up/down through clusters
- **Enter**: Open detailed view for selected cluster
- **q**: Go back to product selection

### Filtering
- **/**: Enter filter mode
- **Esc**: Clear filter and exit filter mode
- **Enter**: Confirm filter and exit filter mode
- **Backspace**: Delete last character in filter

### Creation
- **n** or **c**: Create new Kubernetes cluster (exits to CLI with command)

### Other
- **?**: Show help
- **D**: Toggle debug panel (shows API requests)
- **Ctrl+C**: Quit

## Column Details

| Column | Width | Content | Notes |
|--------|-------|---------|-------|
| Name | 25 chars | Cluster name | Sorted alphabetically |
| Status | 12 chars | Status + icon | Color varies by status |
| Region | 10 chars | OVHcloud region | GRA5, BHS5, SBG5, etc. |
| Version | 10 chars | K8s version | e.g., 1.32, 1.31 |
| Nodes | 6 chars | Node count | Total running nodes |
| Update Policy | 15 chars | Update strategy | ALWAYS_UPDATE, etc. |

## Filtering Examples

### Filter by Name
```
Filter: my-cluster
```
Shows only clusters containing "my-cluster" in the name (case-insensitive)

### Filter by Status
```
Filter: ready
```
Shows only READY clusters

### Filter by Region
```
Filter: gra
```
Shows only clusters in GRA regions

### Filter by Version
```
Filter: 1.32
```
Shows only clusters running Kubernetes 1.32
