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

	killCommand     = app.Command("kill", "Kill a running project")
	killCommandName = killCommand.Arg("name", "Project name.").Required().String()

	rmCommand     = app.Command("rm", "Remove a project")
	rmCommandName = rmCommand.Arg("name", "Project name.").Required().String()

	logCommand     = app.Command("log", "Fetch log of project")
	logCommandName = logCommand.Arg("name", "Project name.").Required().String()

	searchCommand      = app.Command("search", "Search for a project")
	searchCommandName  = searchCommand.Arg("description", "Description").Required().String()
	searchCommandStars = searchCommand.Arg("stars", "Star Count").Int()

	signupCommand  = app.Command("signup", "Signup")
	signupEmail    = signupCommand.Arg("email", "Email.").Required().String()
	signupUsername = signupCommand.Arg("username", "Username.").Required().String()
	signupPassword = signupCommand.Arg("password", "Password.").Required().String()

	psCommand = app.Command("ps", "List running projects")

	imagesCommand = app.Command("images", "List all projects")
)

// runCmd runs a shell command and returns any stderr/stdout
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
	command.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	command.Start()
}

// depcheck does a quick dependency check to ensure
// that all required deps are installed - some auto-install
func depcheck() {
	if runtime.GOOS == "darwin" {
		osCheck()
		qemuCheck()
		cpulimitCheck()
		tuntapCheck()
	}
	if runtime.GOOS == "linux" {
	}
}

// createQemuBlocks returns the set of blocks && lines
func createQemuBlocks(project string, manifest api.Manifest) (string, string) {
	blocks := ""
	drives := ""

	// locked down to one process for now
	volz := manifest.Processes[0].Volumes
	for i := 0; i < len(volz); i++ {
		blocks += "\\\"blk\\\" :  { \\\"source\\\":\\\"dev\\\",,  \\\"path\\\":\\\"/dev/ld" +
			strconv.Itoa(i) + "a\\\",, \\\"fstype\\\":\\\"blk\\\",, \\\"mountpoint\\\":\\\"" +
			volz[i].Mount + "\\\"},, "
		drives += " -drive if=virtio,file=" + projRoot + project + "/volumes/vol" + strconv.Itoa(volz[i].Id) + ",format=raw "
	}

	return blocks, drives
}

// formatEnvs returns the env variables in the format expected for
// rumpkernels
func formatEnvs(menv string) string {
	env := ""
	envs := strings.Split(menv, " ")
	for i := 0; i < len(envs); i++ {
		env += "\\\"env\\\": \\\"" + envs[i] + "\\\",,"
	}

	return env
}

// kvmEnabled returns true if kvm is available
func kvmEnabled() bool {
	cmd := "egrep '(vmx|svm)' /proc/cpuinfo"
	out := strings.TrimSpace(runCmd(cmd))
	if out == "" {
		return false
	} else {
		return true
	}
}

// run runs the unikernel on osx || linux
// locked down to one instance for now
func run(project string) {

	manifest := readManifest(project)

	blocks, drives := createQemuBlocks(project, manifest)

	env := ""
	if manifest.Processes[0].Env != "" {
		env = formatEnvs(manifest.Processes[0].Env)
	}

	projPath := projRoot + project

	ip, gw := getNetwork(projPath)
	setNetwork(projPath, ip, gw)

	appendLn := "\"{ \\\"net\\\" : { \\\"if\\\":\\\"vioif0\\\",,\\\"type\\\":\\\"inet\\\",, \\\"method\\\":\\\"static\\\",, \\\"addr\\\":\\\"" + ip + "\\\",,  \\\"mask\\\":\\\"24\\\",,  \\\"gw\\\":\\\"" + gw + "\\\"},, " + env + blocks + " \\\"cmdline\\\": \\\"" + manifest.Processes[0].Cmdline + "\\\"}\""

	tm := time.Now().Unix()
	pidLn := projPath + "/pids/" + strconv.FormatInt(tm, 10) + ".pid "

	kpath := projRoot + project + "/kernel/"
	if strings.Contains(project, "/") {
		s := strings.Split(project, "/")[1]
		kpath += s
	} else {
		kpath += project
	}

	bootLine := ""

	if manifest.Processes[0].Multiboot {
		bootLine = " -kernel " + kpath + " -append " + appendLn
	} else {
		bootLine = " -hda " + kpath
	}

	kflag := "-no-kvm"
	if runtime.GOOS == "linux" {
		if kvmEnabled() {
			kflag = "-enable-kvm"
		}
	}

	setupNetwork(projPath, gw)

	mac := generateMAC()

	runLan := strconv.Itoa(len(running()) + 1)

	networkLine := "  -net nic,model=virtio,vlan=" + runLan + ",macaddr=" +
		mac + " " + " -net tap,vlan=" + runLan + ",ifname=tap" + runLan + ",script=" + projPath +
		"/ifup.sh,downscript=" + projPath + "/ifdown.sh "

	cmd := "sudo qemu-system-x86_64 " + kflag + drives +
		" -nographic -vga none -serial file:" + projPath + "/logs/blah.log" +
		" -m " + strconv.Itoa(manifest.Processes[0].Memory) +
		networkLine + bootLine +
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

		// enable this for lower osx versions
		o := osCheck()
		if (o != "10.12") && (o != "10.12.2") {
			runCmd("sudo sysctl -w net.inet.ip.fw.enable=1")
		}
	}

	fmt.Println(api.GreenBold("open up http://" + ip + ":3000"))
}

// setToken sets your dprc token for pulling down images from deferpanic
func setToken() {
	dat, err := ioutil.ReadFile(os.Getenv("HOME") + "/.dprc")
	if err != nil {
		fmt.Println(api.RedBold("Have an account yet?\n" +
			"If so you can stick your token in ~/.dprc.\n" +
			"Otherwise signup via:\n\n\tvirgo signup my@email.com username password\n"))
	}
	dtoken := string(dat)

	if dtoken == "" {
		api.RedBold("no token")
		os.Exit(1)
	}

	dtoken = strings.TrimSpace(dtoken)
	api.Cli = api.NewCliImplementation(dtoken)
}

// readManifest de-serializes the project manifest
func readManifest(projectName string) api.Manifest {
	pName := projectName
	if strings.Contains(pName, "/") {
		pName = strings.Split(pName, "/")[1]
	}

	mpath := projRoot + projectName + "/" +
		pName + ".manifest"

	if _, err := os.Stat(mpath); os.IsNotExist(err) {
		fmt.Println(api.RedBold("can't find " + projectName + " manifest - does it exist?"))
		os.Exit(1)
	}

	file, e := ioutil.ReadFile(mpath)
	if e != nil {
		fmt.Println(api.RedBold("Missing Manifest for " + projectName))
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

// rm removes a project locally
func rm(projectName string) {
	os.RemoveAll(projRoot + projectName)
}

// kill kill's the running project
// right now this assumes that there might be multiple instances of the
// same kind running and to kill them all - not sure why that would be
// the case but eh
func kill(projectName string) {
	projPath := projRoot + projectName

	if _, err := os.Stat(projPath); os.IsNotExist(err) {
		fmt.Println(api.RedBold("can't find " + projectName + " - does it exist?"))
		os.Exit(1)
	}

	pidstr := runCmd("cat " + projPath + "/pids/*")
	pids := strings.Split(pidstr, "\n")
	for i := 0; i < len(pids)-1; i++ {
		runCmd("sudo pkill -P " + pids[i])
	}
	runCmd("rm -rf " + projPath + "/pids/*")
	runCmd("rm -rf " + projPath + "/*.sh")
	runCmd("rm -rf " + projPath + "/net")
}

// log for now just does a cat of the logs
func log(projectName string) {
	projPath := projRoot + projectName

	logz := runCmd("cat " + projPath + "/logs/*")
	fmt.Println(logz)
}

// running builds a list of running projects
func running() []string {
	projs := projList(projRoot)

	running := []string{}
	for i := 0; i < len(projs); i++ {
		ppath := projRoot + projs[i] + "/pids"
		files, _ := ioutil.ReadDir(ppath)

		for x := 0; x < len(files); x++ {
			running = append(running, projs[i])
		}
	}

	return running
}

// ps lists the running projects
func ps() {
	pids := running()
	for x := 0; x < len(pids); x++ {
		fmt.Println(pids[x])
	}
}

// pull yanks down a unikernel project
// the project contains the kernel, any volumes attached to the project
// and the project manifest
func pull(projectName string) {
	community := false
	projUser := ""
	projName := projectName
	if strings.Contains(projectName, "/") {
		community = true
		s := strings.Split(projectName, "/")
		projUser = s[0]
		projName = s[1]
	}

	projPath := projRoot + projectName

	setupProjDir(projPath)

	// get manifest
	projs := &api.Projects{}
	err := projs.Manifest(projectName)
	if err != nil {
		fmt.Println(api.RedBold(err.Error()))
		os.Exit(1)
	}

	name := strings.Replace(projectName, "/", "_", -1)
	runCmd("mv " + name + ".manifest " + projPath + "/" + projName + ".manifest")

	// download kernel
	if community {
		projs.DownloadCommunity(projName, projUser, projPath+"/kernel/"+projName)
	} else {
		projs.Download(projName, projPath+"/kernel/"+projName)
	}

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

	if len(os.Args) > 1 && os.Args[1] == "signup" {
		api.Cli = api.NewCliImplementation("")
	} else {
		setToken()
	}

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	case "pull":
		pull(*pullCommandName)
	case "run":
		depcheck()
		run(*runCommandName)
	case "ps":
		ps()
	case "images":
		images()
	case "kill":
		kill(*killCommandName)
	case "rm":
		rm(*rmCommandName)
	case "log":
		log(*logCommandName)
	case "search":
		search := &api.Search{}
		if *searchCommandStars != 0 {
			search.FindWithStars(*searchCommandName, *searchCommandStars)
		} else {
			search.Find(*searchCommandName)
		}
	case "signup":
		users := &api.Users{}
		users.Create(*signupEmail, *signupUsername, *signupPassword)
	}

}
