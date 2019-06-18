package upload

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"os"
	"somethings-funny/common"
	"somethings-funny/sampler/filter"
	"strings"
	"time"
)

const (
	bitsPerByte        = 8
	httpTimeoutSeconds = 30
)

func Upload(ctx context.Context, uploadUrl, token string, in <-chan *filter.WavFile) error {
	bytesPerSample := common.BitDepth / bitsPerByte
	client := &http.Client{
		Timeout: time.Second * httpTimeoutSeconds,
	}
	hostname, _ := os.Hostname()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case f := <-in:
			now := time.Now()
			gzipped, err := gzipData(f.Data)
			if err != nil {
				return err
			}
			duration := fmt.Sprintf("%d sec", len(f.Data)/bytesPerSample/common.SampleRate)
			formData := url.Values(map[string][]string{
				"filename":         {now.Format(time.RFC3339) + ".wav.gz"},
				"content_encoding": {"gzip"},
				"content_type":     {"audio/wav"},
				"data":             {gzipped},
				"meta": {fmt.Sprintf(
					`{"amplitude": "%f", "uploader-hostname": "%s", "duration": "%s"}`,
					f.Amplitude, hostname, duration)},
			})
			req, err := http.NewRequest("POST", uploadUrl, strings.NewReader(formData.Encode()))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Authorization", "Bearer "+token)
			res, err := client.Do(req)
			if err != nil {
				return err
			}

			logrus.Infof("Uploaded. Status: %s | Duration: %s | Amplitude: %f | "+
				"Bytes before compress: %d | Characters in data object: %d",
				res.Status, duration, f.Amplitude, len(f.Data), len(gzipped))
		}
	}
}

func gzipData(data []byte) (string, error) {
	buf := new(bytes.Buffer)
	b64 := base64.NewEncoder(base64.RawURLEncoding, buf)
	gz := gzip.NewWriter(b64)
	_, err := gz.Write(data)
	_ = gz.Close()
	_ = b64.Close()
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}
