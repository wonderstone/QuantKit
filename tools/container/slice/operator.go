package slice

// isContains 判断两个slice是否包含
// if element in wanted is not in sets, 
// which means wanted is not in sets
// return false
func Contains[T comparable](wanted []T, sets []T) bool {
	m := make(map[T]bool)
	for _, v := range sets {
		m[v] = true
	}

	for _, vv := range wanted {
		if !m[vv] {
			return false
		}
	}

	return true
}

func ContainsFunc[T any, S any, Key comparable](
	wanted []T, sets []S, tFunc func(t T) Key, sFunc func(s S) Key,
) (*Key, bool) {
	m := make(map[Key]bool)
	for _, v := range sets {
		m[sFunc(v)] = true
	}

	for _, vv := range wanted {
		k := tFunc(vv)
		if !m[k] {
			return &k, false
		}
	}

	return nil, true
}
