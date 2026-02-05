package cmd

import (
	"sort"
	"strings"
)

// emojiReplacements maps emoji characters to their text equivalents
// This allows emojis to be rendered in PDFs using standard fonts
var emojiReplacements = map[string]string{
	// Status/Check emojis
	"⭕":  "[O]",
	"⭕️": "[O]",
	"❌":  "[X]",
	"❌️": "[X]",
	"✅":  "[v]",
	"✅️": "[v]",
	"✓":  "[v]",
	"✔":  "[v]",
	"✔️": "[v]",
	"☑":  "[v]",
	"☑️": "[v]",
	"❎":  "[X]",
	"❎️": "[X]",

	// Common symbols
	"⚠":  "[!]",
	"⚠️": "[!]",
	"⛔":  "[X]",
	"⛔️": "[X]",
	"🚫":  "[X]",
	"💡":  "[i]",
	"ℹ":  "[i]",
	"ℹ️": "[i]",
	"❓":  "[?]",
	"❔":  "[?]",
	"❗":  "[!]",
	"❕":  "[!]",
	"⚡":  "[*]",
	"⚡️": "[*]",
	"🔥":  "[*]",
	"⭐":  "[*]",
	"⭐️": "[*]",
	"🌟":  "[*]",
	"✨":  "[*]",

	// Arrows
	"➡":  "->",
	"➡️": "->",
	"⬅":  "<-",
	"⬅️": "<-",
	"⬆":  "^",
	"⬆️": "^",
	"⬇":  "v",
	"⬇️": "v",
	"↑":  "^",
	"↓":  "v",
	"←":  "<-",
	"→":  "->",
	"↔":  "<->",
	"↔️": "<->",

	// Numbers/counts
	"1️⃣": "(1)",
	"2️⃣": "(2)",
	"3️⃣": "(3)",
	"4️⃣": "(4)",
	"5️⃣": "(5)",
	"6️⃣": "(6)",
	"7️⃣": "(7)",
	"8️⃣": "(8)",
	"9️⃣": "(9)",
	"🔟":   "(10)",

	// Hands/gestures
	"👍": "[+1]",
	"👎": "[-1]",
	"👌": "[OK]",
	"🤝": "[handshake]",
	"👋": "[wave]",
	"✋": "[stop]",
	"🙌": "[raised hands]",
	"👏": "[clap]",

	// Faces (simplified representations)
	"😀": ":)",
	"😃": ":)",
	"😄": ":D",
	"😁": ":D",
	"😊": ":)",
	"🙂": ":)",
	"😉": ";)",
	"😎": "8)",
	"😢": ":(",
	"😭": ":'(",
	"😱": ":O",
	"😡": ">:(",
	"🤔": "[thinking]",
	"😐": ":|",
	"😑": "-_-",
	"🙄": "[eye roll]",

	// Objects
	"📝":  "[note]",
	"📋":  "[clipboard]",
	"📌":  "[pin]",
	"📎":  "[clip]",
	"🔗":  "[link]",
	"📁":  "[folder]",
	"📂":  "[folder]",
	"📄":  "[doc]",
	"📃":  "[doc]",
	"📑":  "[bookmark]",
	"🔖":  "[bookmark]",
	"🏷":  "[tag]",
	"🏷️": "[tag]",
	"💻":  "[computer]",
	"🖥":  "[desktop]",
	"🖥️": "[desktop]",
	"⌨":  "[keyboard]",
	"⌨️": "[keyboard]",
	"🖱":  "[mouse]",
	"🖱️": "[mouse]",
	"📱":  "[phone]",
	"📧":  "[email]",
	"✉":  "[email]",
	"✉️": "[email]",
	"📞":  "[phone]",
	"🔔":  "[bell]",
	"🔕":  "[muted]",
	"🔒":  "[locked]",
	"🔓":  "[unlocked]",
	"🔑":  "[key]",
	"🗝":  "[key]",
	"🗝️": "[key]",

	// Time
	"⏰":  "[alarm]",
	"⏱":  "[timer]",
	"⏱️": "[timer]",
	"⏲":  "[timer]",
	"⏲️": "[timer]",
	"🕐":  "[1:00]",
	"🕑":  "[2:00]",
	"🕒":  "[3:00]",
	"🕓":  "[4:00]",
	"🕔":  "[5:00]",
	"🕕":  "[6:00]",
	"🕖":  "[7:00]",
	"🕗":  "[8:00]",
	"🕘":  "[9:00]",
	"🕙":  "[10:00]",
	"🕚":  "[11:00]",
	"🕛":  "[12:00]",

	// Weather
	"☀":  "[sun]",
	"☀️": "[sun]",
	"🌤":  "[sun]",
	"🌤️": "[sun]",
	"⛅":  "[cloudy]",
	"⛅️": "[cloudy]",
	"🌥":  "[cloudy]",
	"🌥️": "[cloudy]",
	"☁":  "[cloud]",
	"☁️": "[cloud]",
	"🌧":  "[rain]",
	"🌧️": "[rain]",
	"⛈":  "[storm]",
	"⛈️": "[storm]",
	"🌩":  "[lightning]",
	"🌩️": "[lightning]",
	"❄":  "[snow]",
	"❄️": "[snow]",

	// Hearts
	"❤":  "<3",
	"❤️": "<3",
	"🧡":  "<3",
	"💛":  "<3",
	"💚":  "<3",
	"💙":  "<3",
	"💜":  "<3",
	"🖤":  "<3",
	"🤍":  "<3",
	"🤎":  "<3",
	"💔":  "</3",

	// Misc
	"🎉":  "[party]",
	"🎊":  "[party]",
	"🎁":  "[gift]",
	"🏆":  "[trophy]",
	"🥇":  "[1st]",
	"🥈":  "[2nd]",
	"🥉":  "[3rd]",
	"🎯":  "[target]",
	"🚀":  "[rocket]",
	"💪":  "[strong]",
	"🔍":  "[search]",
	"🔎":  "[search]",
	"📊":  "[chart]",
	"📈":  "[up]",
	"📉":  "[down]",
	"💰":  "[$]",
	"💵":  "[$]",
	"💲":  "[$]",
	"🌍":  "[globe]",
	"🌎":  "[globe]",
	"🌏":  "[globe]",
	"🏠":  "[home]",
	"🏢":  "[building]",
	"🏗":  "[construction]",
	"🏗️": "[construction]",
	"🔧":  "[wrench]",
	"🔨":  "[hammer]",
	"🛠":  "[tools]",
	"🛠️": "[tools]",
	"⚙":  "[gear]",
	"⚙️": "[gear]",
	"🧪":  "[test]",
	"🧬":  "[dna]",
	"🔬":  "[microscope]",
	"🔭":  "[telescope]",
}

// replaceEmojis replaces emoji characters with their text equivalents
// so they can be rendered in PDFs using standard fonts
func replaceEmojis(s string) string {
	// Sort emoji keys by length (longest first) to ensure that emojis with
	// variation selectors (e.g., ➡️) are replaced before their base form (e.g., ➡)
	keys := make([]string, 0, len(emojiReplacements))
	for k := range emojiReplacements {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return len(keys[i]) > len(keys[j])
	})

	result := s
	for _, emoji := range keys {
		result = strings.ReplaceAll(result, emoji, emojiReplacements[emoji])
	}
	return result
}
