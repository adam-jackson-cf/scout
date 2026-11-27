package sum

func Total(values []int) int {
	t := 0
	for _, v := range values {
		t += v
	}
	return t
}
