package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/deferpanic/dpcli/api"
	"gopkg.in/alecthomas/kingpin.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	token  string
	hostOS string

	app = kingpin.New("virgo", "Run Unikernels Locally")

	pullCommand     = app.Command("pull", "Pull a project")
	pullCommandName = pullCommand.Arg("name", "Project name.").Required().String()

	runCommand     = app.Command("run", "Run a project")
	runCommandName = runCommand.Arg("name", "Project name.").Required().String()

	killCommand     = app.Command("kill", "Kill a project")
	killCommandName = killCommand.Arg("name", "Project name.").Required().String()

	logCommand     = app.Command("log", "Fetch log of project")
	logCommandName = logCommand.Arg("name", "Project name.").Required().String()

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

func runAsyncCmd(cmd string) {
	command := exec.Command("/bin/bash", "-c", cmd)
	randomBytes := &bytes.Buffer{}
	command.Stdout = randomBytes
	command.Stderr = randomBytes
	// Start command asynchronously
	command.SysProcAttr = &syscall.SysProcAttr{}
	command.SysProcAttr.Setsid = true
	command.Start()

	//out = randomBytes.Bytes()
}

func depcheck() {
	if runtime.GOOS == "darwin" {
		osCheck()
		qemuCheck()
		tuntapCheck()
	}
	if runtime.GOOS == "linux" {
	}
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

	appendLn := "\"{ \\\"net\\\" : { \\\"if\\\":\\\"vioif0\\\",,\\\"type\\\":\\\"inet\\\",, \\\"method\\\":\\\"static\\\",, \\\"addr\\\":\\\"" + ip + "\\\",,  \\\"mask\\\":\\\"24\\\",,  \\\"gw\\\":\\\"10.1.2.3\\\"},, " + blocks + " \\\"cmdline\\\": \\\"" + manifest.Processes[0].Cmdline + "\\\"}\""

	home := os.Getenv("HOME")
	projPath := home + "/.virgo/projects/" + project
	tm := time.Now().Unix()
	pidLn := projPath + "/pids/" + strconv.FormatInt(tm, 10) + ".pid "

	bootLine := ""
	kpath := home + "/.virgo/projects/" + project + "/kernel/" + project
	if manifest.Processes[0].Multiboot {
		bootLine = " -kernel " + kpath + " -append " + appendLn
	} else {
		bootLine = " -hda " + kpath
	}

	cmd := `sudo qemu-system-x86_64 ` +
		drives +
		" -nographic -vga none -serial file:" + projPath + "/logs/blah.log" +
		" -m " + strconv.Itoa(manifest.Processes[0].Memory) +
		"  -net nic,model=virtio,vlan=0,macaddr=00:16:3e:00:01:01 " +
		" -net tap,vlan=0,script=ifup.sh,downscript=ifdown.sh " +
		bootLine +
		" & echo $! >> " + pidLn

	done := make(chan bool)

	go func() {
		runAsyncCmd(cmd)
		done <- true
	}()

	<-done

	if runtime.GOOS == "darwin" {
		fmt.Println(api.GreenBold("setting sysctl"))
		runCmd("sudo sysctl -w net.inet.ip.forwarding=1")
		runCmd("sudo sysctl -w net.link.ether.inet.proxyall=1")
		runCmd("sudo sysctl -w net.inet.ip.fw.enable=1")
	}

	fmt.Println(api.GreenBold("open up http://" + ip + ":3000"))
}

// setToken sets your dprc token for pulling down images from deferpanic
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

// log for now just does a cat of the logs
func log(projectName string) {
	projPath := "~/.virgo/projects/" + projectName

	logz := runCmd("cat " + projPath + "/logs/*")
	fmt.Println(logz)
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

	if len(os.Args) < 2 {
		fmt.Println(logo)
	}
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
	case "log":
		log(*logCommandName)
	}

}
