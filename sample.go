package main

import "fmt"

func main() {
	sum := addNumbers(2, 6)
	fmt.Println(sum)
}

func addNumbers(a, b int) int {
	return a + b
}
