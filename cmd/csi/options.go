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
	"strings"

	"github.com/alibaba/open-object/pkg/common"
	"github.com/alibaba/open-object/pkg/csi/s3minio"
	"github.com/spf13/pflag"
	cliflag "k8s.io/component-base/cli/flag"
)

type csiOption struct {
	Endpoint     string
	NodeID       string
	Driver       string
	Master       string
	KubeConfig   string
	FeatureGates map[string]bool
}

func (option *csiOption) addFlags(fs *pflag.FlagSet) {
	fs.StringVar(&option.Endpoint, "endpoint", common.DefaultEndpoint, "csi endpoint")
	fs.StringVar(&option.NodeID, "nodeID", "", "node id")
	fs.StringVar(&option.Driver, "driver", common.DefaultDriverName, "csi driver name")
	fs.StringVar(&option.KubeConfig, "kubeconfig", "", "Path to the kubeconfig file to use")
	fs.StringVar(&option.Master, "master", "", "URL/IP for master.")
	fs.Var(cliflag.NewMapStringBool(&option.FeatureGates), "feature-gates", "A set of key=value pairs that describe feature gates for alpha/experimental features. "+
		"Options are:\n"+strings.Join(s3minio.DefaultFeatureGate.KnownFeatures(), "\n"))
}
