/*
Copyright © 2026 Tom Helander <thomas.helander@gmail.com>
*/
package benoit

import "github.com/fatih/color"

var (
	styleTitle   = color.New(color.FgCyan, color.Bold)
	styleRunning = color.New(color.FgYellow)
	styleOK      = color.New(color.FgGreen, color.Bold)
	styleFail    = color.New(color.FgRed, color.Bold)
	styleSkip    = color.New(color.FgBlue)
)

const (
	emojiRun  = "🚀"
	emojiOK   = "✅"
	emojiFail = "❌"
	emojiSkip = "⏭️"
	emojiInfo = "ℹ️ "
)

func PrintStdout(chkID, stdout string) {
	styleSkip.Printf("%s [%-20s] output:\n%s\n", emojiInfo, chkID, stdout)
}

func PrintStderr(chkID, stderr string) {
	styleSkip.Printf("%s [%-20s] error :\n%s\n", emojiInfo, chkID, stderr)
}
