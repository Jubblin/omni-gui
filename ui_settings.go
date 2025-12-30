package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	omniclient "github.com/jubblin/omni-api/internal/client"
)

// createSaveButton creates the save button with settings save logic
func createSaveButton(settingsWindow fyne.Window, parentWindow fyne.Window, fields *SettingsFormFields) *widget.Button {
	return widget.NewButton("Save", func() {
		newSettings := &omniclient.Settings{Endpoint: fields.EndpointEntry.Text}
		
		if fields.AuthSelect.Selected == "OIDC" {
			newSettings.AuthMethod = "oidc"
			newSettings.OIDCIssuerURL = fields.OIDCIssuerEntry.Text
			newSettings.OIDCClientID = fields.OIDCClientIDEntry.Text
			newSettings.OIDCClientSecret = fields.OIDCClientSecretEntry.Text
			newSettings.ServiceAccount = ""
		} else {
			newSettings.AuthMethod = "service_account"
			newSettings.ServiceAccount = fields.ServiceAccountEntry.Text
			newSettings.OIDCIssuerURL = ""
			newSettings.OIDCClientID = ""
			newSettings.OIDCClientSecret = ""
		}

		if err := omniclient.SaveSettings(newSettings); err != nil {
			slog.Error("Failed to save settings", "error", err)
			errorDialog := widget.NewModalPopUp(
				widget.NewLabel(fmt.Sprintf("Failed to save settings: %v", err)),
				settingsWindow.Canvas(),
			)
			errorDialog.Resize(fyne.NewSize(300, 100))
			errorDialog.Show()
			return
		}

		showRestartDialog(settingsWindow, parentWindow)
	})
}

// showSettingsPage displays the settings window
func showSettingsPage(parentWindow fyne.Window) {
	settings, err := omniclient.LoadSettings(false)
	if err != nil {
		slog.Error("Failed to load settings", "error", err)
		settings = &omniclient.Settings{}
	}

	settingsWindow := fyne.CurrentApp().NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(500, 500))
	settingsWindow.CenterOnScreen()

	// Endpoint input
	endpointLabel := widget.NewLabel("Omni Endpoint:")
	endpointEntry := createEntryWithFallback(settings.Endpoint, "OMNI_ENDPOINT")
	endpointEntry.SetPlaceHolder("https://omni.example.com")

	// Auth fields
	serviceAccountEntry, oidcIssuerEntry, oidcClientIDEntry, oidcClientSecretEntry, authSelect, serviceAccountContainer, oidcContainer := createAuthFields(settings)

	// Buttons
	formFields := &SettingsFormFields{
		EndpointEntry:         endpointEntry,
		AuthSelect:            authSelect,
		ServiceAccountEntry:   serviceAccountEntry,
		OIDCIssuerEntry:       oidcIssuerEntry,
		OIDCClientIDEntry:     oidcClientIDEntry,
		OIDCClientSecretEntry: oidcClientSecretEntry,
	}
	saveButton := createSaveButton(settingsWindow, parentWindow, formFields)
	cancelButton := widget.NewButton("Cancel", func() {
		settingsWindow.Close()
	})

	// Language selector
	refreshUI := func() {
		if parentWindow != nil {
			parentWindow.Content().Refresh()
		}
	}
	langLabel := widget.NewLabel("Language:")
	langSelect := createLanguageSelector(nil, refreshUI)

	// Tab content
	appSettingsContent := container.NewVBox(langLabel, langSelect)
	authLabel := widget.NewLabel("Authentication Method:")
	authSettingsContent := container.NewVBox(
		endpointLabel, endpointEntry,
		widget.NewSeparator(),
		authLabel, authSelect,
		widget.NewSeparator(),
		serviceAccountContainer, oidcContainer,
	)

	// Layout
	tabs := container.NewAppTabs(
		container.NewTabItem("Application Settings", container.NewScroll(appSettingsContent)),
		container.NewTabItem("Authentication Settings", container.NewScroll(authSettingsContent)),
	)
	content := container.NewVBox(
		tabs,
		widget.NewSeparator(),
		container.NewHBox(cancelButton, saveButton),
	)

	settingsWindow.SetContent(content)
	settingsWindow.Show()
}

// showRestartDialog displays a dialog asking if the user wants to restart the application
func showRestartDialog(settingsWindow fyne.Window, parentWindow fyne.Window) {
	message := widget.NewLabel("Settings saved!\n\nRestart the application now to apply changes?")
	message.Wrapping = fyne.TextWrapWord
	
	var dialog *widget.PopUp
	
	restartButton := widget.NewButton("Restart Now", func() {
		if dialog != nil {
			dialog.Hide()
		}
		restartApplication(parentWindow)
	})
	
	laterButton := widget.NewButton("Later", func() {
		if dialog != nil {
			dialog.Hide()
		}
	})
	
	buttonContainer := container.NewHBox(laterButton, restartButton)
	dialogContent := container.NewVBox(
		message,
		widget.NewSeparator(),
		buttonContainer,
	)
	
	dialog = widget.NewModalPopUp(
		dialogContent,
		settingsWindow.Canvas(),
	)
	dialog.Resize(fyne.NewSize(350, 120))
	dialog.Show()
}

// restartApplication restarts the application with the same command-line arguments
func restartApplication(parentWindow fyne.Window) {
	slog.Info("Restarting application...")
	
	// Get the executable path
	executable, err := os.Executable()
	if err != nil {
		slog.Error("Failed to get executable path", "error", err)
		// Fallback: try to get from os.Args[0]
		executable = os.Args[0]
		// If it's a relative path, make it absolute
		if !filepath.IsAbs(executable) {
			wd, err := os.Getwd()
			if err == nil {
				executable = filepath.Join(wd, executable)
			}
		}
	}
	
	// Get current command-line arguments (excluding program name)
	args := os.Args[1:]
	
	// Create command to restart
	cmd := exec.Command(executable, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	
	// Start the new process
	if err := cmd.Start(); err != nil {
		slog.Error("Failed to restart application", "error", err)
		// Show error to user
		errorDialog := widget.NewModalPopUp(
			widget.NewLabel(fmt.Sprintf("Failed to restart: %v\n\nPlease restart manually.", err)),
			parentWindow.Canvas(),
		)
		errorDialog.Resize(fyne.NewSize(300, 100))
		errorDialog.Show()
		return
	}
	
	// Close the current application
	slog.Info("New process started, exiting current instance")
	parentWindow.Close()
	os.Exit(0)
}
