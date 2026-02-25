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
	chunks, errs := provider.CompletionStream(ctx, anyllm.CompletionParams{
		Model: activeModel,
		Messages: []anyllm.Message{
			{Role: anyllm.RoleUser, Content: prompt},
		},
	})

	for chunk := range chunks {
		if len(chunk.Choices) > 0 {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}
	if err := <-errs; err != nil {
		fmt.Printf("Error during streaming: %v\n", err)
		return
	}

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
