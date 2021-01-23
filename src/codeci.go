package src

//run suite
func RunSuiteWithCliModel(app, strictModel, configPath string) {
	defer Catch()
	DeployResourceByLayNodes(app, strictModel, configPath)
}


//run suite
func RunSuiteWithAnalyseModel(app, configPath string) {
	defer Catch()
	AnalyseServiceDepOn(app, configPath)
}














