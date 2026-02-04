// SPDX-FileCopyrightText: 2025 OVH SAS <opensource@ovh.net>
//
// SPDX-License-Identifier: Apache-2.0

//go:build !(js && wasm)

package browser

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	"github.com/spf13/cobra"
)

// ViewMode represents the current view mode
type ViewMode int

const (
	ProjectSelectView  ViewMode = iota // Initial view to select a project
	TableView                          // List view for products
	DetailView                         // Detail view for a single item
	NodePoolsView                      // Node pools management view
	NodePoolDetailView                 // Detail view for a single node pool
	LoadingView
	ErrorView
	EmptyView                 // Empty list with creation prompt
	WizardView                // Multi-step wizard for resource creation
	DeleteConfirmView         // Confirmation dialog for deletion
	DebugView                 // Debug panel showing API requests
	KubeUpgradeView           // Kubernetes cluster upgrade selection
	KubePolicyEditView        // Kubernetes cluster policy edit
	KubeDeleteConfirmView     // Kubernetes cluster delete confirmation
	NodePoolScaleView         // Node pool scale view
	NodePoolDeleteConfirmView // Node pool delete confirmation
)

// ASCII OVHcloud logo for loading screen
const ovhcloudASCIILogo = `
   ____  __      __ _    _        _                    _ 
  / __ \ \ \    / /| |  | |      | |                  | |
 | |  | | \ \  / / | |__| |  ___ | |  ___   _   _   __| |
 | |  | |  \ \/ /  |  __  | / __|| | / _ \ | | | | / _` + "`" + ` |
 | |__| |   \  /   | |  | || (__ | || (_) || |_| || (_| |
  \____/     \/    |_|  |_| \___||_| \___/  \__,_| \__,_|
`

// WizardStep represents the current step in the creation wizard
type WizardStep int

const (
	// Instance wizard steps
	WizardStepRegion WizardStep = iota
	WizardStepFlavor
	WizardStepImage
	WizardStepSSHKey
	WizardStepNetwork
	WizardStepFloatingIP // For private network without public network
	WizardStepName
	WizardStepConfirm
	// Kubernetes wizard steps (offset by 100 to avoid conflicts)
	KubeWizardStepRegion WizardStep = iota + 100
	KubeWizardStepVersion
	KubeWizardStepNetwork
	KubeWizardStepSubnet
	KubeWizardStepName
	KubeWizardStepOptions
	KubeWizardStepConfirm
	// Node pool wizard steps (offset by 200)
	NodePoolWizardStepFlavor WizardStep = iota + 200
	NodePoolWizardStepName
	NodePoolWizardStepSize
	NodePoolWizardStepOptions
	NodePoolWizardStepConfirm
)

// ProductType represents a product category
type ProductType int

const (
	ProductInstances ProductType = iota
	ProductKubernetes
	ProductDatabases
	ProductStorage
	ProductNetworks
	ProductProjects
)

// WizardData holds the state for the creation wizard
type WizardData struct {
	step               WizardStep
	regions            []map[string]interface{}
	flavors            []map[string]interface{}
	images             []map[string]interface{}
	sshKeys            []map[string]interface{}
	privateNetworks    []map[string]interface{}
	selectedIndex      int    // Current selection index in the list
	selectedRegion     string // Selected region code
	selectedFlavor     string // Selected flavor ID
	selectedFlavorName string // Selected flavor display name
	selectedImage      string // Selected image ID
	selectedImageName  string // Selected image display name
	selectedSSHKey     string // Selected SSH key ID (empty = no key, "__create_new__" = create)
	selectedSSHKeyName string // Selected SSH key name
	// SSH key creation fields
	creatingSSHKey             bool     // Whether we're in SSH key creation mode
	newSSHKeyName              string   // Name for the new SSH key
	newSSHKeyPublicKey         string   // Public key content
	localPubKeys               []string // List of local .pub files from ~/.ssh
	sshKeyCreateField          int      // 0 = name, 1 = public key selection, 2 = Create/Cancel
	selectedLocalKeyIdx        int      // Index of selected local key (-1 = manual input)
	selectedPrivateNetwork     string   // Selected private network ID (empty = none)
	selectedPrivateNetworkName string   // Selected private network name
	selectedSubnetId           string   // Selected subnet ID for the private network
	usePublicNetwork           bool     // Whether to attach public network
	networkMenuIndex           int      // 0 = public toggle, 1 = private network selection
	instanceName               string   // Name for the new instance
	nameInput                  string   // Current input buffer for name
	isLoading                  bool     // Whether we're loading data
	loadingMessage             string   // Detailed loading message (e.g., "Creating network...")
	errorMsg                   string   // Error message if any
	// Network creation fields
	creatingNetwork    bool   // Whether we're in network creation mode
	newNetworkName     string // Name for the new network
	newNetworkVlanId   int    // VLAN ID (1-4094)
	newNetworkCIDR     string // CIDR for the subnet (default: 10.0.0.0/24)
	newNetworkDHCP     bool   // Enable DHCP for the subnet
	networkCreateField int    // 0 = name, 1 = VLAN ID, 2 = CIDR, 3 = DHCP, 4 = Create/Cancel
	// Floating IP fields (for private network without public network)
	floatingIPs               []map[string]interface{} // Available floating IPs
	selectedFloatingIP        string                   // Selected floating IP ID (empty = none, "__create_new__" = create)
	selectedFloatingIPAddress string                   // Selected floating IP address for display
	createdInstanceId         string                   // ID of the created instance (for floating IP attachment)
	createdInstanceName       string                   // Name of the created instance (for display)
	// Filter for wizard lists
	filterMode  bool   // Whether filter input mode is active in wizard
	filterInput string // Current filter input text for wizard lists
	// Cleanup tracking - IDs of resources created during wizard
	createdSSHKeyId     string // ID of SSH key created during wizard
	createdNetworkId    string // ID of network created during wizard
	createdSubnetId     string // ID of subnet created during wizard
	createdGatewayId    string // ID of gateway created during wizard
	createdFloatingIPId string // ID of floating IP created during wizard
	// Cleanup confirmation
	cleanupPending bool   // Whether we're waiting for cleanup confirmation
	cleanupError   string // Error message that triggered cleanup prompt
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
	kubeNameInput           string                   // Current input buffer for name
	kubePlan                string                   // "free" or "standard"
	kubeUpdatePolicy        string                   // Update policy
	kubeProxyMode           string                   // "iptables" or "ipvs"
	kubePrivateRouting      bool                     // Use private routing as default
	kubeGatewayIP           string                   // vRack gateway IP
	kubeGatewayIPInput      string                   // Current input for gateway IP
	kubeOptionsFieldIndex   int                      // Current field in options step (0-3: plan, policy, proxy, routing flag, 4: gateway IP, 5: buttons)
	kubeConfirmButtonIndex  int                      // 0 = Cancel, 1 = Create
	kubeSubnetMenuIndex     int                      // 0 = nodes subnet, 1 = LB subnet selection
	// Node pool wizard fields
	nodePoolClusterId       string                   // Cluster ID to add node pool to
	nodePoolFlavors         []map[string]interface{} // Available flavors for node pool
	nodePoolName            string                   // Node pool name
	nodePoolNameInput       string                   // Input buffer for name
	nodePoolFlavorName      string                   // Selected flavor name
	nodePoolDesiredNodes    int                      // Desired number of nodes
	nodePoolMinNodes        int                      // Minimum nodes (for autoscale)
	nodePoolMaxNodes        int                      // Maximum nodes (for autoscale)
	nodePoolAutoscale       bool                     // Enable autoscaling
	nodePoolAntiAffinity    bool                     // Enable anti-affinity
	nodePoolMonthlyBilled   bool                     // Monthly billing
	nodePoolSizeFieldIndex  int                      // 0 = desired, 1 = min, 2 = max
	nodePoolOptionsFieldIdx int                      // 0 = autoscale, 1 = anti-affinity, 2 = monthly
	nodePoolConfirmBtnIdx   int                      // 0 = Cancel, 1 = Create
	// Kube upgrade wizard fields
	kubeUpgradeClusterId   string   // Cluster ID for upgrade
	kubeUpgradeVersions    []string // Available upgrade versions
	kubeUpgradeSelectedIdx int      // Selected version index
	// Kube policy edit fields
	kubePolicyClusterId   string // Cluster ID for policy edit
	kubePolicySelectedIdx int    // Selected policy index
	// Kube delete confirmation fields
	kubeDeleteClusterId    string // Cluster ID for deletion
	kubeDeleteClusterName  string // Cluster name for confirmation
	kubeDeleteConfirmInput string // User input for confirmation
	// Node pool scale fields
	nodePoolScaleClusterId string // Cluster ID for scale
	nodePoolScalePoolId    string // Node pool ID for scale
	nodePoolScalePoolName  string // Node pool name for display
	nodePoolScaleDesired   int    // Desired nodes
	nodePoolScaleMin       int    // Min nodes
	nodePoolScaleMax       int    // Max nodes
	nodePoolScaleAutoscale bool   // Autoscale enabled
	nodePoolScaleFieldIdx  int    // Currently selected field
	// Node pool delete fields
	nodePoolDeleteClusterId    string // Cluster ID for deletion
	nodePoolDeletePoolId       string // Node pool ID for deletion
	nodePoolDeletePoolName     string // Node pool name for confirmation
	nodePoolDeleteConfirmInput string // User input for confirmation
}

// Model represents the TUI application state
type Model struct {
	width              int
	height             int
	mode               ViewMode
	previousMode       ViewMode // Previous mode to return to from debug view
	currentProduct     ProductType
	navIdx             int // Index in navigation bar
	table              table.Model
	detailData         map[string]interface{}
	currentData        []map[string]interface{}
	errorMsg           string
	cloudProject       string
	cloudProjectName   string                   // Display name of the selected project
	currentItemName    string                   // Name of the currently viewed item
	notification       string                   // Temporary notification message
	notificationExpiry time.Time                // When the notification should disappear
	projectsList       []map[string]interface{} // Cache of projects for selection
	wizard             WizardData               // Wizard state for resource creation
	selectedAction     int                      // Selected action index in detail view (0-5)
	actionConfirm      bool                     // Whether we're in confirmation mode for an action
	// Filter mode
	filterMode  bool   // Whether filter input mode is active
	filterInput string // Current filter input text
	// Delete confirmation
	deleteTarget       map[string]interface{} // Item to be deleted
	deleteConfirmInput string                 // User input for delete confirmation
	// Debug view
	debugScrollOffset int // Scroll offset for debug log view
	// Instance data cache
	imageMap      map[string]string // imageId -> imageName (for instances)
	floatingIPMap map[string]string // instanceId -> floatingIP address
	// Kubernetes data cache
	kubeNodePools           map[string][]map[string]interface{} // kubeId -> list of node pools
	nodePoolsSelectedIdx    int                                 // Selected index in node pools view
	selectedNodePool        map[string]interface{}              // Currently selected node pool for detail view
	nodePoolDetailActionIdx int                                 // Selected action index in node pool detail view
	nodePoolDetailConfirm   bool                                // Whether we're in confirmation mode
}

// Navigation items for the top bar
type NavItem struct {
	Label   string
	Icon    string
	Product ProductType
	Path    string
}

// Styles
var (
	// Header / Logo
	logoStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7B68EE"))

	// Navigation bar
	navBarStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

	navItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 2)

	navItemSelectedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF7F")).
				Bold(true).
				Padding(0, 2).
				Background(lipgloss.Color("#2a2a2a"))

	// Content area
	contentBoxStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#444444")).
			Padding(1, 2)

	// Title for current product
	productTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#7B68EE")).
				Padding(0, 2)

	// Detail view boxes
	boxTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7B68EE"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Width(18)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

	statusRunningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#00FF7F"))

	statusStoppedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FF6B6B"))

	// Footer
	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#666666")).
			Padding(0, 1)

	// Error and loading
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true).
			Padding(1, 2)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B68EE")).
			Padding(1, 2)
)

// Messages for async operations
type projectsLoadedMsg struct {
	projects   []map[string]interface{}
	err        error
	forProduct ProductType // The product that requested this data
}

type instancesLoadedMsg struct {
	instances     []map[string]interface{}
	imageMap      map[string]string // imageId -> imageName
	floatingIPMap map[string]string // instanceId -> floatingIP address
	err           error
	forProduct    ProductType // The product that requested this data
}

// instancesEnrichedMsg is sent when floating IPs and images are loaded after initial instances display
type instancesEnrichedMsg struct {
	imageMap      map[string]string // imageId -> imageName
	floatingIPMap map[string]string // instanceId -> floatingIP address
}

type dataLoadedMsg struct {
	data       []map[string]interface{}
	err        error
	forProduct ProductType // The product that requested this data
}

// setDefaultProjectMsg is returned after setting the default project
type setDefaultProjectMsg struct {
	projectID   string
	projectName string
	err         error
}

// clearNotificationMsg is sent to clear the notification after timeout
type clearNotificationMsg struct{}

// refreshTickMsg is sent to trigger automatic refresh of data
type refreshTickMsg struct{}

// Wizard-related messages
type regionsLoadedMsg struct {
	regions []map[string]interface{}
	images  []map[string]interface{}
	err     error
}

type flavorsLoadedMsg struct {
	flavors []map[string]interface{}
	err     error
}

type imagesLoadedMsg struct {
	images []map[string]interface{}
	err    error
}

type sshKeysLoadedMsg struct {
	sshKeys []map[string]interface{}
	err     error
}

type sshKeyCreatedMsg struct {
	sshKey map[string]interface{}
	err    error
}

type privateNetworksLoadedMsg struct {
	networks []map[string]interface{}
	err      error
}

type floatingIPsLoadedMsg struct {
	floatingIPs []map[string]interface{}
	err         error
}

type gatewayCreatedMsg struct {
	gateway map[string]interface{}
	err     error
}

type floatingIPCreatedMsg struct {
	floatingIP map[string]interface{}
	err        error
}

type floatingIPAttachedMsg struct {
	instanceName string
	err          error
}

type instanceIPReadyMsg struct {
	instanceId   string
	instanceName string
	privateIP    string
	err          error
}

// Network creation step messages
type networkStepMsg struct {
	step      string                 // "network_created", "creating_subnet", "subnet_created"
	networkId string                 // Network ID for subsequent steps
	network   map[string]interface{} // Network data
	err       error
}

type networkCreatedMsg struct {
	network map[string]interface{}
	err     error
}

type instanceCreatedMsg struct {
	instance map[string]interface{}
	err      error
}

type instanceDeletedMsg struct {
	success    bool
	instanceId string
	err        error
}

type cleanupCompletedMsg struct {
	deletedResources []string
	errors           []string
}

// progressMsg is used to update the loading message during async operations
type progressMsg struct {
	message string
}

// Instance action messages
type instanceActionMsg struct {
	action     string
	instanceId string
	err        error
}

// sshConnectionMsg is returned when SSH action is requested
type sshConnectionMsg struct {
	ip   string
	user string
}

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
	kubeId    string
	nodePools []map[string]interface{}
	err       error
}

// Navigation items for products (shown after project is selected)
func getNavItems() []NavItem {
	return []NavItem{
		{Label: "Instances", Icon: "💻", Product: ProductInstances, Path: "/instances"},
		{Label: "Kubernetes", Icon: "☸️", Product: ProductKubernetes, Path: "/kubernetes"},
		{Label: "Databases", Icon: "🗄️", Product: ProductDatabases, Path: "/databases"},
		{Label: "Storage", Icon: "💾", Product: ProductStorage, Path: "/storage/s3"},
		{Label: "Private networks", Icon: "🌐", Product: ProductNetworks, Path: "/networks/private"},
	}
}

// StartBrowser is the entry point for the browser TUI
func StartBrowser(cmd *cobra.Command, args []string) {
	httpLib.InitClient()

	// Reset creation command
	CreationCommand = ""

	initialModel := Model{
		mode:           LoadingView,
		currentProduct: ProductProjects, // Start with project selection
		navIdx:         0,
	}

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// If a creation command was requested, display it
	if CreationCommand != "" {
		fmt.Println()
		fmt.Println("🚀 To create a new resource, run:")
		fmt.Println()
		fmt.Printf("   %s\n", CreationCommand)
		fmt.Println()
	}
}

// Init initializes the model
func (m Model) Init() tea.Cmd {
	// Start by loading the list of projects
	return m.fetchDataForPath("/projects")
}

// Update handles all messages
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	case projectsLoadedMsg:
		return m.handleProjectsLoaded(msg)

	case instancesLoadedMsg:
		return m.handleInstancesLoaded(msg)

	case instancesEnrichedMsg:
		return m.handleInstancesEnriched(msg)

	case dataLoadedMsg:
		return m.handleDataLoaded(msg)

	case setDefaultProjectMsg:
		return m.handleSetDefaultProject(msg)

	case clearNotificationMsg:
		m.notification = ""
		return m, nil

	case refreshTickMsg:
		// Auto-refresh instances list if we're viewing instances in TableView
		// Only fetch if we're actually viewing the table, then reschedule
		if m.currentProduct == ProductInstances && m.mode == TableView {
			return m, tea.Batch(
				m.fetchDataForPath("/instances"),
				m.scheduleRefresh(),
			)
		}
		// Auto-refresh Kubernetes list if we're viewing Kubernetes in TableView
		if m.currentProduct == ProductKubernetes && m.mode == TableView {
			return m, tea.Batch(
				m.fetchDataForPath("/kubernetes"),
				m.scheduleRefresh(),
			)
		}
		// Not viewing instances or Kubernetes, don't reschedule (will be started again when switching)
		return m, nil

	case creationWizardMsg:
		// For instances and Kubernetes, launch the wizard; for other products, show the CLI command
		if msg.product == ProductInstances {
			m.mode = WizardView
			m.wizard = WizardData{
				step:           WizardStepRegion,
				isLoading:      true,
				loadingMessage: "Loading regions...",
			}
			return m, m.fetchRegions()
		} else if msg.product == ProductKubernetes {
			m.mode = WizardView
			m.wizard = WizardData{
				step:           KubeWizardStepRegion,
				isLoading:      true,
				loadingMessage: "Loading Kubernetes regions...",
			}
			return m, m.fetchKubeRegions()
		}
		// Store the creation command to be displayed after exit
		_, cmd := m.getProductCreationInfo()
		CreationCommand = cmd
		return m, tea.Quit

	case regionsLoadedMsg:
		return m.handleRegionsLoaded(msg)

	case flavorsLoadedMsg:
		return m.handleFlavorsLoaded(msg)

	case imagesLoadedMsg:
		return m.handleImagesLoaded(msg)

	case sshKeysLoadedMsg:
		return m.handleSSHKeysLoaded(msg)

	case sshKeyCreatedMsg:
		return m.handleSSHKeyCreated(msg)

	case privateNetworksLoadedMsg:
		return m.handlePrivateNetworksLoaded(msg)

	case floatingIPsLoadedMsg:
		return m.handleFloatingIPsLoaded(msg)

	case gatewayCreatedMsg:
		return m.handleGatewayCreated(msg)

	case floatingIPCreatedMsg:
		return m.handleFloatingIPCreated(msg)

	case floatingIPAttachedMsg:
		return m.handleFloatingIPAttached(msg)

	case instanceIPReadyMsg:
		return m.handleInstanceIPReady(msg)

	case progressMsg:
		m.wizard.loadingMessage = msg.message
		return m, nil

	case networkStepMsg:
		return m.handleNetworkStep(msg)

	case networkCreatedMsg:
		return m.handleNetworkCreated(msg)

	case instanceCreatedMsg:
		return m.handleInstanceCreated(msg)

	case instanceDeletedMsg:
		return m.handleInstanceDeleted(msg)

	case instanceActionMsg:
		return m.handleInstanceAction(msg)

	case sshConnectionMsg:
		return m.handleSSHConnection(msg)

	case cleanupCompletedMsg:
		return m.handleCleanupCompleted(msg)

	// Kubernetes wizard messages
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

	case kubeNodePoolsLoadedMsg:
		return m.handleKubeNodePoolsLoaded(msg)

	case kubeActionMsg:
		return m.handleKubeAction(msg)

	case launchK9sMsg:
		return m.handleLaunchK9s(msg)

	case switchToNodePoolsViewMsg:
		return m.handleSwitchToNodePoolsView(msg)

	case startNodePoolWizardMsg:
		return m.handleStartNodePoolWizard(msg)

	case nodePoolFlavorsLoadedMsg:
		return m.handleNodePoolFlavorsLoaded(msg)

	case nodePoolCreatedMsg:
		return m.handleNodePoolCreated(msg)

	// Kubernetes upgrade, policy, delete messages
	case startKubeUpgradeWizardMsg:
		return m.handleKubeUpgradeWizard(msg)

	case kubeUpgradeVersionsLoadedMsg:
		return m.handleKubeUpgradeVersionsLoaded(msg)

	case kubeUpgradeMsg:
		return m.handleKubeUpgraded(msg)

	case startKubePolicyEditMsg:
		return m.handleKubePolicyEdit(msg)

	case kubePolicyUpdatedMsg:
		return m.handleKubePolicyUpdated(msg)

	case startKubeDeleteMsg:
		return m.handleKubeDelete(msg)

	case kubeDeletedMsg:
		return m.handleKubeDeleted(msg)

	// Node pool action messages
	case startNodePoolScaleMsg:
		return m.handleStartNodePoolScale(msg)

	case nodePoolScaleMsg:
		return m.handleNodePoolScaled(msg)

	case startNodePoolDeleteMsg:
		return m.handleStartNodePoolDelete(msg)

	case nodePoolDeletedMsg:
		return m.handleNodePoolDeleted(msg)

	case tea.SuspendMsg:
		// TUI has been suspended
		return m, nil

	case tea.ResumeMsg:
		// Program is being resumed
		return m, nil
	}

	return m, nil
}

// CreationCommand stores the command to run after browser exits
var CreationCommand string

// handleSetDefaultProject handles the result of setting a default project
func (m Model) handleSetDefaultProject(msg setDefaultProjectMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.notification = fmt.Sprintf("❌ Failed to set default project: %s", msg.err)
	} else {
		// Update the current cloud project in the browser
		m.cloudProject = msg.projectID
		m.notification = fmt.Sprintf("✅ Default project set to: %s", msg.projectName)
	}
	m.notificationExpiry = time.Now().Add(3 * time.Second)

	// Schedule clearing the notification after 3 seconds
	return m, tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return clearNotificationMsg{}
	})
}

// handleSSHConnection opens an SSH connection to the instance
func (m Model) handleSSHConnection(msg sshConnectionMsg) (tea.Model, tea.Cmd) {
	// Build SSH command - use system defaults (respects ~/.ssh/config)
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		fmt.Sprintf("%s@%s", msg.user, msg.ip),
	}

	// Log the SSH command to debug panel
	sshCmd := "ssh " + strings.Join(args, " ")
	httpLib.BrowserDebugLogger.AddEntry(httpLib.DebugLogEntry{
		Timestamp: time.Now(),
		Method:    "SSH",
		URL:       sshCmd,
	})

	// Execute SSH in the current terminal
	c := exec.Command("ssh", args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	return m, tea.ExecProcess(c, func(err error) tea.Msg {
		if err != nil {
			// Log SSH error to debug panel
			httpLib.BrowserDebugLogger.AddEntry(httpLib.DebugLogEntry{
				Timestamp: time.Now(),
				Method:    "SSH",
				URL:       sshCmd,
				Error:     err.Error(),
			})
			// SSH exit code 255 means connection error, but other exit codes
			// (including non-zero from user commands) should not be treated as SSH errors
			if exitErr, ok := err.(*exec.ExitError); ok {
				if exitErr.ExitCode() != 255 {
					// User exited with non-zero code, not an SSH error
					return instanceActionMsg{action: "ssh", err: nil}
				}
			}
			return instanceActionMsg{action: "ssh", err: err}
		}
		return instanceActionMsg{action: "ssh", err: nil}
	})
}

// View renders the UI
func (m Model) View() string {
	var content strings.Builder

	// Calculate available width
	width := m.width
	if width < 80 {
		width = 80
	}

	// Header with logo
	content.WriteString(m.renderHeader())
	content.WriteString("\n")

	// Navigation bar
	content.WriteString(m.renderNavBar(width))
	content.WriteString("\n\n")

	// Content box with current product
	content.WriteString(m.renderContentBox(width))
	content.WriteString("\n\n")

	// Footer
	content.WriteString(m.renderFooter())

	return content.String()
}

func (m Model) renderHeader() string {
	logo := logoStyle.Render("☁ OVHcloud Manager")
	// Show selected project in header if one is selected
	if m.cloudProject != "" && m.mode != ProjectSelectView {
		projectInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render(fmt.Sprintf(" • Project: %s", m.cloudProjectName))
		return logo + projectInfo
	}
	return logo
}

func (m Model) renderNavBar(width int) string {
	// Don't show nav bar in project selection mode
	if m.mode == ProjectSelectView || m.currentProduct == ProductProjects {
		return ""
	}

	navItems := getNavItems()
	var items []string

	for i, nav := range navItems {
		var style lipgloss.Style
		if i == m.navIdx {
			style = navItemSelectedStyle
		} else {
			style = navItemStyle
		}
		items = append(items, style.Render(fmt.Sprintf("%s %s", nav.Icon, nav.Label)))
	}

	navContent := lipgloss.JoinHorizontal(lipgloss.Top, items...)
	return navBarStyle.Width(width - 2).Render(navContent)
}

func (m Model) renderContentBox(width int) string {
	var titleText string

	// Handle wizard mode with special title
	if m.mode == WizardView {
		// Determine which wizard we're in based on the step
		if m.wizard.step >= 200 {
			// Node pool wizard
			titleText = " 🔧 Add Node Pool "
		} else if m.wizard.step >= 100 {
			// Kubernetes wizard
			titleText = " ☸️  Create Kubernetes Cluster "
		} else {
			// Instance wizard
			titleText = " 🚀 Create Instance "
		}
		title := productTitleStyle.Render(titleText)
		contentStr := m.renderWizardView(width - 6)
		fullContent := title + "\n\n" + contentStr
		return contentBoxStyle.Width(width - 4).Render(fullContent)
	}

	// Handle debug view with special title
	if m.mode == DebugView {
		titleText = " 🔍 Debug - API Requests "
		title := productTitleStyle.Render(titleText)
		contentStr := m.renderDebugView(width - 6)
		fullContent := title + "\n\n" + contentStr
		return contentBoxStyle.Width(width - 4).Render(fullContent)
	}

	// Handle project selection view specially
	if m.mode == ProjectSelectView || m.currentProduct == ProductProjects {
		titleText = " 📦 Select a Project "
	} else {
		navItems := getNavItems()
		currentNav := navItems[m.navIdx]

		// Product title - show item name in detail view
		if m.mode == DetailView && m.currentItemName != "" {
			titleText = fmt.Sprintf(" %s %s > %s ", currentNav.Icon, currentNav.Label, m.currentItemName)
		} else {
			titleText = fmt.Sprintf(" %s %s ", currentNav.Icon, currentNav.Label)
		}
	}
	title := productTitleStyle.Render(titleText)

	// Content based on mode
	var contentStr string
	switch m.mode {
	case LoadingView:
		contentStr = m.renderLoadingView()
	case ErrorView:
		contentStr = errorStyle.Render("❌ Error: " + m.errorMsg)
	case EmptyView:
		contentStr = m.renderEmptyView()
	case ProjectSelectView:
		contentStr = m.renderTable()
	case TableView:
		contentStr = m.renderTable()
	case DetailView:
		contentStr = m.renderDetailView(width - 6)
	case NodePoolsView:
		contentStr = m.renderNodePoolsView(width - 6)
	case NodePoolDetailView:
		contentStr = m.renderNodePoolDetailView(width - 6)
	case DeleteConfirmView:
		contentStr = m.renderDeleteConfirmView()
	case DebugView:
		contentStr = m.renderDebugView(width - 6)
	case KubeUpgradeView:
		contentStr = m.renderKubeUpgradeView(width - 6)
	case KubePolicyEditView:
		contentStr = m.renderKubePolicyEditView(width - 6)
	case KubeDeleteConfirmView:
		contentStr = m.renderKubeDeleteConfirmView(width - 6)
	case NodePoolScaleView:
		contentStr = m.renderNodePoolScaleView(width - 6)
	case NodePoolDeleteConfirmView:
		contentStr = m.renderNodePoolDeleteConfirmView(width - 6)
	}

	// Combine title and content
	fullContent := title + "\n\n" + contentStr

	return contentBoxStyle.Width(width - 4).Render(fullContent)
}

// renderLoadingView displays the loading screen
// Shows ASCII OVHcloud logo only on initial splash screen (loading projects)
func (m Model) renderLoadingView() string {
	var content strings.Builder

	// Show splash screen with logo only when loading projects initially
	if m.currentProduct == ProductProjects && m.cloudProject == "" {
		// Style for the ASCII logo
		logoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B68EE")).
			Bold(true)

		// Style for the loading message
		loadingMsgStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF7F")).
			Bold(true)

		// Add the ASCII logo
		content.WriteString(logoStyle.Render(ovhcloudASCIILogo))
		content.WriteString("\n\n")

		// Add loading message with spinner
		content.WriteString(loadingMsgStyle.Render("        ⏳ Loading projects..."))
		content.WriteString("\n")
	} else {
		// Simple loading message for other cases
		content.WriteString(loadingStyle.Render("⏳ Loading data..."))
	}

	return content.String()
}

// renderDebugView displays the debug panel with API requests
func (m Model) renderDebugView(width int) string {
	var content strings.Builder

	entries := httpLib.BrowserDebugLogger.GetEntries()

	if len(entries) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)
		content.WriteString(emptyStyle.Render("  No API requests recorded yet.\n"))
		content.WriteString(emptyStyle.Render("  Navigate around to see requests appear here.\n"))
	} else {
		// Header
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B68EE")).
			Bold(true)
		content.WriteString(headerStyle.Render(fmt.Sprintf("  📊 %d API requests recorded\n\n", len(entries))))

		// Styles for different status codes
		successStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
		methodStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
		urlStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		reqIdStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF"))

		// Calculate visible entries based on scroll offset
		maxVisible := 15 // Show last 15 entries by default
		startIdx := len(entries) - maxVisible - m.debugScrollOffset
		if startIdx < 0 {
			startIdx = 0
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(entries) {
			endIdx = len(entries)
		}

		// Show entries in reverse order (newest first)
		for i := endIdx - 1; i >= startIdx; i-- {
			entry := entries[i]

			// Format timestamp
			timestamp := timeStyle.Render(entry.Timestamp.Format("15:04:05"))

			// Format method
			method := methodStyle.Render(fmt.Sprintf("%-6s", entry.Method))

			// Format URL (truncate if too long)
			url := entry.URL
			maxUrlLen := width - 60
			if maxUrlLen < 20 {
				maxUrlLen = 20
			}
			if len(url) > maxUrlLen {
				url = url[:maxUrlLen-3] + "..."
			}
			urlFormatted := urlStyle.Render(url)

			// Format status
			var statusFormatted string
			if entry.Error != "" {
				statusFormatted = errorStyle.Render("ERR")
			} else if entry.Method == "SSH" {
				// SSH commands don't have HTTP status codes
				statusFormatted = successStyle.Render("CMD")
			} else if entry.StatusCode >= 200 && entry.StatusCode < 300 {
				statusFormatted = successStyle.Render(fmt.Sprintf("%d", entry.StatusCode))
			} else if entry.StatusCode >= 400 {
				statusFormatted = errorStyle.Render(fmt.Sprintf("%d", entry.StatusCode))
			} else {
				statusFormatted = warnStyle.Render(fmt.Sprintf("%d", entry.StatusCode))
			}

			// Format duration
			duration := timeStyle.Render(fmt.Sprintf("%6s", entry.Duration.Round(time.Millisecond)))

			// Format request ID (display full ID without truncation)
			reqId := "-"
			if entry.RequestID != "" {
				reqId = entry.RequestID
			}
			reqIdFormatted := reqIdStyle.Render(reqId)

			content.WriteString(fmt.Sprintf("  %s %s %s → %s %s\n", timestamp, method, urlFormatted, statusFormatted, duration))

			// Show query string if present
			if entry.QueryString != "" {
				queryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
				content.WriteString(fmt.Sprintf("           Query: %s\n", queryStyle.Render(entry.QueryString)))
			}

			content.WriteString(fmt.Sprintf("           RequestID: %s\n\n", reqIdFormatted))
		}

		// Scroll indicator
		if len(entries) > maxVisible {
			scrollInfo := timeStyle.Render(fmt.Sprintf("  Showing %d-%d of %d (↑↓ to scroll, 'c' to clear)", startIdx+1, endIdx, len(entries)))
			content.WriteString(scrollInfo)
		}
	}

	// Help text
	content.WriteString("\n\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	content.WriteString(helpStyle.Render("  Press 'd' or Esc to close • 'c' to clear logs"))

	return content.String()
}

// renderEmptyView displays an empty state with creation prompt
func (m Model) renderEmptyView() string {
	var content strings.Builder

	// Get product-specific info
	productName, createCmd := m.getProductCreationInfo()

	emptyIcon := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render("📭")

	emptyText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render(fmt.Sprintf("No %s found in this project", productName))

	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("        %s\n\n", emptyIcon))
	content.WriteString(fmt.Sprintf("        %s\n\n", emptyText))

	if createCmd != "" {
		promptStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF7F")).
			Bold(true)

		cmdStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B68EE")).
			Background(lipgloss.Color("#1a1a1a")).
			Padding(0, 1)

		content.WriteString(fmt.Sprintf("        %s\n\n", promptStyle.Render("Press 'c' to create one, or run:")))
		content.WriteString(fmt.Sprintf("        %s\n", cmdStyle.Render(createCmd)))
	}

	return content.String()
}

// renderNodePoolsView displays the node pools management view
func (m Model) renderNodePoolsView(width int) string {
	var content strings.Builder

	if m.detailData == nil {
		return "No cluster selected"
	}

	clusterName := getStringValue(m.detailData, "name", "Unknown")
	clusterId := getStringValue(m.detailData, "id", "")

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B68EE")).
		Bold(true)
	content.WriteString(headerStyle.Render(fmt.Sprintf("  Node Pools - Cluster: %s\n\n", clusterName)))

	// Get node pools for this cluster
	nodePools := m.kubeNodePools[clusterId]

	if len(nodePools) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true)
		content.WriteString(emptyStyle.Render("  No node pools found.\n"))
		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF7F")).
			Bold(true).
			Render("  Press 'c' to create a node pool\n"))
	} else {
		// Format node pools as a simple table without lipgloss for alignment
		content.WriteString("\n")

		// Header
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7B68EE")).
			Bold(true)

		header := "  Name                  Status      Flavor           Nodes       Autoscale"
		content.WriteString(headerStyle.Render(header) + "\n")

		separator := "  " + strings.Repeat("─", 75)
		content.WriteString(headerStyle.Render(separator) + "\n")

		// Rows
		selectedStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#000000")).
			Background(lipgloss.Color("#7B68EE"))

		for i, pool := range nodePools {
			poolName := getStringValue(pool, "name", "Unknown")
			poolStatus := getStringValue(pool, "status", "Unknown")
			flavor := getStringValue(pool, "flavor", "N/A")
			desiredNodes := getIntOrFloatValue(pool, "desiredNodes", 0)
			currentNodes := getIntOrFloatValue(pool, "currentNodes", 0)
			autoscale := getBoolValue(pool, "autoscale", false)
			minNodes := getIntOrFloatValue(pool, "minNodes", 0)
			maxNodes := getIntOrFloatValue(pool, "maxNodes", 0)

			// Truncate and pad to exact width
			nameField := poolName
			if len(nameField) > 20 {
				nameField = nameField[:20]
			}
			for len(nameField) < 20 {
				nameField += " "
			}

			statusField := poolStatus
			if len(statusField) > 10 {
				statusField = statusField[:10]
			}
			for len(statusField) < 10 {
				statusField += " "
			}

			flavorField := flavor
			if len(flavorField) > 15 {
				flavorField = flavorField[:15]
			}
			for len(flavorField) < 15 {
				flavorField += " "
			}

			nodesStr := fmt.Sprintf("%.0f/%.0f", currentNodes, desiredNodes)
			nodesField := nodesStr
			for len(nodesField) < 10 {
				nodesField += " "
			}

			autoscaleStr := "No"
			if autoscale {
				autoscaleStr = fmt.Sprintf("%.0f-%.0f", minNodes, maxNodes)
			}

			line := "  " + nameField + "  " + statusField + "  " + flavorField + "  " + nodesField + "  " + autoscaleStr

			// Highlight selected row
			if i == m.nodePoolsSelectedIdx {
				content.WriteString(selectedStyle.Render(line) + "\n")
			} else {
				content.WriteString(line + "\n")
			}
		}
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render("  Press 'c' to create a node pool, Enter to view details, ↑/↓ to navigate, Escape to go back\n"))
	}

	return content.String()
}

// handleSwitchToNodePoolsView handles switching to node pools management view
func (m Model) handleSwitchToNodePoolsView(msg switchToNodePoolsViewMsg) (tea.Model, tea.Cmd) {
	m.mode = NodePoolsView
	m.nodePoolsSelectedIdx = 0 // Reset selection
	m.wizard.nodePoolClusterId = msg.clusterId
	return m, nil
}

// renderNodePoolDetailView displays detailed information about a single node pool
func (m Model) renderNodePoolDetailView(width int) string {
	var content strings.Builder

	if m.selectedNodePool == nil {
		return "No node pool selected"
	}

	// Get cluster info
	clusterName := getStringValue(m.detailData, "name", "Unknown")

	// Get node pool data
	poolName := getStringValue(m.selectedNodePool, "name", "Unknown")
	poolId := getStringValue(m.selectedNodePool, "id", "N/A")
	poolStatus := getStringValue(m.selectedNodePool, "status", "Unknown")
	flavor := getStringValue(m.selectedNodePool, "flavor", "N/A")
	desiredNodes := getIntOrFloatValue(m.selectedNodePool, "desiredNodes", 0)
	currentNodes := getIntOrFloatValue(m.selectedNodePool, "currentNodes", 0)
	availableNodes := getIntOrFloatValue(m.selectedNodePool, "availableNodes", 0)
	upToDateNodes := getIntOrFloatValue(m.selectedNodePool, "upToDateNodes", 0)
	autoscale := getBoolValue(m.selectedNodePool, "autoscale", false)
	minNodes := getIntOrFloatValue(m.selectedNodePool, "minNodes", 0)
	maxNodes := getIntOrFloatValue(m.selectedNodePool, "maxNodes", 0)
	antiAffinity := getBoolValue(m.selectedNodePool, "antiAffinity", false)
	monthlyBilled := getBoolValue(m.selectedNodePool, "monthlyBilled", false)
	createdAt := getStringValue(m.selectedNodePool, "createdAt", "N/A")
	updatedAt := getStringValue(m.selectedNodePool, "updatedAt", "N/A")

	// Header
	headerLabelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B68EE")).
		Bold(true).
		Width(15).
		Align(lipgloss.Left)
	headerValueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true)

	content.WriteString(fmt.Sprintf("  %s %s\n", headerLabelStyle.Render("Node Pool:"), headerValueStyle.Render(poolName)))
	content.WriteString(fmt.Sprintf("  %s %s\n\n", headerLabelStyle.Render("Cluster:"), headerValueStyle.Render(clusterName)))

	// Actions with selection highlighting
	actions := []string{"Scale", "Delete"}
	var actionParts []string
	for i, action := range actions {
		if i == m.nodePoolDetailActionIdx {
			// Selected action - highlighted
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color("#7B68EE")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 1).
				Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Padding(0, 1).
				Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.nodePoolDetailConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actions[m.nodePoolDetailActionIdx]))
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)
	content.WriteString(actionsBox)
	content.WriteString("\n\n")

	// Create styled sections
	labelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Width(20).
		Align(lipgloss.Left)
	valueStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	successStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF7F"))
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500"))
	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	// Status with color coding
	statusStyle := valueStyle
	switch poolStatus {
	case "READY":
		statusStyle = successStyle
	case "INSTALLING", "UPDATING", "REDEPLOYING":
		statusStyle = warningStyle
	case "ERROR", "DELETING":
		statusStyle = errorStyle
	}

	// Basic Information
	sectionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B68EE")).
		Bold(true)
	content.WriteString(sectionStyle.Render("  Basic Information\n"))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("ID:"), valueStyle.Render(poolId)))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Status:"), statusStyle.Render(poolStatus)))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Flavor:"), valueStyle.Render(flavor)))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Created:"), valueStyle.Render(createdAt)))
	content.WriteString(fmt.Sprintf("  %s %s\n\n", labelStyle.Render("Updated:"), valueStyle.Render(updatedAt)))

	// Node Counts
	content.WriteString(sectionStyle.Render("  Node Counts\n"))
	content.WriteString(fmt.Sprintf("  %s %.0f\n", labelStyle.Render("Desired Nodes:"), desiredNodes))
	content.WriteString(fmt.Sprintf("  %s %.0f\n", labelStyle.Render("Current Nodes:"), currentNodes))
	content.WriteString(fmt.Sprintf("  %s %.0f\n", labelStyle.Render("Available Nodes:"), availableNodes))
	content.WriteString(fmt.Sprintf("  %s %.0f\n\n", labelStyle.Render("Up-to-Date Nodes:"), upToDateNodes))

	// Configuration
	content.WriteString(sectionStyle.Render("  Configuration\n"))

	// Autoscale
	autoscaleStr := "Disabled"
	if autoscale {
		autoscaleStr = fmt.Sprintf("Enabled (%.0f - %.0f nodes)", minNodes, maxNodes)
	}
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Autoscale:"), valueStyle.Render(autoscaleStr)))

	// Anti-affinity
	antiAffinityStr := "Disabled"
	if antiAffinity {
		antiAffinityStr = "Enabled"
	}
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Anti-Affinity:"), valueStyle.Render(antiAffinityStr)))

	// Billing
	billingStr := "Hourly"
	if monthlyBilled {
		billingStr = "Monthly"
	}
	content.WriteString(fmt.Sprintf("  %s %s\n\n", labelStyle.Render("Billing:"), valueStyle.Render(billingStr)))

	// Help text
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString(dimStyle.Render("  ←/→ Navigate actions • Enter Execute • Escape Go back\n"))

	return content.String()
}

// renderKubeUpgradeView displays the Kubernetes upgrade version selection
func (m Model) renderKubeUpgradeView(width int) string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B68EE")).
		Bold(true)
	content.WriteString(headerStyle.Render("  ⬆️  Upgrade Kubernetes Cluster\n\n"))

	// Show loading if still loading
	if m.wizard.isLoading {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			Render("  " + m.wizard.loadingMessage + "\n"))
		return content.String()
	}

	// Show cluster info
	clusterName := getStringValue(m.detailData, "name", "Unknown")
	currentVersion := getStringValue(m.detailData, "version", "Unknown")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Cluster:"), valueStyle.Render(clusterName)))
	content.WriteString(fmt.Sprintf("  %s %s\n\n", labelStyle.Render("Current version:"), valueStyle.Render(currentVersion)))

	// Show available versions
	content.WriteString(headerStyle.Render("  Select target version:\n\n"))

	if len(m.wizard.kubeUpgradeVersions) == 0 {
		content.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Render("  No upgrade versions available. Cluster is up to date.\n"))
	} else {
		selectedStyle := lipgloss.NewStyle().
			Background(lipgloss.Color("#7B68EE")).
			Foreground(lipgloss.Color("#FFFFFF"))
		normalStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF"))

		for i, version := range m.wizard.kubeUpgradeVersions {
			if i == m.wizard.kubeUpgradeSelectedIdx {
				content.WriteString(selectedStyle.Render(fmt.Sprintf("  > %s\n", version)))
			} else {
				content.WriteString(normalStyle.Render(fmt.Sprintf("    %s\n", version)))
			}
		}
	}

	// Help text
	content.WriteString("\n")
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString(dimStyle.Render("  ↑/↓ to navigate • Enter to upgrade • Escape to cancel\n"))

	return content.String()
}

// renderKubePolicyEditView displays the update policy selection
func (m Model) renderKubePolicyEditView(width int) string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B68EE")).
		Bold(true)
	content.WriteString(headerStyle.Render("  ⚙️  Edit Update Policy\n\n"))

	// Show cluster info
	clusterName := getStringValue(m.detailData, "name", "Unknown")
	currentPolicy := getStringValue(m.detailData, "updatePolicy", "Unknown")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Cluster:"), valueStyle.Render(clusterName)))
	content.WriteString(fmt.Sprintf("  %s %s\n\n", labelStyle.Render("Current policy:"), valueStyle.Render(currentPolicy)))

	// Show policy options
	content.WriteString(headerStyle.Render("  Select new policy:\n\n"))

	policies := []struct {
		name        string
		description string
	}{
		{"ALWAYS_UPDATE", "Always update to latest version automatically"},
		{"MINIMAL_DOWNTIME", "Update with minimal service disruption"},
		{"NEVER_UPDATE", "Never auto-update, manual updates only"},
	}

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7B68EE")).
		Foreground(lipgloss.Color("#FFFFFF"))
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))
	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Italic(true)

	for i, policy := range policies {
		if i == m.wizard.kubePolicySelectedIdx {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  > %s\n", policy.name)))
		} else {
			content.WriteString(normalStyle.Render(fmt.Sprintf("    %s\n", policy.name)))
		}
		content.WriteString(descStyle.Render(fmt.Sprintf("      %s\n", policy.description)))
	}

	// Help text
	content.WriteString("\n")
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString(dimStyle.Render("  ↑/↓ to navigate • Enter to apply • Escape to cancel\n"))

	return content.String()
}

// renderKubeDeleteConfirmView displays the cluster deletion confirmation
func (m Model) renderKubeDeleteConfirmView(width int) string {
	var content strings.Builder

	// Header with warning
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)
	content.WriteString(headerStyle.Render("  ⚠️  Delete Kubernetes Cluster\n\n"))

	// Warning message
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700"))
	content.WriteString(warningStyle.Render("  This action is IRREVERSIBLE!\n"))
	content.WriteString(warningStyle.Render("  All data, node pools, and configurations will be permanently deleted.\n\n"))

	// Show cluster info
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Cluster to delete:"), valueStyle.Render(m.wizard.kubeDeleteClusterName)))
	content.WriteString(fmt.Sprintf("  %s %s\n\n", labelStyle.Render("Cluster ID:"), valueStyle.Render(m.wizard.kubeDeleteClusterId)))

	// Confirmation input
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Render(fmt.Sprintf("  Type '%s' to confirm deletion:\n\n", m.wizard.kubeDeleteClusterName)))

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF7F"))
	content.WriteString(inputStyle.Render(fmt.Sprintf("  > %s▌\n", m.wizard.kubeDeleteConfirmInput)))

	// Help text
	content.WriteString("\n")
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString(dimStyle.Render("  Enter to delete (when name matches) • Escape to cancel\n"))

	return content.String()
}

// renderNodePoolScaleView displays the node pool scale interface
func (m Model) renderNodePoolScaleView(width int) string {
	var content strings.Builder

	// Header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B68EE")).
		Bold(true)
	content.WriteString(headerStyle.Render("  📈 Scale Node Pool\n\n"))

	// Show pool info
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(fmt.Sprintf("  %s %s\n\n", labelStyle.Render("Node Pool:"), valueStyle.Render(m.wizard.nodePoolScalePoolName)))

	// Fields
	fields := []struct {
		label    string
		value    int
		selected bool
	}{
		{"Desired Nodes:", m.wizard.nodePoolScaleDesired, m.wizard.nodePoolScaleFieldIdx == 0},
		{"Min Nodes:", m.wizard.nodePoolScaleMin, m.wizard.nodePoolScaleFieldIdx == 1},
		{"Max Nodes:", m.wizard.nodePoolScaleMax, m.wizard.nodePoolScaleFieldIdx == 2},
	}

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7B68EE")).
		Foreground(lipgloss.Color("#FFFFFF"))
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	for _, field := range fields {
		if field.selected {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  > %s %d", field.label, field.value)) + "\n")
		} else {
			content.WriteString(normalStyle.Render(fmt.Sprintf("    %s %d", field.label, field.value)) + "\n")
		}
	}

	// Autoscale toggle
	content.WriteString("\n")
	autoscaleLabel := "Autoscale:"
	autoscaleValue := "Disabled"
	if m.wizard.nodePoolScaleAutoscale {
		autoscaleValue = "Enabled"
	}
	if m.wizard.nodePoolScaleFieldIdx == 3 {
		content.WriteString(selectedStyle.Render(fmt.Sprintf("  > %s %s", autoscaleLabel, autoscaleValue)) + "\n")
	} else {
		content.WriteString(normalStyle.Render(fmt.Sprintf("    %s %s", autoscaleLabel, autoscaleValue)) + "\n")
	}

	// Buttons
	content.WriteString("\n")
	cancelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(0, 1)
	applyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)

	if m.wizard.nodePoolScaleFieldIdx == 4 {
		content.WriteString(selectedStyle.Render("  [Cancel]") + "  " + applyStyle.Render("[Apply]") + "\n")
	} else if m.wizard.nodePoolScaleFieldIdx == 5 {
		content.WriteString(cancelStyle.Render("  [Cancel]") + "  " + selectedStyle.Render("[Apply]") + "\n")
	} else {
		content.WriteString(cancelStyle.Render("  [Cancel]") + "  " + applyStyle.Render("[Apply]") + "\n")
	}

	// Help text
	content.WriteString("\n")
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString(dimStyle.Render("  ↑/↓ Navigate • +/- Adjust value • Space Toggle autoscale • Enter Apply\n"))

	return content.String()
}

// renderNodePoolDeleteConfirmView displays the node pool deletion confirmation
func (m Model) renderNodePoolDeleteConfirmView(width int) string {
	var content strings.Builder

	// Header with warning
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B")).
		Bold(true)
	content.WriteString(headerStyle.Render("  ⚠️  Delete Node Pool\n\n"))

	// Warning message
	warningStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFD700"))
	content.WriteString(warningStyle.Render("  This action will delete all nodes in this pool!\n\n"))

	// Show pool info
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Node Pool to delete:"), valueStyle.Render(m.wizard.nodePoolDeletePoolName)))
	content.WriteString(fmt.Sprintf("  %s %s\n\n", labelStyle.Render("Pool ID:"), valueStyle.Render(m.wizard.nodePoolDeletePoolId)))

	// Confirmation input
	content.WriteString(lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Render(fmt.Sprintf("  Type '%s' to confirm deletion:\n\n", m.wizard.nodePoolDeletePoolName)))

	inputStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF7F"))
	content.WriteString(inputStyle.Render(fmt.Sprintf("  > %s▌\n", m.wizard.nodePoolDeleteConfirmInput)))

	// Help text
	content.WriteString("\n")
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString(dimStyle.Render("  Enter to delete (when name matches) • Escape to cancel\n"))

	return content.String()
}

// renderWizardView displays the multi-step creation wizard
func (m Model) renderWizardView(width int) string {
	var content strings.Builder

	// Check if we're in cleanup confirmation mode
	if m.wizard.cleanupPending {
		return m.renderCleanupConfirmation(width)
	}

	// Progress indicator - build steps dynamically based on configuration
	var steps []string
	var stepMapping []WizardStep // Maps display index to actual step

	// Build steps based on which wizard we're in (determine by first step >= 100)
	if m.wizard.step >= 200 {
		// Node pool wizard
		steps = append(steps, "Flavor", "Name", "Size", "Options", "Confirm")
		stepMapping = append(stepMapping, NodePoolWizardStepFlavor, NodePoolWizardStepName, NodePoolWizardStepSize, NodePoolWizardStepOptions, NodePoolWizardStepConfirm)
	} else if m.wizard.step >= 100 {
		// Kubernetes wizard
		steps = append(steps, "Region", "Version", "Network", "Name", "Options", "Confirm")
		stepMapping = append(stepMapping, KubeWizardStepRegion, KubeWizardStepVersion, KubeWizardStepNetwork, KubeWizardStepName, KubeWizardStepOptions, KubeWizardStepConfirm)
	} else {
		// Instance wizard
		steps = append(steps, "Region", "Flavor", "Image", "SSH Key", "Network")
		stepMapping = append(stepMapping, WizardStepRegion, WizardStepFlavor, WizardStepImage, WizardStepSSHKey, WizardStepNetwork)

		// Add Floating IP step if private network without public network
		if !m.wizard.usePublicNetwork && m.wizard.selectedPrivateNetwork != "" {
			steps = append(steps, "Floating IP")
			stepMapping = append(stepMapping, WizardStepFloatingIP)
		}

		steps = append(steps, "Name", "Confirm")
		stepMapping = append(stepMapping, WizardStepName, WizardStepConfirm)
	}

	// Find current step index in the display
	currentStepIdx := 0
	for i, step := range stepMapping {
		if step == m.wizard.step {
			currentStepIdx = i
			break
		}
		// Handle case where we're at FloatingIP but it's not in the list
		if m.wizard.step == WizardStepFloatingIP && step == WizardStepNetwork {
			currentStepIdx = i + 1
			break
		}
	}

	progressStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	activeStepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	completedStepStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))

	var progressParts []string
	for i, step := range steps {
		var stepStr string
		if i < currentStepIdx {
			stepStr = completedStepStyle.Render(fmt.Sprintf("✓ %s", step))
		} else if i == currentStepIdx {
			stepStr = activeStepStyle.Render(fmt.Sprintf("● %s", step))
		} else {
			stepStr = progressStyle.Render(fmt.Sprintf("○ %s", step))
		}
		progressParts = append(progressParts, stepStr)
	}
	progress := strings.Join(progressParts, progressStyle.Render(" → "))
	content.WriteString(progress + "\n\n")

	// Loading state
	if m.wizard.isLoading {
		msg := "Loading..."
		if m.wizard.loadingMessage != "" {
			msg = m.wizard.loadingMessage
		}
		content.WriteString(loadingStyle.Render("⏳ " + msg))
		return content.String()
	}

	// Render current step (step render functions handle displaying errors in context)
	switch m.wizard.step {
	case WizardStepRegion:
		content.WriteString(m.renderWizardRegionStep(width))
	case WizardStepFlavor:
		content.WriteString(m.renderWizardFlavorStep(width))
	case WizardStepImage:
		content.WriteString(m.renderWizardImageStep(width))
	case WizardStepSSHKey:
		content.WriteString(m.renderWizardSSHKeyStep(width))
	case WizardStepNetwork:
		content.WriteString(m.renderWizardNetworkStep(width))
	case WizardStepFloatingIP:
		content.WriteString(m.renderWizardFloatingIPStep(width))
	case WizardStepName:
		content.WriteString(m.renderWizardNameStep(width))
	case WizardStepConfirm:
		content.WriteString(m.renderWizardConfirmStep(width))
	// Kubernetes wizard steps
	case KubeWizardStepRegion:
		content.WriteString(m.renderKubeWizardRegionStep(width))
	case KubeWizardStepVersion:
		content.WriteString(m.renderKubeWizardVersionStep(width))
	case KubeWizardStepNetwork:
		content.WriteString(m.renderKubeWizardNetworkStep(width))
	case KubeWizardStepSubnet:
		content.WriteString(m.renderKubeWizardSubnetStep(width))
	case KubeWizardStepName:
		content.WriteString(m.renderKubeWizardNameStep(width))
	case KubeWizardStepOptions:
		content.WriteString(m.renderKubeWizardOptionsStep(width))
	case KubeWizardStepConfirm:
		content.WriteString(m.renderKubeWizardConfirmStep(width))
	// Node pool wizard steps
	case NodePoolWizardStepFlavor:
		content.WriteString(m.renderNodePoolWizardFlavorStep(width))
	case NodePoolWizardStepName:
		content.WriteString(m.renderNodePoolWizardNameStep(width))
	case NodePoolWizardStepSize:
		content.WriteString(m.renderNodePoolWizardSizeStep(width))
	case NodePoolWizardStepOptions:
		content.WriteString(m.renderNodePoolWizardOptionsStep(width))
	case NodePoolWizardStepConfirm:
		content.WriteString(m.renderNodePoolWizardConfirmStep(width))
	}

	return content.String()
}

// renderWizardRegionStep renders the region selection step
func (m Model) renderWizardRegionStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select a region:") + "\n")

	// Show filter input if active
	if m.wizard.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s▌", m.wizard.filterInput)) + "\n\n")
	} else if m.wizard.filterInput != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s (press / to edit)", m.wizard.filterInput)) + "\n\n")
	} else {
		content.WriteString("\n")
	}

	filtered := m.getFilteredWizardRegions()
	if len(filtered) == 0 {
		return content.String() + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No regions match filter")
	}

	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))

	// Determine visible range (max 10 items visible)
	maxVisible := 10
	startIdx := 0
	if m.wizard.selectedIndex >= maxVisible {
		startIdx = m.wizard.selectedIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}

	for i := startIdx; i < endIdx; i++ {
		region := filtered[i]
		regionName := getString(region, "name")

		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ "+regionName) + "\n")
		} else {
			content.WriteString(itemStyle.Render("  "+regionName) + "\n")
		}
	}

	// Show scroll indicator if needed
	if len(filtered) > maxVisible {
		scrollInfo := fmt.Sprintf("\n  Showing %d-%d of %d regions", startIdx+1, endIdx, len(filtered))
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(scrollInfo))
	}

	return content.String()
}

// renderWizardFlavorStep renders the flavor selection step
func (m Model) renderWizardFlavorStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render(fmt.Sprintf("Select a flavor (Region: %s):", m.wizard.selectedRegion)) + "\n")

	// Show filter input if active
	if m.wizard.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s▌", m.wizard.filterInput)) + "\n\n")
	} else if m.wizard.filterInput != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s (press / to edit)", m.wizard.filterInput)) + "\n\n")
	} else {
		content.WriteString("\n")
	}

	filtered := m.getFilteredWizardFlavors()
	if len(filtered) == 0 {
		return content.String() + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No flavors match filter")
	}

	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))

	// Determine visible range
	maxVisible := 10
	startIdx := 0
	if m.wizard.selectedIndex >= maxVisible {
		startIdx = m.wizard.selectedIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}

	for i := startIdx; i < endIdx; i++ {
		flavor := filtered[i]
		name := getString(flavor, "name")

		// Get numeric values - HTTP client uses json.Number
		vcpus := getNumericValue(flavor, "vcpus")
		ram := getNumericValue(flavor, "ram")
		disk := getNumericValue(flavor, "disk")
		osType := getString(flavor, "osType")

		// Format disk display - flex flavors have disk=0 (they use attached volumes)
		var diskStr string
		if disk > 0 {
			diskStr = fmt.Sprintf("%5.0f GB", disk)
		} else {
			diskStr = " Volume" // Flex flavors use attached block storage
		}

		displayStr := fmt.Sprintf("%-22s  %2.0f vCPUs  %5.0f GB RAM  %s Disk  [%s]", name, vcpus, ram, diskStr, osType)

		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ "+displayStr) + "\n")
		} else {
			content.WriteString(itemStyle.Render("  "+displayStr) + "\n")
		}
	}

	// Show scroll indicator if needed
	if len(filtered) > maxVisible {
		scrollInfo := fmt.Sprintf("\n  Showing %d-%d of %d flavors", startIdx+1, endIdx, len(filtered))
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(scrollInfo))
	}

	return content.String()
}

// renderWizardImageStep renders the image selection step
func (m Model) renderWizardImageStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render(fmt.Sprintf("Select an image (Region: %s):", m.wizard.selectedRegion)) + "\n")

	// Show filter input if active
	if m.wizard.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s▌", m.wizard.filterInput)) + "\n\n")
	} else if m.wizard.filterInput != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s (press / to edit)", m.wizard.filterInput)) + "\n\n")
	} else {
		content.WriteString("\n")
	}

	filtered := m.getFilteredWizardImages()
	if len(filtered) == 0 {
		return content.String() + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No images match filter")
	}

	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))

	// Determine visible range
	maxVisible := 10
	startIdx := 0
	if m.wizard.selectedIndex >= maxVisible {
		startIdx = m.wizard.selectedIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}

	for i := startIdx; i < endIdx; i++ {
		image := filtered[i]
		name := getString(image, "name")
		imageType := getString(image, "type")

		displayStr := fmt.Sprintf("%-45s  [%s]", truncate(name, 45), imageType)

		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ "+displayStr) + "\n")
		} else {
			content.WriteString(itemStyle.Render("  "+displayStr) + "\n")
		}
	}

	// Show scroll indicator if needed
	if len(filtered) > maxVisible {
		scrollInfo := fmt.Sprintf("\n  Showing %d-%d of %d images", startIdx+1, endIdx, len(filtered))
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(scrollInfo))
	}

	return content.String()
}

// renderWizardSSHKeyStep renders the SSH key selection step
func (m Model) renderWizardSSHKeyStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))
	createKeyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)

	// If in SSH key creation mode, show the creation form
	if m.wizard.creatingSSHKey {
		content.WriteString(titleStyle.Render("Create new SSH key:") + "\n\n")

		inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		activeInputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)

		// Name field (field 0)
		nameLabel := "  Key name: "
		nameValue := m.wizard.newSSHKeyName
		if nameValue == "" {
			nameValue = "_"
		}
		if m.wizard.sshKeyCreateField == 0 {
			content.WriteString(activeInputStyle.Render(nameLabel) + activeInputStyle.Render("["+nameValue+"]") + "\n")
		} else {
			content.WriteString(labelStyle.Render(nameLabel) + inputStyle.Render(nameValue) + "\n")
		}

		// Public key selection (field 1)
		content.WriteString("\n")
		pubKeyLabel := "  Public key: "
		if m.wizard.sshKeyCreateField == 1 {
			content.WriteString(activeInputStyle.Render(pubKeyLabel) + "\n")
		} else {
			content.WriteString(labelStyle.Render(pubKeyLabel) + "\n")
		}

		// List local .pub files
		if len(m.wizard.localPubKeys) > 0 {
			for i, pubKey := range m.wizard.localPubKeys {
				if m.wizard.sshKeyCreateField == 1 && m.wizard.selectedLocalKeyIdx == i {
					content.WriteString(selectedStyle.Render("    ▶ 📄 "+pubKey) + "\n")
				} else {
					content.WriteString(itemStyle.Render("      📄 "+pubKey) + "\n")
				}
			}
		} else {
			content.WriteString(itemStyle.Render("      (no .pub files found in ~/.ssh)") + "\n")
		}

		content.WriteString("\n")

		// Buttons (field 2 = Create, field 3 = Cancel)
		buttonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		activeButtonStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)

		if m.wizard.sshKeyCreateField == 2 {
			content.WriteString("  " + activeButtonStyle.Render("[Create]") + "  " + buttonStyle.Render("[Cancel]") + "\n")
		} else if m.wizard.sshKeyCreateField == 3 {
			content.WriteString("  " + buttonStyle.Render("[Create]") + "  " + activeButtonStyle.Render("[Cancel]") + "\n")
		} else {
			content.WriteString("  " + buttonStyle.Render("[Create]  [Cancel]") + "\n")
		}

		content.WriteString("\n")
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(hintStyle.Render("  Tab/↑↓: Navigate • Enter: Select/Confirm • Esc: Cancel"))

		return content.String()
	}

	content.WriteString(titleStyle.Render("Select an SSH key:") + "\n")

	// Show filter input if active
	if m.wizard.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s▌", m.wizard.filterInput)) + "\n\n")
	} else if m.wizard.filterInput != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s (press / to edit)", m.wizard.filterInput)) + "\n\n")
	} else {
		content.WriteString("\n")
	}

	// Show "Create new key" option first
	if m.wizard.selectedIndex == 0 {
		content.WriteString(selectedStyle.Render("▶ ➕ Create new SSH key") + "\n")
	} else {
		content.WriteString(createKeyStyle.Render("  ➕ Create new SSH key") + "\n")
	}

	// No key option
	if m.wizard.selectedIndex == 1 {
		content.WriteString(selectedStyle.Render("▶ 🚫 No SSH key") + "\n")
	} else {
		content.WriteString(itemStyle.Render("  🚫 No SSH key") + "\n")
	}

	content.WriteString("\n")

	filtered := m.getFilteredWizardSSHKeys()
	if len(filtered) == 0 && m.wizard.filterInput != "" {
		return content.String() + lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("  No SSH keys match filter")
	}

	// Determine visible range (offset by 2 for create and no-key options)
	maxVisible := 8
	listStartIdx := 2 // First SSH key is at index 2
	startIdx := 0
	effectiveIdx := m.wizard.selectedIndex - listStartIdx
	if effectiveIdx >= maxVisible {
		startIdx = effectiveIdx - maxVisible + 1
	}
	if startIdx < 0 {
		startIdx = 0
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}

	for i := startIdx; i < endIdx; i++ {
		sshKey := filtered[i]
		name := getString(sshKey, "name")
		displayIdx := i + listStartIdx // Actual index in the full list

		if displayIdx == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ 🔑 "+name) + "\n")
		} else {
			content.WriteString(itemStyle.Render("  🔑 "+name) + "\n")
		}
	}

	// Show scroll indicator if needed
	if len(filtered) > maxVisible {
		scrollInfo := fmt.Sprintf("\n  Showing %d-%d of %d SSH keys", startIdx+1, endIdx, len(filtered))
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(scrollInfo))
	}

	return content.String()
}

// renderWizardNetworkStep renders the network configuration step
func (m Model) renderWizardNetworkStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)

	// If in network creation mode, show the creation form
	if m.wizard.creatingNetwork {
		content.WriteString(titleStyle.Render("Create new private network:") + "\n\n")

		inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
		activeInputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)

		// Name field (field 0)
		nameLabel := "  Network name: "
		nameValue := m.wizard.newNetworkName
		if nameValue == "" {
			nameValue = "_"
		}
		if m.wizard.networkCreateField == 0 {
			content.WriteString(activeInputStyle.Render(nameLabel) + activeInputStyle.Render("["+nameValue+"]") + "\n")
		} else {
			content.WriteString(labelStyle.Render(nameLabel) + inputStyle.Render(nameValue) + "\n")
		}

		// VLAN ID field (field 1)
		vlanLabel := "  VLAN ID:      "
		vlanValue := fmt.Sprintf("%d", m.wizard.newNetworkVlanId)
		if m.wizard.networkCreateField == 1 {
			content.WriteString(activeInputStyle.Render(vlanLabel) + activeInputStyle.Render("["+vlanValue+"]") + "\n")
		} else {
			content.WriteString(labelStyle.Render(vlanLabel) + inputStyle.Render(vlanValue) + "\n")
		}

		// CIDR field (field 2)
		cidrLabel := "  Subnet CIDR:  "
		cidrValue := m.wizard.newNetworkCIDR
		if cidrValue == "" {
			cidrValue = "10.0.0.0/24"
		}
		if m.wizard.networkCreateField == 2 {
			content.WriteString(activeInputStyle.Render(cidrLabel) + activeInputStyle.Render("["+cidrValue+"]") + "\n")
		} else {
			content.WriteString(labelStyle.Render(cidrLabel) + inputStyle.Render(cidrValue) + "\n")
		}

		// DHCP toggle (field 3)
		dhcpStatus := "[ ]"
		if m.wizard.newNetworkDHCP {
			dhcpStatus = "[✓]"
		}
		dhcpLine := fmt.Sprintf("  Enable DHCP:  %s", dhcpStatus)
		if m.wizard.networkCreateField == 3 {
			content.WriteString(activeInputStyle.Render(dhcpLine) + "\n")
		} else {
			content.WriteString(itemStyle.Render(dhcpLine) + "\n")
		}

		content.WriteString("\n")

		// Create button (field 4)
		if m.wizard.networkCreateField == 4 {
			content.WriteString(selectedStyle.Render("  ▶ [Create Network]") + "\n")
		} else {
			content.WriteString(itemStyle.Render("    [Create Network]") + "\n")
		}

		content.WriteString("\n")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  ↑↓/Tab: Navigate • Space: Toggle DHCP • Enter: Create • Esc: Cancel"))

		return content.String()
	}

	// Normal network selection view
	content.WriteString(titleStyle.Render("Configure network:") + "\n\n")

	// Public network toggle
	publicStatus := "[ ]"
	if m.wizard.usePublicNetwork {
		publicStatus = "[✓]"
	}
	publicLine := fmt.Sprintf("%s Public Network (Internet access)", publicStatus)
	if m.wizard.networkMenuIndex == 0 {
		content.WriteString(selectedStyle.Render("▶ "+publicLine) + "\n")
	} else {
		content.WriteString(itemStyle.Render("  "+publicLine) + "\n")
	}

	content.WriteString("\n")
	content.WriteString(labelStyle.Render("  Private Network:") + "\n")

	// Show filter input if active (only for private network list)
	if m.wizard.filterMode && m.wizard.networkMenuIndex == 1 {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s▌", m.wizard.filterInput)) + "\n")
	} else if m.wizard.filterInput != "" && m.wizard.networkMenuIndex == 1 {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s (press / to edit)", m.wizard.filterInput)) + "\n")
	}

	filtered := m.getFilteredWizardNetworks()
	if len(filtered) == 0 {
		content.WriteString(itemStyle.Render("  No private networks match filter") + "\n")
	} else {
		// Determine visible range
		maxVisible := 8
		startIdx := 0
		if m.wizard.networkMenuIndex == 1 && m.wizard.selectedIndex >= maxVisible {
			startIdx = m.wizard.selectedIndex - maxVisible + 1
		}
		endIdx := startIdx + maxVisible
		if endIdx > len(filtered) {
			endIdx = len(filtered)
		}

		for i := startIdx; i < endIdx; i++ {
			network := filtered[i]
			name := getString(network, "name")
			networkId := getString(network, "id")

			// Just show network name without subnet details
			subnetInfo := ""

			isSelected := m.wizard.networkMenuIndex == 1 && i == m.wizard.selectedIndex

			// Special styling for "Create new" option
			if networkId == "__create_new__" {
				createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
				if isSelected {
					content.WriteString(selectedStyle.Render("  ▶ "+name) + "\n")
				} else {
					content.WriteString(createStyle.Render("    "+name) + "\n")
				}
			} else if isSelected {
				content.WriteString(selectedStyle.Render("  ▶ "+name+subnetInfo) + "\n")
			} else {
				content.WriteString(itemStyle.Render("    "+name+subnetInfo) + "\n")
			}
		}

		// Show scroll indicator if needed
		if len(filtered) > maxVisible {
			scrollInfo := fmt.Sprintf("\n  Showing %d-%d of %d networks", startIdx+1, endIdx, len(filtered))
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(scrollInfo))
		}
	}

	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  Space/Enter: Toggle/Select • ↑↓: Navigate • /: Filter • Enter on network: Continue"))

	return content.String()
}

// renderWizardFloatingIPStep renders the floating IP selection step
func (m Model) renderWizardFloatingIPStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	itemStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Background(lipgloss.Color("#2a2a2a"))

	content.WriteString(titleStyle.Render("Floating IP (for external access):") + "\n\n")

	// Info about private network
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Italic(true)
	content.WriteString(infoStyle.Render("  ℹ️  Since you're using only a private network, you need a Floating IP for external access.") + "\n")
	content.WriteString(infoStyle.Render("     A gateway will be created automatically if needed.") + "\n\n")

	// Show filter input if active
	if m.wizard.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s▌", m.wizard.filterInput)) + "\n\n")
	} else if m.wizard.filterInput != "" {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("  Filter: %s (press / to edit)", m.wizard.filterInput)) + "\n\n")
	}

	filtered := m.getFilteredWizardFloatingIPs()
	if len(filtered) == 0 {
		return content.String() + itemStyle.Render("  No floating IPs available")
	}

	// Determine visible range
	maxVisible := 10
	startIdx := 0
	if m.wizard.selectedIndex >= maxVisible {
		startIdx = m.wizard.selectedIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(filtered) {
		endIdx = len(filtered)
	}

	for i := startIdx; i < endIdx; i++ {
		fip := filtered[i]
		id := getString(fip, "id")
		name := getString(fip, "name")

		isSelected := i == m.wizard.selectedIndex

		// Special styling for special options
		if id == "__none__" {
			if isSelected {
				content.WriteString(selectedStyle.Render("▶ "+name) + "\n")
			} else {
				content.WriteString(itemStyle.Render("  "+name) + "\n")
			}
		} else if id == "__create_new__" {
			createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
			if isSelected {
				content.WriteString(selectedStyle.Render("▶ "+name) + "\n")
			} else {
				content.WriteString(createStyle.Render("  "+name) + "\n")
			}
		} else {
			ip := getString(fip, "ip")
			displayStr := fmt.Sprintf("%s (%s)", ip, name)
			if name == "" {
				displayStr = ip
			}
			if isSelected {
				content.WriteString(selectedStyle.Render("▶ "+displayStr) + "\n")
			} else {
				content.WriteString(itemStyle.Render("  "+displayStr) + "\n")
			}
		}
	}

	// Show scroll indicator if needed
	if len(filtered) > maxVisible {
		scrollInfo := fmt.Sprintf("\n  Showing %d-%d of %d options", startIdx+1, endIdx, len(filtered))
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render(scrollInfo))
	}

	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  ↑↓: Navigate • /: Filter • Enter: Select • ←: Back"))

	return content.String()
}

// renderWizardNameStep renders the instance name input step
func (m Model) renderWizardNameStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Enter instance name:") + "\n\n")

	// Summary of selections
	summaryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE"))

	content.WriteString(summaryStyle.Render("  Region:  ") + valueStyle.Render(m.wizard.selectedRegion) + "\n")
	content.WriteString(summaryStyle.Render("  Flavor:  ") + valueStyle.Render(m.wizard.selectedFlavorName) + "\n")
	content.WriteString(summaryStyle.Render("  Image:   ") + valueStyle.Render(m.wizard.selectedImageName) + "\n")
	sshKeyDisplay := m.wizard.selectedSSHKeyName
	if sshKeyDisplay == "" || sshKeyDisplay == "(No SSH Key)" {
		sshKeyDisplay = "(none)"
	}
	content.WriteString(summaryStyle.Render("  SSH Key: ") + valueStyle.Render(sshKeyDisplay) + "\n\n")

	// Name input
	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7B68EE")).
		Padding(0, 1).
		Width(40)

	inputContent := m.wizard.nameInput
	if inputContent == "" {
		inputContent = lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("my-instance")
	}

	content.WriteString("  " + inputStyle.Render(inputContent+"▌") + "\n\n")
	content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  Type a name and press Enter to continue"))

	return content.String()
}

// renderWizardConfirmStep renders the confirmation step
func (m Model) renderWizardConfirmStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Confirm instance creation:") + "\n\n")

	// Summary box
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(labelStyle.Render("  Name:") + valueStyle.Render(m.wizard.instanceName) + "\n")
	content.WriteString(labelStyle.Render("  Region:") + valueStyle.Render(m.wizard.selectedRegion) + "\n")
	content.WriteString(labelStyle.Render("  Flavor:") + valueStyle.Render(m.wizard.selectedFlavorName) + "\n")
	content.WriteString(labelStyle.Render("  Image:") + valueStyle.Render(m.wizard.selectedImageName) + "\n")
	sshKeyDisplay := m.wizard.selectedSSHKeyName
	if sshKeyDisplay == "" || sshKeyDisplay == "(No SSH Key)" {
		sshKeyDisplay = "(none)"
	}
	content.WriteString(labelStyle.Render("  SSH Key:") + valueStyle.Render(sshKeyDisplay) + "\n")

	// Network info
	networkDisplay := ""
	if m.wizard.usePublicNetwork {
		networkDisplay = "Public"
	} else {
		networkDisplay = "Private only"
	}
	if m.wizard.selectedPrivateNetworkName != "" {
		if m.wizard.usePublicNetwork {
			networkDisplay += " + " + m.wizard.selectedPrivateNetworkName
		} else {
			networkDisplay = m.wizard.selectedPrivateNetworkName
		}
	}
	content.WriteString(labelStyle.Render("  Network:") + valueStyle.Render(networkDisplay) + "\n")

	// Show floating IP info if relevant
	if !m.wizard.usePublicNetwork && m.wizard.selectedPrivateNetwork != "" {
		floatingIPDisplay := "(none)"
		if m.wizard.selectedFloatingIP == "__create_new__" {
			floatingIPDisplay = "(will be created)"
		} else if m.wizard.selectedFloatingIPAddress != "" {
			floatingIPDisplay = m.wizard.selectedFloatingIPAddress
		}
		content.WriteString(labelStyle.Render("  Floating IP:") + valueStyle.Render(floatingIPDisplay) + "\n")
	}

	content.WriteString("\n")

	// Confirmation prompt
	confirmStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF7F")).
		Bold(true)

	cancelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	if m.wizard.selectedIndex == 0 {
		content.WriteString(confirmStyle.Render("  ▶ [Create Instance]") + "    ")
		content.WriteString(cancelStyle.Render("  [Cancel]") + "\n")
	} else {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("    [Create Instance]") + "    ")
		content.WriteString(cancelStyle.Render("  ▶ [Cancel]") + "\n")
	}

	return content.String()
}

// renderKubeWizardRegionStep renders the Kubernetes region selection step
func (m Model) renderKubeWizardRegionStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select region for Kubernetes cluster:") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("Loading regions..."))
		return content.String()
	}

	// Build list of regions
	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)

	for i, region := range m.wizard.kubeRegions {
		regionCode, _ := region["code"].(string)
		regionName, _ := region["name"].(string)

		if regionName == "" {
			regionName = regionCode
		}

		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + regionName + " (" + regionCode + ")"))
		} else {
			content.WriteString(listStyle.Render("  " + regionName + " (" + regionCode + ")"))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter Select • q Cancel"))

	return content.String()
}

// renderKubeWizardVersionStep renders the Kubernetes version selection step
func (m Model) renderKubeWizardVersionStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select Kubernetes version:") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("Loading versions..."))
		return content.String()
	}

	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)

	for i, version := range m.wizard.kubeVersions {
		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + version))
		} else {
			content.WriteString(listStyle.Render("  " + version))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter Select • q Cancel"))

	return content.String()
}

// renderKubeWizardNetworkStep renders the Kubernetes network selection step
func (m Model) renderKubeWizardNetworkStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select private network for Kubernetes:") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("Loading networks..."))
		return content.String()
	}

	// Option to not use a private network
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	if m.wizard.selectedIndex == 0 {
		content.WriteString(selectedStyle.Render("▶ (No private network)"))
	} else {
		content.WriteString(listStyle.Render("  (No private network)"))
	}
	content.WriteString("\n")

	// List networks
	for i, network := range m.wizard.kubeNetworks {
		networkName, _ := network["name"].(string)
		idx := i + 1

		if idx == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + networkName))
		} else {
			content.WriteString(listStyle.Render("  " + networkName))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter Select • q Cancel"))

	return content.String()
}

// renderKubeWizardSubnetStep renders the Kubernetes subnet selection step
func (m Model) renderKubeWizardSubnetStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))

	if m.wizard.kubeSubnetMenuIndex == 0 {
		content.WriteString(titleStyle.Render("Select nodes subnet:") + "\n\n")
	} else {
		content.WriteString(titleStyle.Render("Select load balancer subnet:") + "\n\n")
	}

	if m.wizard.isLoading {
		content.WriteString(loadingStyle.Render("Loading subnets..."))
		return content.String()
	}

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	var subnets []map[string]interface{}
	var selectedSubnet string

	if m.wizard.kubeSubnetMenuIndex == 0 {
		subnets = m.wizard.kubeSubnets
		selectedSubnet = m.wizard.selectedNodesSubnet
	} else {
		subnets = m.wizard.kubeSubnets
		selectedSubnet = m.wizard.selectedLBSubnet
	}

	// Option to use same as nodes subnet (only for LB subnet selection)
	if m.wizard.kubeSubnetMenuIndex == 1 {
		if m.wizard.selectedIndex == 0 {
			content.WriteString(selectedStyle.Render("▶ (Same as nodes subnet)"))
		} else {
			content.WriteString(listStyle.Render("  (Same as nodes subnet)"))
		}
		content.WriteString("\n")
	}

	// List subnets
	for i, subnet := range subnets {
		subnetCIDR, _ := subnet["cidr"].(string)
		idx := i
		if m.wizard.kubeSubnetMenuIndex == 1 {
			idx = i + 1 // Offset by 1 for "same as nodes" option
		}

		var isSelected bool
		if m.wizard.kubeSubnetMenuIndex == 0 {
			isSelected = (subnetCIDR == selectedSubnet)
		} else {
			isSelected = (subnetCIDR == selectedSubnet)
		}

		if idx == m.wizard.selectedIndex || isSelected {
			content.WriteString(selectedStyle.Render("▶ " + subnetCIDR))
		} else {
			content.WriteString(listStyle.Render("  " + subnetCIDR))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter Select • q Cancel"))

	return content.String()
}

// renderKubeWizardNameStep renders the Kubernetes cluster name input step
func (m Model) renderKubeWizardNameStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Enter cluster name:") + "\n\n")

	// Input box
	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).
		Width(40)

	content.WriteString(inputStyle.Render(m.wizard.kubeNameInput) + "\n\n")

	// Validation message
	if m.wizard.kubeName == "" && len(m.wizard.kubeNameInput) > 0 {
		validationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		content.WriteString(validationStyle.Render("Name must be 3-32 alphanumeric characters"))
		content.WriteString("\n\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	content.WriteString(helpStyle.Render("Type to enter • Enter Continue • Backspace Clear"))

	return content.String()
}

// renderKubeWizardOptionsStep renders the Kubernetes advanced options step
func (m Model) renderKubeWizardOptionsStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Advanced options:") + "\n\n")

	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	// Plan selection
	if m.wizard.kubeOptionsFieldIndex == 0 {
		content.WriteString(selectedStyle.Render("▶ Plan: " + m.wizard.kubePlan))
	} else {
		content.WriteString(normalStyle.Render("  Plan: " + m.wizard.kubePlan))
	}
	content.WriteString("\n")

	// Update policy
	if m.wizard.kubeOptionsFieldIndex == 1 {
		content.WriteString(selectedStyle.Render("▶ Update Policy: " + m.wizard.kubeUpdatePolicy))
	} else {
		content.WriteString(normalStyle.Render("  Update Policy: " + m.wizard.kubeUpdatePolicy))
	}
	content.WriteString("\n")

	// Kube-proxy mode
	if m.wizard.kubeOptionsFieldIndex == 2 {
		content.WriteString(selectedStyle.Render("▶ Kube-proxy Mode: " + m.wizard.kubeProxyMode))
	} else {
		content.WriteString(normalStyle.Render("  Kube-proxy Mode: " + m.wizard.kubeProxyMode))
	}
	content.WriteString("\n")

	// Private routing flag
	routingStr := "Disabled"
	if m.wizard.kubePrivateRouting {
		routingStr = "Enabled"
	}
	if m.wizard.kubeOptionsFieldIndex == 3 {
		content.WriteString(selectedStyle.Render("▶ Private Routing: " + routingStr))
	} else {
		content.WriteString(normalStyle.Render("  Private Routing: " + routingStr))
	}
	content.WriteString("\n")

	// Gateway IP (if private routing enabled)
	if m.wizard.kubePrivateRouting {
		if m.wizard.kubeOptionsFieldIndex == 4 {
			content.WriteString(selectedStyle.Render("▶ Gateway IP: " + m.wizard.kubeGatewayIPInput))
		} else {
			content.WriteString(normalStyle.Render("  Gateway IP: " + m.wizard.kubeGatewayIPInput))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter Select • q Cancel"))

	return content.String()
}

// renderKubeWizardConfirmStep renders the Kubernetes cluster confirmation step
func (m Model) renderKubeWizardConfirmStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Confirm Kubernetes cluster creation:") + "\n\n")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(labelStyle.Render("  Name:") + valueStyle.Render(m.wizard.kubeName) + "\n")
	content.WriteString(labelStyle.Render("  Region:") + valueStyle.Render(m.wizard.selectedKubeRegion) + "\n")
	content.WriteString(labelStyle.Render("  Version:") + valueStyle.Render(m.wizard.selectedKubeVersion) + "\n")

	networkDisplay := "(public only)"
	if m.wizard.selectedKubeNetworkName != "" {
		networkDisplay = m.wizard.selectedKubeNetworkName
	}
	content.WriteString(labelStyle.Render("  Network:") + valueStyle.Render(networkDisplay) + "\n")

	content.WriteString(labelStyle.Render("  Plan:") + valueStyle.Render(m.wizard.kubePlan) + "\n")
	content.WriteString(labelStyle.Render("  Update Policy:") + valueStyle.Render(m.wizard.kubeUpdatePolicy) + "\n")
	content.WriteString(labelStyle.Render("  Kube-proxy Mode:") + valueStyle.Render(m.wizard.kubeProxyMode) + "\n")

	routingStr := "Disabled"
	if m.wizard.kubePrivateRouting {
		routingStr = "Enabled (" + m.wizard.kubeGatewayIP + ")"
	}
	content.WriteString(labelStyle.Render("  Private Routing:") + valueStyle.Render(routingStr) + "\n")

	content.WriteString("\n")

	confirmStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF7F"))
	cancelStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF6B6B"))

	if m.wizard.kubeConfirmButtonIndex == 0 {
		content.WriteString(confirmStyle.Render("  ▶ [Create Cluster]") + "    ")
		content.WriteString(cancelStyle.Render("  [Cancel]") + "\n")
	} else {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("    [Create Cluster]") + "    ")
		content.WriteString(cancelStyle.Render("  ▶ [Cancel]") + "\n")
	}

	return content.String()
}

// ========== Node Pool Wizard Render Functions ==========

func (m Model) renderNodePoolWizardFlavorStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Add Node Pool - Select Flavor:") + "\n\n")

	if m.wizard.isLoading {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Render("Loading flavors..."))
		return content.String()
	}

	if m.wizard.errorMsg != "" {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Render(m.wizard.errorMsg))
		return content.String()
	}

	// Show filter input if in filter mode
	if m.wizard.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
		content.WriteString(filterStyle.Render("Filter: ") + m.wizard.filterInput + "_\n\n")
	}

	// Apply filter to flavors
	flavors := m.wizard.nodePoolFlavors
	if m.wizard.filterInput != "" {
		var filtered []map[string]interface{}
		for _, flavor := range flavors {
			name := getString(flavor, "name")
			if strings.Contains(strings.ToLower(name), strings.ToLower(m.wizard.filterInput)) {
				filtered = append(filtered, flavor)
			}
		}
		flavors = filtered
	}

	if len(flavors) == 0 {
		if m.wizard.filterInput != "" {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No flavors match filter"))
		} else {
			content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No flavors available"))
		}
		return content.String()
	}

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	maxVisible := 10
	startIdx := 0
	if m.wizard.selectedIndex >= maxVisible {
		startIdx = m.wizard.selectedIndex - maxVisible + 1
	}
	endIdx := startIdx + maxVisible
	if endIdx > len(flavors) {
		endIdx = len(flavors)
	}

	for i := startIdx; i < endIdx; i++ {
		flavor := flavors[i]
		name := getString(flavor, "name")
		category := getString(flavor, "category")

		// Get specs
		vcpus := getFloatValue(flavor, "vCPUs", 0)
		ram := getFloatValue(flavor, "ram", 0) / 1024 // Convert to GB

		label := fmt.Sprintf("%s (%s) - %d vCPU, %.0fGB RAM", name, category, int(vcpus), ram)

		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render(fmt.Sprintf("  ▶ %s", label)) + "\n")
		} else {
			content.WriteString(normalStyle.Render(fmt.Sprintf("    %s", label)) + "\n")
		}
	}

	if len(flavors) > maxVisible {
		content.WriteString(dimStyle.Render(fmt.Sprintf("\n  (%d/%d - scroll for more)", m.wizard.selectedIndex+1, len(flavors))))
	}

	content.WriteString("\n" + dimStyle.Render("  ↑/↓ Navigate • / Filter • Enter Select • Escape Cancel"))

	return content.String()
}

func (m Model) renderNodePoolWizardNameStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Add Node Pool - Enter Name:") + "\n\n")

	// Show error if present
	if m.wizard.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errorStyle.Render("  ❌ "+m.wizard.errorMsg) + "\n\n")
	}

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))

	content.WriteString(labelStyle.Render("  Flavor: ") + m.wizard.nodePoolFlavorName + "\n\n")
	content.WriteString(labelStyle.Render("  Node pool name: ") + inputStyle.Render(m.wizard.nodePoolNameInput+"▌") + "\n")

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString("\n" + dimStyle.Render("  Enter to continue • Escape to go back"))

	return content.String()
}

func (m Model) renderNodePoolWizardSizeStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Add Node Pool - Configure Size:") + "\n\n")

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	fields := []struct {
		label string
		value int
	}{
		{"Desired nodes", m.wizard.nodePoolDesiredNodes},
		{"Min nodes", m.wizard.nodePoolMinNodes},
		{"Max nodes", m.wizard.nodePoolMaxNodes},
	}

	for i, field := range fields {
		prefix := "    "
		style := normalStyle
		if i == m.wizard.nodePoolSizeFieldIndex {
			prefix = "  ▶ "
			style = selectedStyle
		}
		content.WriteString(style.Render(fmt.Sprintf("%s%s: %d", prefix, field.label, field.value)) + "\n")
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString("\n" + dimStyle.Render("  ↑/↓ Select field • ←/→ Change value • Enter Continue • Escape Back"))

	return content.String()
}

func (m Model) renderNodePoolWizardOptionsStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Add Node Pool - Options:") + "\n\n")

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	options := []struct {
		label   string
		enabled bool
	}{
		{"Autoscale", m.wizard.nodePoolAutoscale},
		{"Anti-affinity", m.wizard.nodePoolAntiAffinity},
		{"Monthly billing", m.wizard.nodePoolMonthlyBilled},
	}

	for i, opt := range options {
		prefix := "    "
		style := normalStyle
		if i == m.wizard.nodePoolOptionsFieldIdx {
			prefix = "  ▶ "
			style = selectedStyle
		}
		checkmark := "[ ]"
		if opt.enabled {
			checkmark = "[✓]"
		}
		content.WriteString(style.Render(fmt.Sprintf("%s%s %s", prefix, checkmark, opt.label)) + "\n")
	}

	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	content.WriteString("\n" + dimStyle.Render("  ↑/↓ Select • Space Toggle • Enter Continue • Escape Back"))

	return content.String()
}

func (m Model) renderNodePoolWizardConfirmStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Confirm Node Pool Creation:") + "\n\n")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(labelStyle.Render("  Name:") + valueStyle.Render(m.wizard.nodePoolName) + "\n")
	content.WriteString(labelStyle.Render("  Flavor:") + valueStyle.Render(m.wizard.nodePoolFlavorName) + "\n")
	content.WriteString(labelStyle.Render("  Desired nodes:") + valueStyle.Render(fmt.Sprintf("%d", m.wizard.nodePoolDesiredNodes)) + "\n")
	content.WriteString(labelStyle.Render("  Min/Max:") + valueStyle.Render(fmt.Sprintf("%d / %d", m.wizard.nodePoolMinNodes, m.wizard.nodePoolMaxNodes)) + "\n")

	autoscaleStr := "No"
	if m.wizard.nodePoolAutoscale {
		autoscaleStr = "Yes"
	}
	content.WriteString(labelStyle.Render("  Autoscale:") + valueStyle.Render(autoscaleStr) + "\n")

	antiAffinityStr := "No"
	if m.wizard.nodePoolAntiAffinity {
		antiAffinityStr = "Yes"
	}
	content.WriteString(labelStyle.Render("  Anti-affinity:") + valueStyle.Render(antiAffinityStr) + "\n")

	monthlyStr := "Hourly"
	if m.wizard.nodePoolMonthlyBilled {
		monthlyStr = "Monthly"
	}
	content.WriteString(labelStyle.Render("  Billing:") + valueStyle.Render(monthlyStr) + "\n")

	content.WriteString("\n")

	createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	cancelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	if m.wizard.nodePoolConfirmBtnIdx == 0 {
		content.WriteString(createStyle.Render("  ▶ [Create Node Pool]") + "    ")
		content.WriteString(dimStyle.Render("[Cancel]") + "\n")
	} else {
		content.WriteString(dimStyle.Render("    [Create Node Pool]") + "    ")
		content.WriteString(cancelStyle.Render("▶ [Cancel]") + "\n")
	}

	return content.String()
}

// renderCleanupConfirmation renders the cleanup confirmation dialog
func (m Model) renderCleanupConfirmation(width int) string {
	var content strings.Builder

	// Error header
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	content.WriteString("\n")
	content.WriteString(titleStyle.Render("⚠️  ERROR DURING CREATION"))
	content.WriteString("\n\n")

	// Error message
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	content.WriteString(errorStyle.Render(m.wizard.cleanupError))
	content.WriteString("\n\n")

	// List created resources
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Resources created:"))
	content.WriteString("\n")

	resourceStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("39"))
	if m.wizard.createdInstanceId != "" {
		content.WriteString("  • " + resourceStyle.Render("Instance: "+m.wizard.createdInstanceName) + "\n")
	}
	if m.wizard.createdNetworkId != "" {
		content.WriteString("  • " + resourceStyle.Render("Network: "+m.wizard.selectedPrivateNetworkName) + "\n")
	}
	if m.wizard.createdGatewayId != "" {
		content.WriteString("  • " + resourceStyle.Render("Gateway") + "\n")
	}
	if m.wizard.createdFloatingIPId != "" {
		content.WriteString("  • " + resourceStyle.Render("Floating IP") + "\n")
	}

	content.WriteString("\n")
	content.WriteString(lipgloss.NewStyle().Bold(true).Render("Do you want to delete these resources?"))
	content.WriteString("\n\n")

	// Options
	yesStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Bold(true)
	noStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	if m.wizard.selectedIndex == 0 {
		content.WriteString(yesStyle.Render("  ▶ [Yes, delete all]") + "    ")
		content.WriteString(noStyle.Render("  [No, keep them]") + "\n")
	} else {
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("    [Yes, delete all]") + "    ")
		content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Render("  ▶ [No, keep them]") + "\n")
	}

	return content.String()
}

// getProductCreationInfo returns the product name and CLI command to create it
func (m Model) getProductCreationInfo() (string, string) {
	switch m.currentProduct {
	case ProductInstances:
		return "instances", fmt.Sprintf("ovhcloud cloud instance create --cloud-project %s", m.cloudProject)
	case ProductKubernetes:
		return "Kubernetes clusters", fmt.Sprintf("ovhcloud cloud kube create --cloud-project %s", m.cloudProject)
	case ProductDatabases:
		return "databases", fmt.Sprintf("ovhcloud cloud database-service create --cloud-project %s", m.cloudProject)
	case ProductStorage:
		return "storage containers", fmt.Sprintf("ovhcloud cloud storage s3 create --cloud-project %s", m.cloudProject)
	case ProductNetworks:
		return "private networks", fmt.Sprintf("ovhcloud cloud network create --cloud-project %s", m.cloudProject)
	default:
		return "resources", ""
	}
}

func (m Model) renderTable() string {
	if m.table.Rows() == nil || len(m.table.Rows()) == 0 {
		if m.filterInput != "" {
			return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No results match filter: " + m.filterInput)
		}
		return lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Render("No data available")
	}

	var content strings.Builder

	// Show filter indicator if filter is active (but not in edit mode)
	if m.filterInput != "" && !m.filterMode {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
		content.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s (press / to edit, Esc to clear)", m.filterInput)) + "\n\n")
	}

	content.WriteString(m.table.View())
	return content.String()
}

func (m Model) renderDeleteConfirmView() string {
	var content strings.Builder
	var instanceName string

	if m.deleteTarget != nil {
		if name, exists := m.deleteTarget["name"].(string); exists {
			instanceName = name
		}
	}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("196"))
	content.WriteString("\n")
	content.WriteString(titleStyle.Render("⚠️  DELETE INSTANCE"))
	content.WriteString("\n\n")

	warningStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	content.WriteString(warningStyle.Render("You are about to delete the following instance:"))
	content.WriteString("\n\n")

	nameStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	content.WriteString("  Instance: ")
	content.WriteString(nameStyle.Render(instanceName))
	content.WriteString("\n\n")

	content.WriteString(warningStyle.Render("This action is IRREVERSIBLE. All data will be lost."))
	content.WriteString("\n\n")

	content.WriteString("To confirm, type the instance name: ")
	content.WriteString(lipgloss.NewStyle().Background(lipgloss.Color("236")).Render(" " + m.deleteConfirmInput + "█ "))
	content.WriteString("\n\n")

	if m.deleteConfirmInput == instanceName && instanceName != "" {
		confirmStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
		content.WriteString(confirmStyle.Render("✓ Press Enter to confirm deletion"))
	} else {
		hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
		content.WriteString(hintStyle.Render("Press Esc to cancel"))
	}

	return content.String()
}

func (m Model) renderDetailView(width int) string {
	if m.detailData == nil {
		return "No data"
	}

	// Determine what type of resource we're viewing
	switch m.currentProduct {
	case ProductInstances:
		return m.renderInstanceDetail(width)
	case ProductKubernetes:
		return m.renderKubernetesDetail(width)
	case ProductProjects:
		return m.renderProjectDetail(width)
	default:
		return m.renderGenericDetail(width)
	}
}

func (m Model) renderInstanceDetail(width int) string {
	var content strings.Builder

	// Get values safely
	_ = getStringValue(m.detailData, "name", "Unknown") // name is shown in title bar
	status := getStringValue(m.detailData, "status", "Unknown")
	id := getStringValue(m.detailData, "id", "N/A")
	region := getStringValue(m.detailData, "region", "N/A")
	created := getStringValue(m.detailData, "created", "N/A")

	// Get flavor name from nested object or fallback to flavorId
	flavorName := "N/A"
	if flavor, ok := m.detailData["flavor"].(map[string]interface{}); ok {
		flavorName = getStringValue(flavor, "name", "N/A")
	}
	if flavorName == "N/A" {
		flavorName = getStringValue(m.detailData, "flavorId", "N/A")
	}

	// Get image name from imageMap or fallback to imageId
	imageId := getStringValue(m.detailData, "imageId", "")
	imageName := "N/A"
	if imageId != "" && m.imageMap != nil {
		if name, ok := m.imageMap[imageId]; ok {
			imageName = name
		}
	}
	if imageName == "N/A" && imageId != "" {
		imageName = imageId
	}

	// Get IP addresses - check all addresses for public and private IPs
	ipv4Public := ""
	ipv4Private := ""
	ipv6Public := ""
	floatingIP := ""
	if addresses, ok := m.detailData["ipAddresses"].([]interface{}); ok {
		for _, addr := range addresses {
			if addrMap, ok := addr.(map[string]interface{}); ok {
				ip := getStringValue(addrMap, "ip", "")
				version := int(getFloatValue(addrMap, "version", 0))
				ipType := getStringValue(addrMap, "type", "")
				if version == 4 {
					if ipType == "public" && ipv4Public == "" {
						ipv4Public = ip
					} else if ipType == "private" && ipv4Private == "" {
						ipv4Private = ip
					}
				} else if version == 6 && ipType == "public" && ipv6Public == "" {
					ipv6Public = ip
				}
			}
		}
	}

	// Check for floating IP
	if m.floatingIPMap != nil {
		if fip, ok := m.floatingIPMap[id]; ok {
			floatingIP = fip
		}
	}

	// Status indicator
	statusIcon := "🟢"
	statusStyle := statusRunningStyle
	if strings.ToLower(status) != "active" && strings.ToLower(status) != "running" {
		statusIcon = "🔴"
		statusStyle = statusStoppedStyle
	}

	// Build the detail view with boxes
	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Left column - Information box
	infoContent := strings.Builder{}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("ID"), valueStyle.Render(truncate(id, 30))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Region"), valueStyle.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Flavor"), valueStyle.Render(truncate(flavorName, 25))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Image"), valueStyle.Render(truncate(imageName, 25))))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Created"), valueStyle.Render(truncate(created, 25))))

	infoBox := renderBox("Informations", infoContent.String(), boxWidth)

	// Right column - Network box
	networkContent := strings.Builder{}
	// Show floating IP if available
	if floatingIP != "" {
		networkContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Floating IP"), valueStyle.Render(floatingIP)))
	}
	// Show public IPv4
	if ipv4Public != "" {
		networkContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("IPv4 Public"), valueStyle.Render(ipv4Public)))
	} else if floatingIP == "" {
		networkContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("IPv4 Public"), valueStyle.Render("N/A")))
	}
	// Show private IPv4 if available
	if ipv4Private != "" {
		networkContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("IPv4 Private"), valueStyle.Render(ipv4Private)))
	}
	// Show IPv6
	if ipv6Public != "" {
		networkContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("IPv6"), valueStyle.Render(truncate(ipv6Public, 35))))
	} else {
		networkContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("IPv6"), valueStyle.Render("N/A")))
	}

	networkBox := renderBox("Réseau", networkContent.String(), boxWidth)

	// Actions box (top) with selectable actions
	// Change Stop to Start if instance is SHUTOFF
	stopStartAction := "Stop"
	if strings.ToUpper(status) == "SHUTOFF" {
		stopStartAction = "Start"
	}
	// Change Rescue Mode to Exit Rescue if instance is in RESCUE
	rescueAction := "Rescue Mode"
	if strings.ToUpper(status) == "RESCUE" {
		rescueAction = "Exit Rescue"
	}
	actions := []string{"SSH", "Reboot", rescueAction, stopStartAction, "Console", "Reinstall"}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			// Selected action - highlighted
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color("#7B68EE")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 1).
				Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Padding(0, 1).
				Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.actionConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actions[m.selectedAction]))
	}
	actionsBox := renderBox("Actions rapides (←/→ pour naviguer, Enter pour exécuter)", actionsContent, width-4)

	// Combine everything
	content.WriteString(actionsBox)
	content.WriteString("\n\n")

	// Side by side boxes
	leftRight := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", networkBox)
	content.WriteString(leftRight)

	return content.String()
}

func (m Model) renderKubernetesDetail(width int) string {
	var content strings.Builder

	clusterName := getStringValue(m.detailData, "name", "Unknown")
	status := getStringValue(m.detailData, "status", "Unknown")
	id := getStringValue(m.detailData, "id", "N/A")
	region := getStringValue(m.detailData, "region", "N/A")
	version := getStringValue(m.detailData, "version", "N/A")
	nodesCount := getFloatValue(m.detailData, "nodesCount", 0)
	updatePolicy := getStringValue(m.detailData, "updatePolicy", "N/A")

	statusIcon := "🟢"
	statusStyle := statusRunningStyle
	if strings.ToLower(status) != "ready" && strings.ToLower(status) != "running" {
		statusIcon = "🟡"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}

	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Cluster info
	infoContent := strings.Builder{}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("ID"), valueStyle.Render(truncate(id, 30))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Region"), valueStyle.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Version"), valueStyle.Render(version)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Nodes"), valueStyle.Render(fmt.Sprintf("%.0f", nodesCount))))

	infoBox := renderBox("Cluster "+clusterName, infoContent.String(), boxWidth)

	// Configuration
	configContent := strings.Builder{}
	configContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Update Policy"), valueStyle.Render(updatePolicy)))

	configBox := renderBox("Configuration", configContent.String(), boxWidth)

	// Actions with selection highlighting
	actions := []string{"Kubeconfig", "K9s", "Manage Pools", "Upgrade", "Policy", "Delete"}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			// Selected action - highlighted
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color("#7B68EE")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Padding(0, 1).
				Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Padding(0, 1).
				Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.actionConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).
			Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actions[m.selectedAction]))
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)

	content.WriteString(actionsBox)
	content.WriteString("\n\n")
	leftRight := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", configBox)
	content.WriteString(leftRight)

	// Node Pools section
	content.WriteString("\n\n")
	nodePoolsContent := strings.Builder{}

	nodePools := m.kubeNodePools[id]
	if len(nodePools) == 0 {
		nodePoolsContent.WriteString(lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true).
			Render("Loading node pools..."))
	} else {
		for i, pool := range nodePools {
			poolName := getStringValue(pool, "name", "Unknown")
			poolStatus := getStringValue(pool, "status", "Unknown")
			flavor := getStringValue(pool, "flavor", "N/A")
			desiredNodes := getIntOrFloatValue(pool, "desiredNodes", 0)
			currentNodes := getIntOrFloatValue(pool, "currentNodes", 0)
			autoscale := getBoolValue(pool, "autoscale", false)

			// Status icon for pool
			poolStatusIcon := "🟢"
			poolStatusStyle := statusRunningStyle
			if strings.ToLower(poolStatus) != "ready" && strings.ToLower(poolStatus) != "running" {
				poolStatusIcon = "🟡"
				poolStatusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
			}

			// Pool header
			poolHeader := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7B68EE")).Render(poolName)
			nodePoolsContent.WriteString(fmt.Sprintf("%s  %s\n", poolHeader, poolStatusStyle.Render(poolStatusIcon+" "+poolStatus)))

			// Pool details
			nodePoolsContent.WriteString(fmt.Sprintf("   %s %s   %s %.0f/%.0f",
				labelStyle.Render("Flavor:"), valueStyle.Render(flavor),
				labelStyle.Render("Nodes:"), currentNodes, desiredNodes))

			if autoscale {
				minNodes := getIntOrFloatValue(pool, "minNodes", 0)
				maxNodes := getIntOrFloatValue(pool, "maxNodes", 0)
				nodePoolsContent.WriteString(fmt.Sprintf("   %s %.0f-%.0f",
					labelStyle.Render("Autoscale:"), minNodes, maxNodes))
			}

			if i < len(nodePools)-1 {
				nodePoolsContent.WriteString("\n\n")
			}
		}
	}

	nodePoolsBox := renderBox(fmt.Sprintf("Node Pools (%d)", len(nodePools)), nodePoolsContent.String(), width-4)
	content.WriteString(nodePoolsBox)

	return content.String()
}

func (m Model) renderProjectDetail(width int) string {
	var content strings.Builder

	name := getStringValue(m.detailData, "description", "Unknown Project")
	projectID := getStringValue(m.detailData, "project_id", "N/A")
	status := getStringValue(m.detailData, "status", "N/A")
	createdAt := getStringValue(m.detailData, "creationDate", "N/A")

	statusIcon := "🟢"
	statusStyle := statusRunningStyle
	if strings.ToLower(status) != "ok" {
		statusIcon = "🟡"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}

	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Project info
	infoContent := strings.Builder{}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Project ID"), valueStyle.Render(truncate(projectID, 30))))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Created"), valueStyle.Render(truncate(createdAt, 25))))

	infoBox := renderBox("Project: "+name, infoContent.String(), boxWidth)

	// Actions
	actionsContent := "[Select as Default] [View Resources] [Settings]"
	actionsBox := renderBox("Actions", actionsContent, boxWidth)

	leftRight := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", actionsBox)
	content.WriteString(leftRight)

	return content.String()
}

func (m Model) renderGenericDetail(width int) string {
	var content strings.Builder

	boxWidth := width - 4

	// Sort keys for consistent display
	keys := make([]string, 0, len(m.detailData))
	for k := range m.detailData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	detailContent := strings.Builder{}
	for _, key := range keys {
		value := m.detailData[key]
		valueStr := fmt.Sprintf("%v", value)
		if len(valueStr) > 50 {
			valueStr = valueStr[:47] + "..."
		}
		detailContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render(key), valueStyle.Render(valueStr)))
	}

	detailBox := renderBox("Details", detailContent.String(), boxWidth)
	content.WriteString(detailBox)

	return content.String()
}

func renderBox(title string, content string, width int) string {
	titleRendered := boxTitleStyle.Render("─ " + title + " ")
	titleLen := lipgloss.Width(titleRendered)

	// Build top border with title
	topBorder := "┌" + titleRendered
	remainingWidth := width - titleLen - 2
	if remainingWidth > 0 {
		topBorder += strings.Repeat("─", remainingWidth) + "┐"
	} else {
		topBorder += "┐"
	}

	// Content lines with side borders
	lines := strings.Split(content, "\n")
	var contentLines []string
	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := width - 4 - lineWidth
		if padding < 0 {
			padding = 0
		}
		contentLines = append(contentLines, "│ "+line+strings.Repeat(" ", padding)+" │")
	}

	// Bottom border
	bottomBorder := "└" + strings.Repeat("─", width-2) + "┘"

	return topBorder + "\n" + strings.Join(contentLines, "\n") + "\n" + bottomBorder
}

func (m Model) renderFooter() string {
	var help string

	// Handle filter mode in table view
	if m.filterMode && m.mode == TableView {
		filterStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true)
		return filterStyle.Render(fmt.Sprintf("Filter: %s▌", m.filterInput)) + "\n" + footerStyle.Render("Type to filter • Enter: Confirm • Esc: Clear & Exit")
	}

	switch m.mode {
	case ProjectSelectView:
		help = "↑↓: Navigate • Enter: Select Project • d: Set Default • q: Quit"
	case TableView:
		if m.filterInput != "" {
			help = "←→: Switch Product • ↑↓: Navigate • /: Edit Filter • Enter: Details • c: Create • Del: Delete • d: Debug • Esc: Clear Filter • q: Quit"
		} else {
			help = "←→: Switch Product • ↑↓: Navigate • /: Filter • Enter: Details • c: Create • Del: Delete • d: Debug • p: Change Project • q: Quit"
		}
	case EmptyView:
		help = "←→: Switch Product • c: Create • d: Debug • p: Change Project • q: Quit"
	case DetailView:
		if m.actionConfirm {
			help = "Enter: Confirm Action • Esc: Cancel"
		} else {
			help = "←→: Select Action • Enter: Execute • d: Debug • Esc: Back to List • q: Quit"
		}
	case WizardView:
		if m.wizard.cleanupPending {
			help = "←→: Select • Enter: Confirm • Esc: Keep resources"
		} else if m.wizard.filterMode {
			help = "Type to filter • Enter: Confirm • Esc: Exit filter"
		} else if m.wizard.step == WizardStepRegion {
			help = "↑↓: Navigate • /: Filter • d: Debug • Enter: Select • Esc: Cancel"
		} else if m.wizard.step == WizardStepFlavor || m.wizard.step == WizardStepImage || m.wizard.step == WizardStepSSHKey {
			help = "↑↓: Navigate • /: Filter • d: Debug • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == WizardStepNetwork && !m.wizard.creatingNetwork {
			help = "↑↓: Navigate • /: Filter • d: Debug • Space: Toggle • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == WizardStepFloatingIP {
			help = "↑↓: Navigate • /: Filter • d: Debug • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == WizardStepName {
			help = "Type: Enter name • Enter: Confirm • ←: Back • Esc: Cancel"
		} else if m.wizard.step == WizardStepConfirm {
			help = "←→: Select • d: Debug • Enter: Confirm • Esc: Cancel"
		} else {
			help = "↑↓: Navigate • d: Debug • Enter: Select • ←: Back • Esc: Cancel"
		}
	case DeleteConfirmView:
		help = "Type instance name to confirm • Enter: Delete • Esc: Cancel"
	case DebugView:
		help = "↑↓: Scroll • c: Clear logs • d/Esc: Close • q: Quit"
	default:
		help = "Enter: Select • q: Quit"
	}

	// Add notification if present
	if m.notification != "" && time.Now().Before(m.notificationExpiry) {
		notificationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF7F")).
			Bold(true)
		if strings.HasPrefix(m.notification, "❌") {
			notificationStyle = notificationStyle.Foreground(lipgloss.Color("#FF6B6B"))
		}
		return notificationStyle.Render(m.notification) + "\n" + footerStyle.Render(help)
	}

	return footerStyle.Render(help)
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle wizard mode separately
	if m.mode == WizardView {
		return m.handleWizardKeyPress(msg)
	}

	// Handle delete confirmation mode
	if m.mode == DeleteConfirmView {
		return m.handleDeleteConfirmKeyPress(msg)
	}

	// Handle debug view mode
	if m.mode == DebugView {
		return m.handleDebugKeyPress(msg)
	}

	// Handle Kubernetes upgrade view
	if m.mode == KubeUpgradeView {
		return m.handleKubeUpgradeKeyPress(msg)
	}

	// Handle Kubernetes policy edit view
	if m.mode == KubePolicyEditView {
		return m.handleKubePolicyEditKeyPress(msg)
	}

	// Handle Kubernetes delete confirmation view
	if m.mode == KubeDeleteConfirmView {
		return m.handleKubeDeleteConfirmKeyPress(msg)
	}

	// Handle Node pool scale view
	if m.mode == NodePoolScaleView {
		return m.handleNodePoolScaleKeyPress(msg)
	}

	// Handle Node pool delete confirmation view
	if m.mode == NodePoolDeleteConfirmView {
		return m.handleNodePoolDeleteConfirmKeyPress(msg)
	}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "left":
		// In NodePoolDetailView, navigate actions
		if m.mode == NodePoolDetailView {
			if m.nodePoolDetailActionIdx > 0 {
				m.nodePoolDetailActionIdx--
				m.nodePoolDetailConfirm = false
			}
			return m, nil
		}
		if m.mode == DetailView && m.currentProduct == ProductInstances {
			if m.selectedAction > 0 {
				m.selectedAction--
				m.actionConfirm = false
			}
			return m, nil
		}
		// In DetailView for Kubernetes, navigate actions
		if m.mode == DetailView && m.currentProduct == ProductKubernetes {
			if m.selectedAction > 0 {
				m.selectedAction--
				m.actionConfirm = false
			}
			return m, nil
		}
		if m.mode != ProjectSelectView && m.currentProduct != ProductProjects {
			if m.navIdx > 0 {
				m.navIdx--
				return m.loadCurrentProduct()
			}
		}
		return m, nil

	case "right":
		// In NodePoolDetailView, navigate actions
		if m.mode == NodePoolDetailView {
			if m.nodePoolDetailActionIdx < 1 { // 2 actions: Scale, Delete
				m.nodePoolDetailActionIdx++
				m.nodePoolDetailConfirm = false
			}
			return m, nil
		}
		// In DetailView, navigate actions
		if m.mode == DetailView && m.currentProduct == ProductInstances {
			if m.selectedAction < 5 { // 6 actions: 0-5
				m.selectedAction++
				m.actionConfirm = false
			}
			return m, nil
		}
		// In DetailView for Kubernetes, navigate actions
		if m.mode == DetailView && m.currentProduct == ProductKubernetes {
			if m.selectedAction < 5 { // 6 actions: 0-5
				m.selectedAction++
				m.actionConfirm = false
			}
			return m, nil
		}
		if m.mode != ProjectSelectView && m.currentProduct != ProductProjects {
			navItems := getNavItems()
			if m.navIdx < len(navItems)-1 {
				m.navIdx++
				return m.loadCurrentProduct()
			}
		}
		return m, nil

	case "p":
		// Go back to project selection
		if m.mode != ProjectSelectView && m.currentProduct != ProductProjects {
			m.currentProduct = ProductProjects
			m.navIdx = 0
			// If we have cached projects, show them directly
			if len(m.projectsList) > 0 {
				m.table = createProjectsTable(m.projectsList, m.width, m.height)
				m.currentData = m.projectsList
				m.mode = ProjectSelectView
				return m, nil
			}
			// Otherwise fetch projects
			m.mode = LoadingView
			return m, m.fetchDataForPath("/projects")
		}
		return m, nil

	case "esc":
		// Clear filter in TableView if active
		if m.mode == TableView && m.filterInput != "" {
			m.filterInput = ""
			m.applyTableFilter()
			return m, nil
		}
		// Go back to node pools view from node pool detail view, or cancel action confirm
		if m.mode == NodePoolDetailView {
			if m.nodePoolDetailConfirm {
				m.nodePoolDetailConfirm = false
			} else {
				m.mode = NodePoolsView
				m.selectedNodePool = nil
				m.nodePoolDetailActionIdx = 0
			}
			return m, nil
		}
		// Go back to detail view from node pools view
		if m.mode == NodePoolsView {
			m.mode = DetailView
			return m, nil
		}
		// Go back to table view from detail view, or cancel action confirm
		if m.mode == DetailView {
			if m.actionConfirm {
				m.actionConfirm = false
			} else {
				m.mode = TableView
				m.selectedAction = 0
			}
			return m, nil
		}
		return m, nil

	case "c":
		// Create resource - available in TableView, EmptyView, and NodePoolsView
		if (m.mode == TableView || m.mode == EmptyView) && m.currentProduct != ProductProjects {
			return m, m.launchCreationWizard()
		}
		// Create node pool from NodePoolsView
		if m.mode == NodePoolsView {
			clusterId := m.wizard.nodePoolClusterId
			region := getStringValue(m.detailData, "region", "")
			return m.handleStartNodePoolWizard(startNodePoolWizardMsg{
				clusterId: clusterId,
				region:    region,
			})
		}
		return m, nil

	case "enter":
		// Handle enter based on current mode
		if m.mode == NodePoolDetailView {
			// Execute selected action on node pool
			if m.nodePoolDetailConfirm {
				// Confirmed - execute the action
				m.nodePoolDetailConfirm = false
				return m, m.executeNodePoolAction(m.nodePoolDetailActionIdx)
			} else {
				// Ask for confirmation (except for Scale which needs a wizard)
				if m.nodePoolDetailActionIdx == 0 {
					// Scale - launch scale view directly
					return m, m.executeNodePoolAction(0)
				}
				m.nodePoolDetailConfirm = true
				return m, nil
			}
		} else if m.mode == DetailView && m.currentProduct == ProductInstances {
			// Execute selected action on instance
			if m.actionConfirm {
				// Confirmed - execute the action
				m.actionConfirm = false
				return m, m.executeInstanceAction(m.selectedAction)
			} else {
				// Ask for confirmation
				m.actionConfirm = true
				return m, nil
			}
		} else if m.mode == DetailView && m.currentProduct == ProductKubernetes {
			// Execute selected action on Kubernetes cluster
			// Manage Pools (index 2) doesn't need confirmation
			if m.selectedAction == 2 {
				// Execute directly without confirmation
				return m, m.executeKubeAction(m.selectedAction)
			} else if m.actionConfirm {
				// Confirmed - execute the action
				m.actionConfirm = false
				return m, m.executeKubeAction(m.selectedAction)
			} else {
				// Ask for confirmation
				m.actionConfirm = true
				return m, nil
			}
		} else if m.mode == ProjectSelectView {
			// Select project and go to products view
			selectedRow := m.table.Cursor()
			if selectedRow >= 0 && selectedRow < len(m.currentData) {
				project := m.currentData[selectedRow]
				m.cloudProject = getStringValue(project, "project_id", "")
				m.cloudProjectName = getStringValue(project, "description", "")
				if m.cloudProjectName == "" {
					m.cloudProjectName = getStringValue(project, "projectName", m.cloudProject)
				}
				// Switch to instances view as default product
				m.currentProduct = ProductInstances
				m.navIdx = 0
				m.mode = LoadingView
				m.detailData = nil
				m.currentData = nil
				return m, m.fetchDataForPath("/instances")
			}
		} else if m.mode == NodePoolsView {
			// In node pools view, show node pool details
			clusterId := getStringValue(m.detailData, "id", "")
			nodePools := m.kubeNodePools[clusterId]
			if m.nodePoolsSelectedIdx >= 0 && m.nodePoolsSelectedIdx < len(nodePools) {
				m.selectedNodePool = nodePools[m.nodePoolsSelectedIdx]
				m.mode = NodePoolDetailView
			}
		} else if m.mode == TableView {
			// In table view, show details
			selectedRow := m.table.Cursor()
			if selectedRow >= 0 && selectedRow < len(m.currentData) {
				m.detailData = m.currentData[selectedRow]
				m.currentItemName = getStringValue(m.detailData, "name", "Item")
				m.mode = DetailView

				// If viewing a Kubernetes cluster, also load node pools
				if m.currentProduct == ProductKubernetes {
					kubeId := getStringValue(m.detailData, "id", "")
					if kubeId != "" {
						return m, m.fetchKubeNodePools(kubeId)
					}
				}
			}
		}
		return m, nil
	case "up", "down", "j", "k":
		// Node pools list navigation
		if m.mode == NodePoolsView {
			clusterId := getStringValue(m.detailData, "id", "")
			nodePools := m.kubeNodePools[clusterId]
			if len(nodePools) > 0 {
				if msg.String() == "down" || msg.String() == "j" {
					if m.nodePoolsSelectedIdx < len(nodePools)-1 {
						m.nodePoolsSelectedIdx++
					}
				} else if msg.String() == "up" || msg.String() == "k" {
					if m.nodePoolsSelectedIdx > 0 {
						m.nodePoolsSelectedIdx--
					}
				}
			}
			return m, nil
		}
		// Table navigation (works in both ProjectSelectView and TableView)
		if m.mode == TableView || m.mode == ProjectSelectView {
			var cmd tea.Cmd
			m.table, cmd = m.table.Update(msg)
			return m, cmd
		}
		return m, nil

	case "r":
		// Refresh current view
		if m.mode == DetailView && m.detailData != nil {
			// Refresh detail view by reloading list and finding the item again
			itemId := getString(m.detailData, "id")
			itemName := m.currentItemName
			m.notification = "⟳ Refreshing..."
			m.notificationExpiry = time.Now().Add(2 * time.Second)
			m.mode = LoadingView

			// Store a flag to return to detail view after loading
			m.detailData = map[string]interface{}{
				"_refreshItemId":   itemId,
				"_refreshItemName": itemName,
			}

			var path string
			switch m.currentProduct {
			case ProductInstances:
				path = "/instances"
			case ProductKubernetes:
				path = "/kubernetes"
			default:
				return m, nil
			}
			return m, m.fetchDataForPath(path)
		} else if m.mode == TableView {
			// Refresh table view
			m.notification = "⟳ Refreshing list..."
			m.notificationExpiry = time.Now().Add(2 * time.Second)
			m.mode = LoadingView
			var path string
			switch m.currentProduct {
			case ProductInstances:
				path = "/instances"
			case ProductKubernetes:
				path = "/kubernetes"
			default:
				return m, nil
			}
			return m, m.fetchDataForPath(path)
		}
		return m, nil

	case "d":
		// In Projects selection view: set selected project as default
		if m.mode == ProjectSelectView || m.currentProduct == ProductProjects {
			var project map[string]interface{}

			if m.mode == ProjectSelectView {
				// Use the selected row in project selection view
				selectedRow := m.table.Cursor()
				if selectedRow >= 0 && selectedRow < len(m.currentData) {
					project = m.currentData[selectedRow]
				}
			}

			if project != nil {
				projectID := getStringValue(project, "project_id", "")
				projectName := getStringValue(project, "description", projectID)
				if projectName == "" {
					projectName = projectID
				}
				return m, m.setDefaultProject(projectID, projectName)
			}
		} else {
			// In other views: toggle debug panel
			m.previousMode = m.mode
			m.mode = DebugView
			m.debugScrollOffset = 0
		}
		return m, nil

	case "delete", "backspace":
		// Delete instance - only in TableView for instances
		if m.mode == TableView && m.currentProduct == ProductInstances {
			selectedRow := m.table.Cursor()
			if selectedRow >= 0 && selectedRow < len(m.currentData) {
				m.deleteTarget = m.currentData[selectedRow]
				m.deleteConfirmInput = ""
				m.mode = DeleteConfirmView
			}
		}
		return m, nil
	}

	return m, nil
}

// handleDebugKeyPress handles key presses in debug view mode
func (m Model) handleDebugKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "d":
		// Close debug view and return to previous mode
		m.mode = m.previousMode
		m.debugScrollOffset = 0
		return m, nil

	case "q", "ctrl+c":
		return m, tea.Quit

	case "up", "k":
		// Scroll up (show older entries)
		entries := httpLib.BrowserDebugLogger.GetEntries()
		maxVisible := 15
		maxOffset := len(entries) - maxVisible
		if maxOffset < 0 {
			maxOffset = 0
		}
		if m.debugScrollOffset < maxOffset {
			m.debugScrollOffset++
		}
		return m, nil

	case "down", "j":
		// Scroll down (show newer entries)
		if m.debugScrollOffset > 0 {
			m.debugScrollOffset--
		}
		return m, nil

	case "c":
		// Clear debug logs
		httpLib.BrowserDebugLogger.Clear()
		m.debugScrollOffset = 0
		return m, nil
	}

	return m, nil
}

// handleDeleteConfirmKeyPress handles key presses in delete confirmation mode
func (m Model) handleDeleteConfirmKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		// Cancel delete and go back to table view
		m.mode = TableView
		m.deleteTarget = nil
		m.deleteConfirmInput = ""
		return m, nil

	case "enter":
		// Check if the input matches the instance name
		if m.deleteTarget != nil {
			instanceName, _ := m.deleteTarget["name"].(string)
			instanceId, _ := m.deleteTarget["id"].(string)
			if m.deleteConfirmInput == instanceName && instanceId != "" {
				// Proceed with deletion
				m.mode = LoadingView
				m.deleteConfirmInput = ""
				return m, m.deleteInstance(instanceId)
			}
		}
		return m, nil

	case "backspace":
		if len(m.deleteConfirmInput) > 0 {
			m.deleteConfirmInput = m.deleteConfirmInput[:len(m.deleteConfirmInput)-1]
		}
		return m, nil

	default:
		// Only accept printable characters for input
		if len(key) == 1 {
			m.deleteConfirmInput += key
		}
		return m, nil
	}
}

// handleFilterKeyPress handles key presses in filter mode for TableView
func (m Model) handleFilterKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc":
		// Cancel filter mode and clear filter
		m.filterMode = false
		m.filterInput = ""
		// Rebuild table with all data
		if m.currentProduct == ProductInstances {
			m.table = createInstancesTable(m.currentData, m.imageMap, m.floatingIPMap, m.width, m.height)
		}
		return m, nil

	case "enter":
		// Confirm filter and exit filter mode
		m.filterMode = false
		// Table is already filtered, just exit filter mode
		return m, nil

	case "backspace":
		if len(m.filterInput) > 0 {
			m.filterInput = m.filterInput[:len(m.filterInput)-1]
			// Rebuild table with filter
			m.applyTableFilter()
		}
		return m, nil

	default:
		// Accept printable characters
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.filterInput += key
			// Rebuild table with filter
			m.applyTableFilter()
		}
	}
	return m, nil
}

// applyTableFilter filters the table based on current filterInput
func (m *Model) applyTableFilter() {
	if m.filterInput == "" {
		// No filter, show all
		switch m.currentProduct {
		case ProductInstances:
			m.table = createInstancesTable(m.currentData, m.imageMap, m.floatingIPMap, m.width, m.height)
		case ProductKubernetes:
			m.table = createKubernetesTable(m.currentData, m.width, m.height)
		default:
			m.table = createGenericTable(m.currentData, m.width, m.height)
		}
		return
	}

	filter := strings.ToLower(m.filterInput)

	switch m.currentProduct {
	case ProductInstances:
		var filtered []map[string]interface{}
		for _, item := range m.currentData {
			name := strings.ToLower(getStringValue(item, "name", ""))
			status := strings.ToLower(getStringValue(item, "status", ""))
			region := strings.ToLower(getStringValue(item, "region", ""))
			if strings.Contains(name, filter) || strings.Contains(status, filter) || strings.Contains(region, filter) {
				filtered = append(filtered, item)
			}
		}
		m.table = createInstancesTable(filtered, m.imageMap, m.floatingIPMap, m.width, m.height)
	case ProductKubernetes:
		var filtered []map[string]interface{}
		for _, item := range m.currentData {
			name := strings.ToLower(getStringValue(item, "name", ""))
			status := strings.ToLower(getStringValue(item, "status", ""))
			region := strings.ToLower(getStringValue(item, "region", ""))
			version := strings.ToLower(getStringValue(item, "version", ""))
			if strings.Contains(name, filter) || strings.Contains(status, filter) || strings.Contains(region, filter) || strings.Contains(version, filter) {
				filtered = append(filtered, item)
			}
		}
		m.table = createKubernetesTable(filtered, m.width, m.height)
	default:
		m.table = createGenericTable(m.currentData, m.width, m.height)
	}
}

// handleWizardKeyPress handles key presses in wizard mode
func (m Model) handleWizardKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Handle cleanup confirmation mode
	if m.wizard.cleanupPending {
		return m.handleCleanupConfirmKeys(key)
	}

	// ctrl+c always quits
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// 'q' quits (except when typing in input fields)
	if key == "q" && m.wizard.step != WizardStepName && m.wizard.step != KubeWizardStepName && m.wizard.step != NodePoolWizardStepName && !m.wizard.filterMode && !m.wizard.creatingSSHKey && !m.wizard.creatingNetwork {
		return m, tea.Quit
	}

	// 'd' opens debug panel (except when typing in input fields)
	// Disable debug shortcut when: in name step, filter mode, creating SSH key, or creating network
	if key == "d" && m.wizard.step != WizardStepName && m.wizard.step != KubeWizardStepName && !m.wizard.filterMode && !m.wizard.creatingSSHKey && !m.wizard.creatingNetwork {
		m.previousMode = m.mode
		m.mode = DebugView
		m.debugScrollOffset = 0
		return m, nil
	}

	// Escape cancels the wizard and goes back to the product view
	// But if in filter mode, just exit filter mode
	if key == "esc" {
		if m.wizard.filterMode {
			m.wizard.filterMode = false
			m.wizard.filterInput = ""
			return m, nil
		}

		// Determine which product we were on and return to it
		returnPath := "/instances"
		if m.wizard.step >= 100 {
			// Kubernetes wizard
			returnPath = "/kubernetes"
		}

		m.wizard = WizardData{}
		m.mode = LoadingView
		return m, m.fetchDataForPath(returnPath)
	}

	// Handle each step differently
	switch m.wizard.step {
	case WizardStepRegion:
		return m.handleWizardRegionKeys(key, msg)
	case WizardStepFlavor:
		return m.handleWizardFlavorKeys(key, msg)
	case WizardStepImage:
		return m.handleWizardImageKeys(key, msg)
	case WizardStepSSHKey:
		return m.handleWizardSSHKeyKeys(key, msg)
	case WizardStepNetwork:
		return m.handleWizardNetworkKeys(key, msg)
	case WizardStepFloatingIP:
		return m.handleWizardFloatingIPKeys(key, msg)
	case WizardStepName:
		return m.handleWizardNameKeys(msg)
	case WizardStepConfirm:
		return m.handleWizardConfirmKeys(key)
	// Kubernetes wizard steps
	case KubeWizardStepRegion:
		return m.handleKubeWizardRegionKeys(key, msg)
	case KubeWizardStepVersion:
		return m.handleKubeWizardVersionKeys(key, msg)
	case KubeWizardStepNetwork:
		return m.handleKubeWizardNetworkKeys(key, msg)
	case KubeWizardStepSubnet:
		return m.handleKubeWizardSubnetKeys(key, msg)
	case KubeWizardStepName:
		return m.handleKubeWizardNameKeys(msg)
	case KubeWizardStepOptions:
		return m.handleKubeWizardOptionsKeys(key, msg)
	case KubeWizardStepConfirm:
		return m.handleKubeWizardConfirmKeys(key)
	// Node pool wizard steps
	case NodePoolWizardStepFlavor:
		return m.handleNodePoolWizardFlavorKeys(key)
	case NodePoolWizardStepName:
		return m.handleNodePoolWizardNameKeys(msg)
	case NodePoolWizardStepSize:
		return m.handleNodePoolWizardSizeKeys(key)
	case NodePoolWizardStepOptions:
		return m.handleNodePoolWizardOptionsKeys(key)
	case NodePoolWizardStepConfirm:
		return m.handleNodePoolWizardConfirmKeys(key)
	}

	return m, nil
}

// handleCleanupConfirmKeys handles key presses in cleanup confirmation mode
func (m Model) handleCleanupConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "left", "right":
		// Toggle between Yes and No
		if m.wizard.selectedIndex == 0 {
			m.wizard.selectedIndex = 1
		} else {
			m.wizard.selectedIndex = 0
		}
		return m, nil

	case "enter":
		if m.wizard.selectedIndex == 0 {
			// Yes, delete all - start cleanup
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Cleaning up resources..."
			m.wizard.cleanupPending = false
			m.notification = "🗑️ Cleaning up created resources..."
			m.notificationExpiry = time.Now().Add(30 * time.Second)
			return m, m.cleanupCreatedResources()
		} else {
			// No, keep them - just exit wizard
			m.notification = "⚠️ Resources kept. You may need to clean them up manually."
			m.notificationExpiry = time.Now().Add(5 * time.Second)
			m.wizard = WizardData{}
			m.mode = LoadingView
			return m, tea.Batch(
				m.fetchDataForPath("/instances"),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
					return clearNotificationMsg{}
				}),
			)
		}

	case "esc":
		// Same as No - keep resources
		m.notification = "⚠️ Resources kept. You may need to clean them up manually."
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.wizard = WizardData{}
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/instances"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
				return clearNotificationMsg{}
			}),
		)
	}

	return m, nil
}

// getFilteredWizardRegions returns filtered regions based on wizard filter input
func (m Model) getFilteredWizardRegions() []map[string]interface{} {
	if m.wizard.filterInput == "" {
		return m.wizard.regions
	}
	filter := strings.ToLower(m.wizard.filterInput)
	var filtered []map[string]interface{}
	for _, region := range m.wizard.regions {
		name := strings.ToLower(getString(region, "name"))
		location := strings.ToLower(getString(region, "datacenterLocation"))
		continent := strings.ToLower(getString(region, "continentCode"))
		if strings.Contains(name, filter) || strings.Contains(location, filter) || strings.Contains(continent, filter) {
			filtered = append(filtered, region)
		}
	}
	return filtered
}

// getFilteredWizardFlavors returns filtered flavors based on wizard filter input
func (m Model) getFilteredWizardFlavors() []map[string]interface{} {
	if m.wizard.filterInput == "" {
		return m.wizard.flavors
	}
	filter := strings.ToLower(m.wizard.filterInput)
	var filtered []map[string]interface{}
	for _, flavor := range m.wizard.flavors {
		name := strings.ToLower(getString(flavor, "name"))
		flavorType := strings.ToLower(getString(flavor, "type"))
		if strings.Contains(name, filter) || strings.Contains(flavorType, filter) {
			filtered = append(filtered, flavor)
		}
	}
	return filtered
}

// getFilteredWizardImages returns filtered images based on wizard filter input
func (m Model) getFilteredWizardImages() []map[string]interface{} {
	if m.wizard.filterInput == "" {
		return m.wizard.images
	}
	filter := strings.ToLower(m.wizard.filterInput)
	var filtered []map[string]interface{}
	for _, image := range m.wizard.images {
		name := strings.ToLower(getString(image, "name"))
		if strings.Contains(name, filter) {
			filtered = append(filtered, image)
		}
	}
	return filtered
}

// getFilteredWizardSSHKeys returns filtered SSH keys based on wizard filter input
func (m Model) getFilteredWizardSSHKeys() []map[string]interface{} {
	if m.wizard.filterInput == "" {
		return m.wizard.sshKeys
	}
	filter := strings.ToLower(m.wizard.filterInput)
	var filtered []map[string]interface{}
	for _, key := range m.wizard.sshKeys {
		name := strings.ToLower(getString(key, "name"))
		// Always include "(No SSH Key)" option
		if name == "(no ssh key)" || strings.Contains(name, filter) {
			filtered = append(filtered, key)
		}
	}
	return filtered
}

// getFilteredWizardNetworks returns filtered private networks based on wizard filter input
func (m Model) getFilteredWizardNetworks() []map[string]interface{} {
	if m.wizard.filterInput == "" {
		return m.wizard.privateNetworks
	}
	filter := strings.ToLower(m.wizard.filterInput)
	var filtered []map[string]interface{}
	for _, network := range m.wizard.privateNetworks {
		name := strings.ToLower(getString(network, "name"))
		id := getString(network, "id")
		// Always include special options
		if id == "__create_new__" || name == "(no private network)" || strings.Contains(name, filter) {
			filtered = append(filtered, network)
		}
	}
	return filtered
}

// getFilteredWizardFloatingIPs returns filtered floating IPs based on wizard filter input
func (m Model) getFilteredWizardFloatingIPs() []map[string]interface{} {
	if m.wizard.filterInput == "" {
		return m.wizard.floatingIPs
	}
	filter := strings.ToLower(m.wizard.filterInput)
	var filtered []map[string]interface{}
	for _, fip := range m.wizard.floatingIPs {
		name := strings.ToLower(getString(fip, "name"))
		ip := strings.ToLower(getString(fip, "ip"))
		id := getString(fip, "id")
		// Always include special options
		if id == "__none__" || id == "__create_new__" || strings.Contains(name, filter) || strings.Contains(ip, filter) {
			filtered = append(filtered, fip)
		}
	}
	return filtered
}

// handleWizardRegionKeys handles region step key presses
func (m Model) handleWizardRegionKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle filter mode
	if m.wizard.filterMode {
		switch key {
		case "enter":
			m.wizard.filterMode = false
			m.wizard.selectedIndex = 0
			return m, nil
		case "backspace":
			if len(m.wizard.filterInput) > 0 {
				m.wizard.filterInput = m.wizard.filterInput[:len(m.wizard.filterInput)-1]
				m.wizard.selectedIndex = 0
			} else {
				m.wizard.filterMode = false
			}
			return m, nil
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.wizard.filterInput += key
				m.wizard.selectedIndex = 0
			}
			return m, nil
		}
	}

	filtered := m.getFilteredWizardRegions()
	switch key {
	case "/":
		m.wizard.filterMode = true
		m.wizard.filterInput = ""
		return m, nil
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(filtered)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if m.wizard.selectedIndex < len(filtered) {
			region := filtered[m.wizard.selectedIndex]
			m.wizard.selectedRegion = getString(region, "name")
			m.wizard.step = WizardStepFlavor
			m.wizard.selectedIndex = 0
			m.wizard.filterInput = ""
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading flavors..."
			return m, m.fetchFlavors(m.wizard.selectedRegion)
		}
	}
	return m, nil
}

// handleWizardFlavorKeys handles flavor step key presses
func (m Model) handleWizardFlavorKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle filter mode
	if m.wizard.filterMode {
		switch key {
		case "enter":
			m.wizard.filterMode = false
			m.wizard.selectedIndex = 0
			return m, nil
		case "backspace":
			if len(m.wizard.filterInput) > 0 {
				m.wizard.filterInput = m.wizard.filterInput[:len(m.wizard.filterInput)-1]
				m.wizard.selectedIndex = 0
			} else {
				m.wizard.filterMode = false
			}
			return m, nil
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.wizard.filterInput += key
				m.wizard.selectedIndex = 0
			}
			return m, nil
		}
	}

	filtered := m.getFilteredWizardFlavors()
	switch key {
	case "/":
		m.wizard.filterMode = true
		m.wizard.filterInput = ""
		return m, nil
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(filtered)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if m.wizard.selectedIndex < len(filtered) {
			flavor := filtered[m.wizard.selectedIndex]
			m.wizard.selectedFlavor = getString(flavor, "id")
			m.wizard.selectedFlavorName = getString(flavor, "name")
			m.wizard.step = WizardStepImage
			m.wizard.selectedIndex = 0
			m.wizard.filterInput = ""
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading images..."
			return m, m.fetchImages(m.wizard.selectedRegion)
		}
	case "left":
		// Go back to region selection
		m.wizard.step = WizardStepRegion
		m.wizard.selectedIndex = 0
		m.wizard.filterInput = ""
	}
	return m, nil
}

// handleWizardImageKeys handles image step key presses
func (m Model) handleWizardImageKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle filter mode
	if m.wizard.filterMode {
		switch key {
		case "enter":
			m.wizard.filterMode = false
			m.wizard.selectedIndex = 0
			return m, nil
		case "backspace":
			if len(m.wizard.filterInput) > 0 {
				m.wizard.filterInput = m.wizard.filterInput[:len(m.wizard.filterInput)-1]
				m.wizard.selectedIndex = 0
			} else {
				m.wizard.filterMode = false
			}
			return m, nil
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.wizard.filterInput += key
				m.wizard.selectedIndex = 0
			}
			return m, nil
		}
	}

	filtered := m.getFilteredWizardImages()
	switch key {
	case "/":
		m.wizard.filterMode = true
		m.wizard.filterInput = ""
		return m, nil
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(filtered)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if m.wizard.selectedIndex < len(filtered) {
			image := filtered[m.wizard.selectedIndex]
			m.wizard.selectedImage = getString(image, "id")
			m.wizard.selectedImageName = getString(image, "name")
			// Go to SSH key selection
			m.wizard.step = WizardStepSSHKey
			m.wizard.selectedIndex = 0
			m.wizard.filterInput = ""
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading SSH keys..."
			return m, m.fetchSSHKeys()
		}
	case "left":
		// Go back to flavor selection
		m.wizard.step = WizardStepFlavor
		m.wizard.selectedIndex = 0
		m.wizard.filterInput = ""
	}
	return m, nil
}

// handleWizardSSHKeyKeys handles SSH key step key presses
func (m Model) handleWizardSSHKeyKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle SSH key creation mode
	if m.wizard.creatingSSHKey {
		return m.handleSSHKeyCreationKeys(key, msg)
	}

	// Handle filter mode
	if m.wizard.filterMode {
		switch key {
		case "enter":
			m.wizard.filterMode = false
			m.wizard.selectedIndex = 2 // Reset to first SSH key
			return m, nil
		case "backspace":
			if len(m.wizard.filterInput) > 0 {
				m.wizard.filterInput = m.wizard.filterInput[:len(m.wizard.filterInput)-1]
				m.wizard.selectedIndex = 2
			} else {
				m.wizard.filterMode = false
			}
			return m, nil
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.wizard.filterInput += key
				m.wizard.selectedIndex = 2
			}
			return m, nil
		}
	}

	filtered := m.getFilteredWizardSSHKeys()
	totalItems := len(filtered) + 2 // +2 for "Create new" and "No key" options

	switch key {
	case "/":
		m.wizard.filterMode = true
		m.wizard.filterInput = ""
		return m, nil
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < totalItems-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if m.wizard.selectedIndex == 0 {
			// Create new SSH key
			m.wizard.creatingSSHKey = true
			m.wizard.newSSHKeyName = ""
			m.wizard.newSSHKeyPublicKey = ""
			m.wizard.sshKeyCreateField = 0
			m.wizard.selectedLocalKeyIdx = 0
			m.wizard.localPubKeys = listLocalSSHPubKeys()
			return m, nil
		} else if m.wizard.selectedIndex == 1 {
			// No SSH key
			m.wizard.selectedSSHKey = ""
			m.wizard.selectedSSHKeyName = "(none)"
			m.wizard.step = WizardStepNetwork
			m.wizard.selectedIndex = 0
			m.wizard.filterInput = ""
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading networks..."
			return m, m.fetchPrivateNetworks()
		} else {
			// Existing SSH key selected
			sshKeyIdx := m.wizard.selectedIndex - 2
			if sshKeyIdx < len(filtered) {
				sshKey := filtered[sshKeyIdx]
				m.wizard.selectedSSHKey = getString(sshKey, "id")
				m.wizard.selectedSSHKeyName = getString(sshKey, "name")
				m.wizard.step = WizardStepNetwork
				m.wizard.selectedIndex = 0
				m.wizard.filterInput = ""
				m.wizard.isLoading = true
				m.wizard.loadingMessage = "Loading networks..."
				return m, m.fetchPrivateNetworks()
			}
		}
	case "left":
		// Go back to image selection
		m.wizard.step = WizardStepImage
		m.wizard.selectedIndex = 0
		m.wizard.filterInput = ""
	}
	return m, nil
}

// listLocalSSHPubKeys returns a list of .pub files in ~/.ssh
func listLocalSSHPubKeys() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	sshDir := home + "/.ssh"
	files, err := os.ReadDir(sshDir)
	if err != nil {
		return nil
	}
	var pubKeys []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".pub") {
			pubKeys = append(pubKeys, f.Name())
		}
	}
	return pubKeys
}

// readLocalSSHPubKey reads the content of a .pub file from ~/.ssh
func readLocalSSHPubKey(filename string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := home + "/.ssh/" + filename
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(content)), nil
}

// handleSSHKeyCreationKeys handles key presses in SSH key creation mode
func (m Model) handleSSHKeyCreationKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "tab", "down":
		if m.wizard.sshKeyCreateField == 0 {
			// Move from name to public key selection
			m.wizard.sshKeyCreateField = 1
			m.wizard.selectedLocalKeyIdx = 0
		} else if m.wizard.sshKeyCreateField == 1 {
			// In public key list, move down or go to buttons
			if m.wizard.selectedLocalKeyIdx < len(m.wizard.localPubKeys)-1 {
				m.wizard.selectedLocalKeyIdx++
			} else {
				m.wizard.sshKeyCreateField = 2
			}
		} else if m.wizard.sshKeyCreateField == 2 {
			m.wizard.sshKeyCreateField = 3
		} else {
			m.wizard.sshKeyCreateField = 0
		}
		return m, nil
	case "shift+tab", "up":
		if m.wizard.sshKeyCreateField == 0 {
			m.wizard.sshKeyCreateField = 3
		} else if m.wizard.sshKeyCreateField == 1 {
			// In public key list, move up or go to name
			if m.wizard.selectedLocalKeyIdx > 0 {
				m.wizard.selectedLocalKeyIdx--
			} else {
				m.wizard.sshKeyCreateField = 0
			}
		} else if m.wizard.sshKeyCreateField == 2 {
			m.wizard.sshKeyCreateField = 1
			if len(m.wizard.localPubKeys) > 0 {
				m.wizard.selectedLocalKeyIdx = len(m.wizard.localPubKeys) - 1
			}
		} else {
			m.wizard.sshKeyCreateField = 2
		}
		return m, nil
	case "enter":
		switch m.wizard.sshKeyCreateField {
		case 1: // Select public key file
			if m.wizard.selectedLocalKeyIdx >= 0 && m.wizard.selectedLocalKeyIdx < len(m.wizard.localPubKeys) {
				filename := m.wizard.localPubKeys[m.wizard.selectedLocalKeyIdx]
				content, err := readLocalSSHPubKey(filename)
				if err == nil {
					m.wizard.newSSHKeyPublicKey = content
					// Auto-fill name from filename if empty
					if m.wizard.newSSHKeyName == "" {
						baseName := strings.TrimSuffix(filename, ".pub")
						m.wizard.newSSHKeyName = baseName
					}
					m.wizard.sshKeyCreateField = 2 // Move to Create button
				}
			}
			return m, nil
		case 2: // Create button
			// Validate inputs
			if m.wizard.newSSHKeyName == "" {
				m.wizard.errorMsg = "SSH key name is required"
				return m, nil
			}
			if m.wizard.newSSHKeyPublicKey == "" {
				m.wizard.errorMsg = "Please select a public key file"
				return m, nil
			}
			// Create SSH key via API
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Creating SSH key..."
			return m, m.createSSHKey()
		case 3: // Cancel button
			m.wizard.creatingSSHKey = false
			m.wizard.newSSHKeyName = ""
			m.wizard.newSSHKeyPublicKey = ""
			m.wizard.sshKeyCreateField = 0
			return m, nil
		}
	case "esc":
		m.wizard.creatingSSHKey = false
		m.wizard.newSSHKeyName = ""
		m.wizard.newSSHKeyPublicKey = ""
		m.wizard.sshKeyCreateField = 0
		return m, nil
	case "backspace":
		if m.wizard.sshKeyCreateField == 0 && len(m.wizard.newSSHKeyName) > 0 {
			m.wizard.newSSHKeyName = m.wizard.newSSHKeyName[:len(m.wizard.newSSHKeyName)-1]
		}
		return m, nil
	default:
		// Handle text input for name field only
		if m.wizard.sshKeyCreateField == 0 {
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.wizard.newSSHKeyName += key
			}
		}
	}
	return m, nil
}

// handleWizardNetworkKeys handles network configuration step key presses
func (m Model) handleWizardNetworkKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// If in network creation mode, delegate to that handler
	if m.wizard.creatingNetwork {
		return m.handleNetworkCreationKeys(key)
	}

	// Handle filter mode (only for private network list, not public toggle)
	if m.wizard.filterMode && m.wizard.networkMenuIndex == 1 {
		switch key {
		case "enter":
			m.wizard.filterMode = false
			m.wizard.selectedIndex = 0
			return m, nil
		case "backspace":
			if len(m.wizard.filterInput) > 0 {
				m.wizard.filterInput = m.wizard.filterInput[:len(m.wizard.filterInput)-1]
				m.wizard.selectedIndex = 0
			} else {
				m.wizard.filterMode = false
			}
			return m, nil
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.wizard.filterInput += key
				m.wizard.selectedIndex = 0
			}
			return m, nil
		}
	}

	filtered := m.getFilteredWizardNetworks()
	switch key {
	case "/":
		// Only allow filter when in private network list
		if m.wizard.networkMenuIndex == 1 {
			m.wizard.filterMode = true
			m.wizard.filterInput = ""
		}
		return m, nil
	case "up", "k":
		if m.wizard.networkMenuIndex == 1 {
			if m.wizard.selectedIndex > 0 {
				m.wizard.selectedIndex--
			} else {
				// Move back to public toggle
				m.wizard.networkMenuIndex = 0
			}
		}
	case "down", "j":
		if m.wizard.networkMenuIndex == 0 {
			// Move to private network list if available
			if len(filtered) > 0 {
				m.wizard.networkMenuIndex = 1
				m.wizard.selectedIndex = 0
			}
		} else if m.wizard.selectedIndex < len(filtered)-1 {
			m.wizard.selectedIndex++
		}
	case " ":
		// Space toggles public network when on that menu item
		if m.wizard.networkMenuIndex == 0 {
			m.wizard.usePublicNetwork = !m.wizard.usePublicNetwork
		}
	case "enter":
		if m.wizard.networkMenuIndex == 0 {
			// When on public toggle, continue to next step
			// If no public network and we'll select a private network, we need floating IP step
			m.wizard.step = WizardStepName
			m.wizard.nameInput = ""
			m.wizard.filterInput = ""
		} else {
			// Check if "Create new" is selected
			if m.wizard.selectedIndex < len(filtered) {
				network := filtered[m.wizard.selectedIndex]
				networkId := getString(network, "id")

				if networkId == "__create_new__" {
					// Enter network creation mode
					m.wizard.creatingNetwork = true
					m.wizard.newNetworkName = ""
					m.wizard.newNetworkVlanId = rand.Intn(4094) + 1 // Random VLAN ID 1-4094
					m.wizard.newNetworkCIDR = "10.0.0.0/24"
					m.wizard.newNetworkDHCP = true
					m.wizard.networkCreateField = 0
					return m, nil
				}

				// Select existing network
				m.wizard.selectedPrivateNetwork = networkId
				m.wizard.selectedPrivateNetworkName = getString(network, "name")
				// Store subnet ID if available - handle both []interface{} and []map[string]interface{}
				m.wizard.selectedSubnetId = ""
				if subnets, ok := network["subnets"].([]map[string]interface{}); ok && len(subnets) > 0 {
					m.wizard.selectedSubnetId = getString(subnets[0], "id")
				} else if subnets, ok := network["subnets"].([]interface{}); ok && len(subnets) > 0 {
					if subnet, ok := subnets[0].(map[string]interface{}); ok {
						m.wizard.selectedSubnetId = getString(subnet, "id")
					}
				}
				// Handle "(No Private Network)" option
				if m.wizard.selectedPrivateNetworkName == "(No Private Network)" {
					m.wizard.selectedPrivateNetwork = ""
					m.wizard.selectedPrivateNetworkName = ""
					m.wizard.selectedSubnetId = ""
				}
			}

			// Decide next step based on network configuration
			if !m.wizard.usePublicNetwork && m.wizard.selectedPrivateNetwork != "" {
				// Private network only - go to floating IP step
				m.wizard.step = WizardStepFloatingIP
				m.wizard.selectedIndex = 0
				m.wizard.filterInput = ""
				m.wizard.isLoading = true
				m.wizard.loadingMessage = "Loading floating IPs..."
				return m, m.fetchFloatingIPs()
			}

			// Go to name input
			m.wizard.step = WizardStepName
			m.wizard.nameInput = ""
			m.wizard.filterInput = ""
		}
	case "left":
		// Go back to SSH key selection and reload SSH keys
		m.wizard.step = WizardStepSSHKey
		m.wizard.selectedIndex = 0
		m.wizard.filterInput = ""
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading SSH keys..."
		return m, m.fetchSSHKeys()
	}
	return m, nil
}

// handleWizardFloatingIPKeys handles floating IP step key presses
func (m Model) handleWizardFloatingIPKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle filter mode
	if m.wizard.filterMode {
		switch key {
		case "enter":
			m.wizard.filterMode = false
			m.wizard.selectedIndex = 0
			return m, nil
		case "backspace":
			if len(m.wizard.filterInput) > 0 {
				m.wizard.filterInput = m.wizard.filterInput[:len(m.wizard.filterInput)-1]
				m.wizard.selectedIndex = 0
			} else {
				m.wizard.filterMode = false
			}
			return m, nil
		default:
			if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
				m.wizard.filterInput += key
				m.wizard.selectedIndex = 0
			}
			return m, nil
		}
	}

	filtered := m.getFilteredWizardFloatingIPs()
	switch key {
	case "/":
		m.wizard.filterMode = true
		m.wizard.filterInput = ""
		return m, nil
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(filtered)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if m.wizard.selectedIndex < len(filtered) {
			fip := filtered[m.wizard.selectedIndex]
			fipId := getString(fip, "id")

			if fipId == "__none__" {
				// No floating IP
				m.wizard.selectedFloatingIP = ""
				m.wizard.selectedFloatingIPAddress = ""
			} else if fipId == "__create_new__" {
				// Will create new floating IP
				m.wizard.selectedFloatingIP = "__create_new__"
				m.wizard.selectedFloatingIPAddress = "(new)"
			} else {
				// Use existing floating IP
				m.wizard.selectedFloatingIP = fipId
				m.wizard.selectedFloatingIPAddress = getString(fip, "ip")
			}

			// Go to name input
			m.wizard.step = WizardStepName
			m.wizard.nameInput = ""
			m.wizard.filterInput = ""
		}
	case "left":
		// Go back to network configuration
		m.wizard.step = WizardStepNetwork
		m.wizard.selectedIndex = 0
		m.wizard.filterInput = ""
		m.wizard.networkMenuIndex = 1
	}
	return m, nil
}

// handleNetworkCreationKeys handles key presses in network creation sub-form
func (m Model) handleNetworkCreationKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up", "k":
		if m.wizard.networkCreateField > 0 {
			m.wizard.networkCreateField--
		}
	case "down", "j":
		if m.wizard.networkCreateField < 4 {
			m.wizard.networkCreateField++
		}
	case "tab":
		m.wizard.networkCreateField = (m.wizard.networkCreateField + 1) % 5
	case " ":
		// Toggle DHCP when on DHCP field (field 3)
		if m.wizard.networkCreateField == 3 {
			m.wizard.newNetworkDHCP = !m.wizard.newNetworkDHCP
		}
	case "enter":
		if m.wizard.networkCreateField == 4 {
			// Create button - validate and create
			if m.wizard.newNetworkName == "" {
				m.wizard.errorMsg = "Network name is required"
				return m, nil
			}
			if m.wizard.newNetworkCIDR == "" {
				m.wizard.newNetworkCIDR = "10.0.0.0/24"
			}
			if m.wizard.newNetworkVlanId < 1 || m.wizard.newNetworkVlanId > 4094 {
				m.wizard.errorMsg = "VLAN ID must be between 1 and 4094"
				return m, nil
			}
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Creating network..."
			m.wizard.errorMsg = ""
			return m, m.createPrivateNetwork()
		}
		// On other fields, move to next field
		if m.wizard.networkCreateField < 4 {
			m.wizard.networkCreateField++
		}
	case "esc":
		// Exit creation mode
		m.wizard.creatingNetwork = false
		m.wizard.errorMsg = ""
	case "backspace":
		// Handle backspace based on current field
		switch m.wizard.networkCreateField {
		case 0: // Name field
			if len(m.wizard.newNetworkName) > 0 {
				m.wizard.newNetworkName = m.wizard.newNetworkName[:len(m.wizard.newNetworkName)-1]
				return m, nil
			}
		case 1: // VLAN ID field
			if m.wizard.newNetworkVlanId >= 10 {
				m.wizard.newNetworkVlanId = m.wizard.newNetworkVlanId / 10
				return m, nil
			} else if m.wizard.newNetworkVlanId > 0 {
				m.wizard.newNetworkVlanId = 0
				return m, nil
			}
		case 2: // CIDR field
			if len(m.wizard.newNetworkCIDR) > 0 {
				m.wizard.newNetworkCIDR = m.wizard.newNetworkCIDR[:len(m.wizard.newNetworkCIDR)-1]
				return m, nil
			}
		}
		// If field is empty, exit creation mode
		m.wizard.creatingNetwork = false
		m.wizard.errorMsg = ""
	default:
		// Handle text input for fields
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			switch m.wizard.networkCreateField {
			case 0: // Name field
				m.wizard.newNetworkName += key
			case 1: // VLAN ID field - only digits
				if key[0] >= '0' && key[0] <= '9' {
					newVal := m.wizard.newNetworkVlanId*10 + int(key[0]-'0')
					if newVal <= 4094 {
						m.wizard.newNetworkVlanId = newVal
					}
				}
			case 2: // CIDR field
				// Only allow valid CIDR characters
				if (key[0] >= '0' && key[0] <= '9') || key[0] == '.' || key[0] == '/' {
					m.wizard.newNetworkCIDR += key
				}
			}
		}
	}
	return m, nil
}

// handleWizardNameKeys handles name input step key presses
func (m Model) handleWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "enter":
		if m.wizard.nameInput != "" {
			m.wizard.instanceName = m.wizard.nameInput
			m.wizard.step = WizardStepConfirm
			m.wizard.selectedIndex = 0 // 0 = Create, 1 = Cancel
		}
	case "left":
		// Go back to appropriate step based on configuration
		if !m.wizard.usePublicNetwork && m.wizard.selectedPrivateNetwork != "" {
			// Go back to floating IP step
			m.wizard.step = WizardStepFloatingIP
			m.wizard.selectedIndex = 0
		} else {
			// Go back to network configuration
			m.wizard.step = WizardStepNetwork
			m.wizard.selectedIndex = 0
			m.wizard.networkMenuIndex = 0
		}
	case "backspace":
		if len(m.wizard.nameInput) > 0 {
			m.wizard.nameInput = m.wizard.nameInput[:len(m.wizard.nameInput)-1]
		}
	default:
		// Accept printable characters for the name
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.wizard.nameInput += key
		}
	}
	return m, nil
}

// handleWizardConfirmKeys handles confirmation step key presses
func (m Model) handleWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	// Prevent multiple submissions while loading
	if m.wizard.isLoading {
		return m, nil
	}

	switch key {
	case "left", "right", "tab":
		// Toggle between Create and Cancel
		if m.wizard.selectedIndex == 0 {
			m.wizard.selectedIndex = 1
		} else {
			m.wizard.selectedIndex = 0
		}
	case "enter":
		if m.wizard.selectedIndex == 0 {
			// Create the instance
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Creating instance..."
			return m, m.createInstance()
		} else {
			// Cancel and go back to instances view
			m.wizard = WizardData{}
			m.mode = LoadingView
			return m, m.fetchDataForPath("/instances")
		}
	case "backspace":
		// Go back to name input
		m.wizard.step = WizardStepName
	}
	return m, nil
}

// Kubernetes wizard key handlers

// handleKubeWizardRegionKeys handles key presses in Kubernetes region selection step
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
		if m.wizard.selectedIndex >= 0 && m.wizard.selectedIndex < len(m.wizard.kubeRegions) {
			region := m.wizard.kubeRegions[m.wizard.selectedIndex]
			if code, ok := region["code"].(string); ok {
				m.wizard.selectedKubeRegion = code
			}
			m.wizard.selectedIndex = 0
			m.wizard.step = KubeWizardStepVersion
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading Kubernetes versions..."
			return m, m.fetchKubeVersions(m.wizard.selectedKubeRegion)
		}
	case "backspace":
		// Go back (but region is first step, so cancel wizard)
		m.wizard = WizardData{}
		m.mode = LoadingView
		return m, m.fetchDataForPath("/kubernetes")
	}
	return m, nil
}

// handleKubeWizardVersionKeys handles key presses in Kubernetes version selection step
func (m Model) handleKubeWizardVersionKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down":
		if m.wizard.selectedIndex < len(m.wizard.kubeVersions)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if m.wizard.selectedIndex >= 0 && m.wizard.selectedIndex < len(m.wizard.kubeVersions) {
			m.wizard.selectedKubeVersion = m.wizard.kubeVersions[m.wizard.selectedIndex]
			m.wizard.selectedIndex = 0
			m.wizard.step = KubeWizardStepNetwork
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading networks..."
			return m, m.fetchKubeNetworks()
		}
	case "backspace":
		m.wizard.step = KubeWizardStepRegion
		m.wizard.selectedIndex = 0
	}
	return m, nil
}

// handleKubeWizardNetworkKeys handles key presses in Kubernetes network selection step
func (m Model) handleKubeWizardNetworkKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxIdx := len(m.wizard.kubeNetworks)

	switch key {
	case "up":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down":
		if m.wizard.selectedIndex < maxIdx {
			m.wizard.selectedIndex++
		}
	case "enter":
		if m.wizard.selectedIndex == 0 {
			// No private network selected
			m.wizard.selectedKubeNetwork = ""
			m.wizard.selectedKubeNetworkName = ""
			m.wizard.step = KubeWizardStepName
			m.wizard.selectedIndex = 0
			m.wizard.kubeNameInput = ""
		} else {
			// Private network selected
			netIdx := m.wizard.selectedIndex - 1
			if netIdx >= 0 && netIdx < len(m.wizard.kubeNetworks) {
				network := m.wizard.kubeNetworks[netIdx]
				if id, ok := network["id"].(string); ok {
					m.wizard.selectedKubeNetwork = id
				}
				if name, ok := network["name"].(string); ok {
					m.wizard.selectedKubeNetworkName = name
				}
				// Load subnets for the selected network
				m.wizard.step = KubeWizardStepSubnet
				m.wizard.selectedIndex = 0
				m.wizard.kubeSubnetMenuIndex = 0
				m.wizard.isLoading = true
				m.wizard.loadingMessage = "Loading subnets..."
				return m, m.fetchKubeSubnets(m.wizard.selectedKubeNetwork)
			}
		}
	case "backspace":
		m.wizard.step = KubeWizardStepVersion
		m.wizard.selectedIndex = 0
	}
	return m, nil
}

// handleKubeWizardSubnetKeys handles key presses in Kubernetes subnet selection step
func (m Model) handleKubeWizardSubnetKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	subnetCount := len(m.wizard.kubeSubnets)

	switch key {
	case "up":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down":
		maxIdx := subnetCount
		if m.wizard.kubeSubnetMenuIndex == 1 {
			maxIdx++ // One extra option for "same as nodes"
		}
		if m.wizard.selectedIndex < maxIdx-1 {
			m.wizard.selectedIndex++
		}
	case "tab":
		// Switch between nodes and LB subnet selection
		if m.wizard.kubeSubnetMenuIndex == 0 {
			m.wizard.kubeSubnetMenuIndex = 1
			m.wizard.selectedIndex = 0
		} else {
			// Done with subnet selection, go to name step
			m.wizard.step = KubeWizardStepName
			m.wizard.selectedIndex = 0
			m.wizard.kubeNameInput = ""
			return m, nil
		}
	case "enter":
		if m.wizard.kubeSubnetMenuIndex == 0 {
			// Select nodes subnet
			if m.wizard.selectedIndex >= 0 && m.wizard.selectedIndex < subnetCount {
				subnet := m.wizard.kubeSubnets[m.wizard.selectedIndex]
				if cidr, ok := subnet["cidr"].(string); ok {
					m.wizard.selectedNodesSubnet = cidr
				}
				if id, ok := subnet["id"].(string); ok {
					m.wizard.selectedNodesSubnetCIDR = id
				}
			}
			// Move to LB subnet selection
			m.wizard.kubeSubnetMenuIndex = 1
			m.wizard.selectedIndex = 0
		} else {
			// Select LB subnet
			if m.wizard.selectedIndex == 0 {
				// Use same as nodes subnet
				m.wizard.selectedLBSubnet = m.wizard.selectedNodesSubnet
				m.wizard.selectedLBSubnetCIDR = m.wizard.selectedNodesSubnetCIDR
			} else {
				subnetIdx := m.wizard.selectedIndex - 1
				if subnetIdx >= 0 && subnetIdx < subnetCount {
					subnet := m.wizard.kubeSubnets[subnetIdx]
					if cidr, ok := subnet["cidr"].(string); ok {
						m.wizard.selectedLBSubnet = cidr
					}
					if id, ok := subnet["id"].(string); ok {
						m.wizard.selectedLBSubnetCIDR = id
					}
				}
			}
			// Move to name step
			m.wizard.step = KubeWizardStepName
			m.wizard.selectedIndex = 0
			m.wizard.kubeNameInput = ""
		}
	case "backspace":
		if m.wizard.kubeSubnetMenuIndex == 1 {
			// Go back to nodes subnet selection
			m.wizard.kubeSubnetMenuIndex = 0
			m.wizard.selectedIndex = 0
		} else {
			// Go back to network selection
			m.wizard.step = KubeWizardStepNetwork
			m.wizard.selectedIndex = 0
		}
	}
	return m, nil
}

// handleKubeWizardNameKeys handles key presses in Kubernetes cluster name input step
func (m Model) handleKubeWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		// Validate name
		name := strings.TrimSpace(m.wizard.kubeNameInput)
		if len(name) >= 3 && len(name) <= 32 {
			m.wizard.kubeName = name
			m.wizard.step = KubeWizardStepOptions
			m.wizard.selectedIndex = 0
			m.wizard.kubeOptionsFieldIndex = 0
			// Set default values for options
			if m.wizard.kubePlan == "" {
				m.wizard.kubePlan = "free"
			}
			if m.wizard.kubeUpdatePolicy == "" {
				m.wizard.kubeUpdatePolicy = "ALWAYS_UPDATE"
			}
			if m.wizard.kubeProxyMode == "" {
				m.wizard.kubeProxyMode = "iptables"
			}
		}
	case tea.KeyBackspace:
		if len(m.wizard.kubeNameInput) > 0 {
			m.wizard.kubeNameInput = m.wizard.kubeNameInput[:len(m.wizard.kubeNameInput)-1]
		}
	case tea.KeyRunes:
		// Allow alphanumeric and hyphens
		for _, r := range msg.Runes {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				if len(m.wizard.kubeNameInput) < 32 {
					m.wizard.kubeNameInput += string(r)
				}
			}
		}
	}
	return m, nil
}

// handleKubeWizardOptionsKeys handles key presses in Kubernetes advanced options step
func (m Model) handleKubeWizardOptionsKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	maxFields := 4
	if m.wizard.kubePrivateRouting {
		maxFields = 5
	}

	switch key {
	case "up":
		if m.wizard.kubeOptionsFieldIndex > 0 {
			m.wizard.kubeOptionsFieldIndex--
		}
	case "down":
		if m.wizard.kubeOptionsFieldIndex < maxFields-1 {
			m.wizard.kubeOptionsFieldIndex++
		}
	case "left", "right":
		// Toggle or cycle values depending on the field
		switch m.wizard.kubeOptionsFieldIndex {
		case 0: // Plan
			if m.wizard.kubePlan == "free" {
				m.wizard.kubePlan = "standard"
			} else {
				m.wizard.kubePlan = "free"
			}
		case 1: // Update policy
			if m.wizard.kubeUpdatePolicy == "ALWAYS_UPDATE" {
				m.wizard.kubeUpdatePolicy = "NEVER_UPDATE"
			} else {
				m.wizard.kubeUpdatePolicy = "ALWAYS_UPDATE"
			}
		case 2: // Kube-proxy mode
			if m.wizard.kubeProxyMode == "iptables" {
				m.wizard.kubeProxyMode = "ipvs"
			} else {
				m.wizard.kubeProxyMode = "iptables"
			}
		case 3: // Private routing toggle
			m.wizard.kubePrivateRouting = !m.wizard.kubePrivateRouting
			if m.wizard.kubePrivateRouting && m.wizard.kubeGatewayIP == "" {
				m.wizard.kubeGatewayIPInput = ""
			}
		case 4: // Gateway IP input (if private routing enabled)
			// Handle as text input
			if msg.Type == tea.KeyRunes {
				for _, r := range msg.Runes {
					if (r >= '0' && r <= '9') || r == '.' {
						if len(m.wizard.kubeGatewayIPInput) < 15 {
							m.wizard.kubeGatewayIPInput += string(r)
						}
					}
				}
			} else if msg.Type == tea.KeyBackspace {
				if len(m.wizard.kubeGatewayIPInput) > 0 {
					m.wizard.kubeGatewayIPInput = m.wizard.kubeGatewayIPInput[:len(m.wizard.kubeGatewayIPInput)-1]
				}
			}
		}
	case "enter":
		// Validate gateway IP if private routing is enabled
		if m.wizard.kubePrivateRouting && m.wizard.kubeGatewayIPInput != "" {
			m.wizard.kubeGatewayIP = m.wizard.kubeGatewayIPInput
		}
		m.wizard.step = KubeWizardStepConfirm
		m.wizard.kubeConfirmButtonIndex = 0
	case "backspace":
		if m.wizard.kubeOptionsFieldIndex == 4 {
			// Clear gateway IP input if focused
			if len(m.wizard.kubeGatewayIPInput) > 0 {
				m.wizard.kubeGatewayIPInput = m.wizard.kubeGatewayIPInput[:len(m.wizard.kubeGatewayIPInput)-1]
			}
		} else {
			// Go back to name step
			m.wizard.step = KubeWizardStepName
			m.wizard.kubeNameInput = m.wizard.kubeName
		}
	}
	return m, nil
}

// handleKubeWizardConfirmKeys handles key presses in Kubernetes confirmation step
func (m Model) handleKubeWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}

	switch key {
	case "left", "right", "tab":
		// Toggle between Create and Cancel
		if m.wizard.kubeConfirmButtonIndex == 0 {
			m.wizard.kubeConfirmButtonIndex = 1
		} else {
			m.wizard.kubeConfirmButtonIndex = 0
		}
	case "enter":
		if m.wizard.kubeConfirmButtonIndex == 0 {
			// Create the cluster
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Creating Kubernetes cluster..."
			return m, m.createKubeClusterWrapper()
		} else {
			// Cancel and go back to Kubernetes view
			m.wizard = WizardData{}
			m.mode = LoadingView
			return m, m.fetchDataForPath("/kubernetes")
		}
	case "backspace":
		// Go back to options step
		m.wizard.step = KubeWizardStepOptions
		m.wizard.kubeOptionsFieldIndex = 0
	}
	return m, nil
}

// createKubeClusterWrapper wraps the cluster creation with proper data formatting
func (m Model) createKubeClusterWrapper() tea.Cmd {
	// Build the creation payload
	payload := map[string]interface{}{
		"name":    m.wizard.kubeName,
		"region":  m.wizard.selectedKubeRegion,
		"version": m.wizard.selectedKubeVersion,
		"plan":    m.wizard.kubePlan,
	}

	// Add network if selected
	if m.wizard.selectedKubeNetwork != "" {
		payload["privateNetworkId"] = m.wizard.selectedKubeNetwork
		payload["nodesSubnetId"] = m.wizard.selectedNodesSubnetCIDR
		if m.wizard.selectedLBSubnetCIDR != "" {
			payload["loadBalancersSubnetId"] = m.wizard.selectedLBSubnetCIDR
		}
	}

	// Add advanced options
	payload["updatePolicy"] = m.wizard.kubeUpdatePolicy
	payload["kubeProxyMode"] = m.wizard.kubeProxyMode

	if m.wizard.kubePrivateRouting && m.wizard.kubeGatewayIP != "" {
		payload["privateNetworkRouting"] = true
		payload["gatewayIP"] = m.wizard.kubeGatewayIP
	}

	return m.createKubeCluster(payload)
}

// creationWizardMsg is sent when the creation wizard should be launched
type creationWizardMsg struct {
	product      ProductType
	cloudProject string
}

// launchCreationWizard prepares to exit the browser and launch the creation command
func (m Model) launchCreationWizard() tea.Cmd {
	return func() tea.Msg {
		return creationWizardMsg{
			product:      m.currentProduct,
			cloudProject: m.cloudProject,
		}
	}
}

// setDefaultProject saves the project ID as the default cloud project
func (m Model) setDefaultProject(projectID, projectName string) tea.Cmd {
	return func() tea.Msg {
		err := config.SetConfigValue(flags.CliConfig, flags.CliConfigPath, "", "default_cloud_project", projectID)
		return setDefaultProjectMsg{
			projectID:   projectID,
			projectName: projectName,
			err:         err,
		}
	}
}

func (m Model) loadCurrentProduct() (Model, tea.Cmd) {
	navItems := getNavItems()
	currentNav := navItems[m.navIdx]
	m.currentProduct = currentNav.Product
	m.mode = LoadingView
	m.detailData = nil
	m.currentData = nil

	// For instances and Kubernetes, start the auto-refresh timer
	if currentNav.Product == ProductInstances || currentNav.Product == ProductKubernetes {
		return m, tea.Batch(
			m.fetchDataForPath(currentNav.Path),
			m.scheduleRefresh(),
		)
	}
	return m, m.fetchDataForPath(currentNav.Path)
}

// Helper functions
func getStringValue(data map[string]interface{}, key string, defaultVal string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
		return fmt.Sprintf("%v", val)
	}
	return defaultVal
}

func getFloatValue(data map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := data[key]; ok {
		if f, ok := val.(float64); ok {
			return f
		}
	}
	return defaultVal
}

func getBoolValue(data map[string]interface{}, key string, defaultVal bool) bool {
	if val, ok := data[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultVal
}

// getIntOrFloatValue extracts a numeric value that could be int or float64 in JSON
func getIntOrFloatValue(data map[string]interface{}, key string, defaultVal float64) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		case json.Number:
			// Handle json.Number type
			if f, err := v.Float64(); err == nil {
				return f
			}
		default:
			// Try to parse as string representation of number
			if str, ok := val.(string); ok {
				var f float64
				fmt.Sscanf(str, "%f", &f)
				return f
			}
		}
	}
	return defaultVal
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// ========== Node Pool Wizard Key Handlers ==========

func (m Model) handleNodePoolWizardFlavorKeys(key string) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}

	switch key {
	case "up":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down":
		if m.wizard.selectedIndex < len(m.wizard.nodePoolFlavors)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if len(m.wizard.nodePoolFlavors) > 0 {
			flavor := m.wizard.nodePoolFlavors[m.wizard.selectedIndex]
			m.wizard.nodePoolFlavorName = getString(flavor, "name")
			m.wizard.step = NodePoolWizardStepName
		}
	case "esc":
		m.wizard = WizardData{}
		m.mode = DetailView
	}
	return m, nil
}

func (m Model) handleNodePoolWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyRunes:
		m.wizard.nodePoolNameInput += string(msg.Runes)
		m.wizard.errorMsg = "" // Clear error on typing
	case tea.KeyBackspace:
		if len(m.wizard.nodePoolNameInput) > 0 {
			m.wizard.nodePoolNameInput = m.wizard.nodePoolNameInput[:len(m.wizard.nodePoolNameInput)-1]
		}
		m.wizard.errorMsg = "" // Clear error on typing
	case tea.KeyEnter:
		if m.wizard.nodePoolNameInput != "" {
			m.wizard.nodePoolName = m.wizard.nodePoolNameInput
			m.wizard.errorMsg = "" // Clear error when moving forward
			m.wizard.step = NodePoolWizardStepSize
		}
	case tea.KeyEscape:
		m.wizard.step = NodePoolWizardStepFlavor
		m.wizard.selectedIndex = 0
		m.wizard.errorMsg = "" // Clear error when going back
	}
	return m, nil
}

func (m Model) handleNodePoolWizardSizeKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		if m.wizard.nodePoolSizeFieldIndex > 0 {
			m.wizard.nodePoolSizeFieldIndex--
		}
	case "down":
		if m.wizard.nodePoolSizeFieldIndex < 2 {
			m.wizard.nodePoolSizeFieldIndex++
		}
	case "left":
		switch m.wizard.nodePoolSizeFieldIndex {
		case 0:
			if m.wizard.nodePoolDesiredNodes > 1 {
				m.wizard.nodePoolDesiredNodes--
			}
		case 1:
			if m.wizard.nodePoolMinNodes > 0 {
				m.wizard.nodePoolMinNodes--
			}
		case 2:
			if m.wizard.nodePoolMaxNodes > m.wizard.nodePoolDesiredNodes {
				m.wizard.nodePoolMaxNodes--
			}
		}
	case "right":
		switch m.wizard.nodePoolSizeFieldIndex {
		case 0:
			if m.wizard.nodePoolDesiredNodes < m.wizard.nodePoolMaxNodes {
				m.wizard.nodePoolDesiredNodes++
			}
		case 1:
			if m.wizard.nodePoolMinNodes < m.wizard.nodePoolDesiredNodes {
				m.wizard.nodePoolMinNodes++
			}
		case 2:
			if m.wizard.nodePoolMaxNodes < 100 {
				m.wizard.nodePoolMaxNodes++
			}
		}
	case "enter":
		m.wizard.step = NodePoolWizardStepOptions
		m.wizard.nodePoolOptionsFieldIdx = 0
	case "esc":
		m.wizard.step = NodePoolWizardStepName
	}
	return m, nil
}

func (m Model) handleNodePoolWizardOptionsKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "up":
		if m.wizard.nodePoolOptionsFieldIdx > 0 {
			m.wizard.nodePoolOptionsFieldIdx--
		}
	case "down":
		if m.wizard.nodePoolOptionsFieldIdx < 2 {
			m.wizard.nodePoolOptionsFieldIdx++
		}
	case " ":
		switch m.wizard.nodePoolOptionsFieldIdx {
		case 0:
			m.wizard.nodePoolAutoscale = !m.wizard.nodePoolAutoscale
		case 1:
			m.wizard.nodePoolAntiAffinity = !m.wizard.nodePoolAntiAffinity
		case 2:
			m.wizard.nodePoolMonthlyBilled = !m.wizard.nodePoolMonthlyBilled
		}
	case "enter":
		m.wizard.step = NodePoolWizardStepConfirm
		m.wizard.nodePoolConfirmBtnIdx = 0
	case "esc":
		m.wizard.step = NodePoolWizardStepSize
	}
	return m, nil
}

func (m Model) handleNodePoolWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}

	switch key {
	case "left", "right", "tab":
		if m.wizard.nodePoolConfirmBtnIdx == 0 {
			m.wizard.nodePoolConfirmBtnIdx = 1
		} else {
			m.wizard.nodePoolConfirmBtnIdx = 0
		}
	case "enter":
		if m.wizard.nodePoolConfirmBtnIdx == 0 {
			// Create the node pool
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Creating node pool..."
			return m, m.createNodePool()
		} else {
			// Cancel
			m.wizard = WizardData{}
			m.mode = DetailView
		}
	case "esc":
		m.wizard.step = NodePoolWizardStepOptions
	}
	return m, nil
}

// getNumericValue extracts a numeric value from a map, handling json.Number type
func getNumericValue(data map[string]interface{}, key string) float64 {
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case json.Number:
			f, _ := v.Float64()
			return f
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func getDefaultCloudProject() (string, error) {
	projectID, err := config.GetConfigValue(flags.CliConfig, "", "default_cloud_project")
	if err != nil || projectID == "" {
		return "", err
	}
	return projectID, nil
}

// handleKubeUpgradeKeyPress handles keyboard input for the Kubernetes upgrade view
func (m Model) handleKubeUpgradeKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.mode = DetailView
		return m, nil

	case "up", "k":
		if m.wizard.kubeUpgradeSelectedIdx > 0 {
			m.wizard.kubeUpgradeSelectedIdx--
		}
		return m, nil

	case "down", "j":
		if m.wizard.kubeUpgradeSelectedIdx < len(m.wizard.kubeUpgradeVersions)-1 {
			m.wizard.kubeUpgradeSelectedIdx++
		}
		return m, nil

	case "enter":
		if len(m.wizard.kubeUpgradeVersions) > 0 {
			selectedVersion := m.wizard.kubeUpgradeVersions[m.wizard.kubeUpgradeSelectedIdx]
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Initiating upgrade..."
			return m, m.upgradeKubeCluster(m.wizard.kubeUpgradeClusterId, selectedVersion)
		}
		return m, nil
	}

	return m, nil
}

// handleKubePolicyEditKeyPress handles keyboard input for the policy edit view
func (m Model) handleKubePolicyEditKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	policies := []string{"ALWAYS_UPDATE", "MINIMAL_DOWNTIME", "NEVER_UPDATE"}

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.mode = DetailView
		return m, nil

	case "up", "k":
		if m.wizard.kubePolicySelectedIdx > 0 {
			m.wizard.kubePolicySelectedIdx--
		}
		return m, nil

	case "down", "j":
		if m.wizard.kubePolicySelectedIdx < len(policies)-1 {
			m.wizard.kubePolicySelectedIdx++
		}
		return m, nil

	case "enter":
		selectedPolicy := policies[m.wizard.kubePolicySelectedIdx]
		return m, m.updateKubePolicy(m.wizard.kubePolicyClusterId, selectedPolicy)
	}

	return m, nil
}

// handleKubeDeleteConfirmKeyPress handles keyboard input for the delete confirmation view
func (m Model) handleKubeDeleteConfirmKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.mode = DetailView
		m.wizard.kubeDeleteConfirmInput = ""
		return m, nil

	case "enter":
		// Check if the input matches the cluster name
		if m.wizard.kubeDeleteConfirmInput == m.wizard.kubeDeleteClusterName {
			return m, m.deleteKubeCluster(m.wizard.kubeDeleteClusterId)
		}
		// Input doesn't match - show error notification
		m.notification = "❌ Cluster name does not match"
		m.notificationExpiry = time.Now().Add(3 * time.Second)
		return m, nil

	case "backspace":
		if len(m.wizard.kubeDeleteConfirmInput) > 0 {
			m.wizard.kubeDeleteConfirmInput = m.wizard.kubeDeleteConfirmInput[:len(m.wizard.kubeDeleteConfirmInput)-1]
		}
		return m, nil

	default:
		// Handle regular character input
		char := msg.String()
		if len(char) == 1 {
			m.wizard.kubeDeleteConfirmInput += char
		}
		return m, nil
	}
}

// handleNodePoolScaleKeyPress handles keyboard input for the node pool scale view
func (m Model) handleNodePoolScaleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = NodePoolDetailView
		m.wizard.nodePoolScaleFieldIdx = 0
		return m, nil

	case "up", "k":
		if m.wizard.nodePoolScaleFieldIdx > 0 {
			m.wizard.nodePoolScaleFieldIdx--
		}
		return m, nil

	case "down", "j":
		if m.wizard.nodePoolScaleFieldIdx < 5 {
			m.wizard.nodePoolScaleFieldIdx++
		}
		return m, nil

	case "+", "=":
		switch m.wizard.nodePoolScaleFieldIdx {
		case 0: // Desired nodes
			if m.wizard.nodePoolScaleDesired < m.wizard.nodePoolScaleMax {
				m.wizard.nodePoolScaleDesired++
			}
		case 1: // Min nodes
			if m.wizard.nodePoolScaleMin < m.wizard.nodePoolScaleMax {
				m.wizard.nodePoolScaleMin++
			}
		case 2: // Max nodes
			if m.wizard.nodePoolScaleMax < 100 {
				m.wizard.nodePoolScaleMax++
			}
		}
		return m, nil

	case "-":
		switch m.wizard.nodePoolScaleFieldIdx {
		case 0: // Desired nodes
			if m.wizard.nodePoolScaleDesired > m.wizard.nodePoolScaleMin {
				m.wizard.nodePoolScaleDesired--
			}
		case 1: // Min nodes
			if m.wizard.nodePoolScaleMin > 0 {
				m.wizard.nodePoolScaleMin--
			}
		case 2: // Max nodes
			if m.wizard.nodePoolScaleMax > m.wizard.nodePoolScaleMin && m.wizard.nodePoolScaleMax > m.wizard.nodePoolScaleDesired {
				m.wizard.nodePoolScaleMax--
			}
		}
		return m, nil

	case " ":
		// Toggle autoscale on field 3
		if m.wizard.nodePoolScaleFieldIdx == 3 {
			m.wizard.nodePoolScaleAutoscale = !m.wizard.nodePoolScaleAutoscale
		}
		return m, nil

	case "enter":
		switch m.wizard.nodePoolScaleFieldIdx {
		case 4: // Cancel button
			m.mode = NodePoolDetailView
			m.wizard.nodePoolScaleFieldIdx = 0
			return m, nil
		case 5: // Apply button
			if m.detailData != nil && m.selectedNodePool != nil {
				kubeID := getString(m.detailData, "id")
				nodePoolID := getString(m.selectedNodePool, "id")
				return m, m.scaleNodePool(
					kubeID,
					nodePoolID,
					m.wizard.nodePoolScaleDesired,
					m.wizard.nodePoolScaleMin,
					m.wizard.nodePoolScaleMax,
					m.wizard.nodePoolScaleAutoscale,
				)
			}
			return m, nil
		}
		return m, nil
	}
	return m, nil
}

// handleNodePoolDeleteConfirmKeyPress handles keyboard input for the node pool delete confirmation view
func (m Model) handleNodePoolDeleteConfirmKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.mode = NodePoolDetailView
		m.wizard.nodePoolDeleteConfirmInput = ""
		return m, nil

	case "backspace":
		if len(m.wizard.nodePoolDeleteConfirmInput) > 0 {
			m.wizard.nodePoolDeleteConfirmInput = m.wizard.nodePoolDeleteConfirmInput[:len(m.wizard.nodePoolDeleteConfirmInput)-1]
		}
		return m, nil

	case "enter":
		// Check if the input matches the node pool name
		if m.selectedNodePool != nil {
			poolName := getString(m.selectedNodePool, "name")
			if m.wizard.nodePoolDeleteConfirmInput == poolName {
				kubeID := getString(m.detailData, "id")
				nodePoolID := getString(m.selectedNodePool, "id")
				return m, m.deleteNodePool(kubeID, nodePoolID)
			}
		}
		return m, nil

	default:
		// Handle regular character input
		char := msg.String()
		if len(char) == 1 {
			m.wizard.nodePoolDeleteConfirmInput += char
		}
		return m, nil
	}
}
