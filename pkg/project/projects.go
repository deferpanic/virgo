package project

import (
	"io/ioutil"
	"log"

	"github.com/deferpanic/dpcli/api"
	"github.com/deferpanic/virgo/pkg/registry"
)

type Projects []Project

func LoadProjects(r registry.Registry) (Projects, error) {
	result := make(Projects, 0)

	fs, err := ioutil.ReadDir(r.Projects())
	if err != nil {
		return nil, err
	}

	for _, fd := range fs {
		if fd.IsDir() {
			r, err := registry.New(fd.Name())
			if err != nil {
				log.Printf("Error initialising registry for project '%s' - %s\n", fd.Name(), err)
			}

			p := New(r)
			if err := p.Load(); err != nil {
				log.Printf("Error loading project '%s' - %s\n", fd.Name(), err)
				continue
			}

			result = append(result, p)
		}
	}

	return result, nil
}
