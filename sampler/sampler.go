package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/orcaman/writerseeker"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/gordonklaus/portaudio"
)

var (
	sampleRate  = 44100
	bitDepth    = 16
	channels    = 1
	audioFormat = 1 // PCM
)

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("Initializing")
	portaudio.Initialize()
	defer portaudio.Terminate()
	fmt.Println("Devices")
	info, err := portaudio.Devices()
	chk(err)
	for _, inf := range info {
		fmt.Println(*inf)
	}
	fmt.Println("Default input")
	info2, err := portaudio.DefaultInputDevice()
	chk(err)
	fmt.Println(*info2)

	buffer := make([]int32, 262144)
	fmt.Println("Opening stream")
	stream, err := portaudio.OpenDefaultStream(channels, 0, float64(sampleRate), len(buffer), buffer)
	chk(err)
	defer stream.Close()
	fmt.Println("Starting stream")
	chk(stream.Start())
	fmt.Println("Reading stream")
	chk(stream.Read())
	fmt.Println("Stopping stream")
	chk(stream.Stop())
	fmt.Println("Done")
	sound := soundBufferToWavReader(&buffer)

	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("BUCKET_NAME")),
		Key:    aws.String(fmt.Sprintf("%d.wav", time.Now().Unix())),
		Body:   sound,
	})
	chk(err)
	for {
		fmt.Println("Hello world")
		time.Sleep(2 * time.Second)
	}
}

func soundBufferToWavReader(soundBuffer *[]int32) *bytes.Reader {
	nSamples := len(*soundBuffer)
	f := &writerseeker.WriterSeeker{}
	// form chunk
	_, err := f.Write([]byte("FORM"))
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(4+8+18+8+8+4*nSamples))) //total bytes
	_, err = f.Write([]byte("AIFF"))
	chk(err)

	// common chunk
	_, err = f.Write([]byte("COMM"))
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(18)))                  //size
	chk(binary.Write(f, binary.BigEndian, int16(1)))                   //channels
	chk(binary.Write(f, binary.BigEndian, int32(nSamples)))            //number of samples
	chk(binary.Write(f, binary.BigEndian, int16(32)))                  //bits per sample
	_, err = f.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}) //80-bit sample rate 44100
	chk(err)

	// sound chunk
	_, err = f.Write([]byte("SSND"))
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(4*nSamples+8))) //size
	chk(binary.Write(f, binary.BigEndian, int32(0)))            //offset
	chk(binary.Write(f, binary.BigEndian, int32(0)))            //block
	chk(binary.Write(f, binary.BigEndian, soundBuffer))
	return f.BytesReader()
}
