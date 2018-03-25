package registry

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/deferpanic/virgo/pkg/tools"
)

func expectedRoot() []string {
	return []string{
		"/tmp/.virgo",
		"/tmp/.virgo/projects",
	}
}

func expectedProject(name string) []string {
	return []string{
		"/tmp/.virgo/projects/" + name,
		"/tmp/.virgo/projects/" + name + "/logs",
		"/tmp/.virgo/projects/" + name + "/kernel",
		"/tmp/.virgo/projects/" + name + "/volumes",

		// we do not create this files, so no test for it
		// "/tmp/.virgo/projects/" + name + "/manifest",
		// "/tmp/.virgo/projects/" + name + "/pid.json",
	}
}

func TestRegistryRoot(t *testing.T) {
	r, err := New("/tmp/.virgo")
	if err != nil {
		t.Fatal(err)
	}
	defer r.purge()

	obtained := r.Structure()

	for _, path := range expectedRoot() {
		if !(tools.StringSlice)(obtained).Contains(path) {
			t.Errorf("obtained structure doesn't contain '%s'\n", path)
		}
	}

	for _, path := range obtained {
		if !(tools.StringSlice)(expectedRoot()).Contains(path) {
			t.Errorf("test structure doesn't contain '%s'\n", path)
		}
	}

	if t.Failed() {
		t.Error(obtained)
		t.FailNow()
	}
}

func TestProjectRoot(t *testing.T) {
	projectName := fmt.Sprintf("testing-%d", rand.Intn(65535))

	r, err := New("/tmp/.virgo")
	if err != nil {
		t.Fatal(err)
	}
	defer r.purge()

	r.AddProject(projectName)

	obtained := r.Project(projectName).Structure()

	for _, path := range expectedProject(projectName) {
		if !(tools.StringSlice)(obtained).Contains(path) {
			t.Errorf("obtained structure doesn't contain '%s'\n", path)
		}
	}

	for _, path := range obtained {
		if !(tools.StringSlice)(expectedProject(projectName)).Contains(path) {
			t.Errorf("test structure doesn't contain '%s'\n", path)
		}
	}

	if t.Failed() {
		t.Errorf("%v\n", obtained)
		t.Errorf("%v\n", expectedProject(projectName))
		t.FailNow()
	}
}

func TestRegistryCommunity(t *testing.T) {
	r, err := New("/tmp/.virgo")
	if err != nil {
		t.Fatal(err)
	}
	defer r.purge()

	if err = r.AddProject("project/asdf"); err != nil {
		t.Error(err)
	}

	if err = r.AddProject("project2/"); err == nil {
		t.Error("Expecting error for empty username")
	}

	if err = r.AddProject("project/asdf/adsf"); err == nil {
		t.Error("Expecting error for wrong format")
	}
}
