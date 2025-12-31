# Usage Guide

This document provides a guide on how to use the Fyne GUI application for browsing and managing Omni resources.

## Basic Navigation

1. **Expand Clusters**: Click on "Clusters" to see all clusters, then expand a cluster to see its MachineSets and ClusterMachines.

2. **Expand Machines**: Click on "Machines" to see all machines, then expand a machine to see its ClusterMachine and MachineStatus.

3. **Expand MachineSets**: Click on "MachineSets" to see all MachineSets, then expand a MachineSet to see its ClusterMachines.

4. **Click Link Nodes**: Click on link nodes (like "Status", "Metrics") to load related resources in the detail pane.

5. **Navigate Relationships**: Use the tree to navigate between related resources (e.g., from a ClusterMachine to its Machine or Cluster).

6. **Refresh Tree Data**: Click the refresh button (circular arrow icon) in the top-left corner, or use the burger menu (☰ Menu) and select "Refresh" to rebuild the internal tree structure. This resets the tree to its initial state, clearing all loaded children and starting fresh.

## Viewing Resource Details

- **Select a Resource**: Click on any resource node in the tree to view its details in the right pane.
- **View JSON**: The selected resource's JSON representation is displayed in the detail pane with proper formatting.
- **Copy JSON**: Select and copy the JSON text using standard keyboard shortcuts (Cmd+C on Mac, Ctrl+C on Windows/Linux).

## Using Settings

1. **Open Settings**: Click the settings button (cog icon) in the top-left corner, or use the burger menu (☰ Menu) and select "Settings".

2. **Application Settings**:
   - Change the UI language (takes effect immediately)
   - Enable/disable empty label filtering (applies immediately, no restart required)
   - Configure ignore environment variables setting (applies immediately, no restart required)
   - Set gRPC debug level (applies immediately, no restart required)

3. **Authentication Settings**:
   - Configure the Omni endpoint
   - Choose authentication method (Service Account or OIDC)
   - Enter credentials for the selected authentication method

4. **Save Settings**: Click "Save" to persist your settings. All application settings (Language, Empty Label Filtering, Ignore Environment Variables, gRPC Debug Level) apply immediately without restart. Only authentication and endpoint changes require a restart.

## Tips

- **Lazy Loading**: Resources are loaded on-demand when you expand nodes, so initial load time is fast.
- **Empty Label Filtering**: By default, nodes with empty labels are hidden. Enable "Show nodes with empty labels in the tree" in Settings (Application Settings tab) to see all nodes. This setting applies immediately without restart.
- **Ignore Environment Variables**: Enable "Ignore environment variables and use only settings file" in Settings (Application Settings tab) to use only settings file values and ignore all environment variables. This setting applies immediately without restart.
- **Logging**: All application logs are written as valid JSON to `omni-api.log` in the execution directory. Each execution starts with a startup banner containing version and configuration information.
- **Orphaned Resources**: ClusterMachines not in a MachineSet are shown directly under their Cluster for easy identification.
- **Link Nodes**: Use link nodes to quickly access related resources without navigating the full hierarchy.
- **Refresh**: Use the Refresh option from the burger menu to reset the tree view after changes in your Omni instance.
