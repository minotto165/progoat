# Progoat 🐐

> [!TIP]
> [日本語のREADMEはこちら](./README.ja.md)

> [!NOTE]
> This project is currently under development.

Progoat is an LLM-powered CLI tool for programming education, inspired by Progate. Generate structured programming courses on any topic, read slides, and solve coding exercises directly from your terminal.

![License](https://img.shields.io/github/license/minotto165/progoat)
![Go Version](https://img.shields.io/github/go-mod/go-version/minotto165/progoat)

## ✨ Features

- **AI-Powered Course Generation**: Create a full course on any programming topic just by providing a prompt.
- **Interactive CLI UI**: Built with [huh](https://github.com/charmbracelet/huh) for a beautiful and smooth user experience.
- **Multi-Provider Support**: Supports OpenAI (GPT-4), Google Gemini, and Anthropic Claude.
- **Hands-on Learning**: Each lesson includes slides, a task description, and boilerplate code to get you started.
- **Local Storage**: Courses are saved locally in `~/.progoat`, so you can learn offline once generated.

## 🚀 Getting Started

### Installation

Make sure you have [Go](https://golang.org/dl/) installed.

```bash
go install github.com/minotto165/progoat@latest
```

Alternatively, clone the repository and build it manually:

```bash
git clone https://github.com/minotto165/progoat.git
cd progoat
go build -o progoat .
```

### Configuration

Before generating courses, you need to set up your API keys.

```bash
progoat config
```
This will open an interactive form where you can choose your LLM provider and enter your API key.

## 📖 Usage

### 1. Generate a Course
Tell the AI what you want to learn.
```bash
progoat generate
```
Example prompt: *"I want to learn the basics of Go concurrency with goroutines and channels."*

### 2. List Your Courses (WIP)
See all the courses you have generated.
```bash
progoat list
```

### 3. Start Learning (WIP)
Begin a learning session for a specific course.
```bash
progoat start
```

### 4. Check Progress (WIP)
Check how far you've come.
```bash
progoat status
```

## 🛠️ Development

If you want to contribute or modify the tool:

1. Clone the repo.
2. Install dependencies: `go mod download`
3. Run the CLI: `go run main.go`

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.

---

Made with ❤️ by [minotto](https://github.com/minotto165)
