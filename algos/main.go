package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	in := bufio.NewReader(os.Stdin)

	var n int
	fmt.Fscan(in, &n)

	parent := make([]int, n)
	for i := 1; i < n; i++ {
		fmt.Fscan(in, &parent[i])
	}

	balance := make([]int, n)
	for i := 0; i < n; i++ {
		fmt.Fscan(in, &balance[i])
	}

	children := make([][]int, n)
	for i := 1; i < n; i++ {
		children[parent[i]] = append(children[parent[i]], i)
	}

	stack := []int{0}
	var order []int
	for len(stack) > 0 {
		u := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		order = append(order, u)
		for _, v := range children[u] {
			stack = append(stack, v)
		}
	}

	effect := make([]int, n)
	var ops int64 = 0

	for i := len(order) - 1; i >= 0; i-- {
		u := order[i]
		sumEffect := 0
		for _, v := range children[u] {
			sumEffect += effect[v]
		}
		need := -balance[u] - sumEffect
		ops += int64(abs(need))
		effect[u] = sumEffect + need
	}

	fmt.Println(ops)
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}