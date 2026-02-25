/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers/anthropic"
	"github.com/mozilla-ai/any-llm-go/providers/gemini"
	"github.com/mozilla-ai/any-llm-go/providers/openai"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Create a new course using AI",
	Long: `Generate a new learning course by providing a topic. 
AI will create lessons, including slides and coding exercises.`,
	Run: func(cmd *cobra.Command, args []string) {
		var prompt string
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
		fmt.Println("Input >", prompt)
		generation(prompt)
	},
}

func generation(prompt string) {
	activeProvider := viper.GetString("active_provider")
	activeModel := viper.GetString(fmt.Sprintf("providers.%s.model", activeProvider))
	activeApiKey := viper.GetString(fmt.Sprintf("providers.%s.api_key", activeProvider))

	var provider anyllm.Provider
	var err error

	switch activeProvider {
	case "openai":
		provider, err = openai.New(anyllm.WithAPIKey(activeApiKey))
	case "gemini":
		provider, err = gemini.New(anyllm.WithAPIKey(activeApiKey))
	case "anthropic":
		provider, err = anthropic.New(anyllm.WithAPIKey(activeApiKey))
	default:
		fmt.Printf("Error: Provider '%s' is not supported or not configured.\n", activeProvider)
		fmt.Println("Please run 'progoat config' first.")
		return
	}

	if err != nil {
		fmt.Println("Error initializing model:", err)
		return
	}
	ctx := context.Background()
	responce, err := provider.Completion(ctx, anyllm.CompletionParams{
		Model: activeModel,
		Messages: []anyllm.Message{
			{
				Role: anyllm.RoleSystem,
				Content: `You are a professional coding instructor. Your task is to generate a structured programming course based on the user's topic. 
Strictly follow these language requirements:
1. Use the same language as the user's prompt for the following fields: "description", "task_description", "title", "slides", and any comments within "initial_code".
2. Use English for all other fields, technical identifiers, and metadata to ensure system compatibility.
3. In "initial_code", provide the actual source code in the target programming language, but ensure all explanatory comments are in the user's language.`,
			},
			{Role: anyllm.RoleUser, Content: prompt},
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
							"language":    map[string]any{"type": "string", "description": "e.g., go, python, javascript"},
							"lessons": map[string]any{
								"type": "array",
								"items": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"lesson_id":        map[string]any{"type": "string"},
										"title":            map[string]any{"type": "string"},
										"slides":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
										"task_description": map[string]any{"type": "string"},
										"initial_code":     map[string]any{"type": "string", "description": "The boilerplate code for the student to start with."},
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
		return
	}

	fmt.Println(responce)
}

func init() {
	rootCmd.AddCommand(generateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// generateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// generateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
