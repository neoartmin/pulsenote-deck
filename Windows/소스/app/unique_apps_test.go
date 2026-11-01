package main

import "testing"

func TestUniqueAppsPreservesSameNamedDifferentPaths(t *testing.T) {
	list := uniqueApps([]foundApp{
		{Name: "Editor", Path: `C:\A\Editor.lnk`},
		{Name: "Editor", Path: `C:\B\Editor.lnk`},
		{Name: "editor", Path: `c:\a\editor.lnk`},
		{Name: "Browser", Path: `C:\Browsers\Browser.lnk`},
	})
	if len(list) != 3 || list[0].Name != "Browser" {
		t.Fatalf("sorted distinct shortcuts: %#v", list)
	}
	if list[1].Path == list[2].Path {
		t.Fatal("distinct names collision")
	}
}
