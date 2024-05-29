package icell

// This is the interface for accessing Cells in pipeline
type ICell interface {
	// pipeline base methods
	Id() string
	Connect(ICell) error
	SetInput(e *Edge)
	SetOutput(e *Edge)

	// non go routine methods
	Stop() error // force stop the cell

	// go routine methods
	Run() // go Run() to start the cell processing, should terminate automatically
}
