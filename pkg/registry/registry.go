package registry

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	cfgDefaultRoot = ".virgo"
	cfgProjectsDir = "projects"
	cfgKernelDir   = "kernel"
	cfgLogsDir     = "logs"
	cfgPidsDir     = "pids"
	cfgVolumesDir  = "volumes"
)

type registry struct {
	name string
	root string
}

// "v ...string" is optional argument, for non-default registry root
func New(name string, v ...string) (r registry, err error) {
	r = registry{
		name: name,
		root: filepath.Join(os.Getenv("HOME"), cfgDefaultRoot),
	}

	if len(v) == 1 {
		r.root = v[0]
	}

	if err = r.initialize(); err != nil {
		return
	}

	return
}

func (r registry) initialize() error {
	if _, err := os.Stat(r.Root()); err != nil {
		if os.IsNotExist(err) {
		} else if os.IsExist(err) {
			fmt.Printf("exists\n")
			return nil
		} else {
			return fmt.Errorf("error initializing registry - %s", err)
		}
	}

	for _, dir := range r.Structure() {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("error creating registry - %s", err)
		}
	}

	return nil
}

func (r registry) purge() error {
	return os.RemoveAll(r.Root())
}

func (r registry) PurgeProject() error {
	return os.RemoveAll(r.Project())
}

func (r registry) Root() string {
	return r.root
}

func (r registry) Projects() string {
	return filepath.Join(r.Root(), cfgProjectsDir)
}

func (r registry) Project() string {
	return filepath.Join(r.Root(), cfgProjectsDir, r.name)
}

func (r registry) LogsDir() string {
	return filepath.Join(r.Project(), cfgLogsDir)
}

func (r registry) PidsDir() string {
	return filepath.Join(r.Project(), cfgPidsDir)
}

func (r registry) KernelDir() string {
	return filepath.Join(r.Project(), cfgKernelDir)
}

func (r registry) VolumesDir() string {
	return filepath.Join(r.Project(), cfgVolumesDir)
}

func (r registry) Structure() []string {
	return []string{
		r.Root(),
		r.Projects(),
		r.Project(),
		r.LogsDir(),
		r.PidsDir(),
		r.KernelDir(),
		r.VolumesDir(),
	}
}
