package main

import (
	"fmt"
	"minishell/shell"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	// Обработка Ctrl+C
	// Мы перехватываем SIGINT, чтобы сам шелл не закрывался.
	// Запущенные дочерние процессы получат сигнал автоматически от ОС, 
	// так как находятся в той же группе процессов (если не отсоединены).
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	
	go func() {
		for range c {
			// При нажатии Ctrl+C просто переходим на новую строку,
			// если мы находимся в ожидании ввода.
			// Если работает дочерний процесс, он получит сигнал и умрет,
			// а shell.Run() продолжит цикл.
			fmt.Print("\n")
		}
	}()

	fmt.Println("Welcome to Minishell! (Type 'exit' or Ctrl+D to quit)")
	shell.Run()
}