package plugin

import "github.com/pab-ops/PluginDemo/framwork/router"

var rList []interface{}

func initRouter() {
	r := router.NewManager()
	for _, m := range rList {
		r.AutoRouter(m)
	}
}

func AutoRouter(module interface{}) {
	rList = append(rList, module)
}
