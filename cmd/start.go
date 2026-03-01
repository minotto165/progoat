/*
Copyright © 2026 minotto
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/eiannone/keyboard"
	tsize "github.com/kopoli/go-terminal-size"
	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers/anthropic"
	"github.com/mozilla-ai/any-llm-go/providers/gemini"
	"github.com/mozilla-ai/any-llm-go/providers/openai"
	"github.com/mozilla-ai/any-llm-go/providers/zai"
	"github.com/pkg/browser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		var courseID string

		if len(args) > 0 {
			courseID = args[0]
		} else {

			courses, err := getCourses()
			if err != nil {
				fmt.Println(err)
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
				fmt.Println(err)
			}
		}

		startCourse(courseID)

	},
}

func startCourse(courseID string) {

	course, err := getCourseStruct(courseID)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	coursePath := filepath.Join(coursesDir, course.ID)

	fmt.Println("[INFO] Course Directory:", coursePath)

	for _, l := range course.Lessons {

		clearScreen()
		slides := l.Slides
		defer keyboard.Close()
		for i, s := range slides {
			out, err := renderWithTerminalWidth(s)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			title := fmt.Sprint(course.Title, " - ", l.Title, ": page ", i+1)
			fmt.Println(title)
			fmt.Print(out)

			fmt.Print("[Enter] Next page")
			fmt.Scanln()
			fmt.Print("\033[1A\033[K")
			fmt.Print("\n\n\n")

		}

		lessonPath := filepath.Clean(filepath.Join(coursePath, l.ID))
		filePath := filepath.Clean(filepath.Join(lessonPath, l.FileName))
		task := fmt.Sprintf("%s\n%s\n\n**File to edit:**\n```text\n%s\n```",
			"## Task:",
			l.TaskDescription,
			filePath,
		)
		out, err := renderWithTerminalWidth(task)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		title := fmt.Sprint(course.Title, " - ", l.Title, ": task")
		fmt.Println(title)

		fmt.Print(out)
		for {

			fmt.Print("Edit and save the file, then hit Enter.")
			fmt.Scanln()

			fmt.Print("\033[1A\033[K")
			fmt.Print("\n\n")

			response, err := judge(l, course.Language, filePath)

			//for DEBUG...
			// response, err = JudgeResult{
			// 	IsCorrect: false,
			// 	Advice:    "TEMP",
			// }, nil

			if err != nil {
				fmt.Println("Error:", err)
				return
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

			out, err = renderWithTerminalWidth(result)

			fmt.Print(out)

			fmt.Print(enterMessage)
			fmt.Scanln()

			if isCorrect {
				break
			}

			fmt.Print("\n")
		}
	}
}

func judge(lesson Lesson, language, filePath string) (JudgeResult, error) {

	var judgeResult JudgeResult

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Running..."
	s.Start()
	defer s.Stop()

	output, err := run(language, filePath)
	if err != nil {
		return judgeResult, err
	}
	s.Stop()

	outputMd := "## Execution output\n"
	outputMd += "> " + output

	out, err := renderWithTerminalWidth(outputMd)
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

	response, err := generate_judgement(lesson.TaskDescription, code_s, output, lesson.CorrectOutput)
	if err != nil {
		return judgeResult, err
	}

	s.Stop()

	json.Unmarshal([]byte(response), &judgeResult)

	return judgeResult, nil

}

func generate_judgement(task, code, out, modelOut string) (string, error) {
	// Set informations
	activeProvider := viper.GetString("active_provider")
	activeModel := viper.GetString(fmt.Sprintf("providers.%s.model", activeProvider))
	activeApiKey := viper.GetString(fmt.Sprintf("providers.%s.api_key", activeProvider))

	// Set model
	var provider anyllm.Provider
	var err error

	switch activeProvider {
	case "openai":
		provider, err = openai.New(anyllm.WithAPIKey(activeApiKey))
	case "gemini":
		provider, err = gemini.New(anyllm.WithAPIKey(activeApiKey))
	case "anthropic":
		provider, err = anthropic.New(anyllm.WithAPIKey(activeApiKey))
	case "zai":
		provider, err = zai.New(anyllm.WithAPIKey(activeApiKey))
	default:
		return "", fmt.Errorf("Provider '%s' is not supported or not configured. Please run 'progoat config' first.\n", activeProvider)
	}

	if err != nil {
		return "", fmt.Errorf("Error during initializing model:%s", err)
	}

	// Generate!
	ctx := context.Background()
	response, err := provider.Completion(ctx, anyllm.CompletionParams{
		Model: activeModel,
		Messages: []anyllm.Message{
			{
				Role:    anyllm.RoleSystem,
				Content: `You are a programming instructor. Compare the student's code and output with the task and model answer. Check if the logic and the output meet the requirements. use Markdown.`,
			},
			{Role: anyllm.RoleUser, Content: "Task:" + task},
			{Role: anyllm.RoleUser, Content: "Model Output:" + modelOut},
			{Role: anyllm.RoleUser, Content: "Student Code:" + code},
			{Role: anyllm.RoleUser, Content: "Student Output:" + out},
		},
		Tools: []anyllm.Tool{
			{
				Type: "function",
				Function: anyllm.Function{
					Name: "judge_code",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"is_correct": map[string]any{"type": "boolean"},
							"advice":     map[string]any{"type": "string", "description": "Super-Short, helpful feedback in the student's language. Use Markdown but don't break a line."},
						},
						"required": []string{"is_correct", "advice"},
					},
				},
			},
		},
		ToolChoice: "required",
	})

	if err != nil {
		return "", err
	}

	if len(response.Choices[0].Message.ToolCalls[0].Function.Arguments) > 0 {
		return response.Choices[0].Message.ToolCalls[0].Function.Arguments, nil
	} else {
		return "", fmt.Errorf("LLM returned an invalid JSON.")
	}
}

func run(language, filePath string) (string, error) {
	cmd := exec.Command("")
	executable := 1
	switch language {
	case "go":
		cmd = exec.Command("go", "run", filePath)
	case "py":
		cmd = exec.Command("python", filePath)
	default:
		executable = 0
		switch language {
		case "html":
			browser.OpenFile(filePath)
		}
	}

	output_s := ""
	if executable == 1 {
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

func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		fmt.Print("\033[H\033[2J")
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

func renderWithTerminalWidth(raw string) (string, error) {
	s, err := tsize.GetSize()
	width := 0
	if err != nil {
		width = 80
		fmt.Println(err)
	} else {
		width = s.Width
	}

	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width-5),
	)

	out, err := r.Render(raw)
	if err != nil {
		return "", err
	}
	return out, nil
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
