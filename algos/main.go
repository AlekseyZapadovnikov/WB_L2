package main

import (
	"container/heap"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type MinHeap []int

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(int))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

func parseTime(s string) int {
	parts := strings.Split(s, ":")
	hour, _ := strconv.Atoi(parts[0])
	minute, _ := strconv.Atoi(parts[1])
	return hour*60 + minute
}

type Trip struct {
	start      int
	end        int
	fromOffice int 
}

func main() {
	var n int
	fmt.Scan(&n)

	allTrips := make([]Trip, 0, n+100000)

	for i := 0; i < n; i++ {
		var s string
		fmt.Scan(&s)
		parts := strings.Split(s, "-")
		start := parseTime(parts[0])
		end := parseTime(parts[1])
		allTrips = append(allTrips, Trip{start, end, 1})
	}

	var m int
	fmt.Scan(&m)

	for i := 0; i < m; i++ {
		var s string
		fmt.Scan(&s)
		parts := strings.Split(s, "-")
		start := parseTime(parts[0])
		end := parseTime(parts[1])
		allTrips = append(allTrips, Trip{start, end, 2})
	}

	sort.Slice(allTrips, func(i, j int) bool {
		return allTrips[i].start < allTrips[j].start
	})

	var freeAt1 MinHeap
	var freeAt2 MinHeap 
	heap.Init(&freeAt1)
	heap.Init(&freeAt2)

	busCount := 0

	for _, trip := range allTrips {
		if trip.fromOffice == 1 {
			if freeAt1.Len() > 0 && freeAt1[0] <= trip.start {
				heap.Pop(&freeAt1)
			} else {
				busCount++
			}
			heap.Push(&freeAt2, trip.end)
		} else {
			if freeAt2.Len() > 0 && freeAt2[0] <= trip.start {
				heap.Pop(&freeAt2)
			} else {
				busCount++
			}
			heap.Push(&freeAt1, trip.end)
		}
	}

	fmt.Println(busCount)
}