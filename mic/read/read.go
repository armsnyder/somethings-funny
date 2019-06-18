package read

import (
	"context"
	"encoding/binary"
)

func Read(ctx context.Context, input <-chan []int16, output chan<- []byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case inp := <-input:
			select {
			case output <- intsToBytes(inp):
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func intsToBytes(intBuffer []int16) []byte {
	output := make([]byte, len(intBuffer)*2)
	for i, v := range intBuffer {
		binary.BigEndian.PutUint16(output[i*2:(i+1)*2], uint16(v))
	}
	return output
}
