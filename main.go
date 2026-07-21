package main

import (
	"fmt"
	"mochi_korov/models"
)

func main() {
	user := models.NewPlayer(1, "TEST")
	res, err := user.RollCubes(3)
	if err != nil {
		panic(err)
	}
	println(res)
	fmt.Printf("user.Turn: %v\n", user.Turn)
	res, err = user.RerollCubes([]uint8{2})
	if err != nil {
		panic(err)
	}
	println(res)
	fmt.Printf("user.Turn: %v\n", user.Turn)
	res, err = user.AddTwoToCubeResult()
	if err != nil {
		panic(err)
	}
	fmt.Printf("user.Turn: %v\n", user.Turn)
}
