/*
Copyright © 2026 minotto
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [CourseID]",
	Short: "Start a learning session",
	Long: `Begin the selected course. 
Read the slides, write your code, and get feedback from the AI judge.`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		courseID := args[0]
		coursePath, err := getCoursePath(courseID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("[INFO] Course Directory:", coursePath)

	},
}

func getCoursePath(courseID string) (string, error) {
	files, err := os.ReadDir(coursesDir)
	if err != nil {
		return "", err
	}

	var path string

	for _, file := range files {
		if file.IsDir() {
			dirName := file.Name()
			if dirName == courseID {
				path = filepath.Join(coursesDir, dirName)
				break
			}
		}
	}
	if path == "" {
		return "", fmt.Errorf("No such a course: %s", courseID)
	}

	return path, nil
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
