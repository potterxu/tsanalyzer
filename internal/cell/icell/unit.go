package icell

type CellUnit interface {
	Data() interface{}
	Format() Format
}

type cellUnit struct {
	data   interface{}
	format Format
}

func NewCellUnit(data interface{}, format Format) *cellUnit {
	return &cellUnit{
		data:   data,
		format: format,
	}
}

func (u *cellUnit) Data() interface{} {
	return u.data
}

func (u *cellUnit) Format() Format {
	return u.format
}
