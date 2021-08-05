package main

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_markdown(t *testing.T) {
	t.Run("Basic Markdown tests", func(t *testing.T) {
		app := &goBlog{
			cfg: &config{
				Server: &configServer{
					PublicAddress: "https://example.com",
				},
			},
		}

		app.initMarkdown()

		// Relative / absolute links

		rendered, err := app.renderMarkdown("[Relative](/relative)", false)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !strings.Contains(string(rendered), `href="/relative"`) {
			t.Errorf("Wrong result, got %v", string(rendered))
		}

		rendered, err = app.renderMarkdown("[Relative](/relative)", true)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !strings.Contains(string(rendered), `href="https://example.com/relative"`) {
			t.Errorf("Wrong result, got %v", string(rendered))
		}
		if strings.Contains(string(rendered), `target="_blank"`) {
			t.Errorf("Wrong result, got %v", string(rendered))
		}

		// External links

		rendered, err = app.renderMarkdown("[External](https://example.com)", true)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !strings.Contains(string(rendered), `target="_blank"`) {
			t.Errorf("Wrong result, got %v", string(rendered))
		}

		// Link title

		rendered, err = app.renderMarkdown(`[With title](https://example.com "Test-Title")`, true)
		if err != nil {
			t.Fatalf("Error: %v", err)
		}
		if !strings.Contains(string(rendered), `title="Test-Title"`) {
			t.Errorf("Wrong result, got %v", string(rendered))
		}

		// Text

		renderedText := app.renderText("**This** *is* [text](/)")
		if renderedText != "This is text" {
			t.Errorf("Wrong result, got \"%v\"", renderedText)
		}

		// Title

		assert.Equal(t, "3. **Test**", app.renderMdTitle("3. **Test**"))
		assert.Equal(t, "Test’s", app.renderMdTitle("Test's"))
		assert.Equal(t, "😂", app.renderMdTitle(":joy:"))
		assert.Equal(t, "<b></b>", app.renderMdTitle("<b></b>"))

		// Template func

		renderedText = string(app.safeRenderMarkdownAsHTML("[Relative](/relative)"))
		assert.Contains(t, renderedText, `href="/relative"`)
	})
}

func Benchmark_markdown(b *testing.B) {
	markdownExample, err := os.ReadFile("testdata/markdownexample.md")
	if err != nil {
		b.Errorf("Failed to read markdown example: %v", err)
	}
	mdExp := string(markdownExample)

	app := &goBlog{
		cfg: &config{
			Server: &configServer{
				PublicAddress: "https://example.com",
			},
		},
	}

	app.initMarkdown()

	b.Run("Benchmark Markdown Rendering", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := app.renderMarkdown(mdExp, true)
			if err != nil {
				b.Errorf("Error: %v", err)
			}
		}
	})
}
