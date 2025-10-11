package channel

import "fmt"

// ConnectionMode represents how a channel connects to the messaging platform
//
// Two modes are supported:
//
// 1. Manual Mode:
//   - User provides existing WAHA session credentials (base_url, session_id, token)
//   - System just connects and uses the session
//   - User is responsible for WAHA setup and QR code scanning
//
// 2. Auto Mode:
//   - System creates and manages WAHA session automatically
//   - System generates QR code for user to scan
//   - Webhooks notify when session is connected
//   - System can restart/manage session lifecycle
type ConnectionMode string

const (
	// ConnectionModeManual - User provides existing WAHA credentials
	// User manages WAHA instance and session
	ConnectionModeManual ConnectionMode = "manual"

	// ConnectionModeAuto - System creates and manages WAHA session
	// System handles QR code, connection status, and lifecycle
	ConnectionModeAuto ConnectionMode = "auto"
)

// IsValid checks if the connection mode is valid
func (cm ConnectionMode) IsValid() bool {
	switch cm {
	case ConnectionModeManual, ConnectionModeAuto:
		return true
	default:
		return false
	}
}

// String returns the string representation
func (cm ConnectionMode) String() string {
	return string(cm)
}

// ParseConnectionMode parses a string into ConnectionMode
func ParseConnectionMode(s string) (ConnectionMode, error) {
	cm := ConnectionMode(s)
	if !cm.IsValid() {
		return "", fmt.Errorf("invalid connection mode: %s (valid: manual, auto)", s)
	}
	return cm, nil
}

// RequiresQRCode returns true if this mode requires QR code scanning
func (cm ConnectionMode) RequiresQRCode() bool {
	return cm == ConnectionModeAuto
}

// RequiresUserCredentials returns true if this mode requires user to provide credentials
func (cm ConnectionMode) RequiresUserCredentials() bool {
	return cm == ConnectionModeManual
}

// CanSystemManageSession returns true if system can manage session lifecycle
func (cm ConnectionMode) CanSystemManageSession() bool {
	return cm == ConnectionModeAuto
}
