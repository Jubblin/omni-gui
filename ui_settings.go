package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	omniclient "github.com/jubblin/omni-api/internal/client"
)

// createSaveButton creates the save button with settings save logic
func createSaveButton(settingsWindow fyne.Window, parentWindow fyne.Window, fields *SettingsFormFields, appState *AppState) *widget.Button {
	return widget.NewButton("Save", func() {
		// Parse gRPC debug level from selected string
		grpcDebugLevel := 0
		if fields.GrpcDebugLevelSelect.Selected != "" {
			// Extract the first character (the level number)
			if len(fields.GrpcDebugLevelSelect.Selected) > 0 {
				fmt.Sscanf(fields.GrpcDebugLevelSelect.Selected, "%d", &grpcDebugLevel)
			}
		}
		
		// Load old settings BEFORE saving to compare what changed
		// Use ignoreEnv=true to get pure file values (not affected by environment variables)
		// This ensures we compare file-to-file, not file-to-env-var-overridden values
		oldSettings, _ := omniclient.LoadSettings(true)
		
		newSettings := &omniclient.Settings{
			Endpoint:        fields.EndpointEntry.Text,
			ShowEmptyLabels: fields.ShowEmptyLabelsCheck.Checked,
			IgnoreEnv:       fields.IgnoreEnvCheck.Checked,
			GrpcDebugLevel:  grpcDebugLevel,
		}
		
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
		
		// Apply settings dynamically
		applySettingsDynamically(newSettings, oldSettings, appState, settingsWindow, parentWindow)
	})
}

// applySettingsDynamically applies settings changes without requiring restart
// 
// Application settings (applied dynamically, no restart needed):
//   - ShowEmptyLabels: Controls tree node filtering
//   - GrpcDebugLevel: Controls gRPC debugging verbosity
//   - IgnoreEnv: Controls whether environment variables are used (affects future settings loading)
//   - Language: Applied immediately via UI refresh (handled separately in showSettingsPage)
//
// Authentication settings (require restart):
//   - Endpoint: Omni API endpoint URL
//   - AuthMethod: Authentication method (service_account or oidc)
//   - ServiceAccount: Service account key
//   - OIDCIssuerURL, OIDCClientID, OIDCClientSecret: OIDC credentials
func applySettingsDynamically(newSettings *omniclient.Settings, oldSettings *omniclient.Settings, appState *AppState, settingsWindow fyne.Window, parentWindow fyne.Window) {
	// Apply ShowEmptyLabels dynamically
	globalShowEmptyLabelNodes := newSettings.ShowEmptyLabels
	showEmptyLabelNodes = globalShowEmptyLabelNodes
	slog.Info("Applied ShowEmptyLabels setting dynamically", "value", globalShowEmptyLabelNodes)
	
	// Apply gRPC debug level dynamically
	setGrpcDebugLevel(newSettings.GrpcDebugLevel)
	slog.Info("Applied gRPC debug level setting dynamically", "level", newSettings.GrpcDebugLevel)
	
	// Refresh the tree to apply the new filtering setting
	if appState != nil && appState.resourceTree != nil {
		refreshTreeData(appState)
		slog.Info("Tree refreshed with new empty label filtering setting")
	}

	// Apply IgnoreEnv setting dynamically
	// Note: This only affects how settings are loaded in the future, not the current client
	// The current client continues to work with its existing configuration
	if oldSettings != nil && oldSettings.IgnoreEnv != newSettings.IgnoreEnv {
		slog.Info("Applied IgnoreEnv setting dynamically", "ignore_env", newSettings.IgnoreEnv, "note", "This affects future settings loading, current client unchanged")
	} else if oldSettings == nil {
		slog.Info("Applied IgnoreEnv setting dynamically", "ignore_env", newSettings.IgnoreEnv)
	}
	
	// Check if authentication or endpoint settings changed (these require restart)
	// Application settings (ShowEmptyLabels, GrpcDebugLevel, IgnoreEnv) are applied dynamically above
	// Only authentication-related settings should trigger a restart
	needsRestart := false
	if oldSettings != nil {
		// Compare only authentication/endpoint settings (these require restart)
		// Application settings are already applied dynamically above
		endpointChanged := oldSettings.Endpoint != newSettings.Endpoint
		authMethodChanged := oldSettings.AuthMethod != newSettings.AuthMethod
		serviceAccountChanged := oldSettings.ServiceAccount != newSettings.ServiceAccount
		oidcIssuerChanged := oldSettings.OIDCIssuerURL != newSettings.OIDCIssuerURL
		oidcClientIDChanged := oldSettings.OIDCClientID != newSettings.OIDCClientID
		oidcClientSecretChanged := oldSettings.OIDCClientSecret != newSettings.OIDCClientSecret
		
		// Verify that application settings did NOT change (sanity check)
		showEmptyLabelsChanged := oldSettings.ShowEmptyLabels != newSettings.ShowEmptyLabels
		grpcDebugLevelChanged := oldSettings.GrpcDebugLevel != newSettings.GrpcDebugLevel
		ignoreEnvChanged := oldSettings.IgnoreEnv != newSettings.IgnoreEnv
		
		if showEmptyLabelsChanged || grpcDebugLevelChanged || ignoreEnvChanged {
			slog.Debug("Application settings changed (applied dynamically, no restart needed)",
				"show_empty_labels_changed", showEmptyLabelsChanged,
				"grpc_debug_level_changed", grpcDebugLevelChanged,
				"ignore_env_changed", ignoreEnvChanged)
		}
		
		if endpointChanged || authMethodChanged || serviceAccountChanged || oidcIssuerChanged || oidcClientIDChanged || oidcClientSecretChanged {
			needsRestart = true
			slog.Info("Restart required due to authentication/endpoint changes",
				"endpoint_changed", endpointChanged,
				"auth_method_changed", authMethodChanged,
				"service_account_changed", serviceAccountChanged,
				"oidc_issuer_changed", oidcIssuerChanged,
				"oidc_client_id_changed", oidcClientIDChanged,
				"oidc_client_secret_changed", oidcClientSecretChanged,
				"old_endpoint", oldSettings.Endpoint,
				"new_endpoint", newSettings.Endpoint,
				"old_auth_method", oldSettings.AuthMethod,
				"new_auth_method", newSettings.AuthMethod)
		} else {
			slog.Info("No authentication/endpoint changes detected - all application settings applied dynamically, no restart needed")
		}
	} else {
		// First time saving settings - check if any authentication settings are being set
		// If endpoint or auth credentials are provided, we might need a restart
		// But if only application settings are being saved, no restart needed
		if newSettings.Endpoint != "" || newSettings.AuthMethod != "" || 
		   newSettings.ServiceAccount != "" || newSettings.OIDCClientID != "" {
			// Authentication settings are being configured for the first time
			// This requires a restart to initialize the client
			needsRestart = true
			slog.Info("Restart required - authentication settings being configured for the first time")
		} else {
			slog.Info("No authentication settings configured - only application settings saved, no restart needed")
		}
	}

	if needsRestart {
		showRestartDialog(settingsWindow, parentWindow)
	} else {
		// Settings applied successfully, close the window
		settingsWindow.Close()
		successDialog := widget.NewModalPopUp(
			widget.NewLabel("Settings saved and applied successfully!"),
			parentWindow.Canvas(),
		)
		successDialog.Resize(fyne.NewSize(300, 100))
		successDialog.Show()
		// Auto-close after 2 seconds
		go func() {
			time.Sleep(2 * time.Second)
			successDialog.Hide()
		}()
	}
}

// showSettingsPage displays the settings window
func showSettingsPage(parentWindow fyne.Window, appState *AppState) {
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

	// Language selector
	refreshUI := func() {
		if parentWindow != nil {
			parentWindow.Content().Refresh()
		}
	}
	langLabel := widget.NewLabel("Language:")
	langSelect := createLanguageSelector(nil, refreshUI)

	// Empty Label Filtering checkbox
	emptyLabelFilteringLabel := widget.NewLabel("Empty Label Filtering:")
	emptyLabelFilteringDescription := widget.NewLabel("By default, nodes with empty labels are filtered out to prevent blank entries in the tree. Enable this option (or use --show-empty-labels command-line flag) to disable filtering and show all nodes, including those with empty labels. This affects tree rendering at all levels: child node loading, tree widget rendering, and node filtering.")
	emptyLabelFilteringDescription.Wrapping = fyne.TextWrapWord
	emptyLabelFilteringCheck := widget.NewCheck("Show nodes with empty labels in the tree (disable filtering)", func(checked bool) {
		// Changes apply dynamically on save
	})
	emptyLabelFilteringCheck.SetChecked(settings.ShowEmptyLabels)

	// Ignore Environment Variables checkbox
	ignoreEnvLabel := widget.NewLabel("Ignore Environment Variables:")
	ignoreEnvDescription := widget.NewLabel("When enabled, the application will only use settings from the settings file and ignore all environment variables. When disabled, environment variables take precedence over settings file values.")
	ignoreEnvDescription.Wrapping = fyne.TextWrapWord
	ignoreEnvCheck := widget.NewCheck("Ignore environment variables and use only settings file", func(checked bool) {
		// Changes apply dynamically on save
	})
	ignoreEnvCheck.SetChecked(settings.IgnoreEnv)

	// gRPC Debug Level selector
	grpcDebugLabel := widget.NewLabel("gRPC Debug Level:")
	grpcDebugDescription := widget.NewLabel("Controls the level of gRPC debugging output. Level 0: Disabled. Level 1: Query dumps only (logs query metadata as JSON). Level 2: Query + Response dumps (logs query and response info). Level 3: Query + Response + UI Display dumps (also creates dump files when resources are displayed).")
	grpcDebugDescription.Wrapping = fyne.TextWrapWord
	grpcDebugSelect := widget.NewSelect([]string{"0 - Disabled", "1 - Query dumps only", "2 - Query + Response dumps", "3 - Full debugging (includes UI dumps)"}, func(selected string) {
		// Changes apply dynamically on save
	})
	// Set initial value based on settings
	if settings.GrpcDebugLevel >= 0 && settings.GrpcDebugLevel <= 3 {
		grpcDebugSelect.SetSelected(fmt.Sprintf("%d - %s", settings.GrpcDebugLevel, 
			[]string{"Disabled", "Query dumps only", "Query + Response dumps", "Full debugging (includes UI dumps)"}[settings.GrpcDebugLevel]))
	} else {
		grpcDebugSelect.SetSelected("0 - Disabled")
	}

	// Buttons
	formFields := &SettingsFormFields{
		EndpointEntry:         endpointEntry,
		AuthSelect:            authSelect,
		ServiceAccountEntry:   serviceAccountEntry,
		OIDCIssuerEntry:       oidcIssuerEntry,
		OIDCClientIDEntry:     oidcClientIDEntry,
		OIDCClientSecretEntry: oidcClientSecretEntry,
		ShowEmptyLabelsCheck:  emptyLabelFilteringCheck,
		IgnoreEnvCheck:        ignoreEnvCheck,
		GrpcDebugLevelSelect:  grpcDebugSelect,
	}
	saveButton := createSaveButton(settingsWindow, parentWindow, formFields, appState)
	cancelButton := widget.NewButton("Cancel", func() {
		settingsWindow.Close()
	})

	// Tab content - utilizing at least 75% of usable space
	appSettingsContent := container.NewVBox(
		langLabel, langSelect,
		widget.NewSeparator(),
		emptyLabelFilteringLabel,
		emptyLabelFilteringDescription,
		emptyLabelFilteringCheck,
		widget.NewSeparator(),
		ignoreEnvLabel,
		ignoreEnvDescription,
		ignoreEnvCheck,
		widget.NewSeparator(),
		grpcDebugLabel,
		grpcDebugDescription,
		grpcDebugSelect,
	)
	authLabel := widget.NewLabel("Authentication Method:")
	authSettingsContent := container.NewVBox(
		endpointLabel, endpointEntry,
		widget.NewSeparator(),
		authLabel, authSelect,
		widget.NewSeparator(),
		serviceAccountContainer, oidcContainer,
	)

	// Layout - tabs utilize at least 75% of usable space
	tabs := container.NewAppTabs(
		container.NewTabItem("Application Settings", container.NewScroll(appSettingsContent)),
		container.NewTabItem("Authentication Settings", container.NewScroll(authSettingsContent)),
	)
	
	// Use Border layout to ensure tabs take up most of the space (at least 75%)
	// Top: tabs (takes remaining space), Bottom: buttons
	buttonContainer := container.NewHBox(cancelButton, saveButton)
	content := container.NewBorder(
		nil,                    // top
		container.NewVBox(      // bottom
			widget.NewSeparator(),
			buttonContainer,
		),
		nil,                    // left
		nil,                    // right
		tabs,                   // center - takes remaining space (at least 75%)
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
	
	// On macOS, GLFW can have issues when starting a new instance while the old one is running
	// Use a shell script approach to properly detach the new process
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		// On macOS, use a shell command that properly detaches the process
		// This avoids GLFW initialization conflicts
		shellCmd := fmt.Sprintf("sleep 0.5 && exec '%s'", executable)
		for _, arg := range args {
			shellCmd += fmt.Sprintf(" '%s'", arg)
		}
		cmd = exec.Command("sh", "-c", shellCmd)
		// Redirect output to avoid conflicts
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else if runtime.GOOS == "windows" {
		// On Windows, use cmd.exe to start the process detached
		allArgs := append([]string{"/C", "start", "/B", executable}, args...)
		cmd = exec.Command("cmd.exe", allArgs...)
	} else {
		// On Linux and other Unix-like systems
		cmd = exec.Command(executable, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		// Set process attributes to detach from parent and create new process group
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
			Pgid:    0,
		}
	}
	
	// Start the new process in the background
	if err := cmd.Start(); err != nil {
		slog.Error("Failed to restart application", "error", err)
		// Show error to user
		errorDialog := widget.NewModalPopUp(
			widget.NewLabel(fmt.Sprintf("Failed to restart: %v\n\nPlease restart manually.", err)),
			parentWindow.Canvas(),
		)
		errorDialog.Resize(fyne.NewSize(300, 150))
		errorDialog.Show()
		return
	}
	
	// Detach from the new process so it can continue independently
	_ = cmd.Process.Release()
	
	// Close the parent window and exit after a brief delay
	// The delay allows the new process to start before we release GLFW resources
	go func() {
		time.Sleep(300 * time.Millisecond)
		slog.Info("New process started, exiting current instance")
		parentWindow.Close()
		os.Exit(0)
	}()
}
