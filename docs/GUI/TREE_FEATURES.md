# Tree Features

This document describes the key features and implementation details of the resource tree in the Fyne GUI application.

## Key Features

### 1. Lazy Loading

- Resources are loaded on-demand when a branch is expanded.
- This improves initial load time and reduces memory usage.

### 2. Folder Grouping

- **Clusters**: Grouped under a "Clusters" folder when listing all clusters.
- **MachineSets**: Grouped under a "MachineSets" folder, sorted alphabetically by name.
- **Machines**: Listed directly as children of the "Machines" resource-type-folder (no nested folder), sorted alphabetically by label.

### 3. Link Nodes

Resource-specific actions appear as clickable leaf nodes under each resource:

- **Clusters**: Status, Metrics, Bootstrap, Kubeconfig, Kubernetes Upgrade, Talos Upgrade, Endpoints, Kubernetes Status, Control Plane Status, Diagnostics, Destroy Status, Workload Proxy Status
- **Machines**: Labels, Extensions, Upgrade Status, Metrics, Config Diff
- **MachineSets**: Status, Destroy Status
- **ClusterMachines**: Status, Config Status, Talos Version, Config

### 4. Reverse Lookups

The tree supports bidirectional navigation:

- **ClusterMachine → Machine**: Via `machine_id` field
- **Machine → ClusterMachine**: Since ClusterMachine ID equals Machine ID
- **ClusterMachine → Cluster**: From `omni.sidero.dev/cluster` label
- **ClusterMachine → MachineSet**: From `omni.sidero.dev/machine-set` label

### 5. Orphaned Resources

- ClusterMachines that don't belong to any MachineSet are shown directly under their Cluster.
- This makes it easy to identify machines that aren't part of a MachineSet.

### 6. Related Resources

- KubernetesVersion resources appear under Clusters when the cluster has a `kubernetes_version` field.
- MachineStatus appears under Machines.

### 7. Refresh Functionality

- The burger menu (☰ Menu) includes a "Refresh" option that rebuilds the internal tree data structure.
- When selected, it:
  - Clears the existing node map
  - Rebuilds the initial tree structure from scratch
  - Resets the tree to its initial state (clearing any loaded children)
  - Refreshes the tree widget display
  - Clears the detail pane
  - Updates the status label to confirm the refresh
- This is useful when you want to reset the tree view or reload the structure after changes in the Omni instance.

## Implementation Details

### Node Types

- **`root`**: The invisible root node
- **`resource-type-folder`**: Top-level folders (Clusters, Machines, MachineSets)
- **`clusters-folder`**: Folder containing all clusters
- **`machinesets-folder`**: Folder containing MachineSets (sorted)
- **`machines-folder`**: Folder containing all machines
- **`link-*`**: Link nodes for actions (always leaves)

### Loading Functions

- **`loadResourceTypeFolderChildren`**: Loads resources when a top-level folder is expanded
- **`loadClusterChildren`**: Loads MachineSets, orphaned ClusterMachines, and KubernetesVersion for a cluster
- **`loadMachineSetChildren`**: Loads ClusterMachines for a MachineSet
- **`loadClusterMachineChildren`**: Loads Machine, Cluster, and MachineSet for a ClusterMachine
- **`loadMachineChildren`**: Loads ClusterMachine (reverse lookup with "(ClusterMachine)" prefix), MachineStatus, and link nodes for a Machine. Pre-loads ClusterMachine children (Machine, Cluster, MachineSet, and link nodes).
- **`addLinkNodes`**: Adds action link nodes based on resource type
