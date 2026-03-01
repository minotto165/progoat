/*
Copyright © 2026 minotto
*/
package cmd

import (
	"fmt"

	"github.com/minotto165/progoat/internal/course"
	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:          "list",
	Short:        "List all generated courses",
	Long:         `Display all learning courses available on your computer.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		courses, err := course.GetCourses(coursesPath)
		if err != nil {
			return err
		}

		maxLength := 30
		fmt.Printf("%-30s %s\n", "COURSE ID", "TITLE")
		fmt.Println("----------------------------------------------------")

		for _, course := range courses {
			id := course.ID
			title := course.Title
			if len(id) > maxLength {
				id = string([]rune(id)[:maxLength-3])
				id += "..."
			}

			fmt.Printf("%-30s %s\n", id, title)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
