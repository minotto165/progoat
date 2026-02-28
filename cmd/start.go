/*
Copyright © 2026 minotto
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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

		for _, l := range course.Lessons {

			slides := l.Slides
			for i, s := range slides {
				clearScreen()
				out, err := glamour.Render(s, "dark")
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				title := fmt.Sprint(course.Title, ": page ", i+1)
				fmt.Println(title)
				fmt.Print(out)

				fmt.Print("Press Enter to next page...")
				fmt.Scanln()
			}

			clearScreen()

			task := fmt.Sprintf("%s\n%s\n\n**File to edit:**\n```text\n%s\n```",
				"## Task:",
				l.TaskDescription,
				filepath.Clean(filepath.Join(coursePath, l.FileName)),
			)
			out, err := glamour.Render(task, "dark")
			if err != nil {
				return
			}

			title := fmt.Sprint(course.Title, ": task")
			fmt.Println(title)

			fmt.Print(out)

			fmt.Print("Edit and save the file, then hit Enter.")
			fmt.Scanln()

		}

	},
}

func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Print("\033[H\033[2J") // Unix系のクリアコマンド
	}
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
