package src

//run suite
func RunSuiteWithCliModel(app, strictModel, configPath string) {
	DeployResourceByLayNodes(app, strictModel, configPath)
}


//run suite
func RunSuiteWithAnalyseModel(app, configPath string) {
	AnalyseServiceDepOn(app, configPath)
}














