package viewEngine

import (
	"html/template"
	"sync"
	"path/filepath"
	"os"
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"
	"io"
	"errors"
	"bytes"
	"github.com/Simbory/mego/watcher"
)

type Engine struct {
	viewDir  string
	viewExt  string
	funcMap  template.FuncMap
	locker   sync.RWMutex
	viewMap  map[string]*engineTemp
	compiled bool

	watcher *watcher.FileWatcher
}

func (engine *Engine) AddFunc(name string, viewFunc interface{}) {
	if len(name) < 1 || viewFunc == nil {
		return
	}
	if engine.funcMap == nil {
		engine.funcMap = make(template.FuncMap)
	}
	if f,ok := engine.funcMap[name]; !ok || f==nil {
		engine.funcMap[name] = viewFunc
	}
}

func (engine *Engine) addView(name string, v *engineTemp) {
	if len(name) == 0 || v == nil {
		return
	}
	if engine.viewMap == nil {
		engine.viewMap = map[string]*engineTemp{}
	}
	engine.viewMap[name] = v
}

func (engine *Engine) getView(name string) *engineTemp {
	if engine.viewMap == nil {
		return nil
	}
	return engine.viewMap[name]
}

func (engine *Engine) getTemplateDeep(file, parent string, t *template.Template) (*template.Template, [][]string, error){
	var fileAbsPath string
	if strings.HasPrefix(file, "../") {
		fileAbsPath = filepath.Join(engine.viewDir, filepath.Dir(parent), file)
	} else {
		fileAbsPath = filepath.Join(engine.viewDir, file)
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
	reg := regexp.MustCompile("[{]{2}[ \t]*template[ \t]+\"([^\"]+)\"")
	allSub := reg.FindAllStringSubmatch(string(data), -1)
	for _, m := range allSub {
		if len(m) == 2 {
			name := m[1]
			if !strings.HasSuffix(name, engine.viewExt) {
				name = name + engine.viewExt
			}
			look := t.Lookup(name)
			if look != nil {
				continue
			}
			t, _, err = engine.getTemplateDeep(name, file, t)
			if err != nil {
				return nil, [][]string{}, err
			}
		}
	}
	return t, allSub, nil
}

func (engine *Engine) getTemplateLoop(temp *template.Template, subMods [][]string, others ...string) (t *template.Template, err error) {
	t = temp
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
					t, subMods1, err = engine.getTemplateDeep(otherFile, "", t)
					if err != nil {
						return nil, err
					} else if subMods1 != nil && len(subMods1) > 0 {
						t, err = engine.getTemplateLoop(t, subMods1, others...)
					}
					break
				}
			}
			//second check define
			for _, otherFile := range others {
				fileAbsPath := filepath.Join(engine.viewDir, otherFile)
				data, err := ioutil.ReadFile(fileAbsPath)
				if err != nil {
					continue
				}
				reg := regexp.MustCompile("{{" + "[ ]*define[ ]+\"([^\"]+)\"")
				allSub := reg.FindAllStringSubmatch(string(data), -1)
				for _, sub := range allSub {
					if len(sub) == 2 && sub[1] == m[1] {
						var subMods1 [][]string
						t, subMods1, err = engine.getTemplateDeep(otherFile, "", t)
						if err != nil {
							return nil, err
						} else if subMods1 != nil && len(subMods1) > 0 {
							t, err = engine.getTemplateLoop(t, subMods1, others...)
						}
						break
					}
				}
			}
		}
	}
	return
}

func (engine *Engine) getTemplate(file string, others ...string) (t *template.Template, err error) {
	t = template.New(file)
	if engine.funcMap != nil {
		t.Funcs(engine.funcMap)
	}
	var subMods [][]string
	t, subMods, err = engine.getTemplateDeep(file, "", t)
	if err != nil {
		return nil, err
	}
	t, err = engine.getTemplateLoop(t, subMods, others...)
	if err != nil {
		return nil, err
	}
	return
}

func (engine *Engine) compile() error {
	engine.locker.Lock()
	defer engine.locker.Unlock()
	if engine.compiled {
		return nil
	}
	engine.AddFunc("include", engine.includeView)
	engine.viewDir = engine.viewDir
	if _, err := os.Stat(engine.viewDir); err != nil {
		if os.IsNotExist(err) {
			return err
		}
		return fmt.Errorf("Cannot open view path: %s", engine.viewDir)
	}
	vf := &engineFile{
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
			t, err := engine.getTemplate(file, v...)
			v := &engineTemp{tpl: t, err: err}
			engine.addView(file, v)
		}
	}
	engine.compiled = true
	return nil
}

func (engine *Engine) includeView(viewName string, data interface{}) template.HTML {
	buf := &bytes.Buffer{}
	err := engine.Render(buf, viewName, data)
	if err != nil {
		panic(err)
	}
	return template.HTML(buf.Bytes())
}

func (engine *Engine) Clear() {
	engine.locker.Lock()
	defer engine.locker.Unlock()

	engine.compiled = false
	engine.viewMap = nil
}

func (engine *Engine) Render(writer io.Writer, viewPath string, viewData interface{}) error {
	if len(viewPath) < 1 {
		return errors.New("The parameter 'viewPath' cannot be empty")
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
	tpl := engine.getView(viewPath)
	if tpl == nil || tpl.tpl == nil {
		return fmt.Errorf("The viewPath '%s' canot be found", viewPath)
	}
	if tpl.err != nil {
		return tpl.err
	}
	err := tpl.tpl.Execute(writer, viewData)
	if err != nil {
		return err
	}
	return nil
}

func NewEngine(rootDir, ext string) (*Engine,error) {
	if len(ext) == 0 {
		ext = ".html"
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	w,err := watcher.NewWatcher()
	if err != nil {
		return nil, err
	}
	engine := &Engine{
		viewDir: rootDir,
		viewExt: strings.ToLower(ext),
		watcher: w,
	}
	stat,err := os.Stat(engine.viewDir)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, fmt.Errorf("The view folder '%s' does not exist.", engine.viewDir)
	}
	engine.watcher.AddHandler(&compileHandler{engine})
	engine.watcher.Start()
	engine.watcher.AddWatch(engine.viewDir, true)
	return engine, nil
}