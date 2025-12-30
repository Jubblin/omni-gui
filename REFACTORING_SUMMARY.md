# Main Package Refactoring Summary

## Overview

The main.go file (originally 2685 lines) has been refactored into logical modules to improve maintainability, reduce cognitive load, and enable easier reuse. The refactoring is ongoing, with core functionality extracted into focused files.

## Current File Structure

### Core Files

- **types.go** - All type definitions

  - `TreeNode`, `AppState`, `DetailComponents`, `ResourceContext`, `SettingsFormFields`
  - `ResourceQuery` struct

- **constants.go** - All constants and version variable

  - Label keys (`labelKeyMachineSet`, etc.)
  - Resource type folders (`resourceTypeFolder`, `clustersFolder`, etc.)
  - Cluster machine label prefix
  - Version variable

### Tree Management

- **tree.go** - Tree node creation, management, navigation, and widget creation

  - Tree node storage (`nodeMap`, `rootNode`, `linkActionMap`)
  - Tree widget creation and callbacks
  - Node selection and branch expansion handling
  - Icon management
  - Functions: `buildInitialTree`, `getChildIDs`, `isBranchNode`, `createTreeNodeWidget`, `setIconForNode`, `updateTreeNodeWidget`, `createResourceTree`, `getNodeByID`, `getRootNode`, `setRootNode`, `addNodeToMap`, `canHaveChildren`, `getIconForResourceType`, `handleNodeSelection`, `handleBranchOpened`, `setupTreeSelection`, `filterEmptyLabelNodes`, `refreshTreeData`

### Resource Management

- **resource_enrichment.go** - Enriching resources with additional data

  - Machine status enrichment
  - ClusterMachine status enrichment
  - Functions: `buildMachineStatusInfo`, `enrichMachineResource`, `enrichClusterMachineResource`

- **resource_labels.go** - Label formatting for different resource types

  - Format functions for each resource type
  - Hostname extraction
  - Resource node creation
  - Functions: `extractHostnameFromResourceMap`, `formatClusterLabel`, `formatMachineLabel`, `formatClusterMachineLabel`, `formatMachineSetLabel`, `formatResourceLabel`, `createResourceNodeFromMap`, `extractResourceInfoFromMap`

### UI Components

- **ui_components.go** - UI component creation

  - Detail pane components
  - Menu creation
  - Language selector
  - Auth field creation
  - Functions: `createDetailComponents`, `setupAppState`, `createBurgerMenu`, `createLanguageSelector`, `createEntryWithFallback`, `createAuthFields`

- **ui_layout.go** - Layout creation

  - Main layout assembly
  - Detail pane layout
  - Functions: `createMainLayout`

- **ui_settings.go** - Settings window

  - Settings form creation
  - Settings save/load
  - Application restart
  - Functions: `createSaveButton`, `showSettingsPage`, `showRestartDialog`, `restartApplication`

- **ui_detail.go** - Detail pane management

  - Detail pane updates
  - Link container management
  - Resource action buttons
  - Functions: `updateDetailPane`, `extractResourceInfo`, `updateDetailTitle`, `loadMachineStatusIfNeeded`, `updateDetailJSON`, `updateDetailLinks`, `clearAllLinkContainers`, `updateResourceActions`, `addActionButton`, `updateMachineIDLinks`, `findMachineIDs`, `updateKubernetesVersionLinks`, `findKubernetesVersions`, `findMachineSetIDs`, `updateMachineSetLinks`

### Main Application

- **main.go** - Main entry point, initialization, and remaining functionality

  - Application startup (`main` function)
  - Window creation
  - Initial setup (`initializeApp`)
  - Resource loading functions (to be extracted to `resource_loader.go`)
  - Action handlers (to be extracted to `actions.go`)
  - Link node creation and management
  - Resource query and search functions

## Refactoring Status

### Completed

✅ Types and constants extracted  
✅ Tree management extracted  
✅ Resource enrichment extracted  
✅ Label formatting extracted  
✅ UI components extracted  
✅ UI layout extracted  
✅ UI settings extracted  
✅ UI detail pane extracted  

### Remaining Work

⏳ Resource loading functions (still in `main.go`)

- `loadNodeChildren`, `loadClusterChildren`, `loadMachineChildren`, `loadMachineSetChildren`, `loadClusterMachineChildren`, `loadKubernetesVersionChildren`
- `loadResourceTypeFolderChildren`, `convertResourcesToNodes`
- `loadClusterMachineForMachine`, `loadMachineStatusNode`
- `findResourcesByK8sVersion`, `findClustersByK8sVersion`, `findClusterMachinesByK8sVersion`, `findClusterMachinesByMachineSet`
- `loadMachineByID`, `loadMachinesByMachineSet`, `loadResourcesByK8sVersion`

⏳ Action handlers (still in `main.go`)

- Cluster actions: `loadClusterStatus`, `loadClusterMetrics`, `loadClusterBootstrap`, `loadClusterKubeconfig`, etc.
- Machine actions: `loadMachineLabels`, `loadMachineExtensions`, `loadMachineUpgradeStatus`, etc.
- MachineSet actions: `loadMachineSetStatus`, `loadMachineSetDestroyStatus`
- ClusterMachine actions: `loadClusterMachineStatus`, `loadClusterMachineConfigStatus`, etc.

⏳ Link node management (still in `main.go`)

- `createLinkNode`, `addLinkNodes`, `addMachineIDLinks`, `addKubernetesVersionLinks`, `addMachineSetLinks`, `addResourceActionLinks`
- `createActionLinkNodes`, `createMachineActionLinkNodes`, `createMachineSetActionLinkNodes`, `createClusterMachineActionLinkNodes`
- `isResourceAlreadyChild`, `buildResourceNodes`

## Benefits

1. **Reduced Cognitive Load**: Each file has a single, clear responsibility
2. **Easier Navigation**: Developers can quickly find relevant code
3. **Better Reusability**: Functions are organized by domain
4. **Improved Maintainability**: Changes are localized to specific files
5. **Easier Testing**: Related functions are grouped together

## Migration Notes

All files remain in the `main` package, so no import changes are needed. The refactoring is purely organizational and maintains backward compatibility.

## File Size Reduction

- **Original**: main.go ~2685 lines
- **Current**: main.go ~1407 lines (47% reduction)
- **Extracted**: ~1278 lines across 9 focused files
