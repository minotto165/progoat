/*
Copyright © 2026 minotto
*/
package cmd

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "progoat",
	Short: "Progate-inspired CLI tool to learn coding with AI",
	Long: `Progoat is a CLI tool that generates custom programming courses using AI. 
It provides an interactive learning experience with explanations and exercises.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".progoat")
	os.MkdirAll(base, 0755)
	configPath := filepath.Join(base, "config.yaml")
	viper.SetConfigType("yaml")
	viper.SetConfigFile(configPath)
	viper.ReadInConfig()
}
