package tools

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/deferpanic/dpcli/api"
)

// custom strings concatenator to avoid separator artefacts of empty params
func Join(a []string, sep string) string {
	result := make([]byte, 0)

	for i, _ := range a {
		if len(a[i]) == 0 {
			continue
		}

		if len(result) > 0 && a[i] != "" {
			result = append(result, []byte(sep)...)
		}

		result = append(result, []byte(a[i])...)
	}

	return string(result)
}

type Slice interface {
	Contains(string) bool
}

type StringSlice []string

func (ss StringSlice) Contains(s string) bool {
	for i, _ := range ss {
		if ss[i] == s {
			return true
		}
	}

	return false
}

func SetToken() error {
	f := os.Getenv("HOME") + "/.dprc"

	dat, err := ioutil.ReadFile(f)
	if err != nil {
		return fmt.Errorf("error reading file '%s' - %s", f, err)
	}

	dtoken := string(dat)

	if dtoken == "" {
		return fmt.Errorf("error reading token - no token found")
	}

	dtoken = strings.TrimSpace(dtoken)
	api.Cli = api.NewCliImplementation(dtoken)

	return nil
}

func ShowFiles(dir string) error {
	fd, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 1, 8, 2, '\t', 0)

	for _, f := range fd {
		if f.IsDir() {
			continue
		}

		fmt.Fprintf(w, "%s\t%d\t%s\n", f.Name(), f.Size(), f.ModTime().String())
	}

	err = w.Flush()
	if err != nil {
		return fmt.Errorf("error flushing log output '%s'", err)
	}

	return nil
}
