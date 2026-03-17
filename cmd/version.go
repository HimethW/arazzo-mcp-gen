/*
 * Copyright (c) 2025, WSO2 LLC. (https://www.wso2.com).
 *
 * WSO2 LLC. licenses this file to you under the Apache License,
 * Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/wso2/arazzo-mcp-gen/cmd.Version=v1.2.3"
//
// Falls back to "dev" when running with `go run`.
var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the arazzo-mcp-gen version",
	Long:  `Print the current version of arazzo-mcp-gen along with build information.`,
	Run: func(cmd *cobra.Command, args []string) {
		short, _ := cmd.Flags().GetBool("short")
		if short {
			fmt.Println(Version)
			return
		}

		fmt.Printf("arazzo-mcp-gen %s\n", Version)
		fmt.Printf("  Go:       %s\n", runtime.Version())
		fmt.Printf("  OS/Arch:  %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	versionCmd.Flags().Bool("short", false, "Print only the version number")
	rootCmd.AddCommand(versionCmd)
}
