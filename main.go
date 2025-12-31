package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
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


func loadNodeChildren(node *TreeNode, appState *AppState, ctx context.Context) {
	if node.Resource == nil {
		return
	}
	
	// Check if this is a resource type folder node
	if node.Type == resourceTypeFolder {
		loadResourceTypeFolderChildren(node, appState, ctx)
		return
	}
	
	resourceID, resourceTypeStr := extractResourceInfoFromMap(node.Resource)
	resourceType := resource.Type(resourceTypeStr)
	slog.Info("Loading children for node", "resource_type", resourceType, "resource_id", resourceID)
	
	// Load regular children first
	switch resourceType {
	case omni.ClusterType:
		loadClusterChildren(node, appState, ctx, resourceID)
	case omni.MachineSetType:
		loadMachineSetChildren(node, appState, ctx, resourceID)
	case omni.ClusterMachineType:
		loadClusterMachineChildren(node, appState, ctx, resourceID)
	case omni.MachineType:
		loadMachineChildren(node, appState, ctx, resourceID, false)
	case omni.KubernetesVersionType:
		loadKubernetesVersionChildren(node, appState, ctx, resourceID)
	}
}

// convertResourcesToNodes converts a list of resources to tree nodes
func convertResourcesToNodes(items []resource.Resource, resourceTypeStr string, stateClient state.State, ctx context.Context) []*TreeNode {
	children := make([]*TreeNode, 0, len(items))
	for _, item := range items {
		resourceMap := resconverter.ToMap(item)
		resourceID, _ := extractResourceInfoFromMap(resourceMap)
		enrichMachineResource(resourceMap, resourceTypeStr, resourceID, stateClient, ctx, 0)
		enrichClusterMachineResource(resourceMap, resourceTypeStr, resourceID, stateClient, ctx, 0)
		childNode := createResourceNodeFromMap(resourceMap, resourceID, resourceTypeStr)
		children = append(children, childNode)
	}
	return children
	}
	
// handleClusterTypeChildren handles children for Cluster resource type
func handleClusterTypeChildren(node *TreeNode, children []*TreeNode) {
		if len(children) > 0 {
			clustersFolderNode := &TreeNode{
				ID:       clustersFolder,
				Type:     clustersFolder,
				Label:    "Clusters",
				Resource: map[string]interface{}{"type": clustersFolder},
				Children: children,
			}
			node.Children = []*TreeNode{clustersFolderNode}
			addNodeToMap(clustersFolderNode)
		} else {
			node.Children = []*TreeNode{}
		}
}

// handleMachineSetTypeChildren handles children for MachineSet resource type
func handleMachineSetTypeChildren(node *TreeNode, children []*TreeNode) {
		if len(children) > 0 {
			sort.Slice(children, func(i, j int) bool {
				return children[i].Label < children[j].Label
			})
			machineSetsFolder := &TreeNode{
				ID:       machinesetsFolder,
				Type:     machinesetsFolder,
				Label:    "MachineSets",
				Resource: map[string]interface{}{"type": machinesetsFolder},
				Children: children,
			}
			node.Children = []*TreeNode{machineSetsFolder}
			addNodeToMap(machineSetsFolder)
		} else {
			node.Children = []*TreeNode{}
		}
}

// handleMachineTypeChildren handles children for Machine resource type
func handleMachineTypeChildren(node *TreeNode, children []*TreeNode) {
		if len(children) > 0 {
			sort.Slice(children, func(i, j int) bool {
				return children[i].Label < children[j].Label
			})
		}
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
}

// handleDefaultTypeChildren handles children for other resource types
func handleDefaultTypeChildren(node *TreeNode, children []*TreeNode) {
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
}

// loadResourceTypeFolderChildren loads all resources of a given type when a top-level resource type folder is expanded
func loadResourceTypeFolderChildren(node *TreeNode, appState *AppState, ctx context.Context) {
	if node.Resource == nil {
		return
	}
	
	resourceTypeStr, ok := node.Resource["resource_type"].(string)
	if !ok {
		return
	}
	
	resourceType := resource.Type(resourceTypeStr)
	
	// List all resources of this type
	md := resource.NewMetadata(omniresources.DefaultNamespace, resourceType, "", resource.VersionUndefined)
	list, err := debugStateList(ctx, appState.stateClient, md)
	if err != nil {
		slog.Error("Failed to list resources", "resource_type", resourceType, "error", err)
		return
	}
	
	// Convert resources to nodes
	children := convertResourcesToNodes(list.Items, resourceTypeStr, appState.stateClient, ctx)
	
	// Handle children based on resource type
	switch resourceType {
	case omni.ClusterType:
		handleClusterTypeChildren(node, children)
	case omni.MachineSetType:
		handleMachineSetTypeChildren(node, children)
	case omni.MachineType:
		handleMachineTypeChildren(node, children)
	default:
		handleDefaultTypeChildren(node, children)
	}
	
	// Add all children to the map
	for _, child := range node.Children {
		addNodeToMap(child)
	}
	
	slog.Info("Loaded resource type folder children", "resource_type", resourceType, "count", len(children))
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
	folderID := fmt.Sprintf("%s-%s", machinesetsFolder, parentID)
	machineSetsFolder := &TreeNode{
		ID:    folderID,
		Type:  machinesetsFolder,
		Label: "MachineSets",
		Resource: map[string]interface{}{
			"type": machinesetsFolder,
		},
		Children: machineSetNodes,
	}
	
	// Add folder and its children to the map
	addNodeToMap(machineSetsFolder)
	
	return machineSetsFolder
}

// loadMachineSetsForCluster loads all MachineSets for a given cluster
func loadMachineSetsForCluster(ctx context.Context, clusterID string, appState *AppState) ([]*TreeNode, map[string]bool) {
	machineSetNodes := make([]*TreeNode, 0)
	machineSetIDs := make(map[string]bool)
	
	machineSetMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineSetType, "", resource.VersionUndefined)
	machineSetList, err := debugStateList(ctx, appState.stateClient, machineSetMD)
	if err != nil {
		return machineSetNodes, machineSetIDs
	}
	
	for _, item := range machineSetList.Items {
		ms, ok := item.(*omni.MachineSet)
		if !ok {
			continue
		}
		labels := ms.Metadata().Labels()
		if labels == nil {
			continue
		}
		cluster, ok := labels.Get(labelKeyCluster)
		if !ok || cluster != clusterID {
			continue
		}
		msID := ms.Metadata().ID()
		machineSetIDs[msID] = true
		resourceMap := resconverter.ToMap(ms)
		childNode := createResourceNodeFromMap(resourceMap, msID, string(omni.MachineSetType))
		machineSetNodes = append(machineSetNodes, childNode)
	}
	
	return machineSetNodes, machineSetIDs
}

// loadOrphanedClusterMachines loads ClusterMachines that don't belong to a MachineSet
func loadOrphanedClusterMachines(ctx context.Context, clusterID string, appState *AppState) []*TreeNode {
	children := make([]*TreeNode, 0)
	
	clusterMachineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, "", resource.VersionUndefined)
	clusterMachineList, err := debugStateList(ctx, appState.stateClient, clusterMachineMD)
	if err != nil {
		return children
	}
	
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
		cluster, ok := labels.Get(labelKeyCluster)
		if !ok || cluster != clusterID {
			continue
		}
		
		// Only add ClusterMachine if it doesn't belong to a MachineSet
		if machineSet, ok := labels.Get(labelKeyMachineSet); ok && machineSet != "" {
			continue
		}
		
		// This ClusterMachine doesn't belong to a MachineSet, add it to cluster
		resourceMap := resconverter.ToMap(cm)
		enrichClusterMachineResource(resourceMap, string(omni.ClusterMachineType), cm.Metadata().ID(), appState.stateClient, ctx, 0)
		cmID := cm.Metadata().ID()
		childNode := createResourceNodeFromMap(resourceMap, cmID, string(omni.ClusterMachineType))
		children = append(children, childNode)
	}
	
	return children
}

// loadKubernetesVersionForCluster loads the KubernetesVersion node if the cluster has one
func loadKubernetesVersionForCluster(ctx context.Context, node *TreeNode, appState *AppState) *TreeNode {
	kv, ok := node.Resource["kubernetes_version"].(string)
	if !ok || kv == "" {
		return nil
	}
	
	kvMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.KubernetesVersionType, kv, resource.VersionUndefined)
	kvRes, err := debugStateGet(ctx, appState.stateClient, kvMD)
	if err != nil {
		return nil
	}
	
	resourceMap := resconverter.ToMap(kvRes)
	return createResourceNodeFromMap(resourceMap, kv, string(omni.KubernetesVersionType))
}

func loadClusterChildren(node *TreeNode, appState *AppState, ctx context.Context, clusterID string) {
	children := make([]*TreeNode, 0)
	
	// Load MachineSets for this cluster
	machineSetNodes, machineSetIDs := loadMachineSetsForCluster(ctx, clusterID, appState)
	if len(machineSetNodes) > 0 {
		if machineSetsFolder := groupMachineSetsIntoFolder(machineSetNodes, clusterID); machineSetsFolder != nil {
			children = append(children, machineSetsFolder)
		}
	}
	
	// Load orphaned ClusterMachines (those not in a MachineSet)
	orphanedClusterMachines := loadOrphanedClusterMachines(ctx, clusterID, appState)
	children = append(children, orphanedClusterMachines...)
	
	// Load KubernetesVersion if present
	if kvNode := loadKubernetesVersionForCluster(ctx, node, appState); kvNode != nil {
		children = append(children, kvNode)
	}
	
	// Filter out any nodes with empty labels
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
	
	// Ensure all children are added to the node map
	// (groupMachineSetsIntoFolder already adds its children, but we need to add orphaned cluster machines and kubernetes version)
	for _, child := range filteredChildren {
		addNodeToMap(child)
	}
	
	slog.Info("Loaded cluster children", 
		"cluster_id", clusterID,
		"total_children", len(filteredChildren),
		"machinesets", len(machineSetIDs),
		"orphaned_clustermachines", len(filteredChildren)-len(machineSetIDs))
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
	
	// Filter out any nodes with empty labels
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
	
	// Ensure all children are added to the node map
	for _, child := range filteredChildren {
		addNodeToMap(child)
	}
	
	slog.Info("Loaded MachineSet children", "machineset_id", machineSetID, "children_count", len(filteredChildren))
}

// loadMachineNodeForClusterMachine loads the Machine node for a ClusterMachine
// This function uses loadMachineNode with skipClusterMachineChild=true to prevent circular references
func loadMachineNodeForClusterMachine(node *TreeNode, appState *AppState, ctx context.Context) *TreeNode {
	machineID, ok := node.Resource["machine_id"].(string)
	if !ok || machineID == "" {
		return nil
	}
	
	// Use loadMachineNode with skipClusterMachineChild=true to prevent circular reference
	// (Machine -> ClusterMachine -> Machine)
	return loadMachineNode(ctx, machineID, appState, true)
}

// loadClusterNodeForClusterMachine loads the Cluster node for a ClusterMachine from labels
func loadClusterNodeForClusterMachine(node *TreeNode, appState *AppState, ctx context.Context) *TreeNode {
	labels, ok := node.Resource["labels"].(map[string]interface{})
	if !ok {
		return nil
	}
	
	clusterID, ok := labels[labelKeyCluster].(string)
	if !ok || clusterID == "" {
		return nil
	}
	
			clusterMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterType, clusterID, resource.VersionUndefined)
			cluster, err := debugStateGet(ctx, appState.stateClient, clusterMD)
	if err != nil {
		return nil
	}
	
				resourceMap := resconverter.ToMap(cluster)
	return createResourceNodeFromMap(resourceMap, clusterID, string(omni.ClusterType))
}

// loadMachineSetNodeForClusterMachine loads the MachineSet node for a ClusterMachine from labels
func loadMachineSetNodeForClusterMachine(node *TreeNode, appState *AppState, ctx context.Context) *TreeNode {
	labels, ok := node.Resource["labels"].(map[string]interface{})
	if !ok {
		return nil
	}
	
	machineSetID, ok := labels[labelKeyMachineSet].(string)
	if !ok || machineSetID == "" {
		return nil
	}
	
			machineSetMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineSetType, machineSetID, resource.VersionUndefined)
			machineSet, err := debugStateGet(ctx, appState.stateClient, machineSetMD)
	if err != nil {
		return nil
	}
	
				resourceMap := resconverter.ToMap(machineSet)
	return createResourceNodeFromMap(resourceMap, machineSetID, string(omni.MachineSetType))
}

func loadClusterMachineChildren(node *TreeNode, appState *AppState, ctx context.Context, clusterMachineID string) {
	// If children are already pre-loaded, use them instead of reloading
	slog.Info("loadClusterMachineChildren called", 
		"clustermachine_id", clusterMachineID,
		"existing_children_count", len(node.Children),
		"node_id", node.ID)
	
	if len(node.Children) > 0 {
		slog.Info("ClusterMachine children already pre-loaded", "clustermachine_id", clusterMachineID, "children_count", len(node.Children))
		// Log what children are pre-loaded
		for i, child := range node.Children {
			slog.Info("Pre-loaded child", 
				"index", i,
				"child_id", child.ID,
				"child_type", child.Type,
				"child_label", child.Label)
		}
		// Ensure all pre-loaded children are in the node map
		for _, child := range node.Children {
			addNodeToMap(child)
		}
		return
	}
	
	slog.Info("No pre-loaded children found, loading ClusterMachine children", "clustermachine_id", clusterMachineID)
	
	children := make([]*TreeNode, 0)
	
	// Get machine_id from ClusterMachine resource
	machineID, ok := node.Resource["machine_id"].(string)
	if !ok || machineID == "" {
		// Fallback: ClusterMachine ID is typically the same as Machine ID
		machineID = clusterMachineID
	}
	
	// Find the Machine for this ClusterMachine
	if machineNode := loadMachineNodeForClusterMachine(node, appState, ctx); machineNode != nil {
		children = append(children, machineNode)
	}
	
	// Add MachineStatus (same ID as Machine)
	machineStatusNode := loadMachineStatusNode(ctx, machineID, appState)
	if machineStatusNode != nil {
		if showEmptyLabelNodes || machineStatusNode.Label != "" {
			children = append(children, machineStatusNode)
			slog.Info("Added MachineStatus to ClusterMachine children", 
				"machine_id", machineID,
				"clustermachine_id", clusterMachineID,
				"machinestatus_label", machineStatusNode.Label,
				"machinestatus_id", machineStatusNode.ID)
		} else {
			slog.Warn("MachineStatus node has empty label, skipping", 
				"machine_id", machineID,
				"clustermachine_id", clusterMachineID)
		}
	} else {
		slog.Debug("MachineStatus not found for ClusterMachine", 
			"machine_id", machineID,
			"clustermachine_id", clusterMachineID)
	}
	
	// Find the Cluster for this ClusterMachine (from labels)
	if clusterNode := loadClusterNodeForClusterMachine(node, appState, ctx); clusterNode != nil {
		children = append(children, clusterNode)
	}
	
	// Find the MachineSet for this ClusterMachine
	if machineSetNode := loadMachineSetNodeForClusterMachine(node, appState, ctx); machineSetNode != nil {
		children = append(children, machineSetNode)
	}
	
	// Filter out any nodes with empty labels
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
	
	// Ensure all children are added to the node map
	for _, child := range filteredChildren {
		addNodeToMap(child)
	}
	
	// Log what children were loaded
	childList := make([]map[string]string, 0, len(filteredChildren))
	for _, child := range filteredChildren {
		childList = append(childList, map[string]string{
			"label": child.Label,
			"type":  child.Type,
			"id":    child.ID,
		})
	}
	
	slog.Info("Loaded ClusterMachine children", 
		"clustermachine_id", clusterMachineID, 
		"children_count", len(filteredChildren),
		"before_filter", len(children),
		"children", childList)
}


// loadClusterMachineForMachine loads ClusterMachine node with its children for a given machine
func loadClusterMachineForMachine(ctx context.Context, machineID string, appState *AppState) *TreeNode {
	clusterMachineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, machineID, resource.VersionUndefined)
	cm, err := debugStateGet(ctx, appState.stateClient, clusterMachineMD)
	if err != nil {
		slog.Debug("ClusterMachine not found for machine", "machine_id", machineID, "error", err)
		return nil
	}

	resourceMap := resconverter.ToMap(cm)
	// Extract the actual ClusterMachine ID from the resource map
	clusterMachineID, _ := extractResourceInfoFromMap(resourceMap)
	if clusterMachineID == "" {
		clusterMachineID = machineID // Fallback to machineID if extraction fails
	}
	
	enrichClusterMachineResource(resourceMap, string(omni.ClusterMachineType), clusterMachineID, appState.stateClient, ctx, 0)
	childNode := createResourceNodeFromMap(resourceMap, clusterMachineID, string(omni.ClusterMachineType))
	
	// Ensure label is never empty - prefix with "(ClusterMachine)" for clarity
	if childNode.Label == "" {
		// If label is still empty after creation, use a fallback
		if clusterMachineID != "" {
			childNode.Label = fmt.Sprintf(clusterMachineLabelPrefix, clusterMachineID)
	} else {
			childNode.Label = fmt.Sprintf(clusterMachineLabelPrefix, machineID)
		}
	} else {
		childNode.Label = fmt.Sprintf(clusterMachineLabelPrefix, childNode.Label)
	}
	
	// Load ClusterMachine children
	childNode.Children = loadClusterMachineChildrenFromMap(ctx, resourceMap, machineID, appState)
	
	// Ensure all pre-loaded children are added to the node map
	for _, child := range childNode.Children {
		addNodeToMap(child)
		slog.Debug("Added pre-loaded child to nodeMap", 
			"child_id", child.ID, 
			"child_type", child.Type, 
			"child_label", child.Label,
			"clustermachine_id", clusterMachineID)
	}
	
	slog.Debug("Pre-loaded ClusterMachine children", 
		"clustermachine_id", clusterMachineID,
		"machine_id", machineID,
		"children_count", len(childNode.Children))
	
	return childNode
}

// loadClusterMachineChildrenFromMap loads children for a ClusterMachine from its resource map
func loadClusterMachineChildrenFromMap(ctx context.Context, resourceMap map[string]interface{}, machineID string, appState *AppState) []*TreeNode {
	children := make([]*TreeNode, 0)
	
	// Add Machine (back reference)
	// Skip ClusterMachine child to prevent circular reference (Machine -> ClusterMachine -> Machine)
	if machineNode := loadMachineNode(ctx, machineID, appState, true); machineNode != nil {
		children = append(children, machineNode)
	}
	
	// Add MachineStatus (same ID as Machine)
	machineStatusNode := loadMachineStatusNode(ctx, machineID, appState)
	if machineStatusNode != nil {
		if showEmptyLabelNodes || machineStatusNode.Label != "" {
			children = append(children, machineStatusNode)
			slog.Debug("Added MachineStatus to ClusterMachine children", "machine_id", machineID, "label", machineStatusNode.Label)
		} else {
			slog.Debug("MachineStatus node has empty label, skipping", "machine_id", machineID)
		}
	} else {
		slog.Debug("MachineStatus not found for ClusterMachine", "machine_id", machineID)
	}
	
	// Add Cluster and MachineSet from labels
	if labels, ok := resourceMap["labels"].(map[string]interface{}); ok {
		if clusterNode := loadClusterNodeFromLabels(ctx, labels, appState); clusterNode != nil {
			children = append(children, clusterNode)
		}
		if machineSetNode := loadMachineSetNodeFromLabels(ctx, labels, appState); machineSetNode != nil {
			children = append(children, machineSetNode)
		}
	}
	
	filteredChildren := filterEmptyLabelNodes(children)
	slog.Debug("loadClusterMachineChildrenFromMap", 
		"machine_id", machineID,
		"before_filter", len(children),
		"after_filter", len(filteredChildren),
		"machinestatus", machineStatusNode != nil)
	return filteredChildren
}

// loadMachineNode loads a Machine node
// skipClusterMachineChild: if true, prevents loading ClusterMachine as a child to avoid circular references
func loadMachineNode(ctx context.Context, machineID string, appState *AppState, skipClusterMachineChild bool) *TreeNode {
	// Check if Machine node already exists in the map (e.g., from Machines folder)
	machineNodeID := fmt.Sprintf("%s-%s", string(omni.MachineType), machineID)
	existingNode := getNodeByID(machineNodeID)
	
	if existingNode != nil {
		// If we need to skip ClusterMachine child and the existing node has it, create a copy without it
		if skipClusterMachineChild {
			// Create a copy of the node without ClusterMachine children
			machineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, machineID, resource.VersionUndefined)
			machine, err := debugStateGet(ctx, appState.stateClient, machineMD)
			if err != nil {
				return nil
			}
			machineResourceMap := resconverter.ToMap(machine)
			enrichMachineResource(machineResourceMap, string(omni.MachineType), machineID, appState.stateClient, ctx, 0)
			machineNode := createResourceNodeFromMap(machineResourceMap, machineID, string(omni.MachineType))
			
			// Load children but filter out ClusterMachine
			loadMachineChildren(machineNode, appState, ctx, machineID, skipClusterMachineChild)
			return machineNode
		}
		slog.Debug("Reusing existing Machine node from map", 
			"machine_id", machineID,
			"existing_children_count", len(existingNode.Children))
		return existingNode
	}
	
	// Machine node doesn't exist yet, create a new one
	machineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, machineID, resource.VersionUndefined)
	machine, err := debugStateGet(ctx, appState.stateClient, machineMD)
	if err != nil {
		return nil
	}
	machineResourceMap := resconverter.ToMap(machine)
	enrichMachineResource(machineResourceMap, string(omni.MachineType), machineID, appState.stateClient, ctx, 0)
	machineNode := createResourceNodeFromMap(machineResourceMap, machineID, string(omni.MachineType))
	
	// Load children (with or without ClusterMachine based on skipClusterMachineChild)
	loadMachineChildren(machineNode, appState, ctx, machineID, skipClusterMachineChild)
	return machineNode
}

// loadClusterNodeFromLabels loads a Cluster node from labels
func loadClusterNodeFromLabels(ctx context.Context, labels map[string]interface{}, appState *AppState) *TreeNode {
	clusterID, ok := labels[labelKeyCluster].(string)
	if !ok || clusterID == "" {
		return nil
	}
	clusterMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterType, clusterID, resource.VersionUndefined)
	cluster, err := debugStateGet(ctx, appState.stateClient, clusterMD)
	if err != nil {
		return nil
	}
	clusterResourceMap := resconverter.ToMap(cluster)
	return createResourceNodeFromMap(clusterResourceMap, clusterID, string(omni.ClusterType))
}

// loadMachineSetNodeFromLabels loads a MachineSet node from labels
func loadMachineSetNodeFromLabels(ctx context.Context, labels map[string]interface{}, appState *AppState) *TreeNode {
	machineSetID, ok := labels[labelKeyMachineSet].(string)
	if !ok || machineSetID == "" {
		return nil
	}
	machineSetMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineSetType, machineSetID, resource.VersionUndefined)
	machineSet, err := debugStateGet(ctx, appState.stateClient, machineSetMD)
	if err != nil {
		return nil
	}
	machineSetResourceMap := resconverter.ToMap(machineSet)
	return createResourceNodeFromMap(machineSetResourceMap, machineSetID, string(omni.MachineSetType))
}

// extractHostnameFromResourceMap extracts hostname from a resource map

// loadMachineStatusNode loads MachineStatus node for a machine
func loadMachineStatusNode(ctx context.Context, machineID string, appState *AppState) *TreeNode {
	machineStatusMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, machineID, resource.VersionUndefined)
	ms, err := debugStateGet(ctx, appState.stateClient, machineStatusMD)
	if err != nil {
		slog.Debug("MachineStatus not found", "machine_id", machineID, "error", err)
		return nil
	}
	resourceMap := resconverter.ToMap(ms)
	statusID := fmt.Sprintf("%s-status", machineID)
	childNode := createResourceNodeFromMap(resourceMap, statusID, string(omni.MachineStatusType))
	
	// Ensure label is never empty
	hostname := extractHostnameFromResourceMap(resourceMap)
	if hostname != "" {
		childNode.Label = fmt.Sprintf("MachineStatus (%s)", hostname)
	} else if childNode.Label == "" {
		// Fallback if createResourceNodeFromMap didn't set a label
		childNode.Label = fmt.Sprintf("MachineStatus (%s)", machineID)
	}
	return childNode
}

func loadMachineChildren(node *TreeNode, appState *AppState, ctx context.Context, machineID string, skipClusterMachineChild bool) {
	children := make([]*TreeNode, 0)
	
	// Load MachineStatus first (same ID as Machine)
	machineStatusNode := loadMachineStatusNode(ctx, machineID, appState)
	if machineStatusNode != nil {
		if showEmptyLabelNodes || machineStatusNode.Label != "" {
			children = append(children, machineStatusNode)
		} else {
			slog.Warn("MachineStatus node has empty label", "machine_id", machineID)
		}
	} else {
		slog.Debug("No MachineStatus found for machine", "machine_id", machineID)
	}
	
	// Load ClusterMachine last (reverse lookup - ClusterMachine ID = Machine ID, prefix label with "(ClusterMachine)", hidden if not found)
	// Skip if skipClusterMachineChild is true to prevent circular references
	var clusterMachineNode *TreeNode
	if !skipClusterMachineChild {
		clusterMachineNode = loadClusterMachineForMachine(ctx, machineID, appState)
		if clusterMachineNode != nil {
			if showEmptyLabelNodes || clusterMachineNode.Label != "" {
				children = append(children, clusterMachineNode)
			} else {
				slog.Warn("ClusterMachine node has empty label", "machine_id", machineID)
			}
		} else {
			slog.Debug("No ClusterMachine found for machine", "machine_id", machineID)
		}
	} else {
		slog.Debug("Skipping ClusterMachine child to prevent circular reference", "machine_id", machineID)
	}
	
	// Filter out any nodes with empty labels from the final children list
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
	
	// Ensure all children are added to the node map
	for _, child := range filteredChildren {
		addNodeToMap(child)
	}
	
	// Build list of child labels and types for logging
	childList := make([]map[string]string, 0, len(filteredChildren))
	for _, child := range filteredChildren {
		childList = append(childList, map[string]string{
			"label": child.Label,
			"type":  child.Type,
			"id":    child.ID,
		})
	}
	
	slog.Info("Loaded Machine children", 
		"machine_id", machineID, 
		"children_count", len(filteredChildren),
		"before_filter", len(children),
		"clustermachine", clusterMachineNode != nil,
		"machinestatus", machineStatusNode != nil,
		"children", childList)
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
	
	// Filter out any nodes with empty labels
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
	
	// Ensure all children are added to the node map
	for _, child := range filteredChildren {
		addNodeToMap(child)
	}
	
	slog.Info("Loaded KubernetesVersion children", "version", version, "children_count", len(filteredChildren))
}





// createSaveButton creates the save button with settings save logic


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





// isResourceAlreadyChild checks if a resource with the given type and ID is already a child
func isResourceAlreadyChild(children []*TreeNode, resourceType, resourceID string) bool {
	for _, child := range children {
		if child.Type == resourceType {
			if id, ok := child.Resource["id"].(string); ok && id == resourceID {
				return true
			}
		}
	}
	return false
}


func findClusterMachinesByMachineSet(ctx context.Context, machineSetID string, appState *AppState) []map[string]interface{} {
	resources := make([]map[string]interface{}, 0)
	clusterMachineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, "", resource.VersionUndefined)
	list, err := debugStateList(ctx, appState.stateClient, clusterMachineMD)
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


func findResourcesByK8sVersion(ctx context.Context, version string, appState *AppState) []map[string]interface{} {
	resources := make([]map[string]interface{}, 0)
	resources = append(resources, findClustersByK8sVersion(ctx, version, appState)...)
	resources = append(resources, findClusterMachinesByK8sVersion(ctx, version, appState)...)
	return resources
}

func findClustersByK8sVersion(ctx context.Context, version string, appState *AppState) []map[string]interface{} {
	resources := make([]map[string]interface{}, 0)
	clusterMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterType, "", resource.VersionUndefined)
	list, err := debugStateList(ctx, appState.stateClient, clusterMD)
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
	list, err := debugStateList(ctx, appState.stateClient, clusterMachineMD)
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


// showEmptyLabelNodes controls whether nodes with empty labels are filtered out
var showEmptyLabelNodes = false

// logStartupBanner logs a startup banner to mark the beginning of a new execution
// All output is structured JSON for consistency
func logStartupBanner(version string, ignoreEnv bool, showEmptyLabels bool, grpcDebugLevel int) {
	startTime := time.Now().Format(time.RFC3339)
	
	// Log banner as structured JSON for programmatic access and consistency
	slog.Info("Application Startup Banner",
		"separator", "==================================================================================",
		"application", "Omni GUI Application",
		"version", version,
		"started_at", startTime,
		"ignore_env", ignoreEnv,
		"show_empty_labels", showEmptyLabels,
		"grpc_debug_level", grpcDebugLevel)
}

// setupLogging configures logging to write to both console (stderr) and a log file
// Returns the log file handle and any error encountered
func setupLogging() (*os.File, error) {
	// Get log file path in the current execution directory
	execDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}
	
	logPath := filepath.Join(execDir, "omni-gui.json")
	
	// Open log file in truncate mode (replace on startup, create if it doesn't exist)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	
	// Create a multi-writer that writes to both stderr and the log file
	multiWriter := io.MultiWriter(os.Stderr, logFile)
	
	// Initialize structured logger with JSON output to both console and file
	logger := slog.New(slog.NewJSONHandler(multiWriter, &slog.HandlerOptions{
		Level:     slog.LevelInfo,
		AddSource: true,
	}))
	slog.SetDefault(logger)
	
	slog.Info("Logging initialized", "log_file", logPath)
	
	return logFile, nil
}

func main() {
	// Load settings to get default values
	settings, err := omniclient.LoadSettings(false)
	ignoreEnvFromSettings := false
	grpcDebugLevelFromSettings := 0
	if err == nil {
		if settings.ShowEmptyLabels {
			showEmptyLabelNodes = true
		}
		ignoreEnvFromSettings = settings.IgnoreEnv
		grpcDebugLevelFromSettings = settings.GrpcDebugLevel
	}
	
	// Parse command-line flags (command-line flags take precedence over settings)
	ignoreEnv := flag.Bool("ignore-env", ignoreEnvFromSettings, "Ignore environment variables and use only settings file")
	showEmptyLabels := flag.Bool("show-empty-labels", showEmptyLabelNodes, "Show nodes with empty labels in the tree (disable filtering)")
	grpcDebugLevel := flag.Int("grpc-debug", grpcDebugLevelFromSettings, "gRPC debug level: 0=disabled, 1=query dumps, 2=query+response, 3=query+response+UI dumps")
	flag.Parse()
	
	// Set global flag for filtering (command-line flag takes precedence)
	showEmptyLabelNodes = *showEmptyLabels
	setGrpcDebugLevel(*grpcDebugLevel)
	
	// Setup logging to both console and file
	logFile, err := setupLogging()
	if err != nil {
		// If we can't set up file logging, at least log to stderr as JSON
		logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level:     slog.LevelInfo,
			AddSource: true,
		}))
		slog.SetDefault(logger)
		slog.Warn("Failed to set up file logging, using stderr only", "error", err)
	} else {
		// Close log file on exit
		defer logFile.Close()
	}
	
	// Log startup banner
	logStartupBanner(Version, *ignoreEnv, showEmptyLabelNodes, *grpcDebugLevel)
	
	slog.Info("Empty label filtering", "filtering_enabled", !showEmptyLabelNodes, "show_empty_labels", showEmptyLabelNodes)
	slog.Info("Ignore environment variables", "ignore_env", *ignoreEnv)
	slog.Info("gRPC debug level", "level", *grpcDebugLevel)
	
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

	burgerMenu := createBurgerMenu(myWindow, appState)
	refreshButton := createRefreshButton(appState)
	settingsButton := createSettingsButton(myWindow, appState)

	setupAppState(appState, resourceTree, nil, statusLabel, detailComponents)
	
	// Ensure root node children are in the map and tree is ready
	if appState.treeRoot != nil {
		slog.Info("Root node has children", "count", len(appState.treeRoot.Children))
		for i, child := range appState.treeRoot.Children {
			slog.Info("Root child", "index", i, "id", child.ID, "label", child.Label, "type", child.Type)
		}
		// Re-add to map to ensure they're accessible
		addNodeToMap(appState.treeRoot)
	}

	mainContent := createMainLayout(burgerMenu, refreshButton, settingsButton, resourceTree, detailComponents, statusLabel)

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
