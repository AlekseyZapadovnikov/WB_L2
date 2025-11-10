package main

import "sync"

func main() {

}

func or(channels ...<-chan interface{}) <-chan interface{} {
    out := make(chan interface{})
    if len(channels) == 0 {
        close(out)
        return out
    }

    var once sync.Once
    for _, ch := range channels {
        go func(c <-chan interface{}) {
            <-c // блокирующая операция
            once.Do(func() { close(out) })
        }(ch)
    }

    return out
}
// все горутины будут висеть и ждать пока придёт сообщение, после конал out закроется