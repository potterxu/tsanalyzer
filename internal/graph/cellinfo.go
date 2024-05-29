package graph

import "github.com/potterxu/tsanalyzer/internal/cell/icell"

type cellInfo struct {
	name   string
	config icell.Config
}

func newCellInfo(name string) *cellInfo {
	return &cellInfo{
		name:   name,
		config: make(icell.Config),
	}
}

func (ci *cellInfo) addProperty(key, value string) {
	ci.config[key] = value
}
