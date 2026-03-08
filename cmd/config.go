/*
Copyright © 2026 minotto
*/
package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/minotto165/progoat/internal/llm"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration and API keys",
	Long: `Set up your AI API keys and choose your preferred AI models. 
Settings are saved locally on your machine.`,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		var provider string
		var apiKey string

		// プロバイダ選択
		err := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Pick a provider").
					Options(
						huh.NewOption("Gemini (auto-fetches latest models)", "gemini"),
						huh.NewOption("OpenAI", "openai"),
						huh.NewOption("Anthropic", "anthropic"),
						huh.NewOption("Z.AI", "zai"),
					).
					Value(&provider),
			),
		).WithTheme(huh.ThemeBase()).Run()

		if err != nil {
			return fmt.Errorf("Cancelled: %w", err)
		}

		var keyTitle string
		switch provider {
		case "gemini":
			keyTitle = "Gemini API Key"
		case "openai":
			keyTitle = "OpenAI API Key"
		case "anthropic":
			keyTitle = "Anthropic API Key"
		case "zai":
			keyTitle = "Z.AI API Key"
		}

		// 設定読み込み
		apiKey = viper.GetString(fmt.Sprintf("providers.%s.api_key", provider))

		// APIキー入力
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(keyTitle).
					EchoMode(huh.EchoModePassword).
					Value(&apiKey),
			),
		).WithTheme(huh.ThemeBase()).Run()
		if err != nil {
			return fmt.Errorf("Cancelled: %w", err)
		}

		// フォールバック
		var geminiDefaultOptions = []huh.Option[string]{
			huh.NewOption("Gemini 3.1 Pro Preview", "gemini-3.1-pro-preview"),
			huh.NewOption("Gemini 3 Flash Preview", "gemini-3-flash-preview"),
			huh.NewOption("Gemini Flash Latest", "gemini-flash-latest"),
			huh.NewOption("Gemini Flash Lite Latest", "gemini-flash-lite-latest"),
			huh.NewOption("Gemini 2.5 Pro", "gemini-2.5-pro"),
		}

		var options []huh.Option[string]
		switch provider {
		case "gemini":
			keyForFetch := apiKey
			if keyForFetch == "" {
				keyForFetch = viper.GetString("providers.gemini.api_key")
			}
			if keyForFetch != "" {
				if fetched, err := llm.FetchGeminiModels(cmd.Context(), keyForFetch); err == nil {
					options = fetched
					fmt.Println("Fetched latest Gemini models")
				} else {
					fmt.Printf("Could not fetch Gemini models (%s), using defaults\n", err)
					options = geminiDefaultOptions
				}
			} else {
				options = geminiDefaultOptions
			}
		case "openai":
			options = []huh.Option[string]{
				huh.NewOption("GPT-5.4", "gpt-5.4"),
				huh.NewOption("GPT-5 mini", "gpt-5-mini"),
				huh.NewOption("GPT-5 nano", "gpt-5-nano"),
			}
		case "anthropic":
			options = []huh.Option[string]{
				huh.NewOption("Claude Opus 4.6", "claude-opus-4-6"),
				huh.NewOption("Claude Sonnet 4.6", "claude-sonnet-4-6"),
				huh.NewOption("Claude Haiku 4.5", "claude-haiku-4-5-20251001"),
			}
		case "zai":
			options = []huh.Option[string]{
				huh.NewOption("GLM-5", "glm-5"),
				huh.NewOption("GLM-4.7", "glm-4.7"),
				huh.NewOption("GLM-4.7-FlashX", "glm-4.7-flashx"),
				huh.NewOption("GLM-4.7-Flash", "glm-4.7-flash"),
			}
		}

		// 設定読み込み
		genModel := viper.GetString(fmt.Sprintf("providers.%s.gen_model", provider))
		judgeModel := viper.GetString(fmt.Sprintf("providers.%s.judge_model", provider))

		var confirm bool

		// 詳細設定
		err = huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Model for course generation").
					Options(options...).
					Value(&genModel),

				huh.NewSelect[string]().
					Title("Model for judging").
					Options(options...).
					Value(&judgeModel),

				huh.NewConfirm().
					Title("Save settings?").
					Affirmative("Save").
					Negative("Cancel").
					Value(&confirm),
			),
		).WithTheme(huh.ThemeBase()).Run()

		if err != nil {
			return fmt.Errorf("Cancelled: %w", err)
		}

		// 保存
		if confirm {
			viper.Set("active_provider", provider)
			if apiKey != "" {
				viper.Set(fmt.Sprintf("providers.%s.api_key", provider), apiKey)
			}
			viper.Set(fmt.Sprintf("providers.%s.gen_model", provider), genModel)
			viper.Set(fmt.Sprintf("providers.%s.judge_model", provider), judgeModel)

			// 設定ファイルがなければ作成、あれば上書き
			if err := viper.WriteConfig(); err != nil {
				viper.SafeWriteConfig()
			}

			fmt.Printf("[Config Changed] provider:%s, gen_model:%s, judge_model:%s\n", provider, genModel, judgeModel)
		} else {
			fmt.Println("Configuration cancelled.")
		}
		return nil
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
