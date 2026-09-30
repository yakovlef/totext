// Package deps detects and, with confirmation, installs totext's runtime
// dependencies: ffmpeg, whisper-cli (whisper.cpp) and a ggml model file.
package deps

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const modelURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main/ggml-large-v3-turbo.bin"
const modelFileName = "ggml-large-v3-turbo.bin"

type Status struct {
	FFmpegPath     string
	WhisperCliPath string
	ModelPath      string
}

func (s Status) OK() bool {
	return s.FFmpegPath != "" && s.WhisperCliPath != "" && s.ModelPath != ""
}

// Check looks for ffmpeg, whisper-cli and a model file without installing anything.
// modelOverride, if non-empty, is used verbatim as the model path (must exist).
func Check(modelOverride string) Status {
	var s Status
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		s.FFmpegPath = p
	}
	if p, err := exec.LookPath("whisper-cli"); err == nil {
		s.WhisperCliPath = p
	}
	s.ModelPath = findModel(modelOverride)
	return s
}

func findModel(override string) string {
	if override != "" {
		if fileExists(override) {
			return override
		}
		return ""
	}
	if p := os.Getenv("TOTEXT_MODEL"); p != "" && fileExists(p) {
		return p
	}
	if p := os.Getenv("WHISPER_MODEL"); p != "" && fileExists(p) {
		return p
	}
	if home, err := os.UserHomeDir(); err == nil {
		legacy := filepath.Join(home, "whisper-models", modelFileName)
		if fileExists(legacy) {
			return legacy
		}
	}
	if managed := managedModelPath(); fileExists(managed) {
		return managed
	}
	return ""
}

func managedModelPath() string {
	dir, err := os.UserCacheDir()
	if err != nil {
		dir = os.TempDir()
	}
	return filepath.Join(dir, "totext", modelFileName)
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

// Ensure prompts the user once for confirmation and then installs whatever
// is missing from Status, using the best available method for the current OS.
func Ensure(s Status) (Status, error) {
	if s.OK() {
		return s, nil
	}

	var missing []string
	if s.FFmpegPath == "" {
		missing = append(missing, "ffmpeg")
	}
	if s.WhisperCliPath == "" {
		missing = append(missing, "whisper-cli (whisper.cpp)")
	}
	if s.ModelPath == "" {
		missing = append(missing, "model "+modelFileName+" (~1.5GB)")
	}

	fmt.Fprintf(os.Stderr, "totext: missing dependencies: %s\n", strings.Join(missing, ", "))
	if !confirm("Install automatically?") {
		return s, fmt.Errorf("dependencies not installed, aborting")
	}

	if s.FFmpegPath == "" || s.WhisperCliPath == "" {
		if err := installBinaries(s); err != nil {
			return s, err
		}
	}
	if s.ModelPath == "" {
		target := managedModelPath()
		if err := downloadModel(target); err != nil {
			return s, err
		}
		s.ModelPath = target
	}

	fresh := Check("")
	if fresh.ModelPath == "" {
		fresh.ModelPath = s.ModelPath
	}
	return fresh, nil
}

func confirm(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N] ", prompt)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return false
	}
	line = strings.ToLower(strings.TrimSpace(line))
	return line == "y" || line == "yes"
}

func installBinaries(s Status) error {
	switch runtime.GOOS {
	case "darwin":
		return installViaBrew(s)
	case "linux":
		if _, err := exec.LookPath("brew"); err == nil {
			return installViaBrew(s)
		}
		return installLinuxFallback(s)
	case "windows":
		return installWindows(s)
	default:
		return fmt.Errorf("unsupported OS: %s, install ffmpeg and whisper-cli manually", runtime.GOOS)
	}
}

func installViaBrew(s Status) error {
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("brew not found, install Homebrew (https://brew.sh) or install ffmpeg/whisper-cpp manually")
	}
	var pkgs []string
	if s.FFmpegPath == "" {
		pkgs = append(pkgs, "ffmpeg")
	}
	if s.WhisperCliPath == "" {
		pkgs = append(pkgs, "whisper-cpp")
	}
	if len(pkgs) == 0 {
		return nil
	}
	return runVisible("brew", append([]string{"install"}, pkgs...)...)
}

func installLinuxFallback(s Status) error {
	if s.FFmpegPath == "" {
		mgr, args := linuxPackageManager("ffmpeg")
		if mgr == "" {
			fmt.Fprintln(os.Stderr, "totext: no supported package manager found, install ffmpeg manually")
		} else if err := runVisible(mgr, args...); err != nil {
			return err
		}
	}
	if s.WhisperCliPath == "" {
		fmt.Fprintln(os.Stderr, "totext: automatic whisper-cli install is not supported on Linux without Homebrew.")
		fmt.Fprintln(os.Stderr, "Install Homebrew (https://brew.sh) and rerun totext doctor, or build whisper.cpp manually:")
		fmt.Fprintln(os.Stderr, "  https://github.com/ggml-org/whisper.cpp#quick-start")
	}
	return nil
}

func linuxPackageManager(pkg string) (string, []string) {
	switch {
	case commandExists("apt-get"):
		return "sudo", []string{"apt-get", "install", "-y", pkg}
	case commandExists("dnf"):
		return "sudo", []string{"dnf", "install", "-y", pkg}
	case commandExists("pacman"):
		return "sudo", []string{"pacman", "-S", "--noconfirm", pkg}
	default:
		return "", nil
	}
}

func installWindows(s Status) error {
	if s.FFmpegPath == "" {
		if commandExists("winget") {
			if err := runVisible("winget", "install", "-e", "--id", "Gyan.FFmpeg"); err != nil {
				return err
			}
		} else {
			fmt.Fprintln(os.Stderr, "totext: winget not found, install ffmpeg manually (https://ffmpeg.org/download.html)")
		}
	}
	if s.WhisperCliPath == "" {
		fmt.Fprintln(os.Stderr, "totext: automatic whisper-cli install is not supported on Windows.")
		fmt.Fprintln(os.Stderr, "Easiest option: install WSL2 + Homebrew and run totext there, or build whisper.cpp yourself:")
		fmt.Fprintln(os.Stderr, "  https://github.com/ggml-org/whisper.cpp#quick-start")
	}
	return nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runVisible(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func downloadModel(target string) error {
	if !confirm(fmt.Sprintf("Download the model (~1.5GB) to %s?", target)) {
		return fmt.Errorf("model not downloaded, aborting")
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	tmp := target + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(modelURL)
	if err != nil {
		os.Remove(tmp)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		os.Remove(tmp)
		return fmt.Errorf("model download failed: HTTP %d", resp.StatusCode)
	}

	fmt.Fprintln(os.Stderr, "totext: downloading model, this will take a couple of minutes...")
	if _, err := io.Copy(out, resp.Body); err != nil {
		os.Remove(tmp)
		return err
	}
	out.Close()
	return os.Rename(tmp, target)
}
