package main

import (
	"fmt"
	"os"

	"github.com/PhilippReinke/blogmd/args"
	"github.com/PhilippReinke/blogmd/project"

	"github.com/alexflint/go-arg"
)

func main() {
	var args args.Args
	arg.MustParse(&args)

	switch {
	case args.Build != nil:
		project := project.NewExistingProject(args.Build.Path)
		if err := project.Build(); err != nil {
			fmt.Println("Could not build project:", err)
			os.Exit(1)
		}

	case args.New != nil:
		if err := project.Create(args.New.Path); err != nil {
			fmt.Println("Could not create new project:", err)
			os.Exit(1)
		}

	case args.Serve != nil:
		project := project.NewExistingProject(args.Serve.Path)
		if err := project.Serve(args.Serve.OpenInBrowser, args.Serve.Watch); err != nil {
			fmt.Println("Could not serve project:", err)
			os.Exit(1)
		}

	case args.Post != nil:
		project := project.NewExistingProject("")
		if err := project.CreatePost(args.Post.Title); err != nil {
			fmt.Println("Could not create new post:", err)
			os.Exit(1)
		}
	}
}
