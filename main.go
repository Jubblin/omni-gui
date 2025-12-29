package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/siderolabs/omni/client/pkg/client"
	omniresources "github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"

	omniclient "github.com/jubblin/omni-api/internal/client"
	"github.com/jubblin/omni-api/internal/i18n"
	resconverter "github.com/jubblin/omni-api/internal/resource"
)

// Version is set at build time via ldflags
var Version = "dev"

const (
	clusterNodePrefixMachineSets     = "machinesets-"
	clusterNodePrefixClusterMachines = "clustermachines-"
	labelKeyMachineSet               = "omni.sidero.dev/machine-set"
	resourceTypePrefix               = "resource-type-"
)

type ResourceQuery struct {
	Type   resource.Type
	IsList bool
}

type TreeNode struct {
	ID       widget.TreeNodeID
	Type     string
	Label    string
	Resource map[string]interface{}
	Children []*TreeNode
}

type AppState struct {
	stateClient              state.State
	selectedResourceType     string
	treeRoot                 *TreeNode
	resourceTree             *widget.Tree
	resourceIDInput          *widget.Entry
	statusLabel              *widget.Label
	detailTitle              *widget.Label
	detailText               *widget.RichText
	currentResource          map[string]interface{}
	machineLinksContainer    *fyne.Container
	versionLinksContainer    *fyne.Container
	machineSetLinksContainer *fyne.Container
	machineRelatedLinksContainer *fyne.Container
	resourceActionsContainer *fyne.Container
}

type DetailComponents struct {
	Title              *widget.Label
	Text               *widget.RichText
	MachineLinks       *fyne.Container
	VersionLinks       *fyne.Container
	MachineSetLinks    *fyne.Container
	MachineRelatedLinks *fyne.Container
	ResourceActions    *fyne.Container
}

type ResourceContext struct {
	AppState    *AppState
	StateClient state.State
	Ctx         context.Context
	Depth       int
}

func getResourceQueries() map[string]ResourceQuery {
	return map[string]ResourceQuery{
		i18n.T("resource.cluster"): {
			Type:   omni.ClusterType,
			IsList: true,
		},
		i18n.T("resource.cluster_machines"): {
			Type:   omni.ClusterMachineType,
			IsList: true,
		},
		i18n.T("resource.etcd_backups"): {
			Type:   omni.EtcdBackupType,
			IsList: true,
		},
		i18n.T("resource.kubernetes_versions"): {
			Type:   omni.KubernetesVersionType,
			IsList: true,
		},
		i18n.T("resource.machine"): {
			Type:   omni.MachineType,
			IsList: true,
		},
		i18n.T("resource.machine_classes"): {
			Type:   omni.MachineClassType,
			IsList: true,
		},
		i18n.T("resource.machine_set"): {
			Type:   omni.MachineSetType,
			IsList: true,
		},
		i18n.T("resource.ongoing_tasks"): {
			Type:   omni.OngoingTaskType,
			IsList: true,
		},
		i18n.T("resource.schematics"): {
			Type:   omni.SchematicType,
			IsList: true,
		},
	}
}

func initializeApp(myWindow fyne.Window, ignoreEnv bool) (*client.Client, *AppState) {
	slog.Info("Creating Omni client")
	omniClient, err := omniclient.NewOmniClient(ignoreEnv)
	if err != nil {
		slog.Error("Failed to create Omni client", "error", err)
		// Format directive is in translation file: "app.error.client" = "Error: %v\n\nPlease set OMNI_ENDPOINT..."
		errorLabel := widget.NewLabel(i18n.T("app.error.client", err)) //nolint
		errorLabel.Wrapping = fyne.TextWrapWord
		myWindow.SetContent(container.NewScroll(errorLabel))
		slog.Info("Error content set on window")
		return nil, nil
	}
	slog.Info("Omni client created successfully")

	// Create root with resource types as top-level nodes
	root := buildInitialTree()
	setRootNode(root)
	addNodeToMap(root)
	
	appState := &AppState{
		stateClient: omniClient.Omni().State(),
		treeRoot:    root,
	}

	return omniClient, appState
}

func buildInitialTree() *TreeNode {
	resourceQueries := getResourceQueries()
	children := make([]*TreeNode, 0, len(resourceQueries))
	
	for name, query := range resourceQueries {
		// Create a node for each resource type
		// Use a special type prefix to identify resource type nodes
		nodeID := fmt.Sprintf("%s%s", resourceTypePrefix, string(query.Type))
		child := &TreeNode{
			ID:       nodeID,
			Type:     fmt.Sprintf("%s%s", resourceTypePrefix, string(query.Type)),
			Label:    name,
			Resource: map[string]interface{}{
				"resource_type": string(query.Type),
				"is_list":       query.IsList,
			},
			Children: []*TreeNode{}, // Will be loaded lazily when expanded
		}
		children = append(children, child)
	}
	
	return &TreeNode{
		ID:       "",
		Type:     "root",
		Label:    "",
		Resource: nil,
		Children: children,
	}
}

func createResourceIDInput() *widget.Entry {
	entry := widget.NewEntry()
	entry.SetPlaceHolder(i18n.T("resource.id.placeholder"))
	return entry
}

// Tree node storage - map for O(1) lookup
var nodeMap = make(map[widget.TreeNodeID]*TreeNode)

func createResourceTree(appState *AppState) *widget.Tree {
	// Clear node map when creating new tree
	nodeMap = make(map[widget.TreeNodeID]*TreeNode)
	
	tree := widget.NewTree(
		// Child IDs callback - returns list of child IDs for a given node ID
		func(id widget.TreeNodeID) []widget.TreeNodeID {
			node := getNodeByID(id)
			if node == nil {
				return []widget.TreeNodeID{}
			}
			childIDs := make([]widget.TreeNodeID, 0, len(node.Children))
			for _, child := range node.Children {
				childIDs = append(childIDs, child.ID)
			}
			return childIDs
		},
		// Is branch callback - determines if a node can be expanded
		func(id widget.TreeNodeID) bool {
			node := getNodeByID(id)
			if node == nil {
				return false
			}
			// If it has children, it's a branch
			if len(node.Children) > 0 {
				return true
			}
			// If it can have children (even if not loaded), it's a branch
			return canHaveChildren(node.Type)
		},
		// Create widget callback - creates the UI widget for a node
		func(branch bool) fyne.CanvasObject {
			// Create a container with icon and label
			icon := widget.NewIcon(nil)
			// Set a minimum size for the icon so it's visible
			icon.Resize(fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize()))
			label := widget.NewLabel("")
			return container.NewHBox(icon, label)
		},
		// Update widget callback - updates the UI widget with node data
		func(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
			// The container is a HBox, access its objects
			box := obj.(*fyne.Container)
			icon := box.Objects[0].(*widget.Icon)
			label := box.Objects[1].(*widget.Label)
			
			// Empty ID is the invisible root - don't show anything
			if id == "" {
				icon.SetResource(nil)
				label.SetText("")
				return
			}
			
			node := getNodeByID(id)
			if node != nil {
				label.SetText(node.Label)
				// Set icon based on resource type - get theme from current app
				app := fyne.CurrentApp()
				if app != nil {
					themeInstance := app.Settings().Theme()
					if themeInstance != nil {
						iconResource := getIconForResourceType(themeInstance, node.Type, branch)
						if iconResource != nil {
							icon.SetResource(iconResource)
							icon.Refresh()
						} else {
							slog.Debug("Icon resource is nil", "resource_type", node.Type, "is_branch", branch)
							icon.SetResource(nil)
						}
					} else {
						slog.Debug("Theme instance is nil")
						icon.SetResource(nil)
					}
				} else {
					slog.Debug("Current app is nil")
					icon.SetResource(nil)
				}
			} else {
				icon.SetResource(nil)
				label.SetText("")
			}
		},
	)
	return tree
}

func getNodeByID(id widget.TreeNodeID) *TreeNode {
	// Empty string is the root
	if id == "" {
		return getRootNode()
	}
	// Look up in map
	return nodeMap[id]
}

func getRootNode() *TreeNode {
	// This will be set in appState.treeRoot
	// We need to access it through a global or pass it differently
	// For now, we'll use a different approach - store root separately
	return rootNode
}

var rootNode *TreeNode

func setRootNode(node *TreeNode) {
	rootNode = node
	if node != nil {
		nodeMap[""] = node
	}
}

func addNodeToMap(node *TreeNode) {
	if node != nil && node.ID != "" {
		nodeMap[node.ID] = node
		// Also add all children recursively
		for _, child := range node.Children {
			addNodeToMap(child)
		}
	}
}

func canHaveChildren(resourceType string) bool {
	// Clusters and MachineSets folders can have children
	if resourceType == "clusters-folder" || resourceType == "machinesets-folder" {
		return true
	}
	
	// Link nodes are always leaves
	if strings.HasPrefix(resourceType, "link-") {
		return false
	}
	
	switch resourceType {
	case string(omni.ClusterType), string(omni.MachineSetType), string(omni.ClusterMachineType), string(omni.MachineType):
		return true
	case string(omni.KubernetesVersionType):
		return true // Can show related resources with this version
	case string(omni.MachineStatusType):
		return false // MachineStatus is a leaf node
	case "placeholder":
		return false
	default:
		return false
	}
}

func getIconForResourceType(themeInstance fyne.Theme, resourceType string, isBranch bool) fyne.Resource {
	if themeInstance == nil {
		slog.Warn("Theme instance is nil in getIconForResourceType")
		return nil
	}
	
	// Clusters and MachineSets folders always use folder icon
	if resourceType == "clusters-folder" || resourceType == "machinesets-folder" {
		return themeInstance.Icon(theme.IconNameFolder)
	}
	
	// Use theme icons for different resource types
	// Get the icon resource from the theme using the theme package constants
	switch resourceType {
	case string(omni.ClusterType):
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameComputer)
	case string(omni.MachineSetType):
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameStorage)
	case string(omni.ClusterMachineType):
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameComputer)
	case string(omni.MachineType):
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameComputer)
	case string(omni.MachineStatusType):
		return themeInstance.Icon(theme.IconNameInfo)
	case string(omni.KubernetesVersionType):
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameDocument)
	case "placeholder":
		return themeInstance.Icon(theme.IconNameFile)
	default:
		// Link nodes use document icon
		if strings.HasPrefix(resourceType, "link-") {
			return themeInstance.Icon(theme.IconNameDocument)
		}
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameFile)
	}
}

func setupTreeSelection(resourceTree *widget.Tree, appState *AppState) {
	// Handle node selection - show details in detail pane or execute link action
	resourceTree.OnSelected = func(id widget.TreeNodeID) {
		if id == "" {
			return
		}
		node := getNodeByID(id)
		if node == nil {
			return
		}
		
		// Check if this is a link node
		if strings.HasPrefix(node.Type, "link-") {
			// Execute the link action
			if action, ok := linkActionMap[id]; ok {
				action()
			}
			return
		}
		
		// Regular resource node - show details
		if node.Resource != nil {
			updateDetailPane(node.Resource, appState)
		}
	}
	
	// Lazy load children when a branch is expanded
	resourceTree.OnBranchOpened = func(id widget.TreeNodeID) {
		if id == "" {
			return
		}
		node := getNodeByID(id)
		if node == nil {
			return
		}
		// Only load if children haven't been loaded yet
		if len(node.Children) == 0 && node.Resource != nil {
			ctx := context.Background()
			loadNodeChildren(node, appState, ctx)
			// Add newly loaded children to the map
			addNodeToMap(node)
			// Refresh tree to show new children
			resourceTree.Refresh()
		}
	}
}

func loadNodeChildren(node *TreeNode, appState *AppState, ctx context.Context) {
	if node.Resource == nil {
		return
	}
	
	// Check if this is a resource type node (like "resource-type-clusters")
	if strings.HasPrefix(node.Type, resourceTypePrefix) {
		loadResourceTypeChildren(node, appState, ctx)
		return
	}
	
	resourceID, resourceType := extractResourceInfoFromMap(node.Resource)
	slog.Info("Loading children for node", "resource_type", resourceType, "resource_id", resourceID)
	
	// Load regular children first
	switch resourceType {
	case string(omni.ClusterType):
		loadClusterChildren(node, appState, ctx, resourceID)
	case string(omni.MachineSetType):
		loadMachineSetChildren(node, appState, ctx, resourceID)
	case string(omni.ClusterMachineType):
		loadClusterMachineChildren(node, appState, ctx, resourceID)
	case string(omni.MachineType):
		loadMachineChildren(node, appState, ctx, resourceID)
	case string(omni.KubernetesVersionType):
		loadKubernetesVersionChildren(node, appState, ctx, resourceID)
	}
	
	// Add link nodes as children after loading regular children
	addLinkNodes(node, appState, ctx)
}

// groupMachineSetsIntoFolder groups MachineSet nodes into a folder and sorts them by name
func groupMachineSetsIntoFolder(machineSetNodes []*TreeNode, parentID string) *TreeNode {
	if len(machineSetNodes) == 0 {
		return nil
	}
	
	// Sort MachineSets by label (name)
	sort.Slice(machineSetNodes, func(i, j int) bool {
		return machineSetNodes[i].Label < machineSetNodes[j].Label
	})
	
	// Create folder node with unique ID based on parent
	folderID := fmt.Sprintf("machinesets-folder-%s", parentID)
	machineSetsFolder := &TreeNode{
		ID:    folderID,
		Type:  "machinesets-folder",
		Label: "MachineSets",
		Resource: map[string]interface{}{
			"type": "machinesets-folder",
		},
		Children: machineSetNodes,
	}
	
	// Add folder and its children to the map
	addNodeToMap(machineSetsFolder)
	
	return machineSetsFolder
}

func loadClusterChildren(node *TreeNode, appState *AppState, ctx context.Context, clusterID string) {
	children := make([]*TreeNode, 0)
	machineSetNodes := make([]*TreeNode, 0)
	
	// First, collect all MachineSets for this cluster and track their IDs
	machineSetIDs := make(map[string]bool)
	machineSetMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineSetType, "", resource.VersionUndefined)
	machineSetList, err := appState.stateClient.List(ctx, machineSetMD)
	if err == nil {
		for _, item := range machineSetList.Items {
			ms, ok := item.(*omni.MachineSet)
			if !ok {
				continue
			}
			labels := ms.Metadata().Labels()
			if labels != nil {
				if cluster, ok := labels.Get("omni.sidero.dev/cluster"); ok && cluster == clusterID {
					msID := ms.Metadata().ID()
					machineSetIDs[msID] = true
					resourceMap := resconverter.ToMap(ms)
					childNode := createResourceNodeFromMap(resourceMap, msID, string(omni.MachineSetType))
					machineSetNodes = append(machineSetNodes, childNode)
				}
			}
		}
	}
	
	// Group MachineSets into a folder if there are any
	if len(machineSetNodes) > 0 {
		machineSetsFolder := groupMachineSetsIntoFolder(machineSetNodes, clusterID)
		if machineSetsFolder != nil {
			children = append(children, machineSetsFolder)
		}
	}
	
	// Find ClusterMachines for this cluster, but only add those that don't belong to a MachineSet
	clusterMachineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, "", resource.VersionUndefined)
	clusterMachineList, err := appState.stateClient.List(ctx, clusterMachineMD)
	if err == nil {
		for _, item := range clusterMachineList.Items {
			cm, ok := item.(*omni.ClusterMachine)
			if !ok {
				continue
			}
			labels := cm.Metadata().Labels()
			if labels == nil {
				continue
			}
			
			// Check if this ClusterMachine belongs to the cluster
			cluster, ok := labels.Get("omni.sidero.dev/cluster")
			if !ok || cluster != clusterID {
				continue
			}
			
			// Only add ClusterMachine if it doesn't belong to a MachineSet
			// (ClusterMachines that belong to MachineSets will be shown under their MachineSet)
			if machineSet, ok := labels.Get(labelKeyMachineSet); ok && machineSet != "" {
				// This ClusterMachine belongs to a MachineSet, skip it at cluster level
				continue
			}
			
			// This ClusterMachine doesn't belong to a MachineSet, add it to cluster
			resourceMap := resconverter.ToMap(cm)
			enrichClusterMachineResource(resourceMap, string(omni.ClusterMachineType), cm.Metadata().ID(), appState.stateClient, ctx, 0)
			cmID := cm.Metadata().ID()
			childNode := createResourceNodeFromMap(resourceMap, cmID, string(omni.ClusterMachineType))
			children = append(children, childNode)
		}
	}
	
	// Add KubernetesVersion as a child if cluster has a kubernetes_version
	if kv, ok := node.Resource["kubernetes_version"].(string); ok && kv != "" {
		kvMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.KubernetesVersionType, kv, resource.VersionUndefined)
		kvRes, err := appState.stateClient.Get(ctx, kvMD)
		if err == nil {
			resourceMap := resconverter.ToMap(kvRes)
			childNode := createResourceNodeFromMap(resourceMap, kv, string(omni.KubernetesVersionType))
			children = append(children, childNode)
		}
	}
	
	node.Children = children
	slog.Info("Loaded cluster children", 
		"cluster_id", clusterID,
		"total_children", len(children),
		"machinesets", len(machineSetIDs),
		"orphaned_clustermachines", len(children)-len(machineSetIDs))
}

// addLinkNodes adds link nodes as children to the given node
func addLinkNodes(node *TreeNode, appState *AppState, ctx context.Context, resourceType, resourceID string) {
	links := make([]*TreeNode, 0)
	
	// Add machine ID links
	if machineIDs := findMachineIDs(node.Resource); len(machineIDs) > 0 {
		for _, machineID := range machineIDs {
			linkID := fmt.Sprintf("link-machine-%s", machineID)
			linkNode := createLinkNodeWithAction(linkID, fmt.Sprintf("Machine: %s", machineID), "machine", func() {
				loadMachineByID(machineID, appState)
			})
			links = append(links, linkNode)
		}
	}
	
	// Add Kubernetes version links
	if versions := findKubernetesVersions(node.Resource); len(versions) > 0 {
		for _, version := range versions {
			linkID := fmt.Sprintf("link-k8s-version-%s", version)
			linkNode := createLinkNodeWithAction(linkID, fmt.Sprintf("Kubernetes Version: %s", version), "k8s-version", func() {
				loadResourcesByK8sVersion(version, appState)
			})
			links = append(links, linkNode)
		}
	}
	
	// Add MachineSet links
	if machineSetIDs := findMachineSetIDs(node.Resource); len(machineSetIDs) > 0 {
		for _, machineSetID := range machineSetIDs {
			linkID := fmt.Sprintf("link-machineset-%s", machineSetID)
			linkNode := createLinkNodeWithAction(linkID, fmt.Sprintf("Machine Set: %s", machineSetID), "machineset", func() {
				loadMachinesByMachineSet(machineSetID, appState)
			})
			links = append(links, linkNode)
		}
	}
	
	// Add resource action links based on resource type
	switch resourceType {
	case string(omni.ClusterType):
		links = append(links, createActionLinkNodes(resourceID, appState)...)
	case string(omni.MachineType):
		links = append(links, createMachineActionLinkNodes(resourceID, appState)...)
	case string(omni.MachineSetType):
		links = append(links, createMachineSetActionLinkNodes(resourceID, appState)...)
	case string(omni.ClusterMachineType):
		links = append(links, createClusterMachineActionLinkNodes(resourceID, appState)...)
	}
	
	// Append link nodes to existing children
	node.Children = append(node.Children, links...)
}

func createLinkNode(id, label, linkType string, action func()) *TreeNode {
	// Store action in a way that can be retrieved later
	// We'll use a closure to capture the action
	return &TreeNode{
		ID:    id,
		Type:  fmt.Sprintf("link-%s", linkType),
		Label: label,
		Resource: map[string]interface{}{
			"link_type": linkType,
			"link_id":   id,
		},
		Children: []*TreeNode{},
	}
}

// linkActionMap stores actions for link nodes
var linkActionMap = make(map[string]func())

func createLinkNodeWithAction(id, label, linkType string, action func()) *TreeNode {
	linkActionMap[id] = action
	return createLinkNode(id, label, linkType, action)
}

func createActionLinkNodes(clusterID string, appState *AppState) []*TreeNode {
	links := []*TreeNode{
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-status-%s", clusterID), "Status", "cluster-status", func() { loadClusterStatus(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-metrics-%s", clusterID), "Metrics", "cluster-metrics", func() { loadClusterMetrics(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-bootstrap-%s", clusterID), "Bootstrap", "cluster-bootstrap", func() { loadClusterBootstrap(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-kubeconfig-%s", clusterID), "Kubeconfig", "cluster-kubeconfig", func() { loadClusterKubeconfig(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-k8s-upgrade-%s", clusterID), "Kubernetes Upgrade", "cluster-k8s-upgrade", func() { loadClusterKubernetesUpgrade(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-talos-upgrade-%s", clusterID), "Talos Upgrade", "cluster-talos-upgrade", func() { loadClusterTalosUpgrade(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-endpoints-%s", clusterID), "Endpoints", "cluster-endpoints", func() { loadClusterEndpoints(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-k8s-status-%s", clusterID), "Kubernetes Status", "cluster-k8s-status", func() { loadClusterKubernetesStatus(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-controlplane-%s", clusterID), "Control Plane Status", "cluster-controlplane", func() { loadClusterControlPlaneStatus(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-diagnostics-%s", clusterID), "Diagnostics", "cluster-diagnostics", func() { loadClusterDiagnostics(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-destroy-%s", clusterID), "Destroy Status", "cluster-destroy", func() { loadClusterDestroyStatus(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-workload-proxy-%s", clusterID), "Workload Proxy Status", "cluster-workload-proxy", func() { loadClusterWorkloadProxyStatus(clusterID, appState) }),
	}
	return links
}

func createMachineActionLinkNodes(machineID string, appState *AppState) []*TreeNode {
	links := []*TreeNode{
		createLinkNodeWithAction(fmt.Sprintf("link-machine-labels-%s", machineID), "Labels", "machine-labels", func() { loadMachineLabels(machineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machine-extensions-%s", machineID), "Extensions", "machine-extensions", func() { loadMachineExtensions(machineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machine-upgrade-%s", machineID), "Upgrade Status", "machine-upgrade", func() { loadMachineUpgradeStatus(machineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machine-metrics-%s", machineID), "Metrics", "machine-metrics", func() { loadMachineMetrics(machineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machine-config-diff-%s", machineID), "Config Diff", "machine-config-diff", func() { loadMachineConfigDiff(machineID, appState) }),
	}
	return links
}

func createMachineSetActionLinkNodes(machineSetID string, appState *AppState) []*TreeNode {
	links := []*TreeNode{
		createLinkNodeWithAction(fmt.Sprintf("link-machineset-status-%s", machineSetID), "Status", "machineset-status", func() { loadMachineSetStatus(machineSetID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machineset-destroy-%s", machineSetID), "Destroy Status", "machineset-destroy", func() { loadMachineSetDestroyStatus(machineSetID, appState) }),
	}
	return links
}

func createClusterMachineActionLinkNodes(clusterMachineID string, appState *AppState) []*TreeNode {
	links := []*TreeNode{
		createLinkNodeWithAction(fmt.Sprintf("link-clustermachine-status-%s", clusterMachineID), "Status", "clustermachine-status", func() { loadClusterMachineStatus(clusterMachineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-clustermachine-config-status-%s", clusterMachineID), "Config Status", "clustermachine-config-status", func() { loadClusterMachineConfigStatus(clusterMachineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-clustermachine-talos-version-%s", clusterMachineID), "Talos Version", "clustermachine-talos-version", func() { loadClusterMachineTalosVersion(clusterMachineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-clustermachine-config-%s", clusterMachineID), "Config", "clustermachine-config", func() { loadClusterMachineConfig(clusterMachineID, appState) }),
	}
	return links
}

func loadMachineSetChildren(node *TreeNode, appState *AppState, ctx context.Context, machineSetID string) {
	children := make([]*TreeNode, 0)
	
	// Find ClusterMachines for this MachineSet
	clusterMachines := findClusterMachinesByMachineSet(ctx, machineSetID, appState)
	for _, resourceMap := range clusterMachines {
		cmID, _ := extractResourceInfoFromMap(resourceMap)
		enrichClusterMachineResource(resourceMap, string(omni.ClusterMachineType), cmID, appState.stateClient, ctx, 0)
		childNode := createResourceNodeFromMap(resourceMap, cmID, string(omni.ClusterMachineType))
		children = append(children, childNode)
	}
	
	node.Children = children
	slog.Info("Loaded MachineSet children", "machineset_id", machineSetID, "children_count", len(children))
}

func loadClusterMachineChildren(node *TreeNode, appState *AppState, ctx context.Context, clusterMachineID string) {
	children := make([]*TreeNode, 0)
	
	// Find the Machine for this ClusterMachine
	if machineID, ok := node.Resource["machine_id"].(string); ok && machineID != "" {
		machineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, machineID, resource.VersionUndefined)
		machine, err := appState.stateClient.Get(ctx, machineMD)
		if err == nil {
			resourceMap := resconverter.ToMap(machine)
			enrichMachineResource(resourceMap, string(omni.MachineType), machineID, appState.stateClient, ctx, 0)
			childNode := createResourceNodeFromMap(resourceMap, machineID, string(omni.MachineType))
			children = append(children, childNode)
		}
	}
	
	// Find the Cluster for this ClusterMachine (from labels)
	if labels, ok := node.Resource["labels"].(map[string]interface{}); ok {
		if clusterID, ok := labels["omni.sidero.dev/cluster"].(string); ok && clusterID != "" {
			clusterMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterType, clusterID, resource.VersionUndefined)
			cluster, err := appState.stateClient.Get(ctx, clusterMD)
			if err == nil {
				resourceMap := resconverter.ToMap(cluster)
				childNode := createResourceNodeFromMap(resourceMap, clusterID, string(omni.ClusterType))
				children = append(children, childNode)
			}
		}
		
		// Find the MachineSet for this ClusterMachine
		if machineSetID, ok := labels[labelKeyMachineSet].(string); ok && machineSetID != "" {
			machineSetMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineSetType, machineSetID, resource.VersionUndefined)
			machineSet, err := appState.stateClient.Get(ctx, machineSetMD)
			if err == nil {
				resourceMap := resconverter.ToMap(machineSet)
				childNode := createResourceNodeFromMap(resourceMap, machineSetID, string(omni.MachineSetType))
				children = append(children, childNode)
			}
		}
	}
	
	node.Children = children
	slog.Info("Loaded ClusterMachine children", "clustermachine_id", clusterMachineID, "children_count", len(children))
}

func loadMachineChildren(node *TreeNode, appState *AppState, ctx context.Context, machineID string) {
	children := make([]*TreeNode, 0)
	
	// Find ClusterMachines that use this Machine
	// ClusterMachine ID is the machine ID
	clusterMachineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, machineID, resource.VersionUndefined)
	cm, err := appState.stateClient.Get(ctx, clusterMachineMD)
	if err == nil {
		resourceMap := resconverter.ToMap(cm)
		enrichClusterMachineResource(resourceMap, string(omni.ClusterMachineType), machineID, appState.stateClient, ctx, 0)
		childNode := createResourceNodeFromMap(resourceMap, machineID, string(omni.ClusterMachineType))
		children = append(children, childNode)
	}
	
	// Find MachineStatus for this Machine (MachineStatus has the same ID as Machine)
	machineStatusMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, machineID, resource.VersionUndefined)
	ms, err := appState.stateClient.Get(ctx, machineStatusMD)
	if err == nil {
		resourceMap := resconverter.ToMap(ms)
		statusID := fmt.Sprintf("%s-status", machineID)
		childNode := createResourceNodeFromMap(resourceMap, statusID, string(omni.MachineStatusType))
		// Set a descriptive label for MachineStatus
		if hostname, ok := resourceMap["hostname"].(string); ok && hostname != "" {
			childNode.Label = fmt.Sprintf("MachineStatus (%s)", hostname)
		} else {
			childNode.Label = "MachineStatus"
		}
		children = append(children, childNode)
	}
	
	node.Children = children
	slog.Info("Loaded Machine children", "machine_id", machineID, "children_count", len(children))
}

func loadKubernetesVersionChildren(node *TreeNode, appState *AppState, ctx context.Context, version string) {
	children := make([]*TreeNode, 0)
	
	// Find all resources with this Kubernetes version
	resources := findResourcesByK8sVersion(ctx, version, appState)
	for _, resourceMap := range resources {
		resourceID, resourceType := extractResourceInfoFromMap(resourceMap)
		enrichMachineResource(resourceMap, resourceType, resourceID, appState.stateClient, ctx, 0)
		enrichClusterMachineResource(resourceMap, resourceType, resourceID, appState.stateClient, ctx, 0)
		childNode := createResourceNodeFromMap(resourceMap, resourceID, resourceType)
		children = append(children, childNode)
	}
	
	node.Children = children
	slog.Info("Loaded KubernetesVersion children", "version", version, "children_count", len(children))
}

func createDetailComponents() *DetailComponents {
	title := widget.NewLabel(i18n.T("tree.select_resource"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	text := widget.NewRichText()
	text.Wrapping = fyne.TextWrapWord

	return &DetailComponents{
		Title:               title,
		Text:                text,
		MachineLinks:        container.NewVBox(),
		VersionLinks:        container.NewVBox(),
		MachineSetLinks:     container.NewVBox(),
		MachineRelatedLinks: container.NewVBox(),
		ResourceActions:     container.NewVBox(),
	}
}

func setupAppState(appState *AppState, resourceTree *widget.Tree, resourceIDInput *widget.Entry, statusLabel *widget.Label, detailComponents *DetailComponents) {
	appState.detailText = detailComponents.Text
	appState.detailTitle = detailComponents.Title
	appState.resourceTree = resourceTree
	if resourceIDInput != nil {
		appState.resourceIDInput = resourceIDInput
	}
	appState.statusLabel = statusLabel
	appState.machineLinksContainer = detailComponents.MachineLinks
	appState.versionLinksContainer = detailComponents.VersionLinks
	appState.machineSetLinksContainer = detailComponents.MachineSetLinks
	appState.machineRelatedLinksContainer = detailComponents.MachineRelatedLinks
	appState.resourceActionsContainer = detailComponents.ResourceActions
}

func createBurgerMenu(myWindow fyne.Window) *widget.Button {
	burgerMenu := widget.NewButton("☰ Menu", nil)
	
	menuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("Settings", func() {
			showSettingsPage(myWindow)
		}),
	}

	menu := fyne.NewMenu("Menu", menuItems...)
	
	burgerMenu.OnTapped = func() {
		popup := widget.NewPopUpMenu(menu, fyne.CurrentApp().Driver().CanvasForObject(burgerMenu))
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(burgerMenu)
		popup.ShowAtPosition(pos.Add(fyne.NewPos(0, burgerMenu.Size().Height)))
	}

	return burgerMenu
}

func createLanguageSelector(appState *AppState, refreshUI func()) *widget.Select {
	languages := i18n.GetSupportedLanguages()
	options := make([]string, 0, len(languages))
	for _, lang := range languages {
		options = append(options, i18n.GetLanguageName(lang))
	}

	langSelect := widget.NewSelect(options, func(selected string) {
		for _, lang := range languages {
			if i18n.GetLanguageName(lang) == selected {
				if err := i18n.SetLanguage(lang); err == nil {
					refreshUI()
				}
				break
			}
		}
	})

	currentLang := i18n.GetCurrentLanguage()
	langSelect.SetSelected(i18n.GetLanguageName(currentLang))

	return langSelect
}

func showSettingsPage(parentWindow fyne.Window) {
	// Always load settings without ignoring env for the settings page
	// The user can see what's currently configured
	settings, err := omniclient.LoadSettings(false)
	if err != nil {
		slog.Error("Failed to load settings", "error", err)
		settings = &omniclient.Settings{}
	}

	// Create settings window
	settingsWindow := fyne.CurrentApp().NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(500, 400))
	settingsWindow.CenterOnScreen()

	// Endpoint input
	endpointLabel := widget.NewLabel("Omni Endpoint:")
	endpointEntry := widget.NewEntry()
	if settings.Endpoint != "" {
		endpointEntry.SetText(settings.Endpoint)
	} else {
		endpointEntry.SetText(os.Getenv("OMNI_ENDPOINT"))
	}
	endpointEntry.SetPlaceHolder("https://omni.example.com")

	// Auth method selection
	authLabel := widget.NewLabel("Authentication Method:")
	authSelect := widget.NewSelect([]string{"Service Account", "OIDC"}, func(selected string) {
		// Selection handler
	})
	
	// Set current selection (check without ignoring env for display purposes)
	currentAuth := omniclient.GetCurrentAuthMethod(false)
	if currentAuth == omniclient.AuthMethodOIDC {
		authSelect.SetSelected("OIDC")
	} else {
		authSelect.SetSelected("Service Account")
	}

	// Service Account fields
	serviceAccountLabel := widget.NewLabel("Service Account Key:")
	serviceAccountEntry := widget.NewEntry()
	serviceAccountEntry.SetPlaceHolder("Base64 encoded service account key")
	serviceAccountEntry.Password = true
	// Load from settings file first, then fall back to environment
	if settings.ServiceAccount != "" {
		serviceAccountEntry.SetText(settings.ServiceAccount)
	} else if os.Getenv("OMNI_SERVICE_ACCOUNT") != "" {
		serviceAccountEntry.SetText(os.Getenv("OMNI_SERVICE_ACCOUNT"))
	} else if os.Getenv("OMNI_SERVICE_ACCOUNT_KEY") != "" {
		serviceAccountEntry.SetText(os.Getenv("OMNI_SERVICE_ACCOUNT_KEY"))
	}

	// OIDC fields
	oidcIssuerLabel := widget.NewLabel("OIDC Issuer URL:")
	oidcIssuerEntry := widget.NewEntry()
	oidcIssuerEntry.SetPlaceHolder("https://oidc-provider.com (optional, auto-derived from endpoint)")
	// Load from settings file first, then fall back to environment
	if settings.OIDCIssuerURL != "" {
		oidcIssuerEntry.SetText(settings.OIDCIssuerURL)
	} else if os.Getenv("OMNI_OIDC_ISSUER_URL") != "" {
		oidcIssuerEntry.SetText(os.Getenv("OMNI_OIDC_ISSUER_URL"))
	}

	oidcClientIDLabel := widget.NewLabel("OIDC Client ID:")
	oidcClientIDEntry := widget.NewEntry()
	oidcClientIDEntry.SetPlaceHolder("your-client-id")
	// Load from settings file first, then fall back to environment
	if settings.OIDCClientID != "" {
		oidcClientIDEntry.SetText(settings.OIDCClientID)
	} else if os.Getenv("OMNI_OIDC_CLIENT_ID") != "" {
		oidcClientIDEntry.SetText(os.Getenv("OMNI_OIDC_CLIENT_ID"))
	}

	oidcClientSecretLabel := widget.NewLabel("OIDC Client Secret:")
	oidcClientSecretEntry := widget.NewEntry()
	oidcClientSecretEntry.SetPlaceHolder("your-client-secret")
	oidcClientSecretEntry.Password = true
	// Load from settings file first, then fall back to environment
	if settings.OIDCClientSecret != "" {
		oidcClientSecretEntry.SetText(settings.OIDCClientSecret)
	} else if os.Getenv("OMNI_OIDC_CLIENT_SECRET") != "" {
		oidcClientSecretEntry.SetText(os.Getenv("OMNI_OIDC_CLIENT_SECRET"))
	}

	// Container for service account fields
	serviceAccountContainer := container.NewVBox(
		serviceAccountLabel,
		serviceAccountEntry,
	)

	// Container for OIDC fields
	oidcContainer := container.NewVBox(
		oidcIssuerLabel,
		oidcIssuerEntry,
		oidcClientIDLabel,
		oidcClientIDEntry,
		oidcClientSecretLabel,
		oidcClientSecretEntry,
	)

	// Show/hide fields based on auth method selection
	updateAuthFields := func(selected string) {
		if selected == "OIDC" {
			serviceAccountContainer.Hide()
			oidcContainer.Show()
		} else {
			serviceAccountContainer.Show()
			oidcContainer.Hide()
		}
	}
	authSelect.OnChanged = updateAuthFields
	updateAuthFields(authSelect.Selected) // Initial state

	// Save button
	saveButton := widget.NewButton("Save", func() {
		// Save settings
		newSettings := &omniclient.Settings{
			Endpoint: endpointEntry.Text,
		}
		
		if authSelect.Selected == "OIDC" {
			newSettings.AuthMethod = "oidc"
			newSettings.OIDCIssuerURL = oidcIssuerEntry.Text
			newSettings.OIDCClientID = oidcClientIDEntry.Text
			newSettings.OIDCClientSecret = oidcClientSecretEntry.Text
			// Clear service account when using OIDC
			newSettings.ServiceAccount = ""
		} else {
			newSettings.AuthMethod = "service_account"
			newSettings.ServiceAccount = serviceAccountEntry.Text
			// Clear OIDC credentials when using service account
			newSettings.OIDCIssuerURL = ""
			newSettings.OIDCClientID = ""
			newSettings.OIDCClientSecret = ""
		}

		if err := omniclient.SaveSettings(newSettings); err != nil {
			slog.Error("Failed to save settings", "error", err)
			errorDialog := widget.NewModalPopUp(
				widget.NewLabel(fmt.Sprintf("Failed to save settings: %v", err)),
				settingsWindow.Canvas(),
			)
			errorDialog.Resize(fyne.NewSize(300, 100))
			errorDialog.Show()
			return
		}

		// Show restart dialog
		showRestartDialog(settingsWindow, parentWindow)
	})

	// Cancel button
	cancelButton := widget.NewButton("Cancel", func() {
		settingsWindow.Close()
	})

	// Layout
	content := container.NewVBox(
		widget.NewLabel("Authentication Settings"),
		widget.NewSeparator(),
		endpointLabel,
		endpointEntry,
		widget.NewSeparator(),
		authLabel,
		authSelect,
		widget.NewSeparator(),
		serviceAccountContainer,
		oidcContainer,
		widget.NewSeparator(),
		container.NewHBox(cancelButton, saveButton),
	)

	scrollContent := container.NewScroll(content)
	settingsWindow.SetContent(scrollContent)
	settingsWindow.Show()
}

// showRestartDialog displays a dialog asking if the user wants to restart the application
func showRestartDialog(settingsWindow fyne.Window, parentWindow fyne.Window) {
	message := widget.NewLabel("Settings saved!\n\nRestart the application now to apply changes?")
	message.Wrapping = fyne.TextWrapWord
	
	var dialog *widget.PopUp
	
	restartButton := widget.NewButton("Restart Now", func() {
		if dialog != nil {
			dialog.Hide()
		}
		restartApplication(parentWindow)
	})
	
	laterButton := widget.NewButton("Later", func() {
		if dialog != nil {
			dialog.Hide()
		}
	})
	
	buttonContainer := container.NewHBox(laterButton, restartButton)
	dialogContent := container.NewVBox(
		message,
		widget.NewSeparator(),
		buttonContainer,
	)
	
	dialog = widget.NewModalPopUp(
		dialogContent,
		settingsWindow.Canvas(),
	)
	dialog.Resize(fyne.NewSize(350, 120))
	dialog.Show()
}

// restartApplication restarts the application with the same command-line arguments
func restartApplication(parentWindow fyne.Window) {
	slog.Info("Restarting application...")
	
	// Get the executable path
	executable, err := os.Executable()
	if err != nil {
		slog.Error("Failed to get executable path", "error", err)
		// Fallback: try to get from os.Args[0]
		executable = os.Args[0]
		// If it's a relative path, make it absolute
		if !filepath.IsAbs(executable) {
			wd, err := os.Getwd()
			if err == nil {
				executable = filepath.Join(wd, executable)
			}
		}
	}
	
	// Get current command-line arguments (excluding program name)
	args := os.Args[1:]
	
	// Create command to restart
	cmd := exec.Command(executable, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// Start the new process
	if err := cmd.Start(); err != nil {
		slog.Error("Failed to restart application", "error", err)
		// Show error to user
		errorDialog := widget.NewModalPopUp(
			widget.NewLabel(fmt.Sprintf("Failed to restart: %v\n\nPlease restart manually.", err)),
			parentWindow.Canvas(),
		)
		errorDialog.Resize(fyne.NewSize(300, 100))
		errorDialog.Show()
		return
	}
	
	// Close the current application
	slog.Info("New process started, exiting current instance")
	parentWindow.Close()
	os.Exit(0)
}

func createMainLayout(burgerMenu *widget.Button, resourceTree *widget.Tree, detailComponents *DetailComponents, statusLabel *widget.Label, langSelect *widget.Select) *container.Split {
	topBar := container.NewBorder(nil, nil, nil, langSelect, burgerMenu)
	// Format directive is in translation file: "app.connected" = "Connected to: %s"
	infoLabel := widget.NewLabel(i18n.T("app.connected", os.Getenv("OMNI_ENDPOINT"))) //nolint
	infoLabel.Wrapping = fyne.TextWrapWord

	// Wrap tree in scroll container to ensure it's visible
	// Add a minimum content to ensure tree is visible even when empty
	treeContainer := container.NewBorder(nil, nil, nil, nil, resourceTree)
	treeScroll := container.NewScroll(treeContainer)
	treeScroll.SetMinSize(fyne.NewSize(300, 400))
	leftPane := container.NewBorder(topBar, infoLabel, nil, nil, treeScroll)
	
	detailScroll := container.NewScroll(container.NewVBox(
		detailComponents.Title,
		detailComponents.ResourceActions,
		detailComponents.Text,
		detailComponents.MachineLinks,
		detailComponents.VersionLinks,
		detailComponents.MachineSetLinks,
		detailComponents.MachineRelatedLinks,
	))
	detailScroll.SetMinSize(fyne.NewSize(400, 0))

	rightPane := container.NewBorder(nil, statusLabel, nil, nil, detailScroll)
	
	split := container.NewHSplit(leftPane, rightPane)
	split.SetOffset(0.3)

	return split
}

func createExecuteQueryFunc(appState *AppState, resourceIDInput *widget.Entry, statusLabel *widget.Label, stateClient state.State) func() {
	return func() {
		resourceQueries := getResourceQueries()
		query, ok := resourceQueries[appState.selectedResourceType]
		if !ok {
			statusLabel.SetText(i18n.T("resource.type.unknown"))
			return
		}

		statusLabel.SetText(i18n.T("resource.querying"))
		ctx := context.Background()

		if query.IsList {
			executeListQuery(ctx, query.Type, appState, stateClient, statusLabel)
		} else {
			resourceID := resourceIDInput.Text
			if resourceID == "" {
				statusLabel.SetText(i18n.T("resource.id.required"))
				return
			}
			executeGetQuery(ctx, query.Type, resourceID, appState, stateClient, statusLabel)
		}
	}
}

func executeListQuery(ctx context.Context, resourceType resource.Type, appState *AppState, stateClient state.State, statusLabel *widget.Label) {
	md := resource.NewMetadata(omniresources.DefaultNamespace, resourceType, "", resource.VersionUndefined)
	list, err := stateClient.List(ctx, md)
	if err != nil {
		handleQueryError(err, appState, statusLabel)
		return
	}

	resources := make([]map[string]interface{}, 0, len(list.Items))
	for _, item := range list.Items {
		resources = append(resources, resconverter.ToMap(item))
	}

	// Clear existing nodes and rebuild
	nodeMap = make(map[widget.TreeNodeID]*TreeNode)
	
	// Create root node with children
	root := &TreeNode{
		ID:       "",
		Type:     "root",
		Label:    "",
		Resource: nil,
		Children: []*TreeNode{},
	}
	
	if len(resources) > 0 {
		resourceNodes := buildResourceNodes(resources, stateClient, ctx)
		
		// For clusters, wrap them in a folder node
		if resourceType == omni.ClusterType {
			clustersFolder := &TreeNode{
				ID:       "clusters-folder",
				Type:     "clusters-folder",
				Label:    "Clusters",
				Resource: map[string]interface{}{
					"type": "clusters-folder",
				},
				Children: resourceNodes,
			}
			root.Children = []*TreeNode{clustersFolder}
			addNodeToMap(clustersFolder)
		} else if resourceType == omni.MachineSetType {
			// Sort MachineSets by label (name)
			sort.Slice(resourceNodes, func(i, j int) bool {
				return resourceNodes[i].Label < resourceNodes[j].Label
			})
			
			// For MachineSets, wrap them in a folder node
			machineSetsFolder := &TreeNode{
				ID:       "machinesets-folder",
				Type:     "machinesets-folder",
				Label:    "MachineSets",
				Resource: map[string]interface{}{
					"type": "machinesets-folder",
				},
				Children: resourceNodes,
			}
			root.Children = []*TreeNode{machineSetsFolder}
			addNodeToMap(machineSetsFolder)
		} else {
			root.Children = resourceNodes
		}
	} else {
		root.Children = []*TreeNode{
			{
				ID:       "no-resources",
				Type:     "placeholder",
				Label:    "No resources found",
				Children: []*TreeNode{},
			},
		}
	}
	
	// Update appState and global root
	appState.treeRoot = root
	setRootNode(root)
	addNodeToMap(root)
	
	// Refresh tree display
	if appState.resourceTree != nil {
		appState.resourceTree.Refresh()
		appState.resourceTree.OpenBranch("")
		appState.resourceTree.Refresh()
	}
	
	statusLabel.SetText(i18n.T("resource.success", fmt.Sprintf("%d resources", len(resources))))
}

func executeGetQuery(ctx context.Context, resourceType resource.Type, resourceID string, appState *AppState, stateClient state.State, statusLabel *widget.Label) {
	md := resource.NewMetadata(omniresources.DefaultNamespace, resourceType, resourceID, resource.VersionUndefined)
	item, err := stateClient.Get(ctx, md)
	if err != nil {
		handleQueryError(err, appState, statusLabel)
		return
	}

	resourceMap := resconverter.ToMap(item)
	nodeID := fmt.Sprintf("%s-%s", resourceType, resourceID)
	
	// Clear and rebuild
	nodeMap = make(map[widget.TreeNodeID]*TreeNode)
	
	root := &TreeNode{
		ID:       "",
		Type:     "root",
		Label:    "",
		Resource: nil,
		Children: []*TreeNode{
			{
				ID:       nodeID,
				Type:     string(resourceType),
				Label:    formatResourceLabel(resourceMap),
				Resource: resourceMap,
				Children: []*TreeNode{},
			},
		},
	}
	
	appState.treeRoot = root
	setRootNode(root)
	addNodeToMap(root)
	
	if appState.resourceTree != nil {
		appState.resourceTree.Refresh()
		appState.resourceTree.OpenBranch("")
		appState.resourceTree.Refresh()
	}
	statusLabel.SetText(i18n.T("resource.success", resourceID))
}

func handleQueryError(err error, appState *AppState, statusLabel *widget.Label) {
	statusLabel.SetText(i18n.T("resource.error", err)) //nolint
	nodeMap = make(map[widget.TreeNodeID]*TreeNode)
	root := &TreeNode{
		ID:       "",
		Type:     "root",
		Label:    "",
		Resource: nil,
		Children: []*TreeNode{},
	}
	appState.treeRoot = root
	setRootNode(root)
	if appState.resourceTree != nil {
		appState.resourceTree.Refresh()
	}
}

func setupResourceIDAutoQuery(resourceIDInput *widget.Entry, appState *AppState, executeQuery func()) {
	resourceIDInput.OnSubmitted = func(_ string) {
		executeQuery()
	}
}

func buildResourceNodes(resources []map[string]interface{}, stateClient state.State, ctx context.Context) []*TreeNode {
	nodes := make([]*TreeNode, 0, len(resources))
	for _, resourceMap := range resources {
		resourceID, resourceType := extractResourceInfoFromMap(resourceMap)
		enrichMachineResource(resourceMap, resourceType, resourceID, stateClient, ctx, 0)
		enrichClusterMachineResource(resourceMap, resourceType, resourceID, stateClient, ctx, 0)
		node := createResourceNodeFromMap(resourceMap, resourceID, resourceType)
		nodes = append(nodes, node)
		// Add to map immediately
		addNodeToMap(node)
	}
	return nodes
}

func extractResourceInfoFromMap(resourceMap map[string]interface{}) (string, string) {
	resourceID := ""
	if id, ok := resourceMap["id"].(string); ok {
		resourceID = id
	}
	resourceType := ""
	if t, ok := resourceMap["type"].(string); ok {
		resourceType = t
	}
	return resourceID, resourceType
}

func enrichMachineResource(resourceMap map[string]interface{}, resourceType, resourceID string, stateClient state.State, ctx context.Context, depth int) {
	if resourceType != string(omni.MachineType) || resourceID == "" || depth > 1 {
		return
	}

	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, resourceID, resource.VersionUndefined)
	machineStatus, err := stateClient.Get(ctx, md)
	if err != nil {
		return
	}

	ms, ok := machineStatus.(*omni.MachineStatus)
	if !ok {
		return
	}

	if ms.TypedSpec().Value.Network != nil && ms.TypedSpec().Value.Network.Hostname != "" {
		resourceMap["hostname"] = ms.TypedSpec().Value.Network.Hostname
	}
}

func enrichClusterMachineResource(resourceMap map[string]interface{}, resourceType, resourceID string, stateClient state.State, ctx context.Context, depth int) {
	if resourceType != string(omni.ClusterMachineType) || resourceID == "" || depth > 1 {
		return
	}

	machineID, ok := resourceMap["machine_id"].(string)
	if !ok || machineID == "" {
		return
	}

	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, machineID, resource.VersionUndefined)
	machineStatus, err := stateClient.Get(ctx, md)
	if err != nil {
		return
	}

	ms, ok := machineStatus.(*omni.MachineStatus)
	if !ok {
		return
	}

	if ms.TypedSpec().Value.Network != nil && ms.TypedSpec().Value.Network.Hostname != "" {
		resourceMap["hostname"] = ms.TypedSpec().Value.Network.Hostname
	}
}

func createResourceNodeFromMap(resourceMap map[string]interface{}, resourceID, resourceType string) *TreeNode {
	label := formatResourceLabel(resourceMap)
	return &TreeNode{
		ID:       fmt.Sprintf("%s-%s", resourceType, resourceID),
		Type:     resourceType,
		Label:    label,
		Resource: resourceMap,
		Children: []*TreeNode{},
	}
}

func formatResourceLabel(resourceMap map[string]interface{}) string {
	resourceID := ""
	if id, ok := resourceMap["id"].(string); ok {
		resourceID = id
	}
	resourceType := ""
	if t, ok := resourceMap["type"].(string); ok {
		resourceType = t
	}

	label := resourceID
	switch resourceType {
	case string(omni.ClusterType):
		if kv, ok := resourceMap["kubernetes_version"].(string); ok {
			label = fmt.Sprintf("%s (K8s: %s)", resourceID, kv)
		}
	case string(omni.MachineType):
		if hostname, ok := resourceMap["hostname"].(string); ok && hostname != "" {
			label = hostname
		} else if addr, ok := resourceMap["management_address"].(string); ok && addr != "" {
			label = fmt.Sprintf("%s (%s)", resourceID, addr)
		}
	case string(omni.ClusterMachineType):
		if hostname, ok := resourceMap["hostname"].(string); ok && hostname != "" {
			label = hostname
		} else if machineID, ok := resourceMap["machine_id"].(string); ok && machineID != "" {
			label = fmt.Sprintf("%s (Machine: %s)", resourceID, machineID)
		}
	case string(omni.MachineSetType):
		if mc, ok := resourceMap["machine_class"].(string); ok {
			label = fmt.Sprintf("%s (Class: %s)", resourceID, mc)
		}
	}
	return label
}

func updateDetailPane(resourceData map[string]interface{}, appState *AppState) {
	appState.currentResource = resourceData

	resourceID, resourceType := extractResourceInfo(resourceData)
	updateDetailTitle(appState, resourceID)
	displayResource := loadMachineStatusIfNeeded(resourceData, resourceType, resourceID, appState)
	updateDetailJSON(displayResource, appState)
	updateDetailLinks(resourceData, resourceType, resourceID, appState)
}

func extractResourceInfo(resourceData map[string]interface{}) (string, string) {
	resourceID := ""
	if id, ok := resourceData["id"].(string); ok {
		resourceID = id
	}
	resourceType := ""
	if t, ok := resourceData["type"].(string); ok {
		resourceType = t
	}
	return resourceID, resourceType
}

func updateDetailTitle(appState *AppState, resourceID string) {
	if resourceID != "" {
		appState.detailTitle.SetText(fmt.Sprintf("Resource Details: %s", resourceID))
	} else {
		appState.detailTitle.SetText("Resource Details")
	}
}

func loadMachineStatusIfNeeded(resourceData map[string]interface{}, resourceType, resourceID string, appState *AppState) map[string]interface{} {
	if resourceType != string(omni.MachineType) || resourceID == "" {
		return resourceData
	}

	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, resourceID, resource.VersionUndefined)
	machineStatus, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		return resourceData
	}

	statusMap := resconverter.ToMap(machineStatus)
	if statusSpec, ok := statusMap["spec"].(map[string]interface{}); ok {
		resourceData["machine_status"] = statusSpec
	}

	return resourceData
}

func updateDetailJSON(resourceData map[string]interface{}, appState *AppState) {
	jsonBytes, err := json.MarshalIndent(resourceData, "", "  ")
	if err != nil {
		appState.detailText.ParseMarkdown(fmt.Sprintf("Error formatting JSON: %v", err))
		return
	}
	appState.detailText.ParseMarkdown(fmt.Sprintf("```json\n%s\n```", string(jsonBytes)))
}

func updateDetailLinks(resourceData map[string]interface{}, resourceType, resourceID string, appState *AppState) {
	clearAllLinkContainers(appState)
	updateResourceActions(resourceType, resourceID, appState)
	updateMachineIDLinks(resourceData, appState)
	updateKubernetesVersionLinks(resourceData, appState)
	updateMachineSetLinks(resourceData, appState)
}

func clearAllLinkContainers(appState *AppState) {
	appState.resourceActionsContainer.RemoveAll()
	appState.machineLinksContainer.RemoveAll()
	appState.versionLinksContainer.RemoveAll()
	appState.machineSetLinksContainer.RemoveAll()
	appState.machineRelatedLinksContainer.RemoveAll()
}

func updateResourceActions(resourceType, resourceID string, appState *AppState) {
	if resourceID == "" {
		return
	}

	actionsLabel := widget.NewLabel("Actions:")
	actionsLabel.TextStyle = fyne.TextStyle{Bold: true}
	appState.resourceActionsContainer.Add(actionsLabel)

	switch resourceType {
	case string(omni.ClusterType):
		addActionButton("Status", func() { loadClusterStatus(resourceID, appState) }, appState)
		addActionButton("Metrics", func() { loadClusterMetrics(resourceID, appState) }, appState)
		addActionButton("Bootstrap", func() { loadClusterBootstrap(resourceID, appState) }, appState)
		addActionButton("Kubeconfig", func() { loadClusterKubeconfig(resourceID, appState) }, appState)
		addActionButton("Kubernetes Upgrade", func() { loadClusterKubernetesUpgrade(resourceID, appState) }, appState)
		addActionButton("Talos Upgrade", func() { loadClusterTalosUpgrade(resourceID, appState) }, appState)
		addActionButton("Endpoints", func() { loadClusterEndpoints(resourceID, appState) }, appState)
		addActionButton("Kubernetes Status", func() { loadClusterKubernetesStatus(resourceID, appState) }, appState)
		addActionButton("Control Plane Status", func() { loadClusterControlPlaneStatus(resourceID, appState) }, appState)
		addActionButton("Diagnostics", func() { loadClusterDiagnostics(resourceID, appState) }, appState)
		addActionButton("Destroy Status", func() { loadClusterDestroyStatus(resourceID, appState) }, appState)
		addActionButton("Workload Proxy Status", func() { loadClusterWorkloadProxyStatus(resourceID, appState) }, appState)

	case string(omni.MachineType):
		addActionButton("Labels", func() { loadMachineLabels(resourceID, appState) }, appState)
		addActionButton("Extensions", func() { loadMachineExtensions(resourceID, appState) }, appState)
		addActionButton("Upgrade Status", func() { loadMachineUpgradeStatus(resourceID, appState) }, appState)
		addActionButton("Metrics", func() { loadMachineMetrics(resourceID, appState) }, appState)
		addActionButton("Config Diff", func() { loadMachineConfigDiff(resourceID, appState) }, appState)

	case string(omni.MachineSetType):
		addActionButton("Status", func() { loadMachineSetStatus(resourceID, appState) }, appState)
		addActionButton("Destroy Status", func() { loadMachineSetDestroyStatus(resourceID, appState) }, appState)

	case string(omni.ClusterMachineType):
		addActionButton("Status", func() { loadClusterMachineStatus(resourceID, appState) }, appState)
		addActionButton("Config Status", func() { loadClusterMachineConfigStatus(resourceID, appState) }, appState)
		addActionButton("Talos Version", func() { loadClusterMachineTalosVersion(resourceID, appState) }, appState)
		addActionButton("Config", func() { loadClusterMachineConfig(resourceID, appState) }, appState)
	}
}

func addActionButton(label string, action func(), appState *AppState) {
	btn := widget.NewButton(label, action)
	appState.resourceActionsContainer.Add(btn)
}

func updateMachineIDLinks(resourceData map[string]interface{}, appState *AppState) {
	machineIDs := findMachineIDs(resourceData)
	if len(machineIDs) == 0 {
		return
	}

	linksLabel := widget.NewLabel(i18n.T("detail.machine_id.label"))
	appState.machineLinksContainer.Add(linksLabel)
	for _, machineID := range machineIDs {
		// Format directive is in translation file: "detail.machine_id" = "Machine ID: %s"
		btn := widget.NewButton(i18n.T("detail.machine_id", machineID), func(id string) func() { //nolint
			return func() {
				loadMachineByID(id, appState)
			}
		}(machineID))
		appState.machineLinksContainer.Add(btn)
	}
}

func findMachineIDs(resourceData map[string]interface{}) []string {
	var machineIDs []string
	if machineID, ok := resourceData["machine_id"].(string); ok && machineID != "" {
		machineIDs = append(machineIDs, machineID)
	}
	return machineIDs
}

func updateKubernetesVersionLinks(resourceData map[string]interface{}, appState *AppState) {
	k8sVersions := findKubernetesVersions(resourceData)
	if len(k8sVersions) == 0 {
		return
	}

	linksLabel := widget.NewLabel(i18n.T("detail.kubernetes_version.label"))
	appState.versionLinksContainer.Add(linksLabel)
	for _, version := range k8sVersions {
		// Format directive is in translation file: "detail.kubernetes_version" = "Kubernetes Version: %s"
		btn := widget.NewButton(i18n.T("detail.kubernetes_version", version), func(v string) func() { //nolint
			return func() {
				loadResourcesByK8sVersion(v, appState)
			}
		}(version))
		appState.versionLinksContainer.Add(btn)
	}
}

func findKubernetesVersions(resourceData map[string]interface{}) []string {
	var versions []string
	if version, ok := resourceData["kubernetes_version"].(string); ok && version != "" {
		versions = append(versions, version)
	}
	return versions
}

func findMachineSetIDs(resourceData map[string]interface{}) []string {
	var machineSetIDs []string
	if labels, ok := resourceData["labels"].(map[string]interface{}); ok {
		if machineSet, ok := labels[labelKeyMachineSet].(string); ok && machineSet != "" {
			machineSetIDs = append(machineSetIDs, machineSet)
		}
	}
	return machineSetIDs
}

func updateMachineSetLinks(resourceData map[string]interface{}, appState *AppState) {
	machineSetIDs := findMachineSetIDs(resourceData)
	if len(machineSetIDs) == 0 {
		return
	}

	linksLabel := widget.NewLabel(i18n.T("detail.machine_set.label"))
	appState.machineSetLinksContainer.Add(linksLabel)
	for _, machineSetID := range machineSetIDs {
		// Format directive is in translation file: "detail.machine_set" = "Machine Set: %s"
		btn := widget.NewButton(i18n.T("detail.machine_set", machineSetID), func(id string) func() { //nolint
			return func() {
				loadMachinesByMachineSet(id, appState)
			}
		}(machineSetID))
		appState.machineSetLinksContainer.Add(btn)
	}
}

func findMachineSetIDs(resourceData map[string]interface{}) []string {
	var machineSetIDs []string
	if labels, ok := resourceData["labels"].(map[string]interface{}); ok {
		if machineSet, ok := labels[labelKeyMachineSet].(string); ok && machineSet != "" {
			machineSetIDs = append(machineSetIDs, machineSet)
		}
	}
	return machineSetIDs
}

// addLinkNodes adds link nodes as children to a resource node
func addLinkNodes(node *TreeNode, appState *AppState, ctx context.Context) {
	if node.Resource == nil {
		return
	}
	
	linkNodes := make([]*TreeNode, 0)
	
	// Add machine_id link if present
	machineIDs := findMachineIDs(node.Resource)
	for _, machineID := range machineIDs {
		// Check if this machine is already a child (to avoid duplicates)
		alreadyChild := false
		for _, child := range node.Children {
			if child.Type == string(omni.MachineType) {
				if id, ok := child.Resource["id"].(string); ok && id == machineID {
					alreadyChild = true
					break
				}
			}
		}
		if !alreadyChild {
			linkNode := createLinkNode("link-machine", fmt.Sprintf("Machine ID: %s", machineID), machineID, func() {
				loadMachineByID(machineID, appState)
			})
			linkNodes = append(linkNodes, linkNode)
		}
	}
	
	// Add kubernetes_version link if present
	k8sVersions := findKubernetesVersions(node.Resource)
	for _, version := range k8sVersions {
		// Check if this version is already a child
		alreadyChild := false
		for _, child := range node.Children {
			if child.Type == string(omni.KubernetesVersionType) {
				if id, ok := child.Resource["id"].(string); ok && id == version {
					alreadyChild = true
					break
				}
			}
		}
		if !alreadyChild {
			linkNode := createLinkNode("link-kubernetes-version", fmt.Sprintf("Kubernetes Version: %s", version), version, func() {
				loadResourcesByK8sVersion(version, appState)
			})
			linkNodes = append(linkNodes, linkNode)
		}
	}
	
	// Add machine_set link if present
	machineSetIDs := findMachineSetIDs(node.Resource)
	for _, machineSetID := range machineSetIDs {
		// Check if this machine set is already a child
		alreadyChild := false
		for _, child := range node.Children {
			if child.Type == string(omni.MachineSetType) {
				if id, ok := child.Resource["id"].(string); ok && id == machineSetID {
					alreadyChild = true
					break
				}
			}
		}
		if !alreadyChild {
			linkNode := createLinkNode("link-machine-set", fmt.Sprintf("Machine Set: %s", machineSetID), machineSetID, func() {
				loadMachinesByMachineSet(machineSetID, appState)
			})
			linkNodes = append(linkNodes, linkNode)
		}
	}
	
	// Append link nodes to existing children
	node.Children = append(node.Children, linkNodes...)
	
	// Add link nodes to the map
	for _, linkNode := range linkNodes {
		addNodeToMap(linkNode)
	}
}

// createLinkNode creates a link node that can be clicked to load a resource
func createLinkNode(linkType, label, targetID string, action func()) *TreeNode {
	linkID := fmt.Sprintf("%s-%s", linkType, targetID)
	
	// Store the action in a map so we can execute it when the link is clicked
	linkActionMap[linkID] = action
	
	return &TreeNode{
		ID:    linkID,
		Type:  linkType,
		Label: label,
		Resource: map[string]interface{}{
			"type":      linkType,
			"id":        targetID,
			"link_type": linkType,
		},
		Children: []*TreeNode{},
	}
}

// linkActionMap stores actions for link nodes
var linkActionMap = make(map[string]func())

func loadMachineByID(machineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, machineID, resource.VersionUndefined)

	machine, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		// Format directives are in translation file: "machine.error.extensions" = "Error loading extensions for machine %s: %v"
		appState.statusLabel.SetText(i18n.T("machine.error.extensions", machineID, err)) //nolint
		return
	}

	machineMap := resconverter.ToMap(machine)
	updateDetailPane(machineMap, appState)
	// Format directive is in translation file: "machine.loaded" = "Loaded machine: %s"
	appState.statusLabel.SetText(i18n.T("machine.loaded", machineID)) //nolint
}

func loadMachinesByMachineSet(machineSetID string, appState *AppState) {
	ctx := context.Background()
	resources := findClusterMachinesByMachineSet(ctx, machineSetID, appState)

	if len(resources) == 0 {
		// Format directive is in translation file: "machineset.not_found" = "No machines found in machine set %s"
		appState.statusLabel.SetText(i18n.T("machineset.not_found", machineSetID)) //nolint
		return
	}

	root := &TreeNode{
		ID:       "",
		Type:     "root",
		Label:    "",
		Resource: nil,
		Children: buildResourceNodes(resources, appState.stateClient, ctx),
	}
	appState.treeRoot = root
	setRootNode(root)
	addNodeToMap(root)
	appState.resourceTree.Refresh()
	// Format directives are in translation file: "machineset.found" = "Found %d machines in machine set %s"
	appState.statusLabel.SetText(i18n.T("machineset.found", len(resources), machineSetID)) //nolint
}

func findClusterMachinesByMachineSet(ctx context.Context, machineSetID string, appState *AppState) []map[string]interface{} {
	resources := make([]map[string]interface{}, 0)
	clusterMachineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, "", resource.VersionUndefined)
	list, err := appState.stateClient.List(ctx, clusterMachineMD)
	if err != nil {
		return resources
	}

	for _, item := range list.Items {
		cm, ok := item.(*omni.ClusterMachine)
		if !ok {
			continue
		}
		labels := cm.Metadata().Labels()
		if labels == nil {
			continue
		}
		machineSet, ok := labels.Get(labelKeyMachineSet)
		if !ok || machineSet != machineSetID {
			continue
		}
		resources = append(resources, resconverter.ToMap(cm))
	}

	return resources
}

func loadResourcesByK8sVersion(version string, appState *AppState) {
	ctx := context.Background()
	resources := findResourcesByK8sVersion(ctx, version, appState)

	if len(resources) == 0 {
		// Format directive is in translation file: "k8s.version.not_found" = "No resources found with Kubernetes version %s"
		appState.statusLabel.SetText(i18n.T("k8s.version.not_found", version)) //nolint
		return
	}

	root := &TreeNode{
		ID:       "",
		Type:     "root",
		Label:    "",
		Resource: nil,
		Children: buildResourceNodes(resources, appState.stateClient, ctx),
	}
	appState.treeRoot = root
	setRootNode(root)
	addNodeToMap(root)
	appState.resourceTree.Refresh()
	// Format directives are in translation file: "k8s.version.found" = "Found %d resources with Kubernetes version %s"
	appState.statusLabel.SetText(i18n.T("k8s.version.found", len(resources), version)) //nolint
}

func findResourcesByK8sVersion(ctx context.Context, version string, appState *AppState) []map[string]interface{} {
	resources := make([]map[string]interface{}, 0)
	resources = append(resources, findClustersByK8sVersion(ctx, version, appState)...)
	resources = append(resources, findClusterMachinesByK8sVersion(ctx, version, appState)...)
	return resources
}

func findClustersByK8sVersion(ctx context.Context, version string, appState *AppState) []map[string]interface{} {
	resources := make([]map[string]interface{}, 0)
	clusterMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterType, "", resource.VersionUndefined)
	list, err := appState.stateClient.List(ctx, clusterMD)
	if err != nil {
		return resources
	}

	for _, item := range list.Items {
		c, ok := item.(*omni.Cluster)
		if !ok {
			continue
		}
		if c.TypedSpec().Value.KubernetesVersion == version {
			resources = append(resources, resconverter.ToMap(c))
		}
	}

	return resources
}

func findClusterMachinesByK8sVersion(ctx context.Context, version string, appState *AppState) []map[string]interface{} {
	resources := make([]map[string]interface{}, 0)
	clusterMachineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, "", resource.VersionUndefined)
	list, err := appState.stateClient.List(ctx, clusterMachineMD)
	if err != nil {
		return resources
	}

	for _, item := range list.Items {
		cm, ok := item.(*omni.ClusterMachine)
		if !ok {
			continue
		}
		if cm.TypedSpec().Value.KubernetesVersion == version {
			resources = append(resources, resconverter.ToMap(cm))
		}
	}

	return resources
}

// Cluster action loaders
func loadClusterStatus(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterStatusType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster status: %s", clusterID))
}

func loadClusterMetrics(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMetricsType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster metrics: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster metrics: %s", clusterID))
}

func loadClusterBootstrap(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterBootstrapStatusType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster bootstrap: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster bootstrap: %s", clusterID))
}

func loadClusterKubeconfig(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.KubeconfigType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading kubeconfig: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded kubeconfig: %s", clusterID))
}

func loadClusterKubernetesUpgrade(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.KubernetesUpgradeStatusType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading Kubernetes upgrade: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded Kubernetes upgrade: %s", clusterID))
}

func loadClusterTalosUpgrade(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.TalosUpgradeStatusType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading Talos upgrade: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded Talos upgrade: %s", clusterID))
}

func loadClusterEndpoints(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterEndpointType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster endpoints: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster endpoints: %s", clusterID))
}

func loadClusterKubernetesStatus(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.KubernetesStatusType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading Kubernetes status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded Kubernetes status: %s", clusterID))
}

func loadClusterControlPlaneStatus(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ControlPlaneStatusType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading control plane status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded control plane status: %s", clusterID))
}

func loadClusterDiagnostics(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterDiagnosticsType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster diagnostics: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster diagnostics: %s", clusterID))
}

func loadClusterDestroyStatus(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterDestroyStatusType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster destroy status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster destroy status: %s", clusterID))
}

func loadClusterWorkloadProxyStatus(clusterID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterWorkloadProxyStatusType, clusterID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading workload proxy status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded workload proxy status: %s", clusterID))
}

// Machine action loaders
func loadMachineLabels(machineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineLabelsType, machineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading machine labels: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded machine labels: %s", machineID))
}

func loadMachineExtensions(machineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineExtensionsType, machineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading machine extensions: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded machine extensions: %s", machineID))
}

func loadMachineUpgradeStatus(machineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineUpgradeStatusType, machineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading machine upgrade status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded machine upgrade status: %s", machineID))
}

func loadMachineMetrics(machineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusMetricsType, machineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading machine metrics: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded machine metrics: %s", machineID))
}

func loadMachineConfigDiff(machineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineConfigDiffType, machineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading machine config diff: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded machine config diff: %s", machineID))
}

// MachineSet action loaders
func loadMachineSetStatus(machineSetID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineSetStatusType, machineSetID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading machine set status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded machine set status: %s", machineSetID))
}

func loadMachineSetDestroyStatus(machineSetID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineSetDestroyStatusType, machineSetID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading machine set destroy status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded machine set destroy status: %s", machineSetID))
}

// ClusterMachine action loaders
func loadClusterMachineStatus(clusterMachineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineStatusType, clusterMachineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster machine status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster machine status: %s", clusterMachineID))
}

func loadClusterMachineConfigStatus(clusterMachineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineConfigStatusType, clusterMachineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster machine config status: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster machine config status: %s", clusterMachineID))
}

func loadClusterMachineTalosVersion(clusterMachineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineTalosVersionType, clusterMachineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster machine Talos version: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster machine Talos version: %s", clusterMachineID))
}

func loadClusterMachineConfig(clusterMachineID string, appState *AppState) {
	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineConfigType, clusterMachineID, resource.VersionUndefined)
	res, err := appState.stateClient.Get(ctx, md)
	if err != nil {
		appState.statusLabel.SetText(fmt.Sprintf("Error loading cluster machine config: %v", err))
		return
	}
	resourceMap := resconverter.ToMap(res)
	updateDetailPane(resourceMap, appState)
	appState.statusLabel.SetText(fmt.Sprintf("Loaded cluster machine config: %s", clusterMachineID))
}

func main() {
	// Parse command-line flags
	ignoreEnv := flag.Bool("ignore-env", false, "Ignore environment variables and use only settings file")
	flag.Parse()
	
	// Initialize structured logger with JSON output to stderr
	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)
	
	slog.Info("Starting Omni GUI Application", "version", Version, "ignoreEnv", *ignoreEnv)
	
	slog.Info("Initializing i18n", "language", "en")
	if err := i18n.Init("en"); err != nil {
		slog.Error("Failed to initialize i18n", "error", err)
		slog.Info("Continuing without i18n")
		// Continue anyway - i18n.T will return the key if translations fail
	} else {
		slog.Info("i18n initialized successfully")
	}

	slog.Info("Creating Fyne application")
	myApp := app.New()
	slog.Info("Fyne app created")
	
	appTitle := i18n.T("app.title")
	slog.Info("App title retrieved", "title", appTitle)
	
	slog.Info("Creating window")
	myWindow := myApp.NewWindow(appTitle)
	slog.Info("Window created")
	
	windowSize := fyne.NewSize(1400, 900)
	myWindow.Resize(windowSize)
	slog.Info("Window resized", "width", windowSize.Width, "height", windowSize.Height)

	slog.Info("Initializing Omni client")
	omniClient, appState := initializeApp(myWindow, *ignoreEnv)
	if omniClient == nil {
		slog.Error("Omni client initialization failed")
		slog.Info("Showing error window and starting event loop")
		myWindow.CenterOnScreen()
		myWindow.Show()
		myWindow.ShowAndRun()
		slog.Info("Error window closed, exiting")
		return
	}
	defer omniClient.Close()
	slog.Info("Omni client initialized successfully")

	resourceTree := createResourceTree(appState)
	setupTreeSelection(resourceTree, appState)
	detailComponents := createDetailComponents()
	statusLabel := widget.NewLabel(i18n.T("app.ready"))

	burgerMenu := createBurgerMenu(myWindow)

	setupAppState(appState, resourceTree, nil, statusLabel, detailComponents)

	refreshUI := func() {
		statusLabel.SetText(i18n.T("app.ready"))
		// Root label is not displayed (empty string ID), so no need to update it
		if appState.detailTitle != nil {
			appState.detailTitle.SetText(i18n.T("tree.select_resource"))
		}
		appState.resourceTree.Refresh()
	}

	langSelect := createLanguageSelector(appState, refreshUI)
	mainContent := createMainLayout(burgerMenu, resourceTree, detailComponents, statusLabel, langSelect)

	myWindow.SetContent(mainContent)
	slog.Info("Window content set")

	// Center the window on screen
	myWindow.CenterOnScreen()
	
	// Ensure tree is opened and visible
	if appState.resourceTree != nil {
		slog.Info("Opening root branch and refreshing tree")
		appState.resourceTree.OpenBranch("")
		appState.resourceTree.Refresh()
		slog.Info("Tree opened and refreshed")
	} else {
		slog.Warn("resourceTree is nil after initial query")

	}
	
	slog.Info("Showing window and starting event loop")
	if myWindow.Content() != nil {
		contentSize := myWindow.Content().Size()
		slog.Info("Window has content", "width", contentSize.Width, "height", contentSize.Height)
	} else {
		slog.Warn("Window has no content")
	}
	myWindow.Show()
	slog.Info("Window.Show() called")
	myWindow.RequestFocus()
	slog.Info("RequestFocus() called")
	slog.Info("Now calling ShowAndRun() - this will block until window closes")
	myWindow.ShowAndRun()
	slog.Info("Application exited")
}
