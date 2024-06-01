package processor

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Comcast/gots/v2/packet"
	"github.com/Comcast/gots/v2/pes"
	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
	"github.com/potterxu/tsanalyzer/tsutil/ts"
)

const (
	VbvName string = "vbv"

	Config_Vbv_Pids = "pids"
	Config_Vbv_Pcr  = "pcr"
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
`
	fmt.Printf(format,
		vbvInputFormats,
		vbvOutputFormats,
		Config_Vbv_Pids,
		Config_Vbv_Pcr)
}

func VbvHelpShort() {
	fmt.Printf("%v : calculate dts-pcr\n", VbvName)
}

type Record struct {
	Pid   int
	Index int64
	Pcr   int64
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
	pids map[int]bool
	pcr  int

	// internal
	accumulator ts.Accumulator
	curVbv      [ts.MAX_PID + 1]*VbvRecord
	lastPcr     *PcrRecord

	vbvs [ts.MAX_PID + 1][]*VbvRecord
}

func NewVbv(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &Vbv{
		accumulator: ts.NewAccumulator(),
		pids:        make(map[int]bool),
		lastPcr:     nil,
	}
	c.ICell = c
	c.Init(stopChan, config)

	var err error
	if pcrStr, ok := config[Config_Vbv_Pcr]; ok {
		if c.pcr, err = strconv.Atoi(pcrStr); err != nil {
			fmt.Println("[vbv] invalid pcr pid", pcrStr)
			return nil, errinfo.ErrInvalidCellConfig
		}
	} else {
		fmt.Println("[vbv] missing pcr pid")
		return nil, errinfo.ErrInvalidCellConfig
	}

	if pidsStr, ok := config[Config_Vbv_Pids]; ok {
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
	return c, nil
}

func (c *Vbv) Run() {
	defer c.StopCell()
	if !c.StartCell() {
		return
	}

	index := int64(0)
workLoop:
	for {
		unit, ok := c.GetInput()
		if !ok {
			break
		}
		switch data := unit.Data().(type) {
		case packet.Packet:
			if !c.processPcrPkt(data, index) {
				break workLoop
			}
			if !c.processVbvPkt(data, index) {
				break workLoop
			}
		default:
			fmt.Println("[vbv] invalid input format")
		}
		index++
	}

	c.showResult()
}

func (c *Vbv) processVbvPkt(pkt packet.Packet, index int64) bool {
	if err := pkt.CheckErrors(); err != nil {
		fmt.Println("[vbv] packet error", err)
		return false
	}
	pid := packet.Pid(&pkt)
	if len(c.pids) > 0 {
		if _, ok := c.pids[pid]; !ok {
			return true
		}
	}

	if c.curVbv[pid] == nil {
		if !pkt.PayloadUnitStartIndicator() {
			// wait for first payload unit start indicator
			return true
		} else {
			c.curVbv[pid] = &VbvRecord{
				Record: Record{
					Pid:   pid,
					Index: index,
				},
				EndIndex: index,
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
		c.curVbv[pid].Dts = -1
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
				c.interpolateVbvPcr(newPcr)
			}
			c.lastPcr = newPcr

		}
	}
	return true
}

func (c *Vbv) interpolateVbvPcr(pcr *PcrRecord) {
	for pid := 0; pid < ts.MAX_PID; pid++ {
		for i := len(c.vbvs[pid]) - 1; i >= 0; i-- {
			vbv := c.vbvs[pid][i]
			if vbv.EndIndex < c.lastPcr.Index {
				break
			}
			if vbv.EndIndex >= c.lastPcr.Index && vbv.EndIndex <= pcr.Index {
				fmt.Println("[vbv] interpolate", pid, vbv.Index, vbv.EndIndex, vbv.Dts, c.lastPcr.Pcr, pcr.Pcr)
				vbv.EndPcr = c.lastPcr.Pcr + (vbv.EndIndex-c.lastPcr.Index)*(pcr.Pcr-c.lastPcr.Pcr)/(pcr.Index-c.lastPcr.Index)
			}
		}
	}
}

func (c *Vbv) showResult() {
	for pid := 0; pid < ts.MAX_PID; pid++ {
		if len(c.vbvs[pid]) == 0 {
			continue
		}
		fmt.Printf("pid %v\n", pid)
		for _, vbv := range c.vbvs[pid] {
			fmt.Printf("  [%v, %v] %v -> %v\n", vbv.Index, vbv.EndIndex, vbv.Dts*300, vbv.EndPcr)
		}
	}
}
