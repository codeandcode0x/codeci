package src
/**
* author: codeandcode0x
* email: codeandcode0x@gmail.com
* data: 2020/05/23
* describe:

* generate node data trees :
* get all nodes from resources
* ResNodes contain all resource data
* RunNodes contain all resource when app run
* LayerNodes contain node layer from node tree
* leafNodes contain node where has no depends
*
* ~ generate node tree from root path
*
* ~ generate node tree from config [app run list]
*
*/

import (
	"github.com/rs/xid"
	"io/ioutil"
	"log"
	"path"
	"strings"
	"encoding/json"
	"os"
	"bufio"
	"codeci/pkg/k8s"
	"path/filepath"
)

//资源
type Resource struct {
	Name string
	Path string
}

//资源节点
type ResNode struct {
	Id string
	Name string
	Depends []string
	Parent string
	ParentId string
	ChildIds []string
	Res Resource
	DataInitPath []string
	IsLeaf bool
}

//层节点 (树)
type LayerNode struct {
	Nodes []*ResNode
}

//节点 map
var treeNodeMap map[string]*[]ResNode

//sql 文件 map
var sqlFileMap map[string]string

//init
func init() {
	treeNodeMap = make(map[string]*[]ResNode, 0)
	sqlFileMap = make(map[string]string, 0)
}

//通过配置文件生成依赖树
func GenerateDependTreeByConfig(apiClient k8s.K8sClientInf, resoucesRoot string, runSvcs []string) []map[string]*ResNode {
	//map
	resourceMap := make(map[string]Resource, 0)
	resNodeMap := make(map[string]*ResNode, 0)
	leafNodeMap := make(map[string]*ResNode, 0)
	runNodeMap := make(map[string]*ResNode, 0)
	//tree nodes
	currentNodes := []*ResNode{}
	allNodes := []LayerNode{}
	//generate node tree
	getAllResourcesFromRoot(apiClient, resoucesRoot, resourceMap, resNodeMap)
	//check service exist
	checkServiceNodeExist(runSvcs, resNodeMap)
	//get run node
	getRunNode(runSvcs, resNodeMap, leafNodeMap, runNodeMap)

	getRunNodeTreeMultiService(runSvcs, resNodeMap, leafNodeMap, runNodeMap)
	for _, appName := range runSvcs {
		currentNodes = append(currentNodes, resNodeMap[appName])
		allNodes = append(allNodes, LayerNode{Nodes: currentNodes})
	}
	getCurrentNodes(currentNodes, runNodeMap, &allNodes)
	layerNodes := removeDuplicateNode(allNodes)
	//write data init status file
	WriteDataStatusFile(sqlFileMap, false)
	//print tree
	// PrintlnRes(layerNodes)
	//返回层节点
	return layerNodes
}

//generate depend tree
func GenerateDependTree(apiClient k8s.K8sClientInf, resoucesRoot, appName string) []map[string]*ResNode {
	//app list
	runSvcs := []string{appName}
	resourceMap := make(map[string]Resource, 0)
	resNodeMap := make(map[string]*ResNode, 0)
	leafNodeMap := make(map[string]*ResNode, 0)
	runNodeMap := make(map[string]*ResNode, 0)
	currentNodes := []*ResNode{}
	allNodes := []LayerNode{}
	//generate node tree
	getAllResourcesFromRoot(apiClient, resoucesRoot, resourceMap, resNodeMap)
	//check service exist
	checkServiceNodeExist(runSvcs, resNodeMap)
	//get run node
	getRunNode(runSvcs, resNodeMap, leafNodeMap, runNodeMap)
	getRunNodeTree(resNodeMap[appName], resNodeMap, leafNodeMap, runNodeMap)
	currentNodes = append(currentNodes, resNodeMap[appName])
	allNodes = append(allNodes, LayerNode{Nodes: currentNodes})
	getCurrentNodes(currentNodes, runNodeMap, &allNodes)
	layerNodes := removeDuplicateNode(allNodes)
	//write data init status file
	WriteDataStatusFile(sqlFileMap, false)
	//print tree
	// PrintlnRes(layerNodes)
	//return
	return layerNodes
}

//remove duplicate node
func removeDuplicateNode(layerNodes []LayerNode) []map[string]*ResNode{
	layerNodesMap := make(map[string]*ResNode, 0)
	layerNodesMapSlice := make([]map[string]*ResNode, 0)
	lnLeafNode := make(map[string]*ResNode, 0)
	layerNodesMapExist := make(map[string]*ResNode, 0)

	for _,layerNode := range layerNodes {
		for _, subNode := range layerNode.Nodes {
			if subNode.IsLeaf == true {
				if _, existLeaf := lnLeafNode[subNode.Name]; existLeaf {
					continue
				}else{
					lnLeafNode[subNode.Name] = subNode
				}
			}else {
				layerNodesMap[subNode.Name] = subNode
			}
		}

		if len(layerNodesMap) > 0 {
			layerNodesMapSlice = append(layerNodesMapSlice, layerNodesMap)
		}
		layerNodesMap = make(map[string]*ResNode, 0)
	}

	if len(lnLeafNode) > 0 {
		layerNodesMapSlice = append(layerNodesMapSlice, lnLeafNode)
	}

	for i:=len(layerNodesMapSlice)-1; i>=0; i-- {
		for _, v := range layerNodesMapSlice[i] {
			if _,exist := layerNodesMapExist[v.Name]; exist {
				delete(layerNodesMapSlice[i], v.Name)
				if len(layerNodesMapSlice[i]) == 0 {
					layerNodesMapSlice = append(layerNodesMapSlice[:i], layerNodesMapSlice[i+1:]...)
				}
				continue
			}else{
				layerNodesMapExist[v.Name] = v
			}
		}
	}

	return layerNodesMapSlice
}

//get run node
func getRunNode(runSvcs []string, resNodeMap, leafNodeMap, runNodeMap map[string]*ResNode) {
	for _, svc := range runSvcs {
		newNode := NewNode(svc, resNodeMap)
		newNode.Id = resNodeMap[svc].Id
		runNodeMap[newNode.Id] = &newNode
	}
}


//get run node tree
func getRunNodeTreeMultiService(svcs []string, resNodeMap, leafNodeMap, runNodeMap map[string]*ResNode) {
	for _,name := range svcs {
		name = strings.TrimSpace(name)
		currentNode := resNodeMap[name]
		newNode := NewNode(name, resNodeMap)
		newNode.Parent = currentNode.Name
		newNode.ParentId = currentNode.Id
		runNodeMap[newNode.Id] = &newNode
		//检测循环依赖
		checkCicleDepend(name, newNode.Id, runNodeMap)
		currentNode.ChildIds = append(currentNode.ChildIds, newNode.Id)
		if len(runNodeMap[newNode.Id].Depends) > 0 {
			getRunNodeTree(runNodeMap[newNode.Id], resNodeMap, leafNodeMap, runNodeMap)
		}
	}
}

//get run node tree
func getRunNodeTree(currentNode *ResNode, resNodeMap, leafNodeMap, runNodeMap map[string]*ResNode) {
	for _,name := range currentNode.Depends {
		name = strings.TrimSpace(name)
		newNode := NewNode(name, resNodeMap)
		newNode.Parent = currentNode.Name
		newNode.ParentId = currentNode.Id
		runNodeMap[newNode.Id] = &newNode
		//检测循环依赖
		checkCicleDepend(name, newNode.Id, runNodeMap)
		currentNode.ChildIds = append(currentNode.ChildIds, newNode.Id)
		if len(runNodeMap[newNode.Id].Depends) > 0 {
			getRunNodeTree(runNodeMap[newNode.Id], resNodeMap, leafNodeMap, runNodeMap)
		}
	}
}

//check cicle depend
func checkCicleDepend(nodeName, currentId string, runNodeMap map[string]*ResNode) {
	checkNodeMap := make(map[string]*ResNode, 0)
	count := 0
	for {
		if runNodeMap[currentId].ParentId == "" { 
			break 
		}
		checkNodeMap[runNodeMap[currentId].Name] = runNodeMap[currentId]
		parentId := runNodeMap[currentId].ParentId
		if nodeName == runNodeMap[currentId].Name {
			count++
		}
		currentId = parentId
	}

	if count >1 {
		log.Fatalf("error: service exist cicle depends ! service: %s in %v",nodeName, checkNodeMap)
	}
}

//new node
func NewNode(index string, resNodeMap map[string]*ResNode) ResNode{
	var resNode ResNode
	checkServiceNodeExist([]string{index}, resNodeMap)
	resNode = *resNodeMap[index]
	resNode.Id = xid.New().String()
	return resNode
}

//get child node
func getChildNode(root *ResNode, resNodeMap, leafNodeMap map[string]*ResNode) {
	for _, depend :=range root.Depends {
		if _, exist := leafNodeMap[depend]; exist {
			leafNodeMap[depend].Parent = root.Name
			leafNodeMap[depend].ParentId = root.Id
			DeployAllResourceFiles(leafNodeMap[depend].Res.Path, leafNodeMap[depend])
		}else{
			resNodeMap[depend].Parent = root.Name
			resNodeMap[depend].ParentId = root.Id
		}
	}
}

//get leaf node service
func getLeafNodeServcie(resNodeMap, leafNodeMap map[string]*ResNode) {
	for _, rm := range resNodeMap {
		if rm.IsLeaf == true {
			leafNodeMap[rm.Name] = resNodeMap[rm.Name]
			// delete(resNodeMap, rm.Name)
		}
	}
}

//check node service
func checkServiceNodeExist(runSvcs []string, resNodeMap map[string]*ResNode) {
	for _,appName := range runSvcs {
		if _,exist := resNodeMap[appName]; !exist {
			log.Fatalln(appName, "service not exist!")
		}
	}
}

//get all resource files
func getAllResourcesFromRoot(apiClient k8s.K8sClientInf, pathName string, resourceMap map[string]Resource, resNodeMap map[string]*ResNode) {
    rd, err := ioutil.ReadDir(pathName)
    if err != nil {
    	log.Fatalln("read file error!", err)
    }
    //resource file path
    for _, fi := range rd {
        if fi.IsDir() {
            getAllResourcesFromRoot(apiClient, pathName +"/"+fi.Name(), resourceMap, resNodeMap)
        } else {
        	filePath := pathName +"/"+fi.Name()
        	if path.Ext(filePath) == ".yaml" && strings.Contains(filePath, "deployment.yaml") {
        		deployment := apiClient.UnmarshalDeployment(GetResourceYaml(filePath))
	        	name := deployment.ObjectMeta.Name
	        	if _, res := resourceMap[name]; !res {
	        		var resNode ResNode
	        		var res Resource
	        		res.Name = name
	        		res.Path = pathName
	        		resourceMap[name] = res

	        		resNode.Id = xid.New().String()
	        		resNode.Name = name
	        		resNode.Res = resourceMap[name]
	        		//depends
	        		depends := deployment.ObjectMeta.Annotations["dependOn"]
	        		if depends != "" {
						resNode.Depends = strings.Split(depends, ",")
					}
					//leaf node
					resNode.IsLeaf = false
					if len(resNode.Depends) == 0 {
						resNode.IsLeaf = true
					}
					//add node
					resNodeMap[name] = &resNode
					//add init data
					addDataInitPath(resNodeMap[name])
	        	}
        	}
        }
    }
}

//add data init path
func addDataInitPath(resNode *ResNode) {
	rd, err := ioutil.ReadDir(resNode.Res.Path)
	if err != nil {
    	log.Fatalln("scan data file - read file error!", err)
    }

	for _, fi := range rd {
	 	if fi.Name() == "data" {
	 		scanDataFile(resNode.Res.Path+"/data", resNode)
	 		break
	 	}
	}
}

//scan data file
func scanDataFile(pathName string, resNode *ResNode) {
	rd, err := ioutil.ReadDir(pathName)
	if err != nil {
    	log.Fatalln("scan data file - read file error!", err)
    }

	for _, fi := range rd {
		if !fi.IsDir() {
			dataExt := path.Ext(fi.Name())
			switch(dataExt) {
			case ".sql":
				resNode.DataInitPath = append(resNode.DataInitPath,  pathName+"/"+fi.Name())
				sqlFileMap[pathName+"/"+fi.Name()] = "false"
				break
			}
		}else{
			scanDataFile(pathName+"/"+fi.Name(), resNode)
		}
	 }
}

//write data init status file
func WriteDataStatusFile(dataInitFileMap map[string]string, writeOver bool) {
	if writeOver {
		WriteDataFile(dataInitFileMap)
	}else{
		wStatus := true
		//data status file whether exist
		dataFilePath := filepath.Join(os.Getenv("HOME"), ".codeci", "data-init-status.json")
		status := FileExist(dataFilePath)
		if status {
			dataJson := ReadDataFile(dataFilePath)
		    if len(dataJson) >0 {
		    	wStatus = false
		    }
		}
		//init data
		if wStatus {
			WriteDataFile(dataInitFileMap)
		}
	}
}

//write data file
func WriteDataFile(dataInitFileMap map[string]string) {
	//create file
	dataFilePath := filepath.Join(os.Getenv("HOME"), ".codeci", "data-init-status.json")
	dataFile, err := os.Create(dataFilePath)
	if err != nil {
		log.Fatalln("create sql status file failed!", err)
	}
	//close file
	defer dataFile.Close()
	//sql file map to json
	sqlBytes, err := json.Marshal(dataInitFileMap)
	if err != nil {
		log.Fatalln("write sql status file failed!")
	}
	//create bufio write
	dataFileWriter := bufio.NewWriter(dataFile)
	_, errWrite := dataFileWriter.WriteString(string(sqlBytes))
	if errWrite != nil {
		log.Fatalln("write sql status file err!", errWrite)
	}
	dataFileWriter.Flush()
}

//get current nodes
func getCurrentNodes(currentNodes []*ResNode, runNodeMap map[string]*ResNode, allNodes *[]LayerNode) {
	tempNodes := currentNodes
	currentNodes = []*ResNode{}
	for _, currentNode := range tempNodes {
		if len(currentNode.ChildIds) > 0 {
			for _, Id := range currentNode.ChildIds {
				if _, exist := runNodeMap[Id]; exist {
					currentNodes = append(currentNodes, runNodeMap[Id])
				}
			}
		}
	}

	var layerNode LayerNode
	layerNode.Nodes = currentNodes
	*allNodes = append(*allNodes, layerNode)
	if len(currentNodes) > 0 {
		getCurrentNodes(currentNodes, runNodeMap, allNodes)
	}
}




