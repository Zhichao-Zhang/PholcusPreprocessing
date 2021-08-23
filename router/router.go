package router

import (
	"PholcusPreprocessing/library/utils"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/net/ghttp"
	"github.com/gogf/gf/os/glog"
)

func Routing() *ghttp.Server{
	s := g.Server()
	s.BindHandler("GET:/v1/views", func(r *ghttp.Request) {
		res , err:=utils.GetViews()
		if err != nil{
			glog.Warning(err)
			response := Response{"101", "Failure",map[string]interface{}{"error": err}}
			r.Response.WriteStatus(500, response)
		}
		response := Response{"100", "SUCCESS",res}
		r.Response.CORS(ghttp.CORSOptions{
			AllowOrigin:      "*",
			AllowMethods:     "POST, GET, OPTIONS, PUT, DELETE,UPDATE,",
		})
		r.Response.WriteStatus(200, response)
		return

	})


	s.BindHandler("GET:/v1/nodeAttrMap", func(r *ghttp.Request) {
		res , err:=utils.GetNodeAttrMap()
		if err != nil{
			glog.Warning(err)
			response := Response{"101", "Failure",map[string]interface{}{"error": err}}
			r.Response.WriteStatus(500, response)
		}
		response := Response{"100", "SUCCESS",res}
		r.Response.CORS(ghttp.CORSOptions{
			AllowOrigin:      "*",
			AllowMethods:     "POST, GET, OPTIONS, PUT, DELETE,UPDATE,",
		})
		r.Response.WriteStatus(200, response)
		return

	})
	s.SetPort(8199)
	return s
}





type Response struct {
	Code string
	Msg string
	Data map[string]interface{}
}
