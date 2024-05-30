package writer

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"reflect"

	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

type FileWriter struct {
	icell.Cell

	filename string
}

func FileWriterHelp() {
	FileWriterHelpShort()
	format := ` Properties:
    %v: filename to write to

`
	fmt.Printf(format, icell.CONFIG_name)
}

func FileWriterHelpShort() {
	fmt.Printf("(filewriter: write content to file)\n")
}

func NewFileWriter(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &FileWriter{}
	c.ICell = c
	c.Init(stopChan, config)

	if filename, ok := config[icell.CONFIG_name]; ok {
		c.filename = filename
	} else {
		fmt.Println("file name not provided for FileWriter")
		FileWriterHelp()
		return nil, errinfo.ErrInvalidCellConfig
	}
	return c, nil
}

func (c *FileWriter) Run() {
	defer c.StopCell()
	if !c.StartCell() {
		return
	}

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
		var err error = nil
		switch data := unit.Data().(type) {
		case []byte:
			err = writeBytes(writer, data)
		case string:
			err = writeBytes(writer, []byte(data))
		default:
			fmt.Printf("Invalid input type %v for FileWriter", reflect.TypeOf(unit.Data()))
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
