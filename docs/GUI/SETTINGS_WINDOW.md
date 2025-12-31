# Settings Window

This document describes the Settings window interface and configuration options.

## Overview

The Settings window (accessed via ☰ Menu → Settings) provides a tabbed interface for configuring application and authentication settings.

## Window Properties

- **Title**: "Settings"
- **Size**: 500x500 pixels
- **Position**: Centered on screen
- **Layout**: Tabbed interface with buttons at the bottom

## Overall Structure

The window uses a vertical layout with three sections:

1. **Top Section**: Tabbed content (scrollable)
2. **Middle Section**: Horizontal separator
3. **Bottom Section**: Action buttons (Cancel and Save)

## Application Settings Tab

Contains a scrollable vertical container with the following components (utilising at least 75% of usable space):

### Language Selector

- Label: "Language:"
- Dropdown selector to choose UI language
- Automatically refreshes UI text when language is changed
- Changes take effect immediately without requiring application restart

### Separator 1

Horizontal line separating sections

### Empty Label Filtering

- Label: "Empty Label Filtering:"
- Description text (word-wrapped):
  - "By default, nodes with empty labels are filtered out to prevent blank entries in the tree. Enable this option to disable filtering and show all nodes, including those with empty labels. This affects tree rendering at all levels: child node loading, tree widget rendering, and node filtering."
- Checkbox: "Show nodes with empty labels in the tree (disable filtering)"
- Checked state is loaded from saved settings
- Changes apply immediately without requiring application restart (tree refreshes automatically)

### Separator 2

Horizontal line separating sections

### Ignore Environment Variables

- Label: "Ignore Environment Variables:"
- Description text (word-wrapped):
  - "When enabled, the application will only use settings from the settings file and ignore all environment variables. When disabled, environment variables take precedence over settings file values."
- Checkbox: "Ignore environment variables and use only settings file"
- Checked state is loaded from saved settings
- Changes apply immediately without requiring application restart (affects future settings loading, current client continues to work)

### Separator 3

Horizontal line separating sections

### gRPC Debug Level

- Label: "gRPC Debug Level:"
- Description text (word-wrapped):
  - "Controls the level of gRPC debugging output. Level 0: Disabled. Level 1: Query dumps only (logs query metadata as JSON). Level 2: Query + Response dumps (logs query and response info). Level 3: Query + Response + UI Display dumps (also creates dump files when resources are displayed)."
- Dropdown selector with options:
  - "0 - Disabled"
  - "1 - Query dumps only"
  - "2 - Query + Response dumps"
  - "3 - Full debugging (includes UI dumps)"
- Selected value is loaded from saved settings
- Changes apply immediately without requiring application restart

## Authentication Settings Tab

Contains a scrollable vertical container with the following components (utilising at least 75% of usable space):

### Omni Endpoint

- Label: "Omni Endpoint:"
- Text entry field
- Placeholder: "<https://omni.example.com>"
- Pre-filled from settings file or `OMNI_ENDPOINT` environment variable

### Separator 4

Horizontal line separating sections

### Authentication Method

- Label: "Authentication Method:"
- Dropdown selector with options: "Service Account" or "OIDC"
- Dynamically shows/hides relevant authentication fields based on selection

### Separator 5

Horizontal line separating sections

### Service Account Container

Shown when "Service Account" is selected:

- **Service Account Key:**
  - Label: "Service Account Key:"
  - Password-masked multi-line text entry field, wrapping string rather than scrolling, should use all available space in the container
  - Placeholder: "Base64 encoded service account key"
  - Pre-filled from settings file or `OMNI_SERVICE_ACCOUNT` / `OMNI_SERVICE_ACCOUNT_KEY` environment variables

### OIDC Container

Shown when "OIDC" is selected:

- **OIDC Issuer URL:**
  - Label: "OIDC Issuer URL:"
  - Text entry field
  - Placeholder: "<https://oidc-provider.com> (optional, auto-derived from endpoint)"
  - Pre-filled from settings file or `OMNI_OIDC_ISSUER_URL` environment variable

- **OIDC Client ID:**
  - Label: "OIDC Client ID:"
  - Text entry field
  - Placeholder: "your-client-id"
  - Pre-filled from settings file or `OMNI_OIDC_CLIENT_ID` environment variable

- **OIDC Client Secret:**
  - Label: "OIDC Client Secret:"
  - Password-masked multi-line text entry field, wrapping string rather than scrolling, should use all available space in the container
  - Placeholder: "your-client-secret"
  - Pre-filled from settings file or `OMNI_OIDC_CLIENT_SECRET` environment variable

## Bottom Section

- **Cancel Button**: Closes the settings window without saving changes
- **Save Button**:
  - Saves all settings to the settings file (`~/.omni-api/settings.json`)
  - Applies application settings dynamically (all settings in Application Settings tab apply immediately)
  - Shows a success message if no restart is required
  - Shows a restart dialog only if authentication or endpoint settings changed
  - Dialog provides two options: "Restart Now" or "Later"

## Behavior Notes

- **Scrollable Content**: Both tabs are scrollable if content exceeds the visible area
- **Dynamic Fields**: Authentication fields show/hide based on the selected authentication method
- **Environment Variable Fallback**: Entry fields pre-fill from environment variables if settings file values are empty (unless "Ignore Environment Variables" is enabled)
- **Password Fields**: Service Account Key and OIDC Client Secret are masked for security
- **Dynamic Application**: Application settings apply immediately without restart:
  - **Language**: Changes apply immediately
  - **Empty Label Filtering**: Changes apply immediately (tree refreshes automatically)
  - **gRPC Debug Level**: Changes apply immediately (debugging output level updates in real-time)
  - **Ignore Environment Variables**: Changes apply immediately (affects future settings loading, current client continues to work)
- **Restart Required**: Only authentication settings require a restart to take effect:
  - **Authentication Method**: Changing between Service Account and OIDC
  - **Omni Endpoint**: Changing the endpoint URL
  - **Authentication Credentials**: Changing service account key or OIDC credentials
- **Settings Persistence**: All settings are saved to `~/.omni-api/settings.json` and persist across application restarts
- **Command-Line Flags**: Command-line flags (e.g., `--show-empty-labels`, `--ignore-env`, `--grpc-debug`) take precedence over settings file values but are not persisted
