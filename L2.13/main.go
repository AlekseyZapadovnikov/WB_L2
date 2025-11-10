// main.go
// Пакет main реализует утилиту для извлечения полей из текстового потока,
// аналог UNIX-утилиты cut.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
)

// Config хранит конфигурацию, полученную из флагов командной строки.
type Config struct {
	// fields - это отсортированный срез уникальных 1-based номеров полей для вывода.
	fields []int
	// delimiter - это разделитель полей.
	delimiter string
	// separatedOnly - флаг, указывающий, что нужно выводить только строки с разделителем.
	separatedOnly bool
}

// parseFields разбирает строку с полями (например, "1,3-5") и возвращает
// отсортированный срез уникальных 1-based номеров полей.
func parseFields(fieldsStr string) ([]int, error) {
	if fieldsStr == "" {
		return nil, errors.New("fields option cannot be empty")
	}

	// Используем map для автоматического удаления дубликатов.
	fieldSet := make(map[int]struct{})

	// Разбираем части, разделенные запятыми.
	parts := strings.Split(fieldsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "-") {
			// Обработка диапазона (например, "3-5").
			rangeParts := strings.Split(part, "-")
			if len(rangeParts) != 2 {
				return nil, fmt.Errorf("invalid range format: %q", part)
			}

			start, err := strconv.Atoi(rangeParts[0])
			if err != nil {
				return nil, fmt.Errorf("invalid start of range in %q: %w", part, err)
			}
			end, err := strconv.Atoi(rangeParts[1])
			if err != nil {
				return nil, fmt.Errorf("invalid end of range in %q: %w", part, err)
			}

			if start < 1 || end < 1 {
				return nil, errors.New("field numbers must be positive")
			}
			if start > end {
				return nil, fmt.Errorf("invalid range: start %d is greater than end %d", start, end)
			}

			for i := start; i <= end; i++ {
				fieldSet[i] = struct{}{}
			}
		} else {
			// Обработка одиночного номера поля.
			fieldNum, err := strconv.Atoi(part)
			if err != nil {
				return nil, fmt.Errorf("invalid field number %q: %w", part, err)
			}
			if fieldNum < 1 {
				return nil, errors.New("field numbers must be positive")
			}
			fieldSet[fieldNum] = struct{}{}
		}
	}

	// Преобразуем map в срез.
	fields := make([]int, 0, len(fieldSet))
	for k := range fieldSet {
		fields = append(fields, k)
	}
	// Сортируем срез для последовательного вывода полей.
	sort.Ints(fields)

	return fields, nil
}

// runCut выполняет основную логику утилиты: читает из reader, обрабатывает и пишет в writer.
func runCut(reader io.Reader, writer io.Writer, cfg Config) error {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()

		// Проверяем, содержит ли строка разделитель.
		containsDelimiter := strings.Contains(line, cfg.delimiter)

		// Если установлен флаг -s и разделителя нет, пропускаем строку.
		if cfg.separatedOnly && !containsDelimiter {
			continue
		}

		// Если разделителя нет (и флаг -s не установлен), выводим строку целиком,
		// как это делает стандартный cut.
		if !containsDelimiter {
			fmt.Fprintln(writer, line)
			continue
		}

		// Разбиваем строку на поля.
		parts := strings.Split(line, cfg.delimiter)
		resultFields := make([]string, 0, len(cfg.fields))

		// Итерируемся по списку необходимых полей.
		for _, fieldNum := range cfg.fields {
			// fieldNum - 1, так как поля 1-based, а индексы в срезе 0-based.
			idx := fieldNum - 1
			if idx < len(parts) {
				// Если такое поле существует, добавляем его в результат.
				resultFields = append(resultFields, parts[idx])
			}
		}

		// Соединяем выбранные поля тем же разделителем и выводим.
		outputLine := strings.Join(resultFields, cfg.delimiter)
		fmt.Fprintln(writer, outputLine)
	}

	// Проверяем на ошибки сканирования (например, при чтении файла).
	return scanner.Err()
}

func main() {
	// Определение и парсинг флагов.
	var (
		fieldsStr     = flag.String("f", "", `select only these fields. Use comma-separated lists and ranges (e.g., "1,3-5")`)
		delimiterStr  = flag.String("d", "\t", "use DELIM instead of TAB for field delimiter")
		separatedOnly = flag.Bool("s", false, "do not print lines not containing delimiters")
	)
	flag.Parse()

	// Валидация аргументов.
	if *fieldsStr == "" {
		fmt.Fprintln(os.Stderr, "error: -f flag is required")
		flag.Usage()
		os.Exit(1)
	}
	
	// Разбираем строку с номерами полей.
	fields, err := parseFields(*fieldsStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing fields: %v\n", err)
		os.Exit(1)
	}

	// Создаем конфигурацию.
	config := Config{
		fields:        fields,
		delimiter:     *delimiterStr,
		separatedOnly: *separatedOnly,
	}

	// Запускаем основную логику, читая из STDIN и пиша в STDOUT.
	if err := runCut(os.Stdin, os.Stdout, config); err != nil {
		fmt.Fprintf(os.Stderr, "error during processing: %v\n", err)
		os.Exit(1)
	}
}