package main

import (
	"context"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"

	"github.com/jubblin/omni-api/internal/i18n"
)

// Tree node storage - map for O(1) lookup
var nodeMap = make(map[widget.TreeNodeID]*TreeNode)

var rootNode *TreeNode

// buildInitialTree creates the initial tree structure
func buildInitialTree() *TreeNode {
	// Build tree structure based on API hierarchy:
	// - Clusters (folder) -> Cluster -> MachineSets (folder) -> MachineSet -> ClusterMachines -> ClusterMachine -> Machine
	// - Machines (folder) -> Machine -> ClusterMachine (reverse lookup) + MachineStatus
	// - MachineSets (folder) -> MachineSet -> ClusterMachines -> ClusterMachine -> Machine
	
	return &TreeNode{
		ID:       "",
		Type:     "root",
		Label:    "",
		Resource: nil,
		Children: []*TreeNode{
			{
				ID:    "clusters-top",
				Type:  resourceTypeFolder,
				Label: i18n.T("resource.cluster"),
				Resource: map[string]interface{}{
					"resource_type": string(omni.ClusterType),
				},
				Children: []*TreeNode{},
			},
			{
				ID:    "machines-top",
				Type:  resourceTypeFolder,
				Label: i18n.T("resource.machine"),
				Resource: map[string]interface{}{
					"resource_type": string(omni.MachineType),
				},
				Children: []*TreeNode{},
			},
			{
				ID:    "machinesets-top",
				Type:  resourceTypeFolder,
				Label: i18n.T("resource.machine_set"),
				Resource: map[string]interface{}{
					"resource_type": string(omni.MachineSetType),
				},
				Children: []*TreeNode{},
			},
		},
	}
}

// getChildIDs returns the list of child IDs for a given node ID
func getChildIDs(id widget.TreeNodeID) []widget.TreeNodeID {
	node := getNodeByID(id)
	if node == nil {
		return []widget.TreeNodeID{}
	}
	childIDs := make([]widget.TreeNodeID, 0, len(node.Children))
	for _, child := range node.Children {
		// Skip nil children to prevent crashes
		if child == nil {
			slog.Warn("getChildIDs: found nil child in node", "node_id", id, "node_type", node.Type)
			continue
		}
		// Skip nodes with empty IDs to prevent crashes
		if child.ID == "" {
			slog.Warn("getChildIDs: found child with empty ID", "node_id", id, "node_type", node.Type, "child_type", child.Type, "child_label", child.Label)
			continue
		}
		// Skip nodes with empty labels to avoid blank leaves (unless flag is set)
		if showEmptyLabelNodes || child.Label != "" {
			childIDs = append(childIDs, child.ID)
		}
	}
	// Log for ClusterMachine and Machine nodes to debug MachineStatus visibility (DEBUG level to reduce noise)
	if (node.Type == string(omni.ClusterMachineType) || node.Type == string(omni.MachineType)) && len(node.Children) > 0 {
		slog.Debug("getChildIDs", 
			"node_id", id,
			"node_type", node.Type,
			"children_count", len(node.Children),
			"child_ids_count", len(childIDs),
			"child_ids", childIDs)
	}
	return childIDs
}

// isBranchNode determines if a node can be expanded
func isBranchNode(id widget.TreeNodeID) bool {
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
}

// createTreeNodeWidget creates the UI widget for a tree node
func createTreeNodeWidget(branch bool) fyne.CanvasObject {
	// Create a container with icon and label
	icon := widget.NewIcon(nil)
	// Set a minimum size for the icon so it's visible
	icon.Resize(fyne.NewSize(theme.IconInlineSize(), theme.IconInlineSize()))
	label := widget.NewLabel("")
	return container.NewHBox(icon, label)
}

// setIconForNode sets the icon for a node based on its type and branch status
func setIconForNode(icon *widget.Icon, nodeType string, branch bool) {
	app := fyne.CurrentApp()
	if app == nil {
		slog.Debug("Current app is nil")
		icon.SetResource(nil)
		return
	}
	
	themeInstance := app.Settings().Theme()
	if themeInstance == nil {
		slog.Debug("Theme instance is nil")
		icon.SetResource(nil)
		return
	}
	
	iconResource := getIconForResourceType(themeInstance, nodeType, branch)
	if iconResource != nil {
		icon.SetResource(iconResource)
		icon.Refresh()
	} else {
		slog.Debug("Icon resource is nil", "resource_type", nodeType, "is_branch", branch)
		icon.SetResource(nil)
	}
}

// updateTreeNodeWidget updates the UI widget with node data
func updateTreeNodeWidget(id widget.TreeNodeID, branch bool, obj fyne.CanvasObject) {
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
	if node == nil {
		icon.SetResource(nil)
		label.SetText("")
		return
	}
	
	// Skip rendering nodes with empty labels to avoid blank leaves (unless flag is set)
	if !showEmptyLabelNodes && node.Label == "" {
		icon.SetResource(nil)
		label.SetText("")
		return
	}
	
	label.SetText(node.Label)
	setIconForNode(icon, node.Type, branch)
}

// createResourceTree creates the resource tree widget
func createResourceTree(appState *AppState) *widget.Tree {
	// Ensure root node and its children are in the map
	// Don't clear the map - we need the nodes that were set up in initializeApp
	if appState.treeRoot != nil {
		setRootNode(appState.treeRoot)
		addNodeToMap(appState.treeRoot)
	}
	
	tree := widget.NewTree(
		getChildIDs,
		isBranchNode,
		createTreeNodeWidget,
		updateTreeNodeWidget,
	)
	return tree
}

// getNodeByID retrieves a node by its ID
func getNodeByID(id widget.TreeNodeID) *TreeNode {
	// Empty string is the root
	if id == "" {
		return getRootNode()
	}
	// Look up in map
	return nodeMap[id]
}

// getRootNode returns the root node
func getRootNode() *TreeNode {
	return rootNode
}

// setRootNode sets the root node
func setRootNode(node *TreeNode) {
	rootNode = node
	if node != nil {
		nodeMap[""] = node
	}
}

// addNodeToMap adds a node and its children to the node map
// It prevents infinite recursion by checking if a node is already in the map
func addNodeToMap(node *TreeNode) {
	if node == nil {
		return
	}
	
	// Check if node is already in the map to prevent infinite recursion
	// (e.g., Machine -> ClusterMachine -> Machine circular reference)
	if node.ID != "" {
		if _, exists := nodeMap[node.ID]; exists {
			// Node already in map, skip to prevent infinite recursion
			return
		}
		nodeMap[node.ID] = node
	}
	
	// Add all children recursively (only if not already in map)
	for _, child := range node.Children {
		addNodeToMap(child)
	}
}

// canHaveChildren determines if a resource type can have children
func canHaveChildren(resourceType string) bool {
	// Root can have children
	if resourceType == "root" {
		return true
	}
	
	// Resource type folders can have children
	if resourceType == resourceTypeFolder {
		return true
	}
	
	// Clusters, MachineSets, and Machines folders can have children
	if resourceType == clustersFolder || resourceType == machinesetsFolder || resourceType == machinesFolder {
		return true
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

// getIconForResourceType returns the appropriate icon for a resource type
func getIconForResourceType(themeInstance fyne.Theme, resourceType string, isBranch bool) fyne.Resource {
	if themeInstance == nil {
		slog.Warn("Theme instance is nil in getIconForResourceType")
		return nil
	}
	
	// Resource type folders and grouping folders always use folder icon
	if resourceType == resourceTypeFolder || resourceType == clustersFolder || resourceType == machinesetsFolder || resourceType == machinesFolder {
		return themeInstance.Icon(theme.IconNameFolder)
	}
	
	// Use theme icons for different resource types
	switch resourceType {
	case string(omni.ClusterType), string(omni.ClusterMachineType), string(omni.MachineType):
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameComputer)
	case string(omni.MachineSetType):
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameStorage)
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
		if isBranch {
			return themeInstance.Icon(theme.IconNameFolder)
		}
		return themeInstance.Icon(theme.IconNameFile)
	}
}

// handleNodeSelection handles the selection of a tree node
func handleNodeSelection(id widget.TreeNodeID, appState *AppState) {
	if id == "" {
		return
	}
	node := getNodeByID(id)
	if node == nil {
		return
	}
	
	// Show details for resource node
	if node.Resource != nil {
		updateDetailPane(node.Resource, appState)
	}
}

// handleBranchOpened handles the expansion of a tree branch
func handleBranchOpened(id widget.TreeNodeID, resourceTree *widget.Tree, appState *AppState) {
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
	} else if len(node.Children) > 0 {
		// Children are already loaded (pre-loaded), ensure they're in the map and refresh
		addNodeToMap(node)
		resourceTree.Refresh()
	}
}

// setupTreeSelection sets up tree selection and branch expansion handlers
func setupTreeSelection(resourceTree *widget.Tree, appState *AppState) {
	// Handle node selection - show details in detail pane
	resourceTree.OnSelected = func(id widget.TreeNodeID) {
		handleNodeSelection(id, appState)
	}
	
	// Lazy load children when a branch is expanded
	resourceTree.OnBranchOpened = func(id widget.TreeNodeID) {
		handleBranchOpened(id, resourceTree, appState)
	}
}

// filterEmptyLabelNodes filters out nodes with empty labels (unless flag is set)
func filterEmptyLabelNodes(nodes []*TreeNode) []*TreeNode {
	if showEmptyLabelNodes {
		// If flag is set, only filter out nil nodes
		filtered := make([]*TreeNode, 0, len(nodes))
		for _, node := range nodes {
			if node != nil {
				filtered = append(filtered, node)
			}
		}
		return filtered
	}
	// Default behavior: filter out nodes with empty labels
	filtered := make([]*TreeNode, 0, len(nodes))
	for _, node := range nodes {
		if node != nil && node.Label != "" {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// refreshTreeData refreshes the tree data by rebuilding the internal structure
func refreshTreeData(appState *AppState) {
	if appState == nil {
		slog.Warn("Cannot refresh: appState is nil")
		return
	}
	
	slog.Info("Refreshing tree data - rebuilding internal structure")
	
	// Clear existing node map
	nodeMap = make(map[widget.TreeNodeID]*TreeNode)
	
	// Rebuild initial tree structure
	root := buildInitialTree()
	appState.treeRoot = root
	setRootNode(root)
	addNodeToMap(root)
	
	// Refresh the tree widget to show the changes
	if appState.resourceTree != nil {
		appState.resourceTree.Refresh()
		// Open root branch to show top-level folders
		appState.resourceTree.OpenBranch("")
		appState.resourceTree.Refresh()
		slog.Info("Tree refreshed successfully")
	} else {
		slog.Warn("Cannot refresh tree widget: resourceTree is nil")
	}
	
	// Update status label
	if appState.statusLabel != nil {
		appState.statusLabel.SetText("Tree data refreshed")
	}
	
	// Clear detail pane
	if appState.detailTitle != nil {
		appState.detailTitle.SetText(i18n.T("tree.select_resource"))
	}
	if appState.detailText != nil {
		appState.detailText.SetText("")
	}
	appState.currentResource = nil
}
