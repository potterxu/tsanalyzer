package cell

import (
	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/cell/impl/reader"
	"github.com/potterxu/tsanalyzer/internal/cell/impl/writer"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

var (
	factory = map[string]func(chan bool, icell.Config) (icell.ICell, error){
		"filereader": reader.NewFileReader,
		"filewriter": writer.NewFileWriter,
	}
)

func NewCell(name string, stopChan chan bool, config map[string]string) (icell.ICell, error) {
	if ctor, ok := factory[name]; ok {
		return ctor(stopChan, config)
	}
	return nil, errinfo.ErrCellNotSupport
}
