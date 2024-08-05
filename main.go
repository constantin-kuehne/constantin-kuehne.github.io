package main

import (
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

func markdownToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}

func findMarkdownFiles(rootDir string) []string {
	mdFiles := make([]string, 0)
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if (filepath.Ext(path) == ".md") && (path != "README.md") {
			mdFiles = append(mdFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Could not walk rootDir")
	}
	return mdFiles
}

func saveHtml(path string, content []byte) {
	if err := os.WriteFile(path, content, 0644); err != nil {
		log.Fatal("Cannot write file")
	}
}

func main() {
	for _, filePath := range findMarkdownFiles(".") {
		md, err := os.ReadFile(filePath)
		if err != nil {
			log.Fatal("Cannot read file")
		}
		dir := filepath.Dir(filePath)
		fileName := strings.TrimSuffix(filepath.Base(filePath), filepath.Ext(filePath))
		newFileName := filepath.Join(dir, fileName+".html")

		html := markdownToHTML(md)
		saveHtml(newFileName, html)
	}
}
