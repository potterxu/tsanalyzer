package converter

import (
	"bytes"
	"fmt"
	"io"
	"reflect"

	"github.com/Comcast/gots/v2/packet"
	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
	"golang.org/x/exp/maps"
)

const (
	BytesConverterName string = "bytes_converter"
)

var (
	bytesConverterOutputFormatMap = map[string]reflect.Type{
		"ts_packet": reflect.TypeFor[packet.Packet](),
	}
)

type BytesConverter struct {
	icell.Cell

	outputFormat reflect.Type

	remainedBytes []byte
}

func BytesConverterHelp() {
	BytesConverterHelpShort()
	format := `  Properties:
    %v: output type %v

`
	fmt.Printf(format, icell.CONFIG_output_type, maps.Keys(bytesConverterOutputFormatMap))
}

func BytesConverterHelpShort() {
	fmt.Printf("%s: convert byte array to type\n", BytesConverterName)
	fmt.Println("  ->cell: []byte")
	fmt.Println("  cell->:", maps.Keys(bytesConverterOutputFormatMap))
	fmt.Println("")
}

func NewBytesConverter(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &BytesConverter{
		remainedBytes: nil,
	}
	c.ICell = c
	c.Init(stopChan, config)

	if outputType, ok := config[icell.CONFIG_output_type]; ok {
		if outputFormat, ok := bytesConverterOutputFormatMap[outputType]; ok {
			c.outputFormat = outputFormat
		} else {
			fmt.Printf("%v=%v not supported for BytesConverter", icell.CONFIG_output_type, outputType)
			BytesConverterHelp()
			return nil, errinfo.ErrInvalidCellConfig
		}
	} else {
		fmt.Printf("%v not provided for BytesConverter\n", icell.CONFIG_output_type)
		BytesConverterHelp()
		return nil, errinfo.ErrInvalidCellConfig
	}
	return c, nil
}

func (c *BytesConverter) Run() {
	defer c.StopCell()
	if !c.StartCell() {
		return
	}

	for {
		unit, ok := c.GetInput()
		if !ok {
			break
		}
		var err error = nil
		switch data := unit.Data().(type) {
		case []byte:
			c.process(data)
		default:
			fmt.Printf("Invalid input type %v for BytesConverter", reflect.TypeOf(unit.Data()))
			err = errinfo.ErrInvalidUnitFormat
		}
		if err != nil {
			break
		}
	}
}

func (c *BytesConverter) process(buffer []byte) {
	data := make([]byte, 0)
	if c.remainedBytes != nil {
		data = append(data, c.remainedBytes...)
		c.remainedBytes = nil
	}
	data = append(data, buffer...)
	remain := len(data)

	switch c.outputFormat {
	case bytesConverterOutputFormatMap["ts_packet"]:

		reader := bytes.NewReader(data)
		var pkt packet.Packet
		for remain >= packet.PacketSize {
			if _, err := io.ReadFull(reader, pkt[:]); err != nil {
				break
			}
			remain -= packet.PacketSize
			c.PutOutput(icell.NewCellUnit(pkt))
		}
	default:
		// not support, drop the buffer
		remain = 0
	}

	if remain > 0 {
		c.remainedBytes = make([]byte, remain)
		copy(c.remainedBytes, data[len(data)-remain:])
	}
}
