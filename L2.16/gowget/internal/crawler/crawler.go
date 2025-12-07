package crawler

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"

	"gowget/internal/downloader"
	"gowget/internal/parser"
	"gowget/internal/storage"
)

type Config struct {
	MaxDepth       int
	MaxConcurrency int
	Timeout        time.Duration
	OutputDir      string
}

type Crawler struct {
	cfg        Config
	client     *downloader.Client
	fs         *storage.FileSystem
	parser     *parser.Processor
	visited    sync.Map // Thread-safe map для посещенных URL
	sem        chan struct{} // Семафор для ограничения горутин
	wg         sync.WaitGroup
}

func New(cfg Config) *Crawler {
	fs := storage.New(cfg.OutputDir)
	return &Crawler{
		cfg:     cfg,
		client:  downloader.New(cfg.Timeout),
		fs:      fs,
		parser:  parser.New(fs),
		sem:     make(chan struct{}, cfg.MaxConcurrency),
	}
}

func (c *Crawler) Start(startURL string) error {
	u, err := url.Parse(startURL)
	if err != nil {
		return fmt.Errorf("invalid start URL: %w", err)
	}

	c.wg.Add(1)
	go c.visit(u, 0)
	c.wg.Wait()
	return nil
}

func (c *Crawler) visit(u *url.URL, depth int) {
	defer c.wg.Done()

	// Нормализация URL (убираем фрагменты #anchor)
	u.Fragment = ""
	uStr := u.String()

	if _, loaded := c.visited.LoadOrStore(uStr, true); loaded {
		return
	}

	if depth > c.cfg.MaxDepth {
		return
	}

	// Занимаем слот в семафоре
	c.sem <- struct{}{}
	defer func() { <-c.sem }()

	log.Printf("[DOWNLOADING] %s", uStr)

	resp, err := c.client.Fetch(uStr)
	if err != nil {
		log.Printf("[ERROR] fetching %s: %v", uStr, err)
		return
	}
	defer resp.Body.Close()

	// Определяем тип контента
	contentType := resp.ContentType
	isHTML := strings.Contains(contentType, "text/html")

	var reader io.Reader = resp.Body
	var newLinks []*url.URL

	// Если HTML, нужно парсить и менять ссылки
	if isHTML {
		res, err := c.parser.Process(u, resp.Body)
		if err != nil {
			log.Printf("[ERROR] parsing HTML %s: %v", uStr, err)
			// Если ошибка парсинга, пробуем сохранить как есть
		} else {
			reader = bytes.NewReader(res.ModifiedHTML)
			newLinks = res.Links
		}
	}

	// Сохраняем файл
	localPath, err := c.fs.Save(u, reader, isHTML)
	if err != nil {
		log.Printf("[ERROR] saving file %s: %v", uStr, err)
	} else {
		log.Printf("[SAVED] %s -> %s", uStr, localPath)
	}

	// Рекурсивно обходим ссылки
	for _, link := range newLinks {
		// Ресурсы (картинки, css) качаем всегда, страницы - только если позволяет глубина
		// Здесь простая логика: увеличиваем глубину только для переходов по ссылкам <a>
		// Для ресурсов depth можно не увеличивать или обрабатывать отдельно.
		// В этой реализации depth увеличивается для всех.
		if depth+1 <= c.cfg.MaxDepth {
			c.wg.Add(1)
			go c.visit(link, depth+1)
		}
	}
}