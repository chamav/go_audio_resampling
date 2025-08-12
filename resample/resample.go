// Package resample exposes a tiny wrapper around FFmpeg's libswresample
// (https://ffmpeg.org/libswresample.html). The binding currently supports
// resampling mono 32‑bit floating point audio between arbitrary input and
// output sample rates. Internally it uses SwrContext and only covers the
// minimal operations required by the example and tests.
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

// Resampler converts audio from one sample rate to another. It assumes both
// the input and output are mono samples encoded as 32‑bit floats in native
// endianness.
type Resampler struct {
	ctx     *C.struct_SwrContext
	inRate  int
	outRate int
}


// New allocates and initializes a Resampler that converts from inRate to
// outRate. The caller must call Close when finished with the resampler.
func New(inRate, outRate int) (*Resampler, error) {
	ctx := C.swr_alloc_set_opts(nil,
		C.long(C.AV_CH_LAYOUT_MONO), C.AV_SAMPLE_FMT_FLT, C.int(outRate),
		C.long(C.AV_CH_LAYOUT_MONO), C.AV_SAMPLE_FMT_FLT, C.int(inRate),
		0, nil) // deprecated but simpler than swr_alloc_set_opts2
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


// Convert resamples the provided mono float32 slice and returns the converted
// data. Passing a nil or empty slice flushes any buffered samples inside the
// SwrContext. After Close is called, Convert returns an error.
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
