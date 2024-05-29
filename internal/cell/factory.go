package cell

import (
	"fmt"

	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/cell/impl/reader"
	"github.com/potterxu/tsanalyzer/internal/cell/impl/writer"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

type factory struct {
	ctor      func(chan bool, icell.Config) (icell.ICell, error)
	shortHelp func()
	help      func()
}

var (
	factories = map[string]*factory{
		"filereader": {reader.NewFileReader, reader.FileReaderHelpShort, reader.FileReaderHelp},
		"filewriter": {writer.NewFileWriter, writer.FileWriterHelpShort, writer.FileWriterHelp},
	}
)

func NewCell(name string, stopChan chan bool, config map[string]string) (icell.ICell, error) {
	if factory, ok := factories[name]; ok {
		return factory.ctor(stopChan, config)
	}
	return nil, errinfo.ErrCellNotSupport
}

func PrintCells() {
	fmt.Println("===Available cells===")
	for _, factory := range factories {
		factory.shortHelp()
	}
}

func CellHelper(name string) {
	if factory, ok := factories[name]; ok {
		fmt.Printf("===Help for cell [%v]===\n", name)
		factory.help()
	} else {
		fmt.Println("No valid cell", name)
	}
}

func Help() {
	fmt.Println("===Help for all cells===")
	for _, factory := range factories {
		factory.help()
	}
}
