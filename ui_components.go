package main

import (
	"os"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	omniclient "github.com/jubblin/omni-api/internal/client"
	"github.com/jubblin/omni-api/internal/i18n"
)

// createDetailComponents creates UI components for the detail pane
func createDetailComponents() *DetailComponents {
	title := widget.NewLabel(i18n.T("tree.select_resource"))
	title.TextStyle = fyne.TextStyle{Bold: true}

	// Use multi-line Entry for copyable text
	// Entry allows text selection and copying (edits will be overwritten on next update)
	text := widget.NewMultiLineEntry()
	text.Wrapping = fyne.TextWrapWord

	return &DetailComponents{
		Title:               title,
		Text:                text,
		MachineLinks:        container.NewVBox(),
		VersionLinks:        container.NewVBox(),
		MachineSetLinks:     container.NewVBox(),
		MachineRelatedLinks: container.NewVBox(),
		ResourceActions:     container.NewVBox(),
	}
}

// setupAppState sets up the AppState with UI components
func setupAppState(appState *AppState, resourceTree *widget.Tree, resourceIDInput *widget.Entry, statusLabel *widget.Label, detailComponents *DetailComponents) {
	appState.detailText = detailComponents.Text
	appState.detailTitle = detailComponents.Title
	appState.resourceTree = resourceTree
	if resourceIDInput != nil {
		appState.resourceIDInput = resourceIDInput
	}
	appState.statusLabel = statusLabel
	appState.machineLinksContainer = detailComponents.MachineLinks
	appState.versionLinksContainer = detailComponents.VersionLinks
	appState.machineSetLinksContainer = detailComponents.MachineSetLinks
	appState.machineRelatedLinksContainer = detailComponents.MachineRelatedLinks
	appState.resourceActionsContainer = detailComponents.ResourceActions
}

// createBurgerMenu creates the hamburger menu button
func createBurgerMenu(myWindow fyne.Window, appState *AppState) *widget.Button {
	burgerMenu := widget.NewButton("☰ Menu", nil)
	
	menuItems := []*fyne.MenuItem{
		fyne.NewMenuItem("Refresh", func() {
			refreshTreeData(appState)
		}),
		fyne.NewMenuItem("Settings", func() {
			showSettingsPage(myWindow)
		}),
	}

	menu := fyne.NewMenu("Menu", menuItems...)
	
	burgerMenu.OnTapped = func() {
		popup := widget.NewPopUpMenu(menu, fyne.CurrentApp().Driver().CanvasForObject(burgerMenu))
		pos := fyne.CurrentApp().Driver().AbsolutePositionForObject(burgerMenu)
		popup.ShowAtPosition(pos.Add(fyne.NewPos(0, burgerMenu.Size().Height)))
	}

	return burgerMenu
}

// createLanguageSelector creates a language selection widget
func createLanguageSelector(appState *AppState, refreshUI func()) *widget.Select {
	languages := i18n.GetSupportedLanguages()
	options := make([]string, 0, len(languages))
	for _, lang := range languages {
		options = append(options, i18n.GetLanguageName(lang))
	}

	langSelect := widget.NewSelect(options, func(selected string) {
		for _, lang := range languages {
			if i18n.GetLanguageName(lang) == selected {
				if err := i18n.SetLanguage(lang); err == nil {
					refreshUI()
				}
				break
			}
		}
	})

	currentLang := i18n.GetCurrentLanguage()
	langSelect.SetSelected(i18n.GetLanguageName(currentLang))

	return langSelect
}

// createEntryWithFallback creates an entry widget with value from settings or environment
func createEntryWithFallback(settingsValue string, envVars ...string) *widget.Entry {
	entry := widget.NewEntry()
	value := settingsValue
	if value == "" {
		for _, envVar := range envVars {
			if v := os.Getenv(envVar); v != "" {
				value = v
				break
			}
		}
	}
	if value != "" {
		entry.SetText(value)
	}
	return entry
}

// createAuthFields creates authentication-related input fields
func createAuthFields(settings *omniclient.Settings) (*widget.Entry, *widget.Entry, *widget.Entry, *widget.Entry, *widget.Select, *fyne.Container, *fyne.Container) {
	// Service Account fields
	serviceAccountLabel := widget.NewLabel("Service Account Key:")
	serviceAccountEntry := createEntryWithFallback(
		settings.ServiceAccount,
		"OMNI_SERVICE_ACCOUNT",
		"OMNI_SERVICE_ACCOUNT_KEY",
	)
	serviceAccountEntry.SetPlaceHolder("Base64 encoded service account key")
	serviceAccountEntry.Password = true

	// OIDC fields
	oidcIssuerLabel := widget.NewLabel("OIDC Issuer URL:")
	oidcIssuerEntry := createEntryWithFallback(settings.OIDCIssuerURL, "OMNI_OIDC_ISSUER_URL")
	oidcIssuerEntry.SetPlaceHolder("https://oidc-provider.com (optional, auto-derived from endpoint)")

	oidcClientIDLabel := widget.NewLabel("OIDC Client ID:")
	oidcClientIDEntry := createEntryWithFallback(settings.OIDCClientID, "OMNI_OIDC_CLIENT_ID")
	oidcClientIDEntry.SetPlaceHolder("your-client-id")

	oidcClientSecretLabel := widget.NewLabel("OIDC Client Secret:")
	oidcClientSecretEntry := createEntryWithFallback(settings.OIDCClientSecret, "OMNI_OIDC_CLIENT_SECRET")
	oidcClientSecretEntry.SetPlaceHolder("your-client-secret")
	oidcClientSecretEntry.Password = true

	// Containers
	serviceAccountContainer := container.NewVBox(serviceAccountLabel, serviceAccountEntry)
	oidcContainer := container.NewVBox(
		oidcIssuerLabel, oidcIssuerEntry,
		oidcClientIDLabel, oidcClientIDEntry,
		oidcClientSecretLabel, oidcClientSecretEntry,
	)

	// Auth method selection
	authSelect := widget.NewSelect([]string{"Service Account", "OIDC"}, nil)
	currentAuth := omniclient.GetCurrentAuthMethod(false)
	if currentAuth == omniclient.AuthMethodOIDC {
		authSelect.SetSelected("OIDC")
	} else {
		authSelect.SetSelected("Service Account")
	}

	// Show/hide fields based on auth method
	updateAuthFields := func(selected string) {
		if selected == "OIDC" {
			serviceAccountContainer.Hide()
			oidcContainer.Show()
		} else {
			serviceAccountContainer.Show()
			oidcContainer.Hide()
		}
	}
	authSelect.OnChanged = updateAuthFields
	updateAuthFields(authSelect.Selected)

	return serviceAccountEntry, oidcIssuerEntry, oidcClientIDEntry, oidcClientSecretEntry, authSelect, serviceAccountContainer, oidcContainer
}
