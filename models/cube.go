package models

import (
	"fmt"
	"math/rand"
	"sync"
)

type CubeResult struct {
	Numbers []uint8
}

func GetCubeNumber(cubCount uint8) (CubeResult, error) {
	if cubCount == 0 || cubCount > 3 {
		return CubeResult{}, fmt.Errorf("invalid cube count: %d", cubCount)
	}
	result := CubeResult{Numbers: make([]uint8, cubCount)}
	wg := sync.WaitGroup{}
	for i := range cubCount {
		wg.Go(func() {
			// Генерация случайного числа от 1 до 6
			result.Numbers[i] = uint8(rand.Intn(6) + 1)
		})
	}
	wg.Wait()
	return result, nil
}
