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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/alibaba/open-object/cmd/connector"
	"github.com/alibaba/open-object/cmd/csi"
	"github.com/alibaba/open-object/cmd/version"
)

const (
	EnvLogLevel = "LogLevel"
	LogPanic    = "Panic"
	LogFatal    = "Fatal"
	LogError    = "Error"
	LogWarn     = "Warn"
	LogInfo     = "Info"
	LogDebug    = "Debug"
	LogTrace    = "Trace"
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
	logLevel := os.Getenv(EnvLogLevel)
	switch logLevel {
	case LogPanic:
		log.SetLevel(log.PanicLevel)
	case LogFatal:
		log.SetLevel(log.FatalLevel)
	case LogError:
		log.SetLevel(log.ErrorLevel)
	case LogWarn:
		log.SetLevel(log.WarnLevel)
	case LogInfo:
		log.SetLevel(log.InfoLevel)
	case LogDebug:
		log.SetLevel(log.DebugLevel)
	case LogTrace:
		log.SetLevel(log.TraceLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}
	addCommands()
}

func main() {
	log.Infof("Version: %s, Commit: %s", VERSION, COMMITID)
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
