package common

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/go-ps"
	"k8s.io/klog/v2"
	"k8s.io/mount-utils"
)

func FuseMount(path string, command string, args []string) error {
	klog.Infof("Mounting fuse with command: %s and args: %s", command, args)

	systemArgs := []string{"systemd-run", "--scope", "--", command}
	systemArgs = append(systemArgs, args...)
	cmdStr := strings.Join(systemArgs, " ")

	klog.Infof("command string is: %s", cmdStr)

	if out, err := ConnectorRunInContainer(cmdStr); err != nil {
		if err != nil {
			return errors.New("fuseMount fail: " + err.Error() + ", out: " + out)
		}
	}

	return waitForMount(path, 10*time.Second)
}

func FuseUmount(path string) error {
	if err := mount.New("").Unmount(path); err != nil {
		return err
	}
	// as fuse quits immediately, we will try to wait until the process is done
	process, err := findFuseMountProcess(path)
	if err != nil {
		klog.Errorf("Error getting PID of fuse mount: %s", err)
		return nil
	}
	if process == nil {
		klog.Warningf("Unable to find PID of fuse mount %s, it must have finished already", path)
		return nil
	}
	klog.Infof("Found fuse pid %v of mount %s, checking if it still runs", process.Pid, path)
	return waitForProcess(process, 1)
}

func RunCommand(cmd string) (string, error) {
	out, err := exec.Command("sh", "-c", cmd).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("Failed to run cmd: " + cmd + ", with out: " + string(out) + ", with error: " + err.Error())
	}
	return string(out), nil
}

// IsDirExisting checks dir exist
func IsDirExisting(filename string) bool {
	_, err := os.Stat(filename)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return true
}

func waitForMount(path string, timeout time.Duration) error {
	var elapsed time.Duration
	var interval = 10 * time.Millisecond
	for {
		notMount, err := mount.New("").IsLikelyNotMountPoint(path)
		if err != nil {
			return err
		}
		if !notMount {
			return nil
		}
		time.Sleep(interval)
		elapsed = elapsed + interval
		if elapsed >= timeout {
			return errors.New("Timeout waiting for mount")
		}
	}
}

func findFuseMountProcess(path string) (*os.Process, error) {
	processes, err := ps.Processes()
	if err != nil {
		return nil, err
	}
	for _, p := range processes {
		cmdLine, err := getCmdLine(p.Pid())
		if err != nil {
			klog.Errorf("Unable to get cmdline of PID %v: %s", p.Pid(), err)
			continue
		}
		if strings.Contains(cmdLine, path) {
			klog.Infof("Found matching pid %v on path %s", p.Pid(), path)
			return os.FindProcess(p.Pid())
		}
	}
	return nil, fmt.Errorf("fuse mount process(%s) not found", path)
}

func getCmdLine(pid int) (string, error) {
	cmdLineFile := fmt.Sprintf("/proc/%v/cmdline", pid)
	cmdLine, err := ioutil.ReadFile(cmdLineFile)
	if err != nil {
		return "", err
	}
	return string(cmdLine), nil
}

func waitForProcess(p *os.Process, backoff int) error {
	if backoff == 20 {
		return fmt.Errorf("Timeout waiting for PID %v to end", p.Pid)
	}
	cmdLine, err := getCmdLine(p.Pid)
	if err != nil {
		klog.Warningf("Error checking cmdline of PID %v, assuming it is dead: %s", p.Pid, err)
		return nil
	}
	if cmdLine == "" {
		// ignore defunct processes
		// TODO: debug why this happens in the first place
		// seems to only happen on k8s, not on local docker
		klog.Warning("Fuse process seems dead, returning")
		return nil
	}
	if err := p.Signal(syscall.Signal(0)); err != nil {
		klog.Warningf("Fuse process does not seem active or we are unprivileged: %s", err)
		return nil
	}
	klog.Infof("Fuse process with PID %v still active, waiting...", p.Pid)
	time.Sleep(time.Duration(backoff*100) * time.Millisecond)
	return waitForProcess(p, backoff+1)
}
