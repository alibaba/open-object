/*
Copyright © 2021 Alibaba Group Holding Ltd.

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
	"github.com/alibaba/open-object/pkg/version"
	"github.com/container-storage-interface/spec/lib/go/csi"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"golang.org/x/net/context"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type Driver interface {
	CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error)
	DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error)
	ControllerExpandVolume(ctx context.Context, req *csi.ControllerExpandVolumeRequest) (*csi.ControllerExpandVolumeResponse, error)
	NodeExpandVolume(ctx context.Context, req *csi.NodeExpandVolumeRequest) (
		*csi.NodeExpandVolumeResponse, error)
	NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error)
	NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error)
	NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error)
	NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error)
	NodeGetVolumeStats(ctx context.Context, req *csi.NodeGetVolumeStatsRequest) (*csi.NodeGetVolumeStatsResponse, error)
}

type FuseDriver struct {
	driver   *csicommon.CSIDriver
	endpoint string

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer
}

// NewFuseDriver initializes the driver
func NewFuseDriver(nodeID, endpoint, driverName string, kubeClient *kubernetes.Clientset) (*FuseDriver, error) {
	driver := csicommon.NewCSIDriver(driverName, version.Version, nodeID)
	if driver == nil {
		klog.Fatalln("Failed to initialize CSI Driver.")
	}

	s3Driver := &FuseDriver{
		endpoint: endpoint,
		driver:   driver,
		ids:      newIdentityServer(driver),
		cs:       newControllerServer(driver),
		ns:       newNodeServer(driver),
	}
	return s3Driver, nil
}

func (s3 *FuseDriver) Run() {
	// Initialize default library driver
	s3.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{
		csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME,
		csi.ControllerServiceCapability_RPC_EXPAND_VOLUME,
	})
	s3.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{
		csi.VolumeCapability_AccessMode_MULTI_NODE_MULTI_WRITER})

	s := csicommon.NewNonBlockingGRPCServer()
	s.Start(s3.endpoint, s3.ids, s3.cs, s3.ns)
	s.Wait()
}
