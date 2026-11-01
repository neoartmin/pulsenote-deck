package main

// Route a simultaneous USB keypress using the identity confirmed on that port.
// Never use the global UI selection to resolve a hardware action.
func resolveHardwareKey(c Config, model int, side string, oneBasedKey int) (Key, bool) {
	if !countValid(model) || oneBasedKey < 1 || oneBasedKey > model {
		return Key{}, false
	}
	profile, ok := c.Models[modelKeyFor(model, side)]
	if !ok || profile.Active < 0 || profile.Active >= len(profile.Profiles) {
		return Key{}, false
	}
	keys := profile.Profiles[profile.Active].Keys
	if oneBasedKey > len(keys) {
		return Key{}, false
	}
	return keys[oneBasedKey-1], true
}

// Three distinct device identities: two 15-key and one 12-key.
func connectedDeviceCount(keys int) int {
	switch keys {
	case 42:
		return 3
	case 27, 30:
		return 2
	case 12, 15:
		return 1
	default:
		return 0
	}
}
