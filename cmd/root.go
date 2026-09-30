package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/yakovlef/totext/internal/deps"
	"github.com/yakovlef/totext/internal/transcribe"
)

var version = "dev"

var (
	lang      string
	srt       bool
	keepAudio bool
	modelFlag string
)

var rootCmd = &cobra.Command{
	Use:     "totext FILE [FILE2 ...]",
	Short:   "Transcribe local audio/video files to text (whisper.cpp)",
	Version: version,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		status := deps.Check(modelFlag)
		if !status.OK() {
			var err error
			status, err = deps.Ensure(status)
			if err != nil {
				return err
			}
		}

		format := "txt"
		if srt {
			format = "srt"
		}
		opt := transcribe.Options{
			FFmpegPath:     status.FFmpegPath,
			WhisperCliPath: status.WhisperCliPath,
			ModelPath:      status.ModelPath,
			Lang:           lang,
			Format:         format,
			KeepAudio:      keepAudio,
		}

		var failed int
		for _, f := range args {
			if _, err := os.Stat(f); err != nil {
				fmt.Fprintf(os.Stderr, "totext: file not found: %s\n", f)
				failed++
				continue
			}
			if err := transcribe.File(f, opt); err != nil {
				fmt.Fprintf(os.Stderr, "totext: %s: %v\n", f, err)
				failed++
			}
		}
		if failed > 0 {
			return fmt.Errorf("%d of %d files failed", failed, len(args))
		}
		return nil
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().StringVarP(&lang, "lang", "l", "ru", "transcription language")
	rootCmd.Flags().BoolVar(&srt, "srt", false, "timestamped subtitles (.srt) instead of .txt")
	rootCmd.Flags().BoolVar(&keepAudio, "keep-audio", false, "keep the intermediate wav file")
	rootCmd.Flags().StringVar(&modelFlag, "model", "", "path to a whisper.cpp ggml model")
	rootCmd.AddCommand(doctorCmd)
}
