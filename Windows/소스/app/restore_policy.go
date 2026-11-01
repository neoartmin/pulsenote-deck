package main

// shouldRescanOnShow separates UI presentation from actual USB status.
// An already authenticated device with a currently present COM port must
// not be disconnected merely because the user reopened the settings window.
func shouldRescanOnShow(port string, connected bool, livePorts []string) bool {
	if !connected || port == "" {
		return true
	}
	return !containsPort(livePorts, port)
}
