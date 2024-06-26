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
	plot       bool
)

// vbvCmd represents the vbv command
var vbvCmd = &cobra.Command{
	Use:   "vbv <filename>",
	Short: "Calculate for vbv, alias for pipe [file_reader ! bytes_converter ! vbv]",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			if err := cmd.Help(); err != nil {
				fmt.Println(err)
			}
			return
		}
		filename := args[0]
		pipe := fmt.Sprintf("file_reader name=%v ! bytes_converter output_format=ts_packet ! vbv pcr=%v pids=%v dir=%v plot=%v",
			filename, pcrPID, streamPIDs, fmt.Sprintf("%v.log", filename), plot)
		pipeArgs := strings.Split(pipe, " ")
		pipeCmd.Run(nil, pipeArgs)
	},
}

func init() {
	rootCmd.AddCommand(vbvCmd)
	vbvCmd.PersistentFlags().Uint32VarP(&pcrPID, "pcr", "p", 32, "pcr pid")
	vbvCmd.PersistentFlags().StringVarP(&streamPIDs, "streams", "s", "32", "stream pids split by \",\"")
	vbvCmd.PersistentFlags().BoolVar(&plot, "plot", false, "plot the results")
}
