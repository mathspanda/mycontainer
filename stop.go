package main

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"io/ioutil"
	"mycontainer/container"
	"os"
	"strconv"
	"syscall"
)

func stopContainer(containerName string) {
	pid, err := GetContainerPidByName(containerName)
	if err != nil {
		log.Errorf("Get container pid by name %s error: %v", pid, err)
		return
	}

	pidInt, err := strconv.Atoi(pid)
	if err != nil {
		log.Errorf("Convert pid from string to int error: %v", err)
		return
	}

	if err := syscall.Kill(pidInt, syscall.SIGTERM); err != nil {
		log.Errorf("Stop container %s error: %v", containerName, err)
		return
	}

	containerInfo, err := GetContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s error: %v", containerName, err)
		return
	}

	containerInfo.Status = container.STOP
	containerInfo.Pid = " "
	newContentBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Json marshal %s error: %v", containerName, err)
		return
	}

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	if err := ioutil.WriteFile(configFilePath, newContentBytes, 0622); err != nil {
		log.Errorf("Write file %s error: %v", configFilePath, err)
	}
}

func removeContainer(containerName string) {
	containerInfo, err := GetContainerInfoByName(containerName)
	if err != nil {
		log.Errorf("Get container %s info error: %v", containerName, err)
		return
	}

	if containerInfo.Status != container.STOP {
		log.Errorf("Couldn't remove running container")
		return
	}

	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Errorf("Remove dir %s error: %v", dirUrl, err)
	}
}

func GetContainerInfoByName(containerName string) (*container.ContainerInfo, error) {
	dirUrl := fmt.Sprintf(container.DefaultInfoLocation, containerName)
	configFilePath := dirUrl + container.ConfigName
	contentBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		log.Errorf("Read file %s error: %v", configFilePath, err)
		return nil, err
	}

	var containerInfo container.ContainerInfo
	if err := json.Unmarshal(contentBytes, &containerInfo); err != nil {
		log.Errorf("GetContainerInfoByName unmarshal error: %v", err)
		return nil, err
	}

	return &containerInfo, nil
}
