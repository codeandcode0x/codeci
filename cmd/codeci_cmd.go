package cmd

import (
    "github.com/spf13/cobra"
    codeci "codeci/src"
)

var apps, strictModel, configPath string

//init
func init() {
    //migration data
    codeciDeployCmd.Flags().StringVarP(&strictModel, "model", "m", "true", "run with strict model")
    codeciDeployCmd.Flags().StringVarP(&apps, "apps", "s", "all", "applications")
    codeciDeployCmd.Flags().StringVarP(&configPath, "config", "c", "", "config path")
    rootCmd.AddCommand(codeciDeployCmd)

    codeciResetCmd.Flags().StringVarP(&apps, "apps", "s", "all", "applications")
    codeciResetCmd.Flags().StringVarP(&configPath, "config", "c", "", "config path")
    rootCmd.AddCommand(codeciResetCmd)

    codeciAnalyseCmd.Flags().StringVarP(&apps, "apps", "s", "all", "applications")
    codeciAnalyseCmd.Flags().StringVarP(&configPath, "config", "c", "", "config path")
    rootCmd.AddCommand(codeciAnalyseCmd)
}

//register deploy command
var codeciDeployCmd = &cobra.Command{
    Use:   "deploy",
    Short: "Deploy service from service folder",
    Long: `deploy service`,
    Run: func(cmd *cobra.Command, args []string) {
        //TO-DO
        if len(args) > 0 {
            if len(args) == 1 {
                codeci.RunSuiteWithCliModel(args[0], "true", configPath)
            }else{
                codeci.RunSuiteWithCliModel(args[0], args[1], configPath)
            }
        }else{
            codeci.RunSuiteWithCliModel(apps, strictModel, configPath)
        }
    },
}

//register reset command
var codeciResetCmd = &cobra.Command{
    Use:   "reset",
    Short: "Reset service",
    Long: `reset service`,
    Run: func(cmd *cobra.Command, args []string) {
        //TO-DO
        strictModel = "reset"
        if len(args) > 0 {
            if len(args) == 1 {
                codeci.RunSuiteWithCliModel(args[0], strictModel, configPath)
            }
        }else{
            codeci.RunSuiteWithCliModel(apps, strictModel, configPath)
        }
    },
}

//register analyse command
var codeciAnalyseCmd = &cobra.Command{
    Use:   "analyse",
    Short: "Analyse services dependency relationship",
    Long: `Analyse services`,
    Run: func(cmd *cobra.Command, args []string) {
        //TO-DO
        codeci.RunSuiteWithAnalyseModel(apps, configPath)
    },
}

//register config command
var codeciConfigCmd = &cobra.Command{
    Use:   "config",
    Short: "Set config",
    Long: `Set config`,
    Run: func(cmd *cobra.Command, args []string) {
        //TO-DO
    },
}







