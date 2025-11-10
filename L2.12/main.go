// main.go
// main пакет реализует утилиту фильтрации текстового потока, аналогичную команде grep.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

// Config хранит конфигурацию, полученную из флагов командной строки.
type Config struct {
	afterLines    int  // -A N: количество строк после совпадения
	beforeLines   int  // -B N: количество строк до совпадения
	contextLines  int  // -C N: количество строк контекста (до и после)
	count         bool // -c: выводить только количество совпадающих строк
	ignoreCase    bool // -i: игнорировать регистр
	invertMatch   bool // -v: инвертировать результат (выводить несовпадающие)
	fixedString   bool // -F: воспринимать шаблон как фиксированную строку
	lineNumbers   bool // -n: выводить номера строк
}

// GrepProcessor обрабатывает входной поток и применяет фильтрацию.
type GrepProcessor struct {
	config      Config
	pattern     string
	inputReader *bufio.Reader
	outputWriter io.Writer
	currentLine int // Текущий номер строки, обрабатываемой из входного потока
	
	// Для контекстного вывода:
	// Queue для хранения строк до совпадения (-B)
	linesBefore []string 
	// Queue для хранения строк после совпадения (-A)
	linesAfter []string

	// Флаг, указывающий, что следующая выведенная строка является частью контекста после совпадения
	printingAfterContext bool 
}

// NewGrepProcessor создает новый экземпляр GrepProcessor.
func NewGrepProcessor(cfg Config, pattern string, reader io.Reader, writer io.Writer) *GrepProcessor {
	gp := &GrepProcessor{
		config:      cfg,
		pattern:     pattern,
		inputReader: bufio.NewReader(reader),
		outputWriter: writer,
		currentLine: 0,
		linesBefore: make([]string, 0, cfg.beforeLines+cfg.contextLines), // Задаем начальную емкость
		linesAfter:  make([]string, 0, cfg.afterLines+cfg.contextLines),
	}
	// Обрабатываем флаг -C: если он установлен, используем его значение для -A и -B
	if cfg.contextLines > 0 {
		gp.config.beforeLines = cfg.contextLines
		gp.config.afterLines = cfg.contextLines
	}
	return gp
}

// readLine читает одну строку из входного потока.
func (gp *GrepProcessor) readLine() (string, error) {
	gp.currentLine++
	line, err := gp.inputReader.ReadString('\n')
	if err != nil {
		// Если это конец файла (EOF), но мы прочитали что-то, вернем прочитанное
		if err == io.EOF && len(line) > 0 {
			return line, nil
		}
		return "", err
	}
	// Удаляем символ новой строки, если он есть
	return strings.TrimSuffix(line, "\n"), nil
}

// match проверяет, совпадает ли строка с шаблоном, с учетом флагов.
func (gp *GrepProcessor) match(line string) (bool, error) {
	var matched bool
	var err error

	if gp.config.fixedString {
		if gp.config.ignoreCase {
			matched = strings.Contains(strings.ToLower(line), strings.ToLower(gp.pattern))
		} else {
			matched = strings.Contains(line, gp.pattern)
		}
	} else {
		// Компилируем регулярное выражение, учитывая флаг -i
		regexPattern := gp.pattern
		if gp.config.ignoreCase {
			regexPattern = "(?i)" + regexPattern // Флаг (?i) для игнорирования регистра
		}
		re, err := regexp.Compile(regexPattern)
		if err != nil {
			return false, fmt.Errorf("invalid regular expression %q: %w", gp.pattern, err)
		}
		matched = re.MatchString(line)
	}

	// Инвертируем результат, если установлен флаг -v
	if gp.config.invertMatch {
		matched = !matched
	}
	return matched, err
}

// printLine выводит строку с номером, если это необходимо.
func (gp *GrepProcessor) printLine(line string, lineNumber int) error {
	if gp.config.count {
		return nil // Если считаем, ничего не печатаем на этом этапе
	}
	
	if gp.config.lineNumbers {
		fmt.Fprintf(gp.outputWriter, "%d:%s\n", lineNumber, line)
	} else {
		fmt.Fprintln(gp.outputWriter, line)
	}
	return nil
}

// Process запускает процесс фильтрации.
func (gp *GrepProcessor) Process() error {
	var matchCount int // Счетчик совпадений для флага -c

	// Основной цикл чтения и обработки строк
	for {
		line, err := gp.readLine()
		if err != nil {
			if err == io.EOF {
				break // Достигнут конец файла, выходим из цикла
			}
			return fmt.Errorf("error reading input: %w", err)
		}

		matched, matchErr := gp.match(line)
		if matchErr != nil {
			return matchErr // Возвращаем ошибку компиляции регулярного выражения
		}

		// --- Обработка контекста (до) ---
		// Если мы находимся в режиме печати контекста после совпадения,
		// и текущая строка не совпала, нам нужно печатать её как строку до.
		if gp.printingAfterContext && !matched {
			gp.printingAfterContext = false // Вышли из контекста после совпадения
		}

		// Если мы печатаем строки до совпадения:
		// Отбрасываем старые строки, если очередь переполнена.
		if len(gp.linesBefore) > gp.config.beforeLines {
			gp.linesBefore = gp.linesBefore[1:]
		}
		// Добавляем текущую строку в очередь строк до.
		gp.linesBefore = append(gp.linesBefore, line)

		// --- Обработка совпадения ---
		if matched {
			matchCount++

			// Если мы печатали контекст после предыдущего совпадения,
			// но нашли новое совпадение, печатаем оставшийся контекст.
			if gp.printingAfterContext {
				// Печатаем строки из linesAfter (которые теперь становятся linesBefore для нового контекста)
				for i, l := range gp.linesAfter {
					if err := gp.printLine(l, gp.currentLine-len(gp.linesAfter)+i+1); err != nil {
						return err
					}
				}
				gp.linesAfter = gp.linesAfter[:0] // Очищаем очередь после печати
			}

			// Печатаем строки, которые были до этого совпадения.
			// Учитываем номера строк для этих строк.
			for i, l := range gp.linesBefore {
				if err := gp.printLine(l, gp.currentLine-len(gp.linesBefore)+i+1); err != nil {
					return err
				}
			}
			gp.linesBefore = gp.linesBefore[:0] // Очищаем очередь после печати

			// Печатаем саму найденную строку.
			if err := gp.printLine(line, gp.currentLine); err != nil {
				return err
			}

			// Устанавливаем флаг, что следующая строка будет частью контекста после.
			gp.printingAfterContext = true
			// Очищаем очередь linesAfter, так как мы начинаем новый контекст.
			gp.linesAfter = gp.linesAfter[:0]

		} else {
			// Если строка не совпала, но мы в режиме печати контекста после предыдущего совпадения:
			if gp.printingAfterContext {
				// Добавляем строку в очередь строк после.
				gp.linesAfter = append(gp.linesAfter, line)
				// Если очередь строк после стала слишком длинной, удаляем старейшую.
				if len(gp.linesAfter) > gp.config.afterLines {
					gp.linesAfter = gp.linesAfter[1:]
				}
			}
		}
	}

	// --- Обработка оставшегося контекста в конце файла ---
	// Если мы были в режиме печати контекста после последнего совпадения,
	// и достигли конца файла, печатаем оставшийся контекст.
	if gp.printingAfterContext {
		for i, l := range gp.linesAfter {
			if err := gp.printLine(l, gp.currentLine-len(gp.linesAfter)+i+1); err != nil {
				return err
			}
		}
	}

	// Если установлен флаг -c, выводим итоговое количество совпадений.
	if gp.config.count {
		fmt.Fprintln(gp.outputWriter, matchCount)
	}

	return nil
}

func main() {
	// Определение и парсинг флагов
	// Используем более человекочитаемые имена флагов, но сохраняем короткие для совместимости.
	var (
		afterLines    = flag.Int("A", 0, "print N lines of trailing context after matching lines")
		beforeLines   = flag.Int("B", 0, "print N lines of leading context before matching lines")
		contextLines  = flag.Int("C", 0, "print N lines of context around matching lines")
		count         = flag.Bool("c", false, "suppress normal output; instead print a count of matching lines")
		ignoreCase    = flag.Bool("i", false, "ignore case distinctions in patterns and input data")
		invertMatch   = flag.Bool("v", false, "select non-matching lines")
		fixedString   = flag.Bool("F", false, "interpret PATTERN as a list of fixed strings, separated by newlines, any of which is to be matched")
		lineNumbers   = flag.Bool("n", false, "prefix each line of output with the 1-based line number within its input file")
	)
	flag.Parse()

	// Проверка на конфликтующие флаги
	if *afterLines > 0 && *contextLines > 0 {
		fmt.Fprintf(os.Stderr, "grep: error: cannot specify both -A and -C\n")
		os.Exit(1)
	}
	if *beforeLines > 0 && *contextLines > 0 {
		fmt.Fprintf(os.Stderr, "grep: error: cannot specify both -B and -C\n")
		os.Exit(1)
	}
	if *count && (*afterLines > 0 || *beforeLines > 0 || *contextLines > 0) {
		// GNU grep игнорирует контекст при count. Мы тоже будем следовать этому.
		// Можно было бы сделать ошибку, но это менее гибко.
		// fmt.Fprintf(os.Stderr, "grep: warning: context flags (-A, -B, -C) are ignored when -c is used\n")
	}

	// Формирование конфигурации
	config := Config{
		afterLines:    *afterLines,
		beforeLines:   *beforeLines,
		contextLines:  *contextLines,
		count:         *count,
		ignoreCase:    *ignoreCase,
		invertMatch:   *invertMatch,
		fixedString:   *fixedString,
		lineNumbers:   *lineNumbers,
	}

	// Определение источника ввода и паттерна
	var reader io.Reader
	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "grep: pattern missing\n")
		flag.Usage()
		os.Exit(1)
	}
	pattern := args[0]
	inputFileName := "stdin" // Для сообщений об ошибках

	if len(args) > 1 { // Если указан файл
		inputFileName = args[1]
		file, err := os.Open(inputFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "grep: %s: %v\n", inputFileName, err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	} else { // Читаем из STDIN
		reader = os.Stdin
	}

	// Создание и запуск обработчика
	processor := NewGrepProcessor(config, pattern, reader, os.Stdout)
	err := processor.Process()
	if err != nil {
		fmt.Fprintf(os.Stderr, "grep: %v\n", err)
		os.Exit(1)
	}
}