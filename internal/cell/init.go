package cell

import (
	"github.com/potterxu/tsanalyzer/internal/cell/impl/converter"
	"github.com/potterxu/tsanalyzer/internal/cell/impl/processor"
	"github.com/potterxu/tsanalyzer/internal/cell/impl/reader"
	"github.com/potterxu/tsanalyzer/internal/cell/impl/writer"
)

// register the cell here
func init() {
	register(type_reader, reader.FileReaderName, reader.NewFileReader, reader.FileReaderHelpShort, reader.FileReaderHelp)
	register(type_converter, converter.BytesConverterName, converter.NewBytesConverter, converter.BytesConverterHelpShort, converter.BytesConverterHelp)
	register(type_writer, writer.FileWriterName, writer.NewFileWriter, writer.FileWriterHelpShort, writer.FileWriterHelp)
	register(type_processor, processor.VbvName, processor.NewVbv, processor.VbvHelpShort, processor.VbvHelp)
}

func register(t cellType, name string, ctor cell_ctor, short cell_short, help cell_help) {
	cells[t] = append(cells[t], name)
	factories[name] = &factory{ctor, short, help}
}
