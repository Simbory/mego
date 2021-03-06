package views

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/simbory/mego/assert"
	"github.com/simbory/mego/fswatcher"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"path"
)

type ViewEngine struct {
	viewDir  string
	viewExt  string
	funcMap  template.FuncMap
	locker   sync.RWMutex
	viewMap  map[string]*tplCache
	compiled bool

	watcher *fswatcher.FileWatcher
}

// AddFunc add new view func to the view files
func (engine *ViewEngine) AddFunc(name string, viewFunc interface{}) {
	if len(name) < 1 || viewFunc == nil {
		return
	}
	if engine.funcMap == nil {
		engine.funcMap = make(template.FuncMap)
	}
	if f, ok := engine.funcMap[name]; !ok || f == nil {
		engine.funcMap[name] = viewFunc
	}
}

func (engine *ViewEngine) addView(name string, v *tplCache) {
	if len(name) == 0 || v == nil {
		return
	}
	if engine.viewMap == nil {
		engine.viewMap = map[string]*tplCache{}
	}
	engine.viewMap[name] = v
}

func (engine *ViewEngine) getView(name string) (*template.Template, error) {
	if engine.viewMap == nil {
		return nil, nil
	}
	if strings.HasPrefix(name, "/") {
		name = strings.TrimLeft(name, "/")
	}
	cacheView := engine.viewMap[name]
	if cacheView == nil {
		return nil, nil
	}
	return cacheView.tpl, cacheView.err
}

func (engine *ViewEngine) getDeep(file, parent string, t *template.Template) (*template.Template, [][]string, error) {
	var fileAbsPath string

	if strings.HasPrefix(file, "../") || strings.HasPrefix(file, "./") {
		fileAbsPath = filepath.Join(engine.viewDir, filepath.Dir(parent), file)
		fileAbsPath = strings.Replace(fileAbsPath, "\\", "/", -1)
		if !strings.HasPrefix(fileAbsPath, engine.viewDir) {
			return nil, nil, errors.New("invalid view file:" + file)
		}
	} else if strings.HasPrefix(file, "/") {
		fileAbsPath = engine.viewDir + strings.TrimLeft(file, "/")
	} else {
		fileAbsPath = filepath.Join(engine.viewDir, file)
	}
	stat, err := os.Stat(fileAbsPath)
	if err != nil || stat.IsDir() {
		return nil, [][]string{}, fmt.Errorf("the partial view '%s' in '%s' cannot be found", file, parent)
	}
	data, err := ioutil.ReadFile(fileAbsPath)
	if err != nil {
		return nil, [][]string{}, err
	}
	t, err = t.New(file).Parse(string(data))
	if err != nil {
		return nil, [][]string{}, err
	}
	reg := regexp.MustCompile("[{]{2}[ \t]*template[ \t]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			name := m[1]
			if !strings.HasSuffix(strings.ToLower(name), engine.viewExt) {
				continue
			}
			look := t.Lookup(name)
			if look != nil {
				continue
			}
			t, _, err = engine.getDeep(name, file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func (engine *ViewEngine) getLoop(temp *template.Template, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = temp
	for _, m := range subMods {
		if len(m) == 2 {
			tpl := t.Lookup(m[1])
			if tpl != nil {
				continue
			}
			//check filename
			for _, otherFile := range others {
				if otherFile == m[1] {
					var subMods1 [][]string
					t, subMods1, err = engine.getDeep(otherFile, "", t)
					if err != nil {
						return nil, err
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = engine.getLoop(t, subMods1, others...)
					}
					break
				}
			}
		}
	}
	return
}

func (engine *ViewEngine) getTplCache(file string, others ...string) *tplCache {
	t := template.New(file)
	if engine.funcMap != nil {
		t.Funcs(engine.funcMap)
	}
	var subMods [][]string
	t, subMods, err := engine.getDeep(file, "", t)
	if err != nil {
		return &tplCache{nil, err}
	}
	t, err = engine.getLoop(t, subMods, others...)
	if err != nil {
		return &tplCache{nil, err}
	}
	return &tplCache{t, nil}
}

func (engine *ViewEngine) compile() error {
	engine.locker.Lock()
	defer engine.locker.Unlock()
	if engine.compiled {
		return nil
	}
	engine.AddFunc("include", engine.includeView)
	if _, err := os.Stat(engine.viewDir); err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("failed to open view directory '%s'", engine.viewDir)
	}
	vf := &file{
		root:    engine.viewDir,
		files:   make(map[string][]string),
		viewExt: engine.viewExt,
	}
	err := filepath.Walk(engine.viewDir, func(path string, f os.FileInfo, err error) error {
		return vf.visit(path, f, err)
	})
	if err != nil {
		return err
	}
	for _, v := range vf.files {
		for _, file := range v {
			v := engine.getTplCache(file, v...)
			engine.addView(file, v)
		}
	}
	engine.compiled = true
	return nil
}

func (engine *ViewEngine) includeView(viewName string, data interface{}) template.HTML {
	str, err := engine.RenderStr(viewName, data)
	assert.PanicErr(err)
	return template.HTML(str)
}

// Clear clear all the cache in the view engine. and re-compile the view files into memory at next call 'Render' or 'RenderStr'
func (engine *ViewEngine) Clear() {
	engine.locker.Lock()
	defer engine.locker.Unlock()

	engine.compiled = false
	engine.viewMap = nil
}

// Render render the view file with given data and then write the result to an io writer.
// viewPath: the relative view file path that will be rendered.
// viewData: the view data
// writer: the given io writer
func (engine *ViewEngine) Render(writer io.Writer, viewPath string, viewData interface{}) error {
	if writer == nil {
		return errors.New("invalid writer")
	}
	if len(viewPath) < 1 {
		return errors.New("invalid viewPath")
	}
	if !strings.HasSuffix(viewPath, engine.viewExt) {
		viewPath = viewPath + engine.viewExt
	}
	if !engine.compiled {
		err := engine.compile()
		if err != nil {
			return err
		}
	}
	tpl, err := engine.getView(viewPath)
	if err != nil {
		return err
	}
	if tpl == nil {
		return fmt.Errorf("the view file '%s' cannot be found", viewPath)
	}
	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, viewData)
	if err != nil {
		return err
	} else {
		io.Copy(writer, buf)
	}
	return nil
}

// RenderStr render the view file with given data then get the result.
// viewPath: the relative view file path that will be rendered
// viewData: the view data
func (engine *ViewEngine) RenderStr(viewPath string, viewData interface{}) (string, error) {
	buf := &bytes.Buffer{}
	err := engine.Render(buf, viewPath, viewData)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// NewEngine create a new view engine.
// rootDir: the root dir of the view files;
// ext: the view file extension(starts with "."), and the default file extension is '.gohtml';
func NewEngine(rootDir, ext string) (*ViewEngine, error) {
	if len(ext) == 0 {
		ext = ".gohtml"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	w, err := fswatcher.NewWatcher()
	if err != nil {
		return nil, err
	}
	rootDir = strings.Replace(path.Clean(rootDir), "\\", "/", -1) + "/"
	engine := &ViewEngine{
		viewDir: rootDir,
		viewExt: strings.ToLower(ext),
		watcher: w,
	}
	stat, err := os.Stat(engine.viewDir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("failed to open the view directory '%s'", engine.viewDir)
	}
	engine.watcher.AddHandler(&compileHandler{engine})
	engine.watcher.Start()
	engine.watcher.AddWatch(engine.viewDir, true)
	return engine, nil
}