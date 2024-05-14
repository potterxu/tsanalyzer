/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"

	"github.com/32bitkid/bitreader"
	"github.com/potterxu/algo/search"
	"github.com/potterxu/mpeg/pes"
	"github.com/potterxu/mpeg/ts"
	"github.com/spf13/cobra"
)

var (
	pcrPID    uint32
	streamPID uint32
)

type PesInfo struct {
	pid        uint32
	length     uint64
	dts        int64
	startIndex uint64
	startPcr   int64
	startVbv   int64
	endIndex   uint64
	endPcr     int64
	endVbv     int64
}

func (p PesInfo) GetHeader() []string {
	return []string{"pid", "length", "dts", "startIndex", "startPcr", "startVbv", "endIndex", "endPcr", "endVbv"}
}

func (p PesInfo) GetRow() []string {
	return []string{fmt.Sprintf("%d", p.pid), fmt.Sprintf("%d", p.length), fmt.Sprintf("%d", p.dts), fmt.Sprintf("%d", p.startIndex), fmt.Sprintf("%d", p.startPcr), fmt.Sprintf("%d", p.startVbv), fmt.Sprintf("%d", p.endIndex), fmt.Sprintf("%d", p.endPcr), fmt.Sprintf("%d", p.endVbv)}
}

// vbvCmd represents the vbv command
var vbvCmd = &cobra.Command{
	Use:   "vbv <filename>",
	Short: "analyze timing info of stream in ts",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Help()
			return
		}
		filename := args[0]
		file, err := os.Open(filename)
		if err != nil {
			panic(err)
		}
		defer file.Close()
		tsReader := bitreader.NewReader(file)
		currentPes := make([]byte, 0)

		pesInfos := make([]PesInfo, 0)
		pcrIdxs := make([]uint64, 0)
		pcrMap := make(map[uint64]int64)

		var index uint64 = 0
		for {
			pkt, err := ts.NewPacket(tsReader)
			if err != nil {
				break
			}
			if pkt.PID == pcrPID && pkt.AdaptationFieldControl&0b10 != 0 && pkt.AdaptationField.PCRFlag {
				pcrIdxs = append(pcrIdxs, index)
				pcrMap[index] = int64(pkt.AdaptationField.PCR)
			}
			if pkt.PID == streamPID {
				if pkt.PayloadUnitStartIndicator {
					if len(currentPes) > 0 {
						pesReader := bitreader.NewReader(bytes.NewReader(currentPes))
						pes, err := pes.NewPacket(pesReader)
						currentPes = make([]byte, 0)
						if err != nil {
							fmt.Println(err)
							continue
						}
						dts := int64(0)
						if pes.Header.PtsDtsFlags == 0b11 {
							dts = int64(pes.Header.DecodingTimeStamp)
						} else if pes.Header.PtsDtsFlags == 0b10 {
							dts = int64(pes.Header.PresentationTimeStamp)
						}
						// convert to 27M
						dts *= 300
						pesInfos = append(pesInfos, PesInfo{startIndex: index, endIndex: index, startPcr: 0, endPcr: 0, dts: dts, pid: pkt.PID, length: 0})
					}
				}
				if len(pesInfos) > 0 {
					pesInfos[len(pesInfos)-1].endIndex = index
					pesInfos[len(pesInfos)-1].length += uint64(len(pkt.Payload))
				}
				currentPes = append(currentPes, pkt.Payload...)
			}
			index++
		}

		// calculate pcr for peses
		interpolatePcr := func(pcrMap map[uint64]int64, pcrIdxs []uint64, idx uint64) int64 {
			lbound, err1 := search.LowerBound(pcrIdxs, idx)
			ubound, err2 := search.UpperBound(pcrIdxs, idx)
			if err1 != nil {
				// no pcr index equal or larger than idx
				return -1
			}
			if pcrIdxs[lbound] == idx {
				// exact match
				return pcrMap[idx]
			}
			// no exact match, need to interpolate
			if err2 != nil || lbound == 0 {
				// no pcr index larger than or smaller than idx
				return -1
			}
			smallerIdx := pcrIdxs[lbound-1]
			largerIdx := pcrIdxs[ubound]
			smallerPcr := pcrMap[uint64(smallerIdx)]
			largerPcr := pcrMap[uint64(largerIdx)]
			return smallerPcr + (largerPcr-smallerPcr)*int64(idx-smallerIdx)/int64(largerIdx-smallerIdx)
		}

		for i, _ := range pesInfos {
			pesInfos[i].startPcr = interpolatePcr(pcrMap, pcrIdxs, pesInfos[i].startIndex)
			pesInfos[i].endPcr = interpolatePcr(pcrMap, pcrIdxs, pesInfos[i].endIndex)
			pesInfos[i].startVbv = pesInfos[i].dts - pesInfos[i].startPcr
			pesInfos[i].endVbv = pesInfos[i].dts - pesInfos[i].endPcr
		}

		// print to file
		if len(pesInfos) > 0 {
			inputPath, _ := filepath.Abs(filename)
			inputDir := filepath.Dir(inputPath)

			logDir := filepath.Join(inputDir, filename+"_log")
			if _, err := os.Stat(logDir); os.IsNotExist(err) {
				os.Mkdir(logDir, 0755)
			}
			logPath := filepath.Join(logDir, fmt.Sprintf("%v_vbv.csv", streamPID))
			outF, err := os.Create(logPath)
			if err != nil {
				panic(err)
			}
			defer outF.Close()
			writer := csv.NewWriter(outF)
			writer.Write(pesInfos[0].GetHeader())

			for _, pes := range pesInfos {
				writer.Write(pes.GetRow())
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(vbvCmd)
	vbvCmd.PersistentFlags().Uint32VarP(&pcrPID, "pcr", "p", 32, "pcr pid")
	vbvCmd.PersistentFlags().Uint32VarP(&streamPID, "stream", "s", 32, "stream pid")
}
