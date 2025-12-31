package main

import (
	"context"
	"encoding/json"
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
	"github.com/cosi-project/runtime/pkg/resource"
	omniresources "github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"

	"github.com/jubblin/omni-api/internal/i18n"
)

// updateDetailPane updates the detail pane with resource data
func updateDetailPane(resourceData map[string]interface{}, appState *AppState) {
	appState.currentResource = resourceData

	resourceID, resourceType := extractResourceInfo(resourceData)
	updateDetailTitle(appState, resourceID)
	displayResource := loadMachineStatusIfNeeded(resourceData, resourceType, resourceID, appState)
	updateDetailJSON(displayResource, appState)
	updateDetailLinks(resourceData, resourceType, resourceID, appState)
}

// extractResourceInfo extracts resource ID and type from resource data
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

// updateDetailTitle updates the detail pane title
func updateDetailTitle(appState *AppState, resourceID string) {
	if resourceID != "" {
		appState.detailTitle.SetText(fmt.Sprintf("Resource Details: %s", resourceID))
	} else {
		appState.detailTitle.SetText("Resource Details")
	}
}

// loadMachineStatusIfNeeded loads MachineStatus if the resource is a Machine
func loadMachineStatusIfNeeded(resourceData map[string]interface{}, resourceType, resourceID string, appState *AppState) map[string]interface{} {
	if resourceType != string(omni.MachineType) || resourceID == "" {
		return resourceData
	}

	ctx := context.Background()
	md := resource.NewMetadata(omniresources.DefaultNamespace, omni.MachineStatusType, resourceID, resource.VersionUndefined)
	machineStatus, err := debugStateGet(ctx, appState.stateClient, md)
	if err != nil {
		return resourceData
	}

	ms, ok := machineStatus.(*omni.MachineStatus)
	if !ok {
		return resourceData
	}

	// Create nested status object with all MachineStatus fields
	statusInfo := make(map[string]interface{})
	spec := ms.TypedSpec().Value
	
	statusInfo["talos_version"] = spec.TalosVersion
	statusInfo["role"] = spec.Role.String()
	statusInfo["maintenance"] = spec.Maintenance
	if spec.LastError != "" {
		statusInfo["last_error"] = spec.LastError
	}
	if spec.Network != nil {
		if spec.Network.Hostname != "" {
			statusInfo["hostname"] = spec.Network.Hostname
		}
	}
	if spec.PlatformMetadata != nil {
		if spec.PlatformMetadata.Platform != "" {
			statusInfo["platform"] = spec.PlatformMetadata.Platform
		}
	}
	if spec.Hardware != nil {
		if spec.Hardware.Arch != "" {
			statusInfo["arch"] = spec.Hardware.Arch
		}
	}
	
	resourceData["status"] = statusInfo
	return resourceData
}

// updateDetailJSON updates the detail pane JSON display
func updateDetailJSON(resourceData map[string]interface{}, appState *AppState) {
	jsonBytes, err := json.MarshalIndent(resourceData, "", "  ")
	if err != nil {
		appState.detailText.SetText(fmt.Sprintf("Error formatting JSON: %v", err))
		return
	}
	// Display JSON as plain text (no markdown formatting) so it's copyable
	appState.detailText.SetText(string(jsonBytes))
}

// updateDetailLinks updates all link containers in the detail pane
func updateDetailLinks(resourceData map[string]interface{}, resourceType, resourceID string, appState *AppState) {
	clearAllLinkContainers(appState)
	updateResourceActions(resourceType, resourceID, appState)
	updateMachineIDLinks(resourceData, appState)
	updateKubernetesVersionLinks(resourceData, appState)
	updateMachineSetLinks(resourceData, appState)
	// Note: Containers are automatically shown/hidden based on content in their update functions
}

// clearAllLinkContainers clears and hides all link containers
func clearAllLinkContainers(appState *AppState) {
	appState.resourceActionsContainer.RemoveAll()
	appState.machineLinksContainer.RemoveAll()
	appState.versionLinksContainer.RemoveAll()
	appState.machineSetLinksContainer.RemoveAll()
	appState.machineRelatedLinksContainer.RemoveAll()
	// Hide all containers when cleared
	appState.resourceActionsContainer.Hide()
	appState.machineLinksContainer.Hide()
	appState.versionLinksContainer.Hide()
	appState.machineSetLinksContainer.Hide()
	appState.machineRelatedLinksContainer.Hide()
}

// updateResourceActions updates the resource actions container based on resource type
func updateResourceActions(resourceType, resourceID string, appState *AppState) {
	if resourceID == "" {
		appState.resourceActionsContainer.Hide()
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
	// Show container if it has content
	if len(appState.resourceActionsContainer.Objects) > 0 {
		appState.resourceActionsContainer.Show()
	}
}

// addActionButton adds an action button to the resource actions container
func addActionButton(label string, action func(), appState *AppState) {
	btn := widget.NewButton(label, action)
	appState.resourceActionsContainer.Add(btn)
}

// updateMachineIDLinks updates machine ID links in the detail pane
func updateMachineIDLinks(resourceData map[string]interface{}, appState *AppState) {
	machineIDs := findMachineIDs(resourceData)
	if len(machineIDs) == 0 {
		appState.machineLinksContainer.Hide()
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
	appState.machineLinksContainer.Show()
}

// findMachineIDs extracts machine IDs from resource data
func findMachineIDs(resourceData map[string]interface{}) []string {
	var machineIDs []string
	if machineID, ok := resourceData["machine_id"].(string); ok && machineID != "" {
		machineIDs = append(machineIDs, machineID)
	}
	return machineIDs
}

// updateKubernetesVersionLinks updates Kubernetes version links in the detail pane
func updateKubernetesVersionLinks(resourceData map[string]interface{}, appState *AppState) {
	k8sVersions := findKubernetesVersions(resourceData)
	if len(k8sVersions) == 0 {
		appState.versionLinksContainer.Hide()
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
	appState.versionLinksContainer.Show()
}

// findKubernetesVersions extracts Kubernetes versions from resource data
func findKubernetesVersions(resourceData map[string]interface{}) []string {
	var versions []string
	if version, ok := resourceData["kubernetes_version"].(string); ok && version != "" {
		versions = append(versions, version)
	}
	return versions
}

// findMachineSetIDs extracts machine set IDs from resource data
func findMachineSetIDs(resourceData map[string]interface{}) []string {
	var machineSetIDs []string
	if labels, ok := resourceData["labels"].(map[string]interface{}); ok {
		if machineSet, ok := labels[labelKeyMachineSet].(string); ok && machineSet != "" {
			machineSetIDs = append(machineSetIDs, machineSet)
		}
	}
	return machineSetIDs
}

// updateMachineSetLinks updates machine set links in the detail pane
func updateMachineSetLinks(resourceData map[string]interface{}, appState *AppState) {
	machineSetIDs := findMachineSetIDs(resourceData)
	if len(machineSetIDs) == 0 {
		appState.machineSetLinksContainer.Hide()
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
	appState.machineSetLinksContainer.Show()
}
