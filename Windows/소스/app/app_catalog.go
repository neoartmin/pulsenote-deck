package main

import (
	"sort"
	"strings"
)

type foundApp struct{ Name, Path string }

// filtering happens in memory and never blocks on filesystem enumeration.
// Results are naturally sorted by the scanner, so typing does not rearrange
// unrelated entries into surprising positions.
func filterApps(items []foundApp, query string) []foundApp {
	q := strings.ToLower(strings.TrimSpace(query))
	out := make([]foundApp, 0, len(items))
	for _, a := range items {
		if q == "" || strings.Contains(strings.ToLower(a.Name), q) || strings.Contains(strings.ToLower(a.Path), q) {
			out = append(out, a)
		}
	}
	return out
}

func pageApps(items []foundApp, page, size int) ([]foundApp, int) {
	if size <= 0 {
		return nil, 0
	}
	pages := (len(items) + size - 1) / size
	if pages == 0 {
		return nil, 0
	}
	if page < 0 {
		page = 0
	}
	if page >= pages {
		page = pages - 1
	}
	start := page * size
	end := start + size
	if end > len(items) {
		end = len(items)
	}
	return items[start:end], page
}

// Keep same-named shortcuts from different applications or directories.
// Earlier the scanner deduplicated by display name, silently dropping apps.
func uniqueApps(items []foundApp) []foundApp {
	found := make(map[string]bool, len(items))
	out := make([]foundApp, 0, len(items))
	for _, app := range items {
		path := strings.ToLower(strings.TrimSpace(app.Path))
		if path == "" || found[path] {
			continue
		}
		found[path] = true
		out = append(out, app)
	}
	sort.Slice(out, func(i, j int) bool {
		n1, n2 := strings.ToLower(out[i].Name), strings.ToLower(out[j].Name)
		if n1 == n2 {
			return strings.ToLower(out[i].Path) < strings.ToLower(out[j].Path)
		}
		return n1 < n2
	})
	return out
}
