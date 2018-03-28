package project

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"

	"github.com/deferpanic/virgo/pkg/network"
	"github.com/deferpanic/virgo/pkg/registry"
	"github.com/deferpanic/virgo/pkg/runner"
	"github.com/deferpanic/virgo/pkg/tools"
)

type Runtime struct {
	ProjectName string
	Process     []runner.ExecRunner
	Network     network.Network
}

type Projects []*Runtime

func LoadProjects(r *registry.Registry) (Projects, error) {
	result := make(Projects, 0)

	b, err := ioutil.ReadFile(r.RuntimeFile())
	if err != nil && os.IsNotExist(err) {
		return result, nil
	}
	if err != nil {
		return nil, fmt.Errorf("error reading %s - %s", r.RuntimeFile(), err)
	}

	if err := json.Unmarshal(b, &result); err != nil {
		return nil, fmt.Errorf("error unmarshalling %s - %s", r.RuntimeFile(), err)
	}

	return result, nil
}

func (ps Projects) GetProjectByName(name string) *Runtime {
	for _, p := range ps {
		if p.ProjectName == name {
			return p
		}
	}

	return nil
}

func (ps Projects) Running() Projects {
	result := make(Projects, 0)

	for _, p := range ps {
		// for now it doesn't matter how many instances are running
		if len(p.Process) > 0 && p.Process[0].IsAlive() {
			result = append(result, p)
		}
	}

	return result
}

func (ps Projects) GetNextNetowrk() (string, string) {
	highIP := net.IP{10, 1, 2, 4}.To4()
	highGw := net.IP{10, 1, 2, 1}.To4()

	if len(ps.Running()) > 0 {
		for _, p := range ps {
			ip := net.ParseIP(p.Network.Ip).To4()
			if ip[2] > highIP[2] {
				highIP = ip
			}
		}

		highIP[2]++
		highGw[2]++

		if highIP[2] == 255 {
			return "", ""
		}
	}

	return highIP.To4().String(), highGw.To4().String()
}

func (ps Projects) NextNum() int {
	return len(ps) + 1
}

func (ps Projects) String() string {
	var result string

	for _, p := range ps {
		pids := []string{}

		result += fmt.Sprintf("%s", p.ProjectName)
		result += fmt.Sprintf("\tGw\tIP\tMAC\n")
		result += fmt.Sprintf("\t%s\t%s\t%s\n", p.Network.Gw, p.Network.Ip, p.Network.Mac)

		for _, instance := range p.Process {
			pids = append(pids, strconv.Itoa(instance.Pid))
		}

		result += fmt.Sprintf("\tPids: %s", tools.Join(pids, ", "))
	}

	if result != "" {
		result += "\n"
	}

	return result
}
