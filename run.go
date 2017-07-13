package main

import (
	log "github.com/Sirupsen/logrus"
	"mycontainer/cgroups"
	"mycontainer/cgroups/subsystems"
	"mycontainer/container"
	"os"
	"strings"
)

func Run(tty bool, command string, res *subsystems.ResourceConfig) {
	parent, _ := container.NewParentProcess(tty, command)
	if parent == nil {
		log.Error("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	cgroupManager := cgroups.NewCgroupManager("mycontainer")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	//sendInitCommand(comArray, writePipe)
	parent.Wait()
}

func sendInitCommand(comArray []string, writePipe *os.File) {
	command := strings.Join(comArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteAt([]byte(command), 0)
	writePipe.Close()
}
