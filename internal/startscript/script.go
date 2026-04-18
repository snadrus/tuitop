package startscript

import (
	"fmt"
	"os/exec"

	"github.com/Gaurav-Gosain/tuios/pkg/tuios"
)

// Script describes how tuitop launches a child program in a terminal window so
// every user gets the same argv and environment (theme path, etc.).
type Script struct {
	WindowTitle string
	Spawn       *tuios.WindowSpawn
}

// YaziFileManager returns a start script that runs yazi with YAZI_CONFIG_HOME set to
// configHome (the host should pass the directory from yaziembed.ConfigDir()).
func YaziFileManager(configHome string) (Script, error) {
	return YaziFileManagerAt(configHome, "My Computer", "")
}

// YaziFileManagerAt runs yazi like YaziFileManager; when openDir is non-empty it is passed
// as the first argument so yazi opens that path.
func YaziFileManagerAt(configHome, windowTitle, openDir string) (Script, error) {
	path, err := exec.LookPath("yazi")
	if err != nil {
		return Script{}, fmt.Errorf("yazi not found in PATH: %w", err)
	}
	var args []string
	if openDir != "" {
		args = []string{openDir}
	}
	return Script{
		WindowTitle: windowTitle,
		Spawn: &tuios.WindowSpawn{
			Program: path,
			Args:    args,
			Env: map[string]string{
				"YAZI_CONFIG_HOME": configHome,
			},
		},
	}, nil
}
