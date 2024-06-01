package ts

import (
	"bytes"

	"github.com/Comcast/gots/v2"
	"github.com/Comcast/gots/v2/packet"
)

const (
	MAX_PID = 8191
)

/* AccumulatorResult is the result of the accumulator
 * Pid: the pid of the accumulated data
 * Data: the accumulated data
 */
type AccumulatorResult struct {
	Pid  int
	Data []byte
}

// Accumulator is used to accumulate ts packet
type Accumulator interface {
	/* Add a packet to the accumulator
	 * return the accumulated data if a full pes packet is accumulated
	 * return whether the accumulated data is ready
	 * return error if any
	 */
	Add(packet.Packet) (*AccumulatorResult, bool, error)

	// force reset the state of the accumulator
	Reset()
}

type accumulator struct {
	payloads [MAX_PID + 1]*bytes.Buffer
}

func NewAccumulator() Accumulator {
	a := &accumulator{}

	for pid := 0; pid <= MAX_PID; pid++ {
		a.payloads[pid] = &bytes.Buffer{}
	}

	return a
}

func (a *accumulator) Add(pkt packet.Packet) (*AccumulatorResult, bool, error) {
	var result *AccumulatorResult = nil
	ready := false

	if err := pkt.CheckErrors(); err != nil {
		a.Reset()
		return nil, ready, err
	}

	done := pkt.PayloadUnitStartIndicator()
	pid := packet.Pid(&pkt)
	if done && a.payloads[pid].Len() > 0 {
		result = a.get(pid)
		ready = true
	}
	if err := a.add(pkt); err != nil {
		return result, ready, err
	}
	return result, ready, nil
}

func (a *accumulator) Reset() {
	for pid := 0; pid <= MAX_PID; pid++ {
		a.reset(pid)
	}
}

func (a *accumulator) reset(pid int) {
	a.payloads[pid].Reset()
}

func (a *accumulator) add(pkt packet.Packet) error {
	pid := packet.Pid(&pkt)
	pay, err := packet.Payload(&pkt)
	if err != nil {
		if err != gots.ErrNoPayload {
			a.reset(pid)
			return err
		}
	} else {
		a.payloads[pid].Write(pay)
	}
	return nil
}

func (a *accumulator) get(pid int) *AccumulatorResult {
	result := &AccumulatorResult{
		Pid:  pid,
		Data: make([]byte, a.payloads[pid].Len()),
	}
	copy(result.Data, a.payloads[pid].Bytes())
	a.reset(pid)
	return result
}
