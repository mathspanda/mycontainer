package main

import (
	log "github.com/Sirupsen/logrus"
	"mycontainer/cgroups"
	"mycontainer/cgroups/subsystems"
	"mycontainer/container"
	"os"
	"strings"
)

func Run(containerName string, tty bool, cmdArray []string, res *subsystems.ResourceConfig, volume string) {
	parent, writePipe := container.NewParentProcess(tty, volume)
	if parent == nil {
		log.Error("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// record container info
	containerName, err := container.RecordContainerInfo(parent.Process.Pid, containerName, cmdArray, volume)
	if err != nil {
		log.Errorf("Record container info error: %v", err)
		return
	}

	// cgroups
	cgroupManager := cgroups.NewCgroupManager("mycontainer")
	defer cgroupManager.Destroy()
	cgroupManager.Set(res)
	cgroupManager.Apply(parent.Process.Pid)

	sendInitCommand(cmdArray, writePipe)

	if tty {
		parent.Wait()
		container.DeleteWorkSpace("/root/", "/root/mnt/", volume)
		os.Exit(0)
	}
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}
