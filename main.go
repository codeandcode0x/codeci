package main

import (
	"codeci/src"
	"os"
	// "log"
)

func main() {
	runSuite()
}

//run suite
func runSuite() {
	params := os.Args
	num := len(params)
	switch(num) {
	case 2:
		app := params[1]
		src.DeployResourceByLayNodes(app, "true")
		break
	case 3:
		app := params[1]
		strictModel := params[2]
		src.DeployResourceByLayNodes(app, strictModel)
		break
	}
}














