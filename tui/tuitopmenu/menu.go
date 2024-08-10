package tuitopmenu

import (
	"embed"
)

//go:embed ../apps/*
var appFiles embed.FS

type NameAndHash struct {
	Name string
	Hash uint64
}

type Menu struct {
	PreviousDefaultApps []NameAndHash
}

func (m *Menu) LoadApps() error {
	apps, err := appFiles.ReadDir("../apps")
	if err != nil {
		return err
	}
	for _, app := range apps {
		if app.IsDir() {
			continue
		}

		cityhash.CityHash64(appFiles.ReadFile(app.Name())) // hash the name

		// TODO
		// if this is in m.PreviousDefaultApps, or existing, continue-for
		// else, add it to the menu. If name already exists, add a number to the end.

	}
}
