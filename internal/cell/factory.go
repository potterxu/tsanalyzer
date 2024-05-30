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
	readers    = []string{reader.FileReaderName}
	converters = []string{}
	processors = []string{}
	writers    = []string{writer.FileWriterName}

	factories = map[string]*factory{
		reader.FileReaderName: {reader.NewFileReader, reader.FileReaderHelpShort, reader.FileReaderHelp},
		writer.FileWriterName: {writer.NewFileWriter, writer.FileWriterHelpShort, writer.FileWriterHelp},
	}
)

func NewCell(name string, stopChan chan bool, config map[string]string) (icell.ICell, error) {
	if factory, ok := factories[name]; ok {
		return factory.ctor(stopChan, config)
	}
	return nil, errinfo.ErrCellNotSupport
}

func printCategoryShort(category string, cells []string) {
	if len(cells) < 1 {
		return
	}
	fmt.Printf("--- %v ---\n", category)
	for _, name := range cells {
		factories[name].shortHelp()
	}
}
func PrintCells() {
	fmt.Println("=== Available Cells ===")
	printCategoryShort("readers", readers)
	printCategoryShort("converters", converters)
	printCategoryShort("processors", processors)
	printCategoryShort("writers", writers)
}

func printCategory(category string, cells []string) {
	if len(cells) < 1 {
		return
	}
	fmt.Printf("--- %v ---\n", category)
	for _, name := range cells {
		factories[name].help()
	}
}
func Help() {
	fmt.Println("===Cell Help===")
	printCategory("readers", readers)
	printCategory("converters", converters)
	printCategory("processors", processors)
	printCategory("writers", writers)
}

func CellHelper(name string) {
	if factory, ok := factories[name]; ok {
		fmt.Printf("===Help for cell [%v]===\n", name)
		factory.help()
	} else {
		fmt.Println("No valid cell", name)
	}
}
