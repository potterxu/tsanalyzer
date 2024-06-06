package icell

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

// Every customized Cell should composite Cell struct
// and replace any required methods
type Cell struct {
	ICell

	id       string
	running  bool
	stopChan chan bool

	// pipeline usage
	input  *Edge
	output *Edge
}

type Config map[string]string

const (
	config_cel_id = "id"
)

// Default interface method
func (c *Cell) Id() string {
	return c.id
}
func (c *Cell) Connect(next ICell) error {
	e, err := newEdge(c.ICell, next)
	if err != nil {
		return err
	}
	c.SetOutput(e)
	next.SetInput(e)
	return nil
}
func (c *Cell) Run() {
	fmt.Println("Please implement Run() method for cell", reflect.TypeOf(c.ICell).Elem().Name())
	c.OnCellStart()
	defer c.OnCellFinished()
}
func (c *Cell) Stop() {
	c.running = false
}
func (c *Cell) SetInput(e *Edge) {
	c.input = e
}
func (c *Cell) SetOutput(e *Edge) {
	c.output = e
}

// General method for custom cells to control the flow
func (c *Cell) Init(stopChan chan bool, config Config) {
	c.stopChan = stopChan
	if v, ok := config[config_cel_id]; ok {
		c.id = v
	} else {
		c.id = uuid.New().String()
	}
	c.running = false
}
func (c *Cell) OnCellStart() {
	c.running = true
}
func (c *Cell) OnCellFinished() {
	c.stopChan <- true
	if c.output != nil {
		c.output.Close()
	}
}
func (c *Cell) Running() bool {
	return c.running
}
func (c *Cell) GetInput() (CellUnit, bool) {
	if c.input != nil {
		v, ok := <-c.input.Channel()
		return v, ok
	}
	return nil, false
}
func (c *Cell) PutOutput(unit CellUnit) {
	if c.output != nil {
		c.output.Channel() <- unit
	}
}
