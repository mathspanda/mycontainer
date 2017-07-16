package container

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func pivotRoot(root string) error {
	if err := syscall.Mount(root, root, "bind", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("Mount rootfs to itself error: %v", err)
	}

	pivotDir := filepath.Join(root, ".pivot_root")
	if err := os.Mkdir(pivotDir, 0700); err != nil {
		return err
	}

	if err := syscall.PivotRoot(root, pivotDir); err != nil {
		return fmt.Errorf("pivot_root %v", err)
	}

	if err := syscall.Chdir("/"); err != nil {
		return fmt.Errorf("chdir / %v", err)
	}

	pivotDir = filepath.Join("/", ".pivot_root")
	if err := syscall.Unmount(pivotDir, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("unmount pivot_root dir %v", err)
	}

	return os.Remove(pivotDir)
}

func setUpMount() {
	pwd, err := os.Getwd()
	if err != nil {
		log.Errorf("Get current location error: %v", err)
		return
	}
	log.Infof("Current location is %v", pwd)

	if err := pivotRoot(pwd); err != nil {
		log.Errorf("Pivot root error: %v", err)
		return
	}

	defaultMountFlags := syscall.MS_NOEXEC | syscall.MS_NOSUID | syscall.MS_NODEV
	syscall.Mount("proc", "/proc", "proc", uintptr(defaultMountFlags), "")
	syscall.Mount("tmpfs", "/dev", "tmpfs", syscall.MS_NOSUID|syscall.MS_STRICTATIME, "mode=755")
}

func NewWorkSpace(volume string, containerName string, imageName string) {
	CreateReadOnlyLayer(imageName)
	CreateWriteLayer(containerName)
	CreateMountPoint(containerName, imageName)
	if volume != "" {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
			log.Infof("mount %q", volumeUrls)
			MountVolume(volumeUrls, containerName)
		} else {
			log.Infof("volume parameter is not correct")
		}
	}
}

func MountVolume(volumeUrls []string, containerName string) error {
	parentUrl := volumeUrls[0]
	if err := os.MkdirAll(parentUrl, 0777); err != nil {
		log.Infof("Mkdir parent dir %s error: %v", parentUrl, err)
	}

	mntUrl := fmt.Sprintf(MntUrl, containerName)
	containerVolumeUrl := mntUrl + "/" + volumeUrls[1]
	if err := os.MkdirAll(containerVolumeUrl, 0777); err != nil {
		log.Infof("Mkdir container dir %s error: %v", containerVolumeUrl, err)
	}

	dirs := "dirs=" + parentUrl
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerVolumeUrl).CombinedOutput()
	if err != nil {
		log.Errorf("Mount volume error: %v", err)
		return err
	}
	return nil
}

func CreateReadOnlyLayer(imageName string) {
	unTarFolderUrl := RootUrl + "/" + imageName + "/"
	imageUrl := RootUrl + "/" + imageName + ".tar"
	exist, err := PathExists(unTarFolderUrl)
	if err != nil {
		log.Errorf("Fail to judge whether dis %s exists: %v", unTarFolderUrl, err)
	}
	if exist == false {
		if err := os.MkdirAll(unTarFolderUrl, 0777); err != nil {
			log.Errorf("Mkdir dir %s error: %v", unTarFolderUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", imageUrl, "-C", unTarFolderUrl).CombinedOutput(); err != nil {
			log.Errorf("untar %s error: %v", imageUrl, err)
		}
	}
}

func CreateWriteLayer(containerName string) {
	writeUrl := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.MkdirAll(writeUrl, 0777); err != nil {
		log.Errorf("Mkdir dir %s error: %v", writeUrl, err)
	}
}

func CreateMountPoint(containerName, imageName string) error {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	if err := os.MkdirAll(mntUrl, 0777); err != nil {
		log.Errorf("Mkdir dir %s error: %v", mntUrl, err)
		return err
	}

	tmpWriteLayer := fmt.Sprintf(WriteLayerUrl, containerName)
	tmpImageLocation := RootUrl + "/" + imageName
	dirs := "dirs=" + tmpWriteLayer + ":" + tmpImageLocation
	_, err := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl).CombinedOutput()
	if err != nil {
		log.Errorf("aufs mount error: %v", err)
		return err
	}
	return nil
}

func DeleteWorkSpace(volume string, containerName string) {
	if volume != "" {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
			DeleteVolumeMountPoint(containerName, volumeUrls)
		}
	}
	DeleteMountPoint(containerName)
	DeleteWriteLayer(containerName)
}

func DeleteMountPoint(containerName string) error {
	mntUrl := fmt.Sprintf(MntUrl, containerName)
	if _, err := exec.Command("umount", mntUrl).CombinedOutput(); err != nil {
		log.Errorf("aufs umount %s error: %v", mntUrl, err)
		return err
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("Remove dir %s error: %v", mntUrl, err)
		return err
	}
	return nil
}

func DeleteVolumeMountPoint(containerName string, volumeUrls []string) error {
	containerUrl := fmt.Sprintf(MntUrl, containerName) + "/" + volumeUrls[1]
	if _, err := exec.Command("umount", containerUrl).CombinedOutput(); err != nil {
		log.Errorf("umount volume %s error: %v", volumeUrls[1], err)
		return err
	}
	return nil
}

func DeleteWriteLayer(containerName string) {
	writeUrl := fmt.Sprintf(WriteLayerUrl, containerName)
	if err := os.RemoveAll(writeUrl); err != nil {
		log.Infof("Remove dir %s error: %v", writeUrl, err)
	}
}

func volumeUrlExtract(volume string) []string {
	volumeUrls := strings.Split(volume, ":")
	return volumeUrls
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
