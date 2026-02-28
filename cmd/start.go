/*
Copyright © 2026 minotto
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/glamour"
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

		course, err := getCourseStruct(courseID)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		coursePath := filepath.Join(coursesDir, course.Title)

		fmt.Println("[INFO] Course Directory:", coursePath)

		for _, c := range course.Lessons {
			slides := c.Slides
			for _, s := range slides {
				out, err := glamour.Render(s, "dark")
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				fmt.Println(out)
				fmt.Println("----------------------")
			}
		}

	},
}

func getCourses() ([]Course, error) {
	files, err := os.ReadDir(coursesDir)
	if err != nil {
		return nil, err
	}
	var courses []Course
	for _, file := range files {
		if file.IsDir() {
			dirName := file.Name()
			coursesJsonPath := filepath.Join(coursesDir, dirName, "course.json")
			coursesJson, err := os.ReadFile(coursesJsonPath)
			if err != nil {
				return nil, err
			}
			// Convert to struct
			var course Course
			err = json.Unmarshal([]byte(coursesJson), &course)
			if err != nil {
				return nil, fmt.Errorf("Error parsing JSON: %s", err)
			}

			// Add to slice
			courses = append(courses, course)
		}
	}
	return courses, nil
}

func getCourseStruct(courseID string) (Course, error) {
	courses, err := getCourses()
	course := Course{}
	found := 0
	if err != nil {
		return course, err
	}

	for _, c := range courses {
		ID := c.ID
		if ID == courseID {
			course = c
			found = 1
			break
		}
	}
	if found == 0 {
		return course, fmt.Errorf("No such a course: %s", courseID)
	}

	return course, nil
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
