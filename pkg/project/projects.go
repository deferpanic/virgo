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
	Process     []*runner.ExecRunner
	Network     []network.Network
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

func (ps Projects) Add(p *Project, r *registry.Registry) error {
	if _, ok := p.Process.(runner.DryRunner); ok {
		return nil
	}

	for i, _ := range ps {
		if ps[i].ProjectName == p.Name() {
			ps[i].Process = append(ps[i].Process, p.Process.(*runner.ExecRunner))
			ps[i].Network = append(ps[i].Network, p.Network)
			return ps.save(r)
		}
	}

	rt := &Runtime{
		ProjectName: p.Name(),
		Process:     []*runner.ExecRunner{p.Process.(*runner.ExecRunner)},
		Network:     []network.Network{p.Network},
	}

	ps = append(ps, rt)

	return ps.save(r)
}

func (ps Projects) Delete(rt *Runtime, r *registry.Registry) error {
	for i, _ := range ps {
		if ps[i].ProjectName == rt.ProjectName {
			ps = append(ps[:i], ps[i+1:]...)
			return ps.save(r)
		}
	}

	return fmt.Errorf("project '%s' not found in runtime", rt.ProjectName)
}

func (ps Projects) save(r *registry.Registry) error {
	wr, err := os.OpenFile(r.RuntimeFile(), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error opening file '%s' - %s", r.RuntimeFile(), err)
	}
	defer wr.Close()

	b, err := json.Marshal(ps)
	if err != nil {
		return fmt.Errorf("error marshalling Projects - %s", err)
	}

	if _, err := wr.Write(b); err != nil {
		return fmt.Errorf("error writing file '%s' - %s", r.RuntimeFile(), err)
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

	if len(ps.Running()) == 0 {
		return highIP.To4().String(), highGw.To4().String()
	}

	if len(ps.Running()) > 0 {
		for _, p := range ps {
			ip := net.ParseIP(p.Network[len(p.Network)-1].Ip).To4()
			if ip[2] > highIP[2] {
				highIP = ip
			}
		}

		highIP[2]++

		if highIP[2] == 255 {
			return "", ""
		}
	}

	highGw[2] = highIP[2]

	return highIP.To4().String(), highGw.To4().String()
}

func (ps Projects) NextNum() int {
	var n int = 0

	for i, _ := range ps {
		n += len(ps[i].Process)
	}

	return n + 1
}

func (ps Projects) String() string {
	var result string

	if len(ps) == 0 {
		return ""
	}

	result += "Projectname\tGw\tIP\tMem\tPids\tMAC\n"

	for _, p := range ps {
		pids := []string{}

                p.Process[0].Mem = p.Process[0].GetMem(p.Process[0].Pid)
		for _, instance := range p.Process {
			pids = append(pids, strconv.Itoa(instance.Pid))
		}

		result += fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\n", p.ProjectName, p.Network[0].Gw, p.Network[0].Ip,                           p.Process[0].Mem,tools.Join(pids, ", "),p.Network[0].Mac)

		for i := 1; i < len(p.Network); i++ {
			result += fmt.Sprintf("\t%s\t%s\t%s\n", p.Network[i].Gw, p.Network[i].Ip, p.Network[i].Mac)
		}

	}

	return result
}
