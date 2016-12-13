package mego

import (
	"errors"
	"fmt"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

const (
	static pathType = iota
	root
	param
	catchAll

	paramBegin    = '<'
	paramBeginStr = "<"
	paramEnd      = '>'
	paramEndStr   = ">"
	pathInfo      = "*pathInfo"
)

var (
	reg1    = regexp.MustCompile("^[a-zA-Z][\\w]*$")
	reg2    = regexp.MustCompile("^[a-zA-Z][\\w]*\\(.+\\)$")
	reg3    = regexp.MustCompile("^[0-9]+$")
	reg4    = regexp.MustCompile("^[0-9]+(~)+[0-9]+$")
	wordReg = regexp.MustCompile("^[\\w]+")
)

// HandlerFunc the route handler function
type ReqHandler func(ctx *Context) interface{}

// ValidationFunc define the route check function
type RouteFunc func(urlPath string, opt RouteOpt) string

type pathType uint8

// RouteOption the route option struct
type RouteOpt interface {
	Validation() string
	HasDefaultValue() bool
	DefaultValue() string
	Setting() string
	MaxLength() int
	MinLength() int
}

type routeOpt struct {
	validation      string
	hasDefaultValue bool
	defaultValue    string
	setting         string
	maxLength       int
	minLength       int
}

func (opt *routeOpt) Validation() string {
	return opt.validation
}

func (opt *routeOpt) HasDefaultValue() bool {
	return opt.hasDefaultValue
}

func (opt *routeOpt) DefaultValue() string {
	return opt.defaultValue
}

func (opt *routeOpt) Setting() string {
	return opt.setting
}

func (opt *routeOpt) MaxLength() int {
	return opt.maxLength
}

func (opt *routeOpt) MinLength() int {
	return opt.minLength
}

type routeNode struct {
	NodeType   pathType
	CurDepth   uint16
	MaxDepth   uint16
	Path       string
	PathSplits []string
	Params     map[string]RouteOpt
	handlers   map[string]ReqHandler
	Children   []*routeNode
}

func (node *routeNode) isLeaf() bool {
	if node.NodeType == root {
		return false
	}
	return node.hasChildren() == false
}

func (node *routeNode) hasChildren() bool {
	return len(node.Children) > 0
}

func (node *routeNode) findChild(path string) *routeNode {
	if !node.hasChildren() {
		return nil
	}
	for _, child := range node.Children {
		if child.Path == path {
			return child
		}
	}
	return nil
}

func (node *routeNode) addChild(childNode *routeNode) error {
	if childNode == nil {
		return errors.New("'childNode' parameter cannot be nil")
	}
	var existChild = node.findChild(childNode.Path)
	if existChild == nil {
		node.Children = append(node.Children, childNode)
		return nil
	}
	if childNode.MaxDepth > existChild.MaxDepth {
		existChild.MaxDepth = childNode.MaxDepth
	}
	if childNode.isLeaf() {
		if existChild.handlers == nil {
			existChild.handlers = make(map[string]ReqHandler)
		}
		// merge handlers
		for hMethod, hFunc := range childNode.handlers {
			if _, ok := existChild.handlers[hMethod]; ok {
				return fmt.Errorf("Duplicate handler for HttpMethod %s in route tree. Path: %s, Depth: %d",
					hMethod,
					existChild.Path,
					existChild.CurDepth)
			}
			existChild.handlers[hMethod] = hFunc
		}
	} else {
		for _, child := range childNode.Children {
			err := existChild.addChild(child)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (node *routeNode) isParamPath(path string) bool {
	return strings.HasPrefix(path, paramBeginStr) && strings.HasSuffix(path, paramEndStr)
}

func (node *routeNode) detectDefault() (bool, map[string]ReqHandler, map[string]string) {
	if !node.hasChildren() {
		return false, nil, nil
	}
	for _, child := range node.Children {
		if child.NodeType != param || len(child.PathSplits) != 1 || !node.isParamPath(child.PathSplits[0]) {
			continue
		}
		paramName := ""
		var opt RouteOpt
		for name, o := range child.Params {
			paramName = name
			opt = o
			break
		}
		if !opt.HasDefaultValue() {
			continue
		}
		if child.handlers != nil {
			return true, child.handlers, map[string]string{paramName: opt.DefaultValue()}
		}
		found, ctrl, routeMap := child.detectDefault()
		if found {
			routeMap[paramName] = opt.DefaultValue()
			return true, ctrl, routeMap
		}
	}
	return false, nil, nil
}

func newRouteNode(routePath, method string, handler ReqHandler) (*routeNode, error) {
	err := checkRoutePath(routePath)
	if err != nil {
		return nil, err
	}
	splitPaths, err := splitURLPath(routePath)
	if err != nil {
		return nil, err
	}
	var length = uint16(len(splitPaths))
	if length == 0 {
		return nil, nil
	}
	if detectNodeType(splitPaths[length-1]) == catchAll {
		length = 255
	}
	var result *routeNode
	var current *routeNode
	for i, p := range splitPaths {
		var child = &routeNode{
			NodeType: detectNodeType(p),
			CurDepth: uint16(i + 1),
			MaxDepth: uint16(length - uint16(i)),
			Path:     p,
		}
		if child.NodeType == param {
			paramPath, params, err := analyzeParamOption(child.Path)
			if err != nil {
				return nil, err
			}
			child.PathSplits = paramPath
			child.Params = params
		}
		if result == nil {
			result = child
			current = result
		} else {
			current.Children = []*routeNode{child}
			current = current.Children[0]
		}
	}
	current.handlers = map[string]ReqHandler{
		strings.ToUpper(method): handler,
	}
	current = result
	for {
		if current == nil {
			break
		}
		if strings.Contains(current.Path, "*") && current.NodeType != catchAll {
			return nil, errors.New("Invalid URL route parameter '" + current.Path + "'")
		}
		if current.NodeType == catchAll && len(current.Children) > 0 {
			return nil, errors.New("Invalid route'" + routePath + ". " +
				"The '*pathInfo' parameter should be at the end of the route. " +
				"For example: '/shell/*pathInfo'.")
		}
		if len(current.Children) > 0 {
			current = current.Children[0]
		} else {
			current = nil
		}
	}
	return result, nil
}

type routeTree struct {
	routeNode
	funcMap   map[string]RouteFunc
	MatchCase bool
}

func (tree *routeTree) addFunc(name string, fun RouteFunc) error {
	if len(name) == 0 {
		return errors.New("The parameter 'name' cannot be empty")
	}
	if fun == nil {
		return errors.New("The parameter 'fun' cannot be nil")
	}
	if _, ok := tree.funcMap[name]; ok {
		return fmt.Errorf("The '%s' function is already exist.", name)
	}
	tree.funcMap[name] = fun
	return nil
}

func (tree *routeTree) lookupDepth(indexNode *routeNode, pathLength uint16, urlParts []string, endWithSlash bool) (found bool, handler map[string]ReqHandler, routeMap map[string]string) {
	found = false
	handler = nil
	routeMap = nil
	if indexNode.MaxDepth+indexNode.CurDepth <= pathLength || indexNode.NodeType == root {
		return
	}
	var routeData = make(map[string]string)
	var curPath = urlParts[indexNode.CurDepth-1]
	if indexNode.NodeType == catchAll {
		// deal with *pathInfo
		var path string
		for _, part := range urlParts[indexNode.CurDepth-1:] {
			path = path + "/" + part
		}
		if endWithSlash {
			path = path + "/"
		}
		routeData["pathInfo"] = strings.TrimLeft(path, "/")
		found = true
		handler = indexNode.handlers
		routeMap = routeData
		return
	} else if indexNode.NodeType == static {
		// deal with static path
		str1 := indexNode.Path
		str2 := curPath
		if !tree.MatchCase {
			str1 = strings.ToLower(str1)
			str2 = strings.ToLower(str2)
		}
		if str1 != str2 {
			return
		}
	} else if indexNode.NodeType == param {
		// deal with dynamic path
		var dynPathSplits []string // the dynamic route paths that to be check
		var str1 string
		var str2 string

		var checkFunc = func(index int) {
			if len(dynPathSplits) > 0 {
				validationStr := curPath[0:index]
				for _, dynPath := range dynPathSplits {
					paramName := dynPath[1 : len(dynPath)-1]
					opt := indexNode.Params[paramName]
					if len(validationStr) == 0 {
						if !opt.HasDefaultValue() {
							return
						}
						routeData[paramName] = opt.DefaultValue()
					} else {
						validateFunc := tree.funcMap[opt.Validation()]
						if validateFunc == nil {
							return
						}
						data := validateFunc(validationStr, opt)
						if len(data) == 0 || len(data) > len(validationStr) {
							return
						}
						routeData[paramName] = data
						validationStr = validationStr[len(data):]
					}
				}
				if len(validationStr) > 0 {
					return
				}
			}
			dynPathSplits = nil
			curPath = curPath[index+len(str1):]
		}

		for _, p := range indexNode.PathSplits {
			if tree.isParamPath(p) {
				dynPathSplits = append(dynPathSplits, p)
				continue
			}
			str1 = p
			str2 = curPath
			if !tree.MatchCase {
				str1 = strings.ToLower(str1)
				str2 = strings.ToLower(str2)
			}
			index := strings.Index(str2, str1)
			if index == -1 {
				return
			}
			checkFunc(index)
		}
		str1 = ""
		checkFunc(len(curPath))
		if len(curPath) != 0 {
			return
		}
	} else {
		return
	}
	if indexNode.CurDepth == pathLength {
		handler = indexNode.handlers
		routeMap = routeData
		// detect default value
		if handler == nil {
			f, c, rm := indexNode.detectDefault()
			if f {
				found = true
				handler = c
				if rm != nil {
					for key, value := range rm {
						routeMap[key] = value
					}
				}
			} else {
				found = false
				routeMap = nil
				handler = nil
			}
		} else {
			found = true
		}
		return
	}
	// check the last url parts
	if !indexNode.hasChildren() {
		return
	}
	for _, child := range indexNode.Children {
		ok, result, rd := tree.lookupDepth(child, pathLength, urlParts, endWithSlash)
		if ok {
			if rd != nil && len(rd) > 0 {
				for key, value := range rd {
					routeData[key] = value
				}
			}
			found = true
			handler = result
			routeMap = routeData
			return
		}
	}
	return
}

func (tree *routeTree) lookup(urlPath string) (map[string]ReqHandler, map[string]string, error) {
	if urlPath == "/" {
		handler := tree.handlers
		if handler == nil {
			f, c, r := tree.detectDefault()
			if f {
				return c, r, nil
			}
			return nil, nil, nil
		}
		return tree.handlers, nil, nil
	}
	urlParts, err := splitURLPath(urlPath)
	if err != nil {
		return nil, nil, err
	}
	var pathLength = uint16(len(urlParts))
	if pathLength == 0 || len(tree.Children) == 0 {
		return nil, nil, nil
	}
	var endWithSlash = strings.HasSuffix(urlPath, "/")
	for _, child := range tree.Children {
		ok, result, rd := tree.lookupDepth(child, pathLength, urlParts, endWithSlash)
		if ok {
			return result, rd, nil
		}
	}
	return nil, nil, nil
}

func (tree *routeTree) addRoute(method, routePath string, handler ReqHandler) {
	if len(routePath) == 0 {
		panic(errors.New("'routePath' param cannot be empty"))
	}
	if handler == nil {
		panic(errors.New("'handler' param cannot be nil"))
	}
	if routePath == "/" {
		if tree.handlers == nil {
			tree.handlers = map[string]ReqHandler{
				strings.ToUpper(method): handler,
			}
		} else {
			tree.handlers[strings.ToUpper(method)] = handler
		}
		return
	}
	branch, err := newRouteNode(routePath, method, handler)
	if err != nil {
		panic(err)
	}
	if err = tree.addChild(branch); err != nil {
		panic(err)
	}
}

func newRouteTree() *routeTree {
	var node = &routeTree{
		funcMap: map[string]RouteFunc{
			"int":  num,
			"any":  any,
			"word": word,
			"enum": enum,
		},
	}
	node.NodeType = root
	node.CurDepth = 0
	node.MaxDepth = 0
	node.Path = "/"
	node.handlers = map[string]ReqHandler{}
	node.MatchCase = runtime.GOOS != "windows"
	return node
}

func isA2Z(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

func isNumber(c byte) bool {
	return (c >= '0' && c <= '9')
}

func splitURLPath(urlPath string) ([]string, error) {
	if len(urlPath) == 0 {
		return nil, errors.New("The URL path is empty")
	}
	p := strings.Trim(urlPath, "/")
	splits := strings.Split(p, "/")
	var result []string
	for _, s := range splits {
		if len(s) == 0 || s == "." {
			continue
		}
		if s == ".." {
			return nil, errors.New("Invalid URL path. The URL path cannot contains '..'")
		}
		result = append(result, s)
	}
	return result, nil
}

func detectNodeType(p string) pathType {
	if p == "/" {
		return root
	}
	if strings.Contains(p, string([]byte{paramBegin})) || strings.Contains(p, string([]byte{paramEnd})) {
		return param
	}
	if p == pathInfo {
		return catchAll
	}
	return static
}

func checkRoutePath(path string) error {
	var routeParams []string
	var paramChars []byte
	var inParamChar = false

	for i := 0; i < len(path); i++ {
		// param begin
		if path[i] == paramBegin {
			if len(paramChars) == 0 {
				inParamChar = true
				continue
			} else {
				return fmt.Errorf("the route param has no closing character '>': %d", i)
			}
		}
		// param end
		if path[i] == paramEnd {
			// check and ensure current route param is not empty
			if len(paramChars) == 0 {
				return fmt.Errorf("Invalid route parameter '<>' or the route parameter has no begining tag '<': %d", i)
			}
			var curParam = strings.Split(string(paramChars), ":")[0]
			for _, tmp := range routeParams {
				if tmp == curParam {
					return fmt.Errorf("Duplicate route param '%s': %d", curParam, i)
				}
			}
			routeParams = append(routeParams, curParam)
			paramChars = make([]byte, 0)
			inParamChar = false
			continue
		}
		if inParamChar {
			if len(paramChars) == 0 {
				if isA2Z(path[i]) {
					paramChars = append(paramChars, path[i])
				} else {
					return fmt.Errorf("Invalid character '%c' at the beginin of the route param: %d", path[i], i)
				}
			} else {
				paramChars = append(paramChars, path[i])
			}
		}
	}
	if len(routeParams) > 255 {
		return errors.New("Too many route params: the maximum number of the route param is 255")
	}
	return nil
}

func splitRouteParam(path string) []string {
	var splits []string
	var byteQueue []byte
	for _, char := range []byte(path) {
		if char == paramEnd {
			byteQueue = append(byteQueue, char)
			if len(byteQueue) > 0 {
				splits = append(splits, string(byteQueue))
				byteQueue = nil
			}
		} else {
			if char == paramBegin && len(byteQueue) > 0 {
				splits = append(splits, string(byteQueue))
				byteQueue = nil
			}
			byteQueue = append(byteQueue, char)
		}
	}
	if len(byteQueue) > 0 {
		splits = append(splits, string(byteQueue))
	}
	return splits
}

func checkParamName(name string) bool {
	return reg1.Match([]byte(name))
}

func checkParamOption(optionStr string) bool {
	return reg2.Match([]byte(optionStr))
}

func checkNumber(opt string) bool {
	return reg3.Match([]byte(opt))
}

func checkNumberRange(optStr string) bool {
	return reg4.Match([]byte(optStr))
}

func analyzeParamOption(path string) ([]string, map[string]RouteOpt, error) {
	splitParams := splitRouteParam(path)
	optionMap := make(map[string]RouteOpt)
	var paramPath []string
	for _, sp := range splitParams {
		if strings.HasSuffix(sp, paramEndStr) && strings.HasPrefix(sp, paramBeginStr) {
			paramStr := strings.Trim(sp, paramBeginStr+paramEndStr)
			splits := strings.Split(paramStr, ":")
			// paramName: the name of the route param (with default value), like 'name', 'name=Steve Jobs' or 'name='
			paramName := splits[0]
			// paramOptionStr: the route param option
			paramOptionStr := ""
			if len(splits) == 1 {
				paramOptionStr = "any"
			}
			if len(splits) == 2 {
				paramOptionStr = splits[1]
				if len(paramOptionStr) == 0 {
					paramOptionStr = "any"
				}
			} else if len(splits) > 2 {
				return nil, nil, errors.New("Invalid route parameter setting: " + sp)
			}
			opt := &routeOpt{}
			var eqIndex = strings.Index(paramName, "=")
			if eqIndex > 0 {
				defaultValue := paramName[eqIndex+1:]
				paramName = paramName[0:eqIndex]
				opt.defaultValue = defaultValue
				opt.hasDefaultValue = true
			} else if !checkParamName(paramName) {
				return nil, nil, errors.New("Invalid route parameter name '" + paramName + "': " + sp)
			} else {
				opt.hasDefaultValue = false
			}
			if checkParamName(paramOptionStr) {
				opt.validation = paramOptionStr
				opt.maxLength = 255
				opt.minLength = 1
			} else if checkParamOption(paramOptionStr) {
				optSplits := strings.Split(paramOptionStr, "(")
				if len(optSplits) != 2 {
					return nil, nil, errors.New("Invalid route parameter setting: " + sp)
				}
				opt.validation = optSplits[0]
				var setting = strings.TrimRight(optSplits[1], ")")
				if strings.Contains(setting, ")") {
					return nil, nil, errors.New("Invalid route parameter setting: " + sp)
				}
				if checkNumber(setting) {
					i, err := strconv.ParseInt(setting, 10, 0)
					if err != nil {
						return nil, nil, err
					}
					opt.maxLength = int(i)
					opt.minLength = int(i)
				} else if checkNumberRange(setting) {
					numbers := strings.Split(setting, "~")
					min, err := strconv.ParseInt(numbers[0], 10, 0)
					if err != nil {
						return nil, nil, err
					}
					max, err := strconv.ParseInt(numbers[1], 10, 0)
					if err != nil {
						return nil, nil, err
					}
					if min < max {
						opt.minLength = int(min)
						opt.maxLength = int(max)
					} else {
						opt.minLength = int(max)
						opt.maxLength = int(min)
					}
				} else {
					opt.maxLength = 255
					opt.minLength = 1
					opt.setting = setting
				}
			}
			optionMap[paramName] = opt
			paramPath = append(paramPath, paramBeginStr+paramName+paramEndStr)
		} else {
			paramPath = append(paramPath, sp)
		}
	}
	return paramPath, optionMap, nil
}

func any(urlPath string, opt RouteOpt) string {
	var length = len(urlPath)
	if length >= opt.MinLength() && length <= opt.MaxLength() {
		return urlPath
	}
	return ""
}

func word(urlPath string, opt RouteOpt) string {
	bytes := wordReg.Find([]byte(urlPath))
	if len(bytes) >= opt.MinLength() && len(bytes) <= opt.MaxLength() {
		return string(bytes)
	}
	return ""
}

func num(urlPath string, opt RouteOpt) string {
	var numBytes []byte
	for _, char := range []byte(urlPath) {
		if isNumber(char) {
			numBytes = append(numBytes, char)
		} else {
			break
		}
		if len(numBytes) >= opt.MaxLength() {
			break
		}
	}
	if len(numBytes) >= opt.MinLength() {
		return string(numBytes)
	}
	return ""
}

func enum(urlPath string, opt RouteOpt) string {
	if len(opt.Setting()) == 0 {
		return ""
	}
	var splits = strings.Split(opt.Setting(), "|")
	for _, value := range splits {
		if strings.HasPrefix(urlPath, value) {
			return value
		}
	}
	return ""
}
