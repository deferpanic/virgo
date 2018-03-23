package project

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/deferpanic/dpcli/api"
	"github.com/deferpanic/virgo/pkg/registry"
	"github.com/deferpanic/virgo/pkg/runner"
)

type Project struct {
	registry.Registry
	manifest api.Manifest
	process  runner.Runner
}

func New(r registry.Registry) *Project {
	return &Project{registry.Registry: r}
}

func (p *Project) Load() error {
	var err error

	r, err := os.OpenFile(p.ManifestFile(), os.O_RDONLY, 0644)
	if err != nil {
		r.Close()
		return err
	}

	if err := json.Unmarshal(r, &p.manifest); err != nil {
		r.Close()
		return fmt.Errorf("unable to load manifest file - %s", err)
	}

	r.Close()

	r, err := os.OpenFile(p.PidFile(), os.O_RDONLY, 0644)
	if err != nil {
		r.Close()
		return err
	}

	if err := json.Unmarshal(r, &p.process); err != nil {
		r.Close()
		return fmt.Errorf("unable to load manifest file - %s", err)
	}

	r.Close()

	return nil
}

func (p *Project) Pull() error {
	var err error
	log.Printf("Pulling project name: %s\n", p.ProjectName())

	p.manifest, err = (*api.Projects).LoadManifest(p.ProjectName())
	if err != nil {
		return err
	}

	// @TODO code below is old - refactor it
	//
	// the only difference here is an image and source path
	if p.IsCommunity() {
		err = (*api.Projects).DownloadCommunity(p.ProjectName(), p.UserName(), p.KernelFile())
	} else {
		err = (*api.Projects).Download(p.ProjectName(), p.KernelFile())
	}

	if err != nil {
		return err
	} else {
		// @TODO add verbosity levels
		log.Println(api.GreenBold("kernel file saved"))
	}

	for i := 0; i < len(p.manifest.Processes); i++ {
		proc := p.manifest.Processes[i]
		for _, volume := range proc.Volumes {
			dst := filepath.Join(p.VolumesDir(), "vol"+strconv.Itoa(volume.Id))

			if err = (*api.Volume).Download(volume.Id, dst); err != nil {
				return err
			} else {
				// @TODO add verbosity levels
				log.Println(api.GreenBold(dst + " file saved"))
			}
		}
	}

	b, err := json.Marshal(p.manifest)
	if err != nil {
		return err
	}

	wr, err := os.OpenFile(p.ManifestFile(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	if _, err := wr.Write(b); err != nil {
		return err
	}

	return nil
}

func (p *Project) Run() error {
	var (
		err      error
		env      string
		kflag    string
		bootline []string
	)

	if len(p.manifest.Processes) == 0 {
		return fmt.Errorf("no processes found in manifest file, unable to proceed")
	}

	blocks, drives := p.createQemuBlocks()

	if p.manifest.Processes[0].Env != "" {
		env = p.formatEnv(p.manifest.Processes[0].Env)
	}

	ip, gw := pkg.GetNetwork(p.Project())
	pkg.SetNetwork(p.Project(), ip, gw)

	appendLn := "\"{ \\\"net\\\" : { \\\"if\\\":\\\"vioif0\\\",,\\\"type\\\":\\\"inet\\\",, \\\"method\\\":\\\"static\\\",, \\\"addr\\\":\\\"" + ip + "\\\",,  \\\"mask\\\":\\\"24\\\",,  \\\"gw\\\":\\\"" + gw + "\\\"},, " + env + blocks + " \\\"cmdline\\\": \\\"" + manifest.Processes[0].Cmdline + "\\\"}\""

	if manifest.Processes[0].Multiboot {
		bootLine = []string{"-kernel", p.KernelFile(), "-append", appendLn}
	} else {
		bootLine = []string{"-hda", p.KernelFile()}
	}

	kflag := "-no-kvm"
	if runtime.GOOS == "linux" {
		if kvmEnabled() {
			kflag = "-enable-kvm"
		}
	}

	if runtime.GOOS == "darwin" {
		if pkg.CheckHAX() {
			log.Println(api.GreenBold("hax is enabled!"))
			kflag = "-accel hax"
		}
	}

	pkg.SetupNetwork(p.Project(), gw)

	mac := pkg.GenerateMAC()

}

func (p *Project) formatEnv(env string) (result string) {
	parts := strings.Split(env, " ")

	for i, _ := range parts {
		result += "\\\"env\\\": \\\"" + parts[i] + "\\\",,"
	}

	return result
}

// locked down to one process for now
//
func (p *Project) createQemuBlocks(project string, manifest api.Manifest) (string, string) {
	blocks := ""
	drives := ""

	if len(manifest.Processes) == 0 {
		return blocks, drives
	}

	for i, volume := range manifest.Processes[0].Volumes {
		blocks += "\\\"blk\\\" :  { \\\"source\\\":\\\"dev\\\",,  \\\"path\\\":\\\"/dev/ld" +
			strconv.Itoa(i) + "a\\\",, \\\"fstype\\\":\\\"blk\\\",, \\\"mountpoint\\\":\\\"" +
			volume.Mount + "\\\"},, "
		drives += " -drive if=virtio,file=" + p.VolumesDir() + "/vol" + strconv.Itoa(volume[i].Id) + ",format=raw "
	}

	return blocks, drives
}
