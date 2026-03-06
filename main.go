package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	glamour "charm.land/glamour/v2"
	styles "charm.land/glamour/v2/styles"

	"charm.land/fang/v2"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type options struct {
	alt     bool
	noThink bool
	style   string
	wrap    int
}

type chunk struct {
	Choices []struct {
		Delta struct {
			Content          string `json:"content"`
			ReasoningContent string `json:"reasoning_content"`
		} `json:"delta"`
	} `json:"choices"`
}

type sseEvent struct {
	ReasoningContent string
	Content          string
	Done             bool
}

func main() {
	var opts options

	cmd := &cobra.Command{
		Use:   "streamd",
		Short: "Stream and render markdown from OpenAI-compatible endpoints",
		Long: `Streams SSE responses from OpenAI-compatible chat completion endpoints
and renders the markdown output in the terminal using glamour.

Supports thinking/reasoning tokens via the 'reasoning_content' field
and inline <think>...</think> tags.`,
		Example: `  curl -s https://api.example.com/v1/chat/completions \
    -H "Content-Type: application/json" \
    -d '{"model":"...","messages":[...],"stream":true}' | streamd

  # Alt-screen with scrollable viewport:
  curl -s ... | streamd --alt

  # Hide thinking:
  curl -s ... | streamd --no-think`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if term.IsTerminal(int(os.Stdin.Fd())) {
				return cmd.Help()
			}
			return run(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.alt, "alt", false, "interactive alt-screen with viewport scrolling")
	cmd.Flags().BoolVar(&opts.noThink, "no-think", false, "hide thinking/reasoning output")
	cmd.Flags().StringVar(&opts.style, "style", "", `glamour style: dark, light, dracula, tokyo-night, pink, ascii`)
	cmd.Flags().IntVarP(&opts.wrap, "wrap", "w", 0, "wrap width (0 = terminal width)")

	if err := fang.Execute(context.Background(), cmd); err != nil {
		os.Exit(1)
	}
}

func resolveStyle(s string) string {
	if s != "" {
		return s
	}
	if e := os.Getenv("GLAMOUR_STYLE"); e != "" {
		return e
	}
	return styles.DarkStyle
}

func termWidth(override int) int {
	if override > 0 {
		return override
	}
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	return 80
}

func run(opts options) error {
	style := resolveStyle(opts.style)
	width := termWidth(opts.wrap)
	showThink := !opts.noThink

	if opts.alt {
		return runTUI(style, width, showThink)
	}
	return runStream(style, width, showThink)
}

// runStream is the non-alt inline streaming mode.
func runStream(style string, width int, showThink bool) error {
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(style),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return err
	}

	var thinking, content strings.Builder
	var inThink bool
	prev := 0
	dirty := false
	last := time.Now()

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var c chunk
		if err := json.Unmarshal([]byte(data), &c); err != nil || len(c.Choices) == 0 {
			continue
		}
		delta := c.Choices[0].Delta

		if delta.ReasoningContent != "" {
			thinking.WriteString(delta.ReasoningContent)
			dirty = true
		}
		if delta.Content != "" {
			td, cd, next := parseContent(delta.Content, inThink)
			thinking.WriteString(td)
			content.WriteString(cd)
			inThink = next
			dirty = true
		}

		if dirty && time.Since(last) > 50*time.Millisecond {
			prev = renderFrame(r, thinking.String(), content.String(), showThink, prev)
			last = time.Now()
			dirty = false
		}
	}

	renderFrame(r, thinking.String(), content.String(), showThink, prev)
	fmt.Println()
	return scanner.Err()
}

// parseContent splits text into thinking/content deltas, handling <think> tags.
func parseContent(text string, inThink bool) (thinkDelta, contentDelta string, newInThink bool) {
	var tb, cb strings.Builder
	rest := text
	for len(rest) > 0 {
		if inThink {
			if i := strings.Index(rest, "</think>"); i >= 0 {
				tb.WriteString(rest[:i])
				rest = rest[i+8:]
				inThink = false
			} else {
				tb.WriteString(rest)
				rest = ""
			}
		} else {
			if i := strings.Index(rest, "<think>"); i >= 0 {
				cb.WriteString(rest[:i])
				rest = rest[i+7:]
				inThink = true
			} else {
				cb.WriteString(rest)
				rest = ""
			}
		}
	}
	return tb.String(), cb.String(), inThink
}

// renderFrame renders a full frame to stdout with synchronized output.
func renderFrame(r *glamour.TermRenderer, thinking, content string, showThink bool, prevLines int) int {
	md := buildMd(thinking, content, showThink)
	if md == "" {
		return prevLines
	}

	out, err := r.Render(md)
	if err != nil {
		out = md
	}

	var buf strings.Builder
	buf.WriteString("\033[?2026h")
	if prevLines > 0 {
		fmt.Fprintf(&buf, "\033[%dA\033[J", prevLines)
	}
	buf.WriteString(out)
	buf.WriteString("\033[?2026l")
	_, _ = os.Stdout.WriteString(buf.String())
	return strings.Count(out, "\n")
}

// buildMd composes the final markdown from thinking + content.
func buildMd(thinking, content string, showThink bool) string {
	var sb strings.Builder
	if showThink && thinking != "" {
		sb.WriteString("*thinking...*\n\n")
		for line := range strings.SplitSeq(strings.TrimRight(thinking, "\n"), "\n") {
			sb.WriteString("> ")
			sb.WriteString(line)
			sb.WriteString("\n")
		}
		if content != "" {
			sb.WriteString("\n---\n\n")
		}
	}
	sb.WriteString(content)
	return sb.String()
}

// readSSE reads SSE events from stdin and sends them to ch.
func readSSE(ch chan<- sseEvent) {
	defer close(ch)
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			ch <- sseEvent{Done: true}
			return
		}

		var c chunk
		if err := json.Unmarshal([]byte(data), &c); err != nil || len(c.Choices) == 0 {
			continue
		}
		delta := c.Choices[0].Delta
		ch <- sseEvent{
			ReasoningContent: delta.ReasoningContent,
			Content:          delta.Content,
		}
	}
	ch <- sseEvent{Done: true}
}
