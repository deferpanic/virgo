package main

import (
	"encoding/json"
	"fmt"
	"github.com/deferpanic/dpcli/api"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	token string
	app   = kingpin.New("virgo", "Run Unikernels Locally")

	pullCommand     = app.Command("pull", "Pull a project")
	pullCommandName = pullCommand.Arg("name", "Project name.").Required().String()

	runCommand     = app.Command("run", "Run a project")
	runCommandName = runCommand.Arg("name", "Project name.").Required().String()

	killCommand     = app.Command("kill", "Kill a project")
	killCommandName = killCommand.Arg("name", "Project name.").Required().String()

	psCommand = app.Command("ps", "List running projects")
)

func runCmd(cmd string) string {
	out, err := exec.Command("/bin/bash", "-c", cmd).CombinedOutput()
	if err != nil {
		fmt.Println(string(out))
		fmt.Println(err)
	}

	return string(out)
}

// osCheck ensures we are dealing with el capitan or above
func osCheck() {
	// osx version
	out := strings.TrimSpace(runCmd("sw_vers -productVersion"))
	if out == "10.11.4" {
		fmt.Println(api.GreenBold("found supported osx version"))
	} else if out == "10.11.5" {
		fmt.Println(api.GreenBold("found supported osx version"))
	} else if out == "10.11.6" {
		fmt.Println(api.GreenBold("found supported osx version"))
	} else {
		fmt.Println(out)
		fmt.Println(api.RedBold("This is only tested on El Capitan - 10.11.4. pf_ctl is used. If using an earlier osx you might need to use natd"))
		os.Exit(1)
	}
}

func qemuCheck() {
	out := strings.TrimSpace(runCmd("which qemu-system-x86_64"))
	if out == "qemu-system-x86_64 not found" {
		fmt.Println(api.RedBold("qemu not found - installing..."))
		runCmd("brew install qemu")
	} else {
		fmt.Println(api.GreenBold("found qemu"))
	}
}

func tuntapCheck() {
	out := strings.TrimSpace(runCmd("sudo kextstat | grep tap"))
	if out != "" {
		fmt.Println(api.GreenBold("found tuntap support"))
	} else {
		fmt.Println(api.RedBold("Please download and install tuntaposx"))
		fmt.Println(api.RedBold("wget http://downloads.sourceforge.net/tuntaposx/tuntap_20150118.tar.gz"))
		os.Exit(1)
	}
}

func depcheck() {
	osCheck()
	qemuCheck()
	tuntapCheck()
}

// createQemuBlocks returns the set of blocks && lines
func createQemuBlocks(project string, manifest api.Manifest) (string, string) {
	blocks := ""
	drives := ""

	home := os.Getenv("HOME")

	// locked down to one process for now
	volz := manifest.Processes[0].Volumes
	for i := 0; i < len(volz); i++ {
		blocks += "\\\"blk\\\" :  { \\\"source\\\":\\\"dev\\\",,  \\\"path\\\":\\\"/dev/ld" +
			strconv.Itoa(i) + "a\\\",, \\\"fstype\\\":\\\"blk\\\",, \\\"mountpoint\\\":\\\"" +
			volz[i].Mount + "\\\"},, "
		drives += " -drive if=virtio,file=" + home + "/.virgo/projects/" + project + "/volumes/vol" + strconv.Itoa(volz[i].Id) + ",format=raw "
	}

	return blocks, drives
}

// run runs the unikernel on osx || linux
// locked down to one instance for now
func run(project string) {

	manifest := readManifest(project)

	blocks, drives := createQemuBlocks(project, manifest)

	ip := "10.1.2.4"

	appendLn := "\"{ \\\"net\\\" : { \\\"if\\\":\\\"vioif0\\\",,\\\"type\\\":\\\"inet\\\",, \\\"method\\\":\\\"static\\\",, \\\"addr\\\":\\\"" + ip + "\\\",,  \\\"mask\\\":\\\"24\\\",,  },, " + blocks + " \\\"cmdline\\\": \\\"" + manifest.Processes[0].Cmdline + "\\\"}"

	home := os.Getenv("HOME")
	projPath := home + "/.virgo/projects/" + project
	tm := time.Now().Unix()
	pidLn := projPath + "/pids/" + strconv.FormatInt(tm, 10) + ".pid"

	cmd := `sudo qemu-system-x86_64 ` +
		drives +
		" -nographic -vga none -serial file:" + projPath + "/logs/blah.log" +
		" -m " + strconv.Itoa(manifest.Processes[0].Memory) +
		"  -net nic,model=virtio,vlan=0,macaddr=00:16:3e:00:01:01 " +
		" -net tap,vlan=0,script=ifup.sh,downscript=ifdown.sh " +
		" -kernel " + home + "/.virgo/projects/" + project + "/kernel/" + project +
		" -append  " + appendLn + "\" & echo $! >> " + pidLn
	go func() {
		runCmd(cmd)
	}()

	fmt.Println(api.GreenBold("setting sysctl"))
	runCmd("sudo sysctl -w net.inet.ip.forwarding=1")
	runCmd("sudo sysctl -w net.link.ether.inet.proxyall=1")
	runCmd("sudo sysctl -w net.inet.ip.fw.enable=1")

	fmt.Println(api.GreenBold("open up http://" + ip + ":3000"))
}

func setToken() {
	dat, err := ioutil.ReadFile(os.Getenv("HOME") + "/.dprc")
	if err != nil {
		fmt.Println(api.RedBold("you can stick your token in ~/.dprc"))
	}
	dtoken := string(dat)

	if dtoken == "" {
		api.RedBold("no token")
		os.Exit(1)
	}

	api.Cli = api.NewCliImplementation(dtoken)
}

// readManifest de-serializes the project manifest
func readManifest(projectName string) api.Manifest {
	mpath := os.Getenv("HOME") + "/.virgo/projects/" + projectName + "/" +
		projectName + ".manifest"

	file, e := ioutil.ReadFile(mpath)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
		os.Exit(1)
	}

	var manifest api.Manifest
	json.Unmarshal(file, &manifest)

	return manifest
}

// setupProjDir sets up the project directory
func setupProjDir(projPath string) {
	// setup directory if not there yet
	runCmd("mkdir -p " + projPath)

	// setup log directory if not there yet
	runCmd("mkdir -p " + projPath + "/logs")

	// setup pid directory if not there yet
	runCmd("mkdir -p " + projPath + "/pids")

	// setup kernel directory if not there yet
	runCmd("mkdir -p " + projPath + "/kernel")

	// setup volumes directory if not there yet
	runCmd("mkdir -p " + projPath + "/volumes")
}

// kill kill's the running project
// right now this assumes that there might be multiple instances of the
// same kind running and to kill them all - not sure why that would be
// the case but eh
func kill(projectName string) {
	projPath := "~/.virgo/projects/" + projectName

	pidstr := runCmd("cat " + projPath + "/pids/*")
	pids := strings.Split(pidstr, "\n")
	for i := 0; i < len(pids)-1; i++ {
		fmt.Println(pids[i])
		runCmd("sudo pkill -P " + pids[i])
	}
	runCmd("rm -rf " + projPath + "/pids/*")
}

// ps lists the running projects
func ps() {
	linez := runCmd("find ~/.virgo/projects/*/pids  -type f")
	nlinez := strings.Split(linez, "\n")
	for i := 0; i < len(nlinez)-1; i++ {
		stuff := strings.Split(nlinez[i], "/")
		fmt.Println(stuff[5])
	}
}

// pull yanks down a unikernel project
// the project contains the kernel, any volumes attached to the project
// and the project manifest
func pull(projectName string) {
	projPath := "~/.virgo/projects/" + projectName

	setupProjDir(projPath)

	// get manifest
	projs := &api.Projects{}
	projs.Manifest(projectName)
	runCmd("mv " + projectName + ".manifest " + projPath)

	// download images
	projs.Download(projectName)
	runCmd("mv " + projectName + " " + projPath + "/kernel/.")

	manifest := readManifest(projectName)

	vols := &api.Volumes{}
	for i := 0; i < len(manifest.Processes); i++ {
		proc := manifest.Processes[i]
		for j := 0; j < len(proc.Volumes); j++ {
			// download volumes
			vols.Download(proc.Volumes[j].Id)
			runCmd("mv vol" + strconv.Itoa(proc.Volumes[j].Id) + " " + projPath + "/volumes/.")
		}
	}

}

func main() {

	b, err := ioutil.ReadFile("dp.txt")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", b)

	setToken()

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "pull":
		pull(*pullCommandName)
	case "run":
		depcheck()
		run(*runCommandName)
	case "ps":
		ps()
	case "kill":
		kill(*killCommandName)
	}

}
