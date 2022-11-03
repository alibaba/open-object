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

package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/klog/v2"

	"github.com/alibaba/open-object/cmd/connector"
	"github.com/alibaba/open-object/cmd/csi"
	"github.com/alibaba/open-object/cmd/version"
)

var (
	MainCmd = &cobra.Command{
		Use: "open-object",
	}
	VERSION  string = ""
	COMMITID string = ""
)

func init() {
	flag.Parse()
	MainCmd.SetGlobalNormalizationFunc(wordSepNormalizeFunc)
	MainCmd.DisableAutoGenTag = true
	addCommands()
}

func main() {
	klog.Infof("Version: %s, Commit: %s", VERSION, COMMITID)
	if err := MainCmd.Execute(); err != nil {
		fmt.Printf("open-object start error: %+v\n", err)
		os.Exit(1)
	}
}

func addCommands() {
	MainCmd.AddCommand(
		csi.Cmd,
		version.Cmd,
		connector.Cmd,
		// doc.Cmd.Cmd,
	)
}

// wordSepNormalizeFunc changes all flags that contain "_" separators
func wordSepNormalizeFunc(f *pflag.FlagSet, name string) pflag.NormalizedName {
	if strings.Contains(name, "_") {
		return pflag.NormalizedName(strings.Replace(name, "_", "-", -1))
	}
	return pflag.NormalizedName(name)
}
