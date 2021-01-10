package cmd

import (
    "github.com/spf13/cobra"
    codeci "codeci/src"
)


var apps, strictModel string

//init
func init() {
    //migration data
    codeciDeployCmd.Flags().StringVarP(&strictModel, "model", "m", "true", "run with strict model")
    codeciDeployCmd.Flags().StringVarP(&apps, "apps", "s", "all", "applications")
    rootCmd.AddCommand(codeciDeployCmd)

    codeciResetCmd.Flags().StringVarP(&apps, "apps", "s", "all", "applications")
    rootCmd.AddCommand(codeciResetCmd)
}

//register command
var codeciDeployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Deploy service from service folder",
    Long: `deploy service`,
    Run: func(cmd *cobra.Command, args []string) {
        //TO-DO
        if len(args) > 0 {
            if len(args) == 1 {
                codeci.RunSuiteWithCliModel(args[0], "true")
            }else{
                codeci.RunSuiteWithCliModel(args[0], args[1])
            }
        }else{
            codeci.RunSuiteWithCliModel(apps, strictModel)
        }
    },
}


//register command
var codeciResetCmd = &cobra.Command{
    Use:   "reset",
    Short: "Reset service from Kubernetes",
    Long: `reset service`,
    Run: func(cmd *cobra.Command, args []string) {
        //TO-DO
        strictModel = "reset"
        if len(args) > 0 {
            if len(args) == 1 {
                codeci.RunSuiteWithCliModel(args[0], strictModel)
            }
        }else{
            codeci.RunSuiteWithCliModel(apps, strictModel)
        }
    },
}



//register command
var codeciAnalyseCmd = &cobra.Command{
    Use:   "analyse",
    Short: "Analyse services from depends relationship",
    Long: `Analyse services`,
    Run: func(cmd *cobra.Command, args []string) {
        //TO-DO
        strictModel = "analyse"
        if len(args) > 0 {
            if len(args) == 1 {
                codeci.RunSuiteWithCliModel(args[0], strictModel)
            }
        }else{
            codeci.RunSuiteWithCliModel(apps, strictModel)
        }
    },
}







