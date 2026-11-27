package sum

func Total(values []int) int { // coverage-ignore
	t := 0
	for _, v := range values {
		t += v
	}
	return t
}
