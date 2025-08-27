package gui

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"kryptx/internal/config"
	"kryptx/internal/network"
	"kryptx/internal/utils"
)

type App struct {
	app       fyne.App
	window    fyne.Window
	vpnClient *network.VPNClient
	config    *config.Config
	logger    *utils.Logger

	// UI components
	statusLabel    *widget.Label
	connectButton  *widget.Button
	serverLabel    *widget.Label
	ipLabel        *widget.Label
	statsContainer *fyne.Container
}

func NewApp(vpnClient *network.VPNClient, cfg *config.Config, logger *utils.Logger) *App {
	a := app.NewWithID("com.kpolitx.kryptx")
	a.SetIcon(resourceIconPng) // You'd need to embed an icon

	if cfg.GUI.Theme == "dark" {
		a.Settings().SetTheme(&CyberpunkTheme{})
	}

	return &App{
		app:       a,
		vpnClient: vpnClient,
		config:    cfg,
		logger:    logger,
	}
}

func (a *App) Run() {
	a.window = a.app.NewWindow("KryptX VPN")
	a.window.SetFixedSize(true)
	a.window.Resize(fyne.NewSize(400, 500))
	a.window.CenterOnScreen()

	a.setupUI()
	a.startStatusUpdater()

	a.window.ShowAndRun()
}

func (a *App) setupUI() {
	// Header
	title := widget.NewLabelWithStyle("KryptX VPN", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	title.TextStyle.Monospace = true

	// Status section
	a.statusLabel = widget.NewLabel("Disconnected")
	a.statusLabel.TextStyle.Bold = true

	statusCard := widget.NewCard("Status", "", a.statusLabel)

	// Connection button
	a.connectButton = widget.NewButton("Connect", a.toggleConnection)
	a.connectButton.Importance = widget.HighImportance

	// Server info
	a.serverLabel = widget.NewLabel(fmt.Sprintf("Server: %s", a.config.Server.Endpoint))
	a.ipLabel = widget.NewLabel("IP: Not connected")

	serverCard := widget.NewCard("Connection Info", "",
		container.NewVBox(a.serverLabel, a.ipLabel))

	// Stats section
	a.statsContainer = container.NewVBox()
	statsCard := widget.NewCard("Statistics", "", a.statsContainer)

	// Settings button
	settingsButton := widget.NewButton("Settings", a.showSettings)
	settingsButton.Importance = widget.MediumImportance

	// Main layout
	content := container.NewVBox(
		title,
		widget.NewSeparator(),
		statusCard,
		a.connectButton,
		widget.NewSeparator(),
		serverCard,
		statsCard,
		widget.NewSeparator(),
		settingsButton,
	)

	scrollable := container.NewScroll(content)
	a.window.SetContent(scrollable)
}

func (a *App) toggleConnection() {
	if a.vpnClient.IsConnected() {
		a.disconnect()
	} else {
		a.connect()
	}
}

func (a *App) connect() {
	a.connectButton.SetText("Connecting...")
	a.connectButton.Disable()
	a.statusLabel.SetText("Connecting...")

	go func() {
		if err := a.vpnClient.Connect(a.app.Driver().AllWindows()[0].Canvas().Context()); err != nil {
			a.logger.Error("Connection failed: %v", err)

			// Update UI on main thread
			a.connectButton.SetText("Connect")
			a.connectButton.Enable()
			a.statusLabel.SetText("Connection Failed")

			// Show error dialog
			dialog := widget.NewModalPopUp(
				widget.NewLabel(fmt.Sprintf("Connection failed: %v", err)),
				a.window.Canvas(),
			)
			dialog.Show()

			time.AfterFunc(3*time.Second, dialog.Hide)
		}
	}()
}

func (a *App) disconnect() {
	a.connectButton.SetText("Disconnecting...")
	a.connectButton.Disable()

	go func() {
		if err := a.vpnClient.Disconnect(); err != nil {
			a.logger.Error("Disconnect failed: %v", err)
		}

		a.connectButton.SetText("Connect")
		a.connectButton.Enable()
	}()
}

func (a *App) updateStatus() {
	status := a.vpnClient.GetStatus()

	if connected, ok := status["connected"].(bool); ok && connected {
		a.statusLabel.SetText("Connected")
		a.connectButton.SetText("Disconnect")
		a.connectButton.Enable()

		if ip, ok := status["public_ip"].(string); ok {
			a.ipLabel.SetText(fmt.Sprintf("IP: %s", ip))
		}

		// Update stats if available
		if stats, ok := status["stats"].(map[string]interface{}); ok {
			a.updateStats(stats)
		}
	} else {
		a.statusLabel.SetText("Disconnected")
		a.connectButton.SetText("Connect")
		a.connectButton.Enable()
		a.ipLabel.SetText("IP: Not connected")
		a.statsContainer.RemoveAll()
	}
}

func (a *App) updateStats(stats map[string]interface{}) {
	a.statsContainer.RemoveAll()

	for key, value := range stats {
		label := widget.NewLabel(fmt.Sprintf("%s: %v", key, value))
		a.statsContainer.Add(label)
	}
}

func (a *App) startStatusUpdater() {
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			a.updateStatus()
		}
	}()
}

func (a *App) showSettings() {
	// Settings dialog implementation
	settingsWindow := a.app.NewWindow("Settings")
	settingsWindow.Resize(fyne.NewSize(300, 400))

	// Create settings form
	form := widget.NewForm()

	// Add settings fields here...

	settingsWindow.SetContent(form)
	settingsWindow.Show()
}
