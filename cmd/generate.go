package cmd

import (
	"bytes"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alexhokl/helper/cli"
	"github.com/alexhokl/helper/iohelper"
	"github.com/go-pdf/fpdf"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

type generateOptions struct {
	markdownFilename string
	pdfFilename      string
	overwrite        bool
	title            string
	author           string
	language         string
	fontPath         string
}

var generateOps generateOptions

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate PDF file from the specified markdown file",
	RunE:  runGenerate,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	flags := generateCmd.Flags()
	flags.StringVarP(&generateOps.markdownFilename, "input", "i", "", "Path to markdown file")
	flags.StringVarP(&generateOps.pdfFilename, "output", "o", "", "Path to output PDF file")
	flags.BoolVarP(&generateOps.overwrite, "overwrite", "f", false, "Overwrite existing PDF file")
	flags.StringVarP(&generateOps.title, "title", "t", "", "Title of the book (defaults to filename)")
	flags.StringVarP(&generateOps.author, "author", "a", "", "Author of the book")
	flags.StringVarP(&generateOps.language, "language", "l", "en", "Language code (e.g., en, ja, zh)")
	flags.StringVar(&generateOps.fontPath, "font", "", "Path to custom TTF font file")

	if err := generateCmd.MarkFlagRequired("input"); err != nil {
		cli.LogUnableToMarkFlagAsRequired("input", err)
	}
	if err := generateCmd.MarkFlagRequired("output"); err != nil {
		cli.LogUnableToMarkFlagAsRequired("output", err)
	}
}

func runGenerate(cmd *cobra.Command, args []string) error {
	if err := validateGenerateOptions(generateOps); err != nil {
		return err
	}

	// Read the Markdown file
	content, err := os.ReadFile(generateOps.markdownFilename)
	if err != nil {
		return fmt.Errorf("failed to read markdown file: %w", err)
	}

	// Convert Markdown to HTML
	htmlContent, err := convertMarkdownToHTML(content)
	if err != nil {
		return fmt.Errorf("failed to convert markdown to HTML: %w", err)
	}

	// Determine title
	title := generateOps.title
	if title == "" {
		// Try to extract title from first H1 heading
		title = extractTitleFromMarkdown(string(content))
		if title == "" {
			// Fall back to filename without extension
			title = strings.TrimSuffix(filepath.Base(generateOps.markdownFilename), filepath.Ext(generateOps.markdownFilename))
		}
	}

	// Create PDF
	if err := createPDF(title, htmlContent, generateOps); err != nil {
		return fmt.Errorf("failed to create PDF: %w", err)
	}

	fmt.Printf("Successfully created %s\n", generateOps.pdfFilename)
	return nil
}

func validateGenerateOptions(options generateOptions) error {
	if !iohelper.IsFileExist(options.markdownFilename) {
		return fmt.Errorf("markdown file %s does not exist", options.markdownFilename)
	}

	if iohelper.IsFileExist(options.pdfFilename) && !options.overwrite {
		return fmt.Errorf("PDF file %s already exists, use option -f to overwrite", options.pdfFilename)
	}

	return nil
}

func convertMarkdownToHTML(content []byte) (string, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
		),
	)

	var buf bytes.Buffer
	if err := md.Convert(content, &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func extractTitleFromMarkdown(content string) string {
	lines := strings.SplitSeq(content, "\n")
	for line := range lines {
		line = strings.TrimSpace(line)
		if after, ok := strings.CutPrefix(line, "# "); ok {
			return after
		}
	}
	return ""
}

// pdfRenderer handles rendering HTML content to PDF
type pdfRenderer struct {
	pdf           *fpdf.Fpdf
	pageWidth     float64
	pageHeight    float64
	leftMargin    float64
	rightMargin   float64
	topMargin     float64
	bottomMargin  float64
	contentWidth  float64
	lineHeight    float64
	fontFamily    string // current font family name
	monoFont      string // monospace font family name
	hasBoldFont   bool   // whether bold variant is available
	hasItalicFont bool   // whether italic variant is available
	inCodeBlock   bool
	inBlockquote  bool
	inList        bool
	listItemNum   int
	isOrderedList bool
}

// newPDFRenderer creates a new PDF renderer with A4 page size
func newPDFRenderer(options generateOptions) (*pdfRenderer, error) {
	pdf := fpdf.New("P", "mm", "A4", "")

	leftMargin := 20.0
	rightMargin := 20.0
	topMargin := 20.0
	bottomMargin := 20.0

	pdf.SetMargins(leftMargin, topMargin, rightMargin)
	pdf.SetAutoPageBreak(true, bottomMargin)

	pageWidth, pageHeight := pdf.GetPageSize()
	contentWidth := pageWidth - leftMargin - rightMargin

	r := &pdfRenderer{
		pdf:           pdf,
		pageWidth:     pageWidth,
		pageHeight:    pageHeight,
		leftMargin:    leftMargin,
		rightMargin:   rightMargin,
		topMargin:     topMargin,
		bottomMargin:  bottomMargin,
		contentWidth:  contentWidth,
		lineHeight:    6.0,
		fontFamily:    "Helvetica",
		monoFont:      "Courier",
		hasBoldFont:   true,
		hasItalicFont: true, // Built-in Helvetica has italic
	}

	// Load custom font or discover CJK font based on language
	if err := r.loadFonts(options); err != nil {
		return nil, err
	}

	return r, nil
}

// loadFonts loads appropriate fonts based on options
func (r *pdfRenderer) loadFonts(options generateOptions) error {
	// If custom font path is specified, use it
	if options.fontPath != "" {
		if !fileExists(options.fontPath) {
			return fmt.Errorf("font file not found: %s", options.fontPath)
		}
		if err := r.loadTTFFont("CustomFont", options.fontPath, ""); err != nil {
			return err
		}
		r.fontFamily = "CustomFont"
		r.hasBoldFont = false   // Custom font may not have bold variant
		r.hasItalicFont = false // Custom font may not have italic variant
		fmt.Printf("Using custom font: %s\n", options.fontPath)
		return nil
	}

	// Check if CJK font is needed
	if !isCJKLanguage(options.language) {
		// Use built-in Helvetica for non-CJK languages
		return nil
	}

	// Discover CJK font
	fontCfg := discoverFont(options.language)
	if fontCfg == nil {
		// Warn user but continue with Helvetica
		fmt.Printf("Warning: No CJK font found for language '%s'.\n", options.language)
		fmt.Printf("CJK characters may not render correctly.\n\n")
		fmt.Printf("To fix this, install a CJK font:\n%s\n\n", getFontInstallInstructions(options.language))
		return nil
	}

	// Load the discovered font
	if err := r.loadTTFFont(fontCfg.name, fontCfg.regular, ""); err != nil {
		return fmt.Errorf("failed to load font %s: %w", fontCfg.name, err)
	}
	r.fontFamily = fontCfg.name
	r.hasItalicFont = false // CJK fonts typically don't have italic variants

	// Load bold variant if available
	if fontCfg.bold != "" {
		if err := r.loadTTFFont(fontCfg.name, fontCfg.bold, "B"); err != nil {
			// Bold loading failed, continue without bold
			r.hasBoldFont = false
		} else {
			r.hasBoldFont = true
		}
	} else {
		r.hasBoldFont = false
	}

	fmt.Printf("Using font: %s (%s)\n", fontCfg.name, fontCfg.regular)
	return nil
}

// loadTTFFont loads a TTF/OTF font file directly
func (r *pdfRenderer) loadTTFFont(familyName, fontPath, style string) error {
	// fpdf's AddUTF8Font expects the font file to be in the fontpath directory
	// Set fontpath to the directory containing the font file
	fontDir := filepath.Dir(fontPath)
	fontFile := filepath.Base(fontPath)

	r.pdf.SetFontLocation(fontDir)
	r.pdf.AddUTF8Font(familyName, style, fontFile)

	if r.pdf.Err() {
		return fmt.Errorf("failed to add font: %w", r.pdf.Error())
	}

	return nil
}

// setMetadata sets PDF document metadata
func (r *pdfRenderer) setMetadata(title, author, language string) {
	r.pdf.SetTitle(title, true)
	if author != "" {
		r.pdf.SetAuthor(author, true)
	}
	r.pdf.SetCreator("markdown-to-pdf", true)
	r.pdf.SetLang(language)
}

// renderCoverPage creates a cover page with the title and optional author
func (r *pdfRenderer) renderCoverPage(title, author string) {
	r.pdf.AddPage()

	// Center the title vertically (roughly 40% from top)
	r.pdf.SetY(r.pageHeight * 0.4)

	// Title
	boldStyle := ""
	if r.hasBoldFont {
		boldStyle = "B"
	}
	r.pdf.SetFont(r.fontFamily, boldStyle, 28)
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.MultiCell(r.contentWidth, 12, title, "", "C", false)

	// Author (if provided)
	if author != "" {
		r.pdf.Ln(20)
		r.pdf.SetFont(r.fontFamily, "", 16)
		r.pdf.SetTextColor(100, 100, 100)
		r.pdf.MultiCell(r.contentWidth, 8, author, "", "C", false)
	}
}

// renderHTML parses and renders HTML content to PDF
func (r *pdfRenderer) renderHTML(htmlContent string) {
	r.pdf.AddPage()
	r.pdf.SetFont(r.fontFamily, "", 11)
	r.pdf.SetTextColor(0, 0, 0)

	// Replace emojis with text equivalents before parsing
	// This ensures emojis can be rendered using standard fonts
	htmlContent = replaceEmojis(htmlContent)

	// Parse HTML and render elements
	elements := parseHTMLElements(htmlContent)
	for _, elem := range elements {
		r.renderElement(elem)
	}
}

// htmlElement represents a parsed HTML element
type htmlElement struct {
	tag        string
	content    string
	attributes map[string]string
	children   []htmlElement
}

// parseHTMLElements parses HTML string into structured elements
func parseHTMLElements(htmlStr string) []htmlElement {
	var elements []htmlElement

	// Pre-process: handle code blocks specially (convert <pre><code>...</code></pre> to <pre>...</pre>)
	codeBlockPattern := regexp.MustCompile(`<pre><code[^>]*>([\s\S]*?)</code></pre>`)
	htmlStr = codeBlockPattern.ReplaceAllString(htmlStr, "<pre>$1</pre>")

	// Pattern for self-closing tags
	selfClosingPattern := regexp.MustCompile(`<(hr|br)\s*/?>`)

	// Pattern for opening tags
	openTagPattern := regexp.MustCompile(`<(h[1-6]|p|pre|blockquote|ul|ol|table)([^>]*)>`)

	pos := 0
	for pos < len(htmlStr) {
		// Try to match self-closing tag
		if loc := selfClosingPattern.FindStringIndex(htmlStr[pos:]); loc != nil && loc[0] == 0 {
			match := selfClosingPattern.FindStringSubmatch(htmlStr[pos:])
			elements = append(elements, htmlElement{tag: match[1]})
			pos += loc[1]
			continue
		}

		// Try to match opening tag
		if loc := openTagPattern.FindStringIndex(htmlStr[pos:]); loc != nil && loc[0] == 0 {
			match := openTagPattern.FindStringSubmatch(htmlStr[pos:])
			tag := match[1]
			attrs := match[2]
			pos += loc[1]

			// Find the corresponding closing tag
			closeTag := fmt.Sprintf("</%s>", tag)
			closeIdx := findClosingTag(htmlStr[pos:], tag)
			if closeIdx == -1 {
				// No closing tag found, skip
				continue
			}

			content := htmlStr[pos : pos+closeIdx]
			pos += closeIdx + len(closeTag)

			elements = append(elements, htmlElement{
				tag:        tag,
				content:    content,
				attributes: parseAttributes(attrs),
			})
			continue
		}

		// Skip whitespace and other characters
		pos++
	}

	return elements
}

// findClosingTag finds the index of the closing tag, handling nested tags
func findClosingTag(s string, tag string) int {
	openTag := fmt.Sprintf("<%s", tag)
	closeTag := fmt.Sprintf("</%s>", tag)
	depth := 1
	pos := 0

	for pos < len(s) && depth > 0 {
		openIdx := strings.Index(s[pos:], openTag)
		closeIdx := strings.Index(s[pos:], closeTag)

		if closeIdx == -1 {
			return -1 // No closing tag found
		}

		if openIdx != -1 && openIdx < closeIdx {
			// Found another opening tag before closing
			// Check if it's actually a complete opening tag (not just a prefix)
			nextChar := ""
			if pos+openIdx+len(openTag) < len(s) {
				nextChar = string(s[pos+openIdx+len(openTag)])
			}
			if nextChar == ">" || nextChar == " " {
				depth++
			}
			pos += openIdx + 1
		} else {
			// Found closing tag
			depth--
			if depth == 0 {
				return pos + closeIdx
			}
			pos += closeIdx + len(closeTag)
		}
	}

	return -1
}

// parseAttributes parses HTML attribute string into a map
func parseAttributes(attrStr string) map[string]string {
	attrs := make(map[string]string)
	attrPattern := regexp.MustCompile(`(\w+)=["']([^"']*)["']`)
	matches := attrPattern.FindAllStringSubmatch(attrStr, -1)
	for _, match := range matches {
		if len(match) > 2 {
			attrs[match[1]] = match[2]
		}
	}
	return attrs
}

// decodeHTMLEntities decodes common HTML entities
func decodeHTMLEntities(s string) string {
	replacements := map[string]string{
		"&amp;":  "&",
		"&lt;":   "<",
		"&gt;":   ">",
		"&quot;": "\"",
		"&#39;":  "'",
		"&nbsp;": " ",
	}
	for entity, char := range replacements {
		s = strings.ReplaceAll(s, entity, char)
	}
	return s
}

// renderElement renders a single HTML element to PDF
func (r *pdfRenderer) renderElement(elem htmlElement) {
	switch elem.tag {
	case "h1":
		r.renderHeading(elem.content, 1)
	case "h2":
		r.renderHeading(elem.content, 2)
	case "h3":
		r.renderHeading(elem.content, 3)
	case "h4":
		r.renderHeading(elem.content, 4)
	case "h5":
		r.renderHeading(elem.content, 5)
	case "h6":
		r.renderHeading(elem.content, 6)
	case "p":
		r.renderParagraph(elem.content)
	case "pre":
		r.renderCodeBlock(elem.content)
	case "blockquote":
		r.renderBlockquote(elem.content)
	case "ul":
		r.renderList(elem.content, false)
	case "ol":
		r.renderList(elem.content, true)
	case "table":
		r.renderTable(elem.content)
	case "hr":
		r.renderHorizontalRule()
	case "text":
		if strings.TrimSpace(elem.content) != "" {
			r.renderParagraph(elem.content)
		}
	}
}

// renderHeading renders a heading with appropriate size
func (r *pdfRenderer) renderHeading(content string, level int) {
	content = stripInlineTags(content)
	content = decodeHTMLEntities(content)

	// Add spacing before heading
	r.pdf.Ln(8)

	// Set font size based on heading level
	sizes := map[int]float64{1: 24, 2: 20, 3: 16, 4: 14, 5: 12, 6: 11}
	size := sizes[level]

	boldStyle := ""
	if r.hasBoldFont {
		boldStyle = "B"
	}
	r.pdf.SetFont(r.fontFamily, boldStyle, size)
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.MultiCell(r.contentWidth, size*0.5, content, "", "L", false)
	r.pdf.Ln(4)

	// Reset font
	r.pdf.SetFont(r.fontFamily, "", 11)
}

// renderParagraph renders a paragraph of text
func (r *pdfRenderer) renderParagraph(content string) {
	content = r.processInlineFormatting(content)
	content = decodeHTMLEntities(content)
	content = strings.TrimSpace(content)

	if content == "" {
		return
	}

	r.pdf.SetFont(r.fontFamily, "", 11)
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.MultiCell(r.contentWidth, r.lineHeight, content, "", "L", false)
	r.pdf.Ln(3)
}

// renderCodeBlock renders a code block with background
func (r *pdfRenderer) renderCodeBlock(content string) {
	content = decodeHTMLEntities(content)
	content = strings.TrimSpace(content)

	// Add spacing
	r.pdf.Ln(4)

	// Calculate block height
	r.pdf.SetFont(r.monoFont, "", 9)
	lines := strings.Split(content, "\n")
	blockHeight := float64(len(lines))*5.0 + 8.0

	// Check if we need a new page
	currentY := r.pdf.GetY()
	if currentY+blockHeight > r.pageHeight-r.bottomMargin {
		r.pdf.AddPage()
	}

	// Draw background
	r.pdf.SetFillColor(244, 244, 244)
	r.pdf.Rect(r.leftMargin, r.pdf.GetY(), r.contentWidth, blockHeight, "F")

	// Draw content
	r.pdf.SetTextColor(0, 0, 0)
	r.pdf.SetXY(r.leftMargin+4, r.pdf.GetY()+4)

	for _, line := range lines {
		r.pdf.CellFormat(r.contentWidth-8, 5, line, "", 1, "L", false, 0, "")
	}

	r.pdf.Ln(6)
	r.pdf.SetFont(r.fontFamily, "", 11)
}

// renderBlockquote renders a blockquote with left border
func (r *pdfRenderer) renderBlockquote(content string) {
	content = stripInlineTags(content)
	content = decodeHTMLEntities(content)
	content = strings.TrimSpace(content)

	r.pdf.Ln(4)

	// Draw left border
	startY := r.pdf.GetY()
	r.pdf.SetDrawColor(200, 200, 200)
	r.pdf.SetLineWidth(1.0)

	// Render text with indentation
	r.pdf.SetX(r.leftMargin + 10)
	italicStyle := ""
	if r.hasItalicFont {
		italicStyle = "I"
	}
	r.pdf.SetFont(r.fontFamily, italicStyle, 11)
	r.pdf.SetTextColor(100, 100, 100)
	r.pdf.MultiCell(r.contentWidth-15, r.lineHeight, content, "", "L", false)

	endY := r.pdf.GetY()
	r.pdf.Line(r.leftMargin+3, startY, r.leftMargin+3, endY)

	r.pdf.Ln(4)
	r.pdf.SetFont(r.fontFamily, "", 11)
	r.pdf.SetTextColor(0, 0, 0)
}

// renderList renders ordered or unordered lists
func (r *pdfRenderer) renderList(content string, ordered bool) {
	r.pdf.Ln(2)
	r.renderListItems(content, ordered, 0, 1)
	r.pdf.Ln(3)
}

// parseListItems extracts list items handling nested tags properly
func parseListItems(content string) []string {
	var items []string
	pos := 0

	for pos < len(content) {
		// Find the next <li> tag
		liStart := strings.Index(content[pos:], "<li>")
		if liStart == -1 {
			break
		}
		liStart += pos + 4 // Move past "<li>"

		// Find the matching </li> using proper nesting handling
		closeIdx := findClosingTag(content[liStart:], "li")
		if closeIdx == -1 {
			break
		}

		itemContent := content[liStart : liStart+closeIdx]
		items = append(items, itemContent)

		pos = liStart + closeIdx + 5 // Move past "</li>"
	}

	return items
}

// renderListItems recursively renders list items with proper nesting support
func (r *pdfRenderer) renderListItems(content string, ordered bool, level int, startNum int) int {
	// Bullet styles for different nesting levels
	bulletStyles := []string{"•", "◦", "▪", "▫"}

	// Indentation per level (in mm)
	indentPerLevel := 8.0
	baseIndent := r.leftMargin + 8

	// Parse list items at current level using proper nested tag handling
	listItems := parseListItems(content)

	itemNum := startNum
	for _, itemContent := range listItems {
		// Check for nested lists within this item
		nestedULPattern := regexp.MustCompile(`(?s)<ul>(.*)</ul>`)
		nestedOLPattern := regexp.MustCompile(`(?s)<ol>(.*)</ol>`)

		nestedUL := nestedULPattern.FindStringSubmatch(itemContent)
		nestedOL := nestedOLPattern.FindStringSubmatch(itemContent)

		// Extract the text content before any nested list
		textContent := itemContent
		if nestedUL != nil {
			textContent = itemContent[:nestedULPattern.FindStringIndex(itemContent)[0]]
		} else if nestedOL != nil {
			textContent = itemContent[:nestedOLPattern.FindStringIndex(itemContent)[0]]
		}

		// Clean up the text content
		textContent = stripInlineTags(textContent)
		textContent = decodeHTMLEntities(textContent)
		textContent = strings.TrimSpace(textContent)

		// Calculate indentation for current level
		currentIndent := baseIndent + (float64(level) * indentPerLevel)

		// Render the list item
		r.pdf.SetX(currentIndent)
		r.pdf.SetFont(r.fontFamily, "", 11)
		r.pdf.SetTextColor(0, 0, 0)

		var bullet string
		if ordered {
			bullet = fmt.Sprintf("%d. ", itemNum)
			itemNum++
		} else {
			// Use different bullet style based on nesting level
			bulletIndex := level % len(bulletStyles)
			bullet = bulletStyles[bulletIndex] + " "
		}

		// Render bullet
		r.pdf.CellFormat(8, r.lineHeight, bullet, "", 0, "L", false, 0, "")

		// Calculate available width for content
		availableWidth := r.contentWidth - 16 - (float64(level) * indentPerLevel)
		if availableWidth < 50 {
			availableWidth = 50 // Minimum width
		}

		// Render item text
		if textContent != "" {
			r.pdf.MultiCell(availableWidth, r.lineHeight, textContent, "", "L", false)
		} else {
			r.pdf.Ln(r.lineHeight)
		}

		// Recursively render nested lists
		if nestedUL != nil {
			r.renderListItems(nestedUL[1], false, level+1, 1)
		}
		if nestedOL != nil {
			r.renderListItems(nestedOL[1], true, level+1, 1)
		}
	}

	return itemNum
}

// renderTable renders an HTML table
func (r *pdfRenderer) renderTable(content string) {
	r.pdf.Ln(4)

	// Parse table rows
	rowPattern := regexp.MustCompile(`<tr>([\s\S]*?)</tr>`)
	headerCellPattern := regexp.MustCompile(`<th[^>]*>([\s\S]*?)</th>`)
	cellPattern := regexp.MustCompile(`<td[^>]*>([\s\S]*?)</td>`)

	rows := rowPattern.FindAllStringSubmatch(content, -1)
	if len(rows) == 0 {
		return
	}

	// Determine number of columns from first row
	firstRow := rows[0][1]
	headerCells := headerCellPattern.FindAllStringSubmatch(firstRow, -1)
	dataCells := cellPattern.FindAllStringSubmatch(firstRow, -1)

	numCols := len(headerCells)
	if numCols == 0 {
		numCols = len(dataCells)
	}
	if numCols == 0 {
		return
	}

	colWidth := r.contentWidth / float64(numCols)

	r.pdf.SetFont(r.fontFamily, "", 10)
	r.pdf.SetDrawColor(200, 200, 200)
	r.pdf.SetLineWidth(0.3)

	for _, row := range rows {
		rowContent := row[1]

		// Check for header cells
		headerCells := headerCellPattern.FindAllStringSubmatch(rowContent, -1)
		if len(headerCells) > 0 {
			boldStyle := ""
			if r.hasBoldFont {
				boldStyle = "B"
			}
			r.pdf.SetFont(r.fontFamily, boldStyle, 10)
			r.pdf.SetFillColor(244, 244, 244)
			for _, cell := range headerCells {
				cellText := stripInlineTags(cell[1])
				cellText = decodeHTMLEntities(cellText)
				r.pdf.CellFormat(colWidth, 8, cellText, "1", 0, "L", true, 0, "")
			}
			r.pdf.Ln(-1)
			continue
		}

		// Regular cells
		r.pdf.SetFont(r.fontFamily, "", 10)
		cells := cellPattern.FindAllStringSubmatch(rowContent, -1)
		for _, cell := range cells {
			cellText := stripInlineTags(cell[1])
			cellText = decodeHTMLEntities(cellText)
			r.pdf.CellFormat(colWidth, 8, cellText, "1", 0, "L", false, 0, "")
		}
		r.pdf.Ln(-1)
	}

	r.pdf.Ln(4)
	r.pdf.SetFont(r.fontFamily, "", 11)
}

// renderHorizontalRule renders a horizontal line
func (r *pdfRenderer) renderHorizontalRule() {
	r.pdf.Ln(6)
	r.pdf.SetDrawColor(200, 200, 200)
	r.pdf.SetLineWidth(0.5)
	y := r.pdf.GetY()
	r.pdf.Line(r.leftMargin, y, r.pageWidth-r.rightMargin, y)
	r.pdf.Ln(6)
}

// stripInlineTags removes inline HTML tags but preserves text
func stripInlineTags(s string) string {
	// Remove common inline tags while keeping content
	patterns := []string{
		`<strong>(.*?)</strong>`,
		`<em>(.*?)</em>`,
		`<code>(.*?)</code>`,
		`<a[^>]*>(.*?)</a>`,
		`<span[^>]*>(.*?)</span>`,
		`<br\s*/?>`,
	}

	result := s
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if pattern == `<br\s*/?>` {
			result = re.ReplaceAllString(result, " ")
		} else {
			result = re.ReplaceAllString(result, "$1")
		}
	}

	// Remove any remaining tags
	result = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(result, "")

	return result
}

// processInlineFormatting processes inline formatting (simplified - fpdf has limited support)
func (r *pdfRenderer) processInlineFormatting(content string) string {
	// For fpdf, we mainly strip tags as it doesn't support inline style changes well
	return stripInlineTags(content)
}

// save writes the PDF to a file
func (r *pdfRenderer) save(filename string) error {
	return r.pdf.OutputFileAndClose(filename)
}

// createPDF generates a PDF from HTML content
func createPDF(title, htmlContent string, options generateOptions) error {
	renderer, err := newPDFRenderer(options)
	if err != nil {
		return err
	}

	// Set document metadata
	renderer.setMetadata(title, options.author, options.language)

	// Create cover page
	renderer.renderCoverPage(title, options.author)

	// Render HTML content
	renderer.renderHTML(htmlContent)

	// Save PDF
	return renderer.save(options.pdfFilename)
}
