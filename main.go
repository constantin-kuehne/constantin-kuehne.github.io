package main

import (
	"context"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	t "constantin-kuehne.github.io/templates"
	"github.com/a-h/templ"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/gosimple/slug"
)

type File struct {
	path string
	info fs.FileInfo
}

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

func findMarkdownFiles(rootDir string) []File {
	mdFiles := make([]File, 0)
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if (filepath.Ext(path) == ".md") && (path != "README.md") {
			mdFiles = append(mdFiles, File{path, info})
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
		log.Fatalf("Cannot write file: %s", err)
	}
}

func Unsafe(html string) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		_, err = io.WriteString(w, html)
		return
	})
}

type FilePath string

type DirPath string

type renderMapFuncType func(site t.SiteInfo, args ...any) (FilePath, templ.Component)

type renderMapType map[DirPath]renderMapFuncType

var renderMap renderMapType

func getTitle(site t.SiteInfo) string {
	var title string
	if strings.Contains(site.FileName, "-") {
		title = slug.Make(strings.Split(site.FileName, "-")[1])
	} else {
		title = slug.Make(site.FileName)
	}

	return title
}

func generateNewFilePath(title string, site t.SiteInfo) FilePath {
	newFilePath := filepath.Join("docs/", site.Dir, title+".html")
	os.MkdirAll(filepath.Dir(newFilePath), os.ModePerm)
	return FilePath(newFilePath)
}

func generatePostsFile(site t.SiteInfo, args ...any) (FilePath, templ.Component) {
	title := getTitle(site)
	return generateNewFilePath(title, site), t.FooterComponent()
}

func generatePostFile(site t.SiteInfo, args ...any) (FilePath, templ.Component) {
	title := getTitle(site)
	return generateNewFilePath(title, site), t.FooterComponent()
}

func main() {
	renderMap = map[DirPath]renderMapFuncType{"./": generatePostsFile, "posts/": generatePostFile}
	for _, file := range findMarkdownFiles("blog/") {
		md, err := os.ReadFile(file.path)
		if err != nil {
			log.Fatalf("Cannot read file %s", err)
		}
		filePath := strings.TrimPrefix(file.path, "blog/")
		dir := filepath.Dir(filePath)
		fileName := strings.TrimSuffix(filepath.Base(file.info.Name()), filepath.Ext(filePath))

		html := markdownToHTML(md)
		site := t.SiteInfo{
			Date:     file.info.ModTime(),
			FileName: fileName,
			Dir:      dir,
			Content:  Unsafe(string(html)),
		}

		handler, exists := renderMap[DirPath(dir)+"/"]
		if !exists {
			log.Fatalf("No handler for dir: %s", dir)
		}
		newFilePath, component := handler(site)
		f, err := os.Create(string(newFilePath))
		if err != nil {
			log.Fatalf("Cannot create file: %s; err: %s", newFilePath, err)
		}

		err = component.Render(context.Background(), f)
		if err != nil {
			log.Fatalf("Cannot Render component: %s; err: %s", component, err)
		}

	}
}
