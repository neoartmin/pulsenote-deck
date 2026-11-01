package main

import "testing"

func TestSearchInstalledApps(t *testing.T) {
	apps := []foundApp{{"Bambu Studio", `C:\Tools\BambuStudio.lnk`}, {"Chrome", `C:\Apps\Chrome.lnk`}, {"Fusion", `C:\Apps\Fusion.lnk`}}
	if len(filterApps(apps, "BAMBU")) != 1 || filterApps(apps, "bambu")[0].Name != "Bambu Studio" {
		t.Fatal("case insensitive names")
	}
	if len(filterApps(apps, "c:\\apps")) != 0 {
	}
	if len(filterApps(apps, `C:\Apps`)) != 2 {
		t.Fatal("path matching")
	}
	if len(filterApps(apps, "absent")) != 0 {
		t.Fatal("missing app matches")
	}
	if len(filterApps(apps, "  ")) != 3 {
		t.Fatal("empty query")
	}
}
func TestPagedSearch(t *testing.T) {
	apps := make([]foundApp, 31)
	for i := range apps {
		apps[i].Name = "App"
	}
	a, p := pageApps(apps, 0, 14)
	if len(a) != 14 || p != 0 {
		t.Fatal("page 0")
	}
	a, p = pageApps(apps, 2, 14)
	if len(a) != 3 || p != 2 {
		t.Fatal("page 2")
	}
	a, p = pageApps(apps, 900, 14)
	if len(a) != 3 || p != 2 {
		t.Fatal("clamp")
	}
	if len(apps) != 31 {
		t.Fatal("modified input")
	}
}
