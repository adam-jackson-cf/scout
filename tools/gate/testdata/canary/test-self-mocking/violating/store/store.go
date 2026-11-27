package store

type Store interface {
	Get(key string) (string, bool)
}

type Memory struct{ values map[string]string }

func NewMemory(values map[string]string) *Memory { return &Memory{values: values} }

func (m *Memory) Get(key string) (string, bool) {
	v, ok := m.values[key]
	return v, ok
}

func Lookup(s Store, key string) string {
	if v, ok := s.Get(key); ok {
		return v
	}
	return "missing"
}
