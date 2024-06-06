package processor

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/Comcast/gots/v2/packet"
	"github.com/Comcast/gots/v2/pes"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
	"github.com/potterxu/tsanalyzer/tsutil/ts"
)

const (
	VbvName string = "vbv"

	config_vbv_pids = "pids"
	config_vbv_pcr  = "pcr"
	config_vbv_dir  = "dir"
	config_vbv_plot = "plot"
)

var (
	vbvInputFormats  []icell.Format = []icell.Format{icell.TS_PACKET}
	vbvOutputFormats []icell.Format = nil
)

func VbvHelp() {
	VbvHelpShort()
	format := `	IO:
	  ->cell: %v
	  cell->: %v
	Properties:
	  %v: select pids to process, split by ","
	  %v: pcr pid
	  %v: output directory
	  %v: plot the result
`
	fmt.Printf(format,
		vbvInputFormats,
		vbvOutputFormats,
		config_vbv_pids,
		config_vbv_pcr,
		config_vbv_dir,
		config_vbv_plot,
	)
}

func VbvHelpShort() {
	fmt.Printf("%v : calculate dts-pcr\n", VbvName)
}

type Record struct {
	Pid    int
	Index  int64
	Pcr    int64
	Packet packet.Packet
}

type VbvRecord struct {
	Record
	EndIndex int64
	EndPcr   int64
	Dts      int64
}

type PcrRecord struct {
	Record
}

type Vbv struct {
	icell.Cell

	// config
	pids      map[int]bool
	pcr       int
	outputDir string
	plot      bool

	// internal
	accumulator    ts.Accumulator
	pendingRecords []*Record
	curVbv         [ts.MAX_PID + 1]*VbvRecord
	lastPcr        *PcrRecord

	vbvs [ts.MAX_PID + 1][]*VbvRecord
}

func NewVbv(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &Vbv{
		accumulator:    ts.NewAccumulator(),
		pids:           make(map[int]bool),
		lastPcr:        nil,
		pendingRecords: make([]*Record, 0),
	}
	c.ICell = c
	c.Init(stopChan, config)

	var err error
	if pcrStr, ok := config[config_vbv_pcr]; ok {
		if c.pcr, err = strconv.Atoi(pcrStr); err != nil {
			fmt.Println("[vbv] invalid pcr pid", pcrStr)
			return nil, errinfo.ErrInvalidCellConfig
		}
	} else {
		fmt.Println("[vbv] missing pcr pid")
		return nil, errinfo.ErrInvalidCellConfig
	}

	if pidsStr, ok := config[config_vbv_pids]; ok {
		pidStrs := strings.Split(pidsStr, ",")
		if len(pidStrs) < 1 {
			fmt.Println("[vbv] missing processing pids")
			return nil, errinfo.ErrInvalidCellConfig
		}
		for _, pidStr := range pidStrs {
			var pid int
			if pid, err = strconv.Atoi(pidStr); err != nil {
				fmt.Println("[vbv] invalid processing pid", pidStr)
				return nil, errinfo.ErrInvalidCellConfig
			}
			c.pids[pid] = true
			c.vbvs[pid] = make([]*VbvRecord, 0)
			c.curVbv[pid] = nil
		}
	} else {
		fmt.Println("[vbv] missing processing pids")
		return nil, errinfo.ErrInvalidCellConfig
	}

	if dir, ok := config[config_vbv_dir]; ok {
		c.outputDir = dir
	} else {
		fmt.Println("[vbv] output to console")
	}

	if plot, ok := config[config_vbv_plot]; ok {
		if plot == "true" {
			c.plot = true
			fmt.Println("[vbv] plot the result")
		}
	}

	return c, nil
}

func (c *Vbv) Run() {
	defer c.OnCellFinished()
	c.OnCellStart()
	defer c.OnCellFinished()

	index := int64(0)
workLoop:
	for {
		unit, ok := c.GetInput()
		if !ok {
			break
		}
		switch reflect.TypeOf(unit.Data()) {
		case icell.FormatToType[icell.TS_PACKET]:
			data := unit.Data().(packet.Packet)
			if !c.processPkt(data, index) {
				break workLoop
			}
			if !c.processPcrPkt(data, index) {
				break workLoop
			}
		default:
			fmt.Println("[vbv] invalid input format")
		}
		index++
	}

	c.showResult()
}

func (c *Vbv) processPkt(pkt packet.Packet, index int64) bool {
	if err := pkt.CheckErrors(); err != nil {
		fmt.Println("[vbv] packet error", err)
		return false
	}
	if _, ok := c.pids[packet.Pid(&pkt)]; !ok {
		return true
	}
	c.pendingRecords = append(c.pendingRecords, &Record{
		Pid:    packet.Pid(&pkt),
		Index:  index,
		Pcr:    -1,
		Packet: pkt,
	})

	return true
}

func (c *Vbv) processPcrPkt(pkt packet.Packet, index int64) bool {
	if err := pkt.CheckErrors(); err != nil {
		fmt.Println("[vbv] packet error", err)
		return false
	}
	if packet.Pid(&pkt) == c.pcr && packet.ContainsAdaptationField(&pkt) {
		af := packet.AdaptationField(pkt)
		hasPcr, err := af.HasPCR()
		if err != nil {
			fmt.Println(err)
			return false
		}
		if hasPcr {
			pcr, err := af.PCR()
			if err != nil {
				fmt.Println(err)
				return false
			}

			newPcr := &PcrRecord{
				Record: Record{
					Pid:   c.pcr,
					Index: index,
					Pcr:   int64(pcr),
				},
			}

			// interpolate vbv
			if c.lastPcr != nil {
				for i, record := range c.pendingRecords {
					c.pendingRecords[i].Pcr = c.lastPcr.Pcr + (record.Index-c.lastPcr.Index)*(newPcr.Pcr-c.lastPcr.Pcr)/(newPcr.Index-c.lastPcr.Index)
				}
				// all the pending packets have pcr now
				// ready to process
				if !c.processPendingPkts() {
					return false
				}
			}
			c.lastPcr = newPcr

		}
	}
	return true
}

func (c *Vbv) processPendingPkts() bool {
	for _, record := range c.pendingRecords {
		pkt := record.Packet
		index := record.Index
		pid := record.Pid
		pcr := record.Pcr

		if c.curVbv[pid] == nil {
			if !pkt.PayloadUnitStartIndicator() {
				// wait for first payload unit start indicator
				continue
			} else {
				c.curVbv[pid] = &VbvRecord{
					Record: Record{
						Pid:   pid,
						Index: index,
					},
					EndIndex: index,
					Dts:      -1,
				}
			}
		}

		result, ready, err := c.accumulator.Add(pkt)
		if err != nil {
			fmt.Println("[vbv] accumulator error", err)
			return false
		}
		if ready {
			pes, err := pes.NewPESHeader(result.Data)
			if err != nil {
				fmt.Println("[vbv] pes error", err)
				return false
			}
			if pes.HasDTS() {
				c.curVbv[pid].Dts = int64(pes.DTS())
			} else if pes.HasPTS() {
				c.curVbv[pid].Dts = int64(pes.PTS())
			}
			c.vbvs[pid] = append(c.vbvs[pid], c.curVbv[pid])
			c.curVbv[pid] = &VbvRecord{
				Record: Record{
					Pid:   pid,
					Index: index,
					Pcr:   -1,
				},
				EndIndex: index,
				EndPcr:   -1,
				Dts:      -1,
			}
		}
		c.curVbv[pid].EndIndex = index
		c.curVbv[pid].EndPcr = pcr
	}
	c.pendingRecords = c.pendingRecords[:0]
	return true
}

func (c *Vbv) showResult() {
	if c.outputDir != "" {
		if err := os.MkdirAll(c.outputDir, 0755); err != nil {
			fmt.Println(err)
			return
		}
	}
	for pid := 0; pid < ts.MAX_PID; pid++ {
		if len(c.vbvs[pid]) == 0 {
			continue
		}
		var writer *bufio.Writer
		if c.outputDir != "" {
			filename := path.Join(c.outputDir, fmt.Sprintf("vbv_%v.txt", pid))
			file, err := os.Create(filename)
			if err != nil {
				fmt.Println(err)
				continue
			}
			defer file.Close()
			writer = bufio.NewWriter(file)
		} else {
			writer = bufio.NewWriter(os.Stdout)
		}

		var plotWriter io.Writer
		var lineChart *charts.Line
		var yAxis []opts.LineData
		if c.plot {
			plotFilename := path.Join(c.outputDir, fmt.Sprintf("vbv_%v.html", pid))
			plotFile, err := os.Create(plotFilename)
			if err != nil {
				fmt.Println(err)
				continue
			}
			defer plotFile.Close()
			plotWriter = plotFile

			lineChart = charts.NewLine()
			lineChart.SetGlobalOptions(
				charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWesteros}),
				charts.WithTitleOpts(opts.Title{
					Title: "VBV for pid " + strconv.Itoa(pid),
				}),
			)
			x := make([]string, len(c.vbvs[pid]))
			for i := range x {
				x[i] = strconv.Itoa(i)
			}
			lineChart.SetXAxis(x)
		}

		if _, err := writer.WriteString(fmt.Sprintf("pid %v\n", pid)); err != nil {
			fmt.Println(err)
			continue
		}
		if _, err := writer.WriteString("  [ index , endIndex ] dts -> pcr vbv\n"); err != nil {
			fmt.Println(err)
			continue
		}
		for _, vbv := range c.vbvs[pid] {
			if c.plot {
				yAxis = append(yAxis, opts.LineData{
					Value: vbv.Dts - vbv.EndPcr/300,
				})
			}
			if _, err := writer.WriteString(fmt.Sprintf("  [ %v , %v ] %v -> %v %v\n", vbv.Index, vbv.EndIndex, vbv.Dts, vbv.EndPcr/300, vbv.Dts-vbv.EndPcr/300)); err != nil {
				fmt.Println(err)
				break
			}
		}
		writer.Flush()

		if c.plot {
			lineChart.AddSeries("vbv", yAxis).
				SetSeriesOptions(
					charts.WithMarkPointNameTypeItemOpts(
						opts.MarkPointNameTypeItem{Name: "max", Type: "max"},
						opts.MarkPointNameTypeItem{Name: "min", Type: "min"},
					),
				)
			if err := lineChart.Render(plotWriter); err != nil {
				fmt.Println(err)
			}
		}
	}
}
