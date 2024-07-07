package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/PhilippReinke/blogmd/browser"
	"github.com/PhilippReinke/blogmd/parser"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/otiai10/copy"
)

const (
	// TODO: implemend dedupe for fsnotify
	watcherDedupeTime = 100 * time.Millisecond
)

var (
	requiredDirs  = []string{"posts", "static", "templates"}
	requiredFiles = []string{
		"templates/base.html",
		"templates/index.html",
		"templates/post.html",
		"templates/rss.html",
		"templates/tags.html",
	}
)

func NewExistingProject(baseDir string) project {
	return project{
		baseDir:      baseDir,
		buildDir:     filepath.Join(baseDir, "build"),
		postsDir:     filepath.Join(baseDir, "posts"),
		staticDir:    filepath.Join(baseDir, "static"),
		templatesDir: filepath.Join(baseDir, "templates"),
	}
}

type project struct {
	baseDir      string
	buildDir     string
	postsDir     string
	staticDir    string
	templatesDir string
}

type PostsInfo []PostInfo

type PostInfo struct {
	FilenameHTML string
	Title        string
	Date         string
	Tags         []string
}

type TagsInfo []string

// Verfiy verifies existence of all required directories and files
func (p project) Verify() error {
	for _, dir := range requiredDirs {
		path := filepath.Join(p.baseDir, dir)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("missing directory '%s' is required", path)
		}
	}
	for _, file := range requiredFiles {
		path := filepath.Join(p.baseDir, file)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return fmt.Errorf("missing file '%s' is required", path)
		}
	}
	return nil
}

func (p project) Build() error {
	// verify project structure
	if err := p.Verify(); err != nil {
		return fmt.Errorf("invalid project: %v", err)
	}

	// parse posts
	posts, err := parser.ReadPosts(p.baseDir)
	if err != nil {
		return fmt.Errorf("could not read posts: %v", err)
	}

	// create build folder
	if err := os.RemoveAll(p.buildDir); err != nil {
		return fmt.Errorf("could not remove directory: %v\n", err)
	}
	if err := os.MkdirAll(p.buildDir, 0700); err != nil {
		return fmt.Errorf("could not create directory '%v': %v\n", p.buildDir, err)
	}

	// home page
	outPath := filepath.Join(p.buildDir, "index.html")
	baseTmplPath := filepath.Join(p.templatesDir, "base.html")
	indexTmplPath := filepath.Join(p.templatesDir, "index.html")
	if err := fillAndSaveTemplate(outPath, baseTmplPath, indexTmplPath, p.assemblePostsInfo(posts)); err != nil {
		return fmt.Errorf("could not fill and save template: %v", err)
	}

	// tags page
	outPath = filepath.Join(p.buildDir, "tags.html")
	tagsTmplPath := filepath.Join(p.templatesDir, "tags.html")
	if err := fillAndSaveTemplate(outPath, baseTmplPath, tagsTmplPath, p.assembleTagsInfo(posts)); err != nil {
		return fmt.Errorf("could not fill and save template: %v", err)
	}

	// rss page
	outPath = filepath.Join(p.buildDir, "rss.html")
	rssTmplPath := filepath.Join(p.templatesDir, "rss.html")
	if err := fillAndSaveTemplate(outPath, baseTmplPath, rssTmplPath, p.assemblePostsInfo(posts)); err != nil {
		return fmt.Errorf("could not fill and save template: %v", err)
	}

	// process posts
	for filename, post := range posts {
		outPath := filepath.Join(p.buildDir, p.postFilenameHTML(filename))
		postTmplPath := filepath.Join(p.templatesDir, "post.html")
		if err := fillAndSaveTemplate(outPath, baseTmplPath, postTmplPath, post); err != nil {
			return fmt.Errorf("could not fill and save template: %v", err)
		}
	}

	// copy static content
	if err = copy.Copy(p.staticDir, p.buildDir); err != nil {
		return fmt.Errorf("could not copy static content from '%v': %v\n", p.staticDir, err)
	}

	return nil
}

func (p project) assemblePostsInfo(posts parser.Posts) PostsInfo {
	var postsInfo PostsInfo
	for filename, post := range posts {
		info := PostInfo{
			FilenameHTML: p.postFilenameHTML(filename),
			Title:        post.Title,
			Date:         post.Date,
			Tags:         post.Tags,
		}
		postsInfo = append(postsInfo, info)
	}
	// sort posts by starting with newest
	sort.Slice(postsInfo, func(i, j int) bool {
		return postsInfo[i].Date > postsInfo[j].Date
	})
	return postsInfo
}

func (p project) assembleTagsInfo(posts parser.Posts) TagsInfo {
	tags := make(map[string]byte)
	for _, post := range posts {
		for _, tag := range post.Tags {
			tags[tag] = byte(0)
		}
	}
	var tagsInfo TagsInfo
	for k := range tags {
		tagsInfo = append(tagsInfo, k)
	}
	// sort alphabetically
	sort.Slice(tagsInfo, func(i, j int) bool {
		return tagsInfo[i] < tagsInfo[j]
	})
	return tagsInfo
}

func (p project) postFilenameHTML(postFilenameMD string) string {
	return fmt.Sprintf("%v.html", postFilenameMD[:len(postFilenameMD)-len(".md")])
}

func fillAndSaveTemplate(outPath, tmplBasePath, tmplContentPath string, data interface{}) error {
	out, err := os.Create(outPath)
	defer out.Close()
	if err != nil {
		return fmt.Errorf("could not create file '%v': %v\n", outPath, err)
	}

	tmpl := template.New("").Funcs(template.FuncMap{"add": func(x, y int) int {
		return x + y
	}})

	tmpl, err = tmpl.New("").ParseFiles(tmplContentPath, tmplBasePath)
	if err != nil {
		return fmt.Errorf("could not parse template files '%v' and '%v': %v\n", tmplContentPath, tmplBasePath, err)
	}

	if err := tmpl.ExecuteTemplate(out, "base", data); err != nil {
		return fmt.Errorf("could not execute template: %v\n", err)
	}
	return nil
}

func (p project) Serve(openInBrowser, withWatch bool) error {
	// build project for the first time
	fmt.Println("Building project...")
	if err := p.Build(); err != nil {
		return fmt.Errorf("cloud not build project: %v", err)
	}

	// rebuild on file change
	if withWatch {
		fmt.Println("Setting up watcher...")
		watcher, err := p.setupWatcher()
		if err != nil {
			return fmt.Errorf("could not setup watcher: %v", err)
		}
		defer watcher.Close()
	}

	// already open page in browser if desired
	fmt.Println("Serving blog under http://localhost:8080...")
	if openInBrowser {
		browser.OpenURL("http://localhost:8080")
	}

	// serve build folder
	e := echo.New()
	e.Static("/", p.buildDir)
	e.HideBanner = true
	return e.Start(":8080")
}

func (p project) setupWatcher() (*fsnotify.Watcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("could not setup watcher: %v", err)
	}

	// handle change events
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if dotFile(event.Name) {
					// ignore dot files
					continue
				}
				if event.Has(fsnotify.Write) {
					fmt.Printf("File modification in %v detected\n", event.Name)
					fmt.Println("Rebuilding project...")
					if err := p.Build(); err != nil {
						fmt.Printf("Cloud not build project: %v\n", err)
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("Watcher encountered error:", err)
			}
		}
	}()

	// watch all required directories
	for _, dir := range requiredDirs {
		path := filepath.Join(p.baseDir, dir)
		if err = watcher.Add(path); err != nil {
			return nil, fmt.Errorf("could not add directory '%v' to watcher: %v", path, err)
		}
	}

	return watcher, nil
}

func dotFile(path string) bool {
	filename := filepath.Base(path)
	return filename[0] == '.'
}

func (p project) CreatePost(title string) error {
	if err := p.Verify(); err != nil {
		return fmt.Errorf("you are not within a valid project: %v", err)
	}

	path := filepath.Join(p.postsDir, filenameFromTitle(title))

	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("file '%s' already exists", path)
	}

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("could not create file '%s': %v", path, err)
	}
	defer file.Close()

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04")

	file.WriteString("---\n")
	file.WriteString("title: " + title + "\n")
	file.WriteString("date: " + formattedTime + "\n")
	file.WriteString("tags: " + "[]\n")
	file.WriteString("---\n")

	return nil
}

func filenameFromTitle(title string) string {
	return strings.ToLower(strings.ReplaceAll(title, " ", "-")) + ".md"
}
