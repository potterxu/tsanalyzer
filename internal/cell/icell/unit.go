package icell

type CellUnit interface {
	Data() interface{}
}

type cellUnit struct {
	data interface{}
}

func NewCellUnit(data interface{}) *cellUnit {
	return &cellUnit{
		data: data,
	}
}

func (u *cellUnit) Data() interface{} {
	return u.data
}
