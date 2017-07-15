package main

import (
	"os/exec"
	log "github.com/Sirupsen/logrus"
)

func commitContainer(imageName string) {
	mntUrl := "/root/mnt/"
	imageTar := "/root/" + imageName + ".tar"
	if _, err := exec.Command("tar", "-czf", imageTar, "-C", mntUrl, ".").CombinedOutput(); err != nil {
		log.Errorf("Tar folder %s error: %v", mntUrl, err)
	}
}
