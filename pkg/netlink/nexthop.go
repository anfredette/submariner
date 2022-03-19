package netlink

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"k8s.io/klog"
)

var nextNextHopId int
var nhMutex sync.Mutex

func GetNextHopId() int {
	nhMutex.Lock()
	nextNextHopId++
	nextHopId := nextNextHopId
	nhMutex.Unlock()
	return nextHopId
}

// RunCommand runs the cmd and returns the combined stdout and stderr
func RunCommand(cmd ...string) (string, error) {
	output, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to run %q: %s (%s)", strings.Join(cmd, " "), err, output)
	}
	return string(output), nil
}

// IsCommandAvailable checks to see if a binary is available in the current path
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	if err != nil {
		return false
	}
	return true
}

// CreateNextHop creates a next hop entry.
func CreateNextHop(nhid int, nextHopAddr string, nextHopDev string) error {
	stdout, err := RunCommand("ip", "nexthop", "add", "id", strconv.Itoa(nhid), "via", nextHopAddr, "dev", nextHopDev)
	if err != nil {
		klog.Info(stdout)
		klog.Info(err)
		return err
	}
	return nil
}

// CreateNextHopGroup creates a next hop group.
func CreateNextHopGroup(nhid int, nextHops []int) error {
	var nextHopList string
	for i, nextHop := range nextHops {
		if i > 0 {
			nextHopList += "/"
		}
		nextHopList += strconv.Itoa(nextHop)
	}

	stdout, err := RunCommand("ip", "nexthop", "add", "id", strconv.Itoa(nhid), "group", nextHopList)
	if err != nil {
		klog.Info(stdout)
		klog.Info(err)
		return err
	}
	return nil
}

// RouteDelete deletes a route. table is optional (use -1 to ignore)
func RouteDelete(network string, table int) error {
	command := []string{"ip", "route", "delete", "to", network}

	if table >= 0 {
		command = append(command, "table", strconv.Itoa(table))
	}

	klog.Infof("RouteDelete: %s", command)
	stdout, err := RunCommand(command...)
	if err != nil {
		klog.Infof("Error adding ip route: %s", stdout)
		klog.Infof("Error adding ip route: %v", err)
		return err
	}
	return nil
}

// RouteAddNextHop adds a route pointing to a nexthop or nexthop group. The following parameters are optional with
// ignore values given in parentheses: table (-1), metric (-1), proto (-1), scope (-1), source ("")
func RouteAddNextHop(network string, nhID int, table int, metric int, proto int, scope int, source string) error {
	// TODO: Need to handle the case where the route already exists, but for now, just try to delete it first.
	_ = RouteDelete(network, table)

	command := []string{"ip", "route", "add", "to", network, "nhid", strconv.Itoa(nhID)}

	if table >= 0 {
		command = append(command, "table", strconv.Itoa(table))
	}

	if metric >= 0 {
		command = append(command, "metric", strconv.Itoa(metric))
	}

	if proto >= 0 {
		command = append(command, "proto", strconv.Itoa(proto))
	}

	if scope >= 0 {
		command = append(command, "scope", strconv.Itoa(scope))
	}

	if source != "" {
		command = append(command, "src", source)
	}

	klog.Infof("RouteAddNextHop: %s", command)
	stdout, err := RunCommand(command...)
	if err != nil {
		klog.Infof("Error adding ip route: %s", stdout)
		klog.Infof("Error adding ip route: %v", err)
		return err
	}
	return nil
}
