package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/procfs"
)

type schedstat struct {
	SumExecRuntime float64
	RunDelay       float64
	Pcount         float64
}

func collectVMProcesses() (procfs.Procs, error) {
	var vm procfs.Procs

	processes, err := procfs.AllProcs()
	if err != nil {
		return nil, err
	}

	for _, p := range processes {
		cmdline, err := p.CmdLine()
		if err != nil {
			return nil, err
			break
		}

		/* skip kernel therad */
		if (len(cmdline)) == 0 {
			continue
		}

		if strings.Index(cmdline[0], "qemu-kvm") > 0 {
			vm = append(vm, p)
		}
	}

	return vm, nil
}

func collectVMTIDs(vm procfs.Proc) ([]int64, error) {
	var tids []int64

	tasks := fmt.Sprintf("/proc/%d/task/", vm.PID)
	d, err := os.Open(tasks)
	if err != nil {
		return nil, fmt.Errorf("could not open %s: %s", d.Name(), err)
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return nil, fmt.Errorf("could not read %s: %s", d.Name(), err)
	}

	for _, n := range names {
		tid, err := strconv.ParseInt(n, 10, 64)
		if err != nil {
			continue
		}
		tids = append(tids, tid)
	}

	return tids, nil
}

// Try to guess VM name from qemu-kvm of cmdline options
func guessVMName(cmdline []string) string {

	if len(cmdline) == 0 {
		return ""
	}

	for i, token := range cmdline {
		t := string(token)
		if strings.Index(t, "-name") == 0 {
			// -name\0www.exmpale.com,qemu:process=qemu:www.exmpale.com
			//  \___i  \___i+1
			name := string(cmdline[i+1])
			// www.exmpale.com,qemu:process=qemu:www.exmpale.com => www.example.com
			name = strings.Split(name, ",")[0]
			// www.exmpale.com -> www-example-com
			name = strings.Replace(name, ".", "-", -1)

			return name
		}
	}

	return ""
}

func collectRundelay(vms procfs.Procs) map[string]float64 {
	var stats = make(map[string]float64)

	for _, vm := range vms {
		var name string

		cmdline, err := vm.CmdLine()
		if err != nil {
			name = string(vm.PID)
		}
		name = guessVMName(cmdline)
		if name == "" {
			name = string(vm.PID)
		}

		stats[name] = 0

		tids, err := collectVMTIDs(vm)
		if err != nil {
			return nil
		}

		for _, tid := range tids {
			path := fmt.Sprintf("/proc/%d/task/%d/schedstat", tid, tid)
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return nil
			}

			values := make([]float64, 10)
			fields := strings.Fields(string(bytes))
			for i, strValue := range fields {
				values[i], err = strconv.ParseFloat(strValue, 64)
				if err != nil {
					return nil
				}
			}

			st := &schedstat{
				SumExecRuntime: values[0],
				RunDelay:       values[1],
				Pcount:         values[2],
			}

			stats[name] += st.RunDelay
		}
	}

	return stats
}

func main() {

	_, err := os.Stat("/proc/self/schedstat")
	if os.IsNotExist(err) {
		fmt.Printf("CONFIG_SCHEDSTATS not be enabled")
		return
	}

	if os.Getenv("MACKEREL_AGENT_PLUGIN_META") != "" {
	} else {
		interval := 1000 * time.Millisecond

		vms, err := collectVMProcesses()
		if err != nil {
			return
		}

		/* delta */
		prevStats := collectRundelay(vms)
		time.Sleep(interval)
		currentStats := collectRundelay(vms)

		now := time.Now()
		for name, _ := range prevStats {

			/* check if VM has been started or shutdowned */
			if _, ok := currentStats[name]; !ok {
				continue
			}

			if _, ok := prevStats[name]; !ok {
				continue
			}

			deltaRunDelay := currentStats[name] - prevStats[name]
			fmt.Printf("vm.%s.run_delay\t%f\t%d\n", name, deltaRunDelay/1000/1000/1000*100, now.Unix())
		}
	}
}
