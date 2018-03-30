package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"text/tabwriter"

	"github.com/deferpanic/dpcli/api"
	"github.com/deferpanic/virgo/pkg/depcheck"
	"github.com/deferpanic/virgo/pkg/network"
	"github.com/deferpanic/virgo/pkg/project"
	"github.com/deferpanic/virgo/pkg/registry"
	"github.com/deferpanic/virgo/pkg/runner"
	"github.com/deferpanic/virgo/pkg/tools"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	token  string
	hostOS string

	app = kingpin.New("virgo", "Run Unikernels Locally")
	dry = app.Flag("dry", "dry run, print commands only").Short('n').Bool()

	pullCommand     = app.Command("pull", "Pull a project")
	pullProjectName = pullCommand.Arg("name", "Project name.").Required().String()

	runCmd         = app.Command("run", "Run a project")
	runProjectName = runCmd.Arg("name", "Project name.").Required().String()

	killCommand     = app.Command("kill", "Kill a running project")
	killProjectName = killCommand.Arg("name", "Project name.").Required().String()

	rmCommand     = app.Command("rm", "Remove a project")
	rmProjectName = rmCommand.Arg("name", "Project name.").Required().String()

	logCommand     = app.Command("log", "Fetch log of project")
	logProjectName = logCommand.Arg("name", "Project name.").Required().String()

	searchCommand      = app.Command("search", "Search for a project")
	searchCommandName  = searchCommand.Arg("description", "Description").Required().String()
	searchCommandStars = searchCommand.Arg("stars", "Star Count").Int()

	signupCommand  = app.Command("signup", "Signup")
	signupEmail    = signupCommand.Arg("email", "Email.").Required().String()
	signupUsername = signupCommand.Arg("username", "Username.").Required().String()
	signupPassword = signupCommand.Arg("password", "Password.").Required().String()

	psCommand = app.Command("ps", "List running projects")

	listCommand = app.Command("list", "List all projects")
	listJson    = listCommand.Flag("json", "output as json").Bool()
)

func main() {
	var (
		stdout, stderr *os.File = os.Stdout, os.Stderr
		process        runner.Runner
	)

	if len(os.Args) < 2 {
		fmt.Println(tools.Logo)
	}

	if len(os.Args) > 1 && os.Args[1] == "signup" {
		api.Cli = api.NewCliImplementation("")
	} else {
		if err := tools.SetToken(); err != nil {
			log.Fatalf("%s\nif you have and account add your token to '~/.dprc' otherwise signup via\nvirgo signup my@email.com username password", err)
		}
	}

	command := kingpin.MustParse(app.Parse(os.Args[1:]))

	if *dry {
		process = runner.NewDryRunner(stdout)
	} else {
		process = runner.NewExecRunner(stdout, stderr, false)
	}

	dep := depcheck.New(process)

	if command == "run" && runtime.GOOS == "darwin" {
		var err error
		log.Println("setting sysctl")

		if _, err = process.Shell("sysctl -w net.inet.ip.forwarding=1"); err != nil {
			log.Fatalf("Error enabling ip forwarding - %s", err)
		}

		if _, err = process.Shell("sysctl -w net.link.ether.inet.proxyall=1"); err != nil {
			log.Fatalf("error enabling proxyall - %s", err)
		}

		// enable this for lower osx versions
		version, err := dep.OsCheck()
		if err != nil {
			log.Fatal(err)
		}
		if dep.IsNeedFw(version) {
			if _, err = process.Shell("sysctl -w net.inet.ip.fw.enable=1"); err != nil {
				log.Fatalf("error enabling ip firewall - %s", err)
			}
		}
	}

	r, err := registry.New()
	if err != nil {
		log.Fatal(err)
	}

	projects, err := project.LoadProjects(r)
	if err != nil {
		log.Fatal(err)
	}

	killProject := func() {
		rt := projects.GetProjectByName(*killProjectName)
		if rt == nil {
			log.Fatalf("Project '%s' isn't running\n", *killProjectName)
		}

		for _, instance := range rt.Process {
			instance.Stop()
		}

		if err := projects.Delete(rt, r); err != nil {
			log.Fatal(err)
		}
	}

	switch command {
	case "pull":
		pr, err := r.AddProject(*pullProjectName)
		if err != nil {
			log.Fatal(err)
		}

		if err := project.Pull(pr); err != nil {
			log.Fatal(err)
		}

	case "run":
		pr, err := r.AddProject(*runProjectName)
		if err != nil {
			log.Fatal(err)
		}

		ip, gw := projects.GetNextNetowrk()
		if ip == "" || gw == "" {
			log.Fatal("Ip range is exceeded, unable to proceed")
		}

		network, err := network.New(pr, ip, gw)
		if err != nil {
			log.Fatal(err)
		}

		p, err := project.New(pr, network, process, projects.NextNum())
		if err != nil {
			log.Fatal(err)
		}

		if err = p.Run(); err != nil {
			log.Fatal(err)
		}

		if err := projects.Add(p, r); err != nil {
			log.Fatal(err)
		}

	case "ps":
		w := tabwriter.NewWriter(os.Stdout, 1, 8, 0, '\t', 0)

		fmt.Fprintf(w, "%s", projects)

	case "kill":
		killProject()

	case "rm":
		killProject()

		if err := r.PurgeProject(*rmProjectName); err != nil {
			log.Fatal(err)
		}

	// Just shows content of log/ directory
	case "log":
		pr, err := r.AddProject(*logProjectName)
		if err != nil {
			log.Fatal(err)
		}

		if err := tools.ShowFiles(pr.LogsDir()); err != nil {
			log.Fatal(err)
		}

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
