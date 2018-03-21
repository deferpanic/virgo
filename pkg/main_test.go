package pkg

// import (
// 	"os"
// 	"testing"
// )

// func TestPs(t *testing.T) {

// 	projRoot = "/tmp/.virgotest/projects/"

// 	// nuke stuff out
// 	os.RemoveAll(projRoot)

// 	// create 3 projects
// 	os.MkdirAll(projRoot+"/myuser/myproj/pids", 0755)
// 	os.MkdirAll(projRoot+"/myuser/anotherproj/pids", 0755)
// 	os.MkdirAll(projRoot+"/anotheruser/anotherproj/pids", 0755)

// 	// every project has to have a manifest
// 	runCmd("touch " + projRoot + "/myuser/myproj/myproj.manifest")
// 	runCmd("touch " + projRoot + "/myuser/anotherproj/anotherproj.manifest")
// 	runCmd("touch " + projRoot + "/anotheruser/anotherproj/anotherproj.manifest")

// 	// pretend processes running
// 	runCmd("touch " + projRoot + "/myuser/myproj/pids/1234.pid")
// 	runCmd("touch " + projRoot + "/anotheruser/anotherproj/pids/4312.pid")

// 	list := running()

// 	if len(list) != 2 {
// 		t.Logf("%v", len(list))
// 		t.Fatalf("not enough running projects")
// 	}

// }
