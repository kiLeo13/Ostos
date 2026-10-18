package platform

// Unit is one of the quantity units supported by the initial product.
type Unit string

const (
	UnitEach       Unit = "each"
	UnitGram       Unit = "g"
	UnitKilogram   Unit = "kg"
	UnitMilliliter Unit = "ml"
	UnitLiter      Unit = "l"
)

func (unit Unit) Valid() bool {
	switch unit {
	case UnitEach, UnitGram, UnitKilogram, UnitMilliliter, UnitLiter:
		return true
	default:
		return false
	}
}
