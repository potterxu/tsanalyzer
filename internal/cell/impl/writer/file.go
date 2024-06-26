package writer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/Comcast/gots/v2/packet"
	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

const (
	FileWriterName string = "file_writer"

	config_filewriter_name string = "name"
)

var (
	fileWriterInputFormats  []icell.Format = []icell.Format{icell.BYTE_SLICE, icell.STRING, icell.TS_PACKET}
	fileWriterOutputFormats []icell.Format = nil
)

type FileWriter struct {
	icell.Cell

	filename string
}

func FileWriterHelp() {
	FileWriterHelpShort()
	format := `	IO:
	  ->cell: %v
	  cell->: %v
	Properties:
	  %v: filename to write to
`
	fmt.Printf(format,
		fileWriterInputFormats,
		fileWriterOutputFormats,
		config_filewriter_name)
}

func FileWriterHelpShort() {
	fmt.Printf("%v : write content to file\n", FileWriterName)
}

func NewFileWriter(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &FileWriter{}
	c.ICell = c
	c.Init(stopChan, config)

	if filename, ok := config[config_filewriter_name]; ok {
		c.filename = filename
	} else {
		fmt.Println("file name not provided for FileWriter")
		FileWriterHelp()
		return nil, errinfo.ErrInvalidCellConfig
	}
	return c, nil
}

func (c *FileWriter) Run() {
	c.OnCellStart()
	defer c.OnCellFinished()

	file, err := os.Create(c.filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()
	for {
		unit, ok := c.GetInput()
		if !ok {
			break
		}
		var err error
		switch reflect.TypeOf(unit.Data()) {
		case icell.FormatToType[icell.BYTE_SLICE]:
			err = writeBytes(writer, unit.Data().([]byte))
		case icell.FormatToType[icell.STRING]:
			err = writeBytes(writer, []byte(unit.Data().(string)))
		case icell.FormatToType[icell.TS_PACKET]:
			data := unit.Data().(packet.Packet)
			bytes := []byte(data[:])
			err = writeBytes(writer, bytes)
		default:
			fmt.Printf("Invalid input type %v for FileWriter", reflect.TypeOf(unit.Data()))
			err = errinfo.ErrInvalidUnitFormat
		}
		if err != nil {
			break
		}
	}

}

func writeBytes(w io.Writer, data []byte) error {
	if _, err := w.Write(data); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}
