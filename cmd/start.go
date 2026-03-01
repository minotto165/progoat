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
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/huh"
	"github.com/minotto165/progoat/internal/course"
	"github.com/minotto165/progoat/internal/llm"
	"github.com/minotto165/progoat/internal/ui"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
)

type JudgeResult struct {
	IsCorrect bool   `json:"is_correct"`
	Advice    string `json:"advice"`
}

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start [CourseID]",
	Short: "Start a learning session",
	Long: `Begin the selected course. 
Read the slides, write your code, and get feedback from the AI judge.`,
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

		err := startCourse(courseID)
		if err != nil {
			return err
		}
		return nil

	},
}

func startCourse(courseID string) error {

	course, err := course.GetCourseStruct(courseID, coursesPath)
	if err != nil {
		return err
	}
	coursePath := filepath.Join(coursesPath, filepath.Base(course.ID))

	fmt.Println("[INFO] Course Directory:", coursePath)

	for _, l := range course.Lessons {

		ui.ClearScreen()
		slides := l.Slides
		for i, s := range slides {
			out, err := ui.RenderWithTerminalWidth(s)
			if err != nil {
				return err
			}
			title := fmt.Sprint(course.Title, " - ", l.Title, ": page ", i+1)
			fmt.Println(title)
			fmt.Print(out)

			fmt.Print("[Enter] Next page")
			fmt.Scanln()
			fmt.Print("\033[1A\033[K")
			fmt.Print("\n\n\n")

		}

		lessonPath := filepath.Clean(filepath.Join(coursePath, filepath.Base(l.ID)))
		filePath := filepath.Clean(filepath.Join(lessonPath, filepath.Base(l.FileName)))
		task := fmt.Sprintf("%s\n%s\n\n**File to edit:**\n```text\n%s\n```",
			"## Task:",
			l.TaskDescription,
			filePath,
		)
		out, err := ui.RenderWithTerminalWidth(task)
		if err != nil {
			return err
		}

		title := fmt.Sprint(course.Title, " - ", l.Title, ": task")
		fmt.Println(title)

		fmt.Print(out)
		for {

			fmt.Print("Edit and save the file, then hit Enter.")
			fmt.Scanln()

			fmt.Print("\033[1A\033[K")
			fmt.Print("\n\n")

			response, err := judge(l, course.ProgrammingLanguage, filePath)

			//for DEBUG...
			// response, err = JudgeResult{
			// 	IsCorrect: false,
			// 	Advice:    "TEMP",
			// }, nil

			if err != nil {
				return err
			}

			isCorrect := response.IsCorrect
			advice := response.Advice

			var result string
			var enterMessage string

			if isCorrect {
				result += "## 🎉 CORRECT!  \n\n"
				enterMessage = "[Enter] Next Lesson"
			} else {
				result += "## ❌ WRONG...  \n\n"
				enterMessage = "[Enter] Retry"
			}

			result += "### AI Advice  \n"
			result += "> " + advice

			out, err = ui.RenderWithTerminalWidth(result)

			fmt.Print(out)

			fmt.Print(enterMessage)
			fmt.Scanln()

			if isCorrect {
				break
			}

			fmt.Print("\n")
		}
	}
	return nil
}

func judge(lesson course.Lesson, language, filePath string) (JudgeResult, error) {

	var judgeResult JudgeResult

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Running..."
	s.Start()
	defer s.Stop()

	output, err := run(language, filePath)
	s.Stop()
	if err != nil {
		return judgeResult, err
	}

	outputMd := "## Execution output\n"
	outputMd += "> " + output

	out, err := ui.RenderWithTerminalWidth(outputMd)
	fmt.Print(out)

	code, err := os.ReadFile(filePath)
	if err != nil {
		return judgeResult, err
	}

	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Judging..."
	s.Start()
	defer s.Stop()

	code_s := string(code)

	response, err := llm.GenerateJudgement(lesson.TaskDescription, code_s, output, lesson.CorrectOutput)
	s.Stop()
	if err != nil {
		return judgeResult, err
	}

	err = json.Unmarshal([]byte(response), &judgeResult)
	if err != nil {
		return judgeResult, err
	}

	return judgeResult, nil

}

func run(language, filePath string) (string, error) {
	cmd := exec.Command("")
	executable := true
	switch language {
	case "go":
		cmd = exec.Command("go", "run", filePath)
	case "py":
		cmd = exec.Command("python", filePath)
	default:
		executable = false
		switch language {
		case "html":
			browser.OpenFile(filePath)
		}
	}

	output_s := ""
	if executable {
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", err
		}
		output_s = string(output)
	} else {
		output_s = string(fmt.Sprint("no output with ", language))
	}

	return output_s, nil
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
