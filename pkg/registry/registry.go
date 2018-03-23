package registry

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	cfgDefaultRoot  = ".virgo"
	cfgProjectsDir  = "projects"
	cfgKernelDir    = "kernel"
	cfgLogsDir      = "logs"
	cfgPidsDir      = "pids"
	cfgVolumesDir   = "volumes"
	cfgManifestFile = "manifest"
	cfgPidFile      = "pid.json"
)

type Registry struct {
	name     string
	root     string
	username string
}

// "v ...string" is optional argument, for non-default registry root
func New(name string, v ...string) (r Registry, err error) {
	if strings.Contains(name, "/") {
		if parts := strings.Split(name, "/"); len(parts) != 2 {
			return Registry{}, fmt.Errorf("wrong format for community project, should be project/username")
		} else {
			name = parts[0]
			if parts[1] == "" {
				return Registry{}, fmt.Errorf("username can't be empty for community projects")
			}
			r.username = parts[1]
		}
	}

	r = Registry{
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

func (r Registry) initialize() error {
	if _, err := os.Stat(r.Root()); err != nil {
		if os.IsNotExist(err) {
		} else if os.IsExist(err) {
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

func (r Registry) purge() error {
	return os.RemoveAll(r.Root())
}

func (r Registry) PurgeProject() error {
	return os.RemoveAll(r.Project())
}

func (r Registry) Root() string {
	return r.root
}

func (r Registry) ProjectName() string {
	return r.name
}

func (r Registry) Projects() string {
	return filepath.Join(r.Root(), cfgProjectsDir)
}

func (r Registry) Project() string {
	return filepath.Join(r.Root(), cfgProjectsDir, r.name)
}

func (r Registry) LogsDir() string {
	return filepath.Join(r.Project(), cfgLogsDir)
}

// func (r Registry) PidsDir() string {
// 	return filepath.Join(r.Project(), cfgPidsDir)
// }

func (r Registry) PidFile() string {
	return filepath.Join(r.Project(), cfgPidFile)
}

func (r Registry) KernelDir() string {
	return filepath.Join(r.Project(), cfgKernelDir)
}

func (r Registry) KernelFile() string {
	return filepath.Join(r.KernelDir(), r.ProjectName())
}

func (r Registry) VolumesDir() string {
	return filepath.Join(r.Project(), cfgVolumesDir)
}

func (r Registry) ManifestFile() string {
	return filepath.Join(r.Project(), cfgManifestFile)
}

func (r Registry) IsCommunity() bool {
	return r.username != ""
}

func (r Registry) UserName() string {
	return r.username
}

func (r Registry) Structure() []string {
	return []string{
		r.Root(),
		r.Projects(),
		r.Project(),
		r.LogsDir(),
		// r.PidsDir(),
		r.KernelDir(),
		r.VolumesDir(),
	}
}
