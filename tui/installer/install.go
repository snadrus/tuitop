package installer

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/snadrus/tuitop/tui/tuiwindow"
	"golang.org/x/xerrors"
	"gopkg.in/yaml.v3"
)

type Installer struct {
	createWindow tuiwindow.CreateWindow
	*Upt
}

func New(createWindow tuiwindow.CreateWindow) *Installer {
	err := EnsureTuitopPath()
	if err != nil {
		fmt.Println(err)
	}
	i := &Installer{
		createWindow: createWindow,
	}

	return i
}

type InstallerYaml struct {
	Name     string      `yaml:"name"`
	Category string      `yaml:"category"`
	CLI      string      `yaml:"cli"`
	Get      AvailableAt `yaml:"get"`
	Alt      altStruct   `yaml:"alt"`
}
type AvailableAt struct {
	Source             string        `yaml:"source"`
	SourceInstructions string        `yaml:"sourceInstructions"`
	Pkg                string        `yaml:"pkg"`
	Either             []AvailableAt `yaml:"either"`
}
type altStruct struct {
	Run string      `yaml:"run"`
	Get AvailableAt `yaml:"get"`
}

// Ensure installs a program from its yaml file definition.
// First detecting it, then pkg, then source, then alt (similarly).
// Returns the path and error.
func (i *Installer) Ensure(yamlReader io.Reader) (path string, err error) {
	// read yaml
	y := InstallerYaml{}
	if err := yaml.NewDecoder(yamlReader).Decode(&y); err != nil {
		return "", xerrors.Errorf("cannot unmarshal yaml: %w", err)
	}
	// detect
	if y.CLI == "" {
		// FUTURE: allow libraries?
		return "", xerrors.Errorf("no CLI defined for %s", y.Name)
	}

	p, err := exec.LookPath(strings.Split(y.CLI, " ")[0])
	if err == nil {
		return p, nil
	}
	var errs error
	for _, availAt := range []AvailableAt{y.Get, y.Alt.Get} {
		p, err = i.installPkg(availAt, y.CLI)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		if p != "" {
			return p, nil
		}
	}
	return "", xerrors.Errorf("cannot install %s: %w", y.Name, errs)
}

// installPkg installs a package, errors are informational. Returns the path or "" if not found.
func (i *Installer) installPkg(get AvailableAt, cli string) (string, error) {
	var errs []error
	for _, a := range get.Either {
		path, err := i.installPkg(a, cli)
		if err == nil {
			return path, nil
		}
		// keep trying if errors.
		errs = append(errs, err)
	}
	if get.Pkg == "" {
		return "", nil
	}
	if i.Upt == nil {
		var err error
		i.Upt, err = NewUpt()
		if err != nil {
			return "", err
		}
		errs = append(errs, err)
	}
	_, err := i.Upt.Search(get.Pkg)
	if err != nil {
		return "", nil // not found
	}
	// Here the package exists.

	errs = append(errs, err)
	err = i.Upt.Install(get.Pkg, i.createWindow)
	if err != nil {
		errs = append(errs, err)
		return "", errors.Join(errs...)
	}

	return exec.LookPath(cli)
}
