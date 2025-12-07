package parser

import (
	"bytes"
	"io"
	"net/url"
	"path/filepath"

	"gowget/internal/storage"

	"golang.org/x/net/html"
)

type Result struct {
	ModifiedHTML []byte
	Links        []*url.URL
}

type Processor struct {
	fs *storage.FileSystem
}

func New(fs *storage.FileSystem) *Processor {
	return &Processor{fs: fs}
}

// Process парсит HTML, находит ссылки для скачивания и переписывает их на локальные.
func (p *Processor) Process(baseURL *url.URL, r io.Reader) (*Result, error) {
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var links []*url.URL
	var f func(*html.Node)

	f = func(n *html.Node) {
		if n.Type == html.ElementNode {
			// Обрабатываем теги, содержащие ссылки
			processNode(n, baseURL, &links, p.fs)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(doc)

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		return nil, err
	}

	return &Result{
		ModifiedHTML: buf.Bytes(),
		Links:        links,
	}, nil
}

func processNode(n *html.Node, baseURL *url.URL, links *[]*url.URL, fs *storage.FileSystem) {
	var attrKey string
	var isResource bool

	switch n.Data {
	case "a":
		attrKey = "href"
		isResource = false
	case "link", "script", "img":
		attrKey = "href"
		if n.Data == "img" || n.Data == "script" {
			attrKey = "src"
		}
		isResource = true // CSS, JS, Images скачиваем всегда
	default:
		return
	}

	for i, a := range n.Attr {
		if a.Key == attrKey {
			parsedLink, err := baseURL.Parse(a.Val) // Резолвим относительные ссылки
			if err != nil {
				continue
			}

			// Проверяем, в том же ли мы домене
			if parsedLink.Hostname() != baseURL.Hostname() {
				continue // Внешние ссылки оставляем как есть
			}

			// Определяем, является ли целью HTML (грубая проверка, можно улучшить через HEAD запрос)
			isHTML := !isResource && (filepath.Ext(parsedLink.Path) == "" || filepath.Ext(parsedLink.Path) == ".html")

			// Добавляем в список для скачивания
			*links = append(*links, parsedLink)

			// Переписываем ссылку на локальную относительную
			relPath, err := fs.ComputeRelativePath(baseURL, parsedLink, isHTML)
			if err == nil {
				n.Attr[i].Val = relPath
			}
			break
		}
	}
}