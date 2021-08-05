package boot

import (
	"PholcusPreprocessing/library/utils"
	"github.com/gogf/gf/os/glog"
)

func init() {
	// Setup the mgm default config
	if err := utils.InitDriver(); err != nil {
		glog.Errorf("连接 Neo4J 数据库失败: %s", err)
	}
}