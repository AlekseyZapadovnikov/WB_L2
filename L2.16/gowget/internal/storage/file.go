package storage

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type FileSystem struct {
	BaseDir string
}

func New(baseDir string) *FileSystem {
	return &FileSystem{BaseDir: baseDir}
}

// Save сохраняет данные (reader) по указанному URL.
// Возвращает локальный путь, куда был сохранен файл.
func (fs *FileSystem) Save(u *url.URL, reader io.Reader, isHTML bool) (string, error) {
	localPath := fs.GetLocalPath(u, isHTML)
	absPath := filepath.Join(fs.BaseDir, localPath)

	if err := os.MkdirAll(filepath.Dir(absPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	f, err := os.Create(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()

	if _, err := io.Copy(f, reader); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	return localPath, nil
}

// GetLocalPath конвертирует URL в относительный путь файловой системы.
func (fs *FileSystem) GetLocalPath(u *url.URL, isHTML bool) string {
	// Убираем схему и хост для создания структуры папок, но хост используем как корневую папку сайта
	host := u.Hostname()
	path := u.Path

	if path == "" || path == "/" {
		path = "/index.html"
	} else if strings.HasSuffix(path, "/") {
		path = path + "index.html"
	} else if isHTML && filepath.Ext(path) == "" {
		// Если это HTML, но нет расширения, добавляем .html
		path = path + ".html"
	}

	// Очищаем путь от потенциально опасных символов (упрощенно)
	// В реальном wget логика сложнее (обработка query params и т.д.)
	return filepath.Join(host, path)
}

// ComputeRelativePath вычисляет относительный путь от sourceURL к targetURL
// для использования в href/src атрибутах.
func (fs *FileSystem) ComputeRelativePath(sourceURL, targetURL *url.URL, targetIsHTML bool) (string, error) {
	srcPath := fs.GetLocalPath(sourceURL, true) // source всегда HTML в нашем контексте
	tgtPath := fs.GetLocalPath(targetURL, targetIsHTML)

	// Получаем директорию исходного файла
	baseDir := filepath.Dir(srcPath)
	
	relPath, err := filepath.Rel(baseDir, tgtPath)
	if err != nil {
		return "", err
	}
	
	// filepath.Rel использует системный разделитель, для URL нужен "/"
	return filepath.ToSlash(relPath), nil
}