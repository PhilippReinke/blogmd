package parser

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/PhilippReinke/godiff/dir"
	"github.com/russross/blackfriday/v2"
	"gopkg.in/yaml.v3"
)

// Posts maps filename to data about the post
type Posts map[string]Post

type Post struct {
	Title   string   `yaml:"title"`
	Date    string   `yaml:"date"`
	Tags    []string `yaml:"tags"`
	Content string
}

func ReadPosts(projectDir string) (Posts, error) {
	dir, err := dir.ReadDir(filepath.Join(projectDir, "posts"))
	if err != nil {
		return nil, err
	}

	posts := make(Posts)
	for filename, fileInfo := range dir.Files {
		if !fileInfo.Mode().IsRegular() {
			// e.g. directory
			continue
		}
		if filepath.Ext(filename) != ".md" {
			continue
		}

		currentFilePath := filepath.Join(dir.Path, filename)
		post, err := readPost(currentFilePath)
		if err != nil {
			return nil, err
		}
		posts[filename] = post
	}

	return posts, nil
}

func readPost(path string) (Post, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return Post{}, fmt.Errorf("error reading file: %v\n", err)
	}

	parts := strings.SplitN(string(content), "---", 3)
	if len(parts) < 3 {
		return Post{}, fmt.Errorf("file does not contain valid YAML front matter.")
	}

	// parse yamls
	yamlContent := parts[1]
	markdownContent := parts[2]
	var post Post
	if err = yaml.Unmarshal([]byte(yamlContent), &post); err != nil {
		return Post{}, fmt.Errorf("error parsing YAML: %v\n", err)
	}

	// markdown to html
	html := string(blackfriday.Run([]byte(markdownContent)))
	post.Content = html

	return post, nil
}
