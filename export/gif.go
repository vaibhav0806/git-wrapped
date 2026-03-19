// export/gif.go
package export

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

const tapeTemplate = `# gh-wrapped VHS tape
Output "%s"
Set FontSize 14
Set FontFamily "JetBrains Mono"
Set Width 1200
Set Height 800
Set Theme { "name": "gh-wrapped", "black": "#08080c", "red": "#ff6b6b", "green": "#0be881", "yellow": "#feca57", "blue": "#48dbfb", "magenta": "#a29bfe", "cyan": "#48dbfb", "white": "#e8e8e8", "brightBlack": "#555555", "brightRed": "#ff4757", "brightGreen": "#0be881", "brightYellow": "#feca57", "brightBlue": "#48dbfb", "brightMagenta": "#ff9ff3", "brightCyan": "#48dbfb", "brightWhite": "#ffffff", "background": "#08080c", "foreground": "#e8e8e8", "selectionBackground": "#333333", "cursorColor": "#e8e8e8" }

Type "gh-wrapped %s --auto" Enter
Sleep 500ms
Wait
`

func GenerateGIF(username string) error {
	if _, err := exec.LookPath("vhs"); err != nil {
		return fmt.Errorf("VHS not found. Install with: brew install vhs")
	}

	outputFile := fmt.Sprintf("gh-wrapped-%s.gif", username)
	absOutput, _ := filepath.Abs(outputFile)

	tapeContent := fmt.Sprintf(tapeTemplate, absOutput, username)
	tapeFile := fmt.Sprintf("gh-wrapped-%s.tape", username)
	if err := os.WriteFile(tapeFile, []byte(tapeContent), 0644); err != nil {
		return fmt.Errorf("write tape: %w", err)
	}
	defer os.Remove(tapeFile)

	fmt.Printf("Recording GIF...\n")
	cmd := exec.Command("vhs", tapeFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("vhs failed: %w", err)
	}

	fmt.Printf("GIF saved to %s\n", absOutput)
	return nil
}
