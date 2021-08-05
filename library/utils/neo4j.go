package utils

import (

	"github.com/gogf/gf/frame/g"
	"github.com/gogf/gf/os/glog"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

var DBDriver *neo4j.Driver

func InitDriver() error {
	dbUri := g.Cfg().GetString("neo4j.link")
	username := g.Cfg().GetString("neo4j.username")
	password := g.Cfg().GetString("neo4j.password")

	driver, err := neo4j.NewDriver(dbUri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		return err
	}
	DBDriver = &driver
	glog.Info("成功连接neo4j")
	return nil
}

//func RunCommand(command string, itemData map[string]interface{}) (error){
//	if DBDriver == nil {
//		if err := InitDriver(); err != nil {
//			glog.Error("连接 Neo4J 数据库失败:", err)
//			return err
//		}
//	}
//	session := (*DBDriver).NewSession(neo4j.SessionConfig{})
//	defer session.Close()
//	_, err := session.WriteTransaction(func(tx neo4j.Transaction) (interface{}, error) {
//		_, err := tx.Run(command, itemData)
//		if err != nil {
//			return err, nil
//		}
//		return nil, nil
//	})
//	if err != nil {
//		glog.Errorf("执行 neo4j 命令[%s]失败: %s", command, err)
//		return err
//	}
//	return nil
//}
//
