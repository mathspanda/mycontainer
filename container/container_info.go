package container

import (
	"encoding/json"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type ContainerInfo struct {
	Pid         string   `json:"pid"`
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Command     string   `json:"command"`
	CreateTime  string   `json:"createTime"`
	Status      string   `json:"status"`
	Volume      string   `json:"volume"`
	PortMapping []string `json:"portmapping"`
}

var (
	RUNNING             string = "running"
	STOP                string = "stopped"
	Exit                string = "exited"
	DefaultInfoLocation string = "/var/run/mycontainer/%s/"
	ConfigName          string = "config.json"
)

func RecordContainerInfo(containerPid int, containerName string, commandArray []string, volume string) (string, error) {
	id := randStringBytes(10)
	createTime := time.Now().Format("2006-01-02 15:04:05")
	command := strings.Join(commandArray, " ")
	if containerName == "" {
		containerName = id
	}

	containerInfo := &ContainerInfo{
		Id:         id,
		Pid:        strconv.Itoa(containerPid),
		Name:       containerName,
		Command:    command,
		CreateTime: createTime,
		Status:     RUNNING,
		Volume:     volume,
	}

	jsonBytes, err := json.Marshal(containerInfo)
	if err != nil {
		log.Errorf("Record container info error: %v", err)
		return "", err
	}

	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.Mkdir(dirUrl, 0622); err != nil {
		log.Errorf("Mkdir dir %s error: %v", dirUrl, err)
		return "", err
	}

	fileName := dirUrl + "/" + ConfigName
	file, err := os.Create(fileName)
	defer file.Close()
	if err != nil {
		log.Errorf("Create file %s error: %v", fileName, err)
		return "", err
	}

	if _, err := file.WriteString(string(jsonBytes)); err != nil {
		log.Errorf("File write string error: %v", err)
		return "", err
	}

	return containerName, nil
}

func DeleteContainerInfo(containerName string) {
	dirUrl := fmt.Sprintf(DefaultInfoLocation, containerName)
	if err := os.RemoveAll(dirUrl); err != nil {
		log.Errorf("Remove dir %s error: %v", dirUrl, err)
	}
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
