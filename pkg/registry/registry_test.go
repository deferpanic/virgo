package registry

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/deferpanic/virgo/pkg/tools"
)

func expected(name string) []string {
	return []string{
		"/tmp/.virgo",
		"/tmp/.virgo/projects",
		"/tmp/.virgo/projects/" + name,
		"/tmp/.virgo/projects/" + name + "/logs",
		// "/tmp/.virgo/projects/" + name + "/pids",
		"/tmp/.virgo/projects/" + name + "/kernel",
		"/tmp/.virgo/projects/" + name + "/volumes",

		// we do not create this file, so no test for it
		// "/tmp/.virgo/projects/" + name + "/manifest",
	}

}

func TestRegistryStructure(t *testing.T) {
	projectName := fmt.Sprintf("testing-%d", rand.Intn(65535))

	r, err := New(projectName, "/tmp/.virgo")
	if err != nil {
		t.Fatal(err)
	}
	defer r.purge()

	obtained := r.Structure()

	for _, path := range expected(projectName) {
		if !(tools.StringSlice)(obtained).Contains(path) {
			t.Errorf("obtained structure doesn't contain '%s'\n", path)
		}
	}

	for _, path := range obtained {
		if !(tools.StringSlice)(expected(projectName)).Contains(path) {
			t.Errorf("test structure doesn't contain '%s'\n", path)
		}
	}

	if t.Failed() {
		t.FailNow()
	}
}

func TestRegistryCommunity(t *testing.T) {
	r, err := New("project/username")
	if err != nil {
		t.Error(err)
	}
	r.purge()

	r, err = New("project/")
	if err == nil {
		t.Error("Expecting error for empty username")
	}
	r.purge()

	r, err = New("project/asdf/adsf")
	if err == nil {
		t.Error("Expecting error for wrong format")
	}
	r.purge()
}
