/*
Copyright © 2026 minotto
*/
package cmd

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/huh"
	"github.com/minotto165/progoat/internal/llm"
	"github.com/spf13/cobra"
)

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [topic]",
	Short: "Create a new course using AI",
	Long: `Generate a new learning course by providing a topic. 
AI will create lessons, including slides and coding exercises.`,
	Args:         cobra.MaximumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		length, _ := cmd.Flags().GetString("length")
		switch length {
		case "short", "medium", "long":
			break
		default:
			return fmt.Errorf("Invalid length. Options: short, medium, long.")

		}

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
				return fmt.Errorf("Cancelled.")
			}
		}

		fmt.Println("Input >", prompt)

		courseTitle := ""

		s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
		s.Suffix = " Generating..."
		s.Start()

		courseTitle, err := llm.GenerateCourse(prompt, length, coursesPath)
		if err != nil {
			return err
		}

		s.Stop()

		if courseTitle != "" {
			fmt.Println("Course generated:", courseTitle)
		}
		return nil
	},
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
