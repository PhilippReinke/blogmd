package args

import "fmt"

var (
	GitCommit = "unknown"
)

type BuildCmd struct {
	Path string `arg:"positional" help:"path to an exisiting project (optional)"`
}

type NewCmd struct {
	Path string `arg:"positional,required" help:"path of new project"`
}

type ServeCmd struct {
	Path          string `arg:"positional" help:"path to an exisiting project (optional)"`
	OpenInBrowser bool   `arg:"-b,--browser" help:"open page in default browser"`
	Watch         bool   `arg:"-w,--watch" help:"automatic rebuild on file change"`
}

type PostCmd struct {
	Title string `arg:"positional,required" help:"title of the post"`
}

type Args struct {
	New   *NewCmd   `arg:"subcommand:new" help:"create a new project"`
	Build *BuildCmd `arg:"subcommand:build" help:"build an existing project"`
	Serve *ServeCmd `arg:"subcommand:serve" help:"serve an existing project"`
	Post  *PostCmd  `arg:"subcommand:post" help:"create a new post from root of the project"`
}

func (Args) Description() string {
	return "blogmd helps you to build a static blog from markdown files."
}

func (Args) Version() string {
	return fmt.Sprintf("GitCommit: %v", GitCommit)
}
