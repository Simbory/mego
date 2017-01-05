package view

import (
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/Simbory/mego"
	"github.com/Simbory/mego/watcher"
	"io"
)

type view struct {
	tpl *template.Template
	err error
}

type viewFile struct {
	viewExt string
	root    string
	files   map[string][]string
}

func (vf *viewFile) visit(paths string, f os.FileInfo, err error) error {
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

type viewContainer struct {
	viewExt      string
	initialized  bool
	viewDir      string
	views        map[string]*view
	funcMaps     template.FuncMap
	compilerLock sync.RWMutex
}

func (vc *viewContainer) addViewFunc(name string, f interface{}) {
	if len(name) < 1 || f == nil {
		return
	}
	if vc.funcMaps == nil {
		vc.funcMaps = make(template.FuncMap)
	}
	vc.funcMaps[name] = f
}

func (vc *viewContainer) addView(name string, v *view) {
	if vc.views == nil {
		vc.views = make(map[string]*view)
	}
	vc.views[name] = v
}

func (vc *viewContainer) getView(name string) *view {
	v, ok := vc.views[name]
	if !ok {
		return nil
	}
	return v
}

func (vc *viewContainer) getTemplate(file, viewExt string, funcMap template.FuncMap, others ...string) (t *template.Template, err error) {
	t = template.New(file)
	if funcMap != nil {
		t.Funcs(funcMap)
	}
	var subMods [][]string
	t, subMods, err = vc.getTemplateDeep(file, viewExt, "", t)
	if err != nil {
		return nil, err
	}
	t, err = vc.getTemplateLoop(t, viewExt, subMods, others...)

	if err != nil {
		return nil, err
	}
	return
}

func (vc *viewContainer) getTemplateDeep(file, viewExt, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string
	if filepath.HasPrefix(file, "../") {
		fileAbsPath = filepath.Join(vc.viewDir, filepath.Dir(parent), file)
	} else {
		fileAbsPath = filepath.Join(vc.viewDir, file)
	}
	stat, err := os.Stat(fileAbsPath)
	if err != nil || stat.IsDir() {
		return nil, [][]string{}, fmt.Errorf("Cannot open the view file %s", file)
	}
	data, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile("{{" + "[ ]*template[ ]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			look := t.Lookup(m[1])
			if look != nil {
				continue
			}
			if !strings.HasSuffix(strings.ToLower(m[1]), viewExt) {
				continue
			}
			t, _, err = vc.getTemplateDeep(m[1], viewExt, file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func (vc *viewContainer) getTemplateLoop(t0 *template.Template, viewExt string, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = t0
	for _, m := range subMods {
		if len(m) == 2 {
			tpl := t.Lookup(m[1])
			if tpl != nil {
				continue
			}
			//first check filename
			for _, otherFile := range others {
				if otherFile == m[1] {
					var subMods1 [][]string
					t, subMods1, err = vc.getTemplateDeep(otherFile, viewExt, "", t)
					if err != nil {
						return nil, err
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = vc.getTemplateLoop(t, viewExt, subMods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherFile := range others {
				fileAbsPath := filepath.Join(vc.viewDir, otherFile)
				data, err := ioutil.ReadFile(fileAbsPath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile("{{" + "[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = vc.getTemplateDeep(otherFile, viewExt, "", t)
						if err != nil {
							return nil, err
						} else if subMods1 != nil && len(subMods1) > 0 {
							t, err = vc.getTemplateLoop(t, viewExt, subMods1, others...)
						}
						break
					}
				}
			}
		}
	}
	return
}

func (vc *viewContainer) compileViews() error {
	vc.compilerLock.Lock()
	defer vc.compilerLock.Unlock()
	dir := viewDir()
	vc.viewDir = dir
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("Cannot open view path: %s", dir)
	}
	vf := &viewFile{
		root:    dir,
		files:   make(map[string][]string),
		viewExt: vc.viewExt,
	}
	err := filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		return vf.visit(path, f, err)
	})
	if err != nil {
		return err
	}
	for _, v := range vf.files {
		for _, file := range v {
			t, err := vc.getTemplate(file, vf.viewExt, vc.funcMaps, v...)
			v := &view{tpl: t, err: err}
			vc.addView(file, v)
		}
	}
	vc.initialized = true
	return nil
}

func (vc *viewContainer) renderView(writer io.Writer, viewPath string, viewData interface{}) error {
	if len(viewPath) < 1 {
		return errors.New("The parameger 'viewPath' cannot be empty")
	}
	if !strings.HasSuffix(viewPath, vc.viewExt) {
		viewPath = viewPath + vc.viewExt
	}
	tpl := vc.getView(viewPath)
	if tpl == nil {
		return fmt.Errorf("The viewPath '%s' canot be found", viewPath)
	}
	if tpl.err != nil {
		return tpl.err
	}
	if tpl.tpl == nil {
		return fmt.Errorf("The viewPath '%s' canot be found", viewPath)
	}
	err := tpl.tpl.Execute(writer, viewData)
	if err != nil {
		return err
	}
	return nil
}

var (
	singleton = &viewContainer{
		viewExt:     ".gohtml",
		initialized: false,
	}
	featuredViewDir = defaultViewDir()
)

func workingDir() string {
	p, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return p
}

func viewDir() string {
	if len(featuredViewDir) > 0 {
		return featuredViewDir
	}
	return defaultViewDir()
}

func init() {
	mego.OnServerStart(func() {
		dir := viewDir()
		stat,err := os.Stat(dir)
		if err != nil {
			panic(err)
		}
		if !stat.IsDir() {
			panic(fmt.Errorf("the view directory does not exists: %s", dir))
		}
		fsWatcher := watcher.Singleton()
		fsWatcher.AddHandler(&fsViewHandler{fsWatcher})
		fsWatcher.Start()
		singleton.compileViews()
		fsWatcher.AddWatch(dir)
		filepath.Walk(dir, func(p string, info os.FileInfo, er error) error {
			if info.IsDir() {
				fsWatcher.AddWatch(p)
			}
			return nil
		})
	})
}

// UseView use mego view component
func UseView(dir string) {
	mego.AssertNotLock()
	if len(dir) > 0 {
		featuredViewDir = fixPath(dirSlash(dir))
	}
	AddViewFunc("include", include)
}

// SetViewExt the the view file extension. The default view file extension is '.gohtml'
func SetViewExt(ext string) {
	mego.AssertNotLock()
	if len(ext) > 0 {
		if !strings.HasPrefix(ext, ".") {
			ext = "." + ext
		}
		singleton.viewExt = ext
	}
}

// AddViewFunc add view function to the view engine
func AddViewFunc(name string, f interface{}) {
	mego.AssertNotLock()
	singleton.addViewFunc(name, f)
}

// View render the view file and return the mego result
func View(viewPath string, data interface{}) mego.Result {
	return &viewResult{
		data:     data,
		viewName: viewPath,
	}
}

// Render render the view file and write the result bytes to the writer
func Render(writer io.Writer, viewName string, data interface{}) error {
	if !singleton.initialized {
		return errors.New("Cannot call this function before it is initialized")
	}
	err := singleton.renderView(writer, viewName, data)
	if err != nil {
		return err
	}
	return nil
}