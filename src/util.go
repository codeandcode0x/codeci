package src

import (
	"github.com/ghodss/yaml"
	appsv1 "k8s.io/api/apps/v1"
	"encoding/json"
	"io/ioutil"
	"sync"
	"path"
	"log"
	"strings"
	"time"
	"fmt"
	"os"
	"os/user"
	"codeci/src/k8s"
	"runtime"
	"bytes"
	"errors"
	"os/exec"
)


//get resource yaml
func GetResourceYaml(filePath string) []byte {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err.Error())
	}
	return bytes
}

//getResourceType
func GetResourceType(bytes []byte) string{
	typeJson, err := yaml.YAMLToJSON(bytes)
	if err != nil {
		panic("get resource type error !")
	}
	var typeIn interface{}
	json.Unmarshal(typeJson, &typeIn)
	return typeIn.(map[string]interface{})["kind"].(string)
}

//init pod data
func InitPodData(apiClient k8s.K8sClientInf, dataType string, namespace, path, padName, containerName, dbUser, dbPasswd, dbName string) {
	queryBytes, _ := ioutil.ReadFile(path)
	query := string(queryBytes)
	query = strings.Replace(query, "`", "\\`", -1)
	querySlice := strings.Split(query, ";")
	sql := ""
	for k, queryStr := range querySlice {
		if len(queryStr) <2 {continue}
		sql = sql +";"+ queryStr
		if k%50 == 0 && k >0 {
			sql = sql + ";"
			command := "mysql -u"+dbUser+" -p"+dbPasswd+" "+dbName+" -e \""+string(sql)+"\""
			_, _, err := apiClient.PodExecCommand(namespace, padName, command, containerName)
			if err != nil {
				log.Println("data exists or sql error!")
			}
			sql = ""
		}
	}
	sql = sql + ";"
	command := "mysql -u"+dbUser+" -p"+dbPasswd+" "+dbName+" -e \""+string(sql)+"\""
	_, _, err := apiClient.PodExecCommand(namespace, padName, command, containerName)
	if err != nil {
		log.Println("data exists or sql error!")
	}
}

//get data init file map
func GetDataInitFile() map[string]string {
	dataInitFileMap := make(map[string]string, 0)
	dataBytes := ReadDataFile("./run/data-init-status.json")
	json.Unmarshal(dataBytes, &dataInitFileMap)
	return dataInitFileMap
}

//check pod status
func CheckPodStatus(apiClient k8s.K8sClientInf, deploy appsv1.Deployment) bool{
	checkStatus := false
	for {
		deploys := apiClient.GetDeployments(deploy.Name)
		if len(deploys) >0 {
			if deploys[0].Status.Replicas >0 {
				for _, condition := range deploys[0].Status.Conditions {
					if condition.Type == "Available" && condition.Status == "True" {
						checkStatus = true
						break
					}
				}
			}
		}

		if checkStatus {
			return true
		}
		fmt.Print(".")
		time.Sleep(2*time.Second)
	}

	return false
}

//check layer nodes (pods) status
func CheckLayNodesStatus(apiClient k8s.K8sClientInf, deploys []appsv1.Deployment) {
	wg := sync.WaitGroup{}
    wg.Add(len(deploys))
	for _, deploy := range deploys {
		go func(apiClient k8s.K8sClientInf, deploy appsv1.Deployment) {
    		CheckPodStatus(apiClient, deploy)
    		wg.Done()
    	}(apiClient, deploy)
	}
	wg.Wait()
}

//pod reset
func PodReset(apiClient k8s.K8sClientInf, resNodeMap map[string]*ResNode, dataInitFileMap map[string]string) map[string]string{
	//scan resource file
	for _, resNode := range resNodeMap {
		ScanResFile(apiClient, resNode.Res.Path)
		//update data init file
		for i:=len(resNode.DataInitPath)-1 ;i>=0;i-- {
			dataInitFileMap[resNode.DataInitPath[i]] = "false"
		}
	}
	return dataInitFileMap
}

//scan data file
func ScanResFile(apiClient k8s.K8sClientInf, pathName string) {
	rd, err := ioutil.ReadDir(pathName)
	if err != nil {
    	log.Fatalln("scan data file - read file error!", err)
    }
    //resource delete
	for _, fi := range rd {
		if !fi.IsDir() {
			dataExt := path.Ext(fi.Name())
			switch(dataExt) {
			case ".yaml":
				bytes := GetResourceYaml(pathName+"/"+fi.Name())
				resourceType := GetResourceType(bytes)
				apiClient.ResDelete(resourceType, bytes)
				break
			}
		}else{
			ScanResFile(apiClient, pathName+"/"+fi.Name())
		}
	 }
}

//read data file
func ReadDataFile(path string) []byte {
	data, err := ioutil.ReadFile(path)
    if err != nil {
    	log.Fatalln("read "+path+" file failed!", err)
    }
    return data
}

//tools print
func PrintlnRes(layerNodes []map[string]*ResNode) {
	index := 0
	for k,v := range layerNodes {
		fmt.Println("[ layer:", k+1, ",", "service num:", len(v), "]")
		for _, subV := range v {
			if index == 0 {
				fmt.Print(subV.Name)
			}else{
				fmt.Print(", ", subV.Name)
			}
			index++
		}
		fmt.Println("\n")
		index = 0
	}
}

//file exist
func FileExist(path string) bool {
  _, err := os.Lstat(path)
  return !os.IsNotExist(err)
}

//resource print
func ResourcePrint(res map[string]*ResNode) {
	for _, v := range res {
		fmt.Println(v.Id, v.ParentId, v.Name, v.Parent)
	}
}

//println
func PrintLog() {
	//to-do list
	fmt.Println("")
	//add logs
}

//get Home
func GetHome() (string, error) {
    user, err := user.Current()
    if nil == err {
        return user.HomeDir, nil
    }

    // cross compile support
    if "windows" == runtime.GOOS {
        return homeWindows()
    }

    // Unix-like system, so just assume Unix
    return homeUnix()
}

func homeUnix() (string, error) {
    // First prefer the HOME environmental variable
    if home := os.Getenv("HOME"); home != "" {
        return home, nil
    }

    // If that fails, try the shell
    var stdout bytes.Buffer
    cmd := exec.Command("sh", "-c", "eval echo ~$USER")
    cmd.Stdout = &stdout
    if err := cmd.Run(); err != nil {
        return "", err
    }

    result := strings.TrimSpace(stdout.String())
    if result == "" {
        return "", errors.New("blank output when reading home directory")
    }

    return result, nil
}

func homeWindows() (string, error) {
    drive := os.Getenv("HOMEDRIVE")
    path := os.Getenv("HOMEPATH")
    home := drive + path
    if drive == "" || path == "" {
        home = os.Getenv("USERPROFILE")
    }
    if home == "" {
        return "", errors.New("HOMEDRIVE, HOMEPATH, and USERPROFILE are blank")
    }

    return home, nil
}


//file or folder check exists
func FFExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

