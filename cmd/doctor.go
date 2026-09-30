package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yakovlef/totext/internal/deps"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check and, if needed, install dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		status := deps.Check(modelFlag)
		printStatus(status)
		if status.OK() {
			return nil
		}
		if _, err := deps.Ensure(status); err != nil {
			return err
		}
		fmt.Println("---")
		printStatus(deps.Check(modelFlag))
		return nil
	},
}

func printStatus(s deps.Status) {
	report("ffmpeg", s.FFmpegPath)
	report("whisper-cli", s.WhisperCliPath)
	report("model", s.ModelPath)
}

func report(name, path string) {
	if path == "" {
		fmt.Printf("✗ %s: not found\n", name)
		return
	}
	fmt.Printf("✓ %s: %s\n", name, path)
}
