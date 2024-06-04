package reader

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

const (
	FileReaderName string = "file_reader"

	Config_FileReader_Size = "size"

	CHUNK_SIZE int = 1 << 10
)

var (
	fileReaderInputFormats  []icell.Format = nil
	fileReaderOutputFormats []icell.Format = []icell.Format{icell.BYTE_SLICE}
)

type FileReader struct {
	icell.Cell

	filename string
	total    uint64
}

func FileReaderHelp() {
	FileReaderHelpShort()
	format := `	IO:
	  ->cell: %v
	  cell->: %v
	Properties:
	  %v: filename to read from
	  %v: optional, total bytes to read
`
	fmt.Printf(format,
		fileReaderInputFormats,
		fileReaderOutputFormats,
		icell.CONFIG_name,
		Config_FileReader_Size)
}

func FileReaderHelpShort() {
	fmt.Printf("%v : read content from file\n", FileReaderName)
}

func NewFileReader(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &FileReader{
		total: 0,
	}
	c.ICell = c
	c.Init(stopChan, config)

	if filename, ok := config[icell.CONFIG_name]; ok {
		c.filename = filename
	} else {
		fmt.Println("file name not provided for FileReader")
		FileReaderHelp()
		return nil, errinfo.ErrInvalidCellConfig
	}

	if tStr, ok := config[Config_FileReader_Size]; ok {
		t, err := strconv.ParseUint(tStr, 10, 64)
		if err != nil {
			fmt.Println("[file_reader] invalid size", tStr)
			return nil, errinfo.ErrInvalidCellConfig
		}
		c.total = t
	}

	return c, nil
}

func (c *FileReader) Run() {
	c.OnCellStart()
	defer c.OnCellFinished()

	file, err := os.Open(c.filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	readBytes := uint64(0)
	for c.Running() {
		buffer := make([]byte, CHUNK_SIZE)
		cnt, err := reader.Read(buffer)
		if err != nil {
			break
		}
		if c.total > 0 && readBytes+uint64(cnt) >= c.total {
			// reach maximum read size
			cnt = int(c.total - readBytes)
			c.PutOutput(icell.NewCellUnit(buffer[:cnt]))
			break
		}
		c.PutOutput(icell.NewCellUnit(buffer[:cnt]))
		readBytes += uint64(cnt)
	}
}
