package s3minio

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/alibaba/open-object/pkg/common"
)

func S3FSMount(url, bucket, mountpoint, AK, SK string) error {
	pwFileContent := AK + ":" + SK
	if err := writeS3fsPassword(pwFileContent); err != nil {
		return err
	}
	// s3fs acs-kok:/ /mnt/s3fs/ -ourl=http://10.254.230.59:9000 -opasswd_file=/root/.passwd-s3fs -ouse_path_request_style -oallow_other -omp_umask=000
	args := []string{
		fmt.Sprintf("%s:/", bucket),
		fmt.Sprintf("%s", mountpoint),
		fmt.Sprintf("-ourl=%s", url),
		fmt.Sprintf("-opasswd_file=%s", filepath.Join(common.ConfigDir, S3FSPassWordFileName)),
		"-ouse_path_request_style",
		"-oallow_other",
		"-omp_umask=0000",
	}

	return common.FuseMount(mountpoint, S3FSCmd, args)
}

func S3FSUmount(mountpoint string) error {
	return common.FuseUmount(mountpoint)
}

func writeS3fsPassword(password string) error {
	if err := makeDir(filepath.Join(common.HostDir, common.ConfigDir)); err != nil {
		return err
	}

	pwFilePath := filepath.Join(common.HostDir, common.ConfigDir, S3FSPassWordFileName)
	pwFile, err := os.Create(pwFilePath)
	if err != nil {
		return err
	}
	err = os.Chmod(pwFilePath, 0600)
	if err != nil {
		return err
	}
	defer pwFile.Close()
	_, err = pwFile.WriteString(password)
	if err != nil {
		return err
	}
	return nil
}

func makeDir(path string) error {
	err := os.MkdirAll(path, os.FileMode(0755))
	if err != nil {
		if !os.IsExist(err) {
			return err
		}
	}
	return nil
}
