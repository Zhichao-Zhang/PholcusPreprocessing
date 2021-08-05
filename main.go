package main

import (
	"PholcusPreprocessing/library/utils"
	_ "PholcusPreprocessing/boot"
	"PholcusPreprocessing/router"
	"github.com/gogf/gf/os/gcmd"
	"github.com/gogf/gf/os/glog"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)


func test()  {
	glog.Info("test ")
	session :=(* utils.DBDriver).NewSession(neo4j.SessionConfig{})
	defer session.Close()
}

func main() {
	// worker server 分离
	_ = gcmd.BindHandle("test", test)
	_= gcmd.BindHandle("server", server)
	_ = gcmd.AutoRun()
}

func server()  {

	s := router.Routing()
	s.Run()

}


