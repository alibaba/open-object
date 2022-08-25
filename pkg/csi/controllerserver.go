/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package csi

import (
	"fmt"
	"net"
	"net/url"

	"github.com/alibaba/open-object/pkg/common"
	"github.com/alibaba/open-object/pkg/csi/s3minio"
	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type controllerServer struct {
	kubeClinet *kubernetes.Clientset
	*csicommon.DefaultControllerServer
}

func newControllerServer(d *csicommon.CSIDriver) *controllerServer {
	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		klog.Fatalf("Error building kubeconfig: %s", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
		kubeClinet:              kubeClient,
	}
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if len(req.GetName()) == 0 {
		return &csi.CreateVolumeResponse{}, status.Error(codes.InvalidArgument, "Name missing in request")
	}
	if req.GetVolumeCapabilities() == nil {
		return &csi.CreateVolumeResponse{}, status.Error(codes.InvalidArgument, "Volume Capabilities missing in request")
	}

	// get driver
	var err error
	driverName := getDriverName(req.GetParameters())
	if driverName == "" {
		return &csi.CreateVolumeResponse{}, status.Errorf(codes.InvalidArgument, "%s not found in storageclass parameters", common.ParamDriverName)
	}
	var driver Driver
	switch driverName {
	case s3minio.DriverName:
		driver, err = getMinIODriver(req.Secrets)
		if err != nil {
			return &csi.CreateVolumeResponse{}, status.Errorf(codes.Internal, "fail to get minio driver: %s", err.Error())
		}
	default:
		return &csi.CreateVolumeResponse{}, status.Errorf(codes.Internal, "unknown driver: %s", driverName)
	}

	// create volume
	return driver.CreateVolume(ctx, req)
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return &csi.DeleteVolumeResponse{}, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	pv, err := cs.kubeClinet.CoreV1().PersistentVolumes().Get(context.Background(), req.GetVolumeId(), metav1.GetOptions{})
	if err != nil {
		return &csi.DeleteVolumeResponse{}, fmt.Errorf("failed to get pv %s info: %s", req.GetVolumeId(), err.Error())
	}

	// get driver
	driverName := getDriverName(pv.Spec.CSI.VolumeAttributes)
	if driverName == "" {
		return &csi.DeleteVolumeResponse{}, status.Errorf(codes.InvalidArgument, "%s not found in pv %s attributes", common.ParamDriverName, pv.Name)
	}
	var driver Driver
	switch driverName {
	case s3minio.DriverName:
		driver, err = getMinIODriver(req.Secrets)
		if err != nil {
			return &csi.DeleteVolumeResponse{}, status.Errorf(codes.Internal, "fail to get minio driver: %s", err.Error())
		}
	default:
		return &csi.DeleteVolumeResponse{}, status.Errorf(codes.Internal, "unknown driver: %s", driverName)
	}

	return driver.DeleteVolume(ctx, req)
}

func (cs *controllerServer) ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error) {
	volumeID := req.GetVolumeId()
	if len(volumeID) == 0 {
		return &csi.ControllerExpandVolumeResponse{}, status.Error(codes.InvalidArgument, "Volume ID missing in request")
	}

	pv, err := cs.kubeClinet.CoreV1().PersistentVolumes().Get(context.Background(), req.GetVolumeId(), metav1.GetOptions{})
	if err != nil {
		return &csi.ControllerExpandVolumeResponse{}, fmt.Errorf("failed to get pv %s info: %s", req.GetVolumeId(), err.Error())
	}

	// get driver
	driverName := getDriverName(pv.Spec.CSI.VolumeAttributes)
	if driverName == "" {
		return &csi.ControllerExpandVolumeResponse{}, status.Errorf(codes.InvalidArgument, "%s not found in pv %s attributes", common.ParamDriverName, pv.Name)
	}
	var driver Driver
	switch driverName {
	case s3minio.DriverName:
		driver, err = getMinIODriver(req.Secrets)
		if err != nil {
			return &csi.ControllerExpandVolumeResponse{}, status.Errorf(codes.Internal, "fail to get minio driver: %s", err.Error())
		}
	default:
		return &csi.ControllerExpandVolumeResponse{}, status.Errorf(codes.Internal, "unknown driver: %s", driverName)
	}

	// expand volume
	return driver.ControllerExpandVolume(ctx, req)
}

func getDriverName(attr map[string]string) string {
	return attr[common.ParamDriverName]
}

func getMinIODriver(secrets map[string]string) (Driver, error) {
	endpoint, err := GetS3EndPoint(secrets[s3minio.SecretMinIOHost])
	if err != nil {
		return nil, err
	}
	klog.Infof("endpoint: %s", endpoint)

	cfg := &s3minio.S3Config{
		AK:       secrets[s3minio.SecretAK],
		SK:       secrets[s3minio.SecretSK],
		Region:   secrets[s3minio.SecretRegion],
		Endpoint: endpoint,
	}
	return s3minio.NewMinIODriver(cfg)
}

func GetS3EndPoint(s3host string) (string, error) {
	// endpoint
	endpoint := ""
	u, err := url.Parse(s3host)
	if err != nil {
		return "", err
	}
	scheme := u.Scheme
	host := u.Hostname()
	port := u.Port()
	// check if is ip
	addr := net.ParseIP(host)
	if addr != nil {
		// is ip
		endpoint = fmt.Sprintf("%s://%s:%s", scheme, addr.String(), port)
	} else {
		// is not ip
		IPs, err := net.LookupIP(host)
		if err != nil {
			return "", fmt.Errorf("fail to lookup %s: %s", host, err.Error())
		}
		if len(IPs) == 1 {
			endpoint = fmt.Sprintf("%s://%s:%s", scheme, IPs[0].String(), port)
		} else if len(IPs) > 1 {
			return "", fmt.Errorf("more than one ip found when lookup host %s: %v", host, IPs)
		} else {
			return "", fmt.Errorf("no ip found when lookup host %s", host)
		}
	}
	return endpoint, nil
}
