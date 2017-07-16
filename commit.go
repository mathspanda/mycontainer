package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"mycontainer/container"
	"os/exec"
)

func commitContainer(containerName, imageName string) {
	mntUrl := fmt.Sprintf(container.MntUrl, containerName) + "/"
	imageTar := container.RootUrl + "/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error: %v", mntUrl, err)
	}
}
