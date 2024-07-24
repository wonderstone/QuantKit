package indicator

// struct tmpRecordFunc
type tmpRecordFunc struct {
	Data    []string
	Header   map[string]int
}

func (t *tmpRecordFunc) Val(fieldName string, header ...map[string]int) string {
	return t.Data[t.Header[fieldName]]
}
func (t *tmpRecordFunc) Update(fieldName, value string, header ...map[string]int){
	t.Data[t.Header[fieldName]] = value
}