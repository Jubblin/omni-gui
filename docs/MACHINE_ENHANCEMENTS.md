# Machine Endpoint Enhancement Options

> 📖 [Back to README](README.md)

This document details the enhancements made to machine-related endpoints, including status consolidation, additional fields, and related resource links.

## Currently Exposed ✅

- **ID** - Machine identifier
- **Namespace** - Resource namespace
- **ManagementAddress** - IP for Talos API access
- **Connected** - Connection status
- **UseGrpcTunnel** - gRPC tunnel usage flag
- **Labels** - All metadata labels (✅ Implemented)
- **Hostname** - Machine hostname from MachineStatus (✅ Implemented)
- **Platform** - Platform information from MachineStatus (✅ Implemented)
- **Arch** - Architecture from MachineStatus (✅ Implemented)
- **TalosVersion** - Talos version from MachineStatus (✅ Implemented)
- **Links** - Self, cluster, labels, extensions, upgrade-status links (✅ Implemented)

## Additional Fields That Can Be Exposed

### 1. **Metadata Labels** ✅ IMPLEMENTED

All labels from machine metadata are now exposed:

- `omni.sidero.dev/cluster` - Cluster association (also linked)
- `omni.sidero.dev/role-controlplane` - Control plane role
- `omni.sidero.dev/address` - Machine address
- `omni.sidero.dev/platform` - Platform information
- Custom user labels

**Implementation**: ✅ Added `Labels map[string]string` field that exposes all metadata labels via `m.Metadata().Labels().Raw()`.

### 2. **Additional Links** ✅ IMPLEMENTED

Links to related resources are now included:

- **self** - `/api/v1/machines/:id` (✅ Implemented)
- **cluster** - `/api/v1/clusters/:id` (if machine is part of a cluster) (✅ Implemented)
- **labels** - `/api/v1/machines/:id/labels` (✅ Implemented)
- **extensions** - `/api/v1/machines/:id/extensions` (✅ Implemented)
- **upgrade-status** - `/api/v1/machines/:id/upgrade-status` (✅ Implemented)
- **status** - `/api/v1/machines/:id/status` (MachineStatus endpoint - deprecated, status is now consolidated)
- **clustermachine** - Not automatically linked (would require reverse lookup)

### 3. **MachineLabels Resource** ✅ IMPLEMENTED

- User-defined labels attached to machines
- Useful for: Machine organization, filtering, custom metadata
- **Endpoint**: `/api/v1/machines/:id/labels` (✅ Implemented)
- **Handler**: `MachineLabelsHandler` in `machinelabels.go`

### 4. **MachineExtensions Resource** ✅ IMPLEMENTED

- List of Talos extensions installed on the machine
- **Fields**: `extensions []string`
- **Endpoint**: `/api/v1/machines/:id/extensions` (✅ Implemented)
- **Handler**: `MachineExtensionsHandler` in `machineextensions.go`
- Useful for: Managing and tracking installed extensions

### 5. **MachineUpgradeStatus Resource** ✅ IMPLEMENTED

- Upgrade status and progress for machines
- **Fields**:
  - `schematic_id` - Target schematic
  - `talos_version` - Target Talos version
  - `current_schematic_id` - Current schematic
  - `current_talos_version` - Current Talos version
  - `phase` - Upgrade phase (Idle, Upgrading, Failed, etc.)
  - `status` - Status message
  - `error` - Error message if failed
  - `is_maintenance` - Maintenance mode flag
- **Endpoint**: `/api/v1/machines/:id/upgrade-status` (✅ Implemented)
- **Handler**: `MachineUpgradeStatusHandler` in `machineupgradestatus.go`
- Useful for: Tracking upgrade progress

### 6. **MachineStatus Summary Fields** ✅ FULLY IMPLEMENTED

All key status fields are now included directly in the machine response:

- **Hostname** - From MachineStatus.Network.Hostname (✅ Implemented)
- **Platform** - From MachineStatus.PlatformMetadata.Platform (✅ Implemented)
- **Architecture** - From MachineStatus.Hardware.Arch (✅ Implemented)
- **TalosVersion** - From MachineStatus.TalosVersion (✅ Implemented)
- **Role** - From MachineStatus.Role (controlplane/worker/none) (✅ Implemented)
- **Maintenance** - From MachineStatus.Maintenance (✅ Implemented)
- **LastError** - From MachineStatus.LastError (✅ Implemented)

**Note**: The implementation fetches MachineStatus resource when available, consolidating all commonly needed fields. The separate `/status` endpoint remains for backward compatibility but is deprecated. Cluster information is already available via labels.

### 7. **Machine Status Metrics** ✅ IMPLEMENTED

- Aggregated metrics for all machines in Omni
- **Resource Type**: `MachineStatusMetricsType`
- **Endpoint**: `/api/v1/machines/:id/metrics` (✅ Implemented)
- **Handler**: `MachineStatusMetricsHandler` in `machinestatusmetrics.go`
- **Fields**: Registered machines count, connected machines count, allocated machines count, pending machines count, platforms distribution, secure boot status, UKI status

## Implementation Status

### ✅ Completed Enhancements

1. **Links to Related Resources** - All links are now included in MachineResponse (including metrics)
2. **Metadata Labels** - All labels are exposed via `Labels` field
3. **MachineLabels Handler** - Implemented at `/api/v1/machines/:id/labels`
4. **MachineExtensions Handler** - Implemented at `/api/v1/machines/:id/extensions`
5. **MachineUpgradeStatus Handler** - Implemented at `/api/v1/machines/:id/upgrade-status`
6. **Status Consolidation** - All status fields (hostname, platform, arch, talos_version, role, maintenance, last_error) are included directly
7. **MachineStatusMetrics Handler** - Implemented at `/api/v1/machines/:id/metrics` for aggregated machine metrics

### 🔄 Potential Future Enhancements

1. **ClusterMachine Link** - Could add reverse lookup to link to ClusterMachine resource
2. **Additional Metrics** - Could add per-machine metrics if available in future Omni versions

## Implementation Considerations

1. **Performance**: Adding MachineStatus fields directly would require an additional resource fetch, which could slow down list operations. Links are preferred.

2. **Data Freshness**: MachineStatus data changes frequently. Including it directly in Machine response might show stale data if cached.

3. **API Design**: Following REST principles, related resources should be separate endpoints with links, not embedded data.

4. **Backward Compatibility**: Adding new fields should not break existing clients.

## Current Response Format

The machine endpoint now returns a consolidated response with all implemented enhancements:

```json
{
  "id": "machine-123",
  "namespace": "default",
  "management_address": "192.168.1.10",
  "connected": true,
  "use_grpc_tunnel": false,
  "labels": {
    "omni.sidero.dev/cluster": "cluster-1",
    "omni.sidero.dev/role-controlplane": "true",
    "custom-label": "value"
  },
  "hostname": "talos-node-1",
  "platform": "metal",
  "arch": "amd64",
  "talos_version": "1.7.0",
  "_links": {
    "self": "http://localhost:8080/api/v1/machines/machine-123",
    "cluster": "http://localhost:8080/api/v1/clusters/cluster-1",
    "labels": "http://localhost:8080/api/v1/machines/machine-123/labels",
    "extensions": "http://localhost:8080/api/v1/machines/machine-123/extensions",
    "upgrade-status": "http://localhost:8080/api/v1/machines/machine-123/upgrade-status"
  }
}
```

**Note**: The `status` link is not included in the main response since status information is now consolidated. The separate `/status` endpoint remains available but is deprecated.

---

## Related Documentation

- 📖 [README.md](README.md) - Main project documentation
- 📋 [RESOURCES.md](RESOURCES.md) - Available Omni resources
- 📊 [TEST_COVERAGE.md](TEST_COVERAGE.md) - Test coverage report
