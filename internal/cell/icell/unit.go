package icell

type CellUnit struct {
	data interface{}
}

func NewCellUnit(data interface{}) *CellUnit {
	return &CellUnit{
		data: data,
	}
}

func (u *CellUnit) Data() interface{} {
	return u.data
}
