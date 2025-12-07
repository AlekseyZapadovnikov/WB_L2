package shell

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Run запускает основной цикл REPL
func Run() {
	reader := bufio.NewReader(os.Stdin)

	for {
		// Вывод приглашения с текущей директорией
		cwd, _ := os.Getwd()
		fmt.Printf("\033[32m%s\033[0m > ", cwd)

		input, err := reader.ReadString('\n')
		if err != nil {
			// Обработка Ctrl+D (EOF)
			fmt.Println("exit")
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "exit" {
			break
		}

		// Обработка логических операторов && и ||
		executeLogic(input)
	}
}

// executeLogic разбирает строку на команды, разделенные && или ||
func executeLogic(input string) {
	// Это упрощенный парсер. Он не поддерживает сложные вложенности и скобки.
	// Мы разбиваем по операторам и выполняем последовательно.
	
	// Токенизация по операторам — сложная задача для strings.Split.
	// Пройдемся вручную или упростим: допустим, разделитель всегда окружен пробелами.
	// Для мини-проекта сделаем рекурсивную обработку первого найденного оператора.

	orIdx := strings.Index(input, " || ")
	andIdx := strings.Index(input, " && ")

	if orIdx != -1 && (andIdx == -1 || orIdx < andIdx) {
		// Нашли || первым
		lhs := input[:orIdx]
		rhs := input[orIdx+4:]
		
		err := ProcessPipeline(lhs)
		if err != nil {
			// Если левая часть упала, выполняем правую
			executeLogic(rhs)
		}
		return
	}

	if andIdx != -1 {
		// Нашли && первым
		lhs := input[:andIdx]
		rhs := input[andIdx+4:]

		err := ProcessPipeline(lhs)
		if err == nil {
			// Если левая часть успешна, выполняем правую
			executeLogic(rhs)
		}
		return
	}

	// Если операторов нет, просто запускаем пайплайн
	if err := ProcessPipeline(input); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}