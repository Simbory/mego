package views

import (
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type tplCache struct {
	tpl *template.Template
	err error
}

type file struct {
	viewExt string
	root    string
	files   map[string][]string
}

func (vf *file) visit(paths string, f os.FileInfo, err error) error {
	if f == nil {
		return err
	}
	if f.IsDir() || (f.Mode()&os.ModeSymlink) > 0 {
		return nil
	}
	if !strings.HasSuffix(strings.ToLower(paths), vf.viewExt) {
		return nil
	}
	replace := strings.NewReplacer("\\", "/")
	a := []byte(paths)
	a = a[len(vf.root):]
	file := strings.TrimLeft(replace.Replace(string(a)), "/")
	subDir := filepath.Dir(file)
	if _, ok := vf.files[subDir]; ok {
		vf.files[subDir] = append(vf.files[subDir], file)
	} else {
		m := make([]string, 1)
		m[0] = file
		vf.files[subDir] = m
	}
	return nil
}
