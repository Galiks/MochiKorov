package game

import (
	"fmt"
	"math/rand"
)

type DiceResult struct {
	Numbers []int `json:"numbers"`
	Sum     int   `json:"sum"`
}

func RollDice(count int) (DiceResult, error) {
	if count == 0 || count > 2 {
		return DiceResult{}, fmt.Errorf("invalid dice count: %d", count)
	}
	nums := make([]int, count)
	sum := 0
	for i := range count {
		nums[i] = rand.Intn(6) + 1
		sum += nums[i]
	}
	return DiceResult{Numbers: nums, Sum: sum}, nil
}
