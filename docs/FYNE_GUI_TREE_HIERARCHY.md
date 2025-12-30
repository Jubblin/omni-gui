# Fyne GUI Tree Hierarchy

This document describes the tree structure hierarchy implemented in the Fyne GUI application for browsing Omni resources.

## Overview

The Fyne GUI displays resources in a hierarchical tree structure that reflects the relationships between different Omni resource types. The tree uses lazy loading - resources are loaded when their parent nodes are expanded.

## Root Level Structure

The root of the tree (invisible) contains three top-level resource type folders:

1. **Clusters** (`resource-type-folder`)
2. **Machines** (`resource-type-folder`)
3. **MachineSets** (`resource-type-folder`)

## 1. Clusters Hierarchy

```
Clusters (resource-type-folder)
└── Clusters (folder) [when expanded, loads all clusters]
    └── Cluster 1
        ├── MachineSets (folder) [sorted by name]
        │   └── MachineSet 1
        │       └── ClusterMachine 1
        │           ├── Machine (via machine_id)
        │           ├── Cluster (reverse lookup from labels)
        │           ├── MachineSet (reverse lookup from labels)
        │           └── [Link nodes: Status, Config Status, Talos Version, Config]
        ├── ClusterMachines (orphaned - not in any MachineSet)
        │   └── ClusterMachine X
        │       ├── Machine
        │       ├── Cluster
        │       └── [Link nodes]
        ├── KubernetesVersion (if cluster has kubernetes_version)
        └── [Link nodes: Status, Metrics, Bootstrap, Kubeconfig, Kubernetes Upgrade, 
             Talos Upgrade, Endpoints, Kubernetes Status, Control Plane Status, 
             Diagnostics, Destroy Status, Workload Proxy Status]
```

### Cluster Children Details

- **MachineSets Folder**: Contains all MachineSets belonging to the cluster (identified by `omni.sidero.dev/cluster` label). MachineSets are sorted alphabetically by name.
- **Orphaned ClusterMachines**: ClusterMachines that belong to the cluster but are not part of any MachineSet (no `omni.sidero.dev/machine-set` label).
- **KubernetesVersion**: If the cluster has a `kubernetes_version` field, the corresponding KubernetesVersion resource is shown as a child.
- **Link Nodes**: Action links that allow accessing related resources like Status, Metrics, Bootstrap, etc.

## 2. Machines Hierarchy

```
Machines (resource-type-folder)
└── Machines (folder) [when expanded, loads all machines]
    └── Machine 1
        ├── ClusterMachine (reverse lookup - ClusterMachine ID = Machine ID)
        │   ├── Machine (back reference)
        │   ├── Cluster
        │   ├── MachineSet
        │   └── [Link nodes]
        ├── MachineStatus (same ID as Machine)
        └── [Link nodes: Labels, Extensions, Upgrade Status, Metrics, Config Diff]
```

### Machine Children Details

- **ClusterMachine**: Reverse lookup - since ClusterMachine ID equals Machine ID, the associated ClusterMachine is shown as a child.
- **MachineStatus**: Status information for the machine (same ID as the Machine). When displayed in the detail pane, MachineStatus information appears as a nested `status` object containing fields like `hostname`, `platform`, `arch`, `talos_version`, `role`, `maintenance`, and `last_error`.
- **Link Nodes**: Action links for machine-specific operations like Labels, Extensions, Upgrade Status, Metrics, and Config Diff.

## 3. MachineSets Hierarchy

```
MachineSets (resource-type-folder)
└── MachineSets (folder) [when expanded, loads all MachineSets, sorted by name]
    └── MachineSet 1
        └── ClusterMachines
            └── ClusterMachine 1
                ├── Machine
                ├── Cluster
                ├── MachineSet
                └── [Link nodes]
```

### MachineSet Children Details

- **ClusterMachines**: All ClusterMachines belonging to the MachineSet (identified by `omni.sidero.dev/machine-set` label).
- Each ClusterMachine shows its related Machine, Cluster, and MachineSet as children.
- **Link Nodes**: Action links for ClusterMachine-specific operations.

## Key Features

### 1. Lazy Loading
- Resources are loaded on-demand when a branch is expanded.
- This improves initial load time and reduces memory usage.

### 2. Folder Grouping
- **Clusters**: Grouped under a "Clusters" folder when listing all clusters.
- **MachineSets**: Grouped under a "MachineSets" folder, sorted alphabetically by name.
- **Machines**: Grouped under a "Machines" folder when listing all machines.

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
- **`loadMachineChildren`**: Loads ClusterMachine (reverse lookup) and MachineStatus for a Machine
- **`addLinkNodes`**: Adds action link nodes based on resource type

### Resource Relationships

The hierarchy reflects the actual API relationships:
- Clusters contain MachineSets (via `omni.sidero.dev/cluster` label)
- MachineSets contain ClusterMachines (via `omni.sidero.dev/machine-set` label)
- ClusterMachines reference Machines (via `machine_id` field)
- Machines can reference ClusterMachines (reverse lookup)

## Usage

1. **Expand Clusters**: Click on "Clusters" to see all clusters, then expand a cluster to see its MachineSets and ClusterMachines.
2. **Expand Machines**: Click on "Machines" to see all machines, then expand a machine to see its ClusterMachine and MachineStatus.
3. **Expand MachineSets**: Click on "MachineSets" to see all MachineSets, then expand a MachineSet to see its ClusterMachines.
4. **Click Link Nodes**: Click on link nodes (like "Status", "Metrics") to load related resources in the detail pane.
5. **Navigate Relationships**: Use the tree to navigate between related resources (e.g., from a ClusterMachine to its Machine or Cluster).
6. **Refresh Tree Data**: Click the burger menu (☰ Menu) in the top-left corner and select "Refresh" to rebuild the internal tree structure. This resets the tree to its initial state, clearing all loaded children and starting fresh.

## Notes

- The tree structure is dynamically built based on resource labels and relationships.
- Orphaned ClusterMachines (not in a MachineSet) are shown at the cluster level for visibility.
- All MachineSets are sorted alphabetically by name for easier navigation.
- Link nodes provide quick access to related resources without navigating the full hierarchy.
