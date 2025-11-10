// main.go
// Пакет main реализует упрощенный аналог UNIX-утилиты sort.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Config хранит конфигурацию, полученную из флагов командной строки.
type Config struct {
	column      int  // -k: номер колонки для сортировки (1-based)
	numeric     bool // -n: сортировать по числовому значению
	reverse     bool // -r: сортировать в обратном порядке
	unique      bool // -u: выводить только уникальные строки
	monthSort   bool // -M: сортировать по названию месяца
	ignoreBlanks bool // -b: игнорировать хвостовые пробелы
	checkSorted bool // -c: проверить, отсортированы ли данные
	humanNumeric bool // -h: сортировать по числовому значению с суффиксами (K, M, G...)
}

// Sorter инкапсулирует логику сортировки.
// Он реализует интерфейс sort.Interface.
type Sorter struct {
	lines  []string
	config Config
}

// NewSorter создает новый экземпляр Sorter.
func NewSorter(lines []string, config Config) *Sorter {
	return &Sorter{
		lines:  lines,
		config: config,
	}
}

// Len возвращает количество строк. Часть sort.Interface.
func (s *Sorter) Len() int {
	return len(s.lines)
}

// Swap меняет местами две строки. Часть sort.Interface.
func (s *Sorter) Swap(i, j int) {
	s.lines[i], s.lines[j] = s.lines[j], s.lines[i]
}

// Less сравнивает две строки в соответствии с заданной конфигурацией. Часть sort.Interface.
func (s *Sorter) Less(i, j int) bool {
	lineA := s.lines[i]
	lineB := s.lines[j]

	valA := s.getComparableValue(lineA)
	valB := s.getComparableValue(lineB)

	var less bool

	switch {
	case s.config.numeric:
		numA, errA := strconv.ParseFloat(valA, 64)
		numB, errB := strconv.ParseFloat(valB, 64)
		// Если одна из строк не число, она считается "меньше"
		if errA != nil {
			numA = 0
		}
		if errB != nil {
			numB = 0
		}
		less = numA < numB
	case s.config.humanNumeric:
		numA := parseHumanNumeric(valA)
		numB := parseHumanNumeric(valB)
		less = numA < numB
	case s.config.monthSort:
		monthA := parseMonth(valA)
		monthB := parseMonth(valB)
		less = monthA < monthB
	default:
		less = valA < valB
	}

	if s.config.reverse {
		return !less
	}
	return less
}

// getComparableValue извлекает и подготавливает значение для сравнения из строки.
// Учитывает флаги -k (колонка) и -b (игнорирование пробелов).
func (s *Sorter) getComparableValue(line string) string {
	if s.config.ignoreBlanks {
		line = strings.TrimRight(line, " \t")
	}

	if s.config.column <= 1 {
		return line
	}
	
	// Используем 0-based индекс для доступа к срезу
	colIndex := s.config.column - 1
	parts := strings.Split(line, "\t")

	if colIndex < len(parts) {
		return parts[colIndex]
	}
	// Если в строке нет нужной колонки, возвращаем пустую строку.
	return ""
}

// Sort запускает сортировку.
func (s *Sorter) Sort() {
	sort.Sort(s)
}

// CheckSorted проверяет, отсортирован ли файл.
// Возвращает true, если отсортирован. В противном случае - false и номер строки с нарушением.
func (s *Sorter) CheckSorted() (bool, int) {
	for i := 1; i < len(s.lines); i++ {
		// Если s.Less(i, i-1) истинно, значит элемент i должен стоять раньше i-1,
		// что является нарушением порядка.
		if s.Less(i, i-1) {
			return false, i + 1 // +1 для 1-based номера строки
		}
	}
	return true, 0
}

// parseHumanNumeric преобразует строку с суффиксами (K, M, G) в число.
func parseHumanNumeric(s string) int64 {
	s = strings.TrimSpace(strings.ToUpper(s))
	multiplier := int64(1)
	suffix := ""

	if len(s) > 1 {
		lastChar := s[len(s)-1]
		switch lastChar {
		case 'K':
			multiplier = 1024
			suffix = "K"
		case 'M':
			multiplier = 1024 * 1024
			suffix = "M"
		case 'G':
			multiplier = 1024 * 1024 * 1024
			suffix = "G"
		case 'T':
			multiplier = 1024 * 1024 * 1024 * 1024
			suffix = "T"
		}
	}
	
	if suffix != "" {
		s = s[:len(s)-1]
	}

	num, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}

	return num * multiplier
}

// monthMap сопоставляет трехбуквенные сокращения месяцев с их порядковым номером.
var monthMap = map[string]time.Month{
	"JAN": time.January, "FEB": time.February, "MAR": time.March,
	"APR": time.April, "MAY": time.May, "JUN": time.June,
	"JUL": time.July, "AUG": time.August, "SEP": time.September,
	"OCT": time.October, "NOV": time.November, "DEC": time.December,
}

// parseMonth преобразует строку с названием месяца в его числовое представление.
func parseMonth(s string) time.Month {
	s = strings.TrimSpace(s)
	if len(s) < 3 {
		return 0 // Неопределенный месяц
	}
	monthAbbr := strings.ToUpper(s[:3])
	if month, ok := monthMap[monthAbbr]; ok {
		return month
	}
	return 0
}

// readLines читает все строки из io.Reader в срез.
func readLines(r io.Reader) ([]string, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

// uniqueLines удаляет дубликаты из отсортированного среза строк.
func uniqueLines(lines []string) []string {
	if len(lines) == 0 {
		return lines
	}
	result := make([]string, 0, len(lines))
	result = append(result, lines[0])
	for i := 1; i < len(lines); i++ {
		if lines[i] != lines[i-1] {
			result = append(result, lines[i])
		}
	}
	return result
}

func main() {
	// Определение и парсинг флагов
	k := flag.Int("k", 1, "sort by column N (1-based, tab-separated)")
	n := flag.Bool("n", false, "sort by numeric value")
	r := flag.Bool("r", false, "sort in reverse order")
	u := flag.Bool("u", false, "output only unique lines")
	M := flag.Bool("M", false, "sort by month name")
	b := flag.Bool("b", false, "ignore trailing blanks")
	c := flag.Bool("c", false, "check if data is sorted")
	h := flag.Bool("h", false, "sort by human-readable numbers (e.g., 2K 1G)")

	flag.Parse()

	config := Config{
		column:      *k,
		numeric:     *n,
		reverse:     *r,
		unique:      *u,
		monthSort:   *M,
		ignoreBlanks: *b,
		checkSorted: *c,
		humanNumeric: *h,
	}

	// Определение источника ввода: файл или STDIN
	var reader io.Reader
	var inputFileName string

	if flag.NArg() > 0 {
		inputFileName = flag.Arg(0)
		file, err := os.Open(inputFileName)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sort: error opening file: %v\n", err)
			os.Exit(1)
		}
		defer file.Close()
		reader = file
	} else {
		inputFileName = "stdin" // Для сообщений об ошибках
		reader = os.Stdin
	}

	// Чтение строк
	lines, err := readLines(reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sort: error reading input: %v\n", err)
		os.Exit(1)
	}

	// Создание и использование сортировщика
	sorter := NewSorter(lines, config)

	// Режим проверки (-c)
	if config.checkSorted {
		isSorted, lineNum := sorter.CheckSorted()
		if !isSorted {
			// Вывод сообщения об ошибке сортировки
			fmt.Fprintf(os.Stderr, "sort: %s:%d: disorder: %s\n", inputFileName, lineNum, sorter.lines[lineNum-1])
			os.Exit(1)
		}
		os.Exit(0) // Если отсортировано, молча выходим с кодом 0
	}
	
	// Основной режим сортировки
	sorter.Sort()

	// Обработка флага -u
	outputLines := sorter.lines
	if config.unique {
		outputLines = uniqueLines(outputLines)
	}

	// Вывод результата
	for _, line := range outputLines {
		fmt.Println(line)
	}
}