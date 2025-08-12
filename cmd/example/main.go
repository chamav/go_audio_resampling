package main

import (
	"fmt"
	"math"

	"github.com/example/go_audio_resampling/resample"
)

func main() {
	inRate := 44100
	outRate := 48000

	r, err := resample.New(inRate, outRate)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	samples := make([]float32, inRate)
	for i := range samples {
		samples[i] = float32(math.Sin(2 * math.Pi * 440 * float64(i) / float64(inRate)))
	}

	out1, err := r.Convert(samples)
	if err != nil {
		panic(err)
	}
	out2, err := r.Convert(nil)
	if err != nil {
		panic(err)
	}
	out := append(out1, out2...)

	fmt.Printf("in samples: %d, out samples: %d\n", len(samples), len(out))
}
