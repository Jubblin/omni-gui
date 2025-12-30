package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"sort"

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
		loadMachineChildren(node, appState, ctx, resourceID)
	case omni.KubernetesVersionType:
		loadKubernetesVersionChildren(node, appState, ctx, resourceID)
	}
	
	// Add link nodes as children after loading regular children
	addLinkNodes(node, appState, ctx)
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
	list, err := appState.stateClient.List(ctx, md)
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
	machineSetList, err := appState.stateClient.List(ctx, machineSetMD)
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
	clusterMachineList, err := appState.stateClient.List(ctx, clusterMachineMD)
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
	kvRes, err := appState.stateClient.Get(ctx, kvMD)
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
	slog.Info("Loaded cluster children", 
		"cluster_id", clusterID,
		"total_children", len(filteredChildren),
		"machinesets", len(machineSetIDs),
		"orphaned_clustermachines", len(filteredChildren)-len(machineSetIDs))
}


func createLinkNodeWithAction(id, label, linkType string, action func()) *TreeNode {
	linkActionMap[id] = action
	return createLinkNode("link-"+linkType, label, id, action)
}

func createActionLinkNodes(clusterID string, appState *AppState) []*TreeNode {
	links := []*TreeNode{
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-status-%s", clusterID), i18n.T("link.cluster.status"), "cluster-status", func() { loadClusterStatus(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-metrics-%s", clusterID), i18n.T("link.cluster.metrics"), "cluster-metrics", func() { loadClusterMetrics(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-bootstrap-%s", clusterID), i18n.T("link.cluster.bootstrap"), "cluster-bootstrap", func() { loadClusterBootstrap(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-kubeconfig-%s", clusterID), i18n.T("link.cluster.kubeconfig"), "cluster-kubeconfig", func() { loadClusterKubeconfig(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-k8s-upgrade-%s", clusterID), i18n.T("link.cluster.k8s_upgrade"), "cluster-k8s-upgrade", func() { loadClusterKubernetesUpgrade(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-talos-upgrade-%s", clusterID), i18n.T("link.cluster.talos_upgrade"), "cluster-talos-upgrade", func() { loadClusterTalosUpgrade(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-endpoints-%s", clusterID), i18n.T("link.cluster.endpoints"), "cluster-endpoints", func() { loadClusterEndpoints(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-k8s-status-%s", clusterID), i18n.T("link.cluster.k8s_status"), "cluster-k8s-status", func() { loadClusterKubernetesStatus(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-controlplane-%s", clusterID), i18n.T("link.cluster.controlplane"), "cluster-controlplane", func() { loadClusterControlPlaneStatus(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-diagnostics-%s", clusterID), i18n.T("link.cluster.diagnostics"), "cluster-diagnostics", func() { loadClusterDiagnostics(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-destroy-%s", clusterID), i18n.T("link.cluster.destroy"), "cluster-destroy", func() { loadClusterDestroyStatus(clusterID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-cluster-workload-proxy-%s", clusterID), i18n.T("link.cluster.workload_proxy"), "cluster-workload-proxy", func() { loadClusterWorkloadProxyStatus(clusterID, appState) }),
	}
	return links
}

func createMachineActionLinkNodes(machineID string, appState *AppState) []*TreeNode {
	links := []*TreeNode{
		createLinkNodeWithAction(fmt.Sprintf("link-machine-labels-%s", machineID), i18n.T("link.machine.labels"), "machine-labels", func() { loadMachineLabels(machineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machine-extensions-%s", machineID), i18n.T("link.machine.extensions"), "machine-extensions", func() { loadMachineExtensions(machineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machine-upgrade-%s", machineID), i18n.T("link.machine.upgrade_status"), "machine-upgrade", func() { loadMachineUpgradeStatus(machineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machine-metrics-%s", machineID), i18n.T("link.machine.metrics"), "machine-metrics", func() { loadMachineMetrics(machineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machine-config-diff-%s", machineID), i18n.T("link.machine.config_diff"), "machine-config-diff", func() { loadMachineConfigDiff(machineID, appState) }),
	}
	return links
}

func createMachineSetActionLinkNodes(machineSetID string, appState *AppState) []*TreeNode {
	links := []*TreeNode{
		createLinkNodeWithAction(fmt.Sprintf("link-machineset-status-%s", machineSetID), i18n.T("link.machineset.status"), "machineset-status", func() { loadMachineSetStatus(machineSetID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-machineset-destroy-%s", machineSetID), i18n.T("link.machineset.destroy"), "machineset-destroy", func() { loadMachineSetDestroyStatus(machineSetID, appState) }),
	}
	return links
}

func createClusterMachineActionLinkNodes(clusterMachineID string, appState *AppState) []*TreeNode {
	links := []*TreeNode{
		createLinkNodeWithAction(fmt.Sprintf("link-clustermachine-status-%s", clusterMachineID), i18n.T("link.clustermachine.status"), "clustermachine-status", func() { loadClusterMachineStatus(clusterMachineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-clustermachine-config-status-%s", clusterMachineID), i18n.T("link.clustermachine.config_status"), "clustermachine-config-status", func() { loadClusterMachineConfigStatus(clusterMachineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-clustermachine-talos-version-%s", clusterMachineID), i18n.T("link.clustermachine.talos_version"), "clustermachine-talos-version", func() { loadClusterMachineTalosVersion(clusterMachineID, appState) }),
		createLinkNodeWithAction(fmt.Sprintf("link-clustermachine-config-%s", clusterMachineID), i18n.T("link.clustermachine.config"), "clustermachine-config", func() { loadClusterMachineConfig(clusterMachineID, appState) }),
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
	
	// Filter out any nodes with empty labels
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
	slog.Info("Loaded MachineSet children", "machineset_id", machineSetID, "children_count", len(filteredChildren))
}

// loadMachineNodeForClusterMachine loads the Machine node for a ClusterMachine
func loadMachineNodeForClusterMachine(node *TreeNode, appState *AppState, ctx context.Context) *TreeNode {
	machineID, ok := node.Resource["machine_id"].(string)
	if !ok || machineID == "" {
		return nil
	}
	
	machineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, machineID, resource.VersionUndefined)
	machine, err := appState.stateClient.Get(ctx, machineMD)
	if err != nil {
		return nil
	}
	
	resourceMap := resconverter.ToMap(machine)
	enrichMachineResource(resourceMap, string(omni.MachineType), machineID, appState.stateClient, ctx, 0)
	return createResourceNodeFromMap(resourceMap, machineID, string(omni.MachineType))
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
	cluster, err := appState.stateClient.Get(ctx, clusterMD)
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
	machineSet, err := appState.stateClient.Get(ctx, machineSetMD)
	if err != nil {
		return nil
	}
	
	resourceMap := resconverter.ToMap(machineSet)
	return createResourceNodeFromMap(resourceMap, machineSetID, string(omni.MachineSetType))
}

func loadClusterMachineChildren(node *TreeNode, appState *AppState, ctx context.Context, clusterMachineID string) {
	children := make([]*TreeNode, 0)
	
	// Find the Machine for this ClusterMachine
	if machineNode := loadMachineNodeForClusterMachine(node, appState, ctx); machineNode != nil {
		children = append(children, machineNode)
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
	slog.Info("Loaded ClusterMachine children", "clustermachine_id", clusterMachineID, "children_count", len(filteredChildren))
}


// loadClusterMachineForMachine loads ClusterMachine node with its children for a given machine
func loadClusterMachineForMachine(ctx context.Context, machineID string, appState *AppState) *TreeNode {
	clusterMachineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterMachineType, machineID, resource.VersionUndefined)
	cm, err := appState.stateClient.Get(ctx, clusterMachineMD)
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
	return childNode
}

// loadClusterMachineChildrenFromMap loads children for a ClusterMachine from its resource map
func loadClusterMachineChildrenFromMap(ctx context.Context, resourceMap map[string]interface{}, machineID string, appState *AppState) []*TreeNode {
	children := make([]*TreeNode, 0)
	
	// Add Machine (back reference)
	if machineNode := loadMachineNode(ctx, machineID, appState); machineNode != nil {
		children = append(children, machineNode)
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
	
	// Add link nodes for ClusterMachine
	clusterMachineLinks := createClusterMachineActionLinkNodes(machineID, appState)
	for _, link := range clusterMachineLinks {
		if link.Label != "" {
			children = append(children, link)
		}
	}
	
	return filterEmptyLabelNodes(children)
}

// loadMachineNode loads a Machine node
func loadMachineNode(ctx context.Context, machineID string, appState *AppState) *TreeNode {
	machineMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineType, machineID, resource.VersionUndefined)
	machine, err := appState.stateClient.Get(ctx, machineMD)
	if err != nil {
		return nil
	}
	machineResourceMap := resconverter.ToMap(machine)
	enrichMachineResource(machineResourceMap, string(omni.MachineType), machineID, appState.stateClient, ctx, 0)
	return createResourceNodeFromMap(machineResourceMap, machineID, string(omni.MachineType))
}

// loadClusterNodeFromLabels loads a Cluster node from labels
func loadClusterNodeFromLabels(ctx context.Context, labels map[string]interface{}, appState *AppState) *TreeNode {
	clusterID, ok := labels[labelKeyCluster].(string)
	if !ok || clusterID == "" {
		return nil
	}
	clusterMD := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterType, clusterID, resource.VersionUndefined)
	cluster, err := appState.stateClient.Get(ctx, clusterMD)
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
	machineSet, err := appState.stateClient.Get(ctx, machineSetMD)
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
	ms, err := appState.stateClient.Get(ctx, machineStatusMD)
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

func loadMachineChildren(node *TreeNode, appState *AppState, ctx context.Context, machineID string) {
	children := make([]*TreeNode, 0)
	
	// Load ClusterMachine (reverse lookup)
	clusterMachineNode := loadClusterMachineForMachine(ctx, machineID, appState)
	if clusterMachineNode != nil {
		if clusterMachineNode.Label == "" {
			slog.Warn("ClusterMachine node has empty label", "machine_id", machineID)
		} else {
			children = append(children, clusterMachineNode)
		}
	} else {
		slog.Debug("No ClusterMachine found for machine", "machine_id", machineID)
	}
	
	// Load MachineStatus
	machineStatusNode := loadMachineStatusNode(ctx, machineID, appState)
	if machineStatusNode != nil {
		if machineStatusNode.Label == "" {
			slog.Warn("MachineStatus node has empty label", "machine_id", machineID)
		} else {
			children = append(children, machineStatusNode)
		}
	} else {
		slog.Debug("No MachineStatus found for machine", "machine_id", machineID)
	}
	
	// Add link nodes: Labels, Extensions, Upgrade Status, Metrics, Config Diff
	machineLinks := createMachineActionLinkNodes(machineID, appState)
	for _, link := range machineLinks {
		if link.Label != "" {
			children = append(children, link)
		}
	}
	
	// Filter out any nodes with empty labels from the final children list
	filteredChildren := filterEmptyLabelNodes(children)
	node.Children = filteredChildren
	slog.Info("Loaded Machine children", 
		"machine_id", machineID, 
		"children_count", len(filteredChildren),
		"before_filter", len(children),
		"clustermachine", clusterMachineNode != nil,
		"machinestatus", machineStatusNode != nil)
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

// addMachineIDLinks adds machine ID link nodes if not already present
func addMachineIDLinks(node *TreeNode, appState *AppState, linkNodes []*TreeNode) []*TreeNode {
	machineIDs := findMachineIDs(node.Resource)
	for _, machineID := range machineIDs {
		if !isResourceAlreadyChild(node.Children, string(omni.MachineType), machineID) {
			linkNode := createLinkNode("link-machine", fmt.Sprintf("Machine ID: %s", machineID), machineID, func() {
				loadMachineByID(machineID, appState)
			})
			linkNodes = append(linkNodes, linkNode)
		}
	}
	return linkNodes
}

// addKubernetesVersionLinks adds Kubernetes version link nodes if not already present
func addKubernetesVersionLinks(node *TreeNode, appState *AppState, linkNodes []*TreeNode) []*TreeNode {
	k8sVersions := findKubernetesVersions(node.Resource)
	for _, version := range k8sVersions {
		if !isResourceAlreadyChild(node.Children, string(omni.KubernetesVersionType), version) {
			linkNode := createLinkNode("link-kubernetes-version", fmt.Sprintf("Kubernetes Version: %s", version), version, func() {
				loadResourcesByK8sVersion(version, appState)
			})
			linkNodes = append(linkNodes, linkNode)
		}
	}
	return linkNodes
}

// addMachineSetLinks adds machine set link nodes if not already present
func addMachineSetLinks(node *TreeNode, appState *AppState, linkNodes []*TreeNode) []*TreeNode {
	machineSetIDs := findMachineSetIDs(node.Resource)
	for _, machineSetID := range machineSetIDs {
		if !isResourceAlreadyChild(node.Children, string(omni.MachineSetType), machineSetID) {
			linkNode := createLinkNode("link-machine-set", fmt.Sprintf("Machine Set: %s", machineSetID), machineSetID, func() {
				loadMachinesByMachineSet(machineSetID, appState)
			})
			linkNodes = append(linkNodes, linkNode)
		}
	}
	return linkNodes
}

// addResourceActionLinks adds resource-specific action link nodes based on resource type
func addResourceActionLinks(resourceID, resourceType string, appState *AppState, linkNodes []*TreeNode) []*TreeNode {
	switch resourceType {
	case string(omni.ClusterType):
		return append(linkNodes, createActionLinkNodes(resourceID, appState)...)
	case string(omni.MachineType):
		return append(linkNodes, createMachineActionLinkNodes(resourceID, appState)...)
	case string(omni.MachineSetType):
		return append(linkNodes, createMachineSetActionLinkNodes(resourceID, appState)...)
	case string(omni.ClusterMachineType):
		return append(linkNodes, createClusterMachineActionLinkNodes(resourceID, appState)...)
	default:
		return linkNodes
	}
}

// addLinkNodes adds link nodes as children to a resource node
func addLinkNodes(node *TreeNode, appState *AppState, ctx context.Context) {
	if node.Resource == nil {
		return
	}
	
	linkNodes := make([]*TreeNode, 0)
	
	// Add machine_id links
	linkNodes = addMachineIDLinks(node, appState, linkNodes)
	
	// Add kubernetes_version links
	linkNodes = addKubernetesVersionLinks(node, appState, linkNodes)
	
	// Add machine_set links
	linkNodes = addMachineSetLinks(node, appState, linkNodes)
	
	// Add resource action links based on resource type
	resourceID, resourceType := extractResourceInfoFromMap(node.Resource)
	linkNodes = addResourceActionLinks(resourceID, resourceType, appState, linkNodes)
	
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
	if action != nil {
		linkActionMap[linkID] = action
	}
	
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

	burgerMenu := createBurgerMenu(myWindow, appState)

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

	mainContent := createMainLayout(burgerMenu, resourceTree, detailComponents, statusLabel)

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
