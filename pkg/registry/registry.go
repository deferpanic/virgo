package registry

import (
	"fmt"
	"io/ioutil"
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
	cfgRuntimeFile  = "runtime.json"
	cfgIfUpFile     = "ifup.sh"
	cfgIfDownFile   = "ifdown.sh"
)

type Project struct {
	name     string
	username string
	root     string
}

type Registry struct {
	root     string
	projects []Project
}

// "v ...string" is optional argument, for non-default registry root
func New(v ...string) (r *Registry, err error) {
	r = &Registry{
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

func (r *Registry) AddProject(name string) (Project, error) {
	p := Project{name: name, root: r.root}

	if strings.Contains(name, "/") {
		if parts := strings.Split(name, "/"); len(parts) != 2 {
			return Project{}, fmt.Errorf("wrong format for community project, should be project/username")
		} else {
			name = parts[0]
			if parts[1] == "" {
				return Project{}, fmt.Errorf("username can't be empty for community projects")
			}
			p.username = parts[1]
		}
	}

	// nothing to initialize for empty project
	if name == "" {
		return Project{}, fmt.Errorf("empty project name, unable to proceed")
	}

	p.name = name
	r.projects = append(r.projects, p)

	for _, dir := range r.Project(name).Structure() {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return Project{}, fmt.Errorf("error creating registry - %s", err)
		}
	}

	return p, nil
}

func (r *Registry) Project(name string) Project {
	for _, project := range r.projects {
		if project.name == name {
			return project
		}
	}

	return Project{}
}

func (r *Registry) ProjectList() []Project {
	return r.projects
}

func (r *Registry) initialize() error {
	if _, err := os.Stat(r.Root()); err != nil {
		if os.IsNotExist(err) {
			for _, dir := range r.Structure() {
				if err := os.MkdirAll(dir, 0755); err != nil {
					return fmt.Errorf("error creating registry - %s", err)
				}
			}
			return nil
		} else if os.IsExist(err) {
			return nil
		} else {
			return fmt.Errorf("error initializing registry - %s", err)
		}
	}

	flist, err := ioutil.ReadDir(r.Projects())
	if err != nil {
		return fmt.Errorf("error reading project directory - %s", err)
	}

	for _, f := range flist {
		if f.IsDir() {
			if _, err := r.AddProject(f.Name()); err != nil {
				return fmt.Errorf("error adding project - %s", err)
			}
		}
	}

	return nil
}

func (r Registry) purge() error {
	return os.RemoveAll(r.Root())
}

func (r Registry) PurgeProject(name string) error {
	return os.RemoveAll(r.Project(name).Root())
}

func (r Registry) Root() string {
	return r.root
}

func (r Registry) Projects() string {
	return filepath.Join(r.root, cfgProjectsDir)
}

func (r Registry) RuntimeFile() string {
	return filepath.Join(r.root, cfgRuntimeFile)
}

func (r Registry) Structure() []string {
	return []string{
		r.Root(),
		r.Projects(),
	}
}

func (p Project) Root() string {
	return filepath.Join(p.root, cfgProjectsDir, p.name)
}

func (p Project) Name() string {
	return p.name
}

func (p Project) LogsDir() string {
	return filepath.Join(p.Root(), cfgLogsDir)
}

func (p Project) PidFile() string {
	return filepath.Join(p.Root(), cfgPidFile)
}

func (p Project) KernelDir() string {
	return filepath.Join(p.Root(), cfgKernelDir)
}

func (p Project) KernelFile() string {
	return filepath.Join(p.Root(), cfgKernelDir, p.name)
}

func (p Project) VolumesDir() string {
	return filepath.Join(p.Root(), cfgVolumesDir)
}

func (p Project) ManifestFile() string {
	return filepath.Join(p.Root(), cfgManifestFile)
}

func (p Project) IfUpFile() string {
	return filepath.Join(p.Root(), cfgIfUpFile)
}

func (p Project) IfDownFile() string {
	return filepath.Join(p.Root(), cfgIfDownFile)
}

func (p Project) IsCommunity() bool {
	return p.username != ""
}

func (p Project) UserName() string {
	return p.username
}

func (p Project) Structure() []string {
	return []string{
		p.Root(),
		p.LogsDir(),
		p.KernelDir(),
		p.VolumesDir(),
	}
}
