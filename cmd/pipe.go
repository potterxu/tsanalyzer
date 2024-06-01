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
	"strings"

	"github.com/potterxu/tsanalyzer/internal/cell"
	"github.com/potterxu/tsanalyzer/internal/graph"
	"github.com/spf13/cobra"
)

var (
	pipeFullHelpFlag bool   = false
	pipeListCellFlag bool   = false
	pipeCellHelp     string = ""
)

// pipeCmd represents the pipe command
var pipeCmd = &cobra.Command{
	Use:   "pipe",
	Short: "Use pipeline to process stream",

	Run: func(cmd *cobra.Command, args []string) {
		runPipe(args)
	},
}

func init() {
	rootCmd.AddCommand(pipeCmd)
	pipeCmd.SetHelpFunc(pipeCmd.PersistentPreRun)
	pipeCmd.PersistentFlags().BoolVarP(&pipeFullHelpFlag, "full", "f", false, "full help for cells")
	pipeCmd.PersistentFlags().BoolVarP(&pipeListCellFlag, "list", "l", false, "list all cells")
	pipeCmd.PersistentFlags().StringVarP(&pipeCellHelp, "cell", "c", "", "help for specific cell")
}

func runPipe(args []string) {
	if pipeFullHelpFlag {
		cell.Help()
		return
	}

	if pipeListCellFlag || args == nil || len(args) == 0 {
		cell.PrintCells()
		return
	}

	if len(pipeCellHelp) > 0 {
		cell.CellHelper(pipeCellHelp)
		return
	}

	g, err := graph.NewGraph(strings.Join(args, " "))
	if err != nil {
		panic(err)
	}
	g.Run()
}
