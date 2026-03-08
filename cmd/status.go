/*
Copyright © 2026 minotto
*/
package cmd

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/minotto165/progoat/internal/course"
	"github.com/minotto165/progoat/internal/ui"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:          "status",
	Short:        "Check your learning progress",
	Long:         `Show your current progress and a list of completed lessons.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {

		fmt.Print("\n")

		progresses, err := course.LoadProgresses(progressPath)
		if err != nil {
			return err
		}

		var lastProgress course.Progress
		for _, p := range progresses {
			if p.LastAccessed.After(lastProgress.LastAccessed) {
				lastProgress = p
			}
		}

		lastCourse, err := course.GetCourseStruct(lastProgress.CourseID, coursesPath)
		if err != nil {
			return err
		}

		//-----------------
		// Header
		//-----------------
		fmt.Println("Progoat Learning Dashboard 🐐")
		fmt.Println("===========================================")
		fmt.Print("\n")

		//-----------------
		// Current Session
		//-----------------
		percentage := 100 * len(lastProgress.CompletedLessons) / lastProgress.TotalLessons

		fmt.Println("[ Current Session ]")
		fmt.Printf("%-10s %s\n", "Course:", lastCourse.Title)
		fmt.Printf("%-10s %s %d%% (%d/%d Lessons)\n", "Progress:", ui.DrawProgressbar(50, 30), percentage, len(lastProgress.CompletedLessons), lastProgress.TotalLessons)
		fmt.Printf("%-10s %s\n", "Next:", lastProgress.CurrentLesson)

		fmt.Print("\n")

		idStyle := lipgloss.NewStyle().Width(15)
		titleStyle := lipgloss.NewStyle().Width(40)
		statusStyle := lipgloss.NewStyle().Width(15)

		//-----------------
		// All Course
		//-----------------
		fmt.Println("[ All Courses ]")

		fmt.Printf("%s %s %s\n",
			idStyle.Render("ID"),
			titleStyle.Render("TITLE"),
			statusStyle.Render("STATUS"))

		for _, p := range progresses {
			c, err := course.GetCourseStruct(p.CourseID, coursesPath)
			if err != nil {
				return err
			}

			progressStatus, _, err := course.LoadProgressStatus(p.CourseID, progressPath)
			if err != nil {
				return err
			}

			var status string

			switch progressStatus {
			case course.Completed:
				status = "✅ Completed"

			case course.InProgress:
				status = "🏃 In Progress"

			default:
				status = "💤 Not Started"
			}
			if err != nil {
				return err
			}

			fmt.Printf("%s %s %s\n",
				idStyle.Render(p.CourseID),
				titleStyle.Render(c.Title),
				statusStyle.Render(status))
		}

		//-----------------
		// Footer
		//-----------------

		fmt.Print("\n")
		fmt.Println("===========================================")
		fmt.Println("Run 'progoat start [ID]' to continue your lesson!")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// statusCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// statusCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
