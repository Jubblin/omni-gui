package main

import (
	"fmt"

	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
)

// extractHostnameFromResourceMap extracts hostname from a resource map
func extractHostnameFromResourceMap(resourceMap map[string]interface{}) string {
	if status, ok := resourceMap["status"].(map[string]interface{}); ok {
		if h, ok := status["hostname"].(string); ok && h != "" {
			return h
		}
	}
	if h, ok := resourceMap["hostname"].(string); ok && h != "" {
		return h
	}
	return ""
}

// formatClusterLabel formats label for Cluster resource
func formatClusterLabel(resourceID string, resourceMap map[string]interface{}) string {
	if kv, ok := resourceMap["kubernetes_version"].(string); ok && kv != "" {
		return fmt.Sprintf("%s (K8s: %s)", resourceID, kv)
	}
	return resourceID
}

// formatMachineLabel formats label for Machine resource
func formatMachineLabel(resourceID string, resourceMap map[string]interface{}) string {
	hostname := extractHostnameFromResourceMap(resourceMap)
	if hostname != "" {
		return hostname
	}
	if addr, ok := resourceMap["management_address"].(string); ok && addr != "" {
		return fmt.Sprintf("%s (%s)", resourceID, addr)
	}
	return resourceID
}

// formatClusterMachineLabel formats label for ClusterMachine resource
func formatClusterMachineLabel(resourceID string, resourceMap map[string]interface{}) string {
	hostname := extractHostnameFromResourceMap(resourceMap)
	if hostname != "" {
		return hostname
	}
	if machineID, ok := resourceMap["machine_id"].(string); ok && machineID != "" {
		return fmt.Sprintf("%s (Machine: %s)", resourceID, machineID)
	}
	// Ensure we never return an empty string
	if resourceID != "" {
		return resourceID
	}
	// Last resort fallback
	return "ClusterMachine"
}

// formatMachineSetLabel formats label for MachineSet resource
func formatMachineSetLabel(resourceID string, resourceMap map[string]interface{}) string {
	if mc, ok := resourceMap["machine_class"].(string); ok && mc != "" {
		return fmt.Sprintf("%s (Class: %s)", resourceID, mc)
	}
	return resourceID
}

// formatResourceLabel formats a label for a resource based on its type
func formatResourceLabel(resourceMap map[string]interface{}) string {
	resourceID := ""
	if id, ok := resourceMap["id"].(string); ok {
		resourceID = id
	}
	resourceType := ""
	if t, ok := resourceMap["type"].(string); ok {
		resourceType = t
	}

	switch resourceType {
	case string(omni.ClusterType):
		return formatClusterLabel(resourceID, resourceMap)
	case string(omni.MachineType):
		return formatMachineLabel(resourceID, resourceMap)
	case string(omni.ClusterMachineType):
		return formatClusterMachineLabel(resourceID, resourceMap)
	case string(omni.MachineSetType):
		return formatMachineSetLabel(resourceID, resourceMap)
	default:
		return resourceID
	}
}

// createResourceNodeFromMap creates a TreeNode from a resource map
func createResourceNodeFromMap(resourceMap map[string]interface{}, resourceID, resourceType string) *TreeNode {
	label := formatResourceLabel(resourceMap)
	// Ensure label is never empty - use resourceID as fallback
	if label == "" {
		if resourceID != "" {
			label = resourceID
		} else {
			label = fmt.Sprintf("%s (no ID)", resourceType)
		}
	}
	return &TreeNode{
		ID:       fmt.Sprintf("%s-%s", resourceType, resourceID),
		Type:     resourceType,
		Label:    label,
		Resource: resourceMap,
		Children: []*TreeNode{},
	}
}

// extractResourceInfoFromMap extracts resource ID and type from a resource map
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
