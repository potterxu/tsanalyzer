package cell

import (
	"fmt"

	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

type cellType int

const (
	type_reader = iota
	type_converter
	type_processor
	type_writer
)

type cell_ctor func(chan bool, icell.Config) (icell.ICell, error)
type cell_short func()
type cell_help func()

type factory struct {
	ctor      cell_ctor
	shortHelp cell_short
	help      cell_help
}

var (
	cells     = map[cellType][]string{}
	factories = map[string]*factory{}
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
	fmt.Println()
}
func PrintCells() {
	printSyntax()
	fmt.Println("=== Available Cells ===")
	printCategoryShort("readers", cells[type_reader])
	printCategoryShort("converters", cells[type_converter])
	printCategoryShort("processors", cells[type_processor])
	printCategoryShort("writers", cells[type_writer])
}

func printCategory(category string, cells []string) {
	if len(cells) < 1 {
		return
	}
	fmt.Printf("--- %v ---\n", category)
	for _, name := range cells {
		factories[name].help()
	}
	fmt.Println()
}
func Help() {
	printSyntax()
	fmt.Println("===Cell Help===")
	printCategory("readers", cells[type_reader])
	printCategory("converters", cells[type_converter])
	printCategory("processors", cells[type_processor])
	printCategory("writers", cells[type_writer])
}

func CellHelper(name string) {
	if factory, ok := factories[name]; ok {
		fmt.Printf("===Help for cell [%v]===\n", name)
		factory.help()
	} else {
		fmt.Println("No valid cell", name)
	}
}

func printSyntax() {
	fmt.Println()
	fmt.Println("Syntax: pipe cell prop1=val1 prop2=val2 ! cell2 prop1=val1 prop2=val2 ! ...")
	fmt.Println()
}
