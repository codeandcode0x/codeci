package src
/**
* author: codeandcode0x
* email: codeandcode0x@gmail.com
* data: 2020/05/23

* describe:
* two ways to deploy apps: 
* 1. apply simple app/service root fold ;
* 2. deploy form config app run list; 
*
* run like this:
* ~ run apps: # cli [run]
*
* ~ run exact app: # cli nginx [namespace] [flag]
*
*/

import (
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/api/apps/v1"
	"path/filepath"
	"github.com/drone/envsubst"
	"io/ioutil"
	"sync"
	"os"
	"log"
	"encoding/json"
	"strings"
	"path"
	"codeci/src/k8s"
)

type Config struct {
	NameSpace string `json:"namespace"`
	ServicePath string `json:"servicepath"`
	DbSrcName  string `json:"dbsrcname"`
	NoCheck []string    `json:"nocheck"`
	AppRun []string  `json:"apprun"`
	DbUser string `json:"dbuser"`
	DbPasswd string `json:"dbpasswd"`
	DbName string `json:"dbname"`
}

var clientset *kubernetes.Clientset
var noCheckNodes, appRunNodes, dataInitFileMap map[string]string
var namespace, dbSrcName, servicePath, dbUser, dbPasswd, dbName string
var appConfig Config
var apps []string
var apiClient k8s.K8sClientInf
var checkLayNodes []appsv1.Deployment
var configPath string


//init
func initKubeConfig() {
	var config *rest.Config
	kubecfgpath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	//kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubecfgpath)
	if err != nil {
		kubecfgpath = filepath.Join("./run", "kubeconfig")
		config, err = clientcmd.BuildConfigFromFlags("", kubecfgpath)
		if err != nil {
			panic("~/.codeci/config or ./kube/config are not exist ! \nyou can vim ~/.codeci/kubeconfig or vim ~/.kube/config")
		}
	}
	//client set
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}

	//get home path
	configPath, _ = GetHome()
	configPath = configPath + "/.codeci"
	//init config
	initConfig(configPath)

	apiClient = &k8s.K8sClient{
		clientset,
		namespace,
	}
}

//init config
func initConfig(cfgPath string) {
	noCheckNodes,appRunNodes = make(map[string]string, 0), make(map[string]string, 0)
	deploycfgpath := filepath.Join(cfgPath, "deployconfig.json")
	cfg, err := ioutil.ReadFile(deploycfgpath)
	err = json.Unmarshal(cfg, &appConfig)
	if err != nil {
		log.Println("warnning: deployconfig file not exist in "+ cfgPath +" ! ")
		deploycfgpath = filepath.Join("./conf/dev", "deployconfig.json")
		cfg, err = ioutil.ReadFile(deploycfgpath)
		err = json.Unmarshal(cfg, &appConfig)
		if err != nil {
			log.Println("error: deployconfig file not exist in ./conf ! ")
		}
	}
	//set env ns
	os.Setenv("NAMESPACE", appConfig.NameSpace)
	namespace = appConfig.NameSpace
	//set db src name
	dbSrcName = appConfig.DbSrcName
	//set resource root
	servicePath = appConfig.ServicePath
	//set no check apps
	for _,v := range appConfig.NoCheck {
		noCheckNodes[v] = v
	}
	//set run apps 
	for _,v := range appConfig.AppRun {
		appRunNodes[v] = v
	}
	//app run list
	apps = appConfig.AppRun
	//set db config
	dbUser = appConfig.DbUser
	dbPasswd = appConfig.DbPasswd
	dbName = appConfig.DbName
}

//deploy resource by layer nodes
func DeployResourceByLayNodes(app, strictModel, cfgPath string) error {
	//init kube config
	initKubeConfig()
	//init config
	if cfgPath != "" {
		initConfig(cfgPath)
	}
	//layer nodes
	var layNodes []map[string]*ResNode
	//get app list
	appList := []string{}
	if app == "all" {
		appList = apps
	}else{
		appList = append(appList, app)
	}
	//get app depend tree
	layNodes = GenerateDependTreeByConfig(apiClient, servicePath, appList)
	//read data init file
	dataInitFileMap = GetDataInitFile()
	//set strict model
	if strictModel == "false" {
		layNodes = layNodes[:1]
	}else if strictModel == "reset" {
		if app != "all" {
			layNodes = layNodes[:1]
		}
		//pod reset
		for _,apps := range layNodes {
			dataInitFileMap = PodReset(apiClient, apps, dataInitFileMap)
		}
		//write data status in file
		WriteDataStatusFile(dataInitFileMap, true)
		return nil
	}

	//deploy services
    for i:=len(layNodes)-1; i>=0 ; i-- {
    	wg := sync.WaitGroup{}
    	wg.Add(len(layNodes[i]))
    	for _,v := range layNodes[i] {
    		go func(path string, resNode *ResNode) {
    			DeployAllResourceFiles(path, resNode)
    			wg.Done()
    		}(v.Res.Path, v)
    		//check layer nodes pods status
    		CheckLayNodesStatus(apiClient, checkLayNodes)
    		checkLayNodes = []appsv1.Deployment{}
    	}
    	wg.Wait()
    }
    //save data status
    WriteDataStatusFile(dataInitFileMap, true)
    return nil
}

//deploy all resource files
func DeployAllResourceFiles(pathname string, resNode *ResNode) error {
    rd, err := ioutil.ReadDir(pathname)
    for _, fi := range rd {
        if fi.IsDir() {
            DeployAllResourceFiles(pathname +"/"+fi.Name(), resNode)
        } else {
        	filePath := pathname +"/"+fi.Name()
        	if path.Ext(filePath) == ".yaml" {
        		DeployResource(filePath, resNode)
        	}
        }
    }
    return err
}

//deploy resource
func DeployResource(filePath string, resNode *ResNode) {
	//get resource bytes
	resourceStr, err := envsubst.EvalEnv(string(GetResourceYaml(filePath)))
	bytes := []byte(resourceStr)
	if err != nil {
		log.Fatalln("replace env params failed, file path:" + filePath + ". error: " + err.Error())
	}
	//get resource type
	resourceType := GetResourceType(bytes)
	switch(resourceType) {
	case "Service":
		spec := apiClient.UnmarshalService(bytes)
		apiClient.DeployService(spec)
		// deployService(spec)
		break
	case "Deployment":
		spec := apiClient.UnmarshalDeployment(bytes)
		apiClient.DeployDeployment(spec)
		//init data
		if len(resNode.DataInitPath)>0 {
			dbDeployments:= apiClient.GetDeployments(dbSrcName)
			if len(dbDeployments) >0 {
				if CheckPodStatus(apiClient, *dbDeployments[0])  {
					pods := apiClient.GetPodsByLabel(dbSrcName)
					podName := pods.Items[0].Name
					for i:=len(resNode.DataInitPath)-1 ;i>=0;i-- {
						queryBytes, _ := ioutil.ReadFile(resNode.DataInitPath[i])
						query := string(queryBytes)
						query = strings.Replace(query, "`", "\\`", -1)
						if _, sqlExist := dataInitFileMap[resNode.DataInitPath[i]]; sqlExist {
							if dataInitFileMap[resNode.DataInitPath[i]] == "false" {
								PrintLog()
								log.Print(resNode.DataInitPath[i])
								dataInitFileMap[resNode.DataInitPath[i]] = "true"
								InitPodData(apiClient, "sql", namespace, resNode.DataInitPath[i], podName, dbSrcName, dbUser, dbPasswd, dbName)
							}
						}else{
							InitPodData(apiClient, "sql", namespace, resNode.DataInitPath[i], podName, dbSrcName, dbUser, dbPasswd, dbName)
							dataInitFileMap[resNode.DataInitPath[i]] = "true"
						}
						
					}
				}
			}
		}
		//check pod
		//检测是否在 no check 中
		if _,exist := noCheckNodes[spec.Name]; !exist {
			CheckPodStatus(apiClient, spec)
			// checkLayNodes = append(checkLayNodes, spec)
		}
		
		break
	case "ConfigMap":
		spec := apiClient.UnmarshalConfigMap(bytes)
		apiClient.DeployConfigMap(spec)
		break
	case "StatefulSet":
		break
	default:
		break
	}
}

//analyse service depend relation ship
func AnalyseServiceDepOn(app, cfgPath string) {
	//init config
	if cfgPath != "" {
		initConfig(cfgPath)
	}
	//layer nodes
	var layNodes []map[string]*ResNode
	//get app list
	appList := []string{}
	if app == "all" {
		appList = apps
	}else{
		appList = append(appList, app)
	}

	layNodes = GenerateDependTreeByConfig(apiClient, servicePath, appList)
	PrintlnRes(layNodes)
}









