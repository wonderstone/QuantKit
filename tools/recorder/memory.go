package recorder

import "github.com/wonderstone/QuantKit/config"

type MemoryRecorder[Struct any] struct {
	sliceData []Struct
	channel   chan any
}

func (m *MemoryRecorder[Struct]) QueryRecord(query ...WithQuery) []any {
	config.ErrorF("MemoryRecorder目前不支持查询")
	return nil
}

func (m *MemoryRecorder[Struct]) Read(data any) error {
	data = m.sliceData
	return nil
}

func NewMemoryRecorder[Struct any](option ...WithOption) *MemoryRecorder[Struct] {
	return &MemoryRecorder[Struct]{
		sliceData: make([]Struct, 0),
		channel:   make(chan any),
	}
}

func (m *MemoryRecorder[Struct]) GetRecord() []Struct {
	return m.sliceData
}

func (m *MemoryRecorder[Struct]) RecordChan() error {
	for d := range m.channel {
		m.sliceData = append(m.sliceData, *d.(*Struct))
	}
	return nil
}

func (m *MemoryRecorder[Struct]) GetChannel() chan any {
	return m.channel
}

func (m *MemoryRecorder[Struct]) Release() {
	close(m.channel)
}
