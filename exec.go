package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"mycontainer/container"
	_ "mycontainer/nsenter"
	"os"
	"os/exec"
	"strings"
)

const ENV_EXEC_PID = "mycontainer_pid"
const ENV_EXEC_CMD = "mycontainer_cmd"

func ExecContainer(containerName string, comArray []string) {
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Exec container GetContainerPidByName %s error: %v", containerName, err)
		return
	}

	cmdStr := strings.Join(comArray, " ")
	cmd := exec.Command("/proc/self/exe", "exec")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	os.Setenv(ENV_EXEC_PID, pid)
	os.Setenv(ENV_EXEC_CMD, cmdStr)

	if err := cmd.Run(); err != nil {
		log.Errorf("Exec container %s error %v", containerName, err)
	}
}

func GetContainerPidByName(containerName string) (string, error) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName

	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return "", err
	}

	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		return "", err
	}

	return containerInfo.Pid, nil
}
