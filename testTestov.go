package l2

import(
	"fmt"
)


func UnnamedReturn() int {
  num := 100

  defer func(num int) {
    fmt.Printf("num value in the 1st defer `before`: %d\n", num)
    num = 777
    fmt.Printf("num value in the 1st after `after`: %d\n", num)
  }(num)

  num = 200

  defer func() {
    fmt.Printf("num value in the 2nd defer `before`: %d\n", num)
    num = 300
    fmt.Printf("num value in the 2nd defer `after`: %d\n", num)
  }()

  fmt.Printf("num value before function exiting: %d\n", num)
  return num
}

// num value before function exiting: 200
// num value in the 2nd defer `before`: 200
// num value in the 2nd defer `after`: 300
// num value in the 1st defer `before`: 100
// num value in the 1st after `after`: 777
// Func result: 200
