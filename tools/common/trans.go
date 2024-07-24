package common

// GetFloat64 change the interface type to float64
func GetFloat64(target interface{}) float64 {
	switch target.(type) {
	case float64:
		return target.(float64)
	case int:
		return float64(target.(int))
	case int64:
		return float64(target.(int64))
	case float32:
		return float64(target.(float32))
	default:
		panic("unknown type")
	}
}
