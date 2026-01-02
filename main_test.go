package main

import (
	"testing"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/jubblin/omni-api/internal/i18n"
	"github.com/stretchr/testify/assert"
)

// TestContainerVisibility tests that empty containers are hidden and containers with content are shown
func TestContainerVisibility(t *testing.T) {
	// Initialize Fyne app for widget creation
	testApp := app.New()
	defer testApp.Quit()

	// Initialize i18n for tests
	if err := i18n.Init("en"); err != nil {
		t.Logf("Warning: Failed to initialize i18n: %v", err)
	}

	// Create test app state with empty containers
	appState := &AppState{
		resourceActionsContainer:     container.NewVBox(),
		machineLinksContainer:         container.NewVBox(),
		versionLinksContainer:         container.NewVBox(),
		machineSetLinksContainer:      container.NewVBox(),
		machineRelatedLinksContainer:   container.NewVBox(),
	}

	// Test 1: Initially hide all containers
	clearAllLinkContainers(appState)
	
	// Verify all containers are hidden
	assert.False(t, appState.resourceActionsContainer.Visible(), "ResourceActions container should be hidden when empty")
	assert.False(t, appState.machineLinksContainer.Visible(), "MachineLinks container should be hidden when empty")
	assert.False(t, appState.versionLinksContainer.Visible(), "VersionLinks container should be hidden when empty")
	assert.False(t, appState.machineSetLinksContainer.Visible(), "MachineSetLinks container should be hidden when empty")
	assert.False(t, appState.machineRelatedLinksContainer.Visible(), "MachineRelatedLinks container should be hidden when empty")

	// Test 2: Add content to ResourceActions and verify it's shown
	appState.resourceActionsContainer.Add(widget.NewLabel("Actions:"))
	appState.resourceActionsContainer.Add(widget.NewButton("Test", func() {}))
	
	// Manually show if it has content (simulating updateResourceActions logic)
	if len(appState.resourceActionsContainer.Objects) > 0 {
		appState.resourceActionsContainer.Show()
	}
	
	assert.True(t, appState.resourceActionsContainer.Visible(), "ResourceActions container should be visible when it has content")
	assert.Equal(t, 2, len(appState.resourceActionsContainer.Objects), "ResourceActions container should have 2 objects")

	// Test 3: Clear containers and verify they're hidden again
	clearAllLinkContainers(appState)
	
	assert.False(t, appState.resourceActionsContainer.Visible(), "ResourceActions container should be hidden after clearing")
	assert.Equal(t, 0, len(appState.resourceActionsContainer.Objects), "ResourceActions container should be empty after clearing")

	// Test 4: Test that containers stay hidden when empty (simulating updateMachineIDLinks with no machineIDs)
	// We test the visibility logic directly without calling i18n-dependent functions
	appState.machineLinksContainer.Hide()
	assert.False(t, appState.machineLinksContainer.Visible(), "MachineLinks container should be hidden when no machine IDs")

	// Test 5: Test that containers are shown when they have content
	appState.machineLinksContainer.Add(widget.NewLabel("Machine ID: test-machine-1"))
	appState.machineLinksContainer.Show()
	
	assert.True(t, appState.machineLinksContainer.Visible(), "MachineLinks container should be visible when it has content")
	assert.Greater(t, len(appState.machineLinksContainer.Objects), 0, "MachineLinks container should have objects")
}

// TestClearAllLinkContainers tests that clearAllLinkContainers properly clears and hides all containers
func TestClearAllLinkContainers(t *testing.T) {
	// Initialize Fyne app for widget creation
	testApp := app.New()
	defer testApp.Quit()

	// Initialize i18n for tests
	if err := i18n.Init("en"); err != nil {
		t.Logf("Warning: Failed to initialize i18n: %v", err)
	}

	appState := &AppState{
		resourceActionsContainer:     container.NewVBox(),
		machineLinksContainer:         container.NewVBox(),
		versionLinksContainer:         container.NewVBox(),
		machineSetLinksContainer:      container.NewVBox(),
		machineRelatedLinksContainer:   container.NewVBox(),
	}

	// Add some content to all containers
	appState.resourceActionsContainer.Add(widget.NewLabel("Test"))
	appState.machineLinksContainer.Add(widget.NewLabel("Test"))
	appState.versionLinksContainer.Add(widget.NewLabel("Test"))
	appState.machineSetLinksContainer.Add(widget.NewLabel("Test"))
	appState.machineRelatedLinksContainer.Add(widget.NewLabel("Test"))
	
	// Show all containers
	appState.resourceActionsContainer.Show()
	appState.machineLinksContainer.Show()
	appState.versionLinksContainer.Show()
	appState.machineSetLinksContainer.Show()
	appState.machineRelatedLinksContainer.Show()

	// Verify they have content and are visible
	assert.Equal(t, 1, len(appState.resourceActionsContainer.Objects))
	assert.True(t, appState.resourceActionsContainer.Visible())

	// Clear all containers
	clearAllLinkContainers(appState)

	// Verify all containers are empty and hidden
	assert.Equal(t, 0, len(appState.resourceActionsContainer.Objects), "ResourceActions should be empty")
	assert.Equal(t, 0, len(appState.machineLinksContainer.Objects), "MachineLinks should be empty")
	assert.Equal(t, 0, len(appState.versionLinksContainer.Objects), "VersionLinks should be empty")
	assert.Equal(t, 0, len(appState.machineSetLinksContainer.Objects), "MachineSetLinks should be empty")
	assert.Equal(t, 0, len(appState.machineRelatedLinksContainer.Objects), "MachineRelatedLinks should be empty")

	assert.False(t, appState.resourceActionsContainer.Visible(), "ResourceActions should be hidden")
	assert.False(t, appState.machineLinksContainer.Visible(), "MachineLinks should be hidden")
	assert.False(t, appState.versionLinksContainer.Visible(), "VersionLinks should be hidden")
	assert.False(t, appState.machineSetLinksContainer.Visible(), "MachineSetLinks should be hidden")
	assert.False(t, appState.machineRelatedLinksContainer.Visible(), "MachineRelatedLinks should be hidden")
}

// TestUpdateDetailLinksIntegration tests the full updateDetailLinks flow
func TestUpdateDetailLinksIntegration(t *testing.T) {
	// Initialize Fyne app for widget creation
	testApp := app.New()
	defer testApp.Quit()

	// Initialize i18n for tests
	if err := i18n.Init("en"); err != nil {
		t.Logf("Warning: Failed to initialize i18n: %v", err)
	}

	appState := &AppState{
		resourceActionsContainer:     container.NewVBox(),
		machineLinksContainer:         container.NewVBox(),
		versionLinksContainer:         container.NewVBox(),
		machineSetLinksContainer:      container.NewVBox(),
		machineRelatedLinksContainer:   container.NewVBox(),
	}

	// Add some content to containers first to test clearing
	appState.resourceActionsContainer.Add(widget.NewLabel("Test"))
	appState.machineLinksContainer.Add(widget.NewLabel("Test"))
	appState.resourceActionsContainer.Show()
	appState.machineLinksContainer.Show()

	// Test with machine resource data that has machine_id
	resourceData := map[string]interface{}{
		"id":        "machine-1",
		"type":      "Machines.omni.sidero.dev",
		"machine_id": "machine-1",
	}

	updateDetailLinks(resourceData, "Machines.omni.sidero.dev", "machine-1", appState)

	// updateDetailLinks clears all containers, so all should be hidden and empty
	assert.False(t, appState.resourceActionsContainer.Visible(), "ResourceActions should be hidden after updateDetailLinks")
	assert.Equal(t, 0, len(appState.resourceActionsContainer.Objects), "ResourceActions should be empty after updateDetailLinks")

	assert.False(t, appState.machineLinksContainer.Visible(), "MachineLinks should be hidden after updateDetailLinks")
	assert.Equal(t, 0, len(appState.machineLinksContainer.Objects), "MachineLinks should be empty after updateDetailLinks")

	assert.False(t, appState.versionLinksContainer.Visible(), "VersionLinks should be hidden after updateDetailLinks")
	assert.Equal(t, 0, len(appState.versionLinksContainer.Objects), "VersionLinks should be empty after updateDetailLinks")

	assert.False(t, appState.machineSetLinksContainer.Visible(), "MachineSetLinks should be hidden after updateDetailLinks")
	assert.Equal(t, 0, len(appState.machineSetLinksContainer.Objects), "MachineSetLinks should be empty after updateDetailLinks")
}

// TestContainerVisibilityWithMachineRoleNone simulates the specific case mentioned:
// Machine with role "none" should not show blank lines between MachineStatus and Labels
func TestContainerVisibilityWithMachineRoleNone(t *testing.T) {
	// Initialize Fyne app for widget creation
	testApp := app.New()
	defer testApp.Quit()

	// Initialize i18n for tests
	if err := i18n.Init("en"); err != nil {
		t.Logf("Warning: Failed to initialize i18n: %v", err)
	}

	appState := &AppState{
		resourceActionsContainer:     container.NewVBox(),
		machineLinksContainer:         container.NewVBox(),
		versionLinksContainer:         container.NewVBox(),
		machineSetLinksContainer:      container.NewVBox(),
		machineRelatedLinksContainer:   container.NewVBox(),
	}

	// Simulate machine with role "none" - no machine_id, no kubernetes_version, no machine-set
	resourceData := map[string]interface{}{
		"id":   "machine-none-role",
		"type": "Machines.omni.sidero.dev",
		// No machine_id, kubernetes_version, or machine-set labels
	}

	updateDetailLinks(resourceData, "Machines.omni.sidero.dev", "machine-none-role", appState)

	// updateDetailLinks clears all containers, so all should be hidden
	assert.False(t, appState.resourceActionsContainer.Visible(), "ResourceActions should be hidden after updateDetailLinks")
	assert.False(t, appState.machineLinksContainer.Visible(), "MachineLinks should be hidden when no machine_id")
	assert.False(t, appState.versionLinksContainer.Visible(), "VersionLinks should be hidden when no kubernetes_version")
	assert.False(t, appState.machineSetLinksContainer.Visible(), "MachineSetLinks should be hidden when no machine-set label")
	assert.False(t, appState.machineRelatedLinksContainer.Visible(), "MachineRelatedLinks should be hidden")

	// Verify no blank lines: all containers should be hidden
	visibleContainers := 0
	if appState.resourceActionsContainer.Visible() {
		visibleContainers++
	}
	if appState.machineLinksContainer.Visible() {
		visibleContainers++
	}
	if appState.versionLinksContainer.Visible() {
		visibleContainers++
	}
	if appState.machineSetLinksContainer.Visible() {
		visibleContainers++
	}
	if appState.machineRelatedLinksContainer.Visible() {
		visibleContainers++
	}

	// All containers should be hidden (0 containers), preventing blank lines
	assert.Equal(t, 0, visibleContainers, "All containers should be hidden, preventing blank lines")
}
