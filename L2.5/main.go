package main

type customError struct {
  msg string
}

func (e *customError) Error() string {
  return e.msg
}

func test() *customError {
  // ... do something
  return nil
}

func main() {
  var err error
  err = test()
  if err != nil {
    println("error")
    return
  }
  println("ok")
}

/*
вывод будет "error", потому что err это интерфейс с значением динамического типа *customError и значением nil => этот интерфейс не nil
подробно:
1) объявляется переменная err типа интерфейс error
2) Функция test() вызывается и возвращает nil типа *customError, ну то есть она возвращает *customError
3) err = test() - в этой строчке err получает значение типа *customError == nil
4) и теперь сам интерфейс не nil
*/