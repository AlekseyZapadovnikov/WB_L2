package main


func main() {
	n1 := []int{1,2,3,0,0,0}
	n2 := []int{2, 5, 6}
	merge(n1, 3, n2, 3)
}


func merge(nums1 []int, m int, nums2 []int, n int) {
    ans := make([]int, 0, n+m)
    p1, p2 := 0, 0

    for p1 < m && p2 < n {
        if nums1[p1] <= nums2[p2] {
            ans = append(ans, nums1[p1])
            p1++
        } else {
           ans = append(ans, nums2[p2])
           p2++
        }
    }

	// fmt.Println(ans[len(ans)-1], nums1[m-1], p1, m-1)
    if ans[len(ans)-1] == nums1[m-1] && p1 == m {
		// fmt.Println("here")
        for p2 < n {
            ans = append(ans, nums2[p2])
            p2++
        }
    } else {
        for p1 < n {
            ans = append(ans, nums1[p1])
            p1++
        }
    }

	// fmt.Println(ans)
}