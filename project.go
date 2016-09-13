package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Project is a representation of a unikernel project
// it contains a set of {processes, volumes, manifest}
type Project struct {
	Name      string
	Community bool
}

var projRoot = os.Getenv("HOME") + "/.virgo/projects/"

// isRoot determines if we are in the root of a project
func isRoot(fname string) bool {
	if strings.Contains(fname, ".manifest") {
		return true
	}
	return false
}

// projList recursively lists project names
func projList(path string) []string {
	var projs = []string{}

	files, _ := ioutil.ReadDir(path)

	for _, f := range files {
		if isRoot(f.Name()) {
			fproj := strings.Replace(path, projRoot+"/", "", -1)
			return []string{fproj}
		}
	}

	for _, f := range files {
		if f.IsDir() {
			s := projList(path + "/" + f.Name())
			projs = append(projs, s...)
		}
	}

	return projs
}

// images lists all the projects available
func images() {
	projs := projList(projRoot)
	for i := 0; i < len(projs); i++ {
		fmt.Println(projs[i])
	}
}
