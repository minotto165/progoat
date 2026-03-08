package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
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
	activeModel := viper.GetString(fmt.Sprintf("providers.%s.gen_model", activeProvider))
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

	if len(response.Choices) == 0 || len(response.Choices[0].Message.ToolCalls) == 0 || response.Choices[0].Message.ToolCalls[0].Function.Arguments == "" {
		return "", fmt.Errorf("LLM returned an invalid or empty response")
	}
	return course.SaveCourse(response.Choices[0].Message.ToolCalls[0].Function.Arguments, coursesPath)

}

func GenerateJudgement(task, code, out, modelOut string) (string, error) {
	// Set informations
	activeProvider := viper.GetString("active_provider")
	activeModel := viper.GetString(fmt.Sprintf("providers.%s.judge_model", activeProvider))
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

	if len(response.Choices) == 0 || len(response.Choices[0].Message.ToolCalls) == 0 || response.Choices[0].Message.ToolCalls[0].Function.Arguments == "" {
		return "", fmt.Errorf("LLM returned an invalid or empty response")
	}
	return response.Choices[0].Message.ToolCalls[0].Function.Arguments, nil
}

//---------------------------
// ↓ VIBE-CODED
//---------------------------

type geminiModel struct {
	Name        string `json:"name"`        // e.g. "models/gemini-2.5-pro"
	DisplayName string `json:"displayName"` // e.g. "Gemini 2.5 Pro"
}

type geminiListModelsResponse struct {
	Models []geminiModel `json:"models"`
}

var reGeminiVersion = regexp.MustCompile(`^gemini-(\d+)(?:\.(\d+))?-`)

func isStandardGemini(id string) bool {
	if !strings.HasPrefix(id, "gemini-") {
		return false
	}
	excludes := []string{
		"tts", "audio", "image-generation", "embedding",
		"computer-use", "deep-research", "robotics",
		"nano", "custom",
	}
	lower := strings.ToLower(id)
	for _, kw := range excludes {
		if strings.Contains(lower, kw) {
			return false
		}
	}

	parts := strings.Split(id, "-")
	last := parts[len(parts)-1]
	if matched, _ := regexp.MatchString(`^\d{3}$`, last); matched {
		return false
	}
	return true
}

func geminiVersionKey(id string) string {
	if strings.Contains(id, "latest") {
		return "zzz-" + id
	}

	major, minor := "000", "000"
	if m := reGeminiVersion.FindStringSubmatch(id); m != nil {
		major = fmt.Sprintf("%03s", m[1])
		if m[2] != "" {
			minor = fmt.Sprintf("%03s", m[2])
		}
	}

	// preview 付きは同バージョン内で後ろ（"z" プレフィックス）
	preview := "a"
	if strings.Contains(id, "preview") {
		preview = "z"
	}

	return fmt.Sprintf("%s-%s-%s-%s", major, minor, preview, id)
}

func FetchGeminiModels(ctx context.Context, apiKey string) ([]huh.Option[string], error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ListModels returned HTTP %d", resp.StatusCode)
	}

	var result geminiListModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	type entry struct {
		id    string
		label string
	}
	var entries []entry

	for _, m := range result.Models {
		id := m.Name
		if len(id) > 7 && id[:7] == "models/" {
			id = id[7:]
		}
		if !isStandardGemini(id) {
			continue
		}
		label := m.DisplayName
		if label == "" {
			label = id
		}
		if !strings.HasPrefix(label, "Gemini") {
			continue
		}
		entries = append(entries, entry{id: id, label: label})
	}

	// バージョン昇順ソート（新しいものが下、latest は末尾）
	sort.Slice(entries, func(i, j int) bool {
		return geminiVersionKey(entries[j].id) < geminiVersionKey(entries[i].id)
	})

	var options []huh.Option[string]
	for _, e := range entries {
		options = append(options, huh.NewOption(e.label, e.id))
	}

	if len(options) == 0 {
		return nil, fmt.Errorf("no models returned")
	}
	return options, nil
}

//---------------------------
// ↑ VIBE-CODED
//---------------------------
