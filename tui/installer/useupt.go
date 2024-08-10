package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path"

	"github.com/snadrus/tuitop/tui/tuiwindow"
)

func installUpt() error {
	cmd := exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/sigoden/upt/main/install.sh | sh -s -- --to ~/.config/tuitop/bin")
	return cmd.Run()
}

// VerifyUpt checks if upt is installed and in the PATH
// It updates the path.
// If upt is not found, it installs it.
func VerifyUpt() (string, error) {
	p, err := exec.LookPath("upt")
	if err != nil {
		if err != exec.ErrNotFound {
			return "", err
		}
		// not found
	}
	if p != "" {
		return p, nil
	}

	err = EnsureTuitopPath()
	if err != nil {
		return "", err
	}
	p, err = exec.LookPath("upt")
	if err != nil {
		if err != exec.ErrNotFound {
			goto doInstall
		}
		return "", err
	}
	return p, nil

doInstall: // path is updated, install upt
	err = installUpt()
	if err != nil {
		return "", err
	}
	p, err = exec.LookPath("upt")
	if err != nil {
		return "", err
	}
	return p, nil
}

func EnsureTuitopPath() error {
	v, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot get user home directory")
	}
	os.Setenv("PATH", os.Getenv("PATH")+":"+path.Join(v, ".config/tuitop/bin"))
	return nil
}

// Upt is a Upt for upt
type Upt struct {
	path string
}

func NewUpt() (*Upt, error) {
	p, err := VerifyUpt()
	if err != nil {
		return nil, err
	}
	return &Upt{path: p}, nil
}

// install a package. Steps:
func (s *Upt) Install(name string, createWindow tuiwindow.CreateWindow) error {
	cmd := exec.Command(s.path, "install", name)
	err := cmd.Run()
	if err == nil {
		return nil
	}
	done := make(chan bool)
	var status int
	createWindow("sudo "+s.path+" install "+name, tuiwindow.WithCloseHandler(func(exitStatus int) {
		status = exitStatus
		done <- true
	}))
	<-done
	if status != 0 {
		return fmt.Errorf("error installing %s", name)
	}
	return nil
}

/*
	func (s *Upt) Remove(name string) error {
		cmd := exec.Command(s.path, "remove", name)
		return cmd.Run()
	}

	func (s *Upt) Update(name string) error {
		cmd := exec.Command(s.path, "update", name)
		return cmd.Run()
	}

		func (s *Upt) Info() (string, error) {
		cmd := exec.Command(s.path, "info")
		out, err := cmd.Output()
		if err != nil {
			return "", err
		}
		return string(out), nil
	}
*/
func (s *Upt) Search(name string) (string, error) {
	cmd := exec.Command(s.path, "search", name)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}
