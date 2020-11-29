package main

func (j *Juice) Juice(key string, values []string) {
	count0 := 0
	count1 := 0

	for _, v := range values {
		if v == "0" {
			count0++
		} else {
			count1++
		}
	}

	if count0 >= count1 {
		j.Emit(key)
	}
}
