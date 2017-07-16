package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"mycontainer/cgroups"
	"mycontainer/cgroups/subsystems"
	"mycontainer/container"
	"os"
	"strings"
	"time"
)

func Run(containerName string, tty bool, cmdArray []string, res *subsystems.ResourceConfig, volume string) {
	id := randStringBytes(10)
	if containerName == "" {
		containerName = id
	}
	fmt.Fprintln(os.Stdout, containerName)

	parent, writePipe := container.NewParentProcess(containerName, tty, volume)
	if parent == nil {
		log.Error("New parent process error")
		return
	}
	if err := parent.Start(); err != nil {
		log.Error(err)
	}

	// record container info
	containerName, err := container.RecordContainerInfo(id, parent.Process.Pid, containerName, cmdArray, volume)
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
		container.DeleteContainerInfo(containerName)
		os.Exit(0)
	}
}

func sendInitCommand(cmdArray []string, writePipe *os.File) {
	command := strings.Join(cmdArray, " ")
	log.Infof("command all is %s", command)
	writePipe.WriteString(command)
	writePipe.Close()
}

func randStringBytes(n int) string {
	letterBytes := "1234567890abcdefghijklmnopqrstuvwxyz"
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
