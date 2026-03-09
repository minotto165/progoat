/*
Copyright © 2026 minotto
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/huh"
	"github.com/minotto165/progoat/internal/course"
	"github.com/spf13/cobra"
)

// removeCmd represents the remove command
var removeCmd = &cobra.Command{
	Use:   "remove [CourseID]",
	Short: "Remove a generated course",
	Long: `Permanently delete a specific course and its files from your computer. 
You can select a course from the list or provide the CourseID as an argument.`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var courseID string

		if len(args) > 0 {
			courseID = args[0]
		} else {

			courses, err := course.GetCourses(coursesPath)
			if err != nil {
				return err
			}

			options := []huh.Option[string]{}
			for _, c := range courses {
				title := c.Title
				id := c.ID
				key := fmt.Sprint(title, "(id: ", id, ")")
				options = append(options, huh.NewOption(key, id))
			}

			form := huh.NewForm(
				huh.NewGroup(
					huh.NewSelect[string]().
						Title("Choose Course").
						Options(options...).
						Value(&courseID),
				),
			).WithTheme(huh.ThemeBase())
			err = form.Run()
			if err != nil {
				return err
			}
		}

		// 安全装置
		baseID := filepath.Base(courseID)
		if baseID == ".." || baseID == "." || baseID == "/" || baseID == "\\" {
			return fmt.Errorf("invalid course ID: %s", courseID)
		}
		path := filepath.Join(coursesPath, baseID)

		err := os.RemoveAll(path)
		if err != nil {
			return err
		}

		fmt.Printf("Deleted %s.", courseID)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// removeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// removeCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
