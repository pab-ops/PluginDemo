package router

import (
	"fmt"
	"github.com/pab-ops/EvopsPlugin/utils"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var goPath string

func init() {
	goPath = os.Getenv("GOPATH")
	if "" == goPath {
		panic("未设置环境变量$GOPATH")
	}
}

func NewManager() *manager {
	rm := new(manager)
	rm.routerMap = make(map[string]handlerModule)

	return rm
}

type handlerModule struct {
	hType    reflect.Type
	funcName string
}

type manager struct {
	routerMap map[string]handlerModule
}

func (rm *manager) ExecRouterModule(key string, args []interface{}) {
	if handler, ok := rm.routerMap[key]; ok {
		reflect.New(handler.hType).MethodByName(handler.funcName).Call([]reflect.Value{reflect.ValueOf(args)})
	} else {
		fmt.Println("404 不存在的channel [", key, "]")
	}
}

func (rm *manager) AutoRouter(module interface{}) {
	v := reflect.ValueOf(module)
	if reflect.Ptr != v.Kind() {
		panic("自动路由类型错误")
	}

	filePath := ""
	t := reflect.Indirect(v).Type()
	wgoPath := filepath.SplitList(goPath)
	for _, wg := range wgoPath {
		wg, _ = filepath.EvalSymlinks(filepath.Join(wg, "pkg", "mod", t.PkgPath()))
		if utils.FileExists(wg) {
			filePath = wg
			break
		}
	}

	fmt.Println(t.Name(), "添加注解路由", t.PkgPath(), filePath)
	rm.parseComment(filePath, t)

	return
}

func (rm *manager) parseComment(pkgPath string, t reflect.Type) {
	fileSet := token.NewFileSet()
	astPkgs, err := parser.ParseDir(fileSet, pkgPath, func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go")
	}, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for _, pkg := range astPkgs {
		for _, fl := range pkg.Files {
			for _, d := range fl.Decls {
				switch declare := d.(type) {
				case *ast.FuncDecl:
					if nil == declare.Recv {
						continue
					}
					exp, _ := declare.Recv.List[0].Type.(*ast.StarExpr)
					if fmt.Sprint(exp.X) != t.Name() {
						continue
					}
					// 解析注释
					comment := declare.Doc.Text()
					txt := ""
					for _, item := range strings.Split(strings.TrimSpace(strings.TrimLeft(comment, "//")), "\n") {
						if strings.HasPrefix(item, "@router") {
							txt = item
						}
					}
					if "" == txt {
						continue
					}
					strList := strings.SplitN(txt, " ", 2)
					fmt.Println(txt, strList)
					if len(strList) < 1 {
						panic("没有填写路由信息")
					}
					rtKey := strings.Trim(strList[1], " ")
					if v, ok := rm.routerMap[rtKey]; ok {
						panic(rtKey + " " + t.String() + "-" + declare.Name.String() + " 与 " + v.hType.String() + "-" + v.funcName + " 路由规则重复")
					}
					rm.routerMap[rtKey] = handlerModule{
						hType:    t,
						funcName: declare.Name.String(),
					}
				}
			}
		}
	}
}
