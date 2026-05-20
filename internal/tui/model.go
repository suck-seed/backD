package tui

import "backD/internal/device"

type Screen int

const (
	ScreenDeviceSelect Screen = iota
	ScreenFileBrowser
	ScreenTemplateSelect
	ScreenBackupProgress
	ScreenBackupSummary
	ScreenSaveTemplate
	ScreenError
)

type Model struct {
	screen Screen
	width  int
	height int

	// Device Selection
	devices        []device.ExternalDevice
	deviceCursor   int
	selectedDevice *device.ExternalDevice
}
