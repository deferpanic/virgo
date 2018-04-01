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

	if name == "" {
		return Project{}, fmt.Errorf("empty project name, unable to proceed")
	}

	if strings.Contains(name, "/") {
		if parts := strings.Split(name, "/"); len(parts) != 2 {
			return Project{}, fmt.Errorf("wrong format for community project, should be project/username")
		} else {
			if parts[1] == "" {
				return Project{}, fmt.Errorf("username can't be empty for community projects")
			}

			p.username = parts[0]
		}
	}

	p.name = name
	r.projects = append(r.projects, p)

	projectroot := filepath.Join(r.Projects(), name)

	if _, err := os.Stat(projectroot); err == nil {
		return p, nil
	}

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
		} else {
			return fmt.Errorf("error initializing registry - %s", err)
		}
	}

	var loadProjects func(root string) error

	loadProjects = func(root string) error {
		list, err := ioutil.ReadDir(root)
		if err != nil {
			return err
		}

		for _, info := range list {
			if !info.IsDir() {
				continue
			}

			manifest := filepath.Join(root, info.Name(), info.Name()+"."+cfgManifestFile)

			if _, err := os.Stat(manifest); err == nil {
				projectName := info.Name()

				if root != r.Projects() {
					projectName = filepath.Join(filepath.Base(root), info.Name())
				}

				if _, err := r.AddProject(projectName); err != nil {
					return err
				}
			} else if os.IsNotExist(err) {
				loadProjects(filepath.Join(root, info.Name()))
			} else {
				fmt.Println(err)
			}
		}

		return nil
	}

	return loadProjects(r.Projects())
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

// Returns project root, e.g.: ~/.virgo/projects/hello
// For community projects root is nested in username/projects directory.
func (p Project) Root() string {
	return filepath.Join(p.root, cfgProjectsDir, p.name)
}

func (p Project) Name() string {
	return p.name
}

func (p Project) LogsDir() string {
	return filepath.Join(p.Root(), cfgLogsDir)
}

func (p Project) KernelDir() string {
	return filepath.Join(p.Root(), cfgKernelDir)
}

func (p Project) KernelFile() string {
	file := filepath.Join(p.Root(), cfgKernelDir, p.Name())

	if p.IsCommunity() {
		name := strings.Replace(p.Name(), "/", "_", -1)
		file = filepath.Join(p.Root(), cfgKernelDir, name)
	}

	return file
}

func (p Project) VolumesDir() string {
	return filepath.Join(p.Root(), cfgVolumesDir)
}

func (p Project) ManifestFile() string {
	file := filepath.Join(p.Root(), p.Name()+"."+cfgManifestFile)

	if p.IsCommunity() {
		parts := strings.Split(p.Name(), "/")
		file = filepath.Join(p.Root(), parts[1]+"."+cfgManifestFile)
	}

	return file
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
