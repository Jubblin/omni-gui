package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosi-project/runtime/pkg/resource"
	omniresources "github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
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
