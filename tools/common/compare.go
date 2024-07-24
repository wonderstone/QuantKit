package common

func Max[Integer int | int64](a, b Integer) Integer {
	if a > b {
		return a
	}

	return b
}

func Min[Integer int | int64](a, b Integer) Integer {
	if a < b {
		return a
	}

	return b
}
