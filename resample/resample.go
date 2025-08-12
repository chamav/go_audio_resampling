// Package resample provides a minimal CGO binding to FFmpeg's
// libswresample for converting mono float32 audio between sample rates.
// See https://ffmpeg.org/doxygen/trunk/group__lswr.html for library docs.
package resample

/*
#cgo pkg-config: libswresample libavutil
#include <libswresample/swresample.h>
#include <libavutil/channel_layout.h>
#include <libavutil/samplefmt.h>
#include <libavutil/opt.h>
#include <libavutil/avutil.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"unsafe"
)

// Resampler wraps FFmpeg's SwrContext for mono float32 audio.
type Resampler struct {
	ctx     *C.struct_SwrContext
	inRate  int
	outRate int
}

// New creates a new Resampler converting from inRate to outRate.
func New(inRate, outRate int) (*Resampler, error) {
	ctx := C.swr_alloc_set_opts(nil,
		C.long(C.AV_CH_LAYOUT_MONO), C.AV_SAMPLE_FMT_FLT, C.int(outRate),
		C.long(C.AV_CH_LAYOUT_MONO), C.AV_SAMPLE_FMT_FLT, C.int(inRate),
		0, nil)
	if ctx == nil {
		return nil, errors.New("swr_alloc_set_opts failed")
	}
	if ret := C.swr_init(ctx); ret < 0 {
		C.swr_free(&ctx)
		return nil, errors.New("swr_init failed")
	}
	return &Resampler{ctx: ctx, inRate: inRate, outRate: outRate}, nil
}

// Close releases the underlying SwrContext.
func (r *Resampler) Close() {
	if r.ctx != nil {
		C.swr_free(&r.ctx)
		r.ctx = nil
	}
}

// Convert resamples the input mono float32 slice. If in is nil or empty,
// remaining buffered samples are flushed.
func (r *Resampler) Convert(in []float32) ([]float32, error) {
	if r.ctx == nil {
		return nil, errors.New("nil context")
	}

	inSamples := C.int(len(in))
	var inBuf unsafe.Pointer
	if len(in) > 0 {
		inBuf = C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
		defer C.free(inBuf)
		*(**C.uint8_t)(inBuf) = (*C.uint8_t)(unsafe.Pointer(&in[0]))
	}

	delay := C.swr_get_delay(r.ctx, C.int64_t(r.inRate)) + C.int64_t(inSamples)
	outSamples := int(C.av_rescale_rnd(delay, C.int64_t(r.outRate), C.int64_t(r.inRate), C.AV_ROUND_UP))
	out := make([]float32, outSamples)

	outBuf := C.malloc(C.size_t(unsafe.Sizeof(uintptr(0))))
	defer C.free(outBuf)
	*(**C.uint8_t)(outBuf) = (*C.uint8_t)(unsafe.Pointer(&out[0]))

	var inArg **C.uint8_t
	if len(in) > 0 {
		inArg = (**C.uint8_t)(inBuf)
	}

	ret := C.swr_convert(r.ctx,
		(**C.uint8_t)(outBuf), C.int(outSamples),
		inArg, inSamples)
	if ret < 0 {
		return nil, errors.New("swr_convert failed")
	}
	return out[:ret], nil
}
