package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderListSimpleUnordered tests rendering of a simple unordered list
func TestRenderListSimpleUnordered(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	content := `<li>Item 1</li><li>Item 2</li><li>Item 3</li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
	// The actual PDF rendering is tested by visual inspection
}

// TestRenderListSimpleOrdered tests rendering of a simple ordered list
func TestRenderListSimpleOrdered(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	content := `<li>First item</li><li>Second item</li><li>Third item</li>`
	renderer.renderList(content, true)

	// If we get here without panic, the test passes
}

// TestRenderListNestedUnordered tests rendering of nested unordered lists
func TestRenderListNestedUnordered(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// HTML structure for nested unordered list
	content := `<li>Level 1 Item 1<ul><li>Level 2 Item 1a</li><li>Level 2 Item 1b</li></ul></li><li>Level 1 Item 2</li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListNestedOrdered tests rendering of nested ordered lists
func TestRenderListNestedOrdered(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// HTML structure for nested ordered list
	content := `<li>First<ol><li>Sub-first</li><li>Sub-second</li></ol></li><li>Second</li>`
	renderer.renderList(content, true)

	// If we get here without panic, the test passes
}

// TestRenderListDeeplyNested tests rendering of deeply nested lists (3+ levels)
func TestRenderListDeeplyNested(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// HTML structure for 3-level deep nested list
	content := `<li>Level 1<ul><li>Level 2<ul><li>Level 3 item</li></ul></li></ul></li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListMixedNesting tests rendering of mixed ordered/unordered nested lists
func TestRenderListMixedNesting(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// Unordered list with nested ordered list
	content := `<li>Unordered item<ol><li>Ordered sub-item 1</li><li>Ordered sub-item 2</li></ol></li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListEmptyItems tests handling of empty list items
func TestRenderListEmptyItems(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	content := `<li></li><li>Non-empty</li><li></li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListWithInlineFormatting tests list items with inline HTML tags
func TestRenderListWithInlineFormatting(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	content := `<li><strong>Bold</strong> text</li><li><em>Italic</em> text</li><li><code>code</code> text</li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListWithHTMLEntities tests list items with HTML entities
func TestRenderListWithHTMLEntities(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	content := `<li>Less than &lt; greater than &gt;</li><li>Ampersand &amp; quote &quot;</li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestConvertMarkdownToHTMLNestedList tests that goldmark correctly converts nested markdown lists to HTML
func TestConvertMarkdownToHTMLNestedList(t *testing.T) {
	markdown := `- Item 1
  - Sub-item 1a
  - Sub-item 1b
- Item 2`

	html, err := convertMarkdownToHTML([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	// Check that nested structure is preserved
	if !strings.Contains(html, "<ul>") {
		t.Error("Expected HTML to contain <ul> tag")
	}
	if !strings.Contains(html, "<li>") {
		t.Error("Expected HTML to contain <li> tag")
	}
	// Nested list should have ul inside li
	if !strings.Contains(html, "</li>") {
		t.Error("Expected HTML to contain closing </li> tag")
	}
}

// TestConvertMarkdownToHTMLOrderedList tests ordered list conversion
func TestConvertMarkdownToHTMLOrderedList(t *testing.T) {
	markdown := `1. First
2. Second
3. Third`

	html, err := convertMarkdownToHTML([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	if !strings.Contains(html, "<ol>") {
		t.Error("Expected HTML to contain <ol> tag")
	}
	if !strings.Contains(html, "<li>") {
		t.Error("Expected HTML to contain <li> tag")
	}
}

// TestStripInlineTags tests the stripInlineTags function
func TestStripInlineTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "strong tag",
			input:    "<strong>bold</strong>",
			expected: "bold",
		},
		{
			name:     "em tag",
			input:    "<em>italic</em>",
			expected: "italic",
		},
		{
			name:     "code tag",
			input:    "<code>code</code>",
			expected: "code",
		},
		{
			name:     "anchor tag",
			input:    `<a href="http://example.com">link</a>`,
			expected: "link",
		},
		{
			name:     "br tag",
			input:    "line1<br/>line2",
			expected: "line1 line2",
		},
		{
			name:     "mixed tags",
			input:    "<strong>bold</strong> and <em>italic</em>",
			expected: "bold and italic",
		},
		{
			name:     "nested tags",
			input:    "<strong><em>bold italic</em></strong>",
			expected: "bold italic",
		},
		{
			name:     "plain text",
			input:    "no tags here",
			expected: "no tags here",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripInlineTags(tt.input)
			if result != tt.expected {
				t.Errorf("stripInlineTags(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestDecodeHTMLEntities tests the decodeHTMLEntities function
func TestDecodeHTMLEntities(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "ampersand",
			input:    "&amp;",
			expected: "&",
		},
		{
			name:     "less than",
			input:    "&lt;",
			expected: "<",
		},
		{
			name:     "greater than",
			input:    "&gt;",
			expected: ">",
		},
		{
			name:     "quote",
			input:    "&quot;",
			expected: "\"",
		},
		{
			name:     "apostrophe",
			input:    "&#39;",
			expected: "'",
		},
		{
			name:     "nbsp",
			input:    "&nbsp;",
			expected: " ",
		},
		{
			name:     "mixed entities",
			input:    "&lt;div&gt; &amp; &quot;text&quot;",
			expected: "<div> & \"text\"",
		},
		{
			name:     "no entities",
			input:    "plain text",
			expected: "plain text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := decodeHTMLEntities(tt.input)
			if result != tt.expected {
				t.Errorf("decodeHTMLEntities(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestFullPDFGenerationWithNestedLists tests end-to-end PDF generation with nested lists
func TestFullPDFGenerationWithNestedLists(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "markdown-to-pdf-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test markdown file
	mdContent := `# Test Document

## Nested Unordered List

- Level 1 Item 1
  - Level 2 Item 1a
  - Level 2 Item 1b
    - Level 3 Item
  - Level 2 Item 1c
- Level 1 Item 2

## Nested Ordered List

1. First
   1. Sub-first
   2. Sub-second
2. Second
`
	mdFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(mdFile, []byte(mdContent), 0644); err != nil {
		t.Fatalf("Failed to write test markdown: %v", err)
	}

	// Read the markdown file
	content, err := os.ReadFile(mdFile)
	if err != nil {
		t.Fatalf("Failed to read markdown file: %v", err)
	}

	// Convert to HTML
	htmlContent, err := convertMarkdownToHTML(content)
	if err != nil {
		t.Fatalf("Failed to convert markdown to HTML: %v", err)
	}

	// Create renderer
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	// Add cover page and render content
	renderer.renderCoverPage("Test Document", "")
	renderer.renderHTML(htmlContent)

	// Save PDF
	pdfFile := filepath.Join(tmpDir, "test.pdf")
	if err := renderer.save(pdfFile); err != nil {
		t.Fatalf("Failed to save PDF: %v", err)
	}

	// Verify PDF was created
	if _, err := os.Stat(pdfFile); os.IsNotExist(err) {
		t.Error("Expected PDF file to be created")
	}
}

// TestRenderListItemsIndentation verifies that indentation increases with nesting level
func TestRenderListItemsIndentation(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// Test that different nesting levels produce different X positions
	// We can't directly test X position, but we can verify the function handles
	// multiple nesting levels without error

	// 4-level deep nesting to test bullet style cycling
	content := `<li>L1<ul><li>L2<ul><li>L3<ul><li>L4</li></ul></li></ul></li></ul></li>`
	renderer.renderList(content, false)

	// If we get here without panic, basic indentation logic works
}

// TestRenderListBulletStyles tests that different bullet styles are used at different levels
func TestRenderListBulletStyles(t *testing.T) {
	// This is more of a visual test, but we verify the function handles
	// 4+ levels of nesting (which cycles through bullet styles)
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// 5 levels deep to test bullet style cycling (we have 4 styles)
	content := `<li>Level 0<ul><li>Level 1<ul><li>Level 2<ul><li>Level 3<ul><li>Level 4</li></ul></li></ul></li></ul></li></ul></li>`
	renderer.renderList(content, false)

	// If we get here without panic, bullet style cycling works
}

// TestParseListItems tests the parseListItems function
func TestParseListItems(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "single item",
			input:    "<li>Item 1</li>",
			expected: []string{"Item 1"},
		},
		{
			name:     "multiple items",
			input:    "<li>Item 1</li><li>Item 2</li><li>Item 3</li>",
			expected: []string{"Item 1", "Item 2", "Item 3"},
		},
		{
			name:     "nested list inside item",
			input:    "<li>Parent<ul><li>Child</li></ul></li>",
			expected: []string{"Parent<ul><li>Child</li></ul>"},
		},
		{
			name:     "deeply nested lists",
			input:    "<li>L1<ul><li>L2<ul><li>L3</li></ul></li></ul></li>",
			expected: []string{"L1<ul><li>L2<ul><li>L3</li></ul></li></ul>"},
		},
		{
			name:     "multiple items with nested lists",
			input:    "<li>A<ul><li>A1</li></ul></li><li>B</li><li>C<ul><li>C1</li></ul></li>",
			expected: []string{"A<ul><li>A1</li></ul>", "B", "C<ul><li>C1</li></ul>"},
		},
		{
			name:     "empty content",
			input:    "",
			expected: nil,
		},
		{
			name:     "no list items",
			input:    "plain text without list items",
			expected: nil,
		},
		{
			name:     "empty list item",
			input:    "<li></li>",
			expected: []string{""},
		},
		{
			name:     "item with newlines",
			input:    "<li>Line 1\nLine 2</li>",
			expected: []string{"Line 1\nLine 2"},
		},
		{
			name:     "nested list with newlines",
			input:    "<li>Parent\n<ul>\n<li>Child</li>\n</ul>\n</li>",
			expected: []string{"Parent\n<ul>\n<li>Child</li>\n</ul>\n"},
		},
		{
			name:     "mixed ordered and unordered nested",
			input:    "<li>Parent<ol><li>Ordered 1</li><li>Ordered 2</li></ol></li>",
			expected: []string{"Parent<ol><li>Ordered 1</li><li>Ordered 2</li></ol>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseListItems(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseListItems(%q) returned %d items, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, item := range result {
				if item != tt.expected[i] {
					t.Errorf("parseListItems(%q)[%d] = %q, want %q", tt.input, i, item, tt.expected[i])
				}
			}
		})
	}
}

// TestFindClosingTag tests the findClosingTag function
func TestFindClosingTag(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		tag      string
		expected int
	}{
		{
			name:     "simple tag",
			input:    "content</li>",
			tag:      "li",
			expected: 7,
		},
		{
			name:     "nested same tag",
			input:    "outer<li>inner</li>more</li>",
			tag:      "li",
			expected: 23,
		},
		{
			name:     "deeply nested",
			input:    "a<li>b<li>c</li>d</li>e</li>",
			tag:      "li",
			expected: 23,
		},
		{
			name:     "no closing tag",
			input:    "content without closing",
			tag:      "li",
			expected: -1,
		},
		{
			name:     "nested ul tags",
			input:    "text<ul><li>item</li></ul>more</ul>",
			tag:      "ul",
			expected: 30,
		},
		{
			name:     "empty content",
			input:    "</li>",
			tag:      "li",
			expected: 0,
		},
		{
			name:     "tag with attributes in nested",
			input:    "text<li class='test'>nested</li>end</li>",
			tag:      "li",
			expected: 35,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := findClosingTag(tt.input, tt.tag)
			if result != tt.expected {
				t.Errorf("findClosingTag(%q, %q) = %d, want %d", tt.input, tt.tag, result, tt.expected)
			}
		})
	}
}

// TestConvertMarkdownToHTMLNestedListWithAsterisks tests nested lists using asterisks
func TestConvertMarkdownToHTMLNestedListWithAsterisks(t *testing.T) {
	markdown := `- level 1
  * level 2`

	html, err := convertMarkdownToHTML([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	// Should have nested ul structure
	if !strings.Contains(html, "<ul>") {
		t.Error("Expected HTML to contain <ul> tag")
	}
	// Should have two li elements
	if strings.Count(html, "<li>") != 2 {
		t.Errorf("Expected 2 <li> tags, got %d", strings.Count(html, "<li>"))
	}
}

// TestConvertMarkdownToHTMLDeeplyNestedList tests 3+ level nested lists
func TestConvertMarkdownToHTMLDeeplyNestedList(t *testing.T) {
	markdown := `- level 1
  - level 2
    - level 3
      - level 4`

	html, err := convertMarkdownToHTML([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	// Should have 4 ul tags (one for each level)
	ulCount := strings.Count(html, "<ul>")
	if ulCount != 4 {
		t.Errorf("Expected 4 <ul> tags for 4 nesting levels, got %d", ulCount)
	}

	// Should have 4 li elements
	liCount := strings.Count(html, "<li>")
	if liCount != 4 {
		t.Errorf("Expected 4 <li> tags, got %d", liCount)
	}
}

// TestConvertMarkdownToHTMLMixedListTypes tests mixed ordered/unordered lists
func TestConvertMarkdownToHTMLMixedListTypes(t *testing.T) {
	markdown := `- Unordered item
  1. Ordered sub-item 1
  2. Ordered sub-item 2
- Another unordered`

	html, err := convertMarkdownToHTML([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	// Should have both ul and ol tags
	if !strings.Contains(html, "<ul>") {
		t.Error("Expected HTML to contain <ul> tag")
	}
	if !strings.Contains(html, "<ol>") {
		t.Error("Expected HTML to contain <ol> tag for nested ordered list")
	}
}

// TestRenderListMultipleSiblingNested tests multiple sibling items each with nested lists
func TestRenderListMultipleSiblingNested(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// Multiple top-level items each with their own nested list
	content := `<li>Parent A<ul><li>Child A1</li><li>Child A2</li></ul></li><li>Parent B<ul><li>Child B1</li></ul></li><li>Parent C<ul><li>Child C1</li><li>Child C2</li><li>Child C3</li></ul></li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListNestedOrderedInUnordered tests ordered list nested inside unordered
func TestRenderListNestedOrderedInUnordered(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	content := `<li>Unordered parent<ol><li>Ordered child 1</li><li>Ordered child 2</li><li>Ordered child 3</li></ol></li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListNestedUnorderedInOrdered tests unordered list nested inside ordered
func TestRenderListNestedUnorderedInOrdered(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	content := `<li>Ordered parent<ul><li>Unordered child 1</li><li>Unordered child 2</li></ul></li><li>Ordered item 2</li>`
	renderer.renderList(content, true)

	// If we get here without panic, the test passes
}

// TestRenderListWithLongContent tests list items with long text content
func TestRenderListWithLongContent(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	longText := "This is a very long list item that should wrap to multiple lines when rendered in the PDF document because it exceeds the available width for the content area."
	content := `<li>` + longText + `<ul><li>Nested item under long content</li></ul></li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListWithSpecialCharacters tests list items with special characters
func TestRenderListWithSpecialCharacters(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	content := `<li>Item with "quotes" and 'apostrophes'</li><li>Item with &lt;tags&gt; escaped</li><li>Item with special chars: @#$%^&*()</li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestRenderListComplexNesting tests complex real-world nesting scenario
func TestRenderListComplexNesting(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// Simulate a table of contents style nested list
	content := `<li>Chapter 1<ul><li>Section 1.1<ul><li>Subsection 1.1.1</li><li>Subsection 1.1.2</li></ul></li><li>Section 1.2</li></ul></li><li>Chapter 2<ul><li>Section 2.1<ul><li>Subsection 2.1.1<ul><li>Paragraph 2.1.1.1</li></ul></li></ul></li></ul></li>`
	renderer.renderList(content, false)

	// If we get here without panic, the test passes
}

// TestParseListItemsPreservesInnerHTML tests that parseListItems preserves inner HTML structure
func TestParseListItemsPreservesInnerHTML(t *testing.T) {
	input := `<li><strong>Bold</strong> and <em>italic</em><ul><li>Nested with <code>code</code></li></ul></li>`
	result := parseListItems(input)

	if len(result) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result))
	}

	expected := `<strong>Bold</strong> and <em>italic</em><ul><li>Nested with <code>code</code></li></ul>`
	if result[0] != expected {
		t.Errorf("parseListItems did not preserve inner HTML.\nGot: %q\nWant: %q", result[0], expected)
	}
}

// TestRenderListSixLevelsDeep tests 6 levels of nesting (beyond bullet style count)
func TestRenderListSixLevelsDeep(t *testing.T) {
	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}
	renderer.pdf.AddPage()

	// 6 levels deep - bullet styles should cycle
	content := `<li>L0<ul><li>L1<ul><li>L2<ul><li>L3<ul><li>L4<ul><li>L5</li></ul></li></ul></li></ul></li></ul></li></ul></li>`
	renderer.renderList(content, false)

	// If we get here without panic, bullet cycling works for 6+ levels
}

// TestFullPDFGenerationWithComplexNestedLists tests end-to-end with the original bug case
func TestFullPDFGenerationWithComplexNestedLists(t *testing.T) {
	// This tests the exact scenario from the bug report
	markdown := `- level 1
  * level 2`

	html, err := convertMarkdownToHTML([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to convert markdown: %v", err)
	}

	renderer, err := newPDFRenderer(generateOptions{language: "en"})
	if err != nil {
		t.Fatalf("Failed to create renderer: %v", err)
	}

	renderer.pdf.AddPage()
	renderer.renderHTML(html)

	// Create temp file to verify PDF can be saved
	tmpDir, err := os.MkdirTemp("", "nested-list-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	pdfFile := filepath.Join(tmpDir, "nested.pdf")
	if err := renderer.save(pdfFile); err != nil {
		t.Fatalf("Failed to save PDF: %v", err)
	}

	// Verify file was created and has content
	info, err := os.Stat(pdfFile)
	if err != nil {
		t.Fatalf("Failed to stat PDF file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("PDF file is empty")
	}
}

// TestParseListItemsWithWhitespace tests handling of whitespace in list items
func TestParseListItemsWithWhitespace(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{
			name:     "leading whitespace",
			input:    "<li>  Item with leading spaces</li>",
			expected: []string{"  Item with leading spaces"},
		},
		{
			name:     "trailing whitespace",
			input:    "<li>Item with trailing spaces  </li>",
			expected: []string{"Item with trailing spaces  "},
		},
		{
			name:     "whitespace between items",
			input:    "<li>Item 1</li>  \n  <li>Item 2</li>",
			expected: []string{"Item 1", "Item 2"},
		},
		{
			name:     "tabs and newlines",
			input:    "<li>Item\twith\ttabs</li>\n<li>Item\nwith\nnewlines</li>",
			expected: []string{"Item\twith\ttabs", "Item\nwith\nnewlines"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseListItems(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("parseListItems(%q) returned %d items, want %d", tt.input, len(result), len(tt.expected))
				return
			}
			for i, item := range result {
				if item != tt.expected[i] {
					t.Errorf("parseListItems(%q)[%d] = %q, want %q", tt.input, i, item, tt.expected[i])
				}
			}
		})
	}
}

// TestParseHTMLElements tests the parseHTMLElements function with list content
func TestParseHTMLElementsWithLists(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedTag string
		shouldFind  bool
	}{
		{
			name:        "simple unordered list",
			input:       "<ul><li>Item</li></ul>",
			expectedTag: "ul",
			shouldFind:  true,
		},
		{
			name:        "simple ordered list",
			input:       "<ol><li>Item</li></ol>",
			expectedTag: "ol",
			shouldFind:  true,
		},
		{
			name:        "nested list structure",
			input:       "<ul><li>Parent<ul><li>Child</li></ul></li></ul>",
			expectedTag: "ul",
			shouldFind:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			elements := parseHTMLElements(tt.input)
			found := false
			for _, elem := range elements {
				if elem.tag == tt.expectedTag {
					found = true
					break
				}
			}
			if found != tt.shouldFind {
				t.Errorf("parseHTMLElements(%q): expected to find tag %q = %v, got %v",
					tt.input, tt.expectedTag, tt.shouldFind, found)
			}
		})
	}
}
