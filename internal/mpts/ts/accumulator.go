package ts

import (
	"bytes"

	"github.com/Comcast/gots/v2"
	"github.com/Comcast/gots/v2/packet"
)

const (
	MAX_PID = 8191
)

type AccumulatorResult struct {
	Pid  int
	Data []byte
}

// Accumulator is used to accumulate ts packet
type Accumulator interface {
	// add ts packet to the accumulator,
	// the accumulator will check the incoming packet triggers the finish condition
	// if finish condition triggered
	// return the previously accumulated bytes
	// and the current packet will be count into next accumulating period
	Add(packet.Packet) (*AccumulatorResult, bool, error)

	// force reset the state of the accumulator
	Reset()
}

type AccumulatorFinish func(packet.Packet) (bool, error)
type accumulator struct {
	finish AccumulatorFinish

	payloads [MAX_PID + 1]*bytes.Buffer
}

// create a new Accumulator
// if the finish function is null, finish function checking PUSI will be used by default
func NewAccumulator(finish AccumulatorFinish) Accumulator {
	a := &accumulator{
		finish: defaultAccumulatorFinish,
	}
	if finish != nil {
		a.finish = finish
	}

	for pid := 0; pid <= MAX_PID; pid++ {
		a.payloads[pid] = &bytes.Buffer{}
	}

	return a
}

// default finish function will return true at PUSI=true packet
func defaultAccumulatorFinish(pkt packet.Packet) (bool, error) {
	if err := pkt.CheckErrors(); err != nil {
		return false, err
	}
	return pkt.PayloadUnitStartIndicator(), nil
}

func (a *accumulator) Add(pkt packet.Packet) (*AccumulatorResult, bool, error) {
	var result *AccumulatorResult = nil
	ready := false

	done, err := a.finish(pkt)
	if err != nil {
		a.Reset()
		return nil, ready, err
	}

	pid := packet.Pid(&pkt)

	if done && a.payloads[pid].Len() > 0 {
		result = &AccumulatorResult{
			Pid:  pid,
			Data: make([]byte, a.payloads[pid].Len()),
		}
		copy(result.Data, a.payloads[pid].Bytes())
		a.reset(pid)
		ready = true
	}
	pay, err := packet.Payload(&pkt)
	if err != nil {
		if err != gots.ErrNoPayload {
			a.reset(pid)
			return result, ready, err
		}
	} else {
		a.payloads[pid].Write(pay)
	}
	return result, ready, nil
}

func (a *accumulator) reset(pid int) {
	a.payloads[pid].Reset()
}
func (a *accumulator) Reset() {
	for pid := 0; pid <= MAX_PID; pid++ {
		a.reset(pid)
	}
}
