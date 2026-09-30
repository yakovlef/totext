# totext

Local audio/video transcription to text via [whisper.cpp](https://github.com/ggml-org/whisper.cpp) — no cloud, runs on your machine's GPU/CPU.

```
totext lecture.mp4
totext --srt talk.mp3
totext -l en interview.wav
```

The output is a `.txt` file (or `.srt` with timestamps via `--srt`) next to the source file.

## Install

### macOS / Linux (Homebrew)

```bash
brew install yakovlef/totext/totext
```

Homebrew installs `ffmpeg` and `whisper-cpp` for you. totext downloads the whisper model (~1.5GB) itself on first run.

### Other ways

Download the binary for your OS from the [releases page](https://github.com/yakovlef/totext/releases) and put it on your `PATH`. On first run, `totext` checks dependencies and offers to install anything missing:

- macOS/Linux — via Homebrew, if available
- Linux without Homebrew — `ffmpeg` via apt/dnf/pacman; `whisper-cli` needs to be installed manually ([instructions](https://github.com/ggml-org/whisper.cpp#quick-start)) or via Homebrew
- Windows — `ffmpeg` via `winget`; `whisper-cli` needs to be built manually or run via WSL2 + Homebrew

Check dependency status any time:

```bash
totext doctor
```

## Flags

| Flag | Default | Description |
|---|---|---|
| `-l, --lang` | `ru` | transcription language |
| `--srt` | off | timestamped subtitles (`.srt`) instead of `.txt` |
| `--keep-audio` | off | keep the intermediate `.wav` file |
| `--model` | — | path to your own whisper.cpp ggml model |

The model path can also be set via `$TOTEXT_MODEL` or `$WHISPER_MODEL`.

## Build from source

```bash
git clone https://github.com/yakovlef/totext
cd totext
go build -o totext .
```

## License

MIT
