package converter

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"slices"

	"github.com/Comcast/gots/v2/packet"
	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

const (
	BytesConverterName string = "bytes_converter"

	config_bytesconverter_outputformat string = "output_format"
)

var (
	bytesConverterInputFormats  = []icell.Format{icell.BYTE_SLICE}
	bytesConverterOutputFormats = []icell.Format{icell.TS_PACKET}
)

type BytesConverter struct {
	icell.Cell

	outputFormat icell.Format

	remainedBytes []byte
}

func BytesConverterHelp() {
	BytesConverterHelpShort()
	format := `	IO:
	  ->cell: %v
	  cell->: %v
	Properties:
	  %v: output format %v
`
	fmt.Printf(format,
		bytesConverterInputFormats,
		bytesConverterOutputFormats,
		config_bytesconverter_outputformat,
		bytesConverterInputFormats)
}

func BytesConverterHelpShort() {
	fmt.Printf("%s: convert byte array to type\n", BytesConverterName)
}

func NewBytesConverter(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &BytesConverter{
		remainedBytes: nil,
	}
	c.ICell = c
	c.Init(stopChan, config)

	if of, ok := config[config_bytesconverter_outputformat]; ok {
		outputFormat := icell.Format(of)
		if slices.Contains(bytesConverterOutputFormats, outputFormat) {
			c.outputFormat = outputFormat
		} else {
			fmt.Printf("%v=%v not supported for BytesConverter", config_bytesconverter_outputformat, outputFormat)
			BytesConverterHelp()
			return nil, errinfo.ErrInvalidCellConfig
		}
	} else {
		fmt.Printf("%v not provided for BytesConverter\n", config_bytesconverter_outputformat)
		BytesConverterHelp()
		return nil, errinfo.ErrInvalidCellConfig
	}
	return c, nil
}

func (c *BytesConverter) Run() {
	c.OnCellStart()
	defer c.OnCellFinished()

	for {
		unit, ok := c.GetInput()
		if !ok {
			break
		}
		var err error = nil
		switch reflect.TypeOf(unit.Data()) {
		case icell.FormatToType[icell.BYTE_SLICE]:
			c.process(unit.Data().([]byte))
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
	case icell.TS_PACKET:
		reader := bytes.NewReader(data)
		var pkt packet.Packet
		for remain >= packet.PacketSize {
			if _, err := io.ReadFull(reader, pkt[:]); err != nil {
				break
			}
			if err := pkt.CheckErrors(); err != nil {
				fmt.Println(err)
				break
			}
			remain -= packet.PacketSize
			c.PutOutput(icell.NewCellUnit(pkt, icell.TS_PACKET))
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
