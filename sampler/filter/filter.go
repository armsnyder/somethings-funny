package filter

import (
	"context"
	"encoding/binary"
	"github.com/armon/circbuf"
	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/orcaman/writerseeker"
	"io/ioutil"
	"math"
	"somethings-funny/common"
	"time"
)

const (
	wavFileLenSeconds = 5
	debounceSeconds   = 1
	bitsPerByte       = 8
)

type WavFile struct {
	Data      []byte
	Amplitude float64
}

func AggregateFilter(ctx context.Context, threshold float64, in <-chan []byte, out chan<- *WavFile,
) error {
	bytesPerSample := common.BitDepth / bitsPerByte
	bigBuff, err := circbuf.NewBuffer(int64(common.SampleRate * bytesPerSample * wavFileLenSeconds))
	if err != nil {
		return err
	}
	start := int64(0)
	end := int64(0)
	maxAmp := float64(0)
	finalSoundBuffer := newIntBuffer(common.SampleRate * wavFileLenSeconds)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case data := <-in:
			_, err = bigBuff.Write(data)
			if err != nil {
				return err
			}

			now := time.Now().Unix()

			// TODO: Recycle soundBuffer
			soundBuffer := make([]int, len(data)/bytesPerSample)
			bytesToInts(data, soundBuffer)
			amp := calcAmplitude(soundBuffer)
			if amp > threshold {
				if amp > maxAmp {
					maxAmp = amp
				}
				if start == 0 {
					start = now
				}
				end = now
			}

			if end != 0 && start != 0 &&
				(now-end > debounceSeconds || now-start > wavFileLenSeconds-debounceSeconds) {

				seeker := new(writerseeker.WriterSeeker)
				enc := wav.NewEncoder(seeker, common.SampleRate,
					common.BitDepth, 1, 1)
				bytesToInts(bigBuff.Bytes(), finalSoundBuffer.Data)
				if err := enc.Write(finalSoundBuffer); err != nil {
					return err
				}
				if err := enc.Close(); err != nil {
					return err
				}
				b, err := ioutil.ReadAll(seeker.BytesReader())
				if err != nil {
					return err
				}
				out <- &WavFile{
					Data:      b,
					Amplitude: maxAmp,
				}
				start = 0
				end = 0
				maxAmp = 0
			}
		}
	}
}

func bytesToInts(byteBuffer []byte, intBuffer []int) {
	bytesPerSample := common.BitDepth / bitsPerByte
	if len(byteBuffer)%bytesPerSample != 0 {
		panic("Bytes cannot be converted to ints")
	}
	if len(intBuffer) < len(byteBuffer)/bytesPerSample {
		panic("Not enough room in intBuffer")
	}
	for i := 0; i < len(byteBuffer)/bytesPerSample; i++ {
		intBuffer[i] = int(int16(binary.BigEndian.Uint16(
			byteBuffer[i*bytesPerSample : (i+1)*bytesPerSample])))
	}
}

func newIntBuffer(size int) *audio.IntBuffer {
	return &audio.IntBuffer{
		Format: &audio.Format{
			NumChannels: 1,
			SampleRate:  common.SampleRate,
		},
		SourceBitDepth: common.BitDepth,
		Data:           make([]int, size),
	}
}

func calcAmplitude(intBuffer []int) float64 {
	if len(intBuffer) == 0 {
		return 0
	}
	maxAmpFromWaveform := 0
	for _, i := range intBuffer {
		amp := i
		if amp < 0 {
			amp = -amp
		}
		if amp > maxAmpFromWaveform {
			maxAmpFromWaveform = amp
		}
	}
	maxPossibleAmp := math.Pow(2, float64(common.BitDepth-1))
	return float64(maxAmpFromWaveform) / maxPossibleAmp
}
