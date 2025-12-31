package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/jubblin/omni-api/internal/i18n"
)

// createMainLayout creates the main application layout
func createMainLayout(burgerMenu *widget.Button, refreshButton *widget.Button, settingsButton *widget.Button, resourceTree *widget.Tree, detailComponents *DetailComponents, statusLabel *widget.Label) *container.Split {
	// Top bar with three buttons: burger menu, refresh, settings (all in top-left corner)
	topBar := container.NewHBox(burgerMenu, refreshButton, settingsButton)
	// Format directive is in translation file: "app.connected" = "Connected to: %s"
	infoLabel := widget.NewLabel(i18n.T("app.connected", os.Getenv("OMNI_ENDPOINT"))) //nolint
	infoLabel.Wrapping = fyne.TextWrapWord

	// Wrap tree in scroll container to ensure it's visible
	// Add a minimum content to ensure tree is visible even when empty
	treeContainer := container.NewBorder(nil, nil, nil, nil, resourceTree)
	treeScroll := container.NewScroll(treeContainer)
	treeScroll.SetMinSize(fyne.NewSize(300, 400))
	leftPane := container.NewBorder(topBar, infoLabel, nil, nil, treeScroll)
	
	// Order matches documentation GUI Layout Overview:
	// Top Section: Title
	// Main Content (Scrollable): JSON Text (expands to fill available space), containers at bottom
	// Bottom Section: Status Label (added via Border container below)
	
	// Containers positioned at bottom (build from bottom of screen)
	// Only show containers when they have content (hide empty containers to avoid blank lines)
	containersBox := container.NewVBox(
		detailComponents.ResourceActions, // Containers: Resource Actions Container
		detailComponents.MachineLinks,    // Containers: Machine Links Container
		detailComponents.VersionLinks,   // Containers: Version Links Container
		detailComponents.MachineSetLinks, // Containers: MachineSet Links Container
		detailComponents.MachineRelatedLinks, // Containers: Machine Related Links Container
	)
	// Initially hide all containers (they'll be shown when populated)
	detailComponents.ResourceActions.Hide()
	detailComponents.MachineLinks.Hide()
	detailComponents.VersionLinks.Hide()
	detailComponents.MachineSetLinks.Hide()
	detailComponents.MachineRelatedLinks.Hide()
	
	// JSON text expands to fill available space, containers at bottom
	// Use Border layout: top=Title, center=Text (expands), bottom=containers
	detailContent := container.NewBorder(
		detailComponents.Title, // Top: Resource Details Title
		containersBox,          // Bottom: Containers (build from bottom)
		nil, nil,
		detailComponents.Text,   // Center: Resource JSON Text - expands to fill available space
	)
	
	detailScroll := container.NewScroll(detailContent)
	detailScroll.SetMinSize(fyne.NewSize(400, 0))

	rightPane := container.NewBorder(nil, statusLabel, nil, nil, detailScroll)
	
	split := container.NewHSplit(leftPane, rightPane)
	split.SetOffset(0.3)

	return split
}
