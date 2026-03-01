package llm

import (
	"context"
	"fmt"

	"github.com/minotto165/progoat/internal/course"
	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers/anthropic"
	"github.com/mozilla-ai/any-llm-go/providers/gemini"
	"github.com/mozilla-ai/any-llm-go/providers/openai"
	"github.com/mozilla-ai/any-llm-go/providers/zai"
	"github.com/spf13/viper"
)

func GenerateCourse(prompt, length, coursesPath string) (string, error) {

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
				Role: anyllm.RoleSystem,
				Content: `You are a professional coding instructor. Your task is to generate a structured programming course based on the user's topic. 
Strictly follow these language requirements:
1. Use the same language as the user's prompt for the following fields: "description", "task_description", "title", "slides", and any comments within "initial_code".
2. Use English for all other fields, technical identifiers, and metadata to ensure system compatibility.
3. In "initial_code", provide the actual source code in the target programming language, but ensure all explanatory comments are in the user's language.
4. Use markdown for the slides to make them easy to read.
5. Course ID should be short and super simple.
6. Course Title should be simple.
7. The first slide of the first lesson MUST be a "Setup Guide". It should explain how to install the necessary environment for the language and how to run the code on a local machine.`,
			},
			{
				Role:    anyllm.RoleUser,
				Content: fmt.Sprintf("Topic: \"\"\"\n%s\n\"\"\"", prompt),
			},
			{
				Role:    anyllm.RoleUser,
				Content: fmt.Sprintf("Course length: %s", length),
			},
		},
		Tools: []anyllm.Tool{
			{
				Type: "function",
				Function: anyllm.Function{
					Name: "generate_course_data",
					Parameters: map[string]any{
						"type": "object",
						"properties": map[string]any{
							"course_id":            map[string]any{"type": "string"},
							"title":                map[string]any{"type": "string"},
							"description":          map[string]any{"type": "string"},
							"programming_language": map[string]any{"type": "string", "description": "The extension of the created code file(e.g., go, py, js),NOT NATURAL LANGUAGE(ja,en...)"},
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
												"2. Write naturally in the student's language (the language used in the prompt). " +
												"3. Do not include page numbers in the markdown string itself." +
												"4. The VERY FIRST slide of the FIRST lesson must be a 'Local Setup Guide' for the programming language (e.g., installation, run commands)."},
										"task_description": map[string]any{"type": "string"},
										"initial_code":     map[string]any{"type": "string", "description": "The boilerplate code for the student to start with."},
										"correct_output":   map[string]any{"type": "string", "description": "The expected standard output (stdout) when the task is correctly implemented."},
										"file_name":        map[string]any{"type": "string", "description": "The name of code file (e.g., main.go, index.html)"},
									},
									"required": []string{"lesson_id", "title", "slides", "task_description", "initial_code", "correct_output"},
								},
							},
						},
						"required": []string{"course_id", "title", "description", "programming_language", "lessons"},
					},
				},
			},
		},
		ToolChoice: "required",
	})

	if err != nil {
		return "", err
	}

	if len(response.Choices) > 0 && len(response.Choices[0].Message.ToolCalls) > 0 && len(response.Choices[0].Message.ToolCalls[0].Function.Arguments) > 0 {

		return course.SaveCourse(response.Choices[0].Message.ToolCalls[0].Function.Arguments, coursesPath)
	} else {
		return "", fmt.Errorf("LLM returned an invalid JSON.")
	}

}

func GenerateJudgement(task, code, out, modelOut string) (string, error) {
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
		return "", fmt.Errorf("failed to initialize model:%w", err)
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

	if len(response.Choices) > 0 && len(response.Choices[0].Message.ToolCalls) > 0 && len(response.Choices[0].Message.ToolCalls[0].Function.Arguments) > 0 {
		return response.Choices[0].Message.ToolCalls[0].Function.Arguments, nil
	} else {
		return "", fmt.Errorf("LLM returned an invalid JSON.")
	}
}
