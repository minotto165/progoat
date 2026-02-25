/*
Copyright © 2026 minotto
*/
package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration and API keys",
	Long: `Set up your AI API keys and choose your preferred AI models. 
Settings are saved locally on your machine.`,
	Run: func(cmd *cobra.Command, args []string) {

		var confirm bool
		var provider string

		form := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Pick a provider").
					Options(
						huh.NewOption("Gemini", "gemini"),
						huh.NewOption("OpenAI", "openai"),
						huh.NewOption("Anthropic", "anthropic"),
					).
					Value(&provider),
			),
		).WithTheme(huh.ThemeBase())

		err := form.Run()
		if err != nil {
			fmt.Println("Configuration cancelled:", err)
			return
		}

		var apiKey string
		var model string

		var keyTitle string
		var options []huh.Option[string]

		switch provider {
		case "gemini":
			keyTitle = "Gemini API Key"
			options = []huh.Option[string]{
				huh.NewOption("Gemini 3 Pro Preview", "gemini-3-pro-preview"),
				huh.NewOption("Gemini 3 Flash Preview", "gemini-3-flash-preview"),
				huh.NewOption("Gemini Flash Latest", "gemini-flash-latest"),
				huh.NewOption("Gemini Flash Lite Latest", "gemini-flash-lite-latest"),
				huh.NewOption("Gemini 2.5 Pro", "gemini-2.5-pro"),
			}
		case "openai":
			keyTitle = "OpenAI API Key"
			options = []huh.Option[string]{
				huh.NewOption("GPT-5.2", "gpt-5.2"),
				huh.NewOption("GPT-5 mini", "gpt-5-mini"),
				huh.NewOption("GPT-5 nano", "gpt-5-nano"),
			}

		case "anthropic":
			keyTitle = "Anthropic API Key"
			options = []huh.Option[string]{
				huh.NewOption("Claude Opus 4.6", "claude-opus-4-6"),
				huh.NewOption("Claude Sonnet 4.5", "claude-sonnet-4-5"),
				huh.NewOption("Claude Haiku 4.5", "claude-haiku-4-5"),
			}
		}

		form = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(keyTitle).
					EchoMode(huh.EchoModePassword).
					Value(&apiKey),

				huh.NewSelect[string]().
					Title("AI model").
					Options(options...).
					Value(&model),

				huh.NewConfirm().
					Title("Save your config?").
					Affirmative("Yes(save)").
					Negative("No(Cancel)").
					Value(&confirm),
			),
		).WithTheme(huh.ThemeBase())

		err = form.Run()
		if err != nil {
			fmt.Println("Configuration cancelled:", err)
			return
		}

		if confirm {
			viper.Set("active_provider", provider)
			viper.Set(fmt.Sprintf("providers.%s.api_key", provider), apiKey)
			viper.Set(fmt.Sprintf("providers.%s.model", provider), model)
			err = viper.WriteConfig()
			if err != nil {
				viper.SafeWriteConfig()
			}

			fmt.Printf("[Config Changed] provider:%s, model:%s\n", provider, model)
		} else {
			fmt.Println("Configuration cancelled.")
		}

	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
