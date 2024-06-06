/*
Copyright © 2024 Potter XU <xujingchuan1995@gmail.com>

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
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// capCmd represents the cap command
var capCmd = &cobra.Command{
	Use:   "cap <net0> <239.1.1.1:1000>",
	Short: "Capture network streams, alias for pipe [mcast_reader ! file_writer]",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
			}
			return
		}

		interfaceName := args[0]
		address := args[1]

		reader := ""
		switch streamType {
		case "ts":
			reader = fmt.Sprintf("mcast_reader intf=%v addr=%v", interfaceName, address)
		default:
			fmt.Printf("Invalid stream type %v\n", streamType)
			return
		}
		writer := fmt.Sprintf("file_writer name=%v", filename)

		pipe := fmt.Sprintf("%v ! %v", reader, writer)
		pipeArgs := strings.Split(pipe, " ")
		pipeCmd.Run(nil, pipeArgs)
	},
}

var (
	streamType string
	filename   string
	// duration   int
)

func init() {
	rootCmd.AddCommand(capCmd)

	capCmd.Flags().StringVarP(&streamType, "type", "t", "ts", "stream type [ts,]")
	capCmd.Flags().StringVarP(&filename, "output", "o", "output.ts", "output filename")
	// capCmd.Flags().IntVarP(&duration, "duration", "d", -1, "capture duration in seconds")
}
