package main

import (
	log "github.com/Sirupsen/logrus"
	"os/exec"
)

func commitContainer(imageName string) {
	mntUrl := "/root/mnt/"
	imageTar := "/root/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error: %v", mntUrl, err)
	}
}
