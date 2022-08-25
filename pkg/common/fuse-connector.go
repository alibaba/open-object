package common

import (
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/sevlyar/go-daemon"
	"k8s.io/klog/v2"
)

func RunConnector() {
	cntxt := &daemon.Context{
		PidFileName: ConnectorPIDFilename,
		PidFilePerm: 0644,
		LogFileName: ConnectorLogFilename,
		LogFilePerm: 0640,
		WorkDir:     ConnectorWorkPath,
		Umask:       027,
	}

	d, err := cntxt.Reborn()
	if err != nil {
		klog.Fatalf("Unable to run connector: %s", err.Error())
	}
	if d != nil {
		return
	}
	defer cntxt.Release()
	klog.Info("Fuse Connector Daemon Is Starting...")

	runFuseProxy()
}

func ConnectorRunInContainer(cmd string) (string, error) {
	c, err := net.Dial("unix", filepath.Join(HostDir, ConnectorSocketPath))
	if err != nil {
		klog.Infof("Fuse connector Dial error: %s", err.Error())
		return err.Error(), err
	}
	defer c.Close()

	_, err = c.Write([]byte(cmd))
	if err != nil {
		klog.Infof("Fuse connector write error: %s", err.Error())
		return err.Error(), err
	}

	buf := make([]byte, 2048)
	n, err := c.Read(buf[:])
	response := string(buf[0:n])
	if strings.HasPrefix(response, "Success") {
		respstr := response[8:]
		return respstr, nil
	}
	return response, errors.New("Fuse connector exec command error:" + response)
}

func runFuseProxy() {
	if IsDirExisting(ConnectorSocketPath) {
		os.Remove(ConnectorSocketPath)
	} else {
		pathDir := filepath.Dir(ConnectorSocketPath)
		if !IsDirExisting(pathDir) {
			os.MkdirAll(pathDir, os.ModePerm)
		}
	}

	klog.Infof("Socket path is ready: %s", ConnectorSocketPath)
	ln, err := net.Listen("unix", ConnectorSocketPath)
	if err != nil {
		klog.Fatalf("Server Listen error: %s", err.Error())
	}
	klog.Info("Daemon Started ...")
	defer ln.Close()

	// Handler to process the command
	for {
		fd, err := ln.Accept()
		if err != nil {
			klog.Infof("Server Accept error: %s", err.Error())
			continue
		}
		go echoServer(fd)
	}
}

func echoServer(c net.Conn) {
	buf := make([]byte, 2048)
	nr, err := c.Read(buf)
	if err != nil {
		klog.Infof("Server Read error: ", err.Error())
		return
	}

	cmd := string(buf[0:nr])
	klog.Infof("Server Receive CSI command: %s", cmd)
	// run command
	if out, err := RunCommand(cmd); err != nil {
		reply := "Fail: " + cmd + ", error: " + err.Error()
		_, err = c.Write([]byte(reply))
		klog.Infof("Server Fail to run cmd:", reply)
	} else {
		out = "Success:" + out
		_, err = c.Write([]byte(out))
		klog.Infof("Success: %s", out)
	}
}
