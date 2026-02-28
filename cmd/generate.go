/*
Copyright © 2026 minotto
*/
package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers/anthropic"
	"github.com/mozilla-ai/any-llm-go/providers/gemini"
	"github.com/mozilla-ai/any-llm-go/providers/openai"
	"github.com/mozilla-ai/any-llm-go/providers/zai"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [topic]",
	Short: "Create a new course using AI",
	Long: `Generate a new learning course by providing a topic. 
AI will create lessons, including slides and coding exercises.`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		length, _ := cmd.Flags().GetString("length")

		var prompt string

		if len(args) > 0 {
			prompt = args[0]
		} else {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewText().
						Title("Prompt").
						Description("Enter what you want to learn...").
						Value(&prompt),
				),
			).WithTheme(huh.ThemeBase())
			err := form.Run()
			if err != nil {
				fmt.Println("Canceled.")
				return
			}
		}

		fmt.Println("Input >", prompt)

		courseTitle := ""

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Generating..."
		s.Start()

		courseTitle = generation(prompt, length) // chanが閉じるまで待つ

		s.Stop()

		if courseTitle != "" {
			fmt.Println("Course generated:", courseTitle)
		}
	},
}

func generation(prompt, length string) string {

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
		fmt.Printf("Error: Provider '%s' is not supported or not configured.\n", activeProvider)
		fmt.Println("Please run 'progoat config' first.")
		return ""
	}

	if err != nil {
		fmt.Println("Error initializing model:", err)
		return ""
	}

	// Generate!
	ctx := context.Background()
	response, err := provider.Completion(ctx, anyllm.CompletionParams{
		Model: activeModel,
		Messages: []anyllm.Message{
			{
				Role: anyllm.RoleSystem,
				Content: `You are a professional coding instructor. Your task is to generate a structured programming course based on the user's topic. 
Strictly follow these language requirements:
1. Use the same language as the user's prompt for the following fields: "description", "task_description", "title", "slides", and any comments within "initial_code".
2. Use English for all other fields, technical identifiers, and metadata to ensure system compatibility.
3. In "initial_code", provide the actual source code in the target programming language, but ensure all explanatory comments are in the user's language.
4. Use markdown for the slides to make them easy to read.
5. Course ID should be short and super simple.
6. Course Title should be simple.`,
			},
			{Role: anyllm.RoleUser, Content: prompt},
			{Role: anyllm.RoleUser, Content: "Course length(short,medium,long):" + length},
		},
		Tools: []anyllm.Tool{
			{
				Type: "function",
				Function: anyllm.Function{
					Name: "generate_course_data",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"course_id":   map[string]any{"type": "string"},
							"title":       map[string]any{"type": "string"},
							"description": map[string]any{"type": "string"},
							"language":    map[string]any{"type": "string", "description": "The extension of the created code file(e.g., go, py, js)"},
							"lessons": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"lesson_id": map[string]any{"type": "string"},
										"title":     map[string]any{"type": "string"},
										"slides": map[string]any{
											"type":  "array",
											"items": map[string]any{"type": "string"},
											"description": "An array of markdown strings, where each element is a single slide page. " +
												"Follow these rules: " +
												"1. Use '##' for headers to define the start of a new slide content. " +
												"2. For language requirements, put the English term first, followed by the Japanese translation in parentheses, like 'Function (ja:スライド)'. " +
												"3. Do not include page numbers in the markdown string itself."},
										"task_description": map[string]any{"type": "string"},
										"initial_code":     map[string]any{"type": "string", "description": "The boilerplate code for the student to start with."},
										"file_name":        map[string]any{"type": "string", "description": "The name of code file (e.g., main.go, index.html)"},
									},
									"required": []string{"lesson_id", "title", "slides", "task_description", "initial_code"},
								},
							},
						},
						"required": []string{"course_id", "title", "description", "language", "lessons"},
					},
				},
			},
		},
		ToolChoice: "required",
	})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return ""
	}

	if len(response.Choices[0].Message.ToolCalls[0].Function.Arguments) > 0 {

		return saveCourse(response.Choices[0].Message.ToolCalls[0].Function.Arguments)
	} else {
		fmt.Println("Error: LLM returned an invalid JSON.")
		return ""
	}

}

func saveCourse(response string) string {

	// JSON to struct
	var course Course
	err := json.Unmarshal([]byte(response), &course)
	if err != nil {
		fmt.Println("Error parsing JSON:", err)
		return ""
	}

	// Crate course directory
	coursePath := filepath.Join(coursesDir, course.ID)
	os.MkdirAll(coursePath, 0755)

	// Update courses.json
	coursesJsonPath := filepath.Join(coursePath, "course.json")
	coursesJson, err := json.Marshal(course) // Convert to string(JSON)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return ""
	}
	coursesJson = []byte(coursesJson) // Convert to []byte

	os.WriteFile(coursesJsonPath, coursesJson, 0755)

	// Create lessons direcotries
	for _, lesson := range course.Lessons {
		lessonPath := filepath.Join(coursePath, lesson.ID)
		os.MkdirAll(lessonPath, 0755)

		// Create slides
		slides := lesson.Slides
		slidesContent, err := json.Marshal(slides)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return ""
		}

		// Write Files
		os.WriteFile(filepath.Join(lessonPath, "slide.json"), slidesContent, 0644)
		os.WriteFile(filepath.Join(lessonPath, "task.md"), []byte(lesson.TaskDescription), 0644)

		ext := course.Language
		os.WriteFile(filepath.Join(lessonPath, "main."+ext), []byte(lesson.InitialCode), 0644)

	}
	return course.Title

}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	generateCmd.Flags().StringP("length", "l", "medium", "Course length (short, medium, long)")
}
