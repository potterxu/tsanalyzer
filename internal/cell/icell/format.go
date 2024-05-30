package icell

import (
	"reflect"

	"github.com/Comcast/gots/v2/packet"
)

type Format string

// Format
const (
	BYTE_SLICE = "[]byte"
	STRING     = "string"
	TS_PACKET  = "ts_packet"
)

var (
	FormatToType = map[string]reflect.Type{
		BYTE_SLICE: reflect.TypeFor[[]byte](),
		STRING:     reflect.TypeFor[string](),
		TS_PACKET:  reflect.TypeFor[packet.Packet](),
	}
)
