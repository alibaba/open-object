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

package version

import (
	"fmt"

	"github.com/alibaba/open-object/pkg/version"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	verbose bool
)

var Cmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of open-local",
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			fmt.Printf("%s\n", version.NameWithVersion(true))
		} else {
			fmt.Printf("%s\n", version.NameWithVersion(false))
		}
	},
}

func init() {
	addFlags(Cmd.Flags())
}

func addFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&verbose, "verbose", false, "show detailed version info")
}
