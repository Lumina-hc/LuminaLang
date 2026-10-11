package stdlib

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed lib
var FS embed.FS

func Extract(dir string) error {
	sub, err := fs.Sub(FS, "lib")
	if err != nil {
		return err
	}
	return fs.WalkDir(sub, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := sub.(interface{ ReadFile(string) ([]byte, error) }).ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(dir, p), data, 0o644)
	})
}
