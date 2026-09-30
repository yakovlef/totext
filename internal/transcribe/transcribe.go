// Package transcribe converts a media file to text via ffmpeg + whisper-cli,
// mirroring the behavior of the original `transcrib` shell script.
package transcribe

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Options struct {
	FFmpegPath     string
	WhisperCliPath string
	ModelPath      string
	Lang           string
	Format         string // "txt" or "srt"
	KeepAudio      bool
}

// File converts and transcribes a single input file, writing the result
// next to it (same base name, .txt or .srt extension).
func File(path string, opt Options) error {
	base := strings.TrimSuffix(path, extOf(path))
	wav := base + ".__totext_tmp.wav"

	fmt.Fprintf(os.Stderr, ">> %s -> audio...\n", path)
	if err := toWav(opt.FFmpegPath, path, wav); err != nil {
		return fmt.Errorf("ffmpeg: %w", err)
	}
	if !opt.KeepAudio {
		defer os.Remove(wav)
	}

	fmt.Fprintf(os.Stderr, ">> %s -> transcribing (%s, %s)...\n", path, opt.Lang, opt.Format)
	if err := runWhisper(opt, wav, base); err != nil {
		return fmt.Errorf("whisper-cli: %w", err)
	}

	fmt.Fprintf(os.Stderr, ">> done: %s.%s\n", base, opt.Format)
	return nil
}

func extOf(path string) string {
	for i := len(path) - 1; i >= 0 && path[i] != '/'; i-- {
		if path[i] == '.' {
			return path[i:]
		}
	}
	return ""
}

func toWav(ffmpegPath, in, out string) error {
	cmd := exec.Command(ffmpegPath, "-nostdin", "-i", in, "-vn", "-ac", "1", "-ar", "16000", "-c:a", "pcm_s16le", out, "-y")
	return cmd.Run()
}

func runWhisper(opt Options, wav, outBase string) error {
	oflag := "-otxt"
	if opt.Format == "srt" {
		oflag = "-osrt"
	}
	cmd := exec.Command(opt.WhisperCliPath,
		"-m", opt.ModelPath,
		"-f", wav,
		"-l", opt.Lang,
		"-t", "8",
		"-et", "2.8",
		oflag,
		"-of", outBase,
	)
	return cmd.Run()
}
