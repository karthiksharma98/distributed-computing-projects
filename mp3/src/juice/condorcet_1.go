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

	if count1 >= count0 {
		j.Emit("", key)
	} else {
		j.Emit("", string(key[0]+key[3]+key[2]+key[2]+key[4]))
	}
}
