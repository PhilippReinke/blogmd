package project

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/PhilippReinke/blogmd/examples"
)

// Create creates a default project with according directories and files under a
// given path
func Create(path string) error {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return fmt.Errorf("a directory '%s' already exists", path)
	}
	if err := copyFS(path, examples.DefaultDir, "default"); err != nil {
		return fmt.Errorf("could not copy default project: %v\n", err)
	}
	return nil
}

func copyFS(dir string, fsys fs.FS, relPath string) error {
	return fs.WalkDir(fsys, ".", func(path string, d fs.DirEntry, err error) error {
		relTarget, _ := filepath.Rel(relPath, filepath.FromSlash(path))
		targ := filepath.Join(dir, relTarget)
		if d.IsDir() {
			if err := os.MkdirAll(targ, 0777); err != nil {
				return err
			}
			return nil
		}
		r, err := fsys.Open(path)
		if err != nil {
			return err
		}
		defer r.Close()
		info, err := r.Stat()
		if err != nil {
			return err
		}
		w, err := os.OpenFile(targ, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0666|info.Mode()&0777)
		if err != nil {
			return err
		}
		if _, err := io.Copy(w, r); err != nil {
			w.Close()
			return fmt.Errorf("copying %s: %v", path, err)
		}
		if err := w.Close(); err != nil {
			return err
		}
		return nil
	})
}
