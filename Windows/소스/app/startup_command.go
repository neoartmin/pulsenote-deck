package main

import "strings"

// Autostart belongs to the actual executable, not any obsolete v7.x portable
// filename left in the registry after the user downloaded a newer version.
func startupMatches(command, executable string) bool {
	command = strings.TrimSpace(command)
	executable = strings.TrimSpace(executable)
	if command == "" || executable == "" {
		return false
	}
	expected := `"` + executable + `" --background`
	return strings.EqualFold(command, expected)
}
