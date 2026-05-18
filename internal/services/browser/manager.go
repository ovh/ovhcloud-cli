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
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/ovh/ovhcloud-cli/internal/config"
	"github.com/ovh/ovhcloud-cli/internal/flags"
	httpLib "github.com/ovh/ovhcloud-cli/internal/http"
	block_storage "github.com/ovh/ovhcloud-cli/internal/services/browser/views/block_storage"
	file_storage "github.com/ovh/ovhcloud-cli/internal/services/browser/views/file_storage"
	object_storage "github.com/ovh/ovhcloud-cli/internal/services/browser/views/object_storage"
	"github.com/ovh/ovhcloud-cli/internal/services/browser/views"
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
	KubeKubeconfigPickerView  // Directory picker for saving kubeconfig
	ComingSoonView            // Coming soon placeholder for unimplemented products
	S3CredentialsView         // S3 user credentials display after creation
	LBPoolDetailView          // Detail view for a single LB pool
	LBListenerDetailView      // Detail view for a single LB listener
	LBL7PolicyDetailView      // Detail view for a single L7 policy
	LBL7RulesView             // List view for L7 rules of a policy
	LBPoolMembersView         // List view for members of a pool
	LBHealthMonitorView       // Detail/edit view for a pool's health monitor
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
	// Volume (Block Storage) wizard steps (offset by 300)
	VolumeWizardStepName WizardStep = iota + 300
	VolumeWizardStepRegion
	VolumeWizardStepType
	VolumeWizardStepAvailabilityZone
	VolumeWizardStepSize
	VolumeWizardStepEncryption
	VolumeWizardStepConfirm
)

const (
	// File Storage wizard steps (offset by 400)
	FileWizardStepName    WizardStep = iota + 400
	FileWizardStepRegion
	FileWizardStepType
	FileWizardStepSize
	FileWizardStepNetwork
	FileWizardStepConfirm

	ObjectWizardStepName        WizardStep = iota + 500
	ObjectWizardStepType
	ObjectWizardStepRegion
	ObjectWizardStepReplication
	ObjectWizardStepVersioning
	ObjectWizardStepObjectLock
	ObjectWizardStepUser
	ObjectWizardStepEncryption
	ObjectWizardStepConfirm
	ObjectWizardStepSwiftType
	ObjectWizardStepSwiftRegion
)

const (
	// S3 User wizard steps
	S3UserWizardStepDescription WizardStep = iota + 600
	S3UserWizardStepConfirm
)

const (
	// Volume Backup / Snapshot wizard steps (offset by 700)
	BackupWizardStepVolume  WizardStep = iota + 700 // pick source volume
	BackupWizardStepType                            // pick Snapshot or Backup
	BackupWizardStepName                            // enter name
	BackupWizardStepConfirm                         // confirm
)

const (
	// Private Network wizard steps (offset by 800)
	PrivNetWizardStepRegion  WizardStep = iota + 800 // choose location
	PrivNetWizardStepName                            // network name
	PrivNetWizardStepVlanID                          // VLAN ID (layer 2 option)
	PrivNetWizardStepSubnet                          // configure subnet CIDR
	PrivNetWizardStepDHCP                            // DHCP distribution options
	PrivNetWizardStepAllocPool                       // IP allocation pool (start/end)
	PrivNetWizardStepGateway                         // gateway options
	PrivNetWizardStepConfirm                         // confirm
)

const (
	GwWizardStepRegion  WizardStep = iota + 900 // select region
	GwWizardStepModel                           // select model/size
	GwWizardStepName                            // enter name
	GwWizardStepNetwork                         // select private network
	GwWizardStepConfirm                         // confirm + create
)

const (
	// Load Balancer wizard steps (offset by 1000)
	LBWizardStepName   WizardStep = iota + 1000 // enter name
	LBWizardStepRegion                          // select region
	LBWizardStepFlavor                          // select size/flavor
	LBWizardStepNetwork                         // select private network (optional)
	LBWizardStepConfirm                         // confirm + create
)

const (
	// Floating IP wizard steps (offset by 1100)
	FIPWizardStepRegion   WizardStep = iota + 1100 // select region
	FIPWizardStepInstance                          // select instance (optional)
	FIPWizardStepConfirm                           // confirm + create
)

const (
	// Workflow wizard steps (offset by 1200)
	WorkflowWizardStepType     WizardStep = iota + 1200 // select workflow type
	WorkflowWizardStepInstance                          // select instance
	WorkflowWizardStepName                              // enter name
	WorkflowWizardStepSchedule                          // define schedule/rotation
	WorkflowWizardStepConfirm                           // confirm + create
)

const (
	// LB Pool wizard steps (offset by 1300)
	LBPoolWizardStepName    WizardStep = iota + 1300 // enter pool name
	LBPoolWizardStepAlgo                             // select algorithm
	LBPoolWizardStepProto                            // select protocol
	LBPoolWizardStepSession                          // session persistence
	LBPoolWizardStepConfirm                          // confirm + create
)

const (
	// LB Listener wizard steps (offset by 1400)
	LBListenerWizardStepName    WizardStep = iota + 1400 // enter listener name
	LBListenerWizardStepProto                            // select protocol
	LBListenerWizardStepPort                             // enter port
	LBListenerWizardStepPool                             // select default pool (optional)
	LBListenerWizardStepConfirm                          // confirm + create
)

const (
	// L7 Policy wizard steps (offset by 1500)
	LBL7PolicyWizardStepName         WizardStep = iota + 1500 // enter policy name
	LBL7PolicyWizardStepPosition                              // enter position
	LBL7PolicyWizardStepAction                                // select action
	LBL7PolicyWizardStepRedirectPool                          // select redirect pool (only for redirectToPool)
	LBL7PolicyWizardStepRedirectUrl                           // enter redirect URL (redirectToUrl / redirectPrefix)
	LBL7PolicyWizardStepConfirm                               // confirm + create
)

const (
	// L7 Rule wizard steps (offset by 1600)
	LBL7RuleWizardStepType       WizardStep = iota + 1600 // select rule type
	LBL7RuleWizardStepCompare                             // select comparison type
	LBL7RuleWizardStepKey                                 // enter key (optional)
	LBL7RuleWizardStepValue                               // enter value
	LBL7RuleWizardStepInvert                              // toggle invert
	LBL7RuleWizardStepConfirm                             // confirm + create
)

const (
	// LB Member wizard steps (offset by 1700)
	LBMemberWizardStepName    WizardStep = iota + 1700 // enter member name
	LBMemberWizardStepIP                              // enter IP address
	LBMemberWizardStepPort                            // enter protocol port
	LBMemberWizardStepWeight                          // enter weight
	LBMemberWizardStepConfirm                         // confirm + save
)

const (
	// LB Health Monitor wizard steps (offset by 1800)
	LBHMWizardStepName          WizardStep = iota + 1800 // enter name
	LBHMWizardStepType                                   // select monitor type
	LBHMWizardStepHttpMethod                             // select HTTP method (http/https only)
	LBHMWizardStepUrlPath                                // enter URL path (http/https only)
	LBHMWizardStepExpectedCodes                          // enter expected HTTP codes (http/https only)
	LBHMWizardStepDelay                                  // enter delay (seconds)
	LBHMWizardStepMaxRetries                             // enter max retries
	LBHMWizardStepMaxRetriesDown                         // enter max retries down
	LBHMWizardStepTimeout                                // enter timeout
	LBHMWizardStepConfirm                                // confirm + save
)

const (
	// Managed Database wizard steps (offset by 1900)
	DBWizardStepName    WizardStep = iota + 1900 // enter service name
	DBWizardStepEngine                           // select engine
	DBWizardStepVersion                          // select version
	DBWizardStepRegion                           // select datacenter/region
	DBWizardStepPlan                             // select plan
	DBWizardStepFlavor                           // select instance flavor
	DBWizardStepNodes                            // number of nodes
	DBWizardStepStorage                          // storage size
	DBWizardStepNetwork                          // network type
	DBWizardStepConfirm                          // confirm + create
)

const (
	// Managed Analytics wizard steps (offset by 2000)
	AnalyticsWizardStepName    WizardStep = iota + 2000 // enter service name
	AnalyticsWizardStepEngine                           // select engine
	AnalyticsWizardStepVersion                          // select version
	AnalyticsWizardStepRegion                           // select datacenter/region
	AnalyticsWizardStepPlan                             // select plan
	AnalyticsWizardStepFlavor                           // select instance flavor
	AnalyticsWizardStepNodes                            // number of nodes
	AnalyticsWizardStepStorage                          // storage size
	AnalyticsWizardStepNetwork                          // network type
	AnalyticsWizardStepConfirm                          // confirm + create
)

// ProductType represents a product category
type ProductType int

const (
	ProductInstances ProductType = iota
	ProductKubernetes
	ProductManagedDatabases
	ProductManagedAnalytics
	ProductStorage         // "Stockage" top-level nav
	ProductStorageBlock    // Block Storage (sous-nav)
	ProductStorageFile     // File Storage (sous-nav)
	ProductStorageBackup   // Volume Backup (sous-nav)
	ProductStorageSnapshot // Volume Snapshot (sous-nav)
	ProductStorageObject   // Object Storage (sous-nav)
	ProductStorageArchive  // Cloud Archive (sous-nav)
	ProductNetworks
	ProductNetworkPrivate // Private Networks (sub-nav)
	ProductNetworkPublic  // Public IPs (sub-nav)
	ProductNetworkGateway // Gateways (sub-nav)
	ProductNetworkLB      // Load Balancers (sub-nav)
	ProductProjects
	ProductCompute         // Compute top-level nav
	ProductInstanceBackup  // Instance Backup (compute sub-nav)
	ProductWorkflow        // Workflow (compute sub-nav)
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
	kubeRegions             []string                 // Available regions for K8s
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
	// Kubeconfig file picker fields
	kubeKubeconfigClusterId   string   // Cluster ID for kubeconfig download
	kubeKubeconfigCurrentDir  string   // Currently browsed directory
	kubeKubeconfigEntries     []string // Subdirectory names in current dir
	kubeKubeconfigSelectedIdx int      // 0="..", 1="[Save here]", 2+= entries
	// Volume (Block Storage) wizard fields
	volumeTypes             []string            // Available volume types for the selected region
	volumeRegionTypeMap     map[string][]string // region name -> []type names (pre-loaded)
	volumeAvailabilityZones []string            // Available availability zones for the region
	volumeTypeAZMap         map[string][]string      // type name -> available AZs (for selected region)
	volumeRegionTypeAZMap   map[string]map[string][]string // region -> type -> AZs (pre-loaded)
	volumeName              string   // Volume name input
	volumeNameInput         string   // Input buffer for volume name
	volumeSize              int      // Volume size in GB
	volumeSizeInput         string   // Input buffer for volume size
	volumeType              string   // Selected volume type
	volumeAvailabilityZone  string   // Selected availability zone
	volumeEncryptionIdx     int      // 0=none, 1=OVHcloud Managed Key
	volumeConfirmBtnIdx     int      // 0 = Create, 1 = Cancel
	// File Storage wizard fields
	fileShareName          string
	fileShareNameInput     string
	fileShareSize          int
	fileShareSizeInput     string
	fileShareType          string   // selected type (e.g., "standard-1az")
	fileShareTypeIdx       int
	fileShareRegions       []string
	fileShareNetworks      []map[string]interface{}
	fileShareSubnets       []map[string]interface{}
	fileShareNetworkId     string
	fileShareNetworkName   string
	fileShareSubnetId      string
	fileShareSubnetCIDR    string
	fileShareNetworkMenuIdx int   // 0=network list, 1=subnet list
	fileShareConfirmBtnIdx  int   // 0=Create, 1=Cancel
	// Object Storage wizard fields
	objectName          string   // Container name
	objectNameInput     string
	objectTypeIdx       int      // 0=Standard, 1=High Performance
	objectRegions       []string // Regions supporting S3
	objectUsers         []map[string]interface{} // Cloud users
	objectUserIdx       int
	objectReplication   bool   // Offsite replication enabled
	objectVersioning    bool   // Versioning enabled
	objectLock          bool   // Object Lock enabled
	objectEncryption    bool   // Encryption enabled (AES256)
	objectConfirmBtnIdx int    // 0=Create, 1=Cancel
	objectSwiftTypeIdx  int    // 0=Static, 1=Private, 2=Public
	objectSwiftRegions  []string // Available regions for Swift
	objectSwiftRegion   string // Selected Swift region
	// S3 User wizard fields
	s3UserDescInput     string // Description input buffer
	s3UserDesc          string // Confirmed description
	s3UserConfirmBtnIdx int    // 0=Create, 1=Cancel
	// Volume Backup / Snapshot wizard fields
	backupVolumes      []map[string]interface{} // loaded block storage volumes
	backupVolumeIdx    int                      // selected volume index
	backupTypeIdx      int                      // 0=Snapshot, 1=Backup
	backupName         string                   // confirmed name
	backupNameInput    string                   // input buffer for name
	backupConfirmBtnIdx int                      // 0=Create, 1=Cancel
	// Private Network wizard fields
	privNetRegions       []map[string]interface{} // [{name, type}]
	privNetRegionIdx     int                      // selected region index
	privNetNameInput     string                   // network name input
	privNetName          string                   // confirmed name
	privNetDefineVlan    bool                     // whether user wants to set a VLAN ID
	privNetVlanInput     string                   // VLAN ID input ("" = auto)
	privNetVlanID        int                      // confirmed VLAN ID (0 = auto)
	privNetEnableSubnet  bool                     // whether to configure a subnet
	privNetCIDRInput     string                   // subnet CIDR input
	privNetCIDR          string                   // confirmed CIDR
	privNetEnableDHCP    bool                     // DHCP distribution enabled
	privNetDHCPFieldIdx  int                      // 0=toggle, 1=Next/Back
	privNetAllocStart    string                   // allocation pool start IP
	privNetAllocEnd      string                   // allocation pool end IP
	privNetAllocField    int                      // 0=start, 1=end
	privNetGatewayMode   int                      // 0=announce first CIDR IP, 1=assign explicit IP
	privNetGatewayInput  string                   // gateway IP input (mode 1)
	privNetGateway       string                   // confirmed gateway IP (mode 1)
	privNetConfirmBtnIdx int                      // 0=Create, 1=Cancel
	privNetIsLocalZone   bool                     // true when selected region is a local zone
	privNetUsedVlanIDs      map[int]bool             // VLAN IDs already in use (to validate before API call)
	privNetAddSubnetMode    bool                     // true when adding subnet to an existing network
	privNetTargetNetworkID  string                   // network ID to add subnet to (add-subnet mode)
	privNetSubnettedRegions map[string]bool          // regions that already have a subnet (add-subnet mode)

	// Gateway wizard fields
	gwNetworkID          string
	gwNetworkName        string
	gwRegion             string
	gwSubnetID           string
	gwModelIdx           int    // index into gatewayModels slice
	gwNameInput          string
	gwName               string
	gwConfirmBtnIdx      int    // 0=Create, 1=Cancel
	gwAvailableRegions   []string                   // regions fetched from API
	gwRegionIdx          int                        // selected region index
	gwAvailableNetworks  []map[string]interface{}   // networks in selected region
	gwNetworkIdx         int                        // selected network index
	// Attach mode: populated when launched from private network detail
	// maps region name -> {"openstackId": "...", "subnetId": "..."}
	gwNetworkRegionMap   map[string]map[string]string
	gwAttachMode         bool // true when wizard was launched from private network detail view

	// Load Balancer wizard fields
	lbName              string
	lbNameInput         string
	lbRegion            string
	lbRegionIdx         int
	lbAvailableRegions  []string
	lbFlavors           []map[string]interface{}
	lbFlavorIdx         int
	lbFlavorId          string
	lbFlavorName        string
	lbNetworks          []map[string]interface{}
	lbNetworkIdx        int    // 0 = Aucun réseau, 1+ = index into lbNetworks
	lbNetworkId         string
	lbNetworkName       string
	lbSubnetId          string
	lbConfirmBtnIdx     int

	// LB Pool wizard fields
	lbPoolLBId          string
	lbPoolLBName        string
	lbPoolLBRegion      string
	lbPoolNameInput     string
	lbPoolName          string
	lbPoolAlgoIdx       int
	lbPoolAlgo          string
	lbPoolProtoIdx      int
	lbPoolProto         string
	lbPoolSessionIdx    int    // 0=None, 1=Source IP
	lbPoolSession       string // "" or "SOURCE_IP"
	lbPoolConfirmIdx    int
	lbPoolEditPoolId    string // non-empty = edit mode (pool ID being edited)

	// LB Listener wizard fields
	lbListenerLBId        string
	lbListenerLBName      string
	lbListenerLBRegion    string
	lbListenerNameInput   string
	lbListenerName        string
	lbListenerProtoIdx    int
	lbListenerProto       string
	lbListenerPortInput   string
	lbListenerPort        int
	lbListenerPoolIdx     int // 0 = no pool, 1+ index into lbPools for this LB
	lbListenerPoolId      string
	lbListenerConfirmIdx  int
	lbListenerEditId      string // non-empty = edit mode (listener ID being edited)

	// L7 Policy wizard fields
	l7PolicyListenerId       string
	l7PolicyListenerName     string
	l7PolicyLBRegion         string
	l7PolicyLBId             string
	l7PolicyNameInput        string
	l7PolicyName             string
	l7PolicyPositionInput    string
	l7PolicyPosition         int
	l7PolicyActionIdx        int
	l7PolicyAction           string
	l7PolicyRedirectPoolIdx  int
	l7PolicyRedirectPoolId   string
	l7PolicyRedirectUrlInput string
	l7PolicyRedirectUrl      string
	l7PolicyConfirmIdx       int
	l7PolicyEditId           string // non-empty = edit mode (policy ID being edited)

	// L7 Rule wizard fields
	l7RulePolicyId    string // policy ID the rule belongs to
	l7RulePolicyName  string // policy name (display)
	l7RuleLBRegion    string // region
	l7RuleTypeIdx     int    // selected index in lbL7RuleTypeOptions
	l7RuleType        string // e.g. "HEADER", "PATH", "HOST_NAME", ...
	l7RuleCompareIdx  int    // selected index in compare options for this type
	l7RuleCompare     string // e.g. "EQUAL_TO", "STARTS_WITH", "REGEX", ...
	l7RuleKeyInput    string // raw input for key field
	l7RuleKey         string // key (for HEADER / COOKIE type)
	l7RuleValueInput  string // raw input for value
	l7RuleValue       string // value to compare against
	l7RuleInvert      bool   // whether to invert the rule
	l7RuleConfirmIdx  int    // 0=Confirm, 1=Cancel
	l7RuleEditId      string // non-empty = edit mode (rule ID being edited)

	// LB Member wizard fields
	lbMemberPoolId      string // pool ID the member belongs to
	lbMemberPoolRegion  string // region
	lbMemberEditId      string // non-empty = edit mode (member ID being edited)
	lbMemberNameInput   string // raw input for member name
	lbMemberName        string // confirmed member name
	lbMemberIPInput     string // raw input for IP address
	lbMemberIP          string // confirmed IP address
	lbMemberPortInput   string // raw input for protocol port
	lbMemberPort        int    // confirmed protocol port
	lbMemberWeightInput string // raw input for weight
	lbMemberWeight      int    // confirmed weight (0-256)
	lbMemberConfirmIdx  int    // 0=Save, 1=Cancel

	// LB Health Monitor wizard fields
	lbHMPoolId             string // pool ID the monitor belongs to
	lbHMPoolRegion         string // region
	lbHMEditId             string // non-empty = edit mode (HM ID being edited)
	lbHMNameInput          string // raw input for name
	lbHMName               string // confirmed name
	lbHMTypeIdx            int    // selected index in lbHMTypeOptions
	lbHMType               string // e.g. "http", "tcp", ...
	lbHMDelayInput         string // raw input for delay
	lbHMDelay              int    // confirmed delay (seconds)
	lbHMMaxRetriesInput    string // raw input for max retries
	lbHMMaxRetries         int    // confirmed max retries
	lbHMMaxRetriesDownInput string // raw input for max retries down
	lbHMMaxRetriesDown     int    // confirmed max retries down
	lbHMTimeoutInput        string // raw input for timeout
	lbHMTimeout             int    // confirmed timeout (seconds)
	lbHMHttpMethodIdx       int    // selected index in lbHMHttpMethodOptions
	lbHMHttpMethod          string // e.g. "GET", "HEAD"
	lbHMUrlPathInput        string // raw input for URL path (http/https)
	lbHMUrlPath             string // confirmed URL path
	lbHMExpectedCodesInput  string // raw input for expected HTTP codes, e.g. "200"
	lbHMExpectedCodes       string // confirmed expected codes
	lbHMConfirmIdx          int    // 0=Save, 1=Cancel

	// Floating IP wizard fields
	fipRegion            string
	fipRegionIdx         int
	fipAvailableRegions  []string
	fipInstances         []map[string]interface{}
	fipInstanceIdx       int   // 0 = standalone (no instance), 1+ = index into fipInstances
	fipInstanceId        string
	fipInstanceName      string
	fipConfirmBtnIdx     int

	// Workflow wizard fields
	wfInstances      []map[string]interface{}
	wfInstanceIdx    int
	wfInstanceId     string
	wfInstanceName   string
	wfRegion         string
	wfName           string
	wfNameInput      string
	wfScheduleIdx    int    // 0=rotation7, 1=rotation14, 2=custom
	wfCron           string
	wfCronInput      string
	wfRotation       int
	wfConfirmBtnIdx  int

	// Managed Database wizard fields
	dbNameInput    string
	dbName         string
	dbEngines      []map[string]interface{} // from capabilities
	dbEngineIdx    int
	dbEngine          string
	dbEngineCategory  string // "operational" or "analysis"
	dbVersionIdx   int
	dbVersion      string
	dbRegionIdx    int
	dbRegion       string
	dbPlanIdx      int
	dbPlan         string
	dbFlavors      []map[string]interface{} // from capabilities
	dbFlavorIdx    int
	dbFlavor       string
	dbCapPlans     []string                 // fallback plan names from capabilities
	dbNodesInput   string
	dbNodes        int
	dbStorageInput string
	dbDiskSize     int
	dbNetworkIdx   int    // 0=public 1=private
	dbNetworkId    string
	dbAvailItems   []map[string]interface{} // from /database/availability
	dbCapsRegions  []string                 // fallback regions from capabilities
	dbConfirmIdx   int                      // 0=Create 1=Cancel
}

// Model represents the TUI application state
type Model struct {
	width              int
	height             int
	mode               ViewMode
	previousMode       ViewMode // Previous mode to return to from debug view
	currentProduct     ProductType
	navIdx             int  // Index in navigation bar
	storageSubIdx      int  // Index in storage sub-navigation (0=Prise en main, 1=Block Storage, ...)
	inStorageSubNav    bool // Whether the keyboard focus is in the storage sub-nav bar
	networkSubIdx      int  // Index in network sub-navigation
	inNetworkSubNav    bool // Whether the keyboard focus is in the network sub-nav bar
	computeSubIdx      int  // Index in compute sub-navigation
	inComputeSubNav    bool // Whether the keyboard focus is in the compute sub-nav bar
	inTableFocus       bool // Whether the keyboard focus is in the table content (third navigation level)
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
	// Object Storage tab rendering
	renderObjectStorageTabs bool
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
	// LB pools cache (lbId -> pools)
	lbPools               map[string][]map[string]interface{}
	selectedLBPool        map[string]interface{}             // Currently selected pool for detail view
	lbPoolDetailActionIdx int                               // Selected action in pool detail view (0=Edit, 1=Delete, 2=Members)
	lbPoolDetailConfirm   bool                              // Whether we're in confirm mode in pool detail
	lbPoolListIdx         int                               // Highlighted pool row in LB detail (-1 = none)
	// LB pool members cache (poolId -> members)
	lbPoolMembers         map[string][]map[string]interface{} // poolID → members
	lbPoolMemberDetailIdx int                               // Currently displayed member index
	lbPoolMemberConfirm   bool                              // Confirm mode for member deletion
	lbMembersSection      int                               // 0=actions bar, 1=member pagination
	lbMembersActionIdx    int                               // Selected action: 0=Create,1=Edit,2=Delete,3=HealthMonitor
	// LB health monitors cache (poolId -> health monitor)
	lbHealthMonitors      map[string]map[string]interface{}  // poolID → health monitor (one per pool)
	lbHMConfirm           bool                              // Confirm mode for HM deletion
	lbHMActionIdx         int                               // Selected action button in HM view (0=Create or Edit, 1=Delete)
	// LB listeners cache (lbId -> listeners)
	lbListeners       map[string][]map[string]interface{}
	lbListenerListIdx int // Highlighted listener row in LB detail (-1 = none)
	selectedLBListener       map[string]interface{} // Currently selected listener for detail view
	lbListenerDetailActionIdx int                   // Selected action in listener detail view (0=Edit, 1=Delete)
	lbListenerDetailConfirm   bool                  // Confirm mode in listener detail
	lbDetailSection           int                   // 0=Listeners block focused, 1=Pools block focused
	lbL7Policies              map[string][]map[string]interface{} // key = listenerId
	lbL7Rules                 map[string][]map[string]interface{} // key = policyId
	lbL7PolicyListIdx         int                                 // Highlighted policy row in listener detail (-1 = none)
	selectedLBL7Policy        map[string]interface{}              // Currently selected policy for detail view
	lbL7PolicyDetailActionIdx int                                 // Selected action in policy detail (0=Edit, 1=Delete)
	lbL7PolicyDetailConfirm   bool                                // Confirm mode in policy detail
	lbL7RuleDetailIdx         int                                 // Currently displayed rule index in L7 Rules view
	lbL7RuleActionIdx         int                                 // Selected action button in L7 Rules view (0=Create,1=Edit,2=Delete)
	lbL7RuleConfirm           bool                                // Confirm mode for rule deletion
	// Background detail-view refresh (set by auto-refresh timer, cleared by data handlers)
	detailRefreshId   string
	detailRefreshName string
	// Block Storage detail view
	volumeDetailView *block_storage.DetailView
	// Snapshot detail view
	snapshotDetailView *block_storage.SnapshotDetailView
	// Backup detail view
	backupDetailView *block_storage.BackupDetailView
	// File Storage detail view
	fileShareDetailView *file_storage.DetailView
	// Object Storage detail view
	objectDetailView *object_storage.DetailView
	// Object Storage user detail view
	objectUserDetailView *object_storage.UserDetailView
	// Object Storage tabs (0=Containers, 1=Users)
	objectStorageTabIdx int
	objectStorageUsers  []map[string]interface{}
	// Managed DB/Analytics detail sub-resources
	dbDetailUsers     []map[string]interface{}
	dbDetailBackups   []map[string]interface{}
	dbDetailDatabases []map[string]interface{}
	dbDetailPools     []map[string]interface{}
	dbDetailLoaded    bool // true once fetchDBDetailSubresources has returned
	dbDetailTab       int  // 0=Service, 1=Users, 2=Backups, 3=Databases, 4=Pools
	// DB user creation state (Users tab)
	dbUserCreateMode  bool                   // true when typing a new username
	dbUserCreateInput string                 // username being typed
	dbUserCreatedData map[string]interface{} // creation result (has password + endpoints)
	// Private Networks tabs (0=Régions vRack, 1=Local Zones)
	privNetTabIdx           int
	privNetLocalZones       []map[string]interface{}
	privNetSelectedSubnet   int  // index of subnet selected for deletion in detail view
	privNetSelectedRegion   int  // index of region selected for deletion in detail view
	// Public IPs tabs (0=Floating IPs, 1=Additional IPs)
	publicIPTabIdx     int
	additionalIPsData  []map[string]interface{}
	// S3 user creation result (for credentials display)
	s3CreatedUser        map[string]interface{}
	s3CreatedCredentials map[string]interface{}
	s3CredentialsSavedPath  string
	s3CredentialsSaveError  string
	s3PendingEnableUser  map[string]interface{} // user being enabled (for credentials display)
	s3CredentialsFromEnable bool // true if S3CredentialsView opened from enable action
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
	data          []map[string]interface{}
	err           error
	forProduct    ProductType // The product that requested this data
	s3Users       []map[string]interface{} // S3 users (for Object Storage)
	additionalIPs []map[string]interface{} // Failover IPs (for ProductNetworkPublic tab 1)
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
	backupName string
	err        error
}

// sshConnectionMsg is returned when SSH action is requested
type sshConnectionMsg struct {
	ip   string
	user string
}

// Kubernetes wizard messages
type kubeRegionsLoadedMsg struct {
	regions []string
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

type volumeRegionsLoadedMsg struct {
	regionNames       []string
	regionTypeMap     map[string][]string
	regionTypeAZMap   map[string]map[string][]string // region -> type -> []AZs
	err               error
}

type volumeTypesLoadedMsg struct {
	types     []string
	typeAZMap map[string][]string // type -> []AZs
	err       error
}

type volumeAZLoadedMsg struct {
	availabilityZones []string
	err               error
}

type volumeCreatedMsg struct {
	volume map[string]interface{}
	err    error
}

type volumeActionDoneMsg struct {
	action int
	err    error
}

type refreshBlockStorageMsg struct{}

type fileShareRegionsLoadedMsg struct {
	regions []string
	err     error
}

type fileShareNetworksLoadedMsg struct {
	networks []map[string]interface{}
	err      error
}

type fileShareSubnetsLoadedMsg struct {
	subnets []map[string]interface{}
	err     error
}

type fileShareCreatedMsg struct {
	share map[string]interface{}
	err   error
}

type objectStorageInitDataLoadedMsg struct {
        regions []string
        users   []map[string]interface{}
        err     error
}

type objectContainerCreatedMsg struct {
        container map[string]interface{}
        err       error
}

type objectContainerActionDoneMsg struct {
	action int
	err    error
}

type swiftContainerUpdatedMsg struct {
	containerName string
	newType       string
	err           error
}

type containerPolicyAddedMsg struct {
	containerName string
	roleName      string
	err           error
}

type s3SecretLoadedMsg struct {
	secret string
	err    error
}

type s3UserActionDoneMsg struct {
	action        int
	newCredential map[string]interface{}
	err           error
}

type s3UserCreatedMsg struct {
	user        map[string]interface{}
	credentials map[string]interface{}
	err         error
}

type s3CredentialsSavedMsg struct {
	filePath    string
	profileName string
	err         error
}

type swiftRegionsLoadedMsg struct {
	regions []string
}

type privNetRegionsLoadedMsg struct {
	regions []map[string]interface{}
	err     error
}

type privNetCreatedMsg struct {
	network map[string]interface{}
	err     error
}

type subnetAddedMsg struct {
	networkID string
	err       error
}

type subnetDeletedMsg struct {
	networkID string
	err       error
}

type regionDeletedMsg struct {
	networkID string
	region    string
	err       error
}

type gatewayDetachedMsg struct {
	networkID string
	err       error
}

type privNetDeletedMsg struct {
	networkName string
	err         error
}

type gwCreatedMsg struct {
	gateway map[string]interface{}
	err     error
}

type gwRegionsLoadedMsg struct {
	regions []string
	err     error
}

type gwNetworksLoadedMsg struct {
	networks []map[string]interface{}
	err      error
}

type gwSubnetLoadedMsg struct {
	subnetID string
	err      error
}

type gwDeletedMsg struct {
	gatewayName string
	err         error
}

type lbCreatedMsg struct {
	lb  map[string]interface{}
	err error
}

type lbRegionsLoadedMsg struct {
	regions []string
	err     error
}

type dbCapabilitiesLoadedMsg struct {
	engines     []map[string]interface{}
	flavors     []map[string]interface{}
	plans       []map[string]interface{}
	availItems  []map[string]interface{}
	capsRegions []string
	err         error
}

type dbCreatedMsg struct {
	dbName string
	err    error
}

type analyticsCreatedMsg struct {
	name string
	err  error
}

type dbDetailSubresourcesMsg struct {
	serviceId string
	users     []map[string]interface{}
	backups   []map[string]interface{}
	databases []map[string]interface{}
	pools     []map[string]interface{}
	err       error
}

type dbServiceDeletedMsg struct {
	serviceId string
	err       error
}

type analyticsEngineAvailLoadedMsg struct {
	engine    string
	availItems []map[string]interface{}
	err       error
}

type lbFlavorsLoadedMsg struct {
	flavors []map[string]interface{}
	err     error
}

type lbNetworksLoadedMsg struct {
	networks []map[string]interface{}
	err      error
}

type lbSubnetLoadedMsg struct {
	subnetID string
	err      error
}

type lbDeletedMsg struct {
	lbName string
	err    error
}

type lbPoolCreatedMsg struct {
	poolName string
	err      error
}

type lbPoolsLoadedMsg struct {
	lbID  string
	pools []map[string]interface{}
	err   error
}

type lbPoolDeletedMsg struct {
	poolName string
	err      error
}

type lbPoolUpdatedMsg struct {
	poolName string
	err      error
}

type lbListenerCreatedMsg struct {
	listenerName string
	err          error
}

type lbListenerDeletedMsg struct {
	listenerName string
	err          error
}

type lbListenerUpdatedMsg struct {
	listenerName string
	err          error
}

type lbListenersLoadedMsg struct {
	lbID      string
	listeners []map[string]interface{}
	err       error
}

type lbL7PolicyCreatedMsg struct {
	policyName string
	err        error
}

type lbL7PoliciesLoadedMsg struct {
	listenerID string
	policies   []map[string]interface{}
	err        error
}

type lbL7RulesLoadedMsg struct {
	policyID string
	rules    []map[string]interface{}
	err      error
}

type lbL7RuleCreatedMsg struct {
	policyID string
	err      error
}

type lbL7RuleDeletedMsg struct {
	policyID string
	err      error
}

type lbL7PolicyDeletedMsg struct {
	policyName string
	err        error
}

type lbL7PolicyUpdatedMsg struct {
	policyName string
	err        error
}

type lbPoolMembersLoadedMsg struct {
	poolID  string
	members []map[string]interface{}
	err     error
}

type lbPoolMemberDeletedMsg struct {
	poolID string
	err    error
}

type lbPoolMemberSavedMsg struct {
	poolID string
	err    error
}

type lbHMLoadedMsg struct {
	poolID string
	hm     map[string]interface{} // nil if not found
	err    error
}

type lbHMSavedMsg struct {
	poolID string
	err    error
}

type lbHMDeletedMsg struct {
	poolID string
	err    error
}

type fipRegionsLoadedMsg struct {
	regions []string
	err     error
}

type fipInstancesLoadedMsg struct {
	instances []map[string]interface{}
	err       error
}

type fipCreatedMsg struct {
	floatingIP map[string]interface{}
	err        error
}

type fipDeletedMsg struct {
	fipIP string
	err   error
}

type workflowDeletedMsg struct {
	name string
	err  error
}

type instanceBackupDeletedMsg struct {
	name string
	err  error
}

type fipDetachedMsg struct {
	fipIP string
	err   error
}

type subnetsLoadedMsg struct {
	networkID string
	subnets   []map[string]any
}

type privNetDetailLoadedMsg struct {
	networkID string
	regions   []interface{}
}

func getNavItems() []NavItem {
	return []NavItem{
		{Label: "Compute", Icon: "💻", Product: ProductCompute, Path: "/instances"},
		{Label: " Kubernetes", Icon: "☸️", Product: ProductKubernetes, Path: "/kubernetes"},
		{Label: " Managed Databases", Icon: "🗄️", Product: ProductManagedDatabases, Path: "/databases"},
		{Label: "Managed Analytics", Icon: "📈", Product: ProductManagedAnalytics, Path: "/analytics"},
		{Label: "Storage", Icon: "💾", Product: ProductStorage, Path: "/storage/block"},
		{Label: "Networks", Icon: "🌐", Product: ProductNetworks, Path: "/networks/private"},
	}
}

type StorageSubItem struct {
	Label   string
	Product ProductType
	Path    string
	Enabled bool
}

func getStorageSubItems() []StorageSubItem {
	return []StorageSubItem{
		{Label: "Block Storage", Product: ProductStorageBlock, Path: "/storage/block", Enabled: true},
		{Label: "File Storage", Product: ProductStorageFile, Path: "/storage/file", Enabled: true},
		{Label: "Volume Backup", Product: ProductStorageBackup, Path: "/storage/backup", Enabled: true},
		{Label: "Volume Snapshot", Product: ProductStorageSnapshot, Path: "/storage/snapshot", Enabled: true},
		{Label: "Object Storage", Product: ProductStorageObject, Path: "/storage/object", Enabled: true},
	}
}

type NetworkSubItem struct {
	Label   string
	Product ProductType
	Path    string
	Enabled bool
}

func getNetworkSubItems() []NetworkSubItem {
	return []NetworkSubItem{
		{Label: "Private Networks", Product: ProductNetworkPrivate, Path: "/networks/private", Enabled: true},
		{Label: "Public IPs", Product: ProductNetworkPublic, Path: "/networks/floatingip", Enabled: true},
		{Label: "Gateways", Product: ProductNetworkGateway, Path: "/networks/gateway", Enabled: true},
		{Label: "Load Balancers", Product: ProductNetworkLB, Path: "/loadbalancer", Enabled: true},
	}
}

type ComputeSubItem struct {
	Label   string
	Product ProductType
	Path    string
	Enabled bool
}

func getComputeSubItems() []ComputeSubItem {
	return []ComputeSubItem{
		{Label: "Instances", Product: ProductInstances, Path: "/instances", Enabled: true},
		{Label: "Instance Backup", Product: ProductInstanceBackup, Path: "/instances/backup", Enabled: true},
		{Label: "Workflow", Product: ProductWorkflow, Path: "/instances/workflow", Enabled: true},
	}
}

// StartBrowser is the entry point for the browser TUI
func StartBrowser(cmd *cobra.Command, args []string) {
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
		if m.currentProduct == ProductInstances {
			if m.mode == TableView {
				return m, tea.Batch(
					m.fetchDataForPath("/instances"),
					m.scheduleRefresh(),
				)
			}
			if m.mode == DetailView && m.detailData != nil {
				m.detailRefreshId = getString(m.detailData, "id")
				m.detailRefreshName = m.currentItemName
				return m, tea.Batch(
					m.fetchDataForPath("/instances"),
					m.scheduleRefresh(),
				)
			}
			// Keep the timer alive even when not in TableView or DetailView
			return m, m.scheduleRefresh()
		}
		// Auto-refresh Kubernetes list if we're viewing Kubernetes in TableView
		if m.currentProduct == ProductKubernetes {
			if m.mode == TableView {
				return m, tea.Batch(
					m.fetchDataForPath("/kubernetes"),
					m.scheduleRefresh(),
				)
			}
			if m.mode == DetailView && m.detailData != nil {
				m.detailRefreshId = getString(m.detailData, "id")
				m.detailRefreshName = m.currentItemName
				return m, tea.Batch(
					m.fetchDataForPath("/kubernetes"),
					m.scheduleRefresh(),
				)
			}
			// Keep the timer alive even when not in TableView or DetailView (e.g. NodePoolsView)
			return m, m.scheduleRefresh()
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
		} else if msg.product == ProductStorageBlock {
			m.mode = WizardView
			m.wizard = WizardData{
				step: VolumeWizardStepName,
			}
			return m, nil
		} else if msg.product == ProductStorageFile {
			m.mode = WizardView
			m.wizard = WizardData{
				step:           FileWizardStepName,
				isLoading:      true,
				loadingMessage: "Loading available regions...",
			}
			return m, m.fetchFileShareRegions()
		} else if msg.product == ProductStorageObject {
			m.mode = WizardView
			m.wizard = WizardData{
				step:           ObjectWizardStepName,
				isLoading:      true,
				loadingMessage: "Loading regions and users...",
			}
			return m, m.fetchObjectStorageInitData()
		} else if msg.product == ProductStorageBackup || msg.product == ProductStorageSnapshot {
			m.mode = WizardView
			m.wizard = WizardData{
				step:           BackupWizardStepVolume,
				isLoading:      true,
				loadingMessage: "Chargement des volumes...",
			}
			return m, m.fetchBackupVolumes()
		} else if msg.product == ProductNetworkPrivate {
			m.mode = WizardView
			// Collect already-used VLAN IDs from the currently loaded networks list
			usedVlans := make(map[int]bool)
			for _, net := range m.currentData {
				switch v := net["vlanId"].(type) {
				case float64:
					usedVlans[int(v)] = true
				case int:
					usedVlans[v] = true
				}
			}
			m.wizard = WizardData{
				step:               PrivNetWizardStepRegion,
				privNetEnableDHCP:  true,
				privNetEnableSubnet: true,
				privNetGatewayMode: 0,
				privNetCIDRInput:   "10.0.0.0/16",
				privNetUsedVlanIDs: usedVlans,
				isLoading:      true,
				loadingMessage: "Loading regions...",
			}
			return m, m.fetchPrivateNetRegionsCmd()
		} else if msg.product == ProductNetworkGateway {
			m.mode = WizardView
			m.wizard = WizardData{
				step:           GwWizardStepRegion,
				isLoading:      true,
				loadingMessage: "Loading regions...",
			}
			return m, m.fetchGwRegions()
		} else if msg.product == ProductManagedDatabases {
			m.mode = WizardView
			m.wizard = WizardData{
				step: DBWizardStepName,
			}
			return m, nil
		} else if msg.product == ProductManagedAnalytics {
			m.mode = WizardView
			m.wizard = WizardData{
				step: AnalyticsWizardStepName,
			}
			return m, nil
		} else if msg.product == ProductNetworkLB {
			m.mode = WizardView
			m.wizard = WizardData{
				step: LBWizardStepName,
			}
			return m, nil
		} else if msg.product == ProductNetworkPublic {
			m.mode = WizardView
			m.wizard = WizardData{
				step:           FIPWizardStepRegion,
				isLoading:      true,
				loadingMessage: "Loading regions...",
			}
			return m, m.fetchFIPRegions()
		} else if msg.product == ProductWorkflow {
			m.mode = WizardView
			m.wizard = WizardData{
				step: WorkflowWizardStepType,
			}
			return m, nil
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

	case kubeconfigReadyForK9sMsg:
		return m.handleKubeconfigReadyForK9s(msg)

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

	case startKubeKubeconfigPickerMsg:
		return m.handleStartKubeKubeconfigPicker(msg)

	// Node pool action messages
	case startNodePoolScaleMsg:
		return m.handleStartNodePoolScale(msg)

	case nodePoolScaleMsg:
		return m.handleNodePoolScaled(msg)

	case startNodePoolDeleteMsg:
		return m.handleStartNodePoolDelete(msg)

	case nodePoolDeletedMsg:
		return m.handleNodePoolDeleted(msg)

	case volumeRegionsLoadedMsg:
		return m.handleVolumeRegionsLoaded(msg)

	case volumeTypesLoadedMsg:
		return m.handleVolumeTypesLoaded(msg)

	case volumeAZLoadedMsg:
		return m.handleVolumeAZLoaded(msg)

	case volumeCreatedMsg:
		return m.handleVolumeCreated(msg)

	case block_storage.ExecuteVolumeActionMsg:
		return m.handleExecuteVolumeAction(msg)

	case backupVolumesLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.backupVolumes = msg.volumes
		m.wizard.backupVolumeIdx = 0
		return m, nil

	case privNetRegionsLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		// In add-subnet mode, filter out regions that already have a subnet
		if m.wizard.privNetAddSubnetMode && len(m.wizard.privNetSubnettedRegions) > 0 {
			var filtered []map[string]interface{}
			for _, r := range msg.regions {
				name, _ := r["name"].(string)
				if !m.wizard.privNetSubnettedRegions[name] {
					filtered = append(filtered, r)
				}
			}
			m.wizard.privNetRegions = filtered
		} else {
			m.wizard.privNetRegions = msg.regions
		}
		m.wizard.privNetRegionIdx = 0
		return m, nil

	case subnetDeletedMsg:
		m.actionConfirm = false
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Failed to delete subnet: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ Subnet deleted successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		// Refresh subnets in the current detail view
		if m.detailData != nil {
			return m, tea.Batch(
				m.fetchNetworkSubnets(msg.networkID),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case regionDeletedMsg:
		m.actionConfirm = false
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Failed to delete region: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Region %s removed from network", msg.region)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		// Reload the private networks list to reflect the change
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/private"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case gatewayDetachedMsg:
		m.actionConfirm = false
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Failed to detach gateway: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ Gateway detached from network"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if m.detailData != nil {
			return m, tea.Batch(
				m.fetchPrivateNetworkDetail(msg.networkID),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case subnetAddedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Failed to add subnet: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = LoadingView
			return m, tea.Batch(
				m.fetchDataForPath("/networks/private"),
				tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		m.notification = "✅ Subnet added successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/private"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case privNetCreatedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(10 * time.Second)
			m.mode = LoadingView
			// Always reload the network list (network may have been created even if subnet failed)
			return m, tea.Batch(
				m.fetchDataForPath("/networks/private"),
				tea.Tick(10*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		netName, _ := msg.network["name"].(string)
		m.notification = fmt.Sprintf("✅ Private network '%s' created successfully", netName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/private"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case privNetDeletedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Private network '%s' deleted successfully", msg.networkName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/private"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case gwDeletedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Gateway '%s' deleted successfully", msg.gatewayName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/gateway"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case gwCreatedMsg:
		attachMode := m.wizard.gwAttachMode
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			if attachMode {
				m.mode = DetailView
			} else {
				m.mode = TableView
			}
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ Gateway created successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if attachMode {
			// Return to private network list
			m.mode = LoadingView
			return m, tea.Batch(
				m.fetchDataForPath("/networks/private"),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/gateway"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case gwRegionsLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.gwAvailableRegions = msg.regions
		m.wizard.gwRegionIdx = 0
		return m, nil

	case gwNetworksLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.gwAvailableNetworks = msg.networks
		m.wizard.gwNetworkIdx = 0
		return m, nil

	case gwSubnetLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.gwSubnetID = msg.subnetID
		m.wizard.step = GwWizardStepConfirm
		return m, nil

	case lbCreatedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		// Prefer name from API response, fall back to wizard input
		lbName := getString(msg.lb, "name")
		if lbName == "" {
			lbName = getString(msg.lb, "id")
		}
		m.notification = fmt.Sprintf("✅ Load Balancer '%s' created successfully", lbName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/loadbalancer"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case subnetsLoadedMsg:
		if m.detailData != nil && getStringValue(m.detailData, "id", "") == msg.networkID {
			m.detailData["_subnets"] = msg.subnets
		}
		return m, nil

	case privNetDetailLoadedMsg:
		if m.detailData != nil && getStringValue(m.detailData, "id", "") == msg.networkID {
			m.detailData["regions"] = msg.regions
		}
		return m, nil

	case fipDeletedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Floating IP %s deleted successfully", msg.fipIP)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/floatingip"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case fipDetachedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Floating IP %s detached successfully", msg.fipIP)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/floatingip"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case lbDeletedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Load Balancer '%s' deleted successfully", msg.lbName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/loadbalancer"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case lbPoolCreatedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = DetailView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Pool \"%s\" created successfully", msg.poolName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.mode = DetailView
		// Refresh pools list
		if m.detailData != nil {
			lbID := getStringValue(m.detailData, "id", "")
			lbRegion := getStringValue(m.detailData, "region", "")
			if lbID != "" && lbRegion != "" {
				return m, tea.Batch(
					m.fetchLBPools(lbID, lbRegion),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbPoolsLoadedMsg:
		if msg.err == nil {
			if m.lbPools == nil {
				m.lbPools = make(map[string][]map[string]interface{})
			}
			m.lbPools[msg.lbID] = msg.pools
		}
		return m, nil

	case lbPoolDeletedMsg:
		m.mode = DetailView
		m.selectedLBPool = nil
		m.lbPoolDetailActionIdx = 0
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Pool \"%s\" deleted successfully", msg.poolName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if m.detailData != nil {
			lbID := getStringValue(m.detailData, "id", "")
			lbRegion := getStringValue(m.detailData, "region", "")
			if lbID != "" && lbRegion != "" {
				return m, tea.Batch(
					m.fetchLBPools(lbID, lbRegion),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbPoolUpdatedMsg:
		m.wizard = WizardData{}
		m.mode = DetailView
		m.selectedLBPool = nil
		m.lbPoolDetailActionIdx = 0
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Pool \"%s\" updated successfully", msg.poolName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if m.detailData != nil {
			lbID := getStringValue(m.detailData, "id", "")
			lbRegion := getStringValue(m.detailData, "region", "")
			if lbID != "" && lbRegion != "" {
				return m, tea.Batch(
					m.fetchLBPools(lbID, lbRegion),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbListenerCreatedMsg:
		m.wizard = WizardData{}
		m.mode = DetailView
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Listener \"%s\" created successfully", msg.listenerName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		// Refresh listeners list
		if m.detailData != nil {
			lbID := getStringValue(m.detailData, "id", "")
			lbRegion := getStringValue(m.detailData, "region", "")
			if lbID != "" && lbRegion != "" {
				return m, tea.Batch(
					m.fetchLBListeners(lbID, lbRegion),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbListenersLoadedMsg:
		if msg.err == nil {
			if m.lbListeners == nil {
				m.lbListeners = make(map[string][]map[string]interface{})
			}
			m.lbListeners[msg.lbID] = msg.listeners
		}
		return m, nil

	case lbListenerDeletedMsg:
		m.mode = DetailView
		m.selectedLBListener = nil
		m.lbListenerDetailActionIdx = 0
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Listener \"%s\" deleted successfully", msg.listenerName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if m.detailData != nil {
			lbID := getStringValue(m.detailData, "id", "")
			lbRegion := getStringValue(m.detailData, "region", "")
			if lbID != "" && lbRegion != "" {
				return m, tea.Batch(
					m.fetchLBListeners(lbID, lbRegion),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbListenerUpdatedMsg:
		m.wizard = WizardData{}
		m.mode = DetailView
		m.selectedLBListener = nil
		m.lbListenerDetailActionIdx = 0
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Listener \"%s\" updated successfully", msg.listenerName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if m.detailData != nil {
			lbID := getStringValue(m.detailData, "id", "")
			lbRegion := getStringValue(m.detailData, "region", "")
			if lbID != "" && lbRegion != "" {
				return m, tea.Batch(
					m.fetchLBListeners(lbID, lbRegion),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbL7PolicyCreatedMsg:
		m.wizard = WizardData{}
		m.mode = LBListenerDetailView
		m.lbL7PolicyListIdx = -1
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ L7 Policy \"%s\" created successfully", msg.policyName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if m.selectedLBListener != nil {
			listenerID := getStringValue(m.selectedLBListener, "id", "")
			region := getStringValue(m.detailData, "region", "")
			if listenerID != "" && region != "" {
				return m, tea.Batch(
					m.fetchLBL7Policies(listenerID, region),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbL7PoliciesLoadedMsg:
		if msg.err == nil {
			if m.lbL7Policies == nil {
				m.lbL7Policies = make(map[string][]map[string]interface{})
			}
			m.lbL7Policies[msg.listenerID] = msg.policies
		}
		return m, nil

	case lbL7RulesLoadedMsg:
		if msg.err == nil {
			if m.lbL7Rules == nil {
				m.lbL7Rules = make(map[string][]map[string]interface{})
			}
			m.lbL7Rules[msg.policyID] = msg.rules
		}
		return m, nil

	case lbL7RuleCreatedMsg:
		m.wizard = WizardData{}
		m.mode = LBL7RulesView
		m.lbL7RuleDetailIdx = 0
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ L7 Rule saved successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		region := getStringValue(m.detailData, "region", "")
		if msg.policyID != "" && region != "" {
			return m, tea.Batch(
				m.fetchLBL7Rules(msg.policyID, region),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbL7RuleDeletedMsg:
		m.lbL7RuleConfirm = false
		m.mode = LBL7RulesView
		m.lbL7RuleDetailIdx = 0
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ L7 Rule deleted successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		region := getStringValue(m.detailData, "region", "")
		if msg.policyID != "" && region != "" {
			return m, tea.Batch(
				m.fetchLBL7Rules(msg.policyID, region),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbL7PolicyDeletedMsg:
		m.mode = LBListenerDetailView
		m.selectedLBL7Policy = nil
		m.lbL7PolicyDetailActionIdx = 0
		m.lbL7PolicyListIdx = -1
		m.lbL7PolicyListIdx = -1
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ L7 Policy \"%s\" deleted successfully", msg.policyName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if m.selectedLBListener != nil {
			listenerID := getStringValue(m.selectedLBListener, "id", "")
			region := getStringValue(m.detailData, "region", "")
			if listenerID != "" && region != "" {
				return m, tea.Batch(
					m.fetchLBL7Policies(listenerID, region),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbL7PolicyUpdatedMsg:
		m.wizard = WizardData{}
		m.mode = LBListenerDetailView
		m.selectedLBL7Policy = nil
		m.lbL7PolicyDetailActionIdx = 0
		m.lbL7PolicyListIdx = -1
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ L7 Policy \"%s\" updated successfully", msg.policyName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		if m.selectedLBListener != nil {
			listenerID := getStringValue(m.selectedLBListener, "id", "")
			region := getStringValue(m.detailData, "region", "")
			if listenerID != "" && region != "" {
				return m, tea.Batch(
					m.fetchLBL7Policies(listenerID, region),
					tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
				)
			}
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbPoolMembersLoadedMsg:
		if msg.err == nil {
			if m.lbPoolMembers == nil {
				m.lbPoolMembers = make(map[string][]map[string]interface{})
			}
			m.lbPoolMembers[msg.poolID] = msg.members
		}
		return m, nil

	case lbPoolMemberDeletedMsg:
		m.lbPoolMemberConfirm = false
		m.mode = LBPoolMembersView
		m.lbPoolMemberDetailIdx = 0
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ Member deleted successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		region := getStringValue(m.detailData, "region", "")
		if msg.poolID != "" && region != "" {
			return m, tea.Batch(
				m.fetchLBMembers(msg.poolID, region),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbPoolMemberSavedMsg:
		poolID := msg.poolID
		m.mode = LBPoolMembersView
		m.lbPoolMemberDetailIdx = 0
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ Member saved successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		region := getStringValue(m.detailData, "region", "")
		if poolID != "" && region != "" {
			return m, tea.Batch(
				m.fetchLBMembers(poolID, region),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbHMLoadedMsg:
		if msg.err == nil {
			if m.lbHealthMonitors == nil {
				m.lbHealthMonitors = make(map[string]map[string]interface{})
			}
			m.lbHealthMonitors[msg.poolID] = msg.hm // can be nil (no HM)
		}
		return m, nil

	case lbHMSavedMsg:
		poolID := msg.poolID
		m.mode = LBHealthMonitorView
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ Health monitor saved successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		region := getStringValue(m.detailData, "region", "")
		if poolID != "" && region != "" {
			return m, tea.Batch(
				m.fetchLBHealthMonitor(poolID, region),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case lbHMDeletedMsg:
		poolID := msg.poolID
		m.lbHMConfirm = false
		m.mode = LBHealthMonitorView
		if m.lbHealthMonitors != nil {
			m.lbHealthMonitors[poolID] = nil
		}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ Health monitor deleted successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		region := getStringValue(m.detailData, "region", "")
		if poolID != "" && region != "" {
			return m, tea.Batch(
				m.fetchLBHealthMonitor(poolID, region),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })

	case dbCapabilitiesLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.dbEngines = msg.engines
		m.wizard.dbFlavors = msg.flavors
		m.wizard.dbAvailItems = msg.availItems
		m.wizard.dbCapsRegions = msg.capsRegions
		// Build capabilities plan names sorted by order for fallback
		for _, p := range msg.plans {
			if name := getStringValue(p, "name", ""); name != "" {
				m.wizard.dbCapPlans = append(m.wizard.dbCapPlans, name)
			}
		}
		m.wizard.dbEngineIdx = 0
		return m, nil

	case dbDetailSubresourcesMsg:
		if msg.err == nil && (m.detailData == nil || getStringValue(m.detailData, "id", "") == msg.serviceId) {
			m.dbDetailUsers = msg.users
			m.dbDetailBackups = msg.backups
			m.dbDetailDatabases = msg.databases
			m.dbDetailPools = msg.pools
			m.dbDetailLoaded = true
		}
		return m, nil

	case dbServiceDeletedMsg:
		m.actionConfirm = false
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Failed to delete service: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = "✅ Database service deleted successfully"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.detailData = nil
		m.dbDetailUsers = nil
		m.dbDetailBackups = nil
		m.dbDetailDatabases = nil
		m.dbDetailPools = nil
		m.dbDetailLoaded = false
		m.dbUserCreateMode = false
		m.dbUserCreateInput = ""
		m.dbUserCreatedData = nil
		m.mode = LoadingView
		path := "/databases"
		if m.currentProduct == ProductManagedAnalytics {
			path = "/analytics"
		}
		return m, tea.Batch(
			m.fetchDataForPath(path),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case dbUserCreatedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Failed to create user: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.dbDetailLoaded = true
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.dbUserCreatedData = msg.user
		// Refresh user list in the background
		if m.detailData != nil {
			engine := getStringValue(m.detailData, "engine", "")
			serviceId := getStringValue(m.detailData, "id", "")
			if engine != "" && serviceId != "" {
				m.dbDetailLoaded = false
				return m, m.fetchDBDetailSubresources(engine, serviceId)
			}
		}
		m.dbDetailLoaded = true
		return m, nil

	case dbCreatedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.notification = fmt.Sprintf("✅ Database service '%s' created successfully!", msg.dbName)
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/databases"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case analyticsEngineAvailLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.dbAvailItems = msg.availItems
		m.wizard.dbEngine = msg.engine
		m.wizard.dbVersionIdx = 0
		m.wizard.dbVersion = ""
		m.wizard.errorMsg = ""
		m.wizard.step = AnalyticsWizardStepVersion
		return m, nil

	case analyticsCreatedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.notification = fmt.Sprintf("✅ Analytics service '%s' created successfully!", msg.name)
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/analytics"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case lbRegionsLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.lbAvailableRegions = msg.regions
		m.wizard.lbRegionIdx = 0
		return m, nil

	case lbFlavorsLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		// Sort flavors by known size order: small < medium < large < xl < xxl etc.
		sizeOrder := map[string]int{
			"small": 1, "s": 1,
			"medium": 2, "m": 2,
			"large": 3, "l": 3,
			"xl": 4,
			"2xl": 5, "xxl": 5,
			"3xl": 6,
		}
		sortRank := func(name string) int {
			n := strings.ToLower(name)
			if v, ok := sizeOrder[n]; ok {
				return v
			}
			return 99
		}
		sort.Slice(msg.flavors, func(i, j int) bool {
			ri := sortRank(getStringValue(msg.flavors[i], "name", ""))
			rj := sortRank(getStringValue(msg.flavors[j], "name", ""))
			if ri != rj {
				return ri < rj
			}
			return getStringValue(msg.flavors[i], "name", "") < getStringValue(msg.flavors[j], "name", "")
		})
		m.wizard.lbFlavors = msg.flavors
		m.wizard.lbFlavorIdx = 0
		return m, nil

	case lbNetworksLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.lbNetworks = msg.networks
		m.wizard.lbNetworkIdx = 0
		return m, nil

	case lbSubnetLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.lbSubnetId = msg.subnetID
		m.wizard.step = LBWizardStepConfirm
		return m, nil

	case fipRegionsLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.fipAvailableRegions = msg.regions
		m.wizard.fipRegionIdx = 0
		return m, nil

	case fipInstancesLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.fipInstances = msg.instances
		m.wizard.fipInstanceIdx = 0
		return m, nil

	case workflowInstancesLoadedMsg:
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		if msg.err != nil {
			m.wizard.errorMsg = msg.err.Error()
			return m, nil
		}
		m.wizard.wfInstances = msg.instances
		m.wizard.wfInstanceIdx = 0
		return m, nil

	case workflowCreatedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Batch(
				m.fetchDataForPath("/instances/workflow"),
				tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		m.notification = fmt.Sprintf("✅ Workflow \"%s\" created successfully!", msg.name)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/instances/workflow"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case workflowDeletedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Workflow \"%s\" deleted successfully", msg.name)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/instances/workflow"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case instanceBackupDeletedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Backup \"%s\" deleted successfully", msg.name)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/instances/backup"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case fipCreatedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Erreur: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		ip := getString(msg.floatingIP, "ip")
		if ip == "" {
			ip = "en cours de provisioning"
		}
		m.notification = fmt.Sprintf("✅ Floating IP %s created successfully", ip)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/networks/floatingip"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case volumeBackupCreatedMsg:
		m.wizard = WizardData{}
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Error: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			m.mode = TableView
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		label := "Snapshot"
		if msg.backupType == "backup" {
			label = "Backup"
		}
		m.notification = fmt.Sprintf("✅ %s '%s' created successfully", label, msg.name)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.mode = LoadingView
		reloadPath := "/storage/snapshot"
		if msg.backupType == "backup" {
			reloadPath = "/storage/backup"
		}
		return m, tea.Batch(
			m.fetchDataForPath(reloadPath),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case file_storage.ExecuteFileShareActionMsg:
		return m.handleExecuteFileShareAction(msg)

	case fileShareActionDoneMsg:
		return m.handleFileShareActionDone(msg)

	case object_storage.ExecuteUserActionMsg:
		return m.handleExecuteUserAction(msg)

	case s3SecretLoadedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Failed to retrieve secret key: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		if m.objectUserDetailView != nil {
			m.objectUserDetailView.SetSecret(msg.secret)
		}
		return m, nil

	case s3UserActionDoneMsg:
		return m.handleS3UserActionDone(msg)

	case object_storage.ExecuteContainerActionMsg:
		containerName := ""
		region := ""
		if msg.Container != nil {
			if n, ok := msg.Container["name"].(string); ok {
				containerName = n
			}
			if r, ok := msg.Container["region"].(string); ok {
				region = r
			}
		}
		switch msg.Action {
		case object_storage.ContainerActionDelete:
			m.notification = fmt.Sprintf("🗑️  Deleting container '%s'...", containerName)
			m.notificationExpiry = time.Now().Add(30 * time.Second)
			return m, m.deleteObjectContainer(containerName, region)
		case object_storage.ContainerActionChangeType:
			newType := ""
			if msg.ExtraData != nil {
				newType, _ = msg.ExtraData["containerType"].(string)
			}
			containerID := getString(msg.Container, "id")
			if containerID == "" {
				containerID = containerName
			}
			m.notification = fmt.Sprintf("🔄 Changing type of '%s' to %s...", containerName, newType)
			m.notificationExpiry = time.Now().Add(30 * time.Second)
			return m, m.updateSwiftContainerType(containerID, newType)
		case object_storage.ContainerActionAddPolicy:
			var userID int64
			roleName := ""
			if msg.ExtraData != nil {
				roleName, _ = msg.ExtraData["roleName"].(string)
				switch v := msg.ExtraData["userId"].(type) {
				case float64:
					userID = int64(v)
				case int64:
					userID = v
				case int:
					userID = int64(v)
				case json.Number:
					userID, _ = v.Int64()
				case string:
					fmt.Sscanf(v, "%d", &userID)
				}
			}
			m.notification = fmt.Sprintf("🔄 Adding user access to '%s'...", containerName)
			m.notificationExpiry = time.Now().Add(30 * time.Second)
			if userID == 0 {
				m.notification = fmt.Sprintf("❌ Failed to resolve user ID (type: %T, val: %v)", msg.ExtraData["userId"], msg.ExtraData["userId"])
				m.notificationExpiry = time.Now().Add(10 * time.Second)
				return m, tea.Tick(10*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
			}
			return m, m.addS3ContainerPolicy(containerName, region, userID, roleName)
		}

	case views.GoBackMsg:
		if m.mode == DetailView && m.currentProduct == ProductStorageBlock {
			m.volumeDetailView = nil
			m.mode = TableView
			return m, nil
		}
		if m.mode == DetailView && m.currentProduct == ProductStorageSnapshot {
			m.snapshotDetailView = nil
			m.mode = TableView
			return m, nil
		}
		if m.mode == DetailView && m.currentProduct == ProductStorageBackup {
			m.backupDetailView = nil
			m.mode = TableView
			return m, nil
		}
		if m.mode == DetailView && m.currentProduct == ProductStorageFile {
			m.fileShareDetailView = nil
			m.mode = TableView
			return m, nil
		}
		if m.mode == DetailView && m.currentProduct == ProductStorageObject {
			if m.objectUserDetailView != nil {
				m.objectUserDetailView = nil
				m.mode = TableView
				return m, nil
			}
			m.objectDetailView = nil
			m.mode = TableView
			return m, nil
		}
		return m, nil

	case volumeActionDoneMsg:
		return m.handleVolumeActionDone(msg)

	case block_storage.ExecuteSnapshotActionMsg:
		return m, m.executeSnapshotAction(msg)

	case snapshotActionDoneMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Error: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.snapshotDetailView = nil
		m.mode = LoadingView
		if msg.action == block_storage.SnapshotActionCreateVolume {
			m.notification = fmt.Sprintf("✅ Volume '%s' created from snapshot", msg.name)
			m.notificationExpiry = time.Now().Add(5 * time.Second)
			// Navigate to Block Storage so handleDataLoaded matches the product
			m.currentProduct = ProductStorageBlock
			m.storageSubIdx = 0 // Block Storage is index 0 in getStorageSubItems()
			return m, tea.Batch(
				m.fetchDataForPath("/storage/block"),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		m.notification = fmt.Sprintf("✅ Snapshot '%s' deleted", msg.name)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		return m, tea.Batch(
			m.fetchDataForPath("/storage/snapshot"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case block_storage.ExecuteBackupActionMsg:
		return m, m.executeBackupAction(msg)

	case block_storage.LoadBackupRestoreVolumesMsg:
		region := fmt.Sprintf("%v", msg.Backup["region"])
		return m, m.fetchVolumesForRegion(region)

	case block_storage.BackupVolumesLoadedMsg:
		if m.backupDetailView != nil {
			m.backupDetailView.SetRestoreVolumes(msg.Volumes)
		}
		return m, nil

	case backupActionDoneMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Error: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.backupDetailView = nil
		m.mode = LoadingView
		if msg.action == block_storage.BackupActionCreateVolume {
			m.notification = fmt.Sprintf("✅ Volume '%s' created from backup", msg.name)
			m.notificationExpiry = time.Now().Add(5 * time.Second)
			m.currentProduct = ProductStorageBlock
			m.storageSubIdx = 0
			return m, tea.Batch(
				m.fetchDataForPath("/storage/block"),
				tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
			)
		}
		action := fmt.Sprintf("Backup '%s' deleted", msg.name)
		if msg.action == block_storage.BackupActionRestore {
			action = fmt.Sprintf("Backup '%s' restored successfully", msg.name)
		}
		m.notification = "✅ " + action
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		return m, tea.Batch(
			m.fetchDataForPath("/storage/backup"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case refreshBlockStorageMsg:
		return m, m.fetchDataForPath("/storage/block")

	case fileShareRegionsLoadedMsg:
		return m.handleFileShareRegionsLoaded(msg)

	case fileShareNetworksLoadedMsg:
		return m.handleFileShareNetworksLoaded(msg)

	case fileShareSubnetsLoadedMsg:
		return m.handleFileShareSubnetsLoaded(msg)

	case fileShareCreatedMsg:
		return m.handleFileShareCreated(msg)

	case objectStorageInitDataLoadedMsg:
		return m.handleObjectStorageInitDataLoaded(msg)

	case objectContainerCreatedMsg:
		return m.handleObjectContainerCreated(msg)

	case objectContainerActionDoneMsg:
		return m.handleObjectContainerActionDone(msg)

	case swiftContainerUpdatedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Error: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Container '%s' type changed to %s", msg.containerName, msg.newType)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.objectDetailView = nil
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/storage/object"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case containerPolicyAddedMsg:
		if msg.err != nil {
			m.notification = fmt.Sprintf("❌ Error adding policy: %s", msg.err.Error())
			m.notificationExpiry = time.Now().Add(8 * time.Second)
			return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
		}
		m.notification = fmt.Sprintf("✅ Access '%s' added to '%s'", msg.roleName, msg.containerName)
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.objectDetailView = nil
		m.detailData = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/storage/object"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)

	case s3UserCreatedMsg:
		return m.handleS3UserCreated(msg)

	case s3CredentialsSavedMsg:
		return m.handleS3CredentialsSaved(msg)

	case swiftRegionsLoadedMsg:
		m.wizard.objectSwiftRegions = msg.regions
		m.wizard.isLoading = false
		m.wizard.loadingMessage = ""
		return m, nil

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

// handleExecuteUserAction dispatches actions from the user detail view.
func (m Model) handleExecuteUserAction(msg object_storage.ExecuteUserActionMsg) (tea.Model, tea.Cmd) {
	if msg.User == nil {
		return m, nil
	}

	// Extract userID
	var userID int64
	switch v := msg.User["_userId"].(type) {
	case float64:
		userID = int64(v)
	case int64:
		userID = v
	case int:
		userID = int64(v)
	case json.Number:
		userID, _ = v.Int64()
	case string:
		fmt.Sscanf(v, "%d", &userID)
	}
	if userID == 0 {
		m.notification = fmt.Sprintf("❌ Failed to resolve user ID (type: %T, val: %v)", msg.User["_userId"], msg.User["_userId"])
		m.notificationExpiry = time.Now().Add(8 * time.Second)
		return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
	}

	switch msg.Action {
	case object_storage.UserActionShowSecret:
		access := fmt.Sprintf("%v", msg.User["access"])
		m.notification = "🔄 Retrieving secret key..."
		m.notificationExpiry = time.Now().Add(30 * time.Second)
		return m, m.getS3Secret(userID, access)
	case object_storage.UserActionEnable:
		m.s3PendingEnableUser = msg.User
		m.notification = "🔄 Activating user..."
		m.notificationExpiry = time.Now().Add(30 * time.Second)
		return m, m.enableS3User(userID)
	case object_storage.UserActionDisable:
		access := fmt.Sprintf("%v", msg.User["access"])
		m.notification = fmt.Sprintf("🔄 Deactivating... (userID=%d, access=%s)", userID, access)
		m.notificationExpiry = time.Now().Add(30 * time.Second)
		return m, m.disableS3User(userID, access)
	case object_storage.UserActionDeleteUser:
		m.notification = "🗑️  Deleting user..."
		m.notificationExpiry = time.Now().Add(30 * time.Second)
		return m, m.deleteCloudUser(userID)
	}
	return m, nil
}

// handleS3UserActionDone handles the result of enable/disable/delete user actions.
func (m Model) handleS3UserActionDone(msg s3UserActionDoneMsg) (tea.Model, tea.Cmd) {
	if msg.err != nil {
		m.notification = fmt.Sprintf("❌ Error: %s", msg.err.Error())
		m.notificationExpiry = time.Now().Add(8 * time.Second)
		return m, tea.Tick(8*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
	}

	switch msg.action {
	case object_storage.UserActionEnable:
		// Show S3CredentialsView like after user creation
		username := ""
		if m.s3PendingEnableUser != nil {
			if u, ok := m.s3PendingEnableUser["_username"].(string); ok && u != "" {
				username = u
			}
		}
		m.s3CreatedUser = map[string]interface{}{"username": username}
		m.s3CreatedCredentials = msg.newCredential
		m.s3CredentialsSavedPath = ""
		m.s3CredentialsSaveError = ""
		m.s3PendingEnableUser = nil
		m.objectUserDetailView = nil
		m.s3CredentialsFromEnable = true
		m.mode = S3CredentialsView
		return m, nil
	case object_storage.UserActionDisable:
		m.notification = "✅ User deactivated"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		m.objectUserDetailView = nil
		m.mode = LoadingView
		return m, tea.Batch(
			m.fetchDataForPath("/storage/object"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)
	case object_storage.UserActionDeleteUser:
		m.objectUserDetailView = nil
		m.mode = LoadingView
		m.notification = "✅ User deleted"
		m.notificationExpiry = time.Now().Add(5 * time.Second)
		return m, tea.Batch(
			m.fetchDataForPath("/storage/object"),
			tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} }),
		)
	}
	return m, nil
}

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

	// Add extra spacing when a project is selected
	if m.cloudProject != "" && m.mode != ProjectSelectView {
		content.WriteString("\n")
	}

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
	experimental := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Bold(true).
		Render(" [EXPERIMENTAL]")

	// Show selected project in header if one is selected
	if m.cloudProject != "" && m.mode != ProjectSelectView {
		projectInfo := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Render(fmt.Sprintf(" • Project: %s", m.cloudProjectName))
		return logo + experimental + projectInfo
	}
	return logo + experimental
}

func (m Model) renderNavBar(width int) string {
	// Don't show nav bar in project selection mode
	if m.mode == ProjectSelectView || m.currentProduct == ProductProjects {
		return ""
	}

	navItems := getNavItems()
	var items []string

	isSubNavFocused := m.inStorageSubNav || m.inNetworkSubNav || m.inComputeSubNav
	isInSubContext := isSubNavFocused || m.inTableFocus

	for i, nav := range navItems {
		var style lipgloss.Style
		if i == m.navIdx && !isInSubContext {
			// Level 1 focus: full green highlight
			style = navItemSelectedStyle
		} else if i == m.navIdx {
			// Sub-level active: dim indicator so user knows which section they're in
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#007733")).
				Padding(0, 2)
		} else {
			style = navItemStyle
		}
		items = append(items, style.Render(fmt.Sprintf("%s %s", nav.Icon, nav.Label)))
	}

	navContent := lipgloss.JoinHorizontal(lipgloss.Top, items...)
	mainNav := navBarStyle.Width(width - 2).Render(navContent)

	// Show storage sub-navigation when on Stockage or any storage sub-product
	isStorageContext := navItems[m.navIdx].Product == ProductStorage ||
		(m.currentProduct >= ProductStorageBlock && m.currentProduct <= ProductStorageArchive)
	if isStorageContext {
		subNav := m.renderStorageSubNav(width)
		return mainNav + "\n" + subNav
	}

	// Show network sub-navigation when on Networks or any network sub-product
	isNetworkContext := navItems[m.navIdx].Product == ProductNetworks ||
		(m.currentProduct >= ProductNetworkPrivate && m.currentProduct <= ProductNetworkLB)
	if isNetworkContext {
		subNav := m.renderNetworkSubNav(width)
		return mainNav + "\n" + subNav
	}

	// Show compute sub-navigation when on Compute or any compute sub-product
	isComputeSubProduct := m.currentProduct == ProductInstances || m.currentProduct == ProductInstanceBackup || m.currentProduct == ProductWorkflow
	isComputeContext := navItems[m.navIdx].Product == ProductCompute || isComputeSubProduct
	if isComputeContext {
		subNav := m.renderComputeSubNav(width)
		return mainNav + "\n" + subNav
	}

	return mainNav
}

func (m Model) renderStorageSubNav(width int) string {
	subItems := getStorageSubItems()
	var items []string

	subItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 2)
	// subItemSelectedStyle := lipgloss.NewStyle().
	// 	Foreground(lipgloss.Color("#00FF7F")).
	// 	Bold(true).
	// 	Padding(0, 2).
	// 	Underline(true)
	subItemDisabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Padding(0, 2)

	// Determine the active sub-item from currentProduct so it stays correct
	// even when storageSubIdx gets out of sync.
	activeSubIdx := m.storageSubIdx
	for i, item := range subItems {
		if item.Product == m.currentProduct {
			activeSubIdx = i
			break
		}
	}

	for i, item := range subItems {
		var style lipgloss.Style
		label := item.Label
		if i == activeSubIdx && m.inStorageSubNav && m.inTableFocus {
			// Level 3: focus moved into the table — show arrow hint, dimmed
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AA55")).Padding(0, 2)
			label = "▼ " + item.Label
		} else if i == activeSubIdx && m.inStorageSubNav {
			// Level 2: sub-nav focused — bright green + bold
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Padding(0, 2)
		} else if i == activeSubIdx {
			// Active item, focus is on main nav — dimmed green
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AA55")).Padding(0, 2)
		} else if !item.Enabled {
			style = subItemDisabledStyle
		} else {
			style = subItemStyle
		}
		items = append(items, style.Render(label))
	}

	borderColor := lipgloss.Color("#333333")
	if m.inStorageSubNav && !m.inTableFocus {
		// Level 2: sub-nav is focused — bright green border
		borderColor = lipgloss.Color("#00FF7F")
	} else if m.inTableFocus {
		// Level 3: focus is inside the table — dim the sub-nav border
		borderColor = lipgloss.Color("#444444")
	}
	subBarStyle := lipgloss.NewStyle().
		Padding(0, 1).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor)
	subContent := lipgloss.JoinHorizontal(lipgloss.Top, items...)
	return subBarStyle.Width(width - 2).Render(subContent)
}

func (m Model) renderNetworkSubNav(width int) string {
	subItems := getNetworkSubItems()
	var items []string

	subItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 2)
	subItemDisabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Padding(0, 2)

	activeSubIdx := m.networkSubIdx
	for i, item := range subItems {
		if item.Product == m.currentProduct {
			activeSubIdx = i
			break
		}
	}

	for i, item := range subItems {
		var style lipgloss.Style
		label := item.Label
		if i == activeSubIdx && m.inNetworkSubNav && m.inTableFocus {
			// Level 3: focus moved into the table — show arrow hint, dimmed
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AA55")).Padding(0, 2)
			label = "▼ " + item.Label
		} else if i == activeSubIdx && m.inNetworkSubNav {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Padding(0, 2)
		} else if i == activeSubIdx {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AA55")).Padding(0, 2)
		} else if !item.Enabled {
			style = subItemDisabledStyle
		} else {
			style = subItemStyle
		}
		items = append(items, style.Render(label))
	}

	borderColor := lipgloss.Color("#333333")
	if m.inNetworkSubNav && !m.inTableFocus {
		// Level 2: sub-nav is focused — bright green border
		borderColor = lipgloss.Color("#00FF7F")
	} else if m.inTableFocus {
		// Level 3: focus is inside the table — dim the sub-nav border
		borderColor = lipgloss.Color("#444444")
	}
	subBarStyle := lipgloss.NewStyle().
		Padding(0, 1).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor)
	subContent := lipgloss.JoinHorizontal(lipgloss.Top, items...)
	return subBarStyle.Width(width - 2).Render(subContent)
}

func (m Model) renderComputeSubNav(width int) string {
	subItems := getComputeSubItems()
	var items []string

	subItemStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 2)
	subItemDisabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Padding(0, 2)

	activeSubIdx := m.computeSubIdx
	for i, item := range subItems {
		if item.Product == m.currentProduct {
			activeSubIdx = i
			break
		}
	}

	for i, item := range subItems {
		var style lipgloss.Style
		label := item.Label
		if i == activeSubIdx && m.inComputeSubNav && m.inTableFocus {
			// Level 3: focus moved into the table — show arrow hint, dimmed
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AA55")).Padding(0, 2)
			label = "▼ " + item.Label
		} else if i == activeSubIdx && m.inComputeSubNav {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true).Padding(0, 2)
		} else if i == activeSubIdx {
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("#00AA55")).Padding(0, 2)
		} else if !item.Enabled {
			style = subItemDisabledStyle
		} else {
			style = subItemStyle
		}
		items = append(items, style.Render(label))
	}

	borderColor := lipgloss.Color("#333333")
	if m.inComputeSubNav && !m.inTableFocus {
		// Level 2: sub-nav is focused — bright green border
		borderColor = lipgloss.Color("#00FF7F")
	} else if m.inTableFocus {
		// Level 3: focus is inside the table — dim the sub-nav border
		borderColor = lipgloss.Color("#444444")
	}
	subBarStyle := lipgloss.NewStyle().
		Padding(0, 1).
		BorderTop(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(borderColor)
	subContent := lipgloss.JoinHorizontal(lipgloss.Top, items...)
	return subBarStyle.Width(width - 2).Render(subContent)
}

func (m Model) renderContentBox(width int) string {
	var titleText string

	// Handle wizard mode with special title
	if m.mode == WizardView {
		// Determine which wizard we're in based on the step
		if m.wizard.step >= 1600 {
			titleText = " 📋 Create L7 Rule "
		} else if m.wizard.step >= 1500 {
			titleText = " 📋 Create L7 Policy "
		} else if m.wizard.step >= 1400 {
			titleText = " 🔊 Create Listener "
		} else if m.wizard.step >= 1300 {
			if m.wizard.lbPoolEditPoolId != "" {
				titleText = " ✏️  Edit Pool "
			} else {
				titleText = " ⚖️  Create Pool "
			}
		} else if m.wizard.step >= 1200 {
			// Workflow wizard
			titleText = " ⚙️  Create a backup Workflow "
		} else if m.wizard.step >= 1100 {
			// Floating IP wizard
			titleText = " 🌐 Create Floating IP "
		} else if m.wizard.step >= 1000 {
			// Load Balancer wizard
			titleText = " ⚖\ufe0f  Create Load Balancer "
		} else if m.wizard.step >= 900 {
			// Gateway wizard
			titleText = " 🌐 Create Gateway "
		} else if m.wizard.step >= 800 {
			// Private Network wizard
			titleText = " 🌐 Create Private Network "
		} else if m.wizard.step >= 700 {
			// Backup/Snapshot wizard
			titleText = " 💾 Create Volume Backup / Snapshot "
		} else if m.wizard.step >= 600 {
			// S3 User wizard
			titleText = " 👤 Create S3 User "
		} else if m.wizard.step >= 500 {
			// Object Storage wizard
			titleText = " 🪣  Create Object Storage Container "
		} else if m.wizard.step >= 400 {
			// File Storage wizard
			titleText = " 🗂️  Create File Share "
		} else if m.wizard.step >= 300 {
			// Volume wizard
			titleText = " 💾 Create Volume "
		} else if m.wizard.step >= 200 {
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

	// Handle S3 credentials view
	if m.mode == S3CredentialsView {
		titleText = " 🔑 S3 User Credentials "
		title := productTitleStyle.Render(titleText)
		contentStr := m.renderS3CredentialsView(width - 6)
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
		if m.mode == DetailView && m.currentProduct == ProductStorageObject && m.objectUserDetailView != nil {
			titleText = m.objectUserDetailView.Title()
		} else if m.mode == DetailView && m.currentProduct == ProductStorageObject && m.objectDetailView != nil {
			titleText = m.objectDetailView.Title()
		} else if m.mode == DetailView && m.currentProduct == ProductStorageSnapshot && m.snapshotDetailView != nil {
			titleText = m.snapshotDetailView.Title()
		} else if m.mode == DetailView && m.currentProduct == ProductStorageBackup && m.backupDetailView != nil {
			titleText = m.backupDetailView.Title()
		} else if m.mode == DetailView && m.currentItemName != "" {
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
		// Add tabs for Object Storage
		if m.currentProduct == ProductStorageObject {
			contentStr = m.renderObjectStorageWithTabs(contentStr, width-6)
		}
		// Add tabs for Private Networks
		if m.currentProduct == ProductNetworkPrivate {
			contentStr = m.renderPrivateNetworksWithTabs(contentStr, width-6)
		}
	case DetailView:
		contentStr = m.renderDetailView(width - 6)
	case NodePoolsView:
		contentStr = m.renderNodePoolsView(width - 6)
	case NodePoolDetailView:
		contentStr = m.renderNodePoolDetailView(width - 6)
	case LBPoolDetailView:
		contentStr = m.renderLBPoolDetailView(width - 6)
	case LBListenerDetailView:
		contentStr = m.renderLBListenerDetailView(width - 6)
	case LBL7PolicyDetailView:
		contentStr = m.renderLBL7PolicyDetailView(width - 6)
	case LBL7RulesView:
		contentStr = m.renderLBL7RulesView(width - 6)
	case LBPoolMembersView:
		contentStr = m.renderLBPoolMembersView(width - 6)
	case LBHealthMonitorView:
		contentStr = m.renderLBHealthMonitorView(width - 6)
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
	case KubeKubeconfigPickerView:
		contentStr = m.renderKubeKubeconfigPickerView(width - 6)
	case ComingSoonView:
		contentStr = m.renderComingSoonView()
	}

	// Combine title and content
	fullContent := title + "\n\n" + contentStr

	// Clamp content height so the nav bars and footer are never pushed off screen.
	// ~12 lines are consumed by header, nav bars, spacing and footer outside this box.
	if m.height > 14 {
		maxLines := m.height - 12
		lines := strings.Split(fullContent, "\n")
		if len(lines) > maxLines {
			lines = lines[:maxLines]
			fullContent = strings.Join(lines, "\n")
		}
	}

	// Level 3 (inTableFocus): highlight content box border in green to show focus is inside the table
	boxStyle := contentBoxStyle
	if m.inTableFocus {
		boxStyle = contentBoxStyle.BorderForeground(lipgloss.Color("#00FF7F"))
	}
	return boxStyle.Width(width - 4).Render(fullContent)
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

// renderComingSoonView displays a placeholder for products not yet implemented
func (m Model) renderComingSoonView() string {
	var content strings.Builder

	navItems := getNavItems()
	currentNav := navItems[m.navIdx]

	iconStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500"))

	textStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	highlightStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B68EE")).
		Bold(true)

	content.WriteString("\n\n")
	content.WriteString(fmt.Sprintf("        %s\n\n", iconStyle.Render("🚧")))
	content.WriteString(fmt.Sprintf("        %s\n\n", highlightStyle.Render(currentNav.Label+" - Coming Soon")))
	content.WriteString(fmt.Sprintf("        %s\n\n", textStyle.Render("This section is not yet implemented in the browser.")))
	content.WriteString(fmt.Sprintf("        %s\n", textStyle.Render("Stay tuned for future updates!")))

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

// renderKubeKubeconfigPickerView displays the directory picker for saving a kubeconfig file.
func (m Model) renderKubeKubeconfigPickerView(width int) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7B68EE")).
		Bold(true)
	content.WriteString(headerStyle.Render("  💾  Save Kubeconfig\n\n"))

	clusterName := getStringValue(m.detailData, "name", "Unknown")
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(fmt.Sprintf("  %s %s\n", labelStyle.Render("Cluster:"), valueStyle.Render(clusterName)))
	content.WriteString(fmt.Sprintf("  %s %s\n\n", labelStyle.Render("Directory:"), valueStyle.Render(m.wizard.kubeKubeconfigCurrentDir)))

	// Build entry list: "..", "[Save here]", then subdirs
	entries := []struct {
		label    string
		isAction bool
	}{
		{"  ..", false},
		{"  [ Save here ]", true},
	}
	for _, d := range m.wizard.kubeKubeconfigEntries {
		entries = append(entries, struct {
			label    string
			isAction bool
		}{"  " + d + "/", false})
	}

	selectedStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7B68EE")).
		Foreground(lipgloss.Color("#FFFFFF"))
	actionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF7F")).
		Bold(true)
	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	for i, e := range entries {
		var rendered string
		if i == m.wizard.kubeKubeconfigSelectedIdx {
			rendered = selectedStyle.Render(fmt.Sprintf("> %s", e.label))
		} else if e.isAction {
			rendered = actionStyle.Render(e.label)
		} else {
			rendered = normalStyle.Render(e.label)
		}
		content.WriteString(rendered + "\n")
	}

	content.WriteString("\n")
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	filename := clusterName + "-kubeconfig.yaml"
	content.WriteString(dimStyle.Render(fmt.Sprintf("  File will be saved as: %s\n", filename)))

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
	if m.wizard.step >= 1600 {
		// L7 Rule wizard
		steps = append(steps, "Type", "Comparison", "Key", "Value", "Invert", "Confirm")
		stepMapping = append(stepMapping, LBL7RuleWizardStepType, LBL7RuleWizardStepCompare, LBL7RuleWizardStepKey, LBL7RuleWizardStepValue, LBL7RuleWizardStepInvert, LBL7RuleWizardStepConfirm)
	} else if m.wizard.step >= 1500 {
		// L7 Policy wizard
		steps = append(steps, "Name", "Position", "Action", "Confirm")
		stepMapping = append(stepMapping, LBL7PolicyWizardStepName, LBL7PolicyWizardStepPosition, LBL7PolicyWizardStepAction, LBL7PolicyWizardStepConfirm)
	} else if m.wizard.step >= 1400 {
		// LB Listener wizard
		steps = append(steps, "Name", "Protocol", "Port", "Pool", "Confirm")
		stepMapping = append(stepMapping, LBListenerWizardStepName, LBListenerWizardStepProto, LBListenerWizardStepPort, LBListenerWizardStepPool, LBListenerWizardStepConfirm)
	} else if m.wizard.step >= 1300 {
		// LB Pool wizard
		steps = append(steps, "Name", "Algorithm", "Protocol", "Session", "Confirm")
		stepMapping = append(stepMapping, LBPoolWizardStepName, LBPoolWizardStepAlgo, LBPoolWizardStepProto, LBPoolWizardStepSession, LBPoolWizardStepConfirm)
	} else if m.wizard.step >= 1200 {
		// Workflow wizard
		steps = append(steps, "Type", "Instance", "Nom", "Planification", "Confirmer")
		stepMapping = append(stepMapping, WorkflowWizardStepType, WorkflowWizardStepInstance, WorkflowWizardStepName, WorkflowWizardStepSchedule, WorkflowWizardStepConfirm)
	} else if m.wizard.step >= 1100 {
		// Floating IP wizard
		steps = append(steps, "Region", "Instance", "Confirm")
		stepMapping = append(stepMapping, FIPWizardStepRegion, FIPWizardStepInstance, FIPWizardStepConfirm)
	} else if m.wizard.step >= 2000 {
		// Managed Analytics wizard
		steps = append(steps, "Name", "Engine", "Version", "Region", "Plan", "Flavor", "Nodes", "Storage", "Network", "Confirm")
		stepMapping = append(stepMapping, AnalyticsWizardStepName, AnalyticsWizardStepEngine, AnalyticsWizardStepVersion, AnalyticsWizardStepRegion, AnalyticsWizardStepPlan, AnalyticsWizardStepFlavor, AnalyticsWizardStepNodes, AnalyticsWizardStepStorage, AnalyticsWizardStepNetwork, AnalyticsWizardStepConfirm)
	} else if m.wizard.step >= 1900 {
		// Managed Database wizard
		steps = append(steps, "Name", "Engine", "Version", "Region", "Plan", "Flavor", "Nodes", "Storage", "Network", "Confirm")
		stepMapping = append(stepMapping, DBWizardStepName, DBWizardStepEngine, DBWizardStepVersion, DBWizardStepRegion, DBWizardStepPlan, DBWizardStepFlavor, DBWizardStepNodes, DBWizardStepStorage, DBWizardStepNetwork, DBWizardStepConfirm)
	} else if m.wizard.step >= 1000 {
		// Load Balancer wizard
		steps = append(steps, "Name", "Region", "Size", "Network", "Confirm")
		stepMapping = append(stepMapping, LBWizardStepName, LBWizardStepRegion, LBWizardStepFlavor, LBWizardStepNetwork, LBWizardStepConfirm)
	} else if m.wizard.step >= 900 {
		// Gateway wizard
		steps = append(steps, "Region", "Size", "Name", "Network", "Confirm")
		stepMapping = append(stepMapping, GwWizardStepRegion, GwWizardStepModel, GwWizardStepName, GwWizardStepNetwork, GwWizardStepConfirm)
	} else if m.wizard.step >= 800 {
		// Private Network wizard
		if m.wizard.privNetAddSubnetMode {
			steps = append(steps, "Region", "Subnet", "DHCP", "IP Pool", "Gateway", "Confirm")
			stepMapping = append(stepMapping, PrivNetWizardStepRegion, PrivNetWizardStepSubnet, PrivNetWizardStepDHCP, PrivNetWizardStepAllocPool, PrivNetWizardStepGateway, PrivNetWizardStepConfirm)
		} else {
			steps = append(steps, "Region", "Name", "VLAN", "Subnet", "DHCP", "IP Pool", "Gateway", "Confirm")
			stepMapping = append(stepMapping, PrivNetWizardStepRegion, PrivNetWizardStepName, PrivNetWizardStepVlanID, PrivNetWizardStepSubnet, PrivNetWizardStepDHCP, PrivNetWizardStepAllocPool, PrivNetWizardStepGateway, PrivNetWizardStepConfirm)
		}
	} else if m.wizard.step >= 700 {
		// Backup/Snapshot wizard
		steps = append(steps, "Volume", "Type", "Name", "Confirm")
		stepMapping = append(stepMapping, BackupWizardStepVolume, BackupWizardStepType, BackupWizardStepName, BackupWizardStepConfirm)
	} else if m.wizard.step >= 600 {
		steps = append(steps, "Description", "Confirm")
		stepMapping = append(stepMapping, S3UserWizardStepDescription, S3UserWizardStepConfirm)
	} else if m.wizard.step >= 500 {
		// Object Storage wizard
		steps = append(steps, "Name", "Type", "Region", "Replication", "Versions", "Lock", "User", "Encryption", "Confirm")
		stepMapping = append(stepMapping, ObjectWizardStepName, ObjectWizardStepType, ObjectWizardStepRegion, ObjectWizardStepReplication, ObjectWizardStepVersioning, ObjectWizardStepObjectLock, ObjectWizardStepUser, ObjectWizardStepEncryption, ObjectWizardStepConfirm)
	} else if m.wizard.step >= 400 {
		// File Storage wizard
		steps = append(steps, "Name", "Region", "Type", "Size", "Network", "Confirm")
		stepMapping = append(stepMapping, FileWizardStepName, FileWizardStepRegion, FileWizardStepType, FileWizardStepSize, FileWizardStepNetwork, FileWizardStepConfirm)
	} else if m.wizard.step >= 300 {
		// Volume wizard
		steps = append(steps, "Name", "Region", "Type", "Avail. Zone", "Size", "Encryption", "Confirm")
		stepMapping = append(stepMapping, VolumeWizardStepName, VolumeWizardStepRegion, VolumeWizardStepType, VolumeWizardStepAvailabilityZone, VolumeWizardStepSize, VolumeWizardStepEncryption, VolumeWizardStepConfirm)
	} else if m.wizard.step >= 200 {
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
	// Volume wizard steps
	case VolumeWizardStepName:
		content.WriteString(m.renderVolumeWizardNameStep(width))
	case VolumeWizardStepRegion:
		content.WriteString(m.renderVolumeWizardRegionStep(width))
	case VolumeWizardStepType:
		content.WriteString(m.renderVolumeWizardTypeStep(width))
	case VolumeWizardStepAvailabilityZone:
		content.WriteString(m.renderVolumeWizardAZStep(width))
	case VolumeWizardStepSize:
		content.WriteString(m.renderVolumeWizardSizeStep(width))
	case VolumeWizardStepEncryption:
		content.WriteString(m.renderVolumeWizardEncryptionStep(width))
	case VolumeWizardStepConfirm:
		content.WriteString(m.renderVolumeWizardConfirmStep(width))
	// File Storage wizard steps
	case FileWizardStepName:
		content.WriteString(m.renderFileWizardNameStep(width))
	case FileWizardStepRegion:
		content.WriteString(m.renderFileWizardRegionStep(width))
	case FileWizardStepType:
		content.WriteString(m.renderFileWizardTypeStep(width))
	case FileWizardStepSize:
		content.WriteString(m.renderFileWizardSizeStep(width))
	case FileWizardStepNetwork:
		content.WriteString(m.renderFileWizardNetworkStep(width))
	case FileWizardStepConfirm:
		content.WriteString(m.renderFileWizardConfirmStep(width))
	case ObjectWizardStepName:
			content.WriteString(m.renderObjectWizardNameStep(width))
	case ObjectWizardStepType:
			content.WriteString(m.renderObjectWizardTypeStep(width))
	case ObjectWizardStepRegion:
			content.WriteString(m.renderObjectWizardRegionStep(width))
	case ObjectWizardStepReplication:
			content.WriteString(m.renderObjectWizardReplicationStep(width))
	case ObjectWizardStepVersioning:
			content.WriteString(m.renderObjectWizardVersioningStep(width))
	case ObjectWizardStepObjectLock:
			content.WriteString(m.renderObjectWizardObjectLockStep(width))
	case ObjectWizardStepUser:
			content.WriteString(m.renderObjectWizardUserStep(width))
	case ObjectWizardStepEncryption:
			content.WriteString(m.renderObjectWizardEncryptionStep(width))
	case ObjectWizardStepConfirm:
			content.WriteString(m.renderObjectWizardConfirmStep(width))
	case ObjectWizardStepSwiftType:
		content.WriteString(m.renderObjectWizardSwiftTypeStep(width))
	case ObjectWizardStepSwiftRegion:
		content.WriteString(m.renderObjectWizardSwiftRegionStep(width))
	// S3 User wizard steps
	case S3UserWizardStepDescription:
		content.WriteString(m.renderS3UserWizardDescStep(width))
	case S3UserWizardStepConfirm:
		content.WriteString(m.renderS3UserWizardConfirmStep(width))
	// Private Network wizard steps
	case PrivNetWizardStepRegion:
		content.WriteString(m.renderPrivNetWizardRegionStep(width))
	case PrivNetWizardStepName:
		content.WriteString(m.renderPrivNetWizardNameStep(width))
	case PrivNetWizardStepVlanID:
		content.WriteString(m.renderPrivNetWizardVlanStep(width))
	case PrivNetWizardStepSubnet:
		content.WriteString(m.renderPrivNetWizardSubnetStep(width))
	case PrivNetWizardStepDHCP:
		content.WriteString(m.renderPrivNetWizardDHCPStep(width))
	case PrivNetWizardStepAllocPool:
		content.WriteString(m.renderPrivNetWizardAllocPoolStep(width))
	case PrivNetWizardStepGateway:
		content.WriteString(m.renderPrivNetWizardGatewayStep(width))
	case PrivNetWizardStepConfirm:
		content.WriteString(m.renderPrivNetWizardConfirmStep(width))
	// Gateway wizard steps
	case GwWizardStepRegion:
		content.WriteString(m.renderGwWizardRegionStep(width))
	case GwWizardStepModel:
		content.WriteString(m.renderGwWizardModelStep(width))
	case GwWizardStepName:
		content.WriteString(m.renderGwWizardNameStep(width))
	case GwWizardStepNetwork:
		content.WriteString(m.renderGwWizardNetworkStep(width))
	case GwWizardStepConfirm:
		content.WriteString(m.renderGwWizardConfirmStep(width))
	// Managed Database wizard steps
	case DBWizardStepName:
		content.WriteString(m.renderDBWizardNameStep(width))
	case DBWizardStepEngine:
		content.WriteString(m.renderDBWizardEngineStep(width))
	case DBWizardStepVersion:
		content.WriteString(m.renderDBWizardVersionStep(width))
	case DBWizardStepRegion:
		content.WriteString(m.renderDBWizardRegionStep(width))
	case DBWizardStepPlan:
		content.WriteString(m.renderDBWizardPlanStep(width))
	case DBWizardStepFlavor:
		content.WriteString(m.renderDBWizardFlavorStep(width))
	case DBWizardStepNodes:
		content.WriteString(m.renderDBWizardNodesStep(width))
	case DBWizardStepStorage:
		content.WriteString(m.renderDBWizardStorageStep(width))
	case DBWizardStepNetwork:
		content.WriteString(m.renderDBWizardNetworkStep(width))
	case DBWizardStepConfirm:
		content.WriteString(m.renderDBWizardConfirmStep(width))
	// Managed Analytics wizard steps
	case AnalyticsWizardStepName:
		content.WriteString(m.renderAnalyticsWizardNameStep(width))
	case AnalyticsWizardStepEngine:
		content.WriteString(m.renderAnalyticsWizardEngineStep(width))
	case AnalyticsWizardStepVersion:
		content.WriteString(m.renderAnalyticsWizardVersionStep(width))
	case AnalyticsWizardStepRegion:
		content.WriteString(m.renderAnalyticsWizardRegionStep(width))
	case AnalyticsWizardStepPlan:
		content.WriteString(m.renderAnalyticsWizardPlanStep(width))
	case AnalyticsWizardStepFlavor:
		content.WriteString(m.renderAnalyticsWizardFlavorStep(width))
	case AnalyticsWizardStepNodes:
		content.WriteString(m.renderAnalyticsWizardNodesStep(width))
	case AnalyticsWizardStepStorage:
		content.WriteString(m.renderAnalyticsWizardStorageStep(width))
	case AnalyticsWizardStepNetwork:
		content.WriteString(m.renderAnalyticsWizardNetworkStep(width))
	case AnalyticsWizardStepConfirm:
		content.WriteString(m.renderAnalyticsWizardConfirmStep(width))
	// Load Balancer wizard steps
	case LBWizardStepName:
		content.WriteString(m.renderLBWizardNameStep(width))
	case LBWizardStepRegion:
		content.WriteString(m.renderLBWizardRegionStep(width))
	case LBWizardStepFlavor:
		content.WriteString(m.renderLBWizardFlavorStep(width))
	case LBWizardStepNetwork:
		content.WriteString(m.renderLBWizardNetworkStep(width))
	case LBWizardStepConfirm:
		content.WriteString(m.renderLBWizardConfirmStep(width))
	// Floating IP wizard steps
	case FIPWizardStepRegion:
		content.WriteString(m.renderFIPWizardRegionStep(width))
	case FIPWizardStepInstance:
		content.WriteString(m.renderFIPWizardInstanceStep(width))
	case FIPWizardStepConfirm:
		content.WriteString(m.renderFIPWizardConfirmStep(width))
	// LB Pool wizard steps
	case LBPoolWizardStepName:
		content.WriteString(m.renderLBPoolWizardNameStep(width))
	case LBPoolWizardStepAlgo:
		content.WriteString(m.renderLBPoolWizardAlgoStep(width))
	case LBPoolWizardStepProto:
		content.WriteString(m.renderLBPoolWizardProtoStep(width))
	case LBPoolWizardStepSession:
		content.WriteString(m.renderLBPoolWizardSessionStep(width))
	case LBPoolWizardStepConfirm:
		content.WriteString(m.renderLBPoolWizardConfirmStep(width))
	// LB Listener wizard steps
	case LBListenerWizardStepName:
		content.WriteString(m.renderLBListenerWizardNameStep(width))
	case LBListenerWizardStepProto:
		content.WriteString(m.renderLBListenerWizardProtoStep(width))
	case LBListenerWizardStepPort:
		content.WriteString(m.renderLBListenerWizardPortStep(width))
	case LBListenerWizardStepPool:
		content.WriteString(m.renderLBListenerWizardPoolStep(width))
	case LBListenerWizardStepConfirm:
		content.WriteString(m.renderLBListenerWizardConfirmStep(width))
	// L7 Policy wizard steps
	case LBL7PolicyWizardStepName:
		content.WriteString(m.renderLBL7PolicyWizardNameStep(width))
	case LBL7PolicyWizardStepPosition:
		content.WriteString(m.renderLBL7PolicyWizardPositionStep(width))
	case LBL7PolicyWizardStepAction:
		content.WriteString(m.renderLBL7PolicyWizardActionStep(width))
	case LBL7PolicyWizardStepRedirectPool:
		content.WriteString(m.renderLBL7PolicyWizardRedirectPoolStep(width))
	case LBL7PolicyWizardStepRedirectUrl:
		content.WriteString(m.renderLBL7PolicyWizardRedirectUrlStep(width))
	case LBL7PolicyWizardStepConfirm:
		content.WriteString(m.renderLBL7PolicyWizardConfirmStep(width))
	// L7 Rule wizard steps
	case LBL7RuleWizardStepType:
		content.WriteString(m.renderLBL7RuleWizardTypeStep(width))
	case LBL7RuleWizardStepCompare:
		content.WriteString(m.renderLBL7RuleWizardCompareStep(width))
	case LBL7RuleWizardStepKey:
		content.WriteString(m.renderLBL7RuleWizardKeyStep(width))
	case LBL7RuleWizardStepValue:
		content.WriteString(m.renderLBL7RuleWizardValueStep(width))
	case LBL7RuleWizardStepInvert:
		content.WriteString(m.renderLBL7RuleWizardInvertStep(width))
	case LBL7RuleWizardStepConfirm:
		content.WriteString(m.renderLBL7RuleWizardConfirmStep(width))
	// LB Member wizard steps
	case LBMemberWizardStepName:
		content.WriteString(m.renderLBMemberWizardNameStep(width))
	case LBMemberWizardStepIP:
		content.WriteString(m.renderLBMemberWizardIPStep(width))
	case LBMemberWizardStepPort:
		content.WriteString(m.renderLBMemberWizardPortStep(width))
	case LBMemberWizardStepWeight:
		content.WriteString(m.renderLBMemberWizardWeightStep(width))
	case LBMemberWizardStepConfirm:
		content.WriteString(m.renderLBMemberWizardConfirmStep(width))
	// LB Health Monitor wizard steps
	case LBHMWizardStepName:
		content.WriteString(m.renderLBHMWizardNameStep(width))
	case LBHMWizardStepType:
		content.WriteString(m.renderLBHMWizardTypeStep(width))
	case LBHMWizardStepHttpMethod:
		content.WriteString(m.renderLBHMWizardHttpMethodStep(width))
	case LBHMWizardStepUrlPath:
		content.WriteString(m.renderLBHMWizardUrlPathStep(width))
	case LBHMWizardStepExpectedCodes:
		content.WriteString(m.renderLBHMWizardExpectedCodesStep(width))
	case LBHMWizardStepDelay:
		content.WriteString(m.renderLBHMWizardDelayStep(width))
	case LBHMWizardStepMaxRetries:
		content.WriteString(m.renderLBHMWizardMaxRetriesStep(width))
	case LBHMWizardStepMaxRetriesDown:
		content.WriteString(m.renderLBHMWizardMaxRetriesDownStep(width))
	case LBHMWizardStepTimeout:
		content.WriteString(m.renderLBHMWizardTimeoutStep(width))
	case LBHMWizardStepConfirm:
		content.WriteString(m.renderLBHMWizardConfirmStep(width))
	// Workflow wizard steps
	case WorkflowWizardStepType:
		content.WriteString(m.renderWorkflowWizardTypeStep(width))
	case WorkflowWizardStepInstance:
		content.WriteString(m.renderWorkflowWizardInstanceStep(width))
	case WorkflowWizardStepName:
		content.WriteString(m.renderWorkflowWizardNameStep(width))
	case WorkflowWizardStepSchedule:
		content.WriteString(m.renderWorkflowWizardScheduleStep(width))
	case WorkflowWizardStepConfirm:
		content.WriteString(m.renderWorkflowWizardConfirmStep(width))
	// Volume Backup / Snapshot wizard steps
	case BackupWizardStepVolume, BackupWizardStepType, BackupWizardStepName, BackupWizardStepConfirm:
		content.WriteString(m.renderBackupWizard(width))
	}
	return content.String()
}

func (m Model) renderWizardRegionStep(width int) string {
	var content strings.Builder

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
		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + region))
		} else {
			content.WriteString(listStyle.Render("  " + region))
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

// ========== Volume Wizard Render Functions ==========

func (m Model) renderVolumeWizardNameStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Enter a name for the volume:") + "\n\n")

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).
		Width(50)
	content.WriteString(inputStyle.Render(m.wizard.volumeNameInput+"▌") + "\n\n")

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	content.WriteString(helpStyle.Render("Type to enter • Enter: Continue • Esc: Cancel"))
	return content.String()
}

func (m Model) renderVolumeWizardRegionStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select region for the volume:") + "\n\n")

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)

	for i, region := range m.wizard.regions {
		name := getString(region, "name")
		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + name))
		} else {
			content.WriteString(listStyle.Render("  " + name))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter Select • Esc Cancel"))
	return content.String()
}

func (m Model) renderVolumeWizardTypeStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select volume type:") + "\n\n")

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)

	for i, vt := range m.wizard.volumeTypes {
		label := volumeTypeDisplayName(vt)
		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + label))
		} else {
			content.WriteString(listStyle.Render("  " + label))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter Select • ← Back • Esc Cancel"))
	return content.String()
}

func (m Model) renderVolumeWizardAZStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Select availability zone:") + "\n\n")

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	listStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	selectedStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#00FF7F")).Padding(0, 1)

	items := m.wizard.volumeAvailabilityZones
	// Prepend a "No preference" option
	allItems := append([]string{"(No preference)"}, items...)
	for i, az := range allItems {
		var label string
		if i == 0 {
			label = az
		} else {
			label = fmt.Sprintf("%s(%s)", m.wizard.selectedRegion, az)
		}
		if i == m.wizard.selectedIndex {
			content.WriteString(selectedStyle.Render("▶ " + label))
		} else {
			content.WriteString(listStyle.Render("  " + label))
		}
		content.WriteString("\n")
	}

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Margin(1, 0, 0, 0)
	content.WriteString(helpStyle.Render("↑↓ Navigate • Enter Select • ← Back • Esc Cancel"))
	return content.String()
}

func (m Model) renderVolumeWizardSizeStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Enter volume size (GB, min: 10, max: 12000):") + "\n\n")

	if m.wizard.errorMsg != "" {
		errStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
		content.WriteString(errStyle.Render("Error: "+m.wizard.errorMsg) + "\n\n")
	}

	inputStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF7F")).
		Padding(0, 1).
		Width(20)
	content.WriteString(inputStyle.Render(m.wizard.volumeSizeInput+"▌") + "\n\n")

	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	content.WriteString(helpStyle.Render("Type size • Enter: Continue • ←: Back • Esc: Cancel"))
	return content.String()
}

func volumeTypeDisplayName(name string) string {
	words := strings.Split(name, "-")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func (m Model) volumeTypeSupportLuks() bool {
	// Only classic, high-speed, high-speed-gen2 have -luks variants
	switch m.wizard.volumeType {
	case "classic", "high-speed", "high-speed-gen2":
		return true
	}
	return false
}

func (m Model) renderVolumeWizardEncryptionStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Encryption") + "\n\n")

	descStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	content.WriteString(descStyle.Render("Enable encryption to add an extra layer of security to your volumes\nand ensure the confidentiality of your data.") + "\n\n")

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F")).Bold(true)
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	disabledStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))

	luksSupported := m.volumeTypeSupportLuks()

	options := []struct {
		label    string
		disabled bool
	}{
		{"None", false},
		{"OVHcloud Managed Key", !luksSupported},
		{"Customer Managed Key  (Coming soon)", true},
	}

	for i, opt := range options {
		cursor := "  "
		if i == m.wizard.volumeEncryptionIdx && !opt.disabled {
			cursor = "▶ "
		}
		var line string
		if opt.disabled {
			line = disabledStyle.Render(cursor + opt.label)
		} else if i == m.wizard.volumeEncryptionIdx {
			line = selectedStyle.Render(cursor + opt.label)
		} else {
			line = normalStyle.Render(cursor + opt.label)
		}
		content.WriteString(line + "\n")
	}

	if !luksSupported {
		content.WriteString("\n" + disabledStyle.Render(fmt.Sprintf("  (Type '%s' does not support encryption)", m.wizard.volumeType)) + "\n")
	}

	content.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	content.WriteString(helpStyle.Render("↑↓: Navigate • Enter: Continue • ←: Back • Esc: Cancel"))
	return content.String()
}

func (m Model) handleVolumeWizardEncryptionKeys(key string) (tea.Model, tea.Cmd) {
	// If type doesn't support luks, force index to 0
	if !m.volumeTypeSupportLuks() {
		m.wizard.volumeEncryptionIdx = 0
	}
	switch key {
	case "up":
		if m.wizard.volumeEncryptionIdx > 0 {
			m.wizard.volumeEncryptionIdx--
		}
	case "down":
		if m.wizard.volumeEncryptionIdx < 1 && m.volumeTypeSupportLuks() {
			m.wizard.volumeEncryptionIdx++
		}
	case "enter":
		m.wizard.step = VolumeWizardStepConfirm
		m.wizard.volumeConfirmBtnIdx = 0
	case "left", "esc":
		m.wizard.step = VolumeWizardStepSize
	}
	return m, nil
}

func (m Model) renderVolumeWizardConfirmStep(width int) string {
	var content strings.Builder

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF"))
	content.WriteString(titleStyle.Render("Confirm volume creation:") + "\n\n")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	content.WriteString(labelStyle.Render("  Name:") + valueStyle.Render(m.wizard.volumeName) + "\n")
	content.WriteString(labelStyle.Render("  Region:") + valueStyle.Render(m.wizard.selectedRegion) + "\n")
	effectiveType := m.wizard.volumeType
	if m.wizard.volumeEncryptionIdx == 1 && !strings.HasSuffix(effectiveType, "-luks") {
		effectiveType += "-luks"
	}
	content.WriteString(labelStyle.Render("  Type:") + valueStyle.Render(effectiveType) + "\n")

	azDisplay := "(No preference)"
	if m.wizard.volumeAvailabilityZone != "" {
		azDisplay = fmt.Sprintf("%s(%s)", m.wizard.selectedRegion, m.wizard.volumeAvailabilityZone)
	}
	content.WriteString(labelStyle.Render("  Avail. Zone:") + valueStyle.Render(azDisplay) + "\n")
	content.WriteString(labelStyle.Render("  Size:") + valueStyle.Render(fmt.Sprintf("%d GB", m.wizard.volumeSize)) + "\n")
	encLabel := "None"
	if m.wizard.volumeEncryptionIdx == 1 {
		encLabel = "OVHcloud Managed Key (LUKS)"
	}
	content.WriteString(labelStyle.Render("  Encryption:") + valueStyle.Render(encLabel) + "\n")

	content.WriteString("\n")

	createStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	cancelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	if m.wizard.volumeConfirmBtnIdx == 0 {
		content.WriteString(createStyle.Render("  ▶ [Create Volume]") + "    ")
		content.WriteString(dimStyle.Render("[Cancel]") + "\n")
	} else {
		content.WriteString(dimStyle.Render("    [Create Volume]") + "    ")
		content.WriteString(cancelStyle.Render("▶ [Cancel]") + "\n")
	}

	return content.String()
}

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
	case ProductManagedDatabases:
		return "databases", fmt.Sprintf("ovhcloud cloud managed-database create --cloud-project %s", m.cloudProject)
	case ProductManagedAnalytics:
		return "analytics", fmt.Sprintf("ovhcloud cloud managed-analytics create --cloud-project %s", m.cloudProject)
	case ProductStorageBlock:
		return "block storage volumes", ""
	case ProductStorageFile:
		return "file shares", ""
	case ProductStorageObject:
		return "object storage containers", ""
	case ProductNetworks, ProductNetworkPrivate:
		return "private networks", fmt.Sprintf("ovhcloud cloud network private create --cloud-project %s", m.cloudProject)
	case ProductNetworkPublic:
		return "floating IPs", ""
	case ProductNetworkGateway:
		return "gateways", ""
	case ProductNetworkLB:
		return "load balancers", ""
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

	// For sub-nav products (Storage / Network), hide the cursor highlight unless
	// the user has entered table focus (Level 3 via Enter / ↓).
	isSubNavProd := m.currentProduct >= ProductStorageBlock && m.currentProduct <= ProductNetworkLB
	if isSubNavProd && !m.inTableFocus {
		s := table.DefaultStyles()
		s.Header = s.Header.
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240")).
			BorderBottom(true).
			Bold(true)
		s.Selected = lipgloss.NewStyle() // no highlight
		m.table.SetStyles(s)
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

// renderObjectStorageWithTabs renders the Object Storage view with tabs
func (m Model) renderObjectStorageWithTabs(tableContent string, width int) string {
	var content strings.Builder

	// Render tabs
	tabActiveStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7B68EE")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 2)

	tabInactiveStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 2)

	tab1 := "My Containers"
	tab2 := "Users"

	var tab1Rendered, tab2Rendered string
	if m.objectStorageTabIdx == 0 {
		tab1Rendered = tabActiveStyle.Render(tab1)
		tab2Rendered = tabInactiveStyle.Render(tab2)
	} else {
		tab1Rendered = tabInactiveStyle.Render(tab1)
		tab2Rendered = tabActiveStyle.Render(tab2)
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, tab1Rendered, "  ", tab2Rendered)
	content.WriteString(tabs + "\n\n")

	// Add the table content
	content.WriteString(tableContent)

	return content.String()
}

func (m Model) renderPrivateNetworksWithTabs(tableContent string, width int) string {
	return tableContent
}

func (m Model) renderPublicIPsWithTabs(tableContent string, width int) string {
	return tableContent
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
	case ProductStorageBlock:
		if m.volumeDetailView != nil {
			return m.volumeDetailView.Render(width, 0)
		}
		return m.renderGenericDetail(width)
	case ProductStorageSnapshot:
		if m.snapshotDetailView != nil {
			return m.snapshotDetailView.Render(width, 0)
		}
		return m.renderGenericDetail(width)
	case ProductStorageBackup:
		if m.backupDetailView != nil {
			return m.backupDetailView.Render(width, 0)
		}
		return m.renderGenericDetail(width)
        case ProductStorageFile:
                if m.fileShareDetailView != nil {
                        return m.fileShareDetailView.Render(width, 0)
                }
                return m.renderGenericDetail(width)
	case ProductNetworkPrivate:
		return m.renderPrivateNetworkDetail(width)
	case ProductManagedDatabases:
		return m.renderManagedDatabaseDetail(width)
	case ProductManagedAnalytics:
		return m.renderManagedDatabaseDetail(width)
	case ProductNetworkGateway:
		return m.renderGatewayDetail(width)
	case ProductNetworkLB:
		return m.renderLBDetail(width)
	case ProductNetworkPublic:
		return m.renderFIPDetail(width)
        case ProductStorageObject:
                if m.objectUserDetailView != nil {
                        return m.objectUserDetailView.Render(width, 0)
                }
                if m.objectDetailView != nil {
                        return m.objectDetailView.Render(width, 0)
                }
                return m.renderGenericDetail(width)
	case ProductWorkflow:
		return m.renderWorkflowDetail(width)
	case ProductInstanceBackup:
		return m.renderInstanceBackupDetail(width)
        default:
                return m.renderGenericDetail(width)
        }
}

func (m Model) renderWorkflowDetail(width int) string {
	var content strings.Builder

	name := getStringValue(m.detailData, "name", "N/A")
	id := getStringValue(m.detailData, "id", "N/A")
	region := getStringValue(m.detailData, "region", "N/A")
	instanceId := getStringValue(m.detailData, "instanceId", "N/A")
	cron := getStringValue(m.detailData, "cron", "N/A")
	rotation := int(getFloatValue(m.detailData, "rotation", 0))
	lastStatus := getStringValue(m.detailData, "lastExecutionStatus", "-")

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	boxWidth := width - 4

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	statusIcon := "⏳"
	switch strings.ToLower(lastStatus) {
	case "success":
		statusIcon = "✅"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	case "error", "failed":
		statusIcon = "❌"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	case "running":
		statusIcon = "🔄"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}

	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(id, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Region"), valueSt.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Instance"), valueSt.Render(truncate(instanceId, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Cron"), valueSt.Render(cron)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Rotation"), valueSt.Render(fmt.Sprintf("%d sauvegardes", rotation))))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Statut"), statusStyle.Render(statusIcon+" "+lastStatus)))
	infoBox := renderBox("Workflow : "+name, infoContent.String(), boxWidth)

	actions := []string{"Supprimer"}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color("#FF6B6B")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.actionConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Esc to cancel", actions[m.selectedAction]))
	}
	actionsBox := renderBox("Actions (Enter to execute, Esc to go back)", actionsContent, width-4)

	content.WriteString(actionsBox + "\n\n")
	content.WriteString(infoBox)
	return content.String()
}

func (m Model) renderInstanceBackupDetail(width int) string {
	var content strings.Builder

	name := getStringValue(m.detailData, "name", "N/A")
	id := getStringValue(m.detailData, "id", "N/A")
	status := getStringValue(m.detailData, "status", "N/A")
	created := getStringValue(m.detailData, "creationDate", "-")
	if len(created) >= 16 {
		created = created[:16]
	}
	minDisk := int(getFloatValue(m.detailData, "minDisk", 0))
	sizeStr := "-"
	if minDisk > 0 {
		sizeStr = fmt.Sprintf("%d GB", minDisk)
	}
	location := getStringValue(m.detailData, "region", "")
	if location == "" {
		if regions, ok := m.detailData["regions"].([]interface{}); ok && len(regions) > 0 {
			var rnames []string
			for _, r := range regions {
				if rs, ok := r.(string); ok {
					rnames = append(rnames, rs)
				}
			}
			location = strings.Join(rnames, ", ")
		}
	}
	if location == "" {
		location = "-"
	}

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	boxWidth := width - 4

	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	statusIcon := "🟡"
	switch strings.ToLower(status) {
	case "active":
		statusIcon = "🟢"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	case "error":
		statusIcon = "🔴"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B"))
	}

	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(id, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Localisation"), valueSt.Render(location)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Taille disque"), valueSt.Render(sizeStr)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Created"), valueSt.Render(created)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Statut"), statusStyle.Render(statusIcon+" "+status)))
	infoBox := renderBox("Instance Backup : "+name, infoContent.String(), boxWidth)

	actions := []string{"Supprimer"}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color("#FF4444")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.actionConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Esc to cancel", actions[m.selectedAction]))
	}
	actionsBox := renderBox("Actions (Enter to execute, Esc to go back)", actionsContent, width-4)

	content.WriteString(actionsBox + "\n\n")
	content.WriteString(infoBox)
	return content.String()
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

	infoBox := renderBox("Information", infoContent.String(), boxWidth)

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

	networkBox := renderBox("Network", networkContent.String(), boxWidth)

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
	actions := []string{"SSH", "Reboot", rescueAction, stopStartAction, "Console", "Reinstall", "Backup", "Delete"}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			// Selected action - highlighted (Delete uses red)
			bg := lipgloss.Color("#7B68EE")
			if action == "Delete" {
				bg = lipgloss.Color("#FF4444")
			}
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(bg).
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
	actionsBox := renderBox("Quick Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)

	// Combine everything
	content.WriteString(actionsBox)
	content.WriteString("\n\n")

	// Side by side boxes
	leftRight := lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", networkBox)
	content.WriteString(leftRight)

	return content.String()
}

func (m Model) renderManagedDatabaseDetail(width int) string {
	var content strings.Builder

	// ── Common data ───────────────────────────────────────────────────────
	dbID := getStringValue(m.detailData, "id", "N/A")
	dbName := getStringValue(m.detailData, "description", "")
	if dbName == "" {
		dbName = dbID
	}
	engineRaw := getStringValue(m.detailData, "engine", "N/A")
	engineDisplay := engineRaw
	version := getStringValue(m.detailData, "version", "")
	if version != "" {
		engineDisplay = engineRaw + " " + version
	}
	isPostgres := strings.EqualFold(engineRaw, "postgresql")

	// ── Tab bar ───────────────────────────────────────────────────────────
	tabActiveStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#7B68EE")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 2)
	tabInactiveStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#333333")).
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 2)
	tabDisabledStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#1a1a1a")).
		Foreground(lipgloss.Color("#444444")).
		Padding(0, 2)

	tabNames := []string{"Service", "Users", "Backups", "Databases", "Pools"}
	var tabParts []string
	for i, name := range tabNames {
		if i == 4 && !isPostgres {
			tabParts = append(tabParts, tabDisabledStyle.Render(name))
		} else if i == m.dbDetailTab {
			tabParts = append(tabParts, tabActiveStyle.Render(name))
		} else {
			tabParts = append(tabParts, tabInactiveStyle.Render(name))
		}
	}
	tabRow := lipgloss.JoinHorizontal(lipgloss.Top, tabParts...)
	tabHint := lipgloss.NewStyle().Foreground(lipgloss.Color("#444444")).Render("  ←/→ navigate tabs")
	content.WriteString(tabRow + tabHint + "\n\n")

	// ── Shared styles ─────────────────────────────────────────────────────
	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	dimSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	headSt := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#AAAAAA")).Width(18)
	fullWidth := width - 4

	switch m.dbDetailTab {
	case 0: // ── Service ──────────────────────────────────────────────────
		plan := getStringValue(m.detailData, "plan", "N/A")
		flavor := getStringValue(m.detailData, "flavor", "N/A")
		status := getStringValue(m.detailData, "status", "N/A")
		createdAt := getStringValue(m.detailData, "createdAt", "N/A")
		if len(createdAt) >= 10 {
			createdAt = createdAt[:10]
		}

		storageStr := "-"
		if storage, ok := m.detailData["storage"].(map[string]interface{}); ok {
			if size, ok := storage["size"].(map[string]interface{}); ok {
				val := getStringValue(size, "value", "")
				unit := getStringValue(size, "unit", "")
				if val != "" {
					storageStr = val + " " + unit
				}
			}
		}

		location := "-"
		nodesCount := 0
		if nodes, ok := m.detailData["nodes"].([]interface{}); ok {
			nodesCount = len(nodes)
			if len(nodes) > 0 {
				if node, ok := nodes[0].(map[string]interface{}); ok {
					if r := getStringValue(node, "region", ""); r != "" {
						location = r
					}
				}
			}
		}

		statusIcon := "🟢"
		statusStyle := statusRunningStyle
		switch strings.ToUpper(status) {
		case "CREATING", "UPDATING", "RESTARTING", "PENDING", "MAINTENANCE":
			statusIcon = "🟡"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
		case "ERROR", "ERROR_INCONSISTENT_SPEC", "DELETING", "SUSPENDED", "LOCKED":
			statusIcon = "🔴"
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444"))
		}

		boxWidth := (width - 6) / 2
		if boxWidth < 35 {
			boxWidth = 35
		}

		var infoContent strings.Builder
		infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
		infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(dbID, 36))))
		infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Engine"), valueSt.Render(engineDisplay)))
		infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Plan"), valueSt.Render(plan)))
		infoBox := renderBox("Database "+dbName, infoContent.String(), boxWidth)

		var cfgContent strings.Builder
		cfgContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Flavor"), valueSt.Render(flavor)))
		cfgContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Storage"), valueSt.Render(storageStr)))
		cfgContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Location"), valueSt.Render(location)))
		cfgContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Nodes"), valueSt.Render(fmt.Sprintf("%d", nodesCount))))
		cfgContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Created"), valueSt.Render(createdAt)))
		cfgBox := renderBox("Configuration", cfgContent.String(), boxWidth)

		actions := []string{"Delete"}
		var actionParts []string
		for i, action := range actions {
			if i == m.selectedAction {
				actionParts = append(actionParts, lipgloss.NewStyle().
					Background(lipgloss.Color("#7B68EE")).
					Foreground(lipgloss.Color("#FFFFFF")).
					Bold(true).Padding(0, 1).Render(action))
			} else {
				actionParts = append(actionParts, lipgloss.NewStyle().
					Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
			}
		}
		actionsContent := strings.Join(actionParts, " ")
		if m.actionConfirm {
			actionsContent += "\n\n" + lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFD700")).Bold(true).
				Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actions[m.selectedAction]))
		}
		actionsBox := renderBox("Actions (Enter to execute)", actionsContent, width-4)

		content.WriteString(actionsBox + "\n\n")
		content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", cfgBox))

	case 1: // ── Users ─────────────────────────────────────────────────────
		if m.dbUserCreatedData != nil {
			// ── User creation result panel ────────────────────────────────
			createdUsername := getStringValue(m.dbUserCreatedData, "username", getStringValue(m.dbUserCreatedData, "name", "—"))
			password := getStringValue(m.dbUserCreatedData, "password", "—")

			// Build URIs from the service's endpoints
			host, port, dbname, scheme, sslMode := "", "", "defaultdb", engineRaw, "require"
			if endpoints, ok := m.detailData["endpoints"].([]interface{}); ok {
				for _, ep := range endpoints {
					epMap, ok := ep.(map[string]interface{})
					if !ok {
						continue
					}
					if !strings.EqualFold(getStringValue(epMap, "component", ""), engineRaw) {
						continue
					}
					host = getStringValue(epMap, "domain", "")
					if p, ok := toFloat64(epMap["port"]); ok {
						port = fmt.Sprintf("%d", int(p))
					}
					if sch := getStringValue(epMap, "scheme", ""); sch != "" {
						scheme = sch
					}
					if sm := getStringValue(epMap, "sslMode", ""); sm != "" {
						sslMode = sm
					}
					if pa := getStringValue(epMap, "path", ""); pa != "" {
						dbname = strings.TrimPrefix(pa, "/")
					}
					break
				}
			}

			highlightSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true)
			uriSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF"))
			connSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#98FB98"))
			warnSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF8C00")).Bold(true)

			var panel strings.Builder
			panel.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Username"), valueSt.Render(createdUsername)))
			panel.WriteString(fmt.Sprintf("%s %s\n\n", labelSt.Render("Password"), highlightSt.Render(password)))

			if host != "" {
				quickURI := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=%s",
					scheme, createdUsername, password, host, port, dbname, sslMode)
				connStr := fmt.Sprintf("host=%s port=%s dbname=%s user=%s password=%s sslmode=%s",
					host, port, dbname, createdUsername, password, sslMode)

				panel.WriteString(labelSt.Render("Connection URI") + "\n")
				panel.WriteString(uriSt.Render(quickURI) + "\n\n")
				panel.WriteString(labelSt.Render("Connection String") + "\n")
				panel.WriteString(connSt.Render(connStr) + "\n\n")
			}

			panel.WriteString(warnSt.Render("⚠  Save your password now — it will not be shown again.") + "\n")
			panel.WriteString(dimSt.Render("   Press Enter or Esc to dismiss"))
			content.WriteString(renderBox("✅ User Created: "+createdUsername, panel.String(), fullWidth))
		} else if m.dbUserCreateMode {
			// ── User creation text input ──────────────────────────────────
			inputSt := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF")).
				Background(lipgloss.Color("#2a2a2a")).
				Padding(0, 1)
			var inputContent strings.Builder
			inputContent.WriteString(labelSt.Render("Username") + "\n\n")
			inputContent.WriteString(inputSt.Render(m.dbUserCreateInput+"▌") + "\n\n")
			inputContent.WriteString(dimSt.Render("Enter → confirm   Esc → cancel"))
			content.WriteString(renderBox("Create New User", inputContent.String(), fullWidth))
		} else {
			// ── Users list with Create action ─────────────────────────────
			createBtn := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7B68EE")).
				Bold(true).Padding(0, 1).Render("[+ Create User]")
			content.WriteString(createBtn + "  " + dimSt.Render("Enter → new user") + "\n\n")

			var usersContent strings.Builder
			if len(m.dbDetailUsers) == 0 {
				if m.dbDetailLoaded {
					usersContent.WriteString(dimSt.Render("None"))
				} else {
					usersContent.WriteString(dimSt.Render("Loading..."))
				}
			} else {
				usersContent.WriteString(fmt.Sprintf("%s%s%s%s\n",
					headSt.Render("Username"),
					headSt.Render("Roles"),
					headSt.Render("Created"),
					headSt.Render("Status")))
				for _, u := range m.dbDetailUsers {
					username := getStringValue(u, "username", getStringValue(u, "name", "—"))
					created := getStringValue(u, "createdAt", "—")
					if len(created) >= 10 {
						created = created[:10]
					}
					userStatus := getStringValue(u, "status", "—")
					var roles []string
					if rawRoles, ok := u["roles"].([]interface{}); ok {
						for _, r := range rawRoles {
							if s, ok := r.(string); ok {
								roles = append(roles, s)
							}
						}
					}
					rolesStr := strings.Join(roles, ", ")
					if rolesStr == "" {
						rolesStr = "—"
					}
					usersContent.WriteString(fmt.Sprintf("%s%s%s%s\n",
						labelSt.Render(truncate(username, 17)),
						labelSt.Render(truncate(rolesStr, 17)),
						labelSt.Render(created),
						valueSt.Render(userStatus)))
				}
			}
			content.WriteString(renderBox(fmt.Sprintf("Users (%d)", len(m.dbDetailUsers)), usersContent.String(), fullWidth))
		}

	case 2: // ── Backups ───────────────────────────────────────────────────
		var backupsContent strings.Builder
		if len(m.dbDetailBackups) == 0 {
			if m.dbDetailLoaded {
				backupsContent.WriteString(dimSt.Render("None"))
			} else {
				backupsContent.WriteString(dimSt.Render("Loading..."))
			}
		} else {
			backupsContent.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
				headSt.Render("Name"),
				headSt.Render("Location"),
				headSt.Render("Created"),
				headSt.Render("Expires"),
				headSt.Render("Status")))
			for _, b := range m.dbDetailBackups {
				bname := getStringValue(b, "name", getStringValue(b, "id", "—"))
				bloc := getStringValue(b, "region", "—")
				bcreated := getStringValue(b, "createdAt", "—")
				if len(bcreated) >= 10 {
					bcreated = bcreated[:10]
				}
				bexpires := getStringValue(b, "expiresAt", "—")
				if len(bexpires) >= 10 {
					bexpires = bexpires[:10]
				}
				bstatus := getStringValue(b, "status", "—")
				backupsContent.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
					labelSt.Render(truncate(bname, 17)),
					labelSt.Render(truncate(bloc, 17)),
					labelSt.Render(bcreated),
					labelSt.Render(bexpires),
					valueSt.Render(bstatus)))
			}
		}
		content.WriteString(renderBox(fmt.Sprintf("Backups (%d)", len(m.dbDetailBackups)), backupsContent.String(), fullWidth))

	case 3: // ── Databases ─────────────────────────────────────────────────
		var dbNamesContent strings.Builder
		if len(m.dbDetailDatabases) == 0 {
			if m.dbDetailLoaded {
				dbNamesContent.WriteString(dimSt.Render("None"))
			} else {
				dbNamesContent.WriteString(dimSt.Render("Loading..."))
			}
		} else {
			for _, d := range m.dbDetailDatabases {
				dbname := getStringValue(d, "name", getStringValue(d, "id", "—"))
				dbNamesContent.WriteString(valueSt.Render("  • "+dbname) + "\n")
			}
		}
		content.WriteString(renderBox(fmt.Sprintf("Databases (%d)", len(m.dbDetailDatabases)), strings.TrimRight(dbNamesContent.String(), "\n"), fullWidth))

	case 4: // ── Pools ─────────────────────────────────────────────────────
		if !isPostgres {
			content.WriteString(renderBox("Connection Pools", dimSt.Render("Connection pools are only available for PostgreSQL."), fullWidth))
		} else {
			var poolsContent strings.Builder
			if len(m.dbDetailPools) == 0 {
				if m.dbDetailLoaded {
					poolsContent.WriteString(dimSt.Render("None"))
				} else {
					poolsContent.WriteString(dimSt.Render("Loading..."))
				}
			} else {
				poolsContent.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
					headSt.Render("Name"),
					headSt.Render("Database"),
					headSt.Render("Mode"),
					headSt.Render("Size"),
					headSt.Render("Username")))
				for _, p := range m.dbDetailPools {
					pname := getStringValue(p, "name", "—")
					pdb := getStringValue(p, "databaseId", getStringValue(p, "database", "—"))
					pmode := getStringValue(p, "mode", "—")
					var psize string
					if v, ok := toFloat64(p["size"]); ok && v > 0 {
						psize = fmt.Sprintf("%d", int(v))
					} else {
						psize = "—"
					}
					puser := getStringValue(p, "userId", getStringValue(p, "username", "—"))
					poolsContent.WriteString(fmt.Sprintf("%s%s%s%s%s\n",
						labelSt.Render(truncate(pname, 17)),
						labelSt.Render(truncate(pdb, 17)),
						labelSt.Render(truncate(pmode, 17)),
						labelSt.Render(psize),
						valueSt.Render(truncate(puser, 17))))
				}
			}
			content.WriteString(renderBox(fmt.Sprintf("Connection Pools (%d)", len(m.dbDetailPools)), strings.TrimRight(poolsContent.String(), "\n"), fullWidth))
		}
	}

	return content.String()
}

func (m Model) renderGatewayDetail_PLACEHOLDER() string { return "" }


func (m Model) renderGatewayDetail(width int) string {
	var content strings.Builder

	gwID := getStringValue(m.detailData, "id", "N/A")
	gwName := getStringValue(m.detailData, "name", "Unknown")
	region := getStringValue(m.detailData, "region", "N/A")
	model := getStringValue(m.detailData, "model", "N/A")
	status := getStringValue(m.detailData, "status", "N/A")

	publicIP := "N/A"
	if ei, ok := m.detailData["externalInformation"].(map[string]interface{}); ok {
		if ips, ok := ei["ips"].([]interface{}); ok && len(ips) > 0 {
			if ipm, ok := ips[0].(map[string]interface{}); ok {
				if v := getStringValue(ipm, "ip", ""); v != "" {
					publicIP = v
				}
			}
		}
	}

	var privateIPs []string
	privateNetwork := "N/A"
	if ifaces, ok := m.detailData["interfaces"].([]interface{}); ok {
		for _, iface := range ifaces {
			if ifm, ok := iface.(map[string]interface{}); ok {
				if v := getStringValue(ifm, "ip", ""); v != "" {
					privateIPs = append(privateIPs, v)
				}
				if privateNetwork == "N/A" {
					if v := getStringValue(ifm, "networkId", ""); v != "" {
						privateNetwork = v
					}
				}
			}
		}
	}
	privateIP := "N/A"
	if len(privateIPs) > 0 {
		privateIP = strings.Join(privateIPs, ", ")
	}

	statusIcon := "🟢"
	statusStyle := statusRunningStyle
	if strings.ToLower(status) != "active" {
		statusIcon = "🟡"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Info box
	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(gwID, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Region"), valueSt.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Model"), valueSt.Render(model)))
	infoBox := renderBox("Gateway "+gwName, infoContent.String(), boxWidth)

	// Network box
	var netContent strings.Builder
	netContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Public IP"), valueSt.Render(publicIP)))
	netContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Private IP"), valueSt.Render(truncate(privateIP, 36))))
	netContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Private Network"), valueSt.Render(truncate(privateNetwork, 36))))
	netBox := renderBox("Network", netContent.String(), boxWidth)

	// Actions
	actions := []string{"Delete"}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color("#7B68EE")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.actionConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actions[m.selectedAction]))
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)

	content.WriteString(actionsBox + "\n\n")
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", netBox))
	return content.String()
}

func (m Model) renderLBDetail(width int) string {
	var content strings.Builder

	lbID := getStringValue(m.detailData, "id", "N/A")
	lbName := getStringValue(m.detailData, "name", "Unknown")
	region := getStringValue(m.detailData, "region", "N/A")

	size := getStringValue(m.detailData, "_flavorName", "")
	if size == "" {
		size = getStringValue(m.detailData, "flavorId", "N/A")
	}

	provisioning := getStringValue(m.detailData, "provisioningStatus", "N/A")
	operating := getStringValue(m.detailData, "operatingStatus", "N/A")
	privateIP := getStringValue(m.detailData, "vipAddress", "N/A")

	privateNetwork := getStringValue(m.detailData, "_networkName", "")
	if privateNetwork == "" {
		privateNetwork = getStringValue(m.detailData, "vipNetworkId", "N/A")
	}

	publicIP := "-"
	if fi, ok := m.detailData["floatingIp"].(map[string]interface{}); ok {
		if v := getStringValue(fi, "ip", ""); v != "" {
			publicIP = v
		}
	}

	statusIcon := "🟢"
	statusStyle := statusRunningStyle
	if strings.ToLower(operating) != "online" && strings.ToLower(operating) != "no_monitor" {
		statusIcon = "🟡"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Info box
	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Operating Status"), statusStyle.Render(statusIcon+" "+operating)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Supply Status"), valueSt.Render(provisioning)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(lbID, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Region"), valueSt.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Size"), valueSt.Render(strings.ToUpper(size))))
	infoBox := renderBox("Load Balancer "+lbName, infoContent.String(), boxWidth)

	// Network box
	var netContent strings.Builder
	netContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Public IP"), valueSt.Render(publicIP)))
	netContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Private IP (VIP)"), valueSt.Render(privateIP)))
	netContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Private Network"), valueSt.Render(truncate(privateNetwork, 36))))
	netBox := renderBox("Network", netContent.String(), boxWidth)

	// Actions: Delete / Listeners / Pools
	actions := []string{"Delete", "Listeners", "Pools"}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			bg := lipgloss.Color("#7B68EE")
			if action == "Delete" {
				bg = lipgloss.Color("#FF4444")
			} else if action == "Pools" {
				bg = lipgloss.Color("#00AA55")
			}
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(bg).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.actionConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actions[m.selectedAction]))
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)

	// Pools section
	var poolsContent strings.Builder
	lbPools := m.lbPools[lbID]
	if len(lbPools) == 0 {
		poolsContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("  No pools. Press «Pools» to create one."))
	} else {
		colName := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(26).Render("Name")
		colAlgo := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20).Render("Algorithm")
		colProto := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(12).Render("Protocol")
		colStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(16).Render("Status")
		poolsContent.WriteString(colName + colAlgo + colProto + colStatus + "\n")
		poolsContent.WriteString(strings.Repeat("─", min(width-8, 74)) + "\n")
		for i, p := range lbPools {
			pName := truncate(getStringValue(p, "name", "-"), 24)
			pAlgo := getStringValue(p, "algorithm", "-")
			pProto := getStringValue(p, "protocol", "-")
			pStatus := getStringValue(p, "operatingStatus", getStringValue(p, "provisioningStatus", "-"))
			statusColor := lipgloss.Color("#00FF7F")
			if strings.ToLower(pStatus) != "online" && strings.ToLower(pStatus) != "active" {
				if strings.ToLower(pStatus) == "error" {
					statusColor = lipgloss.Color("#FF6B6B")
				} else {
					statusColor = lipgloss.Color("#FFD700")
				}
			}
			cursor := "  "
			nameColor := lipgloss.Color("#FFFFFF")
			if i == m.lbPoolListIdx {
				cursor = "▶ "
				nameColor = lipgloss.Color("#00FF7F")
			}
			poolsContent.WriteString(
				cursor +
					lipgloss.NewStyle().Foreground(nameColor).Width(24).Render(pName) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(20).Render(pAlgo) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(12).Render(pProto) +
					lipgloss.NewStyle().Foreground(statusColor).Width(16).Render(pStatus) + "\n",
			)
		}
	}
	poolsFooter := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("\u2191\u2193: Navigate \u2022 Enter: Open")
	poolsBox := renderBox(fmt.Sprintf("Pools (%d)", len(lbPools)), poolsContent.String()+"\n"+poolsFooter, width-4)

	// Listeners section
	lbListeners := m.lbListeners[lbID]
	var listenersContent strings.Builder
	if len(lbListeners) == 0 {
		listenersContent.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("  No listeners. Press «Listeners» to create one."))
	} else {
		hName := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(26).Render("Name")
		hPool := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22).Render("Default pool")
		hProto := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18).Render("Protocol")
		hPort := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(8).Render("Port")
		hStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(14).Render("Status")
		listenersContent.WriteString(hName + hPool + hProto + hPort + hStatus + "\n")
		listenersContent.WriteString(strings.Repeat("─", min(width-8, 88)) + "\n")
		for i, l := range lbListeners {
			lName := truncate(getStringValue(l, "name", "-"), 24)
			lPoolId := getStringValue(l, "defaultPoolId", "")
			lPoolLabel := "-"
			if lPoolId != "" {
				for _, p := range lbPools {
					if getStringValue(p, "id", "") == lPoolId {
						lPoolLabel = truncate(getStringValue(p, "name", lPoolId), 20)
						break
					}
				}
				if lPoolLabel == "-" {
					lPoolLabel = truncate(lPoolId, 8)
				}
			}
			lProto := getStringValue(l, "protocol", "-")
			lPort := fmt.Sprintf("%v", l["port"])
			lStatus := getStringValue(l, "operatingStatus", getStringValue(l, "provisioningStatus", "-"))
			statusColor := lipgloss.Color("#00FF7F")
			if strings.ToLower(lStatus) != "online" {
				if strings.ToLower(lStatus) == "error" {
					statusColor = lipgloss.Color("#FF6B6B")
				} else {
					statusColor = lipgloss.Color("#FFD700")
				}
			}
			cursor := "  "
			nameColor := lipgloss.Color("#FFFFFF")
			if i == m.lbListenerListIdx {
				cursor = "▶ "
				nameColor = lipgloss.Color("#00FF7F")
			}
			listenersContent.WriteString(
				cursor +
					lipgloss.NewStyle().Foreground(nameColor).Width(24).Render(lName) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(22).Render(lPoolLabel) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(18).Render(lProto) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(8).Render(lPort) +
					lipgloss.NewStyle().Foreground(statusColor).Width(14).Render(lStatus) + "\n",
			)
		}
	}
	listenersFooter := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("\u2191\u2193: Navigate \u2022 Enter: Open")
	listenersBox := renderBox(fmt.Sprintf("Listeners (%d)", len(lbListeners)), listenersContent.String()+"\n"+listenersFooter, width-4)

	content.WriteString(actionsBox + "\n\n")
	content.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, infoBox, "  ", netBox))
	content.WriteString("\n\n" + listenersBox)
	content.WriteString("\n\n" + poolsBox)
	return content.String()
}

// renderLBPoolDetailView displays detailed information about a single LB pool.
func (m Model) renderLBPoolDetailView(width int) string {
	var content strings.Builder

	if m.selectedLBPool == nil {
		return "No pool selected"
	}

	poolID := getStringValue(m.selectedLBPool, "id", "N/A")
	poolName := getStringValue(m.selectedLBPool, "name", "N/A")
	algo := getStringValue(m.selectedLBPool, "algorithm", "N/A")
	proto := getStringValue(m.selectedLBPool, "protocol", "N/A")
	operating := getStringValue(m.selectedLBPool, "operatingStatus", "N/A")
	provisioning := getStringValue(m.selectedLBPool, "provisioningStatus", "N/A")
	lbName := getStringValue(m.detailData, "name", "N/A")
	region := getStringValue(m.detailData, "region", "N/A")

	sessionType := "-"
	if sp, ok := m.selectedLBPool["sessionPersistence"].(map[string]interface{}); ok {
		sessionType = getStringValue(sp, "type", "-")
	}

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Actions: Edit / Delete / Members
	actions := []string{"Edit", "Delete", "Members"}
	var actionParts []string
	for i, action := range actions {
		if i == m.lbPoolDetailActionIdx {
			bg := lipgloss.Color("#7B68EE")
			if action == "Delete" {
				bg = lipgloss.Color("#FF4444")
			} else if action == "Members" {
				bg = lipgloss.Color("#00AA55")
			}
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(bg).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 1).Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.lbPoolDetailConfirm && m.lbPoolDetailActionIdx != 2 {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("\u26a0\ufe0f  Press Enter to confirm %s, Esc to cancel", actions[m.lbPoolDetailActionIdx]))
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)
	content.WriteString(actionsBox + "\n\n")

	// Pool info
	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(poolID, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Load Balancer"), valueSt.Render(lbName)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Region"), valueSt.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Algorithm"), valueSt.Render(algo)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Protocol"), valueSt.Render(proto)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Session"), valueSt.Render(sessionType)))

	statusColor := lipgloss.Color("#00FF7F")
	if strings.ToLower(operating) != "online" {
		statusColor = lipgloss.Color("#FFD700")
		if strings.ToLower(operating) == "error" {
			statusColor = lipgloss.Color("#FF6B6B")
		}
	}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Operating Status"),
		lipgloss.NewStyle().Foreground(statusColor).Render(operating)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Supply Status"), valueSt.Render(provisioning)))

	infoBox := renderBox("Pool "+poolName, infoContent.String(), width-4)
	content.WriteString(infoBox)

	return content.String()
}

func (m Model) renderLBListenerDetailView(width int) string {
	var content strings.Builder

	if m.selectedLBListener == nil {
		return "No listener selected"
	}

	listenerID := getStringValue(m.selectedLBListener, "id", "N/A")
	listenerName := getStringValue(m.selectedLBListener, "name", "N/A")
	proto := getStringValue(m.selectedLBListener, "protocol", "N/A")
	port := fmt.Sprintf("%v", m.selectedLBListener["port"])
	operating := getStringValue(m.selectedLBListener, "operatingStatus", "N/A")
	provisioning := getStringValue(m.selectedLBListener, "provisioningStatus", "N/A")
	lbName := getStringValue(m.detailData, "name", "N/A")
	region := getStringValue(m.detailData, "region", "N/A")

	defaultPoolId := getStringValue(m.selectedLBListener, "defaultPoolId", "")
	defaultPoolName := defaultPoolId
	if defaultPoolId != "" {
		lbID := getStringValue(m.detailData, "id", "")
		if pools, ok := m.lbPools[lbID]; ok {
			for _, pool := range pools {
				if getStringValue(pool, "id", "") == defaultPoolId {
					defaultPoolName = getStringValue(pool, "name", defaultPoolId)
					break
				}
			}
		}
	}
	if defaultPoolName == "" {
		defaultPoolName = "-"
	}

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Actions: Edit / Delete / L7 Policies
	actions := []string{"Edit", "Delete", "L7 Policies"}
	var actionParts []string
	for i, action := range actions {
		if i == m.lbListenerDetailActionIdx {
			bg := lipgloss.Color("#7B68EE")
			if action == "Delete" {
				bg = lipgloss.Color("#FF4444")
			} else if action == "L7 Policies" {
				bg = lipgloss.Color("#00AA55")
			}
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(bg).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 1).Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.lbListenerDetailConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("\u26a0\ufe0f  Press Enter to confirm %s, Esc to cancel", actions[m.lbListenerDetailActionIdx]))
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)
	content.WriteString(actionsBox + "\n\n")

	// Listener info
	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(listenerID, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Load Balancer"), valueSt.Render(lbName)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Region"), valueSt.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Protocol"), valueSt.Render(proto)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Port"), valueSt.Render(port)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Default pool"), valueSt.Render(defaultPoolName)))

	statusColor := lipgloss.Color("#00FF7F")
	if strings.ToLower(operating) != "online" {
		statusColor = lipgloss.Color("#FFD700")
		if strings.ToLower(operating) == "error" {
			statusColor = lipgloss.Color("#FF6B6B")
		}
	}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Operating Status"),
		lipgloss.NewStyle().Foreground(statusColor).Render(operating)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Supply Status"), valueSt.Render(provisioning)))

	infoBox := renderBox("Listener "+listenerName, infoContent.String(), width-4)
	content.WriteString(infoBox)

	// L7 Policies section
	listenerID2 := getStringValue(m.selectedLBListener, "id", "")
	l7Policies := m.lbL7Policies[listenerID2]
	var l7Content strings.Builder
	if len(l7Policies) == 0 {
		l7Content.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).
			Render("  No L7 Policies. Select «L7 Policies» to create one."))
	} else {
		hName := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(28).Render("  Name")
		hPos := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(12).Render("Position")
		hAction := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22).Render("Action")
		hStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(16).Render("Status")
		l7Content.WriteString(hName + hPos + hAction + hStatus + "\n")
		l7Content.WriteString(strings.Repeat("─", min(width-8, 78)) + "\n")
		for i, p := range l7Policies {
			pName := truncate(getStringValue(p, "name", "-"), 24)
			pPos := fmt.Sprintf("%v", p["position"])
			pAction := getStringValue(p, "action", "-")
			pStatus := getStringValue(p, "provisioningStatus", getStringValue(p, "operatingStatus", "-"))
			statusColor := lipgloss.Color("#00FF7F")
			if strings.ToLower(pStatus) != "active" && strings.ToLower(pStatus) != "online" {
				if strings.ToLower(pStatus) == "error" {
					statusColor = lipgloss.Color("#FF6B6B")
				} else {
					statusColor = lipgloss.Color("#FFD700")
				}
			}
			cursor := "  "
			nameStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Width(26)
			if i == m.lbL7PolicyListIdx {
				cursor = "▶ "
				nameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7B68EE")).Bold(true).Width(26)
			}
			l7Content.WriteString(
				cursor +
					nameStyle.Render(pName) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(12).Render(pPos) +
					lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC")).Width(22).Render(pAction) +
					lipgloss.NewStyle().Foreground(statusColor).Width(16).Render(pStatus) + "\n",
			)
		}
	}
	l7Footer := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  ↑↓: Navigate • Enter: Open details")
	l7Box := renderBox(fmt.Sprintf("L7 Policies (%d)", len(l7Policies)), l7Content.String()+"\n"+l7Footer, width-4)
	content.WriteString("\n\n" + l7Box)

	return content.String()
}

func (m Model) renderLBL7PolicyDetailView(width int) string {
	var content strings.Builder

	if m.selectedLBL7Policy == nil {
		return "No L7 policy selected"
	}

	policyID := getStringValue(m.selectedLBL7Policy, "id", "N/A")
	policyName := getStringValue(m.selectedLBL7Policy, "name", "N/A")
	action := getStringValue(m.selectedLBL7Policy, "action", "N/A")
	position := fmt.Sprintf("%v", m.selectedLBL7Policy["position"])
	operating := getStringValue(m.selectedLBL7Policy, "operatingStatus", "N/A")
	provisioning := getStringValue(m.selectedLBL7Policy, "provisioningStatus", "N/A")
	listenerName := "-"
	if m.selectedLBListener != nil {
		listenerName = getStringValue(m.selectedLBListener, "name", "-")
	}
	region := getStringValue(m.detailData, "region", "N/A")

	// Redirect target
	redirectTarget := "-"
	if v := getStringValue(m.selectedLBL7Policy, "redirectUrl", ""); v != "" {
		redirectTarget = v
	} else if v := getStringValue(m.selectedLBL7Policy, "redirectPrefix", ""); v != "" {
		redirectTarget = "[prefix] " + v
	} else if poolID := getStringValue(m.selectedLBL7Policy, "redirectPoolId", ""); poolID != "" {
		redirectTarget = poolID
		lbID := getStringValue(m.detailData, "id", "")
		if pools, ok := m.lbPools[lbID]; ok {
			for _, p := range pools {
				if getStringValue(p, "id", "") == poolID {
					redirectTarget = getStringValue(p, "name", poolID)
					break
				}
			}
		}
	}

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Actions: Edit / Delete / L7 Rules
	actions := []string{"Edit", "Delete", "L7 Rules"}
	var actionParts []string
	for i, act := range actions {
		if i == m.lbL7PolicyDetailActionIdx {
			bg := lipgloss.Color("#7B68EE")
			if act == "Delete" {
				bg = lipgloss.Color("#FF4444")
			} else if act == "L7 Rules" {
				bg = lipgloss.Color("#00AA55")
			}
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(bg).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 1).Render(act))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+act+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.lbL7PolicyDetailConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Esc to cancel", actions[m.lbL7PolicyDetailActionIdx]))
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)
	content.WriteString(actionsBox + "\n\n")

	// Policy info
	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(policyID, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Listener"), valueSt.Render(listenerName)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Region"), valueSt.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Action"), valueSt.Render(action)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Position"), valueSt.Render(position)))
	if redirectTarget != "-" {
		infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Redirect Target"), valueSt.Render(redirectTarget)))
	}

	statusColor := lipgloss.Color("#00FF7F")
	if strings.ToLower(operating) != "online" && strings.ToLower(operating) != "active" {
		statusColor = lipgloss.Color("#FFD700")
		if strings.ToLower(operating) == "error" {
			statusColor = lipgloss.Color("#FF6B6B")
		}
	}
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Operating Status"),
		lipgloss.NewStyle().Foreground(statusColor).Render(operating)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Supply Status"), valueSt.Render(provisioning)))

	infoBox := renderBox("L7 Policy "+policyName, infoContent.String(), width-4)
	content.WriteString(infoBox)

	return content.String()
}

func (m Model) renderLBL7RulesView(width int) string {
	var content strings.Builder

	if m.selectedLBL7Policy == nil {
		return "No L7 policy selected"
	}

	policyID := getStringValue(m.selectedLBL7Policy, "id", "")
	policyName := getStringValue(m.selectedLBL7Policy, "name", "N/A")
	rules := m.lbL7Rules[policyID]

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Actions bar — navigate with ←/→, execute with Enter (same pattern as other detail views)
	actions := []string{"+ Create", "✏ Edit", "🗑 Delete"}
	nActions := 1
	if len(rules) > 0 {
		nActions = 3
	}
	actionIdx := m.lbL7RuleActionIdx
	if actionIdx >= nActions {
		actionIdx = 0
	}
	var actionParts []string
	for i := 0; i < nActions; i++ {
		if i == actionIdx {
			bg := lipgloss.Color("#00AA55")
			if i == 2 {
				bg = lipgloss.Color("#FF4444")
			} else if i == 1 {
				bg = lipgloss.Color("#7B68EE")
			}
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(bg).Foreground(lipgloss.Color("#FFFFFF")).Bold(true).Padding(0, 1).Render(actions[i]))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+actions[i]+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.lbL7RuleConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render("⚠️  Press Enter to confirm delete, Esc to cancel")
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)
	content.WriteString(actionsBox + "\n\n")

	// Summary line: rule count + pagination hint
	titleLine := fmt.Sprintf("L7 Rules (%d) — Policy: %s", len(rules), policyName)

	if len(rules) == 0 {
		empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  No L7 Rules defined for this policy.")
		content.WriteString(renderBox(titleLine, empty, width-4))
		return content.String()
	}

	// Clamp index
	idx := m.lbL7RuleDetailIdx
	if idx < 0 {
		idx = 0
	}
	if idx >= len(rules) {
		idx = len(rules) - 1
	}

	// Pagination indicator in title
	pagTitle := fmt.Sprintf("%s  [◄ %d/%d ►]", titleLine, idx+1, len(rules))

	r := rules[idx]
	rType := getStringValue(r, "ruleType", getStringValue(r, "type", "N/A"))
	rComp := getStringValue(r, "compareType", "N/A")
	rKey := getStringValue(r, "key", "-")
	rVal := getStringValue(r, "value", "-")
	rID := getStringValue(r, "id", "N/A")
	invertStr := "No"
	if inv, ok := r["invert"].(bool); ok && inv {
		invertStr = "Yes"
	}
	rStatus := getStringValue(r, "provisioningStatus", getStringValue(r, "operatingStatus", "N/A"))

	var detailContent strings.Builder
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(rID, 36))))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Type"), valueSt.Render(rType)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Comparison"), valueSt.Render(rComp)))
	if rKey != "-" {
		detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Key"), valueSt.Render(rKey)))
	}
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Value"), valueSt.Render(rVal)))
	invertStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))
	if invertStr == "Yes" {
		invertStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Invert"), invertStyle.Render(invertStr)))

	statusColor := lipgloss.Color("#00FF7F")
	if strings.ToLower(rStatus) != "active" && strings.ToLower(rStatus) != "online" {
		statusColor = lipgloss.Color("#FFD700")
		if strings.ToLower(rStatus) == "error" {
			statusColor = lipgloss.Color("#FF6B6B")
		}
	}
	detailContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Prov. Status"),
		lipgloss.NewStyle().Foreground(statusColor).Render(rStatus)))

	content.WriteString(renderBox(pagTitle, detailContent.String(), width-4))
	return content.String()
}

func (m Model) renderLBPoolMembersView(width int) string {
	var content strings.Builder

	if m.selectedLBPool == nil {
		return "No pool selected"
	}

	poolID := getStringValue(m.selectedLBPool, "id", "")
	poolName := getStringValue(m.selectedLBPool, "name", "N/A")
	members := m.lbPoolMembers[poolID]

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(22)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	// Build action buttons — same style as LBPoolDetailView
	type actionDef struct {
		label string
		color string
	}
	allActions := []actionDef{
		{"Create", "#00AA55"},
		{"Edit", "#7B68EE"},
		{"Delete", "#CC3333"},
		{"Health Monitor", "#FF8C00"},
	}
	// When no members, only show Create and Health Monitor
	var actions []actionDef
	var actionLocalIdx []int // maps local idx → allActions idx
	actions = append(actions, allActions[0])
	actionLocalIdx = append(actionLocalIdx, 0)
	if len(members) > 0 {
		actions = append(actions, allActions[1], allActions[2])
		actionLocalIdx = append(actionLocalIdx, 1, 2)
	}
	actions = append(actions, allActions[3])
	actionLocalIdx = append(actionLocalIdx, 3)

	// Clamp action index to valid range
	actionIdx := m.lbMembersActionIdx
	if actionIdx >= len(actions) {
		actionIdx = len(actions) - 1
	}

	var actionParts []string
	for i, a := range actions {
		label := a.label
		if actionLocalIdx[i] == 2 && m.lbPoolMemberConfirm {
			label = "Confirm delete?"
		}
		if i == actionIdx && m.lbMembersSection == 0 {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color(a.color)).Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render(label))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+label+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.lbMembersSection == 0 && len(members) > 0 {
		actionsContent += "   " + lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("↓ browse members")
	}
	actionsBox := renderBox("Actions  ←→: Navigate • Enter: Confirm", actionsContent, width-4)
	content.WriteString(actionsBox + "\n\n")

	titleLine := fmt.Sprintf("Members (%d) — Pool: %s", len(members), poolName)

	if len(members) == 0 {
		empty := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("  No members defined for this pool.")
		content.WriteString(renderBox(titleLine, empty, width-4))
		return content.String()
	}

	// Clamp member index
	idx := m.lbPoolMemberDetailIdx
	if idx < 0 {
		idx = 0
	}
	if idx >= len(members) {
		idx = len(members) - 1
	}

	sectionFocus := ""
	if m.lbMembersSection == 1 {
		sectionFocus = " [← prev  → next  ↑ Actions]"
	}
	pagTitle := fmt.Sprintf("%s  %d/%d%s", titleLine, idx+1, len(members), sectionFocus)

	mem := members[idx]
	memName := getStringValue(mem, "name", "N/A")
	memID := getStringValue(mem, "id", "N/A")
	memAddr := getStringValue(mem, "address", "N/A")
	memPort := fmt.Sprintf("%v", mem["protocolPort"])
	memWeight := fmt.Sprintf("%v", mem["weight"])
	memOpStatus := getStringValue(mem, "operatingStatus", "N/A")
	memProvStatus := getStringValue(mem, "provisioningStatus", "N/A")

	statusColor := lipgloss.Color("#00FF7F")
	if strings.ToLower(memOpStatus) != "online" {
		statusColor = lipgloss.Color("#FFD700")
		if strings.ToLower(memOpStatus) == "error" {
			statusColor = lipgloss.Color("#FF6B6B")
		}
	}

	var detailContent strings.Builder
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(memID, 36))))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Name"), valueSt.Render(memName)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("IP Address"), valueSt.Render(memAddr)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Port"), valueSt.Render(memPort)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Weight"), valueSt.Render(memWeight)))
	detailContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Operating Status"),
		lipgloss.NewStyle().Foreground(statusColor).Render(memOpStatus)))
	detailContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Prov. Status"), valueSt.Render(memProvStatus)))

	content.WriteString(renderBox(pagTitle, detailContent.String(), width-4))
	return content.String()
}

func (m Model) renderFIPDetail(width int) string {
	var content strings.Builder

	fipID := getStringValue(m.detailData, "id", "N/A")
	fipIP := getStringValue(m.detailData, "ip", "N/A")
	region := getStringValue(m.detailData, "region", "N/A")
	status := getStringValue(m.detailData, "status", "N/A")

	// Associated entity (instance or LB)
	associatedTo := "-"
	if entity, ok := m.detailData["associatedEntity"].(map[string]interface{}); ok {
		entityType := getStringValue(entity, "type", "")
		entityID := getStringValue(entity, "id", "")
		if entityID != "" {
			associatedTo = entityType + ": " + entityID
		}
	}

	labelSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(20)
	valueSt := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	boxWidth := (width - 6) / 2
	if boxWidth < 35 {
		boxWidth = 35
	}

	// Status styling
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF7F"))
	statusIcon := "🟢"
	if strings.ToLower(status) != "active" {
		statusIcon = "🟡"
		statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700"))
	}

	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("IP Address"), valueSt.Render(fipIP)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("ID"), valueSt.Render(truncate(fipID, 36))))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Region"), valueSt.Render(region)))
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelSt.Render("Status"), statusStyle.Render(statusIcon+" "+status)))
	infoContent.WriteString(fmt.Sprintf("%s %s", labelSt.Render("Attached to"), valueSt.Render(associatedTo)))
	infoBox := renderBox("Floating IP "+fipIP, infoContent.String(), boxWidth)

	// Actions — "Detach" only shown when attached to something
	actions := []string{"Delete"}
	if associatedTo != "-" {
		actions = append(actions, "Detach")
	}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color("#7B68EE")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.actionConfirm {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actions[m.selectedAction]))
	}
	actionsBox := renderBox("Actions (←/→ to navigate, Enter to execute)", actionsContent, width-4)

	content.WriteString(actionsBox + "\n\n")
	content.WriteString(infoBox)
	return content.String()
}

func (m Model) renderPrivateNetworkDetail(width int) string {
	var content strings.Builder

	netID := getStringValue(m.detailData, "id", "N/A")
	netName := getStringValue(m.detailData, "name", "Unknown")
	vlanID := int(getFloatValue(m.detailData, "vlanId", 0))
	regionType := getStringValue(m.detailData, "_regionType", "region")

	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Width(18)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	boxWidth := (width - 6) / 2
	fullWidth := width - 4

	// Info box (top-left)
	rawRegions, _ := m.detailData["regions"].([]interface{})
	var infoContent strings.Builder
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("ID"), valueStyle.Render(truncate(netID, 36))))
	vlanStr := "automatic"
	if vlanID > 0 { vlanStr = fmt.Sprintf("%d", vlanID) }
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("VLAN ID"), valueStyle.Render(vlanStr)))
	rTypeLabel := "Region (vRack)"
	if regionType == "localzone" { rTypeLabel = "Local Zone" }
	infoContent.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Type"), valueStyle.Render(rTypeLabel)))
	// Regions list
	if len(rawRegions) == 0 {
		infoContent.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("Regions"), lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("none")))
	} else {
		for idx, rv := range rawRegions {
			rm, _ := rv.(map[string]interface{})
			rName := getStringValue(rm, "region", "N/A")
			rStatus := getStringValue(rm, "status", "")
			label := "Regions"
			if idx > 0 { label = "" }
			line := rName
			if rStatus != "" { line += "  (" + rStatus + ")" }
			// Highlight selected region when Delete Region action is active
			if m.selectedAction == 4 && idx == m.privNetSelectedRegion {
				arrow := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true).Render(" ◄")
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF6B6B")).Bold(true).Render(line) + arrow
				if m.actionConfirm {
					line += lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).Render("  ⚠️  Enter to confirm, Esc to cancel")
				}
			} else {
				line = valueStyle.Render(line)
			}
			sep := "\n"
			if idx == len(rawRegions)-1 { sep = "" }
			infoContent.WriteString(fmt.Sprintf("%s %s%s", labelStyle.Render(label), line, sep))
		}
	}
	infoBox := renderBox("Private Network", infoContent.String(), boxWidth)

	// Subnets — full-width list below info row, one entry per subnet
	subnets, _ := m.detailData["_subnets"].([]map[string]any)
	var subnetBoxes []string
	if len(subnets) == 0 {
		subnetBoxes = append(subnetBoxes, renderBox("Subnets (0)", lipgloss.NewStyle().Foreground(lipgloss.Color("#666666")).Render("No subnets"), fullWidth))
	} else {
		for idx, sub := range subnets {
			var sc strings.Builder
			subID := getStringValue(sub, "id", "")
			cidr := getStringValue(sub, "cidr", "N/A")
			gatewayIP := getStringValue(sub, "gatewayIp", "N/A")
			dhcpStr := "N/A"
			if dhcp, ok := sub["dhcpEnabled"].(bool); ok {
				if dhcp { dhcpStr = "enabled" } else { dhcpStr = "disabled" }
			}
			allocPool := "-"
			if pools, ok := sub["ipPools"].([]interface{}); ok && len(pools) > 0 {
				if pool, ok := pools[0].(map[string]interface{}); ok {
					start := getString(pool, "start")
					end := getString(pool, "end")
					if start != "" && end != "" {
						allocPool = start + " – " + end
					}
				}
			}
			if subID != "" {
				sc.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("ID"), valueStyle.Render(truncate(subID, 36))))
			}
			sc.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("CIDR"), valueStyle.Render(cidr)))
			sc.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("Gateway"), valueStyle.Render(gatewayIP)))
			sc.WriteString(fmt.Sprintf("%s %s\n", labelStyle.Render("DHCP"), valueStyle.Render(dhcpStr)))
			sc.WriteString(fmt.Sprintf("%s %s", labelStyle.Render("IP allocated"), valueStyle.Render(allocPool)))
			boxTitle := fmt.Sprintf("Subnet %d/%d", idx+1, len(subnets))
			// Highlight selected subnet when Delete Subnet action is active
			if m.selectedAction == 3 && idx == m.privNetSelectedSubnet {
				boxTitle += "  ◄ selected for deletion"
				if m.actionConfirm {
					sc.WriteString("\n\n" + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFD700")).Bold(true).
						Render("⚠️  Press Enter to confirm deletion, Esc to cancel"))
				}
			}
			subnetBoxes = append(subnetBoxes, renderBox(boxTitle, sc.String(), fullWidth))
		}
	}

	_ = netName

	// Actions
	actions := []string{"Delete", "Assign Gateway", "Add Subnet", "Delete Subnet", "Delete Region", "Detach Gateway"}
	var actionParts []string
	for i, action := range actions {
		if i == m.selectedAction {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Background(lipgloss.Color("#7B68EE")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).Padding(0, 1).Render(action))
		} else {
			actionParts = append(actionParts, lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).Padding(0, 1).Render("["+action+"]"))
		}
	}
	actionsContent := strings.Join(actionParts, " ")
	if m.actionConfirm && m.selectedAction != 3 && m.selectedAction != 4 {
		actionsContent += "\n\n" + lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")).Bold(true).
			Render(fmt.Sprintf("⚠️  Press Enter to confirm %s, Escape to cancel", actions[m.selectedAction]))
	}
	hintAction := "←/→ to navigate, Enter to execute"
	if m.selectedAction == 3 {
		hintAction = "←/→ to navigate • ↑/↓ to select subnet • Enter to delete"
	} else if m.selectedAction == 4 {
		hintAction = "←/→ to navigate • ↑/↓ to select region • Enter to delete"
	}
	actionsBox := renderBox("Actions ("+hintAction+")", actionsContent, width-4)

	content.WriteString(actionsBox + "\n\n")
	content.WriteString(infoBox + "\n\n")
	for _, sb := range subnetBoxes {
		content.WriteString(sb + "\n")
	}
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
		} else if (m.inStorageSubNav || m.inNetworkSubNav || m.inComputeSubNav) && m.inTableFocus {
			if m.currentProduct == ProductNetworkPrivate {
				help = "↑↓: Navigate • ←→: vRack Regions↔Local Zones • Enter: Details • c: Create • /: Filter • d: Debug • Esc: Back • q: Quit"
			} else if m.currentProduct == ProductStorageObject {
				help = "↑↓: Navigate • ←→: Containers↔Users • Enter: Details • c: Create • /: Filter • d: Debug • Esc: Back • q: Quit"
			} else {
				help = "↑↓: Navigate • Enter: Details • c: Create • /: Filter • d: Debug • Esc: Back to Sub-menu • q: Quit"
			}
		} else if m.inStorageSubNav || m.inNetworkSubNav || m.inComputeSubNav {
			help = "←→: Sub-menu • ↓/Enter: Enter Table • ↑/Esc: Back to main nav • d: Debug • p: Change Project • q: Quit"
		} else {
			help = "←→: Switch Product • Enter: Enter Sub-menu • ↑↓: Navigate • /: Filter • Enter: Details • c: Create • Del: Delete • d: Debug • p: Change Project • q: Quit"
		}
	case EmptyView:
		if (m.inStorageSubNav || m.inNetworkSubNav || m.inComputeSubNav) && m.inTableFocus {
			help = "c: Create • d: Debug • Esc: Back to Sub-menu • q: Quit"
		} else if m.inStorageSubNav || m.inNetworkSubNav || m.inComputeSubNav {
			help = "←→: Sub-menu • Enter: Enter Table • ↑/Esc: Back to main nav • c: Create • d: Debug • p: Change Project • q: Quit"
		} else {
			help = "←→: Switch Product • Enter: Enter Sub-menu • c: Create • d: Debug • p: Change Project • q: Quit"
		}
	case DetailView:
		if m.currentProduct == ProductStorageObject && m.objectUserDetailView != nil {
			help = m.objectUserDetailView.HelpText()
		} else if m.currentProduct == ProductStorageObject && m.objectDetailView != nil {
			help = m.objectDetailView.HelpText()
		} else if m.currentProduct == ProductStorageSnapshot && m.snapshotDetailView != nil {
			help = m.snapshotDetailView.HelpText()
		} else if m.currentProduct == ProductStorageBackup && m.backupDetailView != nil {
			help = m.backupDetailView.HelpText()
		} else if m.currentProduct == ProductNetworkPrivate {
			if m.actionConfirm {
				help = "Enter: Confirmer l'action • Esc: Annuler"
			} else {
				help = "←→: Select action • Enter: Execute • Esc: Back to list • q: Quit"
			}
		} else if m.actionConfirm {
			help = "Enter: Confirm Action • Esc: Cancel"
		} else if m.currentProduct == ProductNetworkLB {
			if m.lbListenerListIdx >= 0 {
				help = "↑↓: Navigate listeners • Enter: Open • ←→: Select action • Esc: Back to list • q: Quit"
			} else if m.lbPoolListIdx >= 0 {
				help = "↑↓: Navigate pools • Enter: Open • ←→: Select action • Esc: Back to list • q: Quit"
			} else {
				help = "←→: Select action • Enter: Execute • ↓: Navigate listeners/pools • d: Debug • Esc: Back • q: Quit"
			}
		} else {
			help = "←→: Select Action • Enter: Execute • d: Debug • Esc: Back to List • q: Quit"
		}
	case LBPoolDetailView:
		if m.lbPoolDetailConfirm {
			help = "Enter: Confirm • Esc: Cancel"
		} else {
			help = "←→: Select Action • Enter: Execute • Esc: Back to LB • q: Quit"
		}
	case LBListenerDetailView:
		if m.lbListenerDetailConfirm {
			help = "Enter: Confirm • Esc: Cancel"
		} else if m.lbL7PolicyListIdx >= 0 {
			help = "↑↓: Navigate Policies • Enter: Open • Esc: Deselect • q: Quit"
		} else {
			help = "←→: Select Action • Enter: Execute • ↓: Navigate Policies • Esc: Back to LB • q: Quit"
		}
	case LBL7PolicyDetailView:
		if m.lbL7PolicyDetailConfirm {
			help = "Enter: Confirm • Esc: Cancel"
		} else {
			help = "←→: Select Action • Enter: Execute • Esc: Back to Listener • q: Quit"
		}
	case LBL7RulesView:
		if m.lbL7RuleConfirm {
			help = "Enter: Confirm delete • Esc: Cancel"
		} else {
			help = "←→: Select action • ↑↓: Browse rules • Enter: Execute action • Esc: Back • q: Quit"
		}
	case LBPoolMembersView:
		if m.lbMembersSection == 0 {
			help = "←→: Select action • Enter: Confirm • ↓: Browse members • Esc: Back • q: Quit"
		} else {
			help = "←→ / ↑↓: Navigate members • ↑: Back to actions • Enter: Confirm action • Esc: Back • q: Quit"
		}
	case LBHealthMonitorView:
		if m.lbHMConfirm {
			help = "Enter: Confirm delete • Esc: Cancel"
		} else {
			help = "←→: Select action • Enter: Execute • Esc: Back to Members • q: Quit"
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
		} else if m.wizard.step == VolumeWizardStepName {
			help = "Type: Enter name • Enter: Confirm • Esc: Cancel"
		} else if m.wizard.step == VolumeWizardStepRegion {
			help = "↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == VolumeWizardStepType {
			help = "↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == VolumeWizardStepAvailabilityZone {
			help = "↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == VolumeWizardStepSize {
			help = "Type size in GB • Enter: Confirm • ←: Back • Esc: Cancel"
		} else if m.wizard.step == VolumeWizardStepEncryption {
			help = "↑↓: Select • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == VolumeWizardStepConfirm {
			help = "←→: Select • Enter: Confirm • Esc: Cancel"
		} else if m.wizard.step == FileWizardStepName {
			help = "Type name • Enter: Continue • Esc: Cancel"
		} else if m.wizard.step == FileWizardStepRegion {
			help = "↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == FileWizardStepType {
			help = "↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == FileWizardStepSize {
			help = "Type size in GB • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == FileWizardStepNetwork {
			help = "↑↓: Navigate • Enter: Select/Expand • ←: Back • Esc: Cancel"
		} else if m.wizard.step == FileWizardStepConfirm {
			help = "←→: Select • Enter: Confirm • Esc: Cancel"
		} else if m.wizard.step == S3UserWizardStepDescription {
			help = "Type description • Enter: Continue • Esc: Cancel"
		} else if m.wizard.step == S3UserWizardStepConfirm {
			help = "←→: Select • Enter: Confirm • Esc: Back"
		} else if m.wizard.step == BackupWizardStepVolume {
			help = "↑↓: Navigate • Enter: Select • Esc: Cancel"
		} else if m.wizard.step == BackupWizardStepType {
			help = "↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == BackupWizardStepName {
			help = "Type name • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == BackupWizardStepConfirm {
			help = "←→: Select • Enter: Confirm • Esc: Cancel"
		} else if m.wizard.step == LBL7RuleWizardStepKey {
			help = "Type key name • Enter: Confirm • Esc: Cancel"
		} else if m.wizard.step == LBL7RuleWizardStepValue {
			help = "Type value • Enter: Confirm • Esc: Cancel"
		} else if m.wizard.step == LBMemberWizardStepName {
			help = "Type name • Enter: Continue • Esc: Cancel"
		} else if m.wizard.step == LBMemberWizardStepIP {
			help = "Type IP address • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBMemberWizardStepPort {
			help = "Type port (1-65535) • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBMemberWizardStepWeight {
			help = "Type weight (0-256) • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBMemberWizardStepConfirm {
			help = "←→: Select • Enter: Save • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepName {
			help = "Type name (letters/digits/_-.) • Enter: Continue • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepType {
			help = "↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepHttpMethod {
			help = "↑↓: Navigate • Enter: Select • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepUrlPath {
			help = "Type URL path (e.g. /) • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepExpectedCodes {
			help = "Type expected HTTP codes (e.g. 200) • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepDelay {
			help = "Type delay in seconds • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepMaxRetries {
			help = "Type max retries (1-10) • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepMaxRetriesDown {
			help = "Type max retries down (1-10) • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepTimeout {
			help = "Type timeout in seconds • Enter: Continue • ←: Back • Esc: Cancel"
		} else if m.wizard.step == LBHMWizardStepConfirm {
			help = "←→: Select • Enter: Save • Esc: Cancel"
		} else if m.mode == LBL7RulesView {
			help = "←→: Select action • ↑↓: Browse rules • Enter: Execute • Esc: Back"
		} else {
			help = "↑↓: Navigate • d: Debug • Enter: Select • ←: Back • Esc: Cancel"
		}
	case DeleteConfirmView:
		help = "Type instance name to confirm • Enter: Delete • Esc: Cancel"
	case S3CredentialsView:
		help = "s: Save to ~/.aws/credentials • Enter/Esc: Continue • q: Quit"
	case DebugView:
		help = "↑↓: Scroll • c: Clear logs • d/Esc: Close • q: Quit"
	case KubeKubeconfigPickerView:
		help = "↑↓: Navigate • Enter: Open/Select • Esc: Cancel"
	case ComingSoonView:
		help = "←→: Switch Product • tab/shift+tab: Navigate Sub-menu • d: Debug • p: Change Project • q: Quit"
	default:
		help = "Enter: Select • q: Quit"
	}

	// Add experimental warning
	experimentalNotice := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFA500")).
		Render("⚠️  Experimental feature - Report bugs at: https://github.com/ovh/ovhcloud-cli/issues")

	// Add notification if present
	if m.notification != "" && time.Now().Before(m.notificationExpiry) {
		notificationStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF7F")).
			Bold(true)
		if strings.HasPrefix(m.notification, "❌") {
			notificationStyle = notificationStyle.Foreground(lipgloss.Color("#FF6B6B"))
		}
		return notificationStyle.Render(m.notification) + "\n" + footerStyle.Render(help) + "\n\n" + experimentalNotice
	}

	return footerStyle.Render(help) + "\n\n" + experimentalNotice
}

func (m Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Handle wizard mode separately
	if m.mode == WizardView {
		return m.handleWizardKeyPress(msg)
	}

	// Handle S3 credentials display view
	if m.mode == S3CredentialsView {
		return m.handleS3CredentialsViewKeys(msg)
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

	// Handle kubeconfig directory picker view
	if m.mode == KubeKubeconfigPickerView {
		return m.handleKubeKubeconfigPickerKeyPress(msg)
	}

	// Delegate to block storage detail view when in DetailView for ProductStorageBlock
	if m.mode == DetailView && m.currentProduct == ProductStorageBlock && m.volumeDetailView != nil {
		cmd := m.volumeDetailView.HandleKey(msg)
		return m, cmd
	}

	// Delegate to snapshot detail view
	if m.mode == DetailView && m.currentProduct == ProductStorageSnapshot && m.snapshotDetailView != nil {
		cmd := m.snapshotDetailView.HandleKey(msg)
		return m, cmd
	}

	// Delegate to backup detail view
	if m.mode == DetailView && m.currentProduct == ProductStorageBackup && m.backupDetailView != nil {
		cmd := m.backupDetailView.HandleKey(msg)
		return m, cmd
	}

        // Delegate to file storage detail view when in DetailView for ProductStorageFile
        if m.mode == DetailView && m.currentProduct == ProductStorageFile && m.fileShareDetailView != nil {
                cmd := m.fileShareDetailView.HandleKey(msg)
                return m, cmd
        }

        // Delegate to object storage detail view when in DetailView for ProductStorageObject
        if m.mode == DetailView && m.currentProduct == ProductStorageObject && m.objectDetailView != nil {
                cmd := m.objectDetailView.HandleKey(msg)
                return m, cmd
        }

        // Delegate to object storage user detail view
        if m.mode == DetailView && m.currentProduct == ProductStorageObject && m.objectUserDetailView != nil {
                cmd := m.objectUserDetailView.HandleKey(msg)
                return m, cmd
        }

	// Intercept all keys when DB user creation text input is active
	if m.dbUserCreateMode && m.mode == DetailView &&
		(m.currentProduct == ProductManagedDatabases || m.currentProduct == ProductManagedAnalytics) {
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "esc":
			m.dbUserCreateMode = false
			m.dbUserCreateInput = ""
		case "backspace":
			if len(m.dbUserCreateInput) > 0 {
				m.dbUserCreateInput = m.dbUserCreateInput[:len(m.dbUserCreateInput)-1]
			}
		case "enter":
			username := strings.TrimSpace(m.dbUserCreateInput)
			if username == "" {
				return m, nil
			}
			m.dbUserCreateMode = false
			m.dbUserCreateInput = ""
			m.dbDetailLoaded = false
			return m, m.createDBUser(username)
		default:
			if len(msg.Runes) > 0 {
				m.dbUserCreateInput += string(msg.Runes)
			}
		}
		return m, nil
	}

        switch msg.String() {
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
		// In DetailView for Private Networks, navigate actions (0=Delete, 1=Assign Gateway)
		if m.mode == DetailView && m.currentProduct == ProductNetworkPrivate {
			if m.selectedAction > 0 {
				m.selectedAction--
				m.actionConfirm = false
			}
			return m, nil
		}
		// In LBPoolDetailView, navigate actions
		if m.mode == LBPoolDetailView {
			if m.lbPoolDetailActionIdx > 0 {
				m.lbPoolDetailActionIdx--
				m.lbPoolDetailConfirm = false
			}
			return m, nil
		}
		// In LBL7PolicyDetailView, navigate actions
		if m.mode == LBL7PolicyDetailView {
			if m.lbL7PolicyDetailActionIdx > 0 {
				m.lbL7PolicyDetailActionIdx--
				m.lbL7PolicyDetailConfirm = false
			}
			return m, nil
		}
		// In LBL7RulesView, left/right navigate action buttons
		if m.mode == LBL7RulesView {
			policyID := getStringValue(m.selectedLBL7Policy, "id", "")
			nActions := 1
			if len(m.lbL7Rules[policyID]) > 0 {
				nActions = 3
			}
			if m.lbL7RuleActionIdx > 0 {
				m.lbL7RuleActionIdx--
				m.lbL7RuleConfirm = false
			} else {
				m.lbL7RuleActionIdx = nActions - 1
				m.lbL7RuleConfirm = false
			}
			return m, nil
		}
		// In LBHealthMonitorView: left navigates action buttons
		if m.mode == LBHealthMonitorView {
			poolID := getStringValue(m.selectedLBPool, "id", "")
			hm := m.lbHealthMonitors[poolID]
			if hm != nil {
				if m.lbHMActionIdx > 0 {
					m.lbHMActionIdx--
				} else {
					m.lbHMActionIdx = 1
				}
				m.lbHMConfirm = false
			}
			return m, nil
		}
		// In LBPoolMembersView: left navigates action buttons (section 0) or prev member (section 1)
		if m.mode == LBPoolMembersView && m.selectedLBPool != nil {
			if m.lbMembersSection == 0 {
				// Count available actions
				poolID := getStringValue(m.selectedLBPool, "id", "")
				nActions := 2 // Create + HM
				if len(m.lbPoolMembers[poolID]) > 0 {
					nActions = 4 // Create + Edit + Delete + HM
				}
				if m.lbMembersActionIdx > 0 {
					m.lbMembersActionIdx--
				} else {
					m.lbMembersActionIdx = nActions - 1
				}
			} else {
				poolID := getStringValue(m.selectedLBPool, "id", "")
				count := len(m.lbPoolMembers[poolID])
				if count > 0 {
					if m.lbPoolMemberDetailIdx > 0 {
						m.lbPoolMemberDetailIdx--
					} else {
						m.lbPoolMemberDetailIdx = count - 1
					}
				}
			}
			return m, nil
		}
		// In LBListenerDetailView, navigate actions (only when not navigating policies)
		if m.mode == LBListenerDetailView && m.lbL7PolicyListIdx < 0 {
			if m.lbListenerDetailActionIdx > 0 {
				m.lbListenerDetailActionIdx--
				m.lbListenerDetailConfirm = false
			}
			return m, nil
		}
		// In DetailView for Load Balancers, navigate actions (0=Delete, 1=Listeners, 2=Pools)
		if m.mode == DetailView && m.currentProduct == ProductNetworkLB {
			if m.selectedAction > 0 {
				m.selectedAction--
				m.actionConfirm = false
			}
			return m, nil
		}
		// In DetailView for Workflow, only 1 action (Delete)
		if m.mode == DetailView && m.currentProduct == ProductWorkflow {
			return m, nil
		}
		// In DetailView for Instance Backup, only 1 action (Delete)
		if m.mode == DetailView && m.currentProduct == ProductInstanceBackup {
			return m, nil
		}
		// In DetailView for ManagedDatabases/Analytics, ← switches tabs
		if m.mode == DetailView && (m.currentProduct == ProductManagedDatabases || m.currentProduct == ProductManagedAnalytics) {
			if m.dbDetailTab > 0 {
				m.dbDetailTab--
				m.actionConfirm = false
			}
			return m, nil
		}
		// In DetailView for Floating IPs, navigate actions (0=Delete, 1=Detach)
		if m.mode == DetailView && m.currentProduct == ProductNetworkPublic {
			if m.selectedAction > 0 {
				m.selectedAction--
				m.actionConfirm = false
			}
			return m, nil
		}
		// Object Storage: ←/→ switches between Containers and Users tabs when in table focus
		if m.inStorageSubNav && m.inTableFocus && m.currentProduct == ProductStorageObject &&
			(m.mode == TableView || m.mode == EmptyView) {
			if m.objectStorageTabIdx > 0 {
				m.objectStorageTabIdx = 0
				m.table = createObjectStorageTable(m.currentData, m.width, m.height)
			}
			return m, nil
		}
		// In storage sub-nav (only when focused, not in table)
		if m.inStorageSubNav && !m.inTableFocus && m.mode != DetailView {
			subItems := getStorageSubItems()
			for i, item := range subItems {
				if item.Product == m.currentProduct {
					m.storageSubIdx = i
					break
				}
			}
			m.storageSubIdx = (m.storageSubIdx - 1 + len(subItems)) % len(subItems)
			return m.loadStorageSubProduct()
		}
		// In network sub-nav (only when focused, not in table)
		if m.inNetworkSubNav && !m.inTableFocus && m.mode != DetailView {
			subItems := getNetworkSubItems()
			for i, item := range subItems {
				if item.Product == m.currentProduct {
					m.networkSubIdx = i
					break
				}
			}
			m.networkSubIdx = (m.networkSubIdx - 1 + len(subItems)) % len(subItems)
			return m.loadNetworkSubProduct()
		}
		// In compute sub-nav (only when focused, not in table)
		if m.inComputeSubNav && !m.inTableFocus && m.mode != DetailView {
			subItems := getComputeSubItems()
			for i, item := range subItems {
				if item.Product == m.currentProduct {
					m.computeSubIdx = i
					break
				}
			}
			m.computeSubIdx = (m.computeSubIdx - 1 + len(subItems)) % len(subItems)
			return m.loadComputeSubProduct()
		}
		if m.mode != ProjectSelectView && m.currentProduct != ProductProjects && !m.inTableFocus {
			if m.navIdx > 0 {
				m.navIdx--
				m.inStorageSubNav = false
				m.inNetworkSubNav = false
				m.inComputeSubNav = false
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
			if m.selectedAction < 7 { // 8 actions: 0-7
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
		// In DetailView for Private Networks, navigate actions (0=Delete, 1=Assign Gateway, 2=Add Subnet, 3=Delete Subnet)
		if m.mode == DetailView && m.currentProduct == ProductNetworkPrivate {
			if m.selectedAction < 5 {
				m.selectedAction++
				m.actionConfirm = false
			}
			return m, nil
		}
		// In LBPoolDetailView, navigate actions (0=Edit, 1=Delete, 2=Members)
		if m.mode == LBPoolDetailView {
			if m.lbPoolDetailActionIdx < 2 {
				m.lbPoolDetailActionIdx++
				m.lbPoolDetailConfirm = false
			}
			return m, nil
		}
		// In LBL7PolicyDetailView, navigate actions (0=Edit, 1=Delete, 2=L7 Rules)
		if m.mode == LBL7PolicyDetailView {
			if m.lbL7PolicyDetailActionIdx < 2 {
				m.lbL7PolicyDetailActionIdx++
				m.lbL7PolicyDetailConfirm = false
			}
			return m, nil
		}
		// In LBL7RulesView, left/right navigate action buttons (right)
		if m.mode == LBL7RulesView {
			policyID := getStringValue(m.selectedLBL7Policy, "id", "")
			nActions := 1
			if len(m.lbL7Rules[policyID]) > 0 {
				nActions = 3
			}
			m.lbL7RuleActionIdx = (m.lbL7RuleActionIdx + 1) % nActions
			m.lbL7RuleConfirm = false
			return m, nil
		}
		// In LBHealthMonitorView: right navigates action buttons
		if m.mode == LBHealthMonitorView {
			poolID := getStringValue(m.selectedLBPool, "id", "")
			hm := m.lbHealthMonitors[poolID]
			if hm != nil {
				m.lbHMActionIdx = (m.lbHMActionIdx + 1) % 2
				m.lbHMConfirm = false
			}
			return m, nil
		}
		// In LBPoolMembersView: right navigates action buttons (section 0) or next member (section 1)
		if m.mode == LBPoolMembersView && m.selectedLBPool != nil {
			if m.lbMembersSection == 0 {
				poolID := getStringValue(m.selectedLBPool, "id", "")
				nActions := 2
				if len(m.lbPoolMembers[poolID]) > 0 {
					nActions = 4
				}
				m.lbMembersActionIdx = (m.lbMembersActionIdx + 1) % nActions
			} else {
				poolID := getStringValue(m.selectedLBPool, "id", "")
				count := len(m.lbPoolMembers[poolID])
				if count > 0 {
					m.lbPoolMemberDetailIdx = (m.lbPoolMemberDetailIdx + 1) % count
				}
			}
			return m, nil
		}
		// In LBListenerDetailView, navigate actions (0=Edit, 1=Delete, 2=L7 Policies) — only when not navigating policies
		if m.mode == LBListenerDetailView && m.lbL7PolicyListIdx < 0 {
			if m.lbListenerDetailActionIdx < 2 {
				m.lbListenerDetailActionIdx++
				m.lbListenerDetailConfirm = false
			}
			return m, nil
		}
		// In DetailView for Load Balancers, navigate actions (0=Delete, 1=Listeners, 2=Pools)
		if m.mode == DetailView && m.currentProduct == ProductNetworkLB {
			if m.selectedAction < 2 {
				m.selectedAction++
				m.actionConfirm = false
			}
			return m, nil
		}
		// In DetailView for ManagedDatabases/Analytics, → switches tabs
		if m.mode == DetailView && (m.currentProduct == ProductManagedDatabases || m.currentProduct == ProductManagedAnalytics) {
			maxTab := 4
			if m.dbDetailTab < maxTab {
				m.dbDetailTab++
				m.actionConfirm = false
			}
			return m, nil
		}
		// In DetailView for Workflow, only 1 action (Delete)
		if m.mode == DetailView && m.currentProduct == ProductWorkflow {
			return m, nil
		}
		// In DetailView for Instance Backup, only 1 action (Delete)
		if m.mode == DetailView && m.currentProduct == ProductInstanceBackup {
			return m, nil
		}
		// In DetailView for ManagedDatabases/Analytics, → switches tabs
		if m.mode == DetailView && (m.currentProduct == ProductManagedDatabases || m.currentProduct == ProductManagedAnalytics) {
			if m.dbDetailTab < 4 {
				m.dbDetailTab++
				m.actionConfirm = false
			}
			return m, nil
		}
		// In DetailView for Floating IPs, navigate actions (0=Delete, 1=Detach)
		if m.mode == DetailView && m.currentProduct == ProductNetworkPublic {
			fipAttached := false
			if m.detailData != nil {
				if entity, ok := m.detailData["associatedEntity"].(map[string]interface{}); ok {
					fipAttached = getStringValue(entity, "id", "") != ""
				}
			}
			if fipAttached && m.selectedAction < 1 {
				m.selectedAction++
				m.actionConfirm = false
			}
			return m, nil
		}
		// Object Storage: ←/→ switches between Containers and Users tabs when in table focus
		if m.inStorageSubNav && m.inTableFocus && m.currentProduct == ProductStorageObject &&
			(m.mode == TableView || m.mode == EmptyView) {
			if m.objectStorageTabIdx < 1 {
				m.objectStorageTabIdx = 1
				m.table = createObjectStorageUsersTable(m.objectStorageUsers, m.width, m.height)
			}
			return m, nil
		}
		// In storage sub-nav (only when focused, not in table)
		isStorageSubProduct2 := m.currentProduct >= ProductStorageBlock && m.currentProduct <= ProductStorageArchive
		if m.inStorageSubNav && !m.inTableFocus && isStorageSubProduct2 && m.mode != DetailView {
			subItems := getStorageSubItems()
			for i, item := range subItems {
				if item.Product == m.currentProduct {
					m.storageSubIdx = i
					break
				}
			}
			m.storageSubIdx = (m.storageSubIdx + 1) % len(subItems)
			return m.loadStorageSubProduct()
		}
		// In network sub-nav (only when focused, not in table)
		isNetworkSubProduct2 := m.currentProduct >= ProductNetworkPrivate && m.currentProduct <= ProductNetworkLB
		if m.inNetworkSubNav && !m.inTableFocus && isNetworkSubProduct2 && m.mode != DetailView {
			subItems := getNetworkSubItems()
			for i, item := range subItems {
				if item.Product == m.currentProduct {
					m.networkSubIdx = i
					break
				}
			}
			m.networkSubIdx = (m.networkSubIdx + 1) % len(subItems)
			return m.loadNetworkSubProduct()
		}
		// In compute sub-nav (only when focused, not in table)
		isComputeSubProduct2 := m.currentProduct == ProductInstances || m.currentProduct == ProductInstanceBackup || m.currentProduct == ProductWorkflow
		if m.inComputeSubNav && !m.inTableFocus && isComputeSubProduct2 && m.mode != DetailView {
			subItems := getComputeSubItems()
			for i, item := range subItems {
				if item.Product == m.currentProduct {
					m.computeSubIdx = i
					break
				}
			}
			m.computeSubIdx = (m.computeSubIdx + 1) % len(subItems)
			return m.loadComputeSubProduct()
		}
		if m.mode != ProjectSelectView && m.currentProduct != ProductProjects && !m.inTableFocus {
			navItems := getNavItems()
			if m.navIdx < len(navItems)-1 {
				m.navIdx++
				m.inStorageSubNav = false
				m.inNetworkSubNav = false
				m.inComputeSubNav = false
				return m.loadCurrentProduct()
			}
		}
		return m, nil

	case "t":
		return m, nil

	case "p":
		// Go back to project selection
		if m.mode != ProjectSelectView && m.currentProduct != ProductProjects {
			m.currentProduct = ProductProjects
			m.navIdx = 0
			// If we have cached projects, show them directly
			if len(m.projectsList) > 0 {
				m.table = createProjectsTable(m.projectsList, m.height)
				m.currentData = m.projectsList
				m.mode = ProjectSelectView
				return m, nil
			}
			// Otherwise fetch projects
			m.mode = LoadingView
			return m, m.fetchDataForPath("/projects")
		}
		return m, nil

	case "shift+tab":
		return m, nil

	case "esc":
		// Clear filter in TableView if active
		if m.mode == TableView && m.filterInput != "" {
			m.filterInput = ""
			m.applyTableFilter()
			return m, nil
		}
		// Level 3 → Level 2: exit table focus (back to sub-nav focus)
		if m.inTableFocus && (m.inStorageSubNav || m.inNetworkSubNav || m.inComputeSubNav) && !m.isDeepNavigationView() {
			m.inTableFocus = false
			return m, nil
		}
		// Level 2 → Level 1: exit sub-nav focus (back to main nav)
		if m.inStorageSubNav && !m.inTableFocus && !m.isDeepNavigationView() {
			m.inStorageSubNav = false
			return m, nil
		}
		if m.inNetworkSubNav && !m.inTableFocus && !m.isDeepNavigationView() {
			m.inNetworkSubNav = false
			return m, nil
		}
		if m.inComputeSubNav && !m.inTableFocus && !m.isDeepNavigationView() {
			m.inComputeSubNav = false
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
		// Go back to LB detail from LB pool detail view, or cancel confirm
		if m.mode == LBPoolDetailView {
			if m.lbPoolDetailConfirm {
				m.lbPoolDetailConfirm = false
			} else {
				m.mode = DetailView
				m.selectedLBPool = nil
				m.lbPoolDetailActionIdx = 0
			}
			return m, nil
		}
		// Go back to listener detail from L7 policy detail, or cancel confirm
		if m.mode == LBL7PolicyDetailView {
			if m.lbL7PolicyDetailConfirm {
				m.lbL7PolicyDetailConfirm = false
			} else {
				m.mode = LBListenerDetailView
				m.selectedLBL7Policy = nil
				m.lbL7PolicyDetailActionIdx = 0
				m.lbL7PolicyListIdx = -1
			}
			return m, nil
		}
		// Go back to policy detail from L7 rules view, or cancel delete confirm
		if m.mode == LBL7RulesView {
			if m.lbL7RuleConfirm {
				m.lbL7RuleConfirm = false
			} else {
				m.mode = LBL7PolicyDetailView
			}
			return m, nil
		}
		// Go back to pool detail from pool members view, or cancel delete confirm / exit pagination
		if m.mode == LBPoolMembersView {
			if m.lbPoolMemberConfirm {
				m.lbPoolMemberConfirm = false
			} else if m.lbMembersSection == 1 {
				m.lbMembersSection = 0
			} else {
				m.mode = LBPoolDetailView
			}
			return m, nil
		}
		// Go back to pool members view from health monitor view, or cancel delete confirm
		if m.mode == LBHealthMonitorView {
			if m.lbHMConfirm {
				m.lbHMConfirm = false
			} else {
				m.mode = LBPoolMembersView
			}
			return m, nil
		}
		// Go back to LB detail from LB listener detail view, or cancel confirm / deselect policy
		if m.mode == LBListenerDetailView {
			if m.lbListenerDetailConfirm {
				m.lbListenerDetailConfirm = false
			} else if m.lbL7PolicyListIdx >= 0 {
				m.lbL7PolicyListIdx = -1
			} else {
				m.mode = DetailView
				m.selectedLBListener = nil
				m.lbListenerDetailActionIdx = 0
			}
			return m, nil
		}
		// Go back to detail view from node pools view
		if m.mode == NodePoolsView {
			m.mode = DetailView
			return m, nil
		}
		// Dismiss DB user creation result panel before going back to list
		if m.mode == DetailView && (m.currentProduct == ProductManagedDatabases || m.currentProduct == ProductManagedAnalytics) {
			if m.dbUserCreatedData != nil {
				m.dbUserCreatedData = nil
				return m, nil
			}
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
			// Require table focus (Level 3) for sub-nav products
			isSubNavProd := (m.currentProduct >= ProductStorageBlock && m.currentProduct <= ProductNetworkLB)
			if isSubNavProd && !m.inTableFocus {
				return m, nil
			}
			// Instance Backup: creation is done from an instance detail view, not here
			if m.currentProduct == ProductInstanceBackup {
				return m, nil
			}
			// If viewing S3 users tab, launch user creation wizard
			if m.currentProduct == ProductStorageObject && m.objectStorageTabIdx == 1 {
				m.mode = WizardView
				m.wizard = WizardData{step: S3UserWizardStepDescription}
				return m, nil
			}
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

	case "e":
		// In LBL7RulesView: edit is handled via Enter on the Edit action button
		if false && m.mode == LBL7RulesView && m.selectedLBL7Policy != nil {
			policyID := getStringValue(m.selectedLBL7Policy, "id", "")
			policyName := getStringValue(m.selectedLBL7Policy, "name", "")
			region := getStringValue(m.detailData, "region", "")
			rules := m.lbL7Rules[policyID]
			idx := m.lbL7RuleDetailIdx
			if idx < 0 {
				idx = 0
			}
			if len(rules) > 0 && idx < len(rules) {
				r := rules[idx]
				ruleID := getStringValue(r, "id", "")
				ruleType := getStringValue(r, "ruleType", getStringValue(r, "type", ""))
				compareType := getStringValue(r, "compareType", "")
				ruleKey := getStringValue(r, "key", "")
				ruleValue := getStringValue(r, "value", "")
				ruleInvert := false
				if inv, ok := r["invert"].(bool); ok {
					ruleInvert = inv
				}

				typeIdx := 0
				for i, opt := range lbL7RuleTypeOptions {
					if opt.value == ruleType {
						typeIdx = i
						break
					}
				}
				compareIdx := 0
				validOpts := validCompareOptionsForType(ruleType)
				for i, opt := range validOpts {
					if opt.value == compareType {
						compareIdx = i
						break
					}
				}

				// Only pre-fill key for types that require it
				prefillKey := ""
				for _, t := range lbL7RuleTypeOptions {
					if t.value == ruleType && t.needsKey {
						prefillKey = ruleKey
						break
					}
				}
				m.mode = WizardView
				m.wizard = WizardData{
					step:             LBL7RuleWizardStepType,
					l7RulePolicyId:   policyID,
					l7RulePolicyName: policyName,
					l7RuleLBRegion:   region,
					l7RuleEditId:     ruleID,
					l7RuleTypeIdx:    typeIdx,
					l7RuleType:       ruleType,
					l7RuleCompareIdx: compareIdx,
					l7RuleCompare:    compareType,
					l7RuleKeyInput:   prefillKey,
					l7RuleKey:        prefillKey,
					l7RuleValueInput: ruleValue,
					l7RuleValue:      ruleValue,
					l7RuleInvert:     ruleInvert,
				}
			}
		}
		// In LBHealthMonitorView: edit is handled via Enter on the Edit action button
		if false && m.mode == LBHealthMonitorView && m.selectedLBPool != nil {
			poolID := getStringValue(m.selectedLBPool, "id", "")
			region := getStringValue(m.detailData, "region", "")
			hm := m.lbHealthMonitors[poolID]
			if hm != nil {
				hmID := getStringValue(hm, "id", "")
				hmName := getStringValue(hm, "name", "")
				hmType := getStringValue(hm, "monitorType", "http")
				hmTypeIdx := 0
				for i, opt := range lbHMTypeOptions {
					if opt.value == hmType {
						hmTypeIdx = i
						break
					}
				}
				hmDelay := 5
				if v, ok := hm["delay"]; ok {
					switch d := v.(type) {
					case float64:
						hmDelay = int(d)
					case int:
						hmDelay = d
					}
				}
				hmTimeout := 5
				if v, ok := hm["timeout"]; ok {
					switch d := v.(type) {
					case float64:
						hmTimeout = int(d)
					case int:
						hmTimeout = d
					}
				}
				hmMaxRetries := 3
				if v, ok := hm["maxRetries"]; ok {
					switch d := v.(type) {
					case float64:
						hmMaxRetries = int(d)
					case int:
						hmMaxRetries = d
					}
				}
				hmMaxRetriesDown := 3
				if v, ok := hm["maxRetriesDown"]; ok {
					switch d := v.(type) {
					case float64:
						hmMaxRetriesDown = int(d)
					case int:
						hmMaxRetriesDown = d
					}
				}
				m.mode = WizardView
				// Extract httpMethod, urlPath, expectedCodes from existing httpConfiguration if present
				hmHttpMethod := "GET"
				hmHttpMethodIdx := 2 // GET is index 2 in lbHMHttpMethodOptions
				hmUrlPath := "/"
				hmExpectedCodes := "200"
				if httpCfg, ok := hm["httpConfiguration"].(map[string]interface{}); ok {
					if method, ok := httpCfg["httpMethod"].(string); ok && method != "" {
						hmHttpMethod = method
						for i, opt := range lbHMHttpMethodOptions {
							if opt == hmHttpMethod {
								hmHttpMethodIdx = i
								break
							}
						}
					}
					if up, ok := httpCfg["urlPath"].(string); ok && up != "" {
						hmUrlPath = up
					}
					if ec, ok := httpCfg["expectedCodes"].(string); ok && ec != "" {
						hmExpectedCodes = ec
					}
				}
				m.wizard = WizardData{
					step:                    LBHMWizardStepName,
					lbHMPoolId:              poolID,
					lbHMPoolRegion:          region,
					lbHMEditId:              hmID,
					lbHMNameInput:           hmName,
					lbHMName:                hmName,
					lbHMTypeIdx:             hmTypeIdx,
					lbHMType:                hmType,
					lbHMHttpMethodIdx:       hmHttpMethodIdx,
					lbHMHttpMethod:          hmHttpMethod,
					lbHMUrlPathInput:        hmUrlPath,
					lbHMUrlPath:             hmUrlPath,
					lbHMExpectedCodesInput:  hmExpectedCodes,
					lbHMExpectedCodes:       hmExpectedCodes,
					lbHMDelayInput:          fmt.Sprintf("%d", hmDelay),
					lbHMDelay:               hmDelay,
					lbHMMaxRetriesInput:     fmt.Sprintf("%d", hmMaxRetries),
					lbHMMaxRetries:          hmMaxRetries,
					lbHMMaxRetriesDownInput: fmt.Sprintf("%d", hmMaxRetriesDown),
					lbHMMaxRetriesDown:      hmMaxRetriesDown,
					lbHMTimeoutInput:        fmt.Sprintf("%d", hmTimeout),
					lbHMTimeout:             hmTimeout,
				}
			}
		}
		return m, nil

	case "h":
		// (LBPoolMembersView now uses action-based navigation — h kept for other views if needed)
		return m, nil

	case "enter":
		// Toggle between main nav and sub-nav when on Storage item
		if m.mode != DetailView && m.mode != ProjectSelectView &&
			m.mode != NodePoolsView && m.mode != NodePoolDetailView {
			navItems := getNavItems()
			if m.navIdx < len(navItems) && navItems[m.navIdx].Product == ProductStorage {
				if !m.inStorageSubNav {
					// Level 1 → Level 2: enter sub-nav focus
					m.inStorageSubNav = true
					m.inTableFocus = false
					return m.loadStorageSubProduct()
				} else if !m.inTableFocus {
					// Level 2 → Level 3: enter table focus
					m.inTableFocus = true
					return m, nil
				}
				// Level 3: fall through to normal enter handling
			}
			if m.navIdx < len(navItems) && navItems[m.navIdx].Product == ProductNetworks {
				if !m.inNetworkSubNav {
					// Level 1 → Level 2: enter sub-nav focus
					m.inNetworkSubNav = true
					m.inTableFocus = false
					return m.loadNetworkSubProduct()
				} else if !m.inTableFocus {
					// Level 2 → Level 3: enter table focus
					m.inTableFocus = true
					return m, nil
				}
				// Level 3: fall through to normal enter handling
			}
			if m.navIdx < len(navItems) && navItems[m.navIdx].Product == ProductCompute {
				if !m.inComputeSubNav {
					// Level 1 → Level 2: enter sub-nav focus
					m.inComputeSubNav = true
					m.inTableFocus = false
					return m.loadComputeSubProduct()
				} else if !m.inTableFocus {
					// Level 2 → Level 3: enter table focus
					m.inTableFocus = true
					return m, nil
				}
				// Level 3: fall through to normal enter handling
			}
		}
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
		} else if m.mode == LBPoolDetailView {
			switch m.lbPoolDetailActionIdx {
			case 0: // Edit — launch edit wizard
				region := getStringValue(m.detailData, "region", "")
				poolAlgo := getStringValue(m.selectedLBPool, "algorithm", "roundRobin")
				poolProto := getStringValue(m.selectedLBPool, "protocol", "tcp")
				algoIdx := 0
				protoIdx := 0
				for i, o := range lbPoolAlgoOptions {
					if o.value == poolAlgo {
						algoIdx = i
						break
					}
				}
				for i, o := range lbPoolProtoOptions {
					if o.value == poolProto {
						protoIdx = i
						break
					}
				}
				sessionVal := ""
				if sp, ok := m.selectedLBPool["sessionPersistence"].(map[string]interface{}); ok {
					sessionVal = getStringValue(sp, "type", "")
				}
				sessionIdx := 0
				for i, o := range lbPoolSessionOptions {
					if o.value == sessionVal {
						sessionIdx = i
						break
					}
				}
				poolName := getStringValue(m.selectedLBPool, "name", "")
				m.mode = WizardView
				m.wizard = WizardData{
					step:             LBPoolWizardStepName,
					lbPoolEditPoolId: getStringValue(m.selectedLBPool, "id", ""),
					lbPoolLBId:       getStringValue(m.detailData, "id", ""),
					lbPoolLBName:     getStringValue(m.detailData, "name", ""),
					lbPoolLBRegion:   region,
					lbPoolNameInput:  poolName,
					lbPoolName:       poolName,
					lbPoolAlgoIdx:    algoIdx,
					lbPoolAlgo:       poolAlgo,
					lbPoolProtoIdx:   protoIdx,
					lbPoolProto:      poolProto,
					lbPoolSessionIdx: sessionIdx,
					lbPoolSession:    sessionVal,
				}
				return m, nil
			case 1: // Delete
				if m.lbPoolDetailConfirm {
					m.lbPoolDetailConfirm = false
					return m, m.executeDeleteLBPool()
				}
				m.lbPoolDetailConfirm = true
			case 2: // Members
				poolID := getStringValue(m.selectedLBPool, "id", "")
				region := getStringValue(m.detailData, "region", "")
				m.mode = LBPoolMembersView
				m.lbPoolMemberDetailIdx = 0
				m.lbPoolMemberConfirm = false
				return m, m.fetchLBMembers(poolID, region)
			}
			return m, nil
			// LBPoolMembersView: Enter dispatches the selected action
		} else if m.mode == LBPoolMembersView {
			poolID := getStringValue(m.selectedLBPool, "id", "")
			region := getStringValue(m.detailData, "region", "")
			members := m.lbPoolMembers[poolID]
			idx := m.lbPoolMemberDetailIdx
			if idx < 0 {
				idx = 0
			}
			// Resolve actual action index (when no members, actions are 0=Create, 1=HM)
			actionIdx := m.lbMembersActionIdx
			if len(members) == 0 {
				// Remap: 0=Create stays 0, 1 is HM (maps to logical 3)
				if actionIdx == 1 {
					actionIdx = 3
				}
			}
			switch actionIdx {
			case 0: // Create
				m.lbMembersSection = 0
				m.mode = WizardView
				m.wizard = WizardData{
					step:                LBMemberWizardStepName,
					lbMemberPoolId:      poolID,
					lbMemberPoolRegion:  region,
					lbMemberWeightInput: "1",
				}
			case 1: // Edit
				if len(members) > 0 && idx < len(members) {
					mem := members[idx]
					memberID := getStringValue(mem, "id", "")
					memberName := getStringValue(mem, "name", "")
					memberAddr := getStringValue(mem, "address", "")
					memberPort := 0
					if v, ok := mem["protocolPort"]; ok {
						switch p := v.(type) {
						case float64:
							memberPort = int(p)
						case int:
							memberPort = p
						}
					}
					memberWeight := 1
					if v, ok := mem["weight"]; ok {
						switch w := v.(type) {
						case float64:
							memberWeight = int(w)
						case int:
							memberWeight = w
						}
					}
					m.lbMembersSection = 0
					m.mode = WizardView
					m.wizard = WizardData{
						step:                LBMemberWizardStepName,
						lbMemberPoolId:      poolID,
						lbMemberPoolRegion:  region,
						lbMemberEditId:      memberID,
						lbMemberNameInput:   memberName,
						lbMemberName:        memberName,
						lbMemberIPInput:     memberAddr,
						lbMemberIP:          memberAddr,
						lbMemberPortInput:   fmt.Sprintf("%d", memberPort),
						lbMemberPort:        memberPort,
						lbMemberWeightInput: fmt.Sprintf("%d", memberWeight),
						lbMemberWeight:      memberWeight,
					}
				}
			case 2: // Delete
				if len(members) > 0 && idx < len(members) {
					if m.lbPoolMemberConfirm {
						m.lbPoolMemberConfirm = false
						memberID := getStringValue(members[idx], "id", "")
						return m, m.executeDeleteLBMember(poolID, memberID, region)
					}
					m.lbPoolMemberConfirm = true
				}
			case 3: // Health Monitor
				m.lbMembersSection = 0
				m.mode = LBHealthMonitorView
				m.lbHMConfirm = false
				return m, m.fetchLBHealthMonitor(poolID, region)
			}
			return m, nil
			// LBHealthMonitorView: Enter → execute selected action button
		} else if m.mode == LBHealthMonitorView && m.selectedLBPool != nil {
			poolID := getStringValue(m.selectedLBPool, "id", "")
			region := getStringValue(m.detailData, "region", "")
			hm := m.lbHealthMonitors[poolID]
			if hm == nil {
				// No HM yet — Create
				m.mode = WizardView
				m.wizard = WizardData{
					step:                    LBHMWizardStepName,
					lbHMPoolId:              poolID,
					lbHMPoolRegion:          region,
					lbHMDelayInput:          "5",
					lbHMMaxRetriesInput:     "3",
					lbHMMaxRetriesDownInput: "3",
					lbHMTimeoutInput:        "5",
					lbHMHttpMethodIdx:       2,
					lbHMHttpMethod:          "GET",
					lbHMUrlPathInput:        "/",
					lbHMUrlPath:             "/",
					lbHMExpectedCodesInput:  "200",
					lbHMExpectedCodes:       "200",
				}
			} else {
				switch m.lbHMActionIdx {
				case 0: // Edit
					hmID := getStringValue(hm, "id", "")
					hmName := getStringValue(hm, "name", "")
					hmType := getStringValue(hm, "monitorType", "http")
					hmTypeIdx := 0
					for i, opt := range lbHMTypeOptions {
						if opt.value == hmType { hmTypeIdx = i; break }
					}
					hmDelay := 5
					if v, ok := hm["delay"]; ok { switch d := v.(type) { case float64: hmDelay = int(d); case int: hmDelay = d } }
					hmTimeout := 5
					if v, ok := hm["timeout"]; ok { switch d := v.(type) { case float64: hmTimeout = int(d); case int: hmTimeout = d } }
					hmMaxRetries := 3
					if v, ok := hm["maxRetries"]; ok { switch d := v.(type) { case float64: hmMaxRetries = int(d); case int: hmMaxRetries = d } }
					hmMaxRetriesDown := 3
					if v, ok := hm["maxRetriesDown"]; ok { switch d := v.(type) { case float64: hmMaxRetriesDown = int(d); case int: hmMaxRetriesDown = d } }
					hmHttpMethod := "GET"
					hmHttpMethodIdx := 2
					hmUrlPath := "/"
					hmExpectedCodes := "200"
					if httpCfg, ok := hm["httpConfiguration"].(map[string]interface{}); ok {
						if method, ok := httpCfg["httpMethod"].(string); ok && method != "" {
							hmHttpMethod = method
							for i, opt := range lbHMHttpMethodOptions { if opt == hmHttpMethod { hmHttpMethodIdx = i; break } }
						}
						if up, ok := httpCfg["urlPath"].(string); ok && up != "" { hmUrlPath = up }
						if ec, ok := httpCfg["expectedCodes"].(string); ok && ec != "" { hmExpectedCodes = ec }
					}
					m.mode = WizardView
					m.wizard = WizardData{
						step: LBHMWizardStepName, lbHMPoolId: poolID, lbHMPoolRegion: region, lbHMEditId: hmID,
						lbHMNameInput: hmName, lbHMName: hmName, lbHMTypeIdx: hmTypeIdx, lbHMType: hmType,
						lbHMHttpMethodIdx: hmHttpMethodIdx, lbHMHttpMethod: hmHttpMethod,
						lbHMUrlPathInput: hmUrlPath, lbHMUrlPath: hmUrlPath,
						lbHMExpectedCodesInput: hmExpectedCodes, lbHMExpectedCodes: hmExpectedCodes,
						lbHMDelayInput: fmt.Sprintf("%d", hmDelay), lbHMDelay: hmDelay,
						lbHMTimeoutInput: fmt.Sprintf("%d", hmTimeout), lbHMTimeout: hmTimeout,
						lbHMMaxRetriesInput: fmt.Sprintf("%d", hmMaxRetries), lbHMMaxRetries: hmMaxRetries,
						lbHMMaxRetriesDownInput: fmt.Sprintf("%d", hmMaxRetriesDown), lbHMMaxRetriesDown: hmMaxRetriesDown,
					}
				case 1: // Delete
					if m.lbHMConfirm {
						m.lbHMConfirm = false
						hmID := getStringValue(hm, "id", "")
						return m, m.deleteHealthMonitor(hmID, poolID, region)
					}
					m.lbHMConfirm = true
				}
			}
			return m, nil
			// LBL7RulesView: Enter → execute selected action button
		} else if m.mode == LBL7RulesView && m.selectedLBL7Policy != nil {
			policyID := getStringValue(m.selectedLBL7Policy, "id", "")
			policyName := getStringValue(m.selectedLBL7Policy, "name", "")
			region := getStringValue(m.detailData, "region", "")
			rules := m.lbL7Rules[policyID]
			switch m.lbL7RuleActionIdx {
			case 0: // Create
				m.mode = WizardView
				m.wizard = WizardData{
					step:             LBL7RuleWizardStepType,
					l7RulePolicyId:   policyID,
					l7RulePolicyName: policyName,
					l7RuleLBRegion:   region,
				}
			case 1: // Edit
				idx := m.lbL7RuleDetailIdx
				if idx < 0 { idx = 0 }
				if len(rules) > 0 && idx < len(rules) {
					r := rules[idx]
					ruleID := getStringValue(r, "id", "")
					ruleType := getStringValue(r, "ruleType", getStringValue(r, "type", ""))
					compareType := getStringValue(r, "compareType", "")
					ruleKey := getStringValue(r, "key", "")
					ruleValue := getStringValue(r, "value", "")
					ruleInvert := false
					if inv, ok := r["invert"].(bool); ok { ruleInvert = inv }
					typeIdx := 0
					for i, opt := range lbL7RuleTypeOptions {
						if opt.value == ruleType { typeIdx = i; break }
					}
					compareIdx := 0
					validOpts := validCompareOptionsForType(ruleType)
					for i, opt := range validOpts {
						if opt.value == compareType { compareIdx = i; break }
					}
					prefillKey := ""
					for _, t := range lbL7RuleTypeOptions {
						if t.value == ruleType && t.needsKey { prefillKey = ruleKey; break }
					}
					m.mode = WizardView
					m.wizard = WizardData{
						step:             LBL7RuleWizardStepType,
						l7RulePolicyId:   policyID,
						l7RulePolicyName: policyName,
						l7RuleLBRegion:   region,
						l7RuleEditId:     ruleID,
						l7RuleTypeIdx:    typeIdx,
						l7RuleType:       ruleType,
						l7RuleCompareIdx: compareIdx,
						l7RuleCompare:    compareType,
						l7RuleKey:        prefillKey,
						l7RuleValue:      ruleValue,
						l7RuleInvert:     ruleInvert,
					}
				}
			case 2: // Delete
				idx := m.lbL7RuleDetailIdx
				if idx < 0 { idx = 0 }
				if len(rules) > 0 && idx < len(rules) {
					if m.lbL7RuleConfirm {
						m.lbL7RuleConfirm = false
						ruleID := getStringValue(rules[idx], "id", "")
						return m, m.executeDeleteLBL7Rule(policyID, ruleID, region)
					}
					m.lbL7RuleConfirm = true
				}
			}
			return m, nil
		} else if m.mode == LBL7PolicyDetailView {
			switch m.lbL7PolicyDetailActionIdx {
			case 0: // Edit — launch edit wizard
				policyName := getStringValue(m.selectedLBL7Policy, "name", "")
				policyAction := getStringValue(m.selectedLBL7Policy, "action", "")
				policyPosition := 1
				if v, ok := m.selectedLBL7Policy["position"]; ok {
					switch vt := v.(type) {
					case float64:
						policyPosition = int(vt)
					case int:
						policyPosition = vt
					}
				}
				actionIdx := 0
				for i, opt := range lbL7PolicyActionOptions {
					if opt.value == policyAction {
						actionIdx = i
						break
					}
				}
				redirectPoolId := getStringValue(m.selectedLBL7Policy, "redirectPoolId", "")
				redirectUrl := getStringValue(m.selectedLBL7Policy, "redirectUrl", "")
				if redirectUrl == "" {
					redirectUrl = getStringValue(m.selectedLBL7Policy, "redirectPrefix", "")
				}
				redirectPoolIdx := 0
				if redirectPoolId != "" {
					allPools := m.lbPools[getStringValue(m.detailData, "id", "")]
					for j, p := range allPools {
						if getStringValue(p, "id", "") == redirectPoolId {
							redirectPoolIdx = j
							break
						}
					}
				}
				m.mode = WizardView
				m.wizard = WizardData{
					step:                    LBL7PolicyWizardStepName,
					l7PolicyEditId:          getStringValue(m.selectedLBL7Policy, "id", ""),
					l7PolicyListenerId:      getStringValue(m.selectedLBListener, "id", ""),
					l7PolicyListenerName:    getStringValue(m.selectedLBListener, "name", ""),
					l7PolicyLBRegion:        getStringValue(m.detailData, "region", ""),
					l7PolicyLBId:            getStringValue(m.detailData, "id", ""),
					l7PolicyNameInput:       policyName,
					l7PolicyName:            policyName,
					l7PolicyPositionInput:   fmt.Sprintf("%d", policyPosition),
					l7PolicyPosition:        policyPosition,
					l7PolicyActionIdx:       actionIdx,
					l7PolicyAction:          policyAction,
					l7PolicyRedirectPoolIdx: redirectPoolIdx,
					l7PolicyRedirectPoolId:  redirectPoolId,
					l7PolicyRedirectUrlInput: redirectUrl,
					l7PolicyRedirectUrl:     redirectUrl,
				}
				return m, nil
			case 1: // Delete
				if m.lbL7PolicyDetailConfirm {
					m.lbL7PolicyDetailConfirm = false
					return m, m.executeDeleteLBL7Policy()
				}
				m.lbL7PolicyDetailConfirm = true
			case 2: // L7 Rules
				policyID := getStringValue(m.selectedLBL7Policy, "id", "")
				region := getStringValue(m.detailData, "region", "")
				m.mode = LBL7RulesView
				if policyID != "" && region != "" {
					return m, m.fetchLBL7Rules(policyID, region)
				}
			}
			return m, nil
		} else if m.mode == LBListenerDetailView {
			// If a policy row is selected, Enter opens its detail view
			if listenerID := getStringValue(m.selectedLBListener, "id", ""); listenerID != "" {
				policies := m.lbL7Policies[listenerID]
				if m.lbL7PolicyListIdx >= 0 && m.lbL7PolicyListIdx < len(policies) {
					m.selectedLBL7Policy = policies[m.lbL7PolicyListIdx]
					m.lbL7PolicyDetailActionIdx = 0
					m.lbL7PolicyDetailConfirm = false
					m.mode = LBL7PolicyDetailView
					return m, nil
				}
			}
			switch m.lbListenerDetailActionIdx {
			case 0: // Edit — launch edit wizard
				listenerName := getStringValue(m.selectedLBListener, "name", "")
				listenerProto := getStringValue(m.selectedLBListener, "protocol", "tcp")
				protoIdx := 0
				for i, o := range lbListenerProtoOptions {
					if o.value == listenerProto {
						protoIdx = i
						break
					}
				}
				poolId := getStringValue(m.selectedLBListener, "defaultPoolId", "")
				poolIdx := 0
				if poolId != "" {
					allPools := m.lbPools[getStringValue(m.detailData, "id", "")]
					for _, p := range allPools {
						if lbPoolCompatible(getStringValue(p, "protocol", ""), listenerProto) {
							poolIdx++
							if getStringValue(p, "id", "") == poolId {
								break
							}
						}
					}
				}
				m.mode = WizardView
				m.wizard = WizardData{
					step:               LBListenerWizardStepName,
					lbListenerEditId:   getStringValue(m.selectedLBListener, "id", ""),
					lbListenerLBId:     getStringValue(m.detailData, "id", ""),
					lbListenerLBName:   getStringValue(m.detailData, "name", ""),
					lbListenerLBRegion: getStringValue(m.detailData, "region", ""),
					lbListenerNameInput: listenerName,
					lbListenerName:      listenerName,
					lbListenerProtoIdx:  protoIdx,
					lbListenerProto:     listenerProto,
					lbListenerPoolIdx:   poolIdx,
					lbListenerPoolId:    poolId,
				}
				return m, nil
			case 1: // Delete
				if m.lbListenerDetailConfirm {
					m.lbListenerDetailConfirm = false
					return m, m.executeDeleteLBListener()
				}
				m.lbListenerDetailConfirm = true
			case 2: // L7 Policies — launch L7 policy creation wizard
				m.mode = WizardView
				m.wizard = WizardData{
					step:                 LBL7PolicyWizardStepName,
					l7PolicyListenerId:   getStringValue(m.selectedLBListener, "id", ""),
					l7PolicyListenerName: getStringValue(m.selectedLBListener, "name", ""),
					l7PolicyLBRegion:     getStringValue(m.detailData, "region", ""),
					l7PolicyLBId:         getStringValue(m.detailData, "id", ""),
					l7PolicyPosition:     1,
					l7PolicyPositionInput: "1",
				}
				return m, nil
			}
			return m, nil
		} else if m.mode == DetailView && (m.currentProduct == ProductManagedDatabases || m.currentProduct == ProductManagedAnalytics) {
			switch m.dbDetailTab {
			case 0: // Service tab — Delete action
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.deleteManagedDBService()
				}
				m.actionConfirm = true
			case 1: // Users tab — create new user or dismiss result
				if m.dbUserCreatedData != nil {
					m.dbUserCreatedData = nil
				} else {
					m.dbUserCreateMode = true
					m.dbUserCreateInput = ""
				}
			}
			return m, nil
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
		} else if m.mode == DetailView && m.currentProduct == ProductNetworkGateway {
			switch m.selectedAction {
			case 0:
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeGatewayDelete()
				}
				m.actionConfirm = true
			}
			return m, nil
		} else if m.mode == DetailView && m.currentProduct == ProductNetworkLB {
			lbIDForPool := getStringValue(m.detailData, "id", "")
			// If a listener row is selected, open listener detail
			if listeners, ok := m.lbListeners[lbIDForPool]; ok && m.lbListenerListIdx >= 0 && m.lbListenerListIdx < len(listeners) {
				m.selectedLBListener = listeners[m.lbListenerListIdx]
				m.lbListenerDetailActionIdx = 0
				m.lbListenerDetailConfirm = false
				m.lbL7PolicyListIdx = -1
				m.mode = LBListenerDetailView
				listenerID := getStringValue(m.selectedLBListener, "id", "")
				region := getStringValue(m.detailData, "region", "")
				if listenerID != "" && region != "" {
					return m, m.fetchLBL7Policies(listenerID, region)
				}
				return m, nil
			}
			// If a pool row is selected in the list, Enter opens its detail view
			if pools, ok := m.lbPools[lbIDForPool]; ok && m.lbPoolListIdx >= 0 && m.lbPoolListIdx < len(pools) {
				m.selectedLBPool = pools[m.lbPoolListIdx]
				m.lbPoolDetailActionIdx = 0
				m.lbPoolDetailConfirm = false
				m.mode = LBPoolDetailView
				return m, nil
			}
			switch m.selectedAction {
			case 0: // Supprimer
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeLBDelete()
				}
				m.actionConfirm = true
			case 1: // Listeners — launch listener creation wizard
				m.mode = WizardView
				pools := m.lbPools[getStringValue(m.detailData, "id", "")]
				m.wizard = WizardData{
					step:             LBListenerWizardStepName,
					lbListenerLBId:   getStringValue(m.detailData, "id", ""),
					lbListenerLBName: getStringValue(m.detailData, "name", ""),
					lbListenerLBRegion: getStringValue(m.detailData, "region", ""),
				}
				// pre-load pools list into wizard for pool selection step
				_ = pools // pools accessible via m.lbPools at render time
				return m, nil
			case 2: // Pools — launch pool creation wizard
				m.mode = WizardView
				m.wizard = WizardData{
					step:           LBPoolWizardStepName,
					lbPoolLBId:     getStringValue(m.detailData, "id", ""),
					lbPoolLBName:   getStringValue(m.detailData, "name", ""),
					lbPoolLBRegion: getStringValue(m.detailData, "region", ""),
				}
				return m, nil
			}
			return m, nil
		} else if m.mode == DetailView && m.currentProduct == ProductNetworkPublic {
			switch m.selectedAction {
			case 0: // Supprimer
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeFIPDelete()
				}
				m.actionConfirm = true
			case 1: // Detach
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeFIPDetach()
				}
				m.actionConfirm = true
			}
			return m, nil
	} else if m.mode == DetailView && m.currentProduct == ProductWorkflow {
			switch m.selectedAction {
			case 0: // Supprimer
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeWorkflowDelete()
				}
				m.actionConfirm = true
			}
			return m, nil
		} else if m.mode == DetailView && m.currentProduct == ProductInstanceBackup {
			switch m.selectedAction {
			case 0: // Supprimer
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeInstanceBackupDelete()
				}
				m.actionConfirm = true
			}
			return m, nil
		} else if m.mode == DetailView && m.currentProduct == ProductNetworkPrivate {
			// Private Network detail actions
			switch m.selectedAction {
			case 0: // Delete
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executePrivNetworkDelete()
				}
				m.actionConfirm = true
			case 1: // Assign Gateway
				m.actionConfirm = false
				// Build region list from the network's available regions
				var regionNames []string
				regionMap := make(map[string]map[string]string)
				if regions, ok := m.detailData["regions"].([]interface{}); ok {
					for _, rv := range regions {
						rm, ok := rv.(map[string]interface{})
						if !ok {
							continue
						}
						regionName := getStringValue(rm, "region", "")
						openstackID := getStringValue(rm, "openstackId", "")
						if regionName == "" || openstackID == "" {
							continue
						}
						// Find subnet for this region's OpenStack network
						subnetID := ""
						if subnets, ok := m.detailData["_subnets"].([]map[string]any); ok {
							for _, s := range subnets {
								if getStringValue(s, "networkId", "") == openstackID {
									subnetID = getStringValue(s, "id", "")
									break
								}
							}
							// Fallback: just take first subnet
							if subnetID == "" && len(subnets) > 0 {
								subnetID = getStringValue(subnets[0], "id", "")
							}
						}
						regionNames = append(regionNames, regionName)
						regionMap[regionName] = map[string]string{
							"openstackId": openstackID,
							"subnetId":    subnetID,
						}
					}
				}
				if len(regionNames) == 0 {
					m.notification = "❌ No compatible region: OVH Gateway can only be added to subnets created without a gateway (mode 'OVH Gateway'). Recreate the network with this mode."
					m.notificationExpiry = time.Now().Add(10 * time.Second)
					return m, tea.Tick(10*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
				}
				netName := getStringValue(m.detailData, "name", "")
				if len(regionNames) == 1 {
					// Only one region: pre-select and go directly to model
					rd := regionMap[regionNames[0]]
					if rd["subnetId"] == "" {
						m.notification = "❌ This network has no subnet. Create a subnet first."
						m.notificationExpiry = time.Now().Add(5 * time.Second)
						return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg { return clearNotificationMsg{} })
					}
					m.mode = WizardView
					m.wizard = WizardData{
						step:               GwWizardStepModel,
						gwNetworkName:      netName,
						gwRegion:           regionNames[0],
						gwNetworkID:        rd["openstackId"],
						gwSubnetID:         rd["subnetId"],
						gwNetworkRegionMap: regionMap,
						gwAttachMode:       true,
					}
					return m, nil
				}
				// Multiple regions: let user choose
				m.mode = WizardView
				m.wizard = WizardData{
					step:               GwWizardStepRegion,
					gwNetworkName:      netName,
					gwAvailableRegions: regionNames,
					gwNetworkRegionMap: regionMap,
					gwAttachMode:       true,
				}
				return m, nil
			case 2: // Add Subnet
				m.actionConfirm = false
				netID := getStringValue(m.detailData, "id", "")
				if netID == "" {
					return m, nil
				}
				rType := getStringValue(m.detailData, "_regionType", "region")
				// Build set of regions that already have a subnet
				subnettedRegions := map[string]bool{}
				if subnets, ok := m.detailData["_subnets"].([]map[string]any); ok {
					for _, s := range subnets {
						if r := getStringValue(s, "region", ""); r != "" {
							subnettedRegions[r] = true
						}
					}
				}
				m.mode = WizardView
				m.wizard = WizardData{
					step:                    PrivNetWizardStepRegion,
					privNetAddSubnetMode:    true,
					privNetTargetNetworkID:  netID,
					privNetName:             getStringValue(m.detailData, "name", ""),
					privNetIsLocalZone:      rType == "localzone",
					privNetEnableDHCP:       true,
					privNetEnableSubnet:     true,
					privNetGatewayMode:      0,
					privNetCIDRInput:        "10.0.0.0/16",
					privNetSubnettedRegions: subnettedRegions,
					isLoading:               true,
					loadingMessage:          "Loading regions...",
				}
				return m, m.fetchPrivateNetRegionsCmd()
			case 3: // Delete Subnet
				subnets, _ := m.detailData["_subnets"].([]map[string]any)
				if len(subnets) == 0 {
					return m, nil
				}
				if m.privNetSelectedSubnet >= len(subnets) {
					m.privNetSelectedSubnet = 0
				}
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeSubnetDelete()
				}
				m.actionConfirm = true
			case 4: // Delete Region
				regions, _ := m.detailData["regions"].([]interface{})
				if len(regions) == 0 {
					return m, nil
				}
				if m.privNetSelectedRegion >= len(regions) {
					m.privNetSelectedRegion = 0
				}
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeRegionDelete()
				}
				m.actionConfirm = true
			case 5: // Detach Gateway
				if m.actionConfirm {
					m.actionConfirm = false
					return m, m.executeGatewayDetachFromNetwork()
				}
				m.actionConfirm = true
			}
			return m, nil
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
				return m, tea.Batch(
					m.fetchDataForPath("/instances"),
					m.scheduleRefresh(),
				)
			}
		} else if m.mode == NodePoolsView {
			// In node pools view, show node pool details
			clusterId := getStringValue(m.detailData, "id", "")
			nodePools := m.kubeNodePools[clusterId]
			if m.nodePoolsSelectedIdx >= 0 && m.nodePoolsSelectedIdx < len(nodePools) {
				m.selectedNodePool = nodePools[m.nodePoolsSelectedIdx]
				m.mode = NodePoolDetailView
			}
		}
		// Open detail view from table (replaces former 'v' key)
		if m.mode == TableView {
				isSubNavProd := (m.currentProduct >= ProductStorageBlock && m.currentProduct <= ProductNetworkLB)
			if !isSubNavProd || m.inTableFocus {
				selectedRow := m.table.Cursor()
				if selectedRow >= 0 && selectedRow < len(m.currentData) {
					m.detailData = m.currentData[selectedRow]
					m.currentItemName = getStringValue(m.detailData, "name", "Item")
					m.mode = DetailView
					m.selectedAction = 0
					m.actionConfirm = false

					if m.currentProduct == ProductStorageBlock {
						ctx := &views.Context{Width: m.width, Height: m.height}
						m.volumeDetailView = block_storage.NewDetailView(ctx, m.detailData)
						return m, nil
					}
					if m.currentProduct == ProductStorageFile {
						ctx := &views.Context{Width: m.width, Height: m.height}
						m.fileShareDetailView = file_storage.NewDetailView(ctx, m.detailData)
						return m, nil
					}
					if m.currentProduct == ProductStorageSnapshot {
						ctx := &views.Context{Width: m.width, Height: m.height}
						m.snapshotDetailView = block_storage.NewSnapshotDetailView(ctx, m.detailData)
						return m, nil
					}
					if m.currentProduct == ProductStorageBackup {
						ctx := &views.Context{Width: m.width, Height: m.height}
						m.backupDetailView = block_storage.NewBackupDetailView(ctx, m.detailData)
						return m, nil
					}
					if m.currentProduct == ProductStorageObject {
						if m.objectStorageTabIdx == 1 {
							selectedRow := m.table.Cursor()
							if selectedRow < 0 || selectedRow >= len(m.objectStorageUsers) {
								return m, nil
							}
							ctx := &views.Context{Width: m.width, Height: m.height}
							m.objectUserDetailView = object_storage.NewUserDetailView(ctx, m.objectStorageUsers[selectedRow])
							m.mode = DetailView
							return m, nil
						}
						ctx := &views.Context{Width: m.width, Height: m.height}
						m.objectDetailView = object_storage.NewDetailView(ctx, m.detailData, m.objectStorageUsers)
						return m, nil
					}
					if m.currentProduct == ProductKubernetes {
						kubeId := getStringValue(m.detailData, "id", "")
						if kubeId != "" {
							return m, m.fetchKubeNodePools(kubeId)
						}
					}
					if m.currentProduct == ProductNetworkPrivate {
						netId := getStringValue(m.detailData, "id", "")
						if netId != "" {
							return m, tea.Batch(
								m.fetchNetworkSubnets(netId),
								m.fetchPrivateNetworkDetail(netId),
							)
						}
					}
					if m.currentProduct == ProductNetworkLB {
						lbId := getStringValue(m.detailData, "id", "")
						lbRegion := getStringValue(m.detailData, "region", "")
						m.lbDetailSection = 0   // always start in Listeners section
						m.lbPoolListIdx = -1     // reset pool cursor when entering LB detail
						m.lbListenerListIdx = -1 // reset listener cursor when entering LB detail
						if lbId != "" && lbRegion != "" {
							return m, tea.Batch(
								m.fetchLBPools(lbId, lbRegion),
								m.fetchLBListeners(lbId, lbRegion),
							)
						}
					}
					if m.currentProduct == ProductManagedDatabases || m.currentProduct == ProductManagedAnalytics {
						engine := getStringValue(m.detailData, "engine", "")
						serviceId := getStringValue(m.detailData, "id", "")
						m.dbDetailUsers = nil
						m.dbDetailBackups = nil
						m.dbDetailDatabases = nil
						m.dbDetailPools = nil
						m.dbDetailTab = 0
						m.dbUserCreateMode = false
						m.dbUserCreateInput = ""
						m.dbUserCreatedData = nil
						if engine != "" && serviceId != "" {
							return m, m.fetchDBDetailSubresources(engine, serviceId)
						}
					}
				}
			}
		}
		return m, nil
	case "up", "down", "j", "k":
		key := msg.String()
		isStorageSubProduct := m.currentProduct >= ProductStorageBlock && m.currentProduct <= ProductStorageArchive
		isNetworkSubProduct := m.currentProduct >= ProductNetworkPrivate && m.currentProduct <= ProductNetworkLB
		isComputeSubProduct := m.currentProduct == ProductInstances || m.currentProduct == ProductInstanceBackup || m.currentProduct == ProductWorkflow
		isSubNavProduct := isStorageSubProduct || isNetworkSubProduct || isComputeSubProduct
		navItems := getNavItems()

		// Level 1 → Level 2: ↓ from main nav enters sub-nav for Storage / Networks / Compute
		if (key == "down" || key == "j") && !m.inStorageSubNav && !m.inNetworkSubNav && !m.inComputeSubNav && !m.inTableFocus &&
			m.mode != DetailView && m.mode != ProjectSelectView {
			if navItems[m.navIdx].Product == ProductStorage {
				m.inStorageSubNav = true
				m.inTableFocus = false
				return m.loadStorageSubProduct()
			}
			if navItems[m.navIdx].Product == ProductNetworks {
				m.inNetworkSubNav = true
				m.inTableFocus = false
				return m.loadNetworkSubProduct()
			}
			if navItems[m.navIdx].Product == ProductCompute {
				m.inComputeSubNav = true
				m.inTableFocus = false
				return m.loadComputeSubProduct()
			}
		}

		// In table focus (Level 3): up at row 0 → back to sub-nav focus (Level 2)
		if (key == "up" || key == "k") && m.inTableFocus && isSubNavProduct && m.mode != DetailView && !m.isDeepNavigationView() {
			if m.mode == TableView && m.table.Cursor() == 0 {
				m.inTableFocus = false
				return m, nil
			}
		}
		// In sub-nav focus (Level 2): up → back to main nav (Level 1)
		if (key == "up" || key == "k") && m.inStorageSubNav && !m.inTableFocus && m.mode != DetailView && !m.isDeepNavigationView() {
			m.inStorageSubNav = false
			return m, nil
		}
		if (key == "up" || key == "k") && m.inNetworkSubNav && !m.inTableFocus && m.mode != DetailView && !m.isDeepNavigationView() {
			m.inNetworkSubNav = false
			return m, nil
		}
		if (key == "up" || key == "k") && m.inComputeSubNav && !m.inTableFocus && m.mode != DetailView && !m.isDeepNavigationView() {
			m.inComputeSubNav = false
			return m, nil
		}
		// In sub-nav focus (Level 2): down → enter table focus (Level 3)
		if (key == "down" || key == "j") && isSubNavProduct && (m.inStorageSubNav || m.inNetworkSubNav || m.inComputeSubNav) && !m.inTableFocus && m.mode != DetailView && !m.isDeepNavigationView() {
			m.inTableFocus = true
			return m, nil
		}
		// Node pools list navigation
		if m.mode == NodePoolsView {
			clusterId := getStringValue(m.detailData, "id", "")
			nodePools := m.kubeNodePools[clusterId]
			if len(nodePools) > 0 {
				if key == "down" || key == "j" {
					if m.nodePoolsSelectedIdx < len(nodePools)-1 {
						m.nodePoolsSelectedIdx++
					}
				} else if key == "up" || key == "k" {
					if m.nodePoolsSelectedIdx > 0 {
						m.nodePoolsSelectedIdx--
					}
				}
			}
			return m, nil
		}
		// Private network detail: ↑/↓ to select subnet when Delete Subnet action is active
		if m.mode == DetailView && m.currentProduct == ProductNetworkPrivate && m.selectedAction == 3 {
			subnets, _ := m.detailData["_subnets"].([]map[string]any)
			if len(subnets) > 0 {
				if key == "down" || key == "j" {
					if m.privNetSelectedSubnet < len(subnets)-1 {
						m.privNetSelectedSubnet++
					}
				} else if key == "up" || key == "k" {
					if m.privNetSelectedSubnet > 0 {
						m.privNetSelectedSubnet--
					}
				}
			}
			return m, nil
		}
		// LB detail: ↑/↓ navigate continuously across Listeners then Pools
		if m.mode == DetailView && m.currentProduct == ProductNetworkLB {
			lbID := getStringValue(m.detailData, "id", "")
			listeners := m.lbListeners[lbID]
			pools := m.lbPools[lbID]

			if key == "down" || key == "j" {
				if m.lbDetailSection == 0 {
					// In Listeners section
					if m.lbListenerListIdx < len(listeners)-1 {
						m.lbListenerListIdx++
					} else {
						// Reached end of listeners → jump to Pools section
						m.lbDetailSection = 1
						m.lbListenerListIdx = -1
						if len(pools) > 0 {
							m.lbPoolListIdx = 0
						} else {
							m.lbPoolListIdx = -1
						}
					}
				} else {
					// In Pools section
					if m.lbPoolListIdx < len(pools)-1 {
						m.lbPoolListIdx++
					}
				}
			} else if key == "up" || key == "k" {
				if m.lbDetailSection == 1 {
					// In Pools section
					if m.lbPoolListIdx > 0 {
						m.lbPoolListIdx--
					} else {
						// Reached top of pools → jump back to Listeners section
						m.lbDetailSection = 0
						m.lbPoolListIdx = -1
						if len(listeners) > 0 {
							m.lbListenerListIdx = len(listeners) - 1
						} else {
							m.lbListenerListIdx = -1
						}
					}
				} else {
					// In Listeners section
					if m.lbListenerListIdx > 0 {
						m.lbListenerListIdx--
					} else {
						m.lbListenerListIdx = -1
					}
				}
			}
			return m, nil
		}
		// LBPoolMembersView: ↓ enters pagination from actions; ↑ returns to actions from pagination; in pagination ↑↓ navigate members
		if m.mode == LBPoolMembersView && m.selectedLBPool != nil {
			poolID := getStringValue(m.selectedLBPool, "id", "")
			count := len(m.lbPoolMembers[poolID])
			if (key == "down" || key == "j") && m.lbMembersSection == 0 {
				if count > 0 {
					m.lbMembersSection = 1
				}
			} else if (key == "up" || key == "k") && m.lbMembersSection == 1 {
				m.lbMembersSection = 0
			} else if m.lbMembersSection == 1 && count > 0 {
				if key == "down" || key == "j" {
					m.lbPoolMemberDetailIdx = (m.lbPoolMemberDetailIdx + 1) % count
				} else if key == "up" || key == "k" {
					if m.lbPoolMemberDetailIdx > 0 {
						m.lbPoolMemberDetailIdx--
					} else {
						m.lbPoolMemberDetailIdx = count - 1
					}
				}
			}
			return m, nil
		}
		// LBPoolDetailView / LBL7PolicyDetailView: ↑/↓ not used
		if m.mode == LBPoolDetailView || m.mode == LBL7PolicyDetailView {
			return m, nil
		}
		// LBListenerDetailView: ↑/↓ to navigate L7 policies list
		if m.mode == LBListenerDetailView {
			listenerID := getStringValue(m.selectedLBListener, "id", "")
			policies := m.lbL7Policies[listenerID]
			if key == "down" || key == "j" {
				if len(policies) > 0 && m.lbL7PolicyListIdx < len(policies)-1 {
					m.lbL7PolicyListIdx++
				}
			} else {
				if m.lbL7PolicyListIdx > 0 {
					m.lbL7PolicyListIdx--
				} else {
					m.lbL7PolicyListIdx = -1
				}
			}
			return m, nil
		}
		// Private network detail: ↑/↓ to select region when Delete Region action is active
		if m.mode == DetailView && m.currentProduct == ProductNetworkPrivate && m.selectedAction == 4 {
			regions, _ := m.detailData["regions"].([]interface{})
			if len(regions) > 0 {
				if key == "down" || key == "j" {
					if m.privNetSelectedRegion < len(regions)-1 {
						m.privNetSelectedRegion++
					}
				} else if key == "up" || key == "k" {
					if m.privNetSelectedRegion > 0 {
						m.privNetSelectedRegion--
					}
				}
			}
			return m, nil
		}
		// Table navigation: for sub-nav products requires Level 3 (inTableFocus); non-sub-nav always allowed
		if m.mode == TableView || m.mode == ProjectSelectView {
			if !isSubNavProduct || m.inTableFocus {
				var cmd tea.Cmd
				m.table, cmd = m.table.Update(msg)
				return m, cmd
			}
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
			return m, tea.Batch(
				m.fetchDataForPath(path),
				m.scheduleRefresh(),
			)
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
			return m, tea.Batch(
				m.fetchDataForPath(path),
				m.scheduleRefresh(),
			)
		}
		return m, nil

	case "d":
		// In LBL7RulesView: delete is handled via Enter on the Delete action button
		if false && m.mode == LBL7RulesView && m.selectedLBL7Policy != nil {
			policyID := getStringValue(m.selectedLBL7Policy, "id", "")
			rules := m.lbL7Rules[policyID]
			if len(rules) > 0 {
				idx := m.lbL7RuleDetailIdx
				if idx < 0 {
					idx = 0
				}
				if idx < len(rules) {
					if m.lbL7RuleConfirm {
						// Second press: execute delete
						m.lbL7RuleConfirm = false
						ruleID := getStringValue(rules[idx], "id", "")
						region := getStringValue(m.detailData, "region", "")
						return m, m.executeDeleteLBL7Rule(policyID, ruleID, region)
					}
					// First press: ask for confirm
					m.lbL7RuleConfirm = true
				}
			}
			return m, nil
		}
		// In LBHealthMonitorView: delete is handled via Enter on the Delete action button
		if false && m.mode == LBHealthMonitorView && m.selectedLBPool != nil {
			poolID := getStringValue(m.selectedLBPool, "id", "")
			region := getStringValue(m.detailData, "region", "")
			hm := m.lbHealthMonitors[poolID]
			if hm != nil {
				if m.lbHMConfirm {
					m.lbHMConfirm = false
					hmID := getStringValue(hm, "id", "")
					return m, m.deleteHealthMonitor(hmID, poolID, region)
				}
				m.lbHMConfirm = true
			}
			return m, nil
		}
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
		if m.mode == TableView {
			selectedRow := m.table.Cursor()
			if selectedRow >= 0 && selectedRow < len(m.currentData) {
				switch m.currentProduct {
				case ProductInstances:
					m.deleteTarget = m.currentData[selectedRow]
					m.deleteConfirmInput = ""
					m.mode = DeleteConfirmView
				case ProductKubernetes:
					cluster := m.currentData[selectedRow]
					clusterId := getStringValue(cluster, "id", "")
					clusterName := getStringValue(cluster, "name", "")
					if clusterId != "" {
						m.wizard.kubeDeleteClusterId = clusterId
						m.wizard.kubeDeleteClusterName = clusterName
						m.wizard.kubeDeleteConfirmInput = ""
						m.mode = KubeDeleteConfirmView
					}
				}
			}
		}
		return m, nil

	case "q", "ctrl+c":
		return m, tea.Quit
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
		case ProductStorageBlock:
			m.table = createBlockStorageTable(m.currentData, m.width, m.height)
		case ProductStorageFile:
			m.table = createFileStorageTable(m.currentData, m.width, m.height)
		case ProductStorageObject:
			m.table = createObjectStorageTable(m.currentData, m.width, m.height)
		case ProductNetworkPrivate:
			m.table = createPrivateNetworksTable(m.currentData, m.width, m.height)
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
	case ProductStorageObject:
		var filtered []map[string]interface{}
		for _, item := range m.currentData {
			name := strings.ToLower(getStringValue(item, "name", ""))
			region := strings.ToLower(getStringValue(item, "region", ""))
			if strings.Contains(name, filter) || strings.Contains(region, filter) {
				filtered = append(filtered, item)
			}
		}
		m.table = createObjectStorageTable(filtered, m.width, m.height)
	default:
		m.table = createGenericTable(m.currentData, m.width, m.height)
	}
}

// handleWizardKeyPress handles key presses in wizard mode
func (m Model) isDeepNavigationView() bool {
	switch m.mode {
	case DetailView, WizardView,
		NodePoolDetailView, NodePoolsView,
		LBPoolDetailView, LBListenerDetailView,
		LBL7PolicyDetailView, LBL7RulesView,
		LBPoolMembersView, LBHealthMonitorView:
		return true
	}
	return false
}

// isWizardTextInputStep returns true when the current wizard step is a free-form text or numeric
// input field, so that single-character shortcut keys (q, d, p, etc.) are suppressed.
func (m Model) isWizardTextInputStep() bool {
	if m.wizard.filterMode || m.wizard.creatingSSHKey || m.wizard.creatingNetwork {
		return true
	}
	switch m.wizard.step {
	case WizardStepName,
		KubeWizardStepName,
		NodePoolWizardStepName,
		VolumeWizardStepName,
		VolumeWizardStepSize,
		FileWizardStepName,
		FileWizardStepSize,
		ObjectWizardStepName,
		S3UserWizardStepDescription,
		BackupWizardStepName,
		PrivNetWizardStepName,
		PrivNetWizardStepVlanID,
		PrivNetWizardStepSubnet,
		PrivNetWizardStepAllocPool,
		PrivNetWizardStepGateway,
		GwWizardStepName,
		LBWizardStepName,
		LBPoolWizardStepName,
		LBListenerWizardStepName,
		LBListenerWizardStepPort,
		LBL7PolicyWizardStepName,
		LBL7PolicyWizardStepPosition,
		LBL7PolicyWizardStepRedirectUrl,
		LBL7RuleWizardStepKey,
		WorkflowWizardStepName,
		WorkflowWizardStepSchedule,
		LBMemberWizardStepName,
		LBMemberWizardStepIP,
		LBMemberWizardStepPort,
		LBMemberWizardStepWeight,
		LBHMWizardStepName,
		LBHMWizardStepUrlPath,
		LBHMWizardStepExpectedCodes,
		LBHMWizardStepDelay,
		LBHMWizardStepMaxRetries,
		LBHMWizardStepMaxRetriesDown,
		LBHMWizardStepTimeout,
		DBWizardStepName,
		DBWizardStepNodes,
		DBWizardStepStorage,
		AnalyticsWizardStepName,
		AnalyticsWizardStepNodes:
		return true
	}
	// Value step is text input only when not a bool-value type
	if m.wizard.step == LBL7RuleWizardStepValue && !m.isBoolValueType() {
		return true
	}
	return false
}

func (m Model) handleWizardKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Block all key input while an async operation is in progress.
	if m.wizard.isLoading {
		if key == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil
	}

	// Handle cleanup confirmation mode
	if m.wizard.cleanupPending {
		return m.handleCleanupConfirmKeys(key)
	}

	// ctrl+c always quits
	if key == "ctrl+c" {
		return m, tea.Quit
	}

	// 'q' quits (except when typing in input fields)
	if key == "q" && !m.isWizardTextInputStep() {
		return m, tea.Quit
	}

	// 'd' opens debug panel (except when typing in input fields)
	if key == "d" && !m.isWizardTextInputStep() {
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
		if m.wizard.step >= 1800 {
			// LB Health Monitor wizard: return to health monitor view
			m.wizard = WizardData{}
			m.mode = LBHealthMonitorView
			return m, nil
		} else if m.wizard.step >= 1700 {
			// LB Member wizard: return to pool members view
			m.wizard = WizardData{}
			m.mode = LBPoolMembersView
			return m, nil
		} else if m.wizard.step >= 1500 {
			// L7 Policy wizard: return to listener detail view
			m.wizard = WizardData{}
			m.mode = LBListenerDetailView
			m.lbL7PolicyListIdx = -1
			return m, nil
		} else if m.wizard.step >= 1400 {
			// LB Listener wizard: return to LB detail view
			m.wizard = WizardData{}
			m.mode = DetailView
			return m, nil
		} else if m.wizard.step >= 1300 {
			// LB Pool wizard: return to LB detail view
			m.wizard = WizardData{}
			m.mode = DetailView
			return m, nil
		} else if m.wizard.step >= 1200 {
			// Workflow wizard: return to workflow list
			returnPath = "/instances/workflow"
		} else if m.wizard.step >= 1100 {
			// Floating IP wizard: return to public IPs list
			returnPath = "/networks/floatingip"
		} else if m.wizard.step >= 1000 {
			// Load Balancer wizard: return to LB list
			returnPath = "/loadbalancer"
		} else if m.wizard.step >= 900 {
			// Gateway wizard: return to private network detail view
			m.wizard = WizardData{}
			m.mode = DetailView
			return m, nil
		} else if m.wizard.step >= 800 {
			returnPath = "/networks/private"
		} else if m.wizard.step >= 700 {
			returnPath = "/storage/backup"
		} else if m.wizard.step >= 500 {
			returnPath = "/storage/object"
		} else if m.wizard.step >= 400 {
			returnPath = "/storage/file"
		} else if m.wizard.step >= 300 {
			// Volume wizard
			returnPath = "/storage/block"
		} else if m.wizard.step >= 100 {
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
	// Volume wizard steps
	case VolumeWizardStepName:
		return m.handleVolumeWizardNameKeys(msg)
	case VolumeWizardStepRegion:
		return m.handleVolumeWizardRegionKeys(key, msg)
	case VolumeWizardStepType:
		return m.handleVolumeWizardTypeKeys(key, msg)
	case VolumeWizardStepAvailabilityZone:
		return m.handleVolumeWizardAZKeys(key, msg)
	case VolumeWizardStepSize:
		return m.handleVolumeWizardSizeKeys(msg)
	case VolumeWizardStepEncryption:
		return m.handleVolumeWizardEncryptionKeys(key)
	case VolumeWizardStepConfirm:
		return m.handleVolumeWizardConfirmKeys(key)
	// File Storage wizard steps
	case FileWizardStepName:
		return m.handleFileWizardNameKeys(msg)
	case FileWizardStepRegion:
		return m.handleFileWizardRegionKeys(key, msg)
	case FileWizardStepType:
		return m.handleFileWizardTypeKeys(key, msg)
	case FileWizardStepSize:
		return m.handleFileWizardSizeKeys(msg)
	case FileWizardStepNetwork:
		return m.handleFileWizardNetworkKeys(key, msg)
	case FileWizardStepConfirm:
		return m.handleFileWizardConfirmKeys(key)
	// Object Storage wizard steps
	case ObjectWizardStepName:
			return m.handleObjectWizardNameKeys(msg)
	case ObjectWizardStepType:
			return m.handleObjectWizardTypeKeys(key)
	case ObjectWizardStepRegion:
			return m.handleObjectWizardRegionKeys(key)
	case ObjectWizardStepReplication:
		switch key {
		case "left", "h", "y":
			m.wizard.objectReplication = true
		case "right", "l", "n":
			m.wizard.objectReplication = false
		case "enter":
			m.wizard.step = ObjectWizardStepVersioning
		case "esc":
			m.wizard.step = ObjectWizardStepRegion
		}
		return m, nil
	case ObjectWizardStepVersioning:
		switch key {
		case "left", "h", "y":
			m.wizard.objectVersioning = true
		case "right", "l", "n":
			m.wizard.objectVersioning = false
		case "enter":
			m.wizard.step = ObjectWizardStepObjectLock
		case "esc":
			m.wizard.step = ObjectWizardStepReplication
		}
		return m, nil
	case ObjectWizardStepObjectLock:
		switch key {
		case "left", "h", "y":
			m.wizard.objectLock = true
		case "right", "l", "n":
			m.wizard.objectLock = false
		case "enter":
			m.wizard.step = ObjectWizardStepUser
		case "esc":
			m.wizard.step = ObjectWizardStepVersioning
		}
		return m, nil
	case ObjectWizardStepUser:
			return m.handleObjectWizardUserKeys(key)
	case ObjectWizardStepEncryption:
		switch key {
		case "up", "k":
			m.wizard.objectEncryption = false
		case "down", "j":
			m.wizard.objectEncryption = true
		case "enter":
			m.wizard.step = ObjectWizardStepConfirm
		case "esc":
			m.wizard.step = ObjectWizardStepUser
		}
		return m, nil
	case ObjectWizardStepConfirm:
		return m.handleObjectWizardConfirmKeys(key)
	case ObjectWizardStepSwiftType:
		return m.handleObjectWizardSwiftTypeKeys(key)
	case ObjectWizardStepSwiftRegion:
		return m.handleObjectWizardSwiftRegionKeys(key)
	case S3UserWizardStepDescription:
		return m.handleS3UserWizardDescKeys(msg)
	case S3UserWizardStepConfirm:
		return m.handleS3UserWizardConfirmKeys(key)
	case BackupWizardStepVolume, BackupWizardStepType, BackupWizardStepName, BackupWizardStepConfirm:
		return m.handleBackupWizardKeys(msg)
	// Private Network wizard steps
	case PrivNetWizardStepRegion:
		return m.handlePrivNetWizardRegionKeys(key)
	case PrivNetWizardStepName:
		return m.handlePrivNetWizardNameKeys(msg)
	case PrivNetWizardStepVlanID:
		return m.handlePrivNetWizardVlanKeys(msg)
	case PrivNetWizardStepSubnet:
		return m.handlePrivNetWizardSubnetKeys(msg)
	case PrivNetWizardStepDHCP:
		return m.handlePrivNetWizardDHCPKeys(key)
	case PrivNetWizardStepAllocPool:
		return m.handlePrivNetWizardAllocPoolKeys(msg)
	case PrivNetWizardStepGateway:
		return m.handlePrivNetWizardGatewayKeys(msg)
	case PrivNetWizardStepConfirm:
		return m.handlePrivNetWizardConfirmKeys(key)
	// Gateway wizard steps
	case GwWizardStepRegion:
		return m.handleGwWizardRegionKeys(key)
	case GwWizardStepModel:
		return m.handleGwWizardModelKeys(key)
	case GwWizardStepName:
		return m.handleGwWizardNameKeys(msg)
	case GwWizardStepNetwork:
		return m.handleGwWizardNetworkKeys(key)
	case GwWizardStepConfirm:
		return m.handleGwWizardConfirmKeys(key)
	// Managed Database wizard steps
	case DBWizardStepName:
		return m.handleDBWizardNameKeys(msg)
	case DBWizardStepEngine:
		return m.handleDBWizardEngineKeys(key)
	case DBWizardStepVersion:
		return m.handleDBWizardVersionKeys(key)
	case DBWizardStepRegion:
		return m.handleDBWizardRegionKeys(key)
	case DBWizardStepPlan:
		return m.handleDBWizardPlanKeys(key)
	case DBWizardStepFlavor:
		return m.handleDBWizardFlavorKeys(key)
	case DBWizardStepNodes:
		return m.handleDBWizardNodesKeys(msg)
	case DBWizardStepStorage:
		return m.handleDBWizardStorageKeys(msg)
	case DBWizardStepNetwork:
		return m.handleDBWizardNetworkKeys(key)
	case DBWizardStepConfirm:
		return m.handleDBWizardConfirmKeys(key)
	// Managed Analytics wizard steps
	case AnalyticsWizardStepName:
		return m.handleAnalyticsWizardNameKeys(msg)
	case AnalyticsWizardStepEngine:
		return m.handleAnalyticsWizardEngineKeys(key)
	case AnalyticsWizardStepVersion:
		return m.handleAnalyticsWizardVersionKeys(key)
	case AnalyticsWizardStepRegion:
		return m.handleAnalyticsWizardRegionKeys(key)
	case AnalyticsWizardStepPlan:
		return m.handleAnalyticsWizardPlanKeys(key)
	case AnalyticsWizardStepFlavor:
		return m.handleAnalyticsWizardFlavorKeys(key)
	case AnalyticsWizardStepNodes:
		return m.handleAnalyticsWizardNodesKeys(msg)
	case AnalyticsWizardStepStorage:
		return m.handleAnalyticsWizardStorageKeys(msg)
	case AnalyticsWizardStepNetwork:
		return m.handleAnalyticsWizardNetworkKeys(key)
	case AnalyticsWizardStepConfirm:
		return m.handleAnalyticsWizardConfirmKeys(key)
	// Load Balancer wizard steps
	case LBWizardStepName:
		return m.handleLBWizardNameKeys(msg)
	case LBWizardStepRegion:
		return m.handleLBWizardRegionKeys(key)
	case LBWizardStepFlavor:
		return m.handleLBWizardFlavorKeys(key)
	case LBWizardStepNetwork:
		return m.handleLBWizardNetworkKeys(key)
	case LBWizardStepConfirm:
		return m.handleLBWizardConfirmKeys(key)
	// Floating IP wizard steps
	case FIPWizardStepRegion:
		return m.handleFIPWizardRegionKeys(key)
	case FIPWizardStepInstance:
		return m.handleFIPWizardInstanceKeys(key)
	case FIPWizardStepConfirm:
		return m.handleFIPWizardConfirmKeys(key)
	// LB Pool wizard steps
	case LBPoolWizardStepName:
		return m.handleLBPoolWizardNameKeys(key)
	case LBPoolWizardStepAlgo:
		return m.handleLBPoolWizardAlgoKeys(key)
	case LBPoolWizardStepProto:
		return m.handleLBPoolWizardProtoKeys(key)
	case LBPoolWizardStepSession:
		return m.handleLBPoolWizardSessionKeys(key)
	case LBPoolWizardStepConfirm:
		return m.handleLBPoolWizardConfirmKeys(key)
	// LB Listener wizard steps
	case LBListenerWizardStepName:
		return m.handleLBListenerWizardNameKeys(key)
	case LBListenerWizardStepProto:
		return m.handleLBListenerWizardProtoKeys(key)
	case LBListenerWizardStepPort:
		return m.handleLBListenerWizardPortKeys(key)
	case LBListenerWizardStepPool:
		return m.handleLBListenerWizardPoolKeys(key)
	case LBListenerWizardStepConfirm:
		return m.handleLBListenerWizardConfirmKeys(key)
	// L7 Policy wizard steps
	case LBL7PolicyWizardStepName:
		return m.handleLBL7PolicyWizardNameKeys(key)
	case LBL7PolicyWizardStepPosition:
		return m.handleLBL7PolicyWizardPositionKeys(key)
	case LBL7PolicyWizardStepAction:
		return m.handleLBL7PolicyWizardActionKeys(key)
	case LBL7PolicyWizardStepRedirectPool:
		return m.handleLBL7PolicyWizardRedirectPoolKeys(key)
	case LBL7PolicyWizardStepRedirectUrl:
		return m.handleLBL7PolicyWizardRedirectUrlKeys(key)
	case LBL7PolicyWizardStepConfirm:
		return m.handleLBL7PolicyWizardConfirmKeys(key)
	// L7 Rule wizard steps
	case LBL7RuleWizardStepType:
		return m.handleLBL7RuleWizardTypeKeys(key)
	case LBL7RuleWizardStepCompare:
		return m.handleLBL7RuleWizardCompareKeys(key)
	case LBL7RuleWizardStepKey:
		return m.handleLBL7RuleWizardKeyKeys(key)
	case LBL7RuleWizardStepValue:
		return m.handleLBL7RuleWizardValueKeys(key)
	case LBL7RuleWizardStepInvert:
		return m.handleLBL7RuleWizardInvertKeys(key)
	case LBL7RuleWizardStepConfirm:
		return m.handleLBL7RuleWizardConfirmKeys(key)
	// LB Member wizard steps
	case LBMemberWizardStepName:
		return m.handleLBMemberWizardNameKeys(key)
	case LBMemberWizardStepIP:
		return m.handleLBMemberWizardIPKeys(key)
	case LBMemberWizardStepPort:
		return m.handleLBMemberWizardPortKeys(key)
	case LBMemberWizardStepWeight:
		return m.handleLBMemberWizardWeightKeys(key)
	case LBMemberWizardStepConfirm:
		return m.handleLBMemberWizardConfirmKeys(key)
	// LB Health Monitor wizard steps
	case LBHMWizardStepName:
		return m.handleLBHMWizardNameKeys(key)
	case LBHMWizardStepType:
		return m.handleLBHMWizardTypeKeys(key)
	case LBHMWizardStepHttpMethod:
		return m.handleLBHMWizardHttpMethodKeys(key)
	case LBHMWizardStepUrlPath:
		return m.handleLBHMWizardUrlPathKeys(key)
	case LBHMWizardStepExpectedCodes:
		return m.handleLBHMWizardExpectedCodesKeys(key)
	case LBHMWizardStepDelay:
		return m.handleLBHMWizardDelayKeys(key)
	case LBHMWizardStepMaxRetries:
		return m.handleLBHMWizardMaxRetriesKeys(key)
	case LBHMWizardStepMaxRetriesDown:
		return m.handleLBHMWizardMaxRetriesDownKeys(key)
	case LBHMWizardStepTimeout:
		return m.handleLBHMWizardTimeoutKeys(key)
	case LBHMWizardStepConfirm:
		return m.handleLBHMWizardConfirmKeys(key)
	// Workflow wizard steps
	case WorkflowWizardStepType:
		return m.handleWorkflowWizardTypeKeys(key)
	case WorkflowWizardStepInstance:
		return m.handleWorkflowWizardInstanceKeys(key)
	case WorkflowWizardStepName:
		return m.handleWorkflowWizardNameKeys(key)
	case WorkflowWizardStepSchedule:
		return m.handleWorkflowWizardScheduleKeys(key)
	case WorkflowWizardStepConfirm:
		return m.handleWorkflowWizardConfirmKeys(key)
	}
	return m, nil
}

func (m Model) handleCleanupConfirmKeys(key string) (tea.Model, tea.Cmd) {
	switch key {
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
			m.wizard.selectedKubeRegion = m.wizard.kubeRegions[m.wizard.selectedIndex]
			m.wizard.selectedIndex = 0
			m.wizard.step = KubeWizardStepVersion
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading Kubernetes versions..."
			return m, m.fetchKubeVersions()
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
	m.detailData = nil
	m.currentData = nil
	m.inStorageSubNav = false
	m.inNetworkSubNav = false
	m.inComputeSubNav = false
	m.inTableFocus = false

	// For Networks, go to default sub-item (Private Networks = index 0)
	if currentNav.Product == ProductNetworks {
		m.networkSubIdx = 0
		return m.loadNetworkSubProduct()
	}

	// For Stockage, go to default sub-item (Block Storage = index 0)
	if currentNav.Product == ProductStorage {
		m.storageSubIdx = 0
		return m.loadStorageSubProduct()
	}

	// For Compute, go to default sub-item (Instances = index 0)
	if currentNav.Product == ProductCompute {
		m.computeSubIdx = 0
		return m.loadComputeSubProduct()
	}

	m.mode = LoadingView

	// For Kubernetes, start the auto-refresh timer
	if currentNav.Product == ProductKubernetes {
		return m, tea.Batch(
			m.fetchDataForPath(currentNav.Path),
			m.scheduleRefresh(),
		)
	}
	return m, m.fetchDataForPath(currentNav.Path)
}

func (m Model) loadStorageSubProduct() (Model, tea.Cmd) {
	subItems := getStorageSubItems()
	sub := subItems[m.storageSubIdx]
	m.currentProduct = sub.Product
	m.detailData = nil
	m.currentData = nil
	m.volumeDetailView = nil
	m.inTableFocus = false

	if !sub.Enabled {
		m.mode = ComingSoonView
		return m, nil
	}

	m.mode = LoadingView
	return m, m.fetchDataForPath(sub.Path)
}

func (m Model) loadNetworkSubProduct() (Model, tea.Cmd) {
	subItems := getNetworkSubItems()
	sub := subItems[m.networkSubIdx]
	m.currentProduct = sub.Product
	m.detailData = nil
	m.currentData = nil
	m.inTableFocus = false
	m.publicIPTabIdx = 0

	if !sub.Enabled {
		m.mode = ComingSoonView
		return m, nil
	}

	m.mode = LoadingView
	return m, m.fetchDataForPath(sub.Path)
}

func (m Model) loadComputeSubProduct() (Model, tea.Cmd) {
	subItems := getComputeSubItems()
	sub := subItems[m.computeSubIdx]
	m.currentProduct = sub.Product
	m.detailData = nil
	m.currentData = nil
	m.inTableFocus = false

	if !sub.Enabled {
		m.mode = ComingSoonView
		return m, nil
	}

	m.mode = LoadingView
	// Instances need the auto-refresh timer
	if sub.Product == ProductInstances {
		return m, tea.Batch(
			m.fetchDataForPath(sub.Path),
			m.scheduleRefresh(),
		)
	}
	return m, m.fetchDataForPath(sub.Path)
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

// ========== Volume Wizard Key Handlers ==========

func (m Model) handleVolumeWizardNameKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		name := strings.TrimSpace(m.wizard.volumeNameInput)
		if name == "" {
			m.wizard.errorMsg = "Volume name cannot be empty"
			return m, nil
		}
		m.wizard.volumeName = name
		m.wizard.errorMsg = ""
		m.wizard.step = VolumeWizardStepRegion
		m.wizard.selectedIndex = 0
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Loading regions..."
		return m, m.fetchVolumeRegions()
	case tea.KeyBackspace:
		if len(m.wizard.volumeNameInput) > 0 {
			m.wizard.volumeNameInput = m.wizard.volumeNameInput[:len(m.wizard.volumeNameInput)-1]
		}
	case tea.KeyRunes:
		m.wizard.volumeNameInput += string(msg.Runes)
	}
	return m, nil
}

func (m Model) handleVolumeWizardRegionKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}
	regions := m.wizard.regions
	switch key {
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(regions)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if len(regions) == 0 {
			return m, nil
		}
		selected := regions[m.wizard.selectedIndex]
		m.wizard.selectedRegion = getString(selected, "name")
		m.wizard.errorMsg = ""
		if types, ok := m.wizard.volumeRegionTypeMap[m.wizard.selectedRegion]; ok {
			m.wizard.volumeTypes = types
			m.wizard.volumeTypeAZMap = m.wizard.volumeRegionTypeAZMap[m.wizard.selectedRegion]
			m.wizard.step = VolumeWizardStepType
			m.wizard.selectedIndex = 0
		} else {
			// Fallback: fetch on demand
			m.wizard.step = VolumeWizardStepType
			m.wizard.selectedIndex = 0
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Loading volume types..."
			return m, m.fetchVolumeTypes(m.wizard.selectedRegion)
		}
	}
	return m, nil
}

func (m Model) handleVolumeWizardTypeKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}
	types := m.wizard.volumeTypes
	switch key {
	case "up", "k":
		if m.wizard.selectedIndex > 0 {
			m.wizard.selectedIndex--
		}
	case "down", "j":
		if m.wizard.selectedIndex < len(types)-1 {
			m.wizard.selectedIndex++
		}
	case "enter":
		if len(types) == 0 {
			return m, nil
		}
		m.wizard.volumeType = types[m.wizard.selectedIndex]
		m.wizard.errorMsg = ""
		m.wizard.selectedIndex = 0
		m.wizard.volumeAvailabilityZones = nil
		m.wizard.volumeAvailabilityZone = ""
		m.wizard.step = VolumeWizardStepAvailabilityZone
		m.wizard.isLoading = true
		m.wizard.loadingMessage = "Chargement des zones..."
		return m, m.fetchVolumeAvailabilityZones(m.wizard.selectedRegion)
	case "left":
		m.wizard.step = VolumeWizardStepRegion
		m.wizard.selectedIndex = 0
	}
	return m, nil
}

func (m Model) handleVolumeWizardAZKeys(key string, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}
	// Items = "(No preference)" + actual AZs
	totalItems := 1 + len(m.wizard.volumeAvailabilityZones)
	switch key {
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
			m.wizard.volumeAvailabilityZone = "" // No preference
		} else {
			m.wizard.volumeAvailabilityZone = m.wizard.volumeAvailabilityZones[m.wizard.selectedIndex-1]
		}
		m.wizard.errorMsg = ""
		m.wizard.step = VolumeWizardStepSize
		if m.wizard.volumeSize > 0 {
			m.wizard.volumeSizeInput = fmt.Sprintf("%d", m.wizard.volumeSize)
		} else {
			m.wizard.volumeSizeInput = ""
		}
	case "left":
		m.wizard.step = VolumeWizardStepType
		m.wizard.selectedIndex = 0
	}
	return m, nil
}

func (m Model) handleVolumeWizardSizeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		sizeStr := strings.TrimSpace(m.wizard.volumeSizeInput)
		size, err := strconv.Atoi(sizeStr)
		if err != nil || size < 10 || size > 12000 {
			m.wizard.errorMsg = "Size must be between 10 and 12000 GB"
			return m, nil
		}
		m.wizard.volumeSize = size
		m.wizard.errorMsg = ""
		m.wizard.step = VolumeWizardStepEncryption
	case tea.KeyBackspace:
		if len(m.wizard.volumeSizeInput) > 0 {
			m.wizard.volumeSizeInput = m.wizard.volumeSizeInput[:len(m.wizard.volumeSizeInput)-1]
		}
	case tea.KeyLeft:
		// Go back to AZ step only if that step was actually shown (type had AZ choices)
		if len(m.wizard.volumeAvailabilityZones) > 0 {
			m.wizard.step = VolumeWizardStepAvailabilityZone
		} else {
			m.wizard.step = VolumeWizardStepType
		}
		m.wizard.selectedIndex = 0
	case tea.KeyRunes:
		for _, r := range msg.Runes {
			if r >= '0' && r <= '9' {
				m.wizard.volumeSizeInput += string(r)
			}
		}
	}
	return m, nil
}

func (m Model) handleVolumeWizardConfirmKeys(key string) (tea.Model, tea.Cmd) {
	if m.wizard.isLoading {
		return m, nil
	}
	switch key {
	case "right", "tab":
		if m.wizard.volumeConfirmBtnIdx == 0 {
			m.wizard.volumeConfirmBtnIdx = 1
		} else {
			m.wizard.volumeConfirmBtnIdx = 0
		}
	case "enter":
		if m.wizard.volumeConfirmBtnIdx == 0 {
			// Create the volume
			m.wizard.isLoading = true
			m.wizard.loadingMessage = "Creating volume..."
			return m, m.createVolume()
		}
		// Cancel
		m.wizard = WizardData{}
		m.mode = LoadingView
		return m, m.fetchDataForPath("/storage/block")
	case "left", "esc":
		m.wizard.step = VolumeWizardStepEncryption
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

// handleKubeKubeconfigPickerKeyPress handles keyboard input for the kubeconfig directory picker.
func (m Model) handleKubeKubeconfigPickerKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	totalEntries := 2 + len(m.wizard.kubeKubeconfigEntries) // "..", "[Save here]", dirs

	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit

	case "esc":
		m.mode = DetailView
		return m, nil

	case "up", "k":
		if m.wizard.kubeKubeconfigSelectedIdx > 0 {
			m.wizard.kubeKubeconfigSelectedIdx--
		}
		return m, nil

	case "down", "j":
		if m.wizard.kubeKubeconfigSelectedIdx < totalEntries-1 {
			m.wizard.kubeKubeconfigSelectedIdx++
		}
		return m, nil

	case "enter":
		idx := m.wizard.kubeKubeconfigSelectedIdx
		if idx == 0 {
			// Navigate to parent directory
			parent := filepath.Dir(m.wizard.kubeKubeconfigCurrentDir)
			m.wizard.kubeKubeconfigCurrentDir = parent
			m.wizard.kubeKubeconfigEntries = listSubdirs(parent)
			m.wizard.kubeKubeconfigSelectedIdx = 0
		} else if idx == 1 {
			// Save here: trigger download
			clusterName := getStringValue(m.detailData, "name", "Unknown")
			destDir := m.wizard.kubeKubeconfigCurrentDir
			clusterId := m.wizard.kubeKubeconfigClusterId
			m.mode = DetailView
			return m, func() tea.Msg {
				return m.downloadKubeconfigToPath(clusterId, clusterName, destDir)
			}
		} else {
			// Navigate into selected subdirectory
			newDir := filepath.Join(m.wizard.kubeKubeconfigCurrentDir, m.wizard.kubeKubeconfigEntries[idx-2])
			m.wizard.kubeKubeconfigCurrentDir = newDir
			m.wizard.kubeKubeconfigEntries = listSubdirs(newDir)
			m.wizard.kubeKubeconfigSelectedIdx = 0
		}
		return m, nil
	}

	return m, nil
}

// listSubdirs returns a sorted list of subdirectory names in dir (non-hidden only).
func listSubdirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			dirs = append(dirs, e.Name())
		}
	}
	return dirs
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

