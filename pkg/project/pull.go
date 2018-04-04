package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/deferpanic/dpcli/api"
	"github.com/deferpanic/virgo/pkg/registry"
)

func Pull(pr registry.Project) error {
	var (
		err      error
		manifest api.Manifest
	)

	if manifest, err = api.LoadManifest(pr.Name()); err != nil {
		return err
	}

	ap := &api.Projects{}
	if pr.IsCommunity() {
		parts := strings.Split(pr.Name(), "/")
		err = ap.DownloadCommunity(parts[1], pr.UserName(), pr.KernelFile())
	} else {
		err = ap.Download(pr.Name(), pr.KernelFile())
	}
	if err != nil {
		return err
	}

	v := &api.Volumes{}

	for i := 0; i < len(manifest.Processes); i++ {
		proc := manifest.Processes[i]
		for _, volume := range proc.Volumes {
			dst := filepath.Join(pr.VolumesDir(), "vol"+strconv.Itoa(volume.Id))

			if err = v.Download(volume.Id, dst); err != nil {
				return err
			}
		}
	}

	b, err := json.Marshal(manifest)
	if err != nil {
		return err
	}

	wr, err := os.OpenFile(pr.ManifestFile(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if _, err := wr.Write(b); err != nil {
		return err
	}

	return nil
}
