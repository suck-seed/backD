package backup

// ProcessMsg is sent on progress channel after every file decision
type ProcessMsg struct {
	Stats Stats
	Done  bool  // true on the final message
	Err   error // non-nil only when the entire run failed before it started
}

// Options control how a backup run behaves
type Options struct {

	// Sources is list of absolute source path (files or directories)
	Sources []string

	// DeviceMount is root mount of the external storage device
	// /run/media/nothing/Sandesh Extra
	DeviceMount string
}
