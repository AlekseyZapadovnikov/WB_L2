package main

import (
  "fmt"
  "math/rand"
  "time"
)

func asChan(vs ...int) <-chan int {
  c := make(chan int)
  go func() {
    for _, v := range vs {
      c <- v
      time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
    }
  close(c)
}()
  return c
}

func merge(a, b <-chan int) <-chan int {
  c := make(chan int)
  go func() {
    for {
      select {
        case v, ok := <-a:
          if ok {
            c <- v
          } else {
            a = nil
          }
        case v, ok := <-b:
          if ok {
            c <- v
          } else {
            b = nil
          }
        }
        if a == nil && b == nil {
          close(c)
          return
        }
     }
   }()
  return c
}

  func main() {
    rand.Seed(time.Now().Unix())
    a := asChan(1, 3, 5, 7)
    b := asChan(2, 4, 6, 8)
    c := merge(a, b)
    for v := range c {
    fmt.Print(v)
  }
}

/*
asChan пишет данные в канал в отдельной горутине, после этого закрывает его и возвращает этоот канал
запускается 2 раза функция asChan, которая возвращает 2 канала с числами

функция merge принимает на вход 2 канала, и запускает горутину, которая читает из этих каналов в бесконечном цикле с помощью select
если из какого-то канала прочитано значение, оно отправляется в результирующий канал
если канал закрыт, то он присваивается nil, чтобы больше не читать из него

Присваивание a = nil — это ключевой трюк. Операции с nil-каналом (чтение или запись) в select блокируются навсегда.
Устанавливая закрытый канал в nil,
мы "выключаем" соответствующий case, и select больше никогда не будет его рассматривать.

тк все работает конкурентно/паралельно вывод программы предсказать невозможно
*/