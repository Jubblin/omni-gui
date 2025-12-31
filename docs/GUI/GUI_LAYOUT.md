# GUI Layout

This document describes the main window layout and structure of the Fyne GUI application.

## Overview

The Fyne GUI uses a horizontal split layout dividing the window into two main panes:

- **Empty Label Filtering** - Configurable in Settings (Application Settings tab)
  - By default, nodes with empty labels are filtered out to prevent blank entries in the tree
  - Enable "Show nodes with empty labels in the tree" in Settings (Application Settings tab) to disable filtering and show all nodes, including those with empty labels. This setting applies immediately without restart.
  - This affects tree rendering at all levels: child node loading, tree widget rendering, and node filtering

## Left Pane (30% width by default)

### Top Section

- **Burger Menu Button** (☰) - Located in the top-left corner
  - Provides access to application menu options
  - Options include: "Refresh" (rebuilds tree structure) and "Settings" (opens settings window)
  - Opens a popup menu when clicked

- **Refresh Button** (circular arrow icon) - Located in the top-left corner
  - reloads data and rebuilds tree structure)

- **Settings Button** (cog type icon) - Located in the top-left corner
  - opens settings window

### Main Content

- **Resource Tree** - Scrollable hierarchical tree view
  - Displays hierarchical tree of Omni resources
  - Three top-level resource type folders:
    1. **Clusters** - Contains clusters and all related resources
    2. **Machines** - Contains machines and all related resources  
    3. **MachineSets** - Contains machine sets and all related resources
  - Uses lazy loading - children are loaded when nodes are expanded
  - Custom icons for different resource types (folders, computers, storage, documents, etc.)
  - Minimum size: 300x400 pixels to ensure visibility
  - **Empty Label Filtering**: By default, nodes with empty labels are filtered out to prevent blank entries. Enable "Show nodes with empty labels in the tree" in Settings (Application Settings tab) to disable filtering and show all nodes. This setting applies immediately without restart.

### Bottom Section

- **Connection Info Label** - Located at the bottom
  - Displays "Connected to: [OMNI_ENDPOINT]"
  - Text wrapping enabled for long endpoint URLs

## Right Pane (70% width by default)

### Top Section

- **Resource Details Title** - Bold label
  - Shows "Resource Details: [resource_id]" when a resource is selected
  - Shows "Resource Details" when no resource is selected
  - Updates dynamically when tree selection changes

### Main Content (Scrollable)

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

### Bottom Section

- **Status Label** - Located at the bottom
  - Shows application status messages (e.g., "Ready")
  - Updates during operations to provide user feedback

## Layout Features

1. **Resizable Split**: The horizontal splitter can be dragged to adjust the width of left/right panes
2. **Scrollable Content**: Both panes have independent scrolling when content exceeds visible area
3. **Responsive Sizing**: Minimum sizes ensure components remain usable at different window sizes
4. **Dynamic Content**: Right pane content updates based on tree selection
5. **Text Selection**: Right pane JSON text is fully selectable and copyable

## Window Properties

- **Default Size**: 1400x900 pixels
- **Centered**: Window automatically centers on screen when opened
- **Theme Support**: Uses Fyne's built-in theme system (supports light/dark mode)

The layout provides a clean separation between resource navigation (left) and resource details (right), with easy access to settings via the burger menu.
