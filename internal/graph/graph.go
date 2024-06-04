package graph

import (
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"

	"github.com/potterxu/tsanalyzer/internal/cell"
	"github.com/potterxu/tsanalyzer/internal/cell/icell"
	"github.com/potterxu/tsanalyzer/internal/errinfo"
)

type Graph struct {
	pipeline  []icell.ICell
	stopChans []chan bool

	wg sync.WaitGroup
}

func getCellInfos(gDesc string) ([]*cellInfo, bool) {
	cellDescList := strings.Split(strings.TrimSpace(gDesc), "!")
	if len(cellDescList) < 1 {
		fmt.Println("No cell description found")
		return nil, false
	}
	cellInfos := make([]*cellInfo, len(cellDescList))
	for i, cDesc := range cellDescList {
		if len(cDesc) < 1 {
			fmt.Println("No cell description before !")
			return nil, false
		}
		cArgs := strings.Split(strings.TrimSpace(cDesc), " ")
		if len(cArgs) < 1 {
			fmt.Println("No cell description before !")
			return nil, false
		}
		cellName := cArgs[0]
		cellInfo := newCellInfo(cellName)

		if len(cArgs) > 1 {
			for _, property := range cArgs[1:] {
				pArgs := strings.Split(strings.TrimSpace(property), "=")
				if len(pArgs) != 2 {
					fmt.Printf("Invalid property for cell %v: %v\n", cellName, property)
					cell.CellHelper(cellName)
					return nil, false
				}
				cellInfo.addProperty(pArgs[0], pArgs[1])
			}
		}
		cellInfos[i] = cellInfo
	}
	fmt.Printf("Create pipeline: ")
	for i, info := range cellInfos {
		if i == 0 {
			fmt.Printf("%v ", info.name)
		} else {
			fmt.Printf("-> %v ", info.name)
		}
	}
	fmt.Println()
	return cellInfos, true
}

func connectPipeline(pipeline []icell.ICell) error {
	for i := range pipeline {
		if i >= len(pipeline)-1 {
			break
		}
		if err := pipeline[i].Connect(pipeline[i+1]); err != nil {
			return err
		}
	}
	return nil
}

func NewGraph(gDesc string) (*Graph, error) {
	cellInfos, ok := getCellInfos(gDesc)
	if !ok {
		return nil, errinfo.ErrFailedToBuildGraph
	}
	graph := &Graph{
		pipeline:  make([]icell.ICell, len(cellInfos)),
		stopChans: make([]chan bool, len(cellInfos)),
	}

	for i, info := range cellInfos {
		graph.stopChans[i] = make(chan bool, 1)
		var err error
		fmt.Printf("Create cell %v: %v\n", info.name, info.config)
		graph.pipeline[i], err = cell.NewCell(info.name, graph.stopChans[i], info.config)
		if err != nil {
			return nil, err
		}
	}

	if err := connectPipeline(graph.pipeline); err != nil {
		return nil, err
	}

	return graph, nil
}

func (g *Graph) Run() {
	// Start running cell from the end of pipeline
	for i := len(g.pipeline) - 1; i >= 0; i-- {
		g.wg.Add(1)
		go g.pipeline[i].Run()
		go func(index int) {
			<-g.stopChans[index]
			fmt.Printf("Cell %v finished\n", reflect.TypeOf(g.pipeline[index]).Elem().Name())
			g.wg.Done()
		}(i)
	}
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		// stop the first cell of the pipeline to stop the whole graph
		g.pipeline[0].Stop()
	}()
	g.wg.Wait()
	fmt.Println("Graph finished")
}
