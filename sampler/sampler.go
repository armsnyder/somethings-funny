package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
	"time"

	"github.com/orcaman/writerseeker"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func chk(err error) {
	if err != nil {
		panic(err)
	}
}

type container struct {
	buf  []int16
	lock sync.Mutex
}

func bytesToInts(byteBuffer *[]byte, intBuffer *[]int16) {
	if len(*byteBuffer)%2 != 0 {
		panic("Bytes cannot be converted to ints because they are not a multiple of 2")
	}
	if len(*intBuffer) < len(*byteBuffer)/2 {
		panic("Not enough room in intBuffer")
	}
	for i := 0; i < len(*byteBuffer)/2; i++ {
		(*intBuffer)[i] = int16(binary.BigEndian.Uint16((*byteBuffer)[i*2 : (i+1)*2]))
	}
}

func intsToBytes(intBuffer *[]int16, byteBuffer *[]byte) {
	if len(*byteBuffer) < len(*intBuffer)*2 {
		panic("Not enough room in byteBuffer")
	}
	for i, v := range *intBuffer {
		binary.BigEndian.PutUint16((*byteBuffer)[i*2:(i+1)*2], uint16(v))
	}
}

func toWav(in *[]int16) *bytes.Reader {
	nSamples := len(*in)
	f := &writerseeker.WriterSeeker{}
	// form chunk
	_, err := f.Write([]byte("FORM"))
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(4+8+18+8+8+2*nSamples))) //total bytes
	_, err = f.Write([]byte("AIFF"))
	chk(err)

	// common chunk
	_, err = f.Write([]byte("COMM"))
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(18)))                  //size
	chk(binary.Write(f, binary.BigEndian, int16(1)))                   //channels
	chk(binary.Write(f, binary.BigEndian, int32(nSamples)))            //number of samples
	chk(binary.Write(f, binary.BigEndian, int16(16)))                  //bits per sample
	_, err = f.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}) //80-bit sample rate 44100
	chk(err)

	// sound chunk
	_, err = f.Write([]byte("SSND"))
	chk(err)
	chk(binary.Write(f, binary.BigEndian, int32(2*nSamples+8))) //size
	chk(binary.Write(f, binary.BigEndian, int32(0)))            //offset
	chk(binary.Write(f, binary.BigEndian, int32(0)))            //block

	byteBuff := make([]byte, nSamples*2)
	intsToBytes(in, &byteBuff)
	_, err = f.Write(byteBuff)
	chk(err)
	return f.BytesReader()
}

func upload(container *container) {
	defer func() {
		err := recover()
		if err != nil {
			log.Fatalln("Exiting with error", err)
		}
	}()
	fmt.Println("Initializing uploader")
	sess := session.Must(session.NewSession())
	sqsClient := sqs.New(sess)
	queueURL := aws.String(os.Getenv("SQS_URL"))
	uploader := s3manager.NewUploader(sess)
	fmt.Println("Listening for invocations")
	for {
		output, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{QueueUrl: queueURL, WaitTimeSeconds: aws.Int64(20)})
		chk(err)
		if len(output.Messages) > 0 {
			fmt.Println("Received invocation")
			container.lock.Lock()
			fmt.Println("Converting to WAV")
			sound := toWav(&container.buf)
			container.lock.Unlock()
			fmt.Println("Uploading WAV")
			_, err = uploader.Upload(&s3manager.UploadInput{
				Bucket: aws.String(os.Getenv("BUCKET_NAME")),
				Key:    aws.String(fmt.Sprintf("%d.wav", time.Now().Unix())),
				Body:   sound,
			})
			chk(err)
			fmt.Println("Deleting invocation from queue")
			entries := make([]*sqs.DeleteMessageBatchRequestEntry, len(output.Messages))
			for i, m := range output.Messages {
				entries[i] = &sqs.DeleteMessageBatchRequestEntry{Id: m.MessageId, ReceiptHandle: m.ReceiptHandle}
			}
			_, err = sqsClient.DeleteMessageBatch(&sqs.DeleteMessageBatchInput{QueueUrl: queueURL, Entries: entries})
			chk(err)
			fmt.Println("Invocation complete")
		}
	}
}

func main() {
	fmt.Println("Running")
	fmt.Println("Dialing unix socket")
	conn, err := net.Dial("unix", "/mic/mic.sock")
	chk(err)
	defer func() {
		fmt.Println("Closing socket")
		conn.Close()
	}()
	readBuffer := make([]byte, 512)
	soundBuffer := make([]int16, 256)
	container := &container{make([]int16, 44100*30), sync.Mutex{}}
	go upload(container)
	fmt.Println("Listening to sound")
	for {
		_, err := conn.Read(readBuffer)
		chk(err)
		bytesToInts(&readBuffer, &soundBuffer)
		container.lock.Lock()
		copy(container.buf, container.buf[len(soundBuffer):])
		copy(container.buf[len(container.buf)-len(soundBuffer):], soundBuffer)
		container.lock.Unlock()
	}
}
