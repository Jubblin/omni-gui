# Tree Hierarchy

This document describes the hierarchical tree structure used to display Omni resources in the Fyne GUI application.

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
Machines (resource-type-folder) [when expanded, loads list of machines directly]
└── Machine 1 [when expanded, loads the machine details]
    ├── MachineStatus (same ID as Machine)
    ├── [Link nodes: Labels, Extensions, Upgrade Status, Metrics, Config Diff]
    └── (ClusterMachine) ClusterMachine (reverse lookup - ClusterMachine ID = Machine ID, prefix label with "(ClusterMachine)", hidden if not found)
        ├── Machine (back reference)
        ├── Cluster (from labels)
        ├── MachineSet (from labels, if applicable)
        └── [Link nodes: Status, Config Status, Talos Version, Config]
```

### Machine Children Details

- **ClusterMachine**: Reverse lookup - since ClusterMachine ID equals Machine ID, the associated ClusterMachine is shown as a child with the label prefixed with "(ClusterMachine)" for clarity. The ClusterMachine node is only shown if it exists (hidden if not found). When expanded, it shows:
  - **Machine**: Back reference to the Machine itself
  - **Cluster**: The cluster this ClusterMachine belongs to (from `omni.sidero.dev/cluster` label)
  - **MachineSet**: The MachineSet this ClusterMachine belongs to (from `omni.sidero.dev/machine-set` label, if applicable)
  - **Link Nodes**: ClusterMachine-specific action links (Status, Config Status, Talos Version, Config)
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

## Resource Relationships

The hierarchy reflects the actual API relationships:

- Clusters contain MachineSets (via `omni.sidero.dev/cluster` label)
- MachineSets contain ClusterMachines (via `omni.sidero.dev/machine-set` label)
- ClusterMachines reference Machines (via `machine_id` field)
- Machines can reference ClusterMachines (reverse lookup)

## Notes

- The tree structure is dynamically built based on resource labels and relationships.
- Orphaned ClusterMachines (not in a MachineSet) are shown at the cluster level for visibility.
- All MachineSets are sorted alphabetically by name for easier navigation.
- Link nodes provide quick access to related resources without navigating the full hierarchy.
