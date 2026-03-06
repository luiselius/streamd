<div align="center">
  <h1>streamd</h1>
  <p>A CLI tool that streams OpenAI-compatible chat completion responses and renders markdown in the terminal using glamour.</p>

  <a href="https://github.com/Gaurav-Gosain/streamd/releases"><img src="https://img.shields.io/github/release/Gaurav-Gosain/streamd.svg" alt="Latest Release"></a>
  <a href="https://pkg.go.dev/github.com/Gaurav-Gosain/streamd?tab=doc"><img src="https://godoc.org/github.com/Gaurav-Gosain/streamd?status.svg" alt="GoDoc"></a>
</div>

---

streamd takes streaming SSE output from any OpenAI-compatible endpoint (piped via `curl` or similar) and renders it as beautifully formatted markdown in the terminal. It supports reasoning/thinking tokens and offers both an inline streaming mode and an interactive alt-screen mode with scrolling.

<details>
<summary>Table of Contents</summary>

- [Installation](#installation)
- [Usage](#usage)
- [Features](#features)
- [Modes](#modes)
- [Thinking / Reasoning Support](#thinking--reasoning-support)
- [Development](#development)
- [License](#license)

</details>

## Installation

### Package Managers

**Homebrew (macOS/Linux):**
```bash
brew tap Gaurav-Gosain/tap
brew install streamd
```

### Other Methods

- **[GitHub Releases](https://github.com/Gaurav-Gosain/streamd/releases)** - Download pre-built binaries
- **Go Install:** `go install github.com/Gaurav-Gosain/streamd@latest`
- **Build from Source:** See [Development](#development) below

**Requirements:**
- A terminal with true color support (most modern terminals work fine)
- Go 1.25+ (if building from source)

## Usage

```bash
# Stream a chat completion with rendered markdown
curl -s https://api.example.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{"model":"...","messages":[{"role":"user","content":"hello"}],"stream":true}' | streamd

# Alt-screen mode with scrollable viewport
curl -s ... | streamd --alt

# Hide thinking/reasoning output
curl -s ... | streamd --no-think

# Use a different glamour theme
curl -s ... | streamd --style dracula
```

### Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--alt` | | `false` | Interactive alt-screen with viewport scrolling |
| `--no-think` | | `false` | Hide thinking/reasoning output |
| `--style` | | `dark` | Glamour style: `dark`, `light`, `dracula`, `tokyo-night`, `pink`, `ascii` |
| `--wrap` | `-w` | `0` | Word wrap width (0 = terminal width) |

The style can also be set via the `GLAMOUR_STYLE` environment variable.

## Features

- **Live markdown rendering** using [glamour](https://github.com/charmbracelet/glamour) (v2-exp) with syntax-highlighted code blocks, styled headings, lists, tables, and more
- **Flicker-free streaming** via synchronized terminal output (`DECSYNC`) and debounced re-rendering
- **Thinking/reasoning support** for models that expose chain-of-thought, via the `reasoning_content` SSE field or inline `<think>...</think>` tags
- **Interactive alt-screen mode** powered by [bubbletea](https://github.com/charmbracelet/bubbletea) with a scrollable viewport, scroll progress bar, and keyboard navigation
- **Multiple themes** including dark, light, dracula, tokyo-night, and pink
- **Auto-detected terminal width** for proper word wrapping
- **Styled CLI help** via [fang](https://github.com/charmbracelet/fang)

## Modes

### Inline Mode (default)

Content is rendered directly in the terminal as it streams. The output uses ANSI cursor control with synchronized output to re-render in place without flicker. Best for quick queries where you want the output to remain in your scrollback.

### Alt-Screen Mode (`--alt`)

Opens a full-screen interactive viewport. The stream auto-scrolls to follow new content. Once streaming is complete (or at any point), you can scroll through the output and quit when ready.

#### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j` / `k` | Scroll down / up |
| `d` / `u` | Half-page down / up |
| `pgdn` / `pgup` | Page down / up |
| `g` / `G` | Go to top / bottom |
| `q` / `esc` / `ctrl+c` | Quit |

The viewport auto-scrolls to follow the stream. Scrolling up pauses auto-scroll; pressing `G` resumes it.

## Thinking / Reasoning Support

streamd supports models that expose their reasoning process. It handles two common patterns:

1. **`reasoning_content` field** (OpenAI-compatible) - Reasoning tokens arrive in a separate `reasoning_content` field in the SSE delta, used by models like Qwen, DeepSeek, and others.

2. **`<think>` tags** - Some models wrap their reasoning in `<think>...</think>` tags within the regular `content` field.

Both patterns are detected automatically. Thinking content is rendered as an italic header followed by a blockquote, separated from the main response by a horizontal rule.

Use `--no-think` to hide reasoning output entirely.

## Development

Contributions are welcome. Feel free to open issues or pull requests.

**Build from source:**
```bash
git clone https://github.com/Gaurav-Gosain/streamd.git
cd streamd
go build -o streamd .
./streamd --help
```

### Dependencies

streamd is built on the [Charm](https://charm.sh) ecosystem:

| Library | Purpose |
|---------|---------|
| [glamour v2](https://github.com/charmbracelet/glamour) | Markdown rendering |
| [bubbletea v2](https://github.com/charmbracelet/bubbletea) | TUI framework (alt-screen mode) |
| [bubbles v2](https://github.com/charmbracelet/bubbles) | Viewport component |
| [lipgloss v2](https://github.com/charmbracelet/lipgloss) | Terminal styling |
| [fang](https://github.com/charmbracelet/fang) | Styled CLI help |

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=Gaurav-Gosain/streamd&type=Date&theme=dark)](https://star-history.com/#Gaurav-Gosain/streamd&Date)

<p style="display:flex;flex-wrap:wrap;">
<img alt="GitHub Language Count" src="https://img.shields.io/github/languages/count/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
<img alt="GitHub Top Language" src="https://img.shields.io/github/languages/top/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
<img alt="Repo Size" src="https://img.shields.io/github/repo-size/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
<img alt="GitHub Issues" src="https://img.shields.io/github/issues/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
<img alt="GitHub Closed Issues" src="https://img.shields.io/github/issues-closed/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
<img alt="GitHub Pull Requests" src="https://img.shields.io/github/issues-pr/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
<img alt="GitHub Closed Pull Requests" src="https://img.shields.io/github/issues-pr-closed/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
<img alt="GitHub Contributors" src="https://img.shields.io/github/contributors/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
<img alt="GitHub Last Commit" src="https://img.shields.io/github/last-commit/Gaurav-Gosain/streamd" style="padding:5px;margin:5px;" />
</p>

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
