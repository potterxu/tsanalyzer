/*
Copyright © 2024 Potter XU <xujingchuan1995@gmail.com>
*/
package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	pcrPID     uint32
	streamPIDs string
)

// vbvCmd represents the vbv command
var vbvCmd = &cobra.Command{
	Use: `vbv -p pcrpid -s streampids <filename>
	pcrpid is the pcr pid
	streampids is the stream pids, split by ","`,
	Short: "alias for pipe [file_reader ! bytes_converter ! vbv]",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
			}
			return
		}
		filename := args[0]
		pipe := fmt.Sprintf("file_reader name=%v ! bytes_converter output_format=ts_packet ! vbv pcr=%v pids=%v dir=%v",
			filename, pcrPID, streamPIDs, fmt.Sprintf("%v.log", filename))
		pipeArgs := strings.Split(pipe, " ")
		pipeCmd.Run(nil, pipeArgs)
	},
}

func init() {
	rootCmd.AddCommand(vbvCmd)
	vbvCmd.PersistentFlags().Uint32VarP(&pcrPID, "pcr", "p", 32, "pcr pid")
	vbvCmd.PersistentFlags().StringVarP(&streamPIDs, "stream", "s", "32", "stream pid")
}
