# Available Omni Resources

> 📖 [Back to README](README.md)

This document provides a comprehensive overview of all Omni resources available in the GUI application, their implementation status, and the underlying gRPC API calls used to access them.

## API Call Methods

Resources are accessed using the Omni client's state interface, which provides gRPC-based access:

- **`state.List(ctx, metadata)`** - Lists all resources of a given type
- **`state.Get(ctx, metadata)`** - Retrieves a single resource by ID
- **`state.Watch(ctx, metadata)`** - Watches for resource changes (not currently used in GUI)

All API calls use `resource.NewMetadata()` to create metadata with:
- Namespace: `omniresources.DefaultNamespace`
- Type: The specific resource type (e.g., `omni.ClusterType`)
- ID: Resource identifier (empty string for List operations)
- Version: `resource.VersionUndefined`

## Currently Implemented ✅

### Core Resources

- **Clusters** (`ClusterType`)
  - List clusters, get cluster details
  - Status, metrics, bootstrap information
  - Endpoints, kubeconfig (⚠️ sensitive)
  - Kubernetes and Talos upgrade status
  - Kubernetes status, Kubernetes nodes, control plane status
  - Diagnostics, destroy status, workload proxy status
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.ClusterType, "", version))` - List all clusters
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterType, clusterID, version))` - Get single cluster
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterStatusType, clusterID, version))` - Get cluster status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterMetricsType, clusterID, version))` - Get cluster metrics
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.KubeconfigType, clusterID, version))` - Get kubeconfig
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.KubernetesUpgradeStatusType, clusterID, version))` - Get Kubernetes upgrade status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.TalosUpgradeStatusType, clusterID, version))` - Get Talos upgrade status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterEndpointType, clusterID, version))` - Get cluster endpoints
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.KubernetesStatusType, clusterID, version))` - Get Kubernetes status
    - `state.List(ctx, resource.NewMetadata(namespace, omni.ClusterKubernetesNodesType, "", version))` - List Kubernetes nodes (filtered by cluster)
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ControlPlaneStatusType, clusterID, version))` - Get control plane status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterDiagnosticsType, clusterID, version))` - Get diagnostics
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterDestroyStatusType, clusterID, version))` - Get destroy status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterWorkloadProxyStatusType, clusterID, version))` - Get workload proxy status

- **Machines** (`MachineType`)
  - List machines, get machine details
  - Status information consolidated in main response (hostname, platform, arch, talos_version, role, maintenance, last_error)
  - Labels, extensions, upgrade status, metrics, config diff
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.MachineType, "", version))` - List all machines
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineType, machineID, version))` - Get single machine
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineStatusType, machineID, version))` - Get machine status (consolidated in main response)
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineLabelsType, machineID, version))` - Get machine labels
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineExtensionsType, machineID, version))` - Get machine extensions
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineUpgradeStatusType, machineID, version))` - Get upgrade status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineStatusMetricsType, machineID, version))` - Get machine metrics
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineConfigDiffType, machineID, version))` - Get config diff

- **MachineSets** (`MachineSetType`)
  - List machine sets, get machine set details
  - Machine set status, destroy status
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.MachineSetType, "", version))` - List all machine sets
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineSetType, machineSetID, version))` - Get single machine set
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineSetStatusType, machineSetID, version))` - Get machine set status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineSetDestroyStatusType, machineSetID, version))` - Get destroy status

- **MachineSetNodes** (`MachineSetNodeType`)
  - List machine set nodes, get node details
  - Filtering by machine set (via labels: `omni.sidero.dev/machine-set`)
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.MachineSetNodeType, "", version))` - List all machine set nodes (filter by machine set label)
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineSetNodeType, nodeID, version))` - Get single machine set node

- **ClusterMachines** (`ClusterMachineType`)
  - List cluster machines, get cluster machine details
  - Status, config status, Talos version information, machine configuration
  - Filtering by cluster (via labels: `omni.sidero.dev/cluster`)
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.ClusterMachineType, "", version))` - List all cluster machines (filter by cluster label)
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterMachineType, clusterMachineID, version))` - Get single cluster machine
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterMachineStatusType, clusterMachineID, version))` - Get cluster machine status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterMachineConfigStatusType, clusterMachineID, version))` - Get config status
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterMachineTalosVersionType, clusterMachineID, version))` - Get Talos version
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ClusterMachineConfigType, clusterMachineID, version))` - Get machine configuration

### Configuration & Patches

- **ConfigPatches** (`ConfigPatchType`)
  - List config patches, get patch details
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.ConfigPatchType, "", version))` - List all config patches
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.ConfigPatchType, patchID, version))` - Get single config patch

- **MachineClass** (`MachineClassType`)
  - List machine classes, get machine class details
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.MachineClassType, "", version))` - List all machine classes
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.MachineClassType, classID, version))` - Get single machine class

### Machine Management

- **MachineLabels** (`MachineLabelsType`)
  - Get machine labels

- **MachineExtensions** (`MachineExtensionsType`)
  - Get machine extensions

- **MachineUpgradeStatus** (`MachineUpgradeStatusType`)
  - Get machine upgrade status

- **MachineSetStatus** (`MachineSetStatusType`)
  - Get machine set status

### Kubernetes Management

- **Kubeconfigs** (`KubeconfigType`)
  - Get cluster kubeconfig (⚠️ contains sensitive credentials)
  - **API Calls**:
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.KubeconfigType, clusterID, version))` - Get cluster kubeconfig

- **KubernetesUpgradeStatus** (`KubernetesUpgradeStatusType`)
  - Get Kubernetes upgrade status
  - **API Calls**:
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.KubernetesUpgradeStatusType, clusterID, version))` - Get Kubernetes upgrade status

- **TalosUpgradeStatus** (`TalosUpgradeStatusType`)
  - Get Talos OS upgrade status
  - **API Calls**:
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.TalosUpgradeStatusType, clusterID, version))` - Get Talos upgrade status

- **KubernetesStatus** (`KubernetesStatusType`)
  - Get Kubernetes cluster status with nodes and static pods

- **ClusterKubernetesNodes** (`ClusterKubernetesNodesType`)
  - List and get Kubernetes nodes in a cluster

- **KubernetesVersion** (`KubernetesVersionType`)
  - List and get available Kubernetes versions
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.KubernetesVersionType, "", version))` - List all Kubernetes versions
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.KubernetesVersionType, versionID, version))` - Get single Kubernetes version

### Cluster Management

- **ClusterEndpoints** (`ClusterEndpointType`)
  - Get cluster management endpoints

- **ClusterMachineStatus** (`ClusterMachineStatusType`)
  - Get cluster machine status

- **ClusterMachineConfigStatus** (`ClusterMachineConfigStatusType`)
  - Get cluster machine configuration status

- **ClusterMachineTalosVersion** (`ClusterMachineTalosVersionType`)
  - Get cluster machine Talos version

- **ClusterMachineConfig** (`ClusterMachineConfigType`)
  - Get cluster machine configuration

- **ControlPlaneStatus** (`ControlPlaneStatusType`)
  - Get control plane health status

- **ClusterDiagnostics** (`ClusterDiagnosticsType`)
  - Cluster diagnostic information with node diagnostics

- **ClusterDestroyStatus** (`ClusterDestroyStatusType`)
  - Cluster destruction status

- **ClusterWorkloadProxyStatus** (`ClusterWorkloadProxyStatusType`)
  - Workload proxy status with exposed services count

### Backup & Recovery

- **EtcdBackups** (`EtcdBackupType`)
  - List etcd backups, get backup details
  - Filtering by cluster (via labels: `omni.sidero.dev/cluster`)
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.EtcdBackupType, "", version))` - List all etcd backups (filter by cluster label)
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.EtcdBackupType, backupID, version))` - Get single etcd backup

- **EtcdBackupStatus** (`EtcdBackupStatusType`)
  - Get etcd backup status
  - **API Calls**:
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.EtcdBackupStatusType, backupID, version))` - Get etcd backup status

- **EtcdManualBackup** (`EtcdManualBackupType`)
  - List and get etcd manual backup requests
  - Filtering by cluster (via labels: `omni.sidero.dev/cluster`)
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.EtcdManualBackupType, "", version))` - List all etcd manual backups (filter by cluster label)
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.EtcdManualBackupType, backupID, version))` - Get single etcd manual backup

### Configuration

- **Schematics** (`SchematicType`)
  - List schematics, get schematic details

- **SchematicConfiguration** (`SchematicConfigurationType`)
  - List and get schematic configurations

### Tasks & Operations

- **OngoingTask** (`OngoingTaskType`)
  - List ongoing tasks, get task details
  - Filtering by resource (via `spec.ResourceId` field)
  - **API Calls**:
    - `state.List(ctx, resource.NewMetadata(namespace, omni.OngoingTaskType, "", version))` - List all ongoing tasks (filter by resource ID in application code)
    - `state.Get(ctx, resource.NewMetadata(namespace, omni.OngoingTaskType, taskID, version))` - Get single ongoing task

### Configuration Management

- **ExtensionsConfiguration** (`ExtensionsConfigurationType`)
  - List and get extensions configurations

- **KernelArgs** (`KernelArgsType`)
  - List and get kernel args configurations

### Infrastructure

- **LoadBalancerConfig** (`LoadBalancerConfigType`)
  - List and get load balancer configurations

- **LoadBalancerStatus** (`LoadBalancerStatusType`)
  - Get load balancer status

- **ExposedService** (`ExposedServiceType`)
  - List and get exposed services

### Machine Provisioning

- **MachineRequestSet** (`MachineRequestSetType`)
  - List and get machine request sets

### Image Management

- **ImagePullRequest** (`ImagePullRequestType`)
  - List and get image pull requests

- **ImagePullStatus** (`ImagePullStatusType`)
  - Get image pull operation status

### Installation & Infrastructure

- **InstallationMedia** (`InstallationMediaType`)
  - List and get installation media (ISO, disk images)

- **InfraMachineConfig** (`InfraMachineConfigType`)
  - List and get infrastructure machine configurations
  - Filtering by machine ID

- **MachineConfigDiff** (`MachineConfigDiffType`)
  - Machine configuration differences

## High Priority Resources to Add

All high-priority resources have been implemented! ✅

The following resources remain for future consideration:

### Security-Sensitive Resources

- **ClusterSecrets** (`ClusterSecretsType`)
  - Secrets stored for clusters
  - Useful for: Secret management
  - ⚠️ Security: Contains sensitive data
  - **Priority**: Low - Security sensitive, requires careful consideration for read-only access or redaction

## Medium Priority Resources

All medium-priority resources have been implemented! ✅

## Low Priority / Internal Resources

These are typically internal or less commonly accessed:

- **ClusterConfigVersion** - Internal config versioning
- **ClusterUUID** - Internal UUID tracking
- **MachineStatusLink** - Internal status links
- **MachineStatusSnapshot** - Internal snapshots
- **BackupData** - Internal backup data
- **ClusterMachineIdentity** - Internal identity management
- **ClusterMachineEncryptionKey** - Internal encryption keys
- **ClusterMachineTemplate** - Internal templates

## Security Considerations

Resources marked with ⚠️ contain sensitive information:

- **Kubeconfigs** - Contains cluster credentials
- **ClusterSecrets** - Contains secret data
- **ClusterMachineEncryptionKey** - Contains encryption keys

These should be:

- Protected with authentication/authorization
- Only exposed to authorized users
- Considered for read-only access or redaction

## Implementation Statistics

- **Total Resources Implemented**: 38+ resource types
- **Test Coverage**: 69.3% for handlers (exceeds 70% target threshold)
- **Coverage**: Comprehensive coverage of Omni resources including all high and medium-priority items

## Recommended Next Implementation Order

All high and medium-priority resources have been implemented! ✅

Remaining resources to consider:

1. **ClusterSecrets** - Security sensitive, requires careful consideration
2. Low-priority/internal resources (see Low Priority / Internal Resources section)

## API Features

### API Call Patterns

All resources follow consistent API call patterns using the Omni client's state interface:

1. **List Operations**: Use `state.List(ctx, metadata)` with empty resource ID to retrieve all resources of a type
2. **Get Operations**: Use `state.Get(ctx, metadata)` with specific resource ID to retrieve a single resource
3. **Filtering**: Performed in application code by:
   - Checking resource labels (e.g., `omni.sidero.dev/cluster`, `omni.sidero.dev/machine-set`)
   - Checking resource spec fields (e.g., `spec.ResourceId` for ongoing tasks)
4. **Metadata Creation**: All calls use `resource.NewMetadata()` with:
   - Namespace: `omniresources.DefaultNamespace`
   - Type: Specific resource type constant (e.g., `omni.ClusterType`)
   - ID: Resource identifier (empty string for List operations)
   - Version: `resource.VersionUndefined`

**Example API Call:**
```go
// List all clusters
md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterType, "", resource.VersionUndefined)
items, err := stateClient.List(ctx, md)

// Get a specific cluster
md := resource.NewMetadata(omniresources.DefaultNamespace, omni.ClusterType, clusterID, resource.VersionUndefined)
cluster, err := stateClient.Get(ctx, md)
```

### Filtering Support

Resources can be filtered in application code:

- Cluster machines: Filtered via `omni.sidero.dev/cluster` label
- Etcd backups: Filtered via `omni.sidero.dev/cluster` label
- Machine set nodes: Filtered via `omni.sidero.dev/machine-set` label
- Ongoing tasks: Filtered via `spec.ResourceId` field

---

## Related Documentation

- 📖 [README.md](README.md) - Main project documentation
- 📊 [TEST_COVERAGE.md](TEST_COVERAGE.md) - Test coverage report
- 🔧 [MACHINE_ENHANCEMENTS.md](MACHINE_ENHANCEMENTS.md) - Machine endpoint enhancements
