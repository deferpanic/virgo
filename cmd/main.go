package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/deferpanic/dpcli/api"
	"github.com/deferpanic/virgo/pkg"
	"github.com/deferpanic/virgo/pkg/registry"
	"github.com/deferpanic/virgo/pkg/runner"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	process runner.Runner

	token  string
	hostOS string

	app     = kingpin.New("virgo", "Run Unikernels Locally")
	dry     = app.Flag("dry", "dry run, print commands only").Short('n').Bool()
	verbose = app.Flag("verbose", "verbose output").Short('v').Bool()

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

// depcheck does a quick dependency check to ensure
// that all required deps are installed - some auto-install
func depcheck() {
	if runtime.GOOS == "darwin" {
		pkg.OsCheck()
		pkg.QemuCheck()
		pkg.CpulimitCheck()
		pkg.TuntapCheck()
	}
	if runtime.GOOS == "linux" {
	}
}

// createQemuBlocks returns the set of blocks && lines

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

// run runs the unikernel on osx || linux
// locked down to one instance for now
func run(r registry.Registry) {
	var err error

	manifest := readManifest(r)

	// blocks, drives := createQemuBlocks(project, manifest)

	// env := ""
	// if manifest.Processes[0].Env != "" {
	// 	env = formatEnvs(manifest.Processes[0].Env)
	// }

	// projPath := pkg.ProjRoot + project

	// ip, gw := pkg.GetNetwork(projPath)
	// pkg.SetNetwork(projPath, ip, gw)

	// appendLn := "\"{ \\\"net\\\" : { \\\"if\\\":\\\"vioif0\\\",,\\\"type\\\":\\\"inet\\\",, \\\"method\\\":\\\"static\\\",, \\\"addr\\\":\\\"" + ip + "\\\",,  \\\"mask\\\":\\\"24\\\",,  \\\"gw\\\":\\\"" + gw + "\\\"},, " + env + blocks + " \\\"cmdline\\\": \\\"" + manifest.Processes[0].Cmdline + "\\\"}\""

	// kpath := pkg.ProjRoot + project + "/kernel/"
	// if strings.Contains(project, "/") {
	// 	s := strings.Split(project, "/")[1]
	// 	kpath += s
	// } else {
	// 	kpath += project
	// }

	// bootLine := ""

	// if manifest.Processes[0].Multiboot {
	// 	bootLine = " -kernel " + kpath + " -append " + appendLn
	// } else {
	// 	bootLine = " -hda " + kpath
	// }

	// kflag := "-no-kvm"
	// if runtime.GOOS == "linux" {
	// 	if kvmEnabled() {
	// 		kflag = "-enable-kvm"
	// 	}
	// }

	// if runtime.GOOS == "darwin" {
	// 	if pkg.CheckHAX() {
	// 		fmt.Println(api.GreenBold("hax is enabled!"))
	// 		kflag = "-accel hax"
	// 	}
	// }

	// pkg.SetupNetwork(projPath, gw)

	// mac := pkg.GenerateMAC()

	runLan := strconv.Itoa(len(running()) + 1)

	// @TODO, save pid into pidLn
	cmd := "sudo"
	args := []string{
		"qemu-system-x86_64", kflag, drives, "-nographic", "-vga", "none", "-serial",
		"file:", projPath + "/logs/blah.log", "-m", strconv.Itoa(manifest.Processes[0].Memory),
		"-net", "nic,model=virtio,vlan=" + runLan + ",macaddr=" + mac,
		"-net", "tap,vlan=" + runLan + ",ifname=tap" + runLan + ",script=" + projPath +
			"/ifup.sh,downscript=" + projPath + "/ifdown.sh ", bootLine,
	}

	process.SetDetached(true)
	if err = process.Exec(cmd, args...); err != nil {
		log.Fatalf("Error running %s %s\n", cmd, pkg.Join(args, " "))
	}
	process.SaveState(r.PidsDir())

	if runtime.GOOS == "darwin" {
		fmt.Println(api.GreenBold("setting sysctl"))

		if _, err = process.Run("sysctl", []string{"-w", "net.inet.ip.forwarding=1"}...); err != nil {
			log.Fatal("Error enabling ip forwarding")
		}

		if _, err = process.Run("sysctl", []string{"-w", "net.link.ether.inet.proxyall=1"}...); err != nil {
			log.Fatal("Error enabling proxyall")
		}

		// enable this for lower osx versions
		o := pkg.OsCheck()
		if pkg.NeedsFW(o) {
			if _, err = process.Run("sysctl", []string{"-w", "net.inet.ip.fw.enable=1"}...); err != nil {
				log.Fatal("Error enabling ip firewall")
			}
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
// func readManifest(projectName string) api.Manifest {
func readManifest(r registry.Registry) api.Manifest {
	// @TODO Find a use-case, then fix naming
	//
	// pName := projectName
	// if strings.Contains(pName, "/") {
	// 	pName = strings.Split(pName, "/")[1]
	// }

	if _, err := os.Stat(r.ManifestFile()); err != nil {
		log.Fatal(api.RedBold(r.ProjectName() + " error - " + err.Error()))
	}

	file, err := ioutil.ReadFile(r.ManifestFile())
	if err != nil {
		log.Fatal(api.RedBold("error reading " + projectName + " manifest - " + err.Error()))
	}

	var manifest api.Manifest
	json.Unmarshal(file, &manifest)

	return manifest
}

// rm removes a project locally
func rm(projectName string) {
	os.RemoveAll(pkg.ProjRoot + projectName)
}

// kill kill's the running project
// right now this assumes that there might be multiple instances of the
// same kind running and to kill them all - not sure why that would be
// the case but eh
func kill(projectName string) {
	projPath := pkg.ProjRoot + projectName

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
	projPath := pkg.ProjRoot + projectName

	logz := runCmd("cat " + projPath + "/logs/*")
	fmt.Println(logz)
}

// running builds a list of running projects
func running() []string {
	projs := ProjList(pkg.ProjRoot)

	running := []string{}
	for i := 0; i < len(projs); i++ {
		ppath := pkg.ProjRoot + projs[i] + "/pids"
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
}

func main() {
	var (
		stdout, stderr *os.File = nil, nil
	)

	if len(os.Args) < 2 {
		fmt.Println(logo)
	}

	if len(os.Args) > 1 && os.Args[1] == "signup" {
		api.Cli = api.NewCliImplementation("")
	} else {
		setToken()
	}

	if *verbose {
		stdout = os.Stdout
		stderr = os.Stderr
	}

	if *dry {
		process = runner.NewDryRunner(stdout)
	} else {
		process = runner.NewExecRunner(stdout, stderr, false)
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
		Images()
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
