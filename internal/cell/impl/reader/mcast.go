package reader

import (
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/potterxu/tsanalyzer/internal/cell/icell"
)

const (
	McastReaderName string = "mcast_reader"

	config_mcastreader_interface string = "intf"
	config_mcastreader_address   string = "addr"
)

var (
	mcastReaderInputFormats  []icell.Format = nil
	mcastReaderOutputFormats []icell.Format = []icell.Format{icell.BYTE_SLICE}
)

func NewMcastReader(stopChan chan bool, config icell.Config) (icell.ICell, error) {
	c := &mcastReader{}
	c.ICell = c
	c.Init(stopChan, config)

	if intf, ok := config[config_mcastreader_interface]; ok {
		c.intfName = intf
	} else {
		return nil, errors.New("interface name not found")
	}

	if addr, ok := config[config_mcastreader_address]; ok {
		c.address = addr
	} else {
		return nil, errors.New("multicast address not found")
	}

	return c, nil
}

func McastReaderHelp() {
	McastReaderHelpShort()
	format := `	IO:
	  ->cell: %v
	  cell->: %v
	Properties:
	  %v: interface name
	  %v: multicast address, e.g "239.1.1.1:1000"
`
	fmt.Printf(format,
		mcastReaderInputFormats,
		mcastReaderOutputFormats,
		config_mcastreader_interface,
		config_mcastreader_address,
	)
}

func McastReaderHelpShort() {
	fmt.Printf("%v : write content to file\n", McastReaderName)
}

type mcastReader struct {
	icell.Cell

	intfName string
	address  string
}

func (c *mcastReader) Run() {
	c.OnCellStart()
	defer c.OnCellFinished()

	intf, err := net.InterfaceByName(c.intfName)
	if err != nil {
		fmt.Println("[mcast_reader]", err)
		return
	}

	addr, err := net.ResolveUDPAddr("udp", c.address)
	if err != nil {
		fmt.Println("[mcast_reader]", err)
		return
	}

	conn, err := net.ListenMulticastUDP("udp", intf, addr)

	if err != nil {
		fmt.Println("[mcast_reader] Error connecting:", err)
		return
	}

	defer conn.Close()

	for c.Running() {
		// read network stream
		buffer := make([]byte, 1316)
		err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		if err != nil {
			fmt.Println("[mcast_reader] Error setting read deadline:", err)
			return
		}
		n, _, err := conn.ReadFrom(buffer)
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				continue
			}
			fmt.Println("[mcast_reader] Error reading:", err)
			return
		}
		data := buffer[:n]
		c.PutOutput(icell.NewCellUnit(data, icell.BYTE_SLICE))
	}
}
