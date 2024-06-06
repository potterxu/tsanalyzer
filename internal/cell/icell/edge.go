package icell

import (
	"reflect"

	"github.com/google/uuid"
)

const (
	EDGE_BUFFER int = 10
)

type Edge struct {
	id string

	src ICell
	dst ICell

	unitType reflect.Type
	channel  chan CellUnit
	open     bool
}

func newEdge(src, dst ICell) (*Edge, error) {
	e := &Edge{
		id:      uuid.NewString(),
		src:     src,
		dst:     dst,
		channel: make(chan CellUnit, EDGE_BUFFER),
		open:    true,
	}
	return e, nil
}

func (e *Edge) Id() string {
	return e.id
}

func (e *Edge) Src() ICell {
	return e.src
}

func (e *Edge) Dst() ICell {
	return e.dst
}

func (e *Edge) UnitType() reflect.Type {
	return e.unitType
}

func (e *Edge) Channel() chan CellUnit {
	return e.channel
}

func (e *Edge) Close() {
	if e.open {
		close(e.channel)
	}
}
