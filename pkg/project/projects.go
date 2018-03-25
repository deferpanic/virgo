package project

import (
	"log"

	"github.com/deferpanic/virgo/pkg/registry"
)

type Projects []*Project

func LoadProjects(r *registry.Registry) (Projects, error) {
	result := make(Projects, 0)

	for _, rp := range r.ProjectList() {
		p := New(rp)
		if err := p.Load(); err != nil {
			log.Printf("Error loading project '%s' - %s\n", rp.Name(), err)
		}

		result = append(result, p)
	}

	return result, nil
}

func (ps Projects) GetProjectByName(name string) *Project {
	for _, project := range ps {
		if project.Name() == name {
			return project
		}
	}

	return nil
}

func (ps Projects) Running() Projects {
	result := make(Projects, 0)

	for _, project := range ps {
		if project.process.IsAlive() {
			result = append(result, project)
		}
	}

	return result
}
