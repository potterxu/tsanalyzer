package reader

import (
	"bufio"
	"fmt"
	"os"

	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

const (
	FileReaderName string = "file_reader"
	CHUNK_SIZE     int    = 188
)

type FileReader struct {
	icell.Cell

	filename string
}

func FileReaderHelp() {
	FileReaderHelpShort()
	format := `  Properties:
    %v: filename to read from

`
	fmt.Printf(format, icell.CONFIG_name)
}

func FileReaderHelpShort() {
	fmt.Printf("%v : read content from file\n", FileReaderName)
	fmt.Println("  ->cell: N/A")
	fmt.Println("  cell->: []byte")
	fmt.Println("")
}

func NewFileReader(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &FileReader{}
	c.ICell = c
	c.Init(stopChan, config)

	if filename, ok := config[icell.CONFIG_name]; ok {
		c.filename = filename
	} else {
		fmt.Println("file name not provided for FileReader")
		FileReaderHelp()
		return nil, errinfo.ErrInvalidCellConfig
	}
	return c, nil
}

func (c *FileReader) Run() {
	defer c.StopCell()
	if !c.StartCell() {
		return
	}

	file, err := os.Open(c.filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	buffer := make([]byte, CHUNK_SIZE)
	for {
		cnt, err := reader.Read(buffer)
		if err != nil {
			break
		}
		c.PutOutput(icell.NewCellUnit(buffer[:cnt]))
	}
}
