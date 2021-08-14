package utils

import (
	"github.com/gogf/gf/encoding/gjson"
	"github.com/gogf/gf/os/glog"
	"github.com/gogf/gf/util/gconv"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"strings"
)

// 获取有效接口集合 validIntSet，后面的allEdge和allNode需要筛选
func GetValidInt(session neo4j.Session)  (map[string]bool,  error) {
	validInt, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		records, err := tx.Run("MATCH (n1:Device)-[r]-(n2:Interface) RETURN n2", map[string]interface{}{})
		allEntry, err := records.Collect()
		if err != nil {
			return nil, err
		}
		return allEntry, nil
	})
	if err != nil {
		glog.Errorf("执行 neo4Xj 命令失败: %s", err)
	}
	validIntSet := make(map[string]bool)

	for _, v := range validInt.([]*neo4j.Record) {
		r := v.Values[0]
		temp := gconv.String(r)
		jsonS := gjson.New(temp)
		id := jsonS.GetString("Id")
		if _, ok := validIntSet[id]; ok || len(id) == 0{
			continue
		}
		validIntSet[id] = true
	}

	return validIntSet, err
}

func FixRepeat(input map[string]interface{})  map[string]interface{}{
	nodeSet := make(map[string]bool)
	edgeSet := make(map[string]bool)
	resNode :=  []map[string]string{}
	resEdge :=  []map[string]string{}

	nodes := gconv.Interfaces(input["nodes"])
	edges := gconv.Interfaces(input["edges"])

	for _, n := range nodes{

		if _,ok := nodeSet[n.(map[string]string)["id"]]; ok{
			glog.Warning("repeated node id" , n.(map[string]string)["id"])
			continue
		}else {
			resNode = append(resNode, n.(map[string]string))
			nodeSet[n.(map[string]string)["id"]] = true
		}

	}

	for _, e := range edges{

		if _,ok := edgeSet[e.(map[string]string)["id"]]; ok{
			glog.Warning("repeated edge id" , e.(map[string]string)["id"])
			continue
		}else {
			resEdge = append(resEdge, e.(map[string]string))
			edgeSet[e.(map[string]string)["id"]] = true
		}

	}
	res := make(map[string]interface{})
	res["nodes"] = nodes
	res["edges"] = edges
	glog.Info("arr vs set")
	glog.Info(len(nodes), len(nodeSet), len(edges), len(edgeSet))
	return res





}


func GetViews() (map[string]interface{}, error) {
	session := (*DBDriver).NewSession(neo4j.SessionConfig{})
	defer session.Close()


	// 读入全部数据：结点 and 边
	allEntry, err := session.ReadTransaction(func(tx neo4j.Transaction) (interface{}, error) {
		records, err := tx.Run("MATCH (n1)-[r]-(n2) RETURN n1,r,n2", map[string]interface{}{})
		allEntry, err := records.Collect()
		if err != nil {
			return nil, err
		}
		return allEntry, nil
	})
	if err != nil {
		glog.Errorf("执行 neo4Xj 命令失败: %s", err)
	}
	// 负责去重的set
	nodesSet := map[string]bool{}
	edgesSet := map[string]bool{}
	allIntSet := map[string]bool{} // 用于处理validInt过滤(终端过滤）

	// 存储 不重复nodes和edges的数组
	allNodes := []map[string]string{}
	allEdges := []map[string]string{}


	for _, v := range allEntry.([]*neo4j.Record) {

		// 读入neo4j Record
		// 处理node
		n1 := v.Values[0]
		n2 := v.Values[2]

		for _, n := range []interface{}{n1, n2}{

			node, err := NodeHandler(n, nodesSet)
			if err != nil{
				glog.Warning(err)
			}
			if node != nil{
				allNodes = append(allNodes, node)
			}

			if node["category"] == "Interface"{
				allIntSet[node["id"]] = true
			}

		}

		// 处理edge
		r := v.Values[1]
		edge, err := EdgeHandler(r, edgesSet)
		if err != nil{
			glog.Warning(err)
		}

		if edge != nil{
			allEdges = append(allEdges, edge)
		}

	}


	glog.Info("获取allNodes, allEdges + set计数 完成，计数为：")
	glog.Info(len(allNodes), len(nodesSet))
	glog.Info(len(allEdges), len(edgesSet))

	// 如果需要过滤掉终端设备的话

	validIntSet, err := GetValidInt(session)
	if err != nil{
		glog.Warning(err)
	}
	glog.Info(len(validIntSet))
	// 获取validIntSet后，找到补集，inValidIntSet好用
	invalidIntSet := make(map[string]bool)

	for key, _ := range allIntSet{
		if _,ok := validIntSet[key]; !ok{
			invalidIntSet[key] = true
		}
	}

	//glog.Info(len(invalidIntSet))

	// 混合视图
	mixViewNodes := []map[string]string{}
	mixViewEdges := []map[string]string{}

/*	for _, node := range allNodes{
		if node["category"] == "Interface"{
			if _, ok := invalidIntSet[node["id"]]; ok{
				// invalid node, 跳过
				continue
			}
}

		mixViewNodes = append(mixViewNodes, node)
	}

	for _, edge := range allEdges{
		if _, ok := invalidIntSet[edge["source"]]; ok{
			// invalid node, 跳过
			continue
		}
		if _, ok := invalidIntSet[edge["target"]]; ok{
			// invalid node, 跳过
			continue
		}

		mixViewEdges = append(mixViewEdges, edge)
	}
*/
	mixViewNodes = allNodes
	mixViewEdges = allEdges

	glog.Info("mixView 处理完成， 计数为：")
	glog.Info(len(mixViewNodes), len(mixViewEdges))


	
	mixView := make(map[string]interface{})
	mixView["nodes"] = mixViewNodes
	mixView["edges"] = mixViewEdges
//	glog.Info(mixView)

	newMixView := FixRepeat(mixView)




	// 逻辑视图

	NotL3Device := make(map[string]bool)
	seenNodes :=  make(map[string]bool)
	LogicViewNodes := []interface{}{}
	LogicViewEdges := []interface{}{}

	for _, node := range mixViewNodes{
		if node["category"] == "Device"{
			if node["deviceType"] == "L3Switch" || node["deviceType"] == "Router"{
				seenNodes[node["id"]] = true
			}else{
				NotL3Device[node["id"]] = true
			}
		}

	}

	update := true

	for update{
		update = false
		//glog.Info(len(seenNodes))
		for _, edge := range mixViewEdges{
			if _, ok:= seenNodes[edge["source"]]; ok {
				if _, ok := seenNodes[edge["target"]]; !ok{
					if _, ok := NotL3Device[edge["target"]]; !ok{
						seenNodes[edge["target"]] = true
						update = true
					}
				}
			}
		}
	}
	//glog.Info(len(seenNodes))
	logicNodeSet := make(map[string]bool)
	logicEdgeSet := make(map[string]bool)

	for _, node := range mixViewNodes{
		if _, ok := seenNodes[node["id"]]; ok{
			if _, ok := NotL3Device[node["id"]]; !ok{
				if _, ok := logicNodeSet[node["id"]]; ok{
					continue
				}
				LogicViewNodes = append(LogicViewNodes, node)
				logicNodeSet[node["id"]] = true
			}
		}
	}

	for _, edge := range mixViewEdges{
		if _, ok := seenNodes[edge["source"]]; ok{
			if _, ok := seenNodes[edge["target"]]; ok{
				if _, ok := logicEdgeSet[edge["id"]]; ok{
					continue
				}
				LogicViewEdges = append(LogicViewEdges, edge)
				logicEdgeSet[edge["id"]] = true
			}
		}
	}

	glog.Info("逻辑视图完成，计数为：")
	glog.Info(len(LogicViewNodes), len(LogicViewEdges))
	logicView := make(map[string]interface{})
	logicView["nodes"] = LogicViewNodes
	logicView["edges"] = LogicViewEdges
	//glog.Info(logicView)


	newLogicView := FixRepeat(logicView)

	// 物理视图
	NetworkSet := make(map[string]bool)
	PhysViewNodes := []interface{}{}
	PhysViewEdges := []interface{}{}



	for _, node := range mixViewNodes{
		if node["category"] == "Interface"{
			if _,ok := invalidIntSet[node["id"]]; ok{
				continue
			}
		}


		if node["category"] == "Network"{
			NetworkSet[node["id"]] = true
			continue
		}
		PhysViewNodes = append(PhysViewNodes, node)
	}

	for _, edge := range mixViewEdges{
		
		if _,ok := invalidIntSet[edge["source"]]; ok{
			continue
		}

		if _,ok := invalidIntSet[edge["target"]]; ok{
			continue
		}

		if _, ok := NetworkSet[edge["source"]]; ok{
			continue
		}

		if _, ok := NetworkSet[edge["target"]]; ok{
			continue
		}

		PhysViewEdges = append(PhysViewEdges, edge)
	}

	glog.Info("物理视图完成，计数为：")
	glog.Info(len(PhysViewNodes), len(PhysViewEdges))
	physView := make(map[string]interface{})
	physView["nodes"] = PhysViewNodes
	physView["edges"] = PhysViewEdges
	//glog.Info(physView)
	newPhysView := FixRepeat(physView)

	res := make(map[string]interface{})
	res["mixed"] = newMixView
	res["logical"] = newLogicView
	res["physical"] = newPhysView
	glog.Info(res["mixed"])
	return res, nil

}


func NodeHandler(n interface{}, nodesSet map[string]bool) (node map[string]string, error error) {
	// 处理nodes，关注n1和n2
	temp := gconv.String(n)
	jsonS := gjson.New(temp)
	id := jsonS.GetString("Id")
	// 去重检查 + 加入
	if _, ok := nodesSet[id]; ok == true{
		//glog.Info("发现重复结点", id)
		return nil, nil
	}
	nodesSet[id] = true

	// 读取其他数据
	// 检查categorys数组格式
	categorys := jsonS.GetArray("Labels")
	category := ""
	if len(categorys) == 1{
		category = categorys[0].(string)
	}else {
		glog.Warning("category格式出现异常", categorys)
	}

	// 初始化node
	node = map[string]string{
		"id" : id,
		"category": category,
	}

	// 处理Device的情况
	if category == "Device"{
		deviceType := jsonS.GetString("Props.type")
		node["deviceType"] = deviceType

		name := jsonS.GetString("Props.name")
		if len(name) == 0{
			name = "null"
		}
		node["name"] = name
	}

	// 处理Interface的情况
	if category == "Interface"{
		// 【终端过滤：interface处理】
		name := jsonS.GetString("Props.ipv4")
		if len(name) == 0{
			name = "null"
		}
		node["name"] = name
	}

	// 处理Network的情况
	if category == "Network"{
		name := jsonS.GetString("Props.network")
		if len(name) == 0{
			name = "null"
		}
		node["name"] = name
	}

	return node, nil
}



func EdgeHandler(r interface{}, edgeSet map[string]bool) (edge map[string]string, error error) {
	temp := gconv.String(r)
	jsonS := gjson.New(temp)
	id := jsonS.GetString("Id")


	// 去重检查 + 加入
	if _, ok := edgeSet[id]; ok == true{
		//glog.Info("发现重复结点", id)
		return nil, nil
	}
	edgeSet[id] = true

	// 读取其他数据
	// source target relationType
	source := jsonS.GetString("StartId")
	target := jsonS.GetString("EndId")
	relationType := jsonS.GetString("Type")


	// 初始化edge
	edge = map[string]string{
		"id" : strings.Join([]string{"edge", id}, "_"),
		"source" : source,
		"target" : target,
		"relationType" : relationType,
	}

	// 处理from，需要判断Props内容
	from := "Unknown"
	if lldp := jsonS.GetInt("Props.lldp"); lldp == 1{
		from = "lldp"
	}else if stp := jsonS.GetInt("Props.stp"); stp == 1{
		from = "stp"
	}

	edge["from"] = from
	return  edge, nil
}

