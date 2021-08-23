package utils

import (
	"encoding/json"
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/util/gconv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

func GetNodeAttrMap() (map[string]interface{}, error) {
	session := (*DBDriver).NewSession(neo4j.SessionConfig{})
	defer session.Close()

	allNodes, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		records, err := tx.Run("MATCH (n:Device) RETURN n", map[string]interface{}{})
		allEntry, err := records.Collect()
		if err != nil {
			return nil, err
		}
		return allEntry, nil
	})
	if err != nil {
		glog.Errorf("执行 neo4Xj 命令失败: %s", err)
	}
	result := map[string]interface{}{}
	for _, v := range allNodes.([]*neo4j.Record) {
		node := v.Values[0]
		temp := gconv.String(node)
		jsonS := gjson.New(temp)
		id := jsonS.GetString("Id")
		props := jsonS.GetMap("Props")
		if props["rawData"] == nil{
			// lldp Device
			glog.Info("LLDP")
			nodeAttr := map[string]interface{}{}
			nodeAttr["desc"] = props["desc"]
			nodeAttr["identifier"] = props["identifier"]
			nodeAttr["interfaceName"] = props["interfaceName"]
			nodeAttr["name"] = props["name"]
			nodeAttr["type"] = props["type"]
			nodeAttr["vendor"] = props["vendor"]
			result[id] =nodeAttr
		}else{
			// snmp Device
			glog.Info("SNMP")
			nodeAttr := map[string]interface{}{}
			rawData := map[string]interface{}{}
			err := json.Unmarshal([]byte(props["rawData"].(string)),&rawData )
			if err != nil{
				glog.Warning(err)
			}
			nodeAttr["identifier"] = props["identifier"]
			nodeAttr["canIPForwarding"] = props["canIPForwarding"]
			nodeAttr["isLayer2Device"] = props["isLayer2Device"]
			nodeAttr["name"] = props["name"]
			nodeAttr["ArpTable"] = rawData["ArpTable"]
			nodeAttr["FdpTable"] = rawData["FdpTable"]
			nodeAttr["IPRouteTable"] = rawData["IPRouteTable"]
			nodeAttr["InterfaceTable"] = rawData["InterfaceTable"]
			nodeAttr["LLdpTable"] = rawData["LLdpTable"]
			nodeAttr["MetaData"] = rawData["MetaData"]
			nodeAttr["PhysMediaTable"] = rawData["PhysMediaTable"]
			nodeAttr["StpTable"] = rawData["StpTable"]
			result[id] = nodeAttr

		}


	}

	return result, nil


}
