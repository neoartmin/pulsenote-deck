package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const version = "7.5.8"

var palette = []string{"파랑", "빨강", "회색", "보라", "주황", "초록", "노랑"}

// Muted keycap colors reduce glare while preserving all seven saved color
// names and the existing screen layout. Win32 COLORREF is 0x00BBGGRR.
// Changing the rendering palette does NOT rewrite users' saved settings.
var paletteRGB = []uint32{
	0xD5B494, // 파랑 #94B4D5, soft blue
	0xA098D9, // 빨강 #D998A0, dusty rose
	0xE8DED6, // 회색 #D6DEE8, cool gray
	0xD5A6B9, // 보라 #B9A6D5, lavender
	0x8FB4E4, // 주황 #E4B48F, apricot
	0xB4C9A0, // 초록 #A0C9B4, sage
	0x9ED0DE, // 노랑 #DED09E, sand
}

// Subtle lip highlights, coordinated with the lower-saturation fills.
var paletteHighlightRGB = []uint32{
	0xE9D3BF, 0xBFB9F0, 0xF9F6F4, 0xEACAD9,
	0xB6D1F2, 0xD5E5C5, 0xBBE4F0,
}

func keycapHighlight(color string) uint32 {
	for i, c := range palette {
		if color == c {
			return paletteHighlightRGB[i]
		}
	}
	return paletteHighlightRGB[2]
}

// Every softened cap is light enough for the original dark foreground.
// Light text on pastel blue/purple would fail readability contrast checks.
func lightKeycap(color string) bool {
	return isColor(color)
}

const maxProfiles = 8

type Key struct {
	Label string `json:"label"`
	Type  string `json:"type"`
	Value string `json:"value"`
	Color string `json:"color"`
}
type Profile struct {
	Name string `json:"name"`
	Keys []Key  `json:"keys"`
}
type Model struct {
	Active   int       `json:"active"`
	Profiles []Profile `json:"profiles"`
}
type Config struct {
	Version           int              `json:"version"`
	Selected          int              `json:"selected"`
	LastVerifiedModel int              `json:"lastVerifiedModel,omitempty"`
	SelectedSide      string           `json:"selectedSide,omitempty"`
	LastVerifiedSide  string           `json:"lastVerifiedSide,omitempty"`
	PreferredCOM      string           `json:"preferredCOM"`
	Models            map[string]Model `json:"models"`
}

var validTypes = map[string]bool{"none": true, "shortcut": true, "program": true, "website": true, "folder": true, "volume_up": true, "volume_down": true, "volume_mute": true, "media_play": true, "media_next": true, "media_prev": true}

func keyDefault(n int) Key {
	return Key{Label: fmt.Sprintf("버튼 %02d", n), Type: "none", Color: "회색"}
}
func newModel(count int) Model {
	p := Profile{Name: "기본 프로필", Keys: make([]Key, count)}
	for i := range p.Keys {
		p.Keys[i] = keyDefault(i + 1)
	}
	return Model{Profiles: []Profile{p}}
}
func defaults() Config {
	return Config{Version: 7, Selected: 15, SelectedSide: "R", Models: map[string]Model{"12": newModel(12), "15": newModel(15), "15L": newModel(15)}}
}
func isColor(s string) bool {
	for _, c := range palette {
		if c == s {
			return true
		}
	}
	return false
}
func countValid(n int) bool { return n == 12 || n == 15 }
func validSide(side string) string {
	if side == "L" {
		return "L"
	}
	return "R"
}

// The original "15" profile remains the PN15 right-hand profile, so no
// existing PN15 settings are discarded. The left-hand model has its own keys.
func modelKeyFor(n int, side string) string {
	if n == 15 && validSide(side) == "L" {
		return "15L"
	}
	return strconv.Itoa(n)
}
func modelKey(c Config) string { return modelKeyFor(c.Selected, c.SelectedSide) }

// Legacy 15-key firmware identifies as PN15 right; dedicated firmware uses
// distinct side tokens. Never infer orientation from the COM port or keypress.
func parseDeviceInfo(s string) (int, string) {
	switch s {
	case "PND:INFO:12:C3":
		return 12, "R"
	case "PND:INFO:15:C3", "PND:INFO:15:R:C3":
		return 15, "R"
	case "PND:INFO:15:L:C3":
		return 15, "L"
	default:
		return 0, ""
	}
}

func normalize(c Config) Config {
	// Keep PN12 profiles and old PN8 archival data, but default to the
	// sole PN15 GRIP hardware layout on a fresh installation.
	if !countValid(c.Selected) {
		c.Selected = 15
	}
	c.SelectedSide = validSide(c.SelectedSide)
	c.LastVerifiedSide = validSide(c.LastVerifiedSide)
	if c.Models == nil {
		c.Models = map[string]Model{}
	}
	for _, count := range []int{12, 15} {
		id := strconv.Itoa(count)
		m, ok := c.Models[id]
		if !ok || len(m.Profiles) == 0 {
			m = newModel(count)
		}
		if len(m.Profiles) > maxProfiles {
			m.Profiles = m.Profiles[:maxProfiles]
		}
		if m.Active < 0 || m.Active >= len(m.Profiles) {
			m.Active = 0
		}
		for j := range m.Profiles {
			p := &m.Profiles[j]
			if strings.TrimSpace(p.Name) == "" {
				p.Name = fmt.Sprintf("프로필 %d", j+1)
			}
			if len(p.Keys) > count {
				p.Keys = p.Keys[:count]
			}
			for len(p.Keys) < count {
				p.Keys = append(p.Keys, keyDefault(len(p.Keys)+1))
			}
			for k := range p.Keys {
				q := &p.Keys[k]
				if !validTypes[q.Type] {
					q.Type = "none"
					q.Value = ""
				}
				if !isColor(q.Color) {
					q.Color = "회색"
				}
				if q.Label == "" {
					q.Label = fmt.Sprintf("버튼 %02d", k+1)
				}
				if len([]rune(q.Label)) > 45 {
					q.Label = string([]rune(q.Label)[:45])
				}
				if len(q.Value) > 4096 {
					q.Value = q.Value[:4096]
				}
			}
		}
		c.Models[id] = m
	}
	// New left-handed profile is separate from the original PN15 profile.
	left, ok := c.Models["15L"]
	if !ok || len(left.Profiles) == 0 {
		left = newModel(15)
	}
	if len(left.Profiles) > maxProfiles {
		left.Profiles = left.Profiles[:maxProfiles]
	}
	if left.Active < 0 || left.Active >= len(left.Profiles) {
		left.Active = 0
	}
	for j := range left.Profiles {
		p := &left.Profiles[j]
		if strings.TrimSpace(p.Name) == "" {
			p.Name = fmt.Sprintf("프로필 %d", j+1)
		}
		if len(p.Keys) > 15 {
			p.Keys = p.Keys[:15]
		}
		for len(p.Keys) < 15 {
			p.Keys = append(p.Keys, keyDefault(len(p.Keys)+1))
		}
		for k := range p.Keys {
			q := &p.Keys[k]
			if !validTypes[q.Type] {
				q.Type = "none"
				q.Value = ""
			}
			if !isColor(q.Color) {
				q.Color = "회색"
			}
			if q.Label == "" {
				q.Label = fmt.Sprintf("버튼 %02d", k+1)
			}
			if len([]rune(q.Label)) > 45 {
				q.Label = string([]rune(q.Label)[:45])
			}
			if len(q.Value) > 4096 {
				q.Value = q.Value[:4096]
			}
		}
	}
	c.Models["15L"] = left
	c.Version = 7
	c.PreferredCOM = strings.ToUpper(strings.TrimSpace(c.PreferredCOM))
	return c
}

// Offline editing uses the last selected UI tab, independent of USB identity.
// Existing profiles are never reset just because a device is disconnected.
func restoreAutomaticModel(c Config) Config {
	return normalize(c)
}
func setEditorModel(c *Config, count int, side string) bool {
	if !countValid(count) {
		return false
	}
	side = validSide(side)
	if c.Selected == count && c.SelectedSide == side {
		return false
	}
	c.Selected = count
	c.SelectedSide = side
	return true
}

func active(c Config) (int, Model, Profile) {
	n := c.Selected
	m := c.Models[modelKey(c)]
	return n, m, m.Profiles[m.Active]
}
func selectedKey(c Config, idx int) (Key, bool) {
	n, _, p := active(c)
	if idx < 0 || idx >= n {
		return Key{}, false
	}
	return p.Keys[idx], true
}
func mutateKey(c *Config, idx int, fn func(*Key)) bool {
	n, m, p := active(*c)
	if idx < 0 || idx >= n {
		return false
	}
	fn(&p.Keys[idx])
	m.Profiles[m.Active] = p
	c.Models[modelKey(*c)] = m
	return true
}
func getConfigPath() string {
	dir := os.Getenv("APPDATA")
	if dir == "" {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "PulseNoteDeck", "config_unified.json")
}

var store = struct {
	sync.RWMutex
	C    Config
	Path string
	Err  error
}{}

// atomicWriteFile writes through a unique temporary file so interruption
// cannot leave a truncated main settings file or .bak snapshot.
func atomicWriteFile(path string, data []byte) error {
	directory := filepath.Dir(path)
	f, err := os.CreateTemp(directory, ".pnd-settings-*.tmp")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	if err = f.Chmod(0600); err != nil {
		f.Close()
		return err
	}
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func saveJSON(path string, c Config) error {
	b, e := json.MarshalIndent(normalize(c), "", "  ")
	if e != nil {
		return e
	}
	if e = os.MkdirAll(filepath.Dir(path), 0700); e != nil {
		return e
	}
	// Only preserve a known-valid snapshot; never overwrite damaged profiles.
	if _, statErr := os.Stat(path); statErr == nil {
		if _, readErr := readJSON(path); readErr != nil {
			return fmt.Errorf("기존 설정 손상, 덮어쓰기 중단: %w", readErr)
		}
		old, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		// Frequent UI refreshes must not rewrite unchanged settings or rotate
		// away the last useful backup of a genuinely different configuration.
		if bytes.Equal(old, b) {
			return nil
		}
		if readErr = atomicWriteFile(path+".bak", old); readErr != nil {
			return readErr
		}
	}
	return atomicWriteFile(path, b)
}
func readJSON(path string) (Config, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return Config{}, e
	}
	var c Config
	if e = json.Unmarshal(b, &c); e != nil {
		return Config{}, e
	}
	if c.Version != 7 {
		return Config{}, errors.New("알 수 없는 설정 버전")
	}
	return normalize(c), nil
}
func loadStore() {
	store.Path = getConfigPath()
	c, e := readJSON(store.Path)
	if e == nil {
		store.C = restoreAutomaticModel(c)
		return
	}
	// Recover only a verified backup. Preserve corrupted bytes as a separately
	// named file instead of silently discarding the user's original data.
	if recovered, re := readJSON(store.Path + ".bak"); re == nil {
		if !errors.Is(e, os.ErrNotExist) {
			// Keep previous damaged snapshots instead of overwriting them.
			broken := store.Path + ".corrupt"
			for i := 1; ; i++ {
				if _, statErr := os.Stat(broken); errors.Is(statErr, os.ErrNotExist) {
					break
				}
				broken = fmt.Sprintf("%s.corrupt.%d", store.Path, i)
			}
			if mvErr := os.Rename(store.Path, broken); mvErr != nil {
				store.Err = fmt.Errorf("기존 파일 보존 실패: %w", mvErr)
				store.C = restoreAutomaticModel(recovered)
				return
			}
		}
		if saveErr := saveJSON(store.Path, recovered); saveErr != nil {
			store.Err = saveErr
		}
		store.C = restoreAutomaticModel(recovered)
		return
	}
	if !errors.Is(e, os.ErrNotExist) {
		store.Err = fmt.Errorf("설정 파일 보호: %w", e)
		store.C = defaults()
		return
	}
	store.C = defaults()
	_ = importLegacy(&store.C, filepath.Join(filepath.Dir(store.Path), "config_v4.json"))
	store.C = normalize(store.C)
}

// Import in memory only, keep v6 config_v4.json untouched. Unknown formats are ignored.
func importLegacy(c *Config, path string) error {
	b, e := os.ReadFile(path)
	if e != nil {
		return e
	}
	var root struct {
		Profiles []struct {
			Name string `json:"name"`
			Keys []struct {
				Label string `json:"label"`
				Name  string `json:"name"`
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"keys"`
		} `json:"profiles"`
		Active int `json:"active"`
	}
	if e = json.Unmarshal(b, &root); e != nil {
		return e
	}
	if len(root.Profiles) == 0 {
		return errors.New("no compatible profiles")
	}
	m := newModel(12)
	m.Profiles = nil
	for _, old := range root.Profiles {
		p := Profile{Name: old.Name, Keys: make([]Key, 12)}
		for j := 0; j < 12; j++ {
			p.Keys[j] = keyDefault(j + 1)
			if j < len(old.Keys) {
				v := old.Keys[j]
				p.Keys[j].Label = v.Label
				if p.Keys[j].Label == "" {
					p.Keys[j].Label = v.Name
				}
				p.Keys[j].Type = v.Type
				p.Keys[j].Value = v.Value
			}
		}
		m.Profiles = append(m.Profiles, p)
	}
	m.Active = root.Active
	c.Models["12"] = m
	return nil
}
func snap() Config { store.RLock(); defer store.RUnlock(); return store.C }
func update(fn func(*Config)) error {
	store.Lock()
	defer store.Unlock()
	if store.Err != nil {
		return store.Err
	}
	orig := store.C
	c := copyConfig(orig)
	fn(&c)
	c = normalize(c)
	if err := saveJSON(store.Path, c); err != nil {
		return err
	}
	store.C = c
	return nil
}
func copyConfig(c Config) Config {
	out := c
	out.Models = make(map[string]Model, len(c.Models))
	for k, v := range c.Models {
		v2 := v
		v2.Profiles = make([]Profile, len(v.Profiles))
		for i, p := range v.Profiles {
			q := p
			q.Keys = append([]Key(nil), p.Keys...)
			v2.Profiles[i] = q
		}
		out.Models[k] = v2
	}
	return out
}
func colorValue(color string) uint32 {
	for i, c := range palette {
		if c == color {
			return paletteRGB[i]
		}
	}
	return paletteRGB[2]
}
func parseInfo(s string) int { n, _ := parseDeviceInfo(s); return n }
func parseKey(s string) (int, bool) {
	if !strings.HasPrefix(s, "KEY:") {
		return 0, false
	}
	v := s[4:]
	if len(v) == 0 || len(v) > 2 {
		return 0, false
	}
	for _, r := range v {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, e := strconv.Atoi(v)
	return n, e == nil && n >= 1 && n <= 15
}
