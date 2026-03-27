package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// fontConfig holds paths to font files for a font family
type fontConfig struct {
	name    string // font family name to use in PDF
	regular string // path to regular weight font
	bold    string // path to bold weight font (optional)
}

// languageFontPreferences maps language codes to preferred font families
// Listed in order of preference - first available font will be used
// Only fonts with TTF/OTF files are included (TTC files are not supported by fpdf)
var languageFontPreferences = map[string][]string{
	// Japanese
	"ja": {
		"NotoSansJP",
		"NotoSansCJKjp",
		"YuGothic",
		"HiraKakuProN",
	},
	// Simplified Chinese
	"zh": {
		"NotoSansSC",
		"NotoSansCJKsc",
		"SimHei",
	},
	// Traditional Chinese
	"zh-TW": {
		"NotoSansTC",
		"NotoSansCJKtc",
	},
	"zh-HK": {
		"NotoSansHK",
		"NotoSansCJKhk",
	},
	// Korean
	"ko": {
		"NotoSansKR",
		"NotoSansCJKkr",
		"AppleGothic",
		"Malgun Gothic",
	},
}

// fontFilePatterns maps font family names to possible file name patterns
// Order matters: .ttf and .otf are preferred over .ttc (which requires extraction)
var fontFilePatterns = map[string][]string{
	// Japanese - prioritize TTF/OTF over TTC
	"NotoSansJP":    {"NotoSansJP-Regular.ttf", "NotoSansJP-Regular.otf", "NotoSansJP[wght].ttf"},
	"NotoSansCJKjp": {"NotoSansCJKjp-Regular.ttf", "NotoSansCJKjp-Regular.otf"},
	"YuGothic":      {"YuGothic-Medium.otf", "YuGothic-Regular.otf", "yugothic.ttf"},
	"HiraKakuProN":  {"HiraKakuProN-W3.otf"},

	// Simplified Chinese
	"NotoSansSC":    {"NotoSansSC-Regular.ttf", "NotoSansSC-Regular.otf", "NotoSansSC[wght].ttf"},
	"NotoSansCJKsc": {"NotoSansCJKsc-Regular.ttf", "NotoSansCJKsc-Regular.otf"},
	"SimHei":        {"simhei.ttf", "SimHei.ttf"},

	// Traditional Chinese
	"NotoSansTC":    {"NotoSansTC-Regular.ttf", "NotoSansTC-Regular.otf", "NotoSansTC[wght].ttf"},
	"NotoSansCJKtc": {"NotoSansCJKtc-Regular.ttf", "NotoSansCJKtc-Regular.otf"},

	// Hong Kong
	"NotoSansHK":    {"NotoSansHK-Regular.ttf", "NotoSansHK-Regular.otf", "NotoSansHK[wght].ttf"},
	"NotoSansCJKhk": {"NotoSansCJKhk-Regular.ttf", "NotoSansCJKhk-Regular.otf"},

	// Korean
	"NotoSansKR":    {"NotoSansKR-Regular.ttf", "NotoSansKR-Regular.otf", "NotoSansKR[wght].ttf"},
	"NotoSansCJKkr": {"NotoSansCJKkr-Regular.ttf", "NotoSansCJKkr-Regular.otf"},
	"AppleGothic":   {"AppleGothic.ttf"},
	"Malgun Gothic": {"malgun.ttf", "MalgunGothic.ttf"},
}

// boldFontFilePatterns maps font family names to bold variant file patterns
var boldFontFilePatterns = map[string][]string{
	"NotoSansJP":    {"NotoSansJP-Bold.ttf", "NotoSansJP-Bold.otf"},
	"NotoSansCJKjp": {"NotoSansCJKjp-Bold.ttf", "NotoSansCJKjp-Bold.otf"},
	"YuGothic":      {"YuGothic-Bold.otf"},
	"HiraKakuProN":  {"HiraKakuProN-W6.otf"},

	"NotoSansSC":    {"NotoSansSC-Bold.ttf", "NotoSansSC-Bold.otf"},
	"NotoSansCJKsc": {"NotoSansCJKsc-Bold.ttf", "NotoSansCJKsc-Bold.otf"},

	"NotoSansTC":    {"NotoSansTC-Bold.ttf", "NotoSansTC-Bold.otf"},
	"NotoSansCJKtc": {"NotoSansCJKtc-Bold.ttf", "NotoSansCJKtc-Bold.otf"},

	"NotoSansHK":    {"NotoSansHK-Bold.ttf", "NotoSansHK-Bold.otf"},
	"NotoSansCJKhk": {"NotoSansCJKhk-Bold.ttf", "NotoSansCJKhk-Bold.otf"},

	"NotoSansKR":    {"NotoSansKR-Bold.ttf", "NotoSansKR-Bold.otf"},
	"NotoSansCJKkr": {"NotoSansCJKkr-Bold.ttf", "NotoSansCJKkr-Bold.otf"},
	"Malgun Gothic": {"malgunbd.ttf", "MalgunGothicBold.ttf"},
}

// getFontDirectories returns the list of directories to search for fonts
func getFontDirectories() []string {
	var dirs []string

	homeDir, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		dirs = []string{
			"/System/Library/Fonts",
			"/System/Library/Fonts/Supplemental",
			"/Library/Fonts",
			filepath.Join(homeDir, "Library/Fonts"),
			// Homebrew font locations
			"/opt/homebrew/share/fonts",
			"/usr/local/share/fonts",
		}
		// Add macOS font asset directories (for fonts like YuGothic)
		assetBase := "/System/Library/AssetsV2/com_apple_MobileAsset_Font7"
		if entries, err := os.ReadDir(assetBase); err == nil {
			for _, entry := range entries {
				if entry.IsDir() {
					assetPath := filepath.Join(assetBase, entry.Name(), "AssetData")
					if info, err := os.Stat(assetPath); err == nil && info.IsDir() {
						dirs = append(dirs, assetPath)
					}
				}
			}
		}
	case "linux":
		dirs = []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
			filepath.Join(homeDir, ".fonts"),
			filepath.Join(homeDir, ".local/share/fonts"),
			// Snap and Flatpak locations
			"/snap/common/fonts",
		}
	case "windows":
		winDir := os.Getenv("WINDIR")
		if winDir == "" {
			winDir = "C:\\Windows"
		}
		dirs = []string{
			filepath.Join(winDir, "Fonts"),
			filepath.Join(homeDir, "AppData", "Local", "Microsoft", "Windows", "Fonts"),
		}
	}

	return dirs
}

// findFontFile searches for a font file in the system font directories
// Only returns TrueType-based fonts (not CFF/PostScript-based OpenType)
func findFontFile(patterns []string) string {
	fontDirs := getFontDirectories()

	for _, dir := range fontDirs {
		for _, pattern := range patterns {
			// Direct match
			path := filepath.Join(dir, pattern)
			if fileExists(path) && isTrueTypeFont(path) {
				return path
			}

			// Search in subdirectories (fonts are often organized in folders)
			matches, err := filepath.Glob(filepath.Join(dir, "*", pattern))
			if err == nil && len(matches) > 0 {
				for _, match := range matches {
					if isTrueTypeFont(match) {
						return match
					}
				}
			}

			// Case-insensitive search on case-sensitive filesystems
			if runtime.GOOS != "windows" {
				found := findFontCaseInsensitive(dir, pattern)
				if found != "" && isTrueTypeFont(found) {
					return found
				}
			}
		}
	}

	return ""
}

// findFontCaseInsensitive searches for a font file case-insensitively
func findFontCaseInsensitive(baseDir, pattern string) string {
	patternLower := strings.ToLower(pattern)

	var result string
	filepath.Walk(baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if strings.ToLower(info.Name()) == patternLower {
			result = path
			return filepath.SkipAll // Stop walking
		}
		return nil
	})

	return result
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// isTrueTypeFont checks if a font file is TrueType-based (not CFF/PostScript)
// fpdf only supports TrueType outlines, not CFF-based OpenType fonts
func isTrueTypeFont(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Read the first 4 bytes to check the signature
	header := make([]byte, 4)
	if _, err := f.Read(header); err != nil {
		return false
	}

	// TrueType signatures:
	// - 0x00010000 - TrueType
	// - 0x74727565 ("true") - TrueType
	// - 0x4F54544F ("OTTO") - OpenType with CFF outlines (NOT supported by fpdf)

	// Check for "OTTO" (CFF-based OpenType) - NOT supported
	if header[0] == 'O' && header[1] == 'T' && header[2] == 'T' && header[3] == 'O' {
		return false
	}

	// Check for valid TrueType signatures
	// 0x00010000
	if header[0] == 0x00 && header[1] == 0x01 && header[2] == 0x00 && header[3] == 0x00 {
		return true
	}
	// "true"
	if header[0] == 't' && header[1] == 'r' && header[2] == 'u' && header[3] == 'e' {
		return true
	}

	return false
}

// discoverFont attempts to find a suitable font for the given language
func discoverFont(language string) *fontConfig {
	// Normalize language code
	lang := normalizeLanguageCode(language)

	// Get preferred fonts for this language
	preferences, ok := languageFontPreferences[lang]
	if !ok {
		// Try base language (e.g., "zh" for "zh-CN")
		if idx := strings.Index(lang, "-"); idx > 0 {
			preferences, ok = languageFontPreferences[lang[:idx]]
		}
	}

	if !ok {
		return nil // No CJK font needed for this language
	}

	// Search for each preferred font in order
	for _, fontFamily := range preferences {
		patterns, ok := fontFilePatterns[fontFamily]
		if !ok {
			continue
		}

		regularPath := findFontFile(patterns)
		if regularPath == "" {
			continue
		}

		// Found a font, now look for bold variant
		var boldPath string
		if boldPatterns, ok := boldFontFilePatterns[fontFamily]; ok {
			boldPath = findFontFile(boldPatterns)
		}

		return &fontConfig{
			name:    fontFamily,
			regular: regularPath,
			bold:    boldPath,
		}
	}

	return nil
}

// normalizeLanguageCode normalizes language codes to a consistent format
func normalizeLanguageCode(lang string) string {
	lang = strings.ToLower(lang)
	lang = strings.ReplaceAll(lang, "_", "-")

	// Map common variants
	switch lang {
	case "zh-cn", "zh-hans":
		return "zh"
	case "zh-tw", "zh-hant":
		return "zh-TW"
	case "zh-hk":
		return "zh-HK"
	case "jp":
		return "ja"
	case "kr":
		return "ko"
	}

	return lang
}

// isCJKLanguage returns true if the language code indicates a CJK language
func isCJKLanguage(language string) bool {
	lang := normalizeLanguageCode(language)
	switch {
	case lang == "ja":
		return true
	case lang == "ko":
		return true
	case strings.HasPrefix(lang, "zh"):
		return true
	}
	return false
}

// getFontInstallInstructions returns instructions for installing fonts
func getFontInstallInstructions(language string) string {
	lang := normalizeLanguageCode(language)

	var fontName string
	switch {
	case lang == "ja":
		fontName = "Noto Sans JP"
	case strings.HasPrefix(lang, "zh"):
		if lang == "zh-TW" || lang == "zh-HK" {
			fontName = "Noto Sans TC"
		} else {
			fontName = "Noto Sans SC"
		}
	case lang == "ko":
		fontName = "Noto Sans KR"
	default:
		fontName = "Noto Sans CJK"
	}

	var brewCask string
	switch {
	case lang == "ja":
		brewCask = "font-noto-sans-jp"
	case lang == "zh-TW" || lang == "zh-HK":
		brewCask = "font-noto-sans-tc"
	case strings.HasPrefix(lang, "zh"):
		brewCask = "font-noto-sans-sc"
	case lang == "ko":
		brewCask = "font-noto-sans-kr"
	default:
		brewCask = "font-noto-sans-cjk"
	}

	var instructions string
	switch runtime.GOOS {
	case "darwin":
		instructions = "Install using Homebrew:\n" +
			"  brew install --cask " + brewCask + "\n\n" +
			"Or install the full CJK collection (note: not detected automatically):\n" +
			"  brew install --cask font-noto-sans-cjk\n\n" +
			"Or download from Google Fonts:\n" +
			"  https://fonts.google.com/noto/specimen/" + strings.ReplaceAll(fontName, " ", "+")
	case "linux":
		instructions = `Install using your package manager:
  # Debian/Ubuntu
  sudo apt install fonts-noto-cjk

  # Fedora
  sudo dnf install google-noto-sans-cjk-fonts

  # Arch
  sudo pacman -S noto-fonts-cjk`
	case "windows":
		instructions = `Download from Google Fonts and install:
  https://fonts.google.com/noto/specimen/Noto+Sans+JP

Or install via Windows Settings > Personalization > Fonts`
	}

	return "Recommended font: " + fontName + "\n\n" + instructions
}
