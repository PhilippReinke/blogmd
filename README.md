# blogmd

A simple blog builder that converts markdown posts to a static website. So,
basically what [Hugo](https://github.com/gohugoio/hugo) does and which I
recommend over `blogmd` as this is just a little fun project for me.

## Usage

```sh
# new blog (with example project)
blogmd new my-blog

# create new blog post
cd my-blog
blogmd post "Hello blogmd"

# build
blogmd build <path to project>
# omit path when already in root of project

# serve
blogmd serve -w -b <path to project>
# omit path when already in root of project
# -w means watch which will rebuild project on change
# -b means open in default browser

# also use
blogmd --help
blogmd <command> --help
```

## Installation

```sh
# option 1
git clone https://github.com/PhilippReinke/blogmd.git
cd blogmd
make install

# option 2
go install github.com/PhilippReinke/blogmd@latest
```

## Required project structure

The following directory structure is required

```
my-site
|- posts
|- static
|- templates
 |- base.html
 |- index.html
 |- post.html
 |- rss.html
 |- tags.html
```

| directory or file    | description                                         |
|:---------------------|:----------------------------------------------------|
| posts                | Blog posts in markdown format                       |
| static               | Things like favicon.ico or robots.txt               |
| templates            | Go HTML templates that will be filled with markdown |
| templates/base.html  | required                                            |
| templates/index.html | required                                            |
| templates/post.html  | required                                            |
| templates/rss.html   | required                                            |
| templates/tags.html  | required                                            |
