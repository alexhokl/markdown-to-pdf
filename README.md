# markdown-to-pdf

A command-line tool to convert Markdown files to PDF format.

## Usage

```bash
markdown-to-pdf generate -i input.md -o output.pdf
```

### Options

- `-i, --input` - Path to the markdown file (required)
- `-o, --output` - Path to output PDF file (required)
- `-t, --title` - Title of the book (defaults to first H1 heading or filename)
- `-f, --overwrite` - Overwrite existing PDF file

