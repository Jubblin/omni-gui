# Fyne GUI Tree Hierarchy

This document describes the tree structure hierarchy implemented in the Fyne GUI application for browsing Omni resources.

## GUI Layout Overview

The Fyne GUI uses a horizontal split layout dividing the window into two main panes:

### Left Pane (30% width by default)

**Top Section:**

- **Burger Menu Button** (☰ Menu) - Located in the top-left corner
  - Provides access to application menu options
  - Options include: "Refresh" (rebuilds tree structure) and "Settings" (opens settings window)
  - Opens a popup menu when clicked

**Main Content:**

- **Resource Tree** - Scrollable hierarchical tree view
  - Displays hierarchical tree of Omni resources
  - Three top-level resource type folders:
    1. **Clusters** - Contains clusters and all related resources
    2. **Machines** - Contains machines and all related resources  
    3. **MachineSets** - Contains machine sets and all related resources
  - Uses lazy loading - children are loaded when nodes are expanded
  - Custom icons for different resource types (folders, computers, storage, documents, etc.)
  - Minimum size: 300x400 pixels to ensure visibility

**Bottom Section:**

- **Connection Info Label** - Located at the bottom
  - Displays "Connected to: [OMNI_ENDPOINT]"
  - Text wrapping enabled for long endpoint URLs

### Right Pane (70% width by default)

**Top Section:**

- **Resource Details Title** - Bold label
  - Shows "Resource Details: [resource_id]" when a resource is selected
  - Shows "Resource Details" when no resource is selected
  - Updates dynamically when tree selection changes

**Main Content (Scrollable):**

- **Resource JSON Text** - Multi-line, selectable, copyable text field
  - Displays formatted JSON (with indentation) of the selected resource
  - Text wrapping enabled for long content
  - **Text can be selected and copied** using standard keyboard shortcuts (Cmd+C on Mac, Ctrl+C on Windows/Linux)
  - Editable (though edits will be overwritten on next update)
  - Uses all available space in pane not used by other artifacts

#### Containers (build from bottom of the screen)

- **Resource Actions Container** - Action buttons/links (if applicable)
- **Machine Links Container** - Machine-related links (displayed when viewing machine resources)
- **Version Links Container** - Version-related links (displayed when applicable)
- **MachineSet Links Container** - Machine set-related links (displayed when applicable)
- **Machine Related Links Container** - Additional machine-related links (displayed when applicable)

**Bottom Section:**

- **Status Label** - Located at the bottom
  - Shows application status messages (e.g., "Ready")
  - Updates during operations to provide user feedback

### Layout Features

1. **Resizable Split**: The horizontal splitter can be dragged to adjust the width of left/right panes
2. **Scrollable Content**: Both panes have independent scrolling when content exceeds visible area
3. **Responsive Sizing**: Minimum sizes ensure components remain usable at different window sizes
4. **Dynamic Content**: Right pane content updates based on tree selection
5. **Text Selection**: Right pane JSON text is fully selectable and copyable

### Window Properties

- **Default Size**: 1400x900 pixels
- **Centered**: Window automatically centers on screen when opened
- **Theme Support**: Uses Fyne's built-in theme system (supports light/dark mode)

The layout provides a clean separation between resource navigation (left) and resource details (right), with easy access to settings via the burger menu.

### Settings Window

The Settings window (accessed via ☰ Menu → Settings) contains:

**Application Settings tab:**

- **Language Selector** - Dropdown to choose UI language
  - Automatically refreshes UI text when language is changed
  - Changes take effect immediately without requiring application restart

**Authentication Settings tab:**

- **Omni Endpoint** - Configuration for the Omni API endpoint
- **Authentication Method** - Choose between "Service Account" or "OIDC"
- **Service Account Key** - Base64 encoded service account key (when using Service Account)
- **OIDC Settings** - Issuer URL, Client ID, and Client Secret (when using OIDC)

Settings can be saved and the application will prompt to restart to apply authentication changes.

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
    ├── (ClusterMachine) ClusterMachine (reverse lookup - ClusterMachine ID = Machine ID, prefix label with "(ClusterMachine)", hidden if not found)
    │   ├── Machine (back reference)
    │   ├── Cluster (from labels)
    │   ├── MachineSet (from labels, if applicable)
    │   └── [Link nodes: Status, Config Status, Talos Version, Config]
    ├── MachineStatus (same ID as Machine)
    └── [Link nodes: Labels, Extensions, Upgrade Status, Metrics, Config Diff]
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
