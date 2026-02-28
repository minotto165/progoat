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
	"github.com/eiannone/keyboard"
	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers/anthropic"
	"github.com/mozilla-ai/any-llm-go/providers/gemini"
	"github.com/mozilla-ai/any-llm-go/providers/openai"
	"github.com/mozilla-ai/any-llm-go/providers/zai"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		out, err := glamour.Render(task, "dark")
		if err != nil {
			return
		}

		title := fmt.Sprint(course.Title, ": task")
		fmt.Println(title)

		fmt.Print(out)

		fmt.Print("Edit and save the file, then hit Enter.")
		fmt.Scanln()

		err = judge(l, course.Language, filePath)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Scanln()

	}
}

func judge(lesson Lesson, language, filePath string) error {

	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Running..."
	s.Start()
	defer s.Stop()

	output, err := run(language, filePath)
	if err != nil {
		return err
	}
	s.Stop()

	fmt.Println("Output:")
	fmt.Println("  ", output)

	code, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	s = spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = " Judging..."
	s.Start()
	defer s.Stop()

	code_s := string(code)

	response, err := generate(lesson.TaskDescription, code_s, output, lesson.CorrectOutput)
	if err != nil {
		return err
	}

	s.Stop()

	fmt.Println(response)

	return nil

}

func generate(task, code, out, modelOut string) (string, error) {
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
							"advice":     map[string]any{"type": "string", "description": "Short, helpful feedback in the student's language"},
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
	switch language {
	case "go":
		cmd = exec.Command("go", "run", filePath)
	case "py":
		cmd = exec.Command("python", filePath)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	output_s := string(output)
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
