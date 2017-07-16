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

func NewWorkSpace(rootUrl string, mntUrl string, volume string) {
	CreateReadOnlyLayer(rootUrl)
	CreateWriteLayer(rootUrl)
	CreateMountPoint(rootUrl, mntUrl)
	if volume != "" {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
			log.Infof("mount %q", volumeUrls)
			MountVolume(mntUrl, volumeUrls)
		} else {
			log.Infof("volume parameter is not correct")
		}
	}
}

func MountVolume(mntUrl string, volumeUrls []string) {
	parentUrl := volumeUrls[0]
	if err := os.Mkdir(parentUrl, 0777); err != nil {
		log.Infof("Mkdir parent dir %s error: %v", parentUrl, err)
	}

	containerUrl := mntUrl + volumeUrls[1]
	if err := os.Mkdir(containerUrl, 0777); err != nil {
		log.Infof("Mkdir container dir %s error: %v", containerUrl, err)
	}

	dirs := "dirs=" + parentUrl
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("Mount volume error: %v", err)
	}
}

func CreateReadOnlyLayer(rootUrl string) {
	busyboxUrl := rootUrl + "busybox/"
	busyboxTarUrl := rootUrl + "busybox.tar"
	exist, err := PathExists(busyboxUrl)
	if err != nil {
		log.Errorf("Fail to judge whether dis %s exists: %v", busyboxUrl, err)
	}
	if exist == false {
		if err := os.Mkdir(busyboxUrl, 0777); err != nil {
			log.Errorf("Mkdir dir %s error: %v", busyboxUrl, err)
		}
		if _, err := exec.Command("tar", "-xvf", busyboxTarUrl, "-C", busyboxUrl).CombinedOutput(); err != nil {
			log.Errorf("untar %s error: %v", busyboxTarUrl, err)
		}
	}
}

func CreateWriteLayer(rootUrl string) {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.Mkdir(writeUrl, 0777); err != nil {
		log.Errorf("Mkdir dir %s error: %v", writeUrl, err)
	}
}

func CreateMountPoint(rootUrl string, mntUrl string) {
	if err := os.Mkdir(mntUrl, 0777); err != nil {
		log.Errorf("Mkdir dir %s error: %v", mntUrl, err)
	}

	dirs := "dirs=" + rootUrl + "writeLayer:" + rootUrl + "busybox"
	cmd := exec.Command("mount", "-t", "aufs", "-o", dirs, "none", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("aufs mount error: %v", err)
	}
}

func DeleteWorkSpace(rootUrl string, mntUrl string, volume string) {
	if volume != "" {
		volumeUrls := volumeUrlExtract(volume)
		if len(volumeUrls) == 2 && volumeUrls[0] != "" && volumeUrls[1] != "" {
			DeleteVolumeMountPoint(mntUrl, volumeUrls)
		}
	}
	DeleteMountPoint(mntUrl)
	DeleteWriteUrl(rootUrl)
}

func DeleteMountPoint(mntUrl string) {
	cmd := exec.Command("umount", mntUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("aufs umount %s error: %v", mntUrl, err)
	}
	if err := os.RemoveAll(mntUrl); err != nil {
		log.Errorf("Remove dir %s error: %v", mntUrl, err)
	}
}

func DeleteVolumeMountPoint(mntUrl string, volumeUrls []string) {
	containerUrl := mntUrl + volumeUrls[1]
	cmd := exec.Command("umount", containerUrl)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Errorf("umount volume %s error: %v", volumeUrls[1], err)
	}
}

func DeleteWriteUrl(rootUrl string) {
	writeUrl := rootUrl + "writeLayer/"
	if err := os.RemoveAll(writeUrl); err != nil {
		log.Errorf("Remove dir %s error: %v", writeUrl, err)
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
