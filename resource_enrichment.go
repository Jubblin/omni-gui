package main

import (
	"context"

	"github.com/cosi-project/runtime/pkg/resource"
	"github.com/cosi-project/runtime/pkg/state"
	omniresources "github.com/siderolabs/omni/client/pkg/omni/resources"
	"github.com/siderolabs/omni/client/pkg/omni/resources/omni"
)

// buildMachineStatusInfo extracts status information from a MachineStatus resource
func buildMachineStatusInfo(ms *omni.MachineStatus) map[string]interface{} {
	statusInfo := make(map[string]interface{})
	spec := ms.TypedSpec().Value
	statusInfo["talos_version"] = spec.TalosVersion
	statusInfo["role"] = spec.Role.String()
	statusInfo["maintenance"] = spec.Maintenance
	if spec.LastError != "" {
		statusInfo["last_error"] = spec.LastError
	}
	if spec.Network != nil && spec.Network.Hostname != "" {
		statusInfo["hostname"] = spec.Network.Hostname
	}
	if spec.PlatformMetadata != nil && spec.PlatformMetadata.Platform != "" {
		statusInfo["platform"] = spec.PlatformMetadata.Platform
	}
	if spec.Hardware != nil && spec.Hardware.Arch != "" {
		statusInfo["arch"] = spec.Hardware.Arch
	}
	return statusInfo
}

// enrichMachineResource enriches a Machine resource with MachineStatus information
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

	resourceMap["status"] = buildMachineStatusInfo(ms)
}

// enrichClusterMachineResource enriches a ClusterMachine resource with MachineStatus information
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

	resourceMap["status"] = buildMachineStatusInfo(ms)
}
