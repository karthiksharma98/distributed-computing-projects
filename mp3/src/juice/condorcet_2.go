package main

import (
	"fmt"
	"strconv"
	"strings"
)

func (j *Juice) Juice(key string, values []string) {

	numCandidates := 10

	votes := make([]int, numCandidates)
	for _, v := range values {
		pairWinner, _ := strconv.Atoi(string(strings.TrimSpace(v)[1]))
		votes[pairWinner]++
	}

	for i, v := range votes {
		if v == numCandidates-1 {
			j.Emit(fmt.Sprint(i), " is the condorcet winner!")
			return
		}
	}

	// if reached here then no condorcet winner
	maxCount := 0
	winnerSet := ""
	for i, v := range votes {
		if v > maxCount {
			maxCount = v
		}
	}

	for i, v := range votes {
		if v == maxCount {
			winnerSet = winnerSet + fmt.Sprint(i) + ","
		}
	}

	j.Emit(winnerSet, " have the highest condorcet counts, no winner.")

}