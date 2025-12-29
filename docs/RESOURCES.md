# Available Omni Resources for API Exposure

> 📖 [Back to README](README.md)

This document provides a comprehensive overview of all Omni resources available for API exposure, their implementation status, and detailed endpoint information.

## Currently Implemented ✅

### Core Resources

- **Clusters** (`ClusterType`)
  - List clusters, get cluster details
  - Status, metrics, bootstrap information
  - Endpoints, kubeconfig (⚠️ sensitive)
  - Kubernetes and Talos upgrade status
  - Kubernetes status, Kubernetes nodes, control plane status
  - Diagnostics, destroy status, workload proxy status
  - **Endpoints**: `/api/v1/clusters`, `/api/v1/clusters/:id`, `/api/v1/clusters/:id/status`, `/api/v1/clusters/:id/metrics`, `/api/v1/clusters/:id/bootstrap`, `/api/v1/clusters/:id/kubeconfig`, `/api/v1/clusters/:id/kubernetes-upgrade`, `/api/v1/clusters/:id/talos-upgrade`, `/api/v1/clusters/:id/endpoints`, `/api/v1/clusters/:id/kubernetes-status`, `/api/v1/clusters/:id/kubernetes-nodes`, `/api/v1/clusters/:id/kubernetes-nodes/:node`, `/api/v1/clusters/:id/controlplane-status`, `/api/v1/clusters/:id/diagnostics`, `/api/v1/clusters/:id/destroy-status`, `/api/v1/clusters/:id/workload-proxy-status`

- **Machines** (`MachineType`)
  - List machines, get machine details
  - Status information consolidated in main response (hostname, platform, arch, talos_version, role, maintenance, last_error)
  - Labels, extensions, upgrade status, metrics, config diff
  - **Endpoints**: `/api/v1/machines`, `/api/v1/machines/:id`, `/api/v1/machines/:id/status` (deprecated), `/api/v1/machines/:id/labels`, `/api/v1/machines/:id/extensions`, `/api/v1/machines/:id/upgrade-status`, `/api/v1/machines/:id/metrics`, `/api/v1/machines/:id/config-diff`

- **MachineSets** (`MachineSetType`)
  - List machine sets, get machine set details
  - Machine set status, destroy status
  - **Endpoints**: `/api/v1/machinesets`, `/api/v1/machinesets/:id`, `/api/v1/machinesets/:id/status`, `/api/v1/machinesets/:id/destroy-status`

- **MachineSetNodes** (`MachineSetNodeType`)
  - List machine set nodes, get node details
  - Filtering by machine set
  - **Endpoints**: `/api/v1/machinesetnodes`, `/api/v1/machinesetnodes/:id`

- **ClusterMachines** (`ClusterMachineType`)
  - List cluster machines, get cluster machine details
  - Status, config status, Talos version information, machine configuration
  - Filtering by cluster
  - **Endpoints**: `/api/v1/clustermachines`, `/api/v1/clustermachines/:id`, `/api/v1/clustermachines/:id/status`, `/api/v1/clustermachines/:id/config-status`, `/api/v1/clustermachines/:id/talos-version`, `/api/v1/clustermachines/:id/config`

### Configuration & Patches

- **ConfigPatches** (`ConfigPatchType`)
  - List config patches, get patch details
  - **Endpoints**: `/api/v1/configpatches`, `/api/v1/configpatches/:id`

- **MachineClass** (`MachineClassType`)
  - List machine classes, get machine class details
  - **Endpoints**: `/api/v1/machineclasses`, `/api/v1/machineclasses/:id`

### Machine Management

- **MachineLabels** (`MachineLabelsType`)
  - Get machine labels
  - **Endpoints**: `/api/v1/machines/:id/labels`

- **MachineExtensions** (`MachineExtensionsType`)
  - Get machine extensions
  - **Endpoints**: `/api/v1/machines/:id/extensions`

- **MachineUpgradeStatus** (`MachineUpgradeStatusType`)
  - Get machine upgrade status
  - **Endpoints**: `/api/v1/machines/:id/upgrade-status`

- **MachineSetStatus** (`MachineSetStatusType`)
  - Get machine set status
  - **Endpoints**: `/api/v1/machinesets/:id/status`

### Kubernetes Management

- **Kubeconfigs** (`KubeconfigType`)
  - Get cluster kubeconfig (⚠️ contains sensitive credentials)
  - **Endpoints**: `/api/v1/clusters/:id/kubeconfig`

- **KubernetesUpgradeStatus** (`KubernetesUpgradeStatusType`)
  - Get Kubernetes upgrade status
  - **Endpoints**: `/api/v1/clusters/:id/kubernetes-upgrade`

- **TalosUpgradeStatus** (`TalosUpgradeStatusType`)
  - Get Talos OS upgrade status
  - **Endpoints**: `/api/v1/clusters/:id/talos-upgrade`

- **KubernetesStatus** (`KubernetesStatusType`)
  - Get Kubernetes cluster status with nodes and static pods
  - **Endpoints**: `/api/v1/clusters/:id/kubernetes-status`

- **ClusterKubernetesNodes** (`ClusterKubernetesNodesType`)
  - List and get Kubernetes nodes in a cluster
  - **Endpoints**: `/api/v1/clusters/:id/kubernetes-nodes`, `/api/v1/clusters/:id/kubernetes-nodes/:node`

- **KubernetesVersion** (`KubernetesVersionType`)
  - List and get available Kubernetes versions
  - **Endpoints**: `/api/v1/kubernetes-versions`, `/api/v1/kubernetes-versions/:id`

### Cluster Management

- **ClusterEndpoints** (`ClusterEndpointType`)
  - Get cluster management endpoints
  - **Endpoints**: `/api/v1/clusters/:id/endpoints`

- **ClusterMachineStatus** (`ClusterMachineStatusType`)
  - Get cluster machine status
  - **Endpoints**: `/api/v1/clustermachines/:id/status`

- **ClusterMachineConfigStatus** (`ClusterMachineConfigStatusType`)
  - Get cluster machine configuration status
  - **Endpoints**: `/api/v1/clustermachines/:id/config-status`

- **ClusterMachineTalosVersion** (`ClusterMachineTalosVersionType`)
  - Get cluster machine Talos version
  - **Endpoints**: `/api/v1/clustermachines/:id/talos-version`

- **ClusterMachineConfig** (`ClusterMachineConfigType`)
  - Get cluster machine configuration
  - **Endpoints**: `/api/v1/clustermachines/:id/config`

- **ControlPlaneStatus** (`ControlPlaneStatusType`)
  - Get control plane health status
  - **Endpoints**: `/api/v1/clusters/:id/controlplane-status`

- **ClusterDiagnostics** (`ClusterDiagnosticsType`)
  - Cluster diagnostic information with node diagnostics
  - **Endpoints**: `/api/v1/clusters/:id/diagnostics`

- **ClusterDestroyStatus** (`ClusterDestroyStatusType`)
  - Cluster destruction status
  - **Endpoints**: `/api/v1/clusters/:id/destroy-status`

- **ClusterWorkloadProxyStatus** (`ClusterWorkloadProxyStatusType`)
  - Workload proxy status with exposed services count
  - **Endpoints**: `/api/v1/clusters/:id/workload-proxy-status`

### Backup & Recovery

- **EtcdBackups** (`EtcdBackupType`)
  - List etcd backups, get backup details
  - Filtering by cluster
  - **Endpoints**: `/api/v1/etcdbackups`, `/api/v1/etcdbackups/:id`

- **EtcdBackupStatus** (`EtcdBackupStatusType`)
  - Get etcd backup status
  - **Endpoints**: `/api/v1/etcdbackups/:id/status`

- **EtcdManualBackup** (`EtcdManualBackupType`)
  - List and get etcd manual backup requests
  - Filtering by cluster
  - **Endpoints**: `/api/v1/etcd-manual-backups`, `/api/v1/etcd-manual-backups/:id`

### Configuration

- **Schematics** (`SchematicType`)
  - List schematics, get schematic details
  - **Endpoints**: `/api/v1/schematics`, `/api/v1/schematics/:id`

- **SchematicConfiguration** (`SchematicConfigurationType`)
  - List and get schematic configurations
  - **Endpoints**: `/api/v1/schematic-configurations`, `/api/v1/schematic-configurations/:id`

### Tasks & Operations

- **OngoingTask** (`OngoingTaskType`)
  - List ongoing tasks, get task details
  - Filtering by resource
  - **Endpoints**: `/api/v1/ongoingtasks`, `/api/v1/ongoingtasks/:id`

### Configuration Management

- **ExtensionsConfiguration** (`ExtensionsConfigurationType`)
  - List and get extensions configurations
  - **Endpoints**: `/api/v1/extensions-configurations`, `/api/v1/extensions-configurations/:id`

- **KernelArgs** (`KernelArgsType`)
  - List and get kernel args configurations
  - **Endpoints**: `/api/v1/kernel-args`, `/api/v1/kernel-args/:id`

### Infrastructure

- **LoadBalancerConfig** (`LoadBalancerConfigType`)
  - List and get load balancer configurations
  - **Endpoints**: `/api/v1/loadbalancer-configs`, `/api/v1/loadbalancer-configs/:id`

- **LoadBalancerStatus** (`LoadBalancerStatusType`)
  - Get load balancer status
  - **Endpoints**: `/api/v1/loadbalancers/:id/status`

- **ExposedService** (`ExposedServiceType`)
  - List and get exposed services
  - **Endpoints**: `/api/v1/exposed-services`, `/api/v1/exposed-services/:id`

### Machine Provisioning

- **MachineRequestSet** (`MachineRequestSetType`)
  - List and get machine request sets
  - **Endpoints**: `/api/v1/machine-request-sets`, `/api/v1/machine-request-sets/:id`

### Image Management

- **ImagePullRequest** (`ImagePullRequestType`)
  - List and get image pull requests
  - **Endpoints**: `/api/v1/image-pull-requests`, `/api/v1/image-pull-requests/:id`

- **ImagePullStatus** (`ImagePullStatusType`)
  - Get image pull operation status
  - **Endpoints**: `/api/v1/image-pull-requests/:id/status`

### Installation & Infrastructure

- **InstallationMedia** (`InstallationMediaType`)
  - List and get installation media (ISO, disk images)
  - **Endpoints**: `/api/v1/installation-medias`, `/api/v1/installation-medias/:id`

- **InfraMachineConfig** (`InfraMachineConfigType`)
  - List and get infrastructure machine configurations
  - Filtering by machine ID
  - **Endpoints**: `/api/v1/infra-machine-configs`, `/api/v1/infra-machine-configs/:id`

- **MachineConfigDiff** (`MachineConfigDiffType`)
  - Machine configuration differences
  - **Endpoints**: `/api/v1/machines/:id/config-diff`

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
- **Total Endpoints**: 70+ API endpoints
- **Test Coverage**: 69.3% for handlers (exceeds 70% target threshold)
- **Coverage**: Comprehensive coverage of Omni resources including all high and medium-priority items

## Recommended Next Implementation Order

All high and medium-priority resources have been implemented! ✅

Remaining resources to consider:

1. **ClusterSecrets** - Security sensitive, requires careful consideration
2. Low-priority/internal resources (see Low Priority / Internal Resources section)

## API Features

### Hypermedia Controls

All responses include `_links` objects with URLs to related resources, enabling easy API navigation.

### Filtering Support

Many list endpoints support query parameters for filtering:

- Cluster machines: `?cluster=<cluster-id>`
- Etcd backups: `?cluster=<cluster-id>`
- Machine set nodes: `?machineset=<machineset-id>`
- Ongoing tasks: `?resource=<resource-id>`

### Status Consolidation

Machine endpoints consolidate status information directly in responses, reducing API calls. The separate status endpoint remains for backward compatibility but is deprecated.

### Full URL Generation

All links in responses are full URLs (scheme + host + path) for easy consumption by clients.

---

## Related Documentation

- 📖 [README.md](README.md) - Main project documentation
- 📊 [TEST_COVERAGE.md](TEST_COVERAGE.md) - Test coverage report
- 🔧 [MACHINE_ENHANCEMENTS.md](MACHINE_ENHANCEMENTS.md) - Machine endpoint enhancements
