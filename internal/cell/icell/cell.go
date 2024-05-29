package icell

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
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
	CONFIG_id          = "id"
	CONFIG_name        = "name"
	CONFIG_input_type  = "input_type"
	CONFIG_output_type = "output_type"
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
	c.StartCell()
	c.StopCell()
}
func (c *Cell) Stop() error {
	if !c.StopCell() {
		return errinfo.ErrCellAlreadyStop
	}
	return nil
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
	if v, ok := config[CONFIG_id]; ok {
		c.id = v
	} else {
		c.id = uuid.New().String()
	}
	c.running = false
}
func (c *Cell) StartCell() bool {
	if !c.running {
		c.running = true
		return true
	}
	return false
}
func (c *Cell) StopCell() bool {
	if c.running {
		c.running = false
		c.stopChan <- true
		if c.output != nil {
			c.output.Close()
		}
		return true
	}
	return false
}
func (c *Cell) Running() bool {
	return c.running
}
func (c *Cell) GetInput() (interface{}, bool) {
	if c.input != nil {
		v, ok := <-c.input.Channel()
		return v, ok
	}
	return nil, false
}
func (c *Cell) PutOutput(data interface{}) {
	if c.output != nil {
		c.output.Channel() <- data
	}
}
