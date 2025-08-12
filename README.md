# go_audio_resampling

Minimal Go binding to FFmpeg's [libswresample](https://ffmpeg.org/libswresample.html) for converting mono float32 audio between sample rates.

## Prerequisites

- Go 1.24 or later
- FFmpeg development files providing `libswresample` and `libavutil`
  - Debian/Ubuntu: `sudo apt-get install ffmpeg libswresample-dev libavutil-dev`

## Usage

```go
r, err := resample.New(44100, 48000)
if err != nil {
        // handle error
}
defer r.Close()

out, err := r.Convert(in)
// Flush any buffered audio by calling Convert(nil)
```

A complete example generating a sine wave and resampling it from 44.1 kHz to
48 kHz can be run with:

```sh
go run ./cmd/example
```

## Testing

```sh
go vet ./...
go test ./...
```

## License

MIT
