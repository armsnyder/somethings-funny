package upload

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"time"
)

func UploadHandler(region, bucket *string) gin.HandlerFunc {
	s3c := s3.New(session.Must(session.NewSession(&aws.Config{Region: region})))
	return func(c *gin.Context) {
		filename := c.PostForm("filename")
		if filename == "" {
			_ = c.AbortWithError(400, errors.New("missing required field: filename"))
			return
		}
		data := c.PostForm("data")
		if data == "" {
			_ = c.AbortWithError(400, errors.New("missing required field: data"))
			return
		}
		var metadata map[string]*string
		meta := c.PostForm("meta")
		if meta != "" {
			metadata = make(map[string]*string)
			if err := json.Unmarshal([]byte(meta), &metadata); err != nil {
				_ = c.AbortWithError(400, errors.Wrap(err, "meta field is not valid json"))
				return
			}
		}
		contentType := c.PostForm("content_type")
		contentEncoding := c.PostForm("content_encoding")
		dataBytes, err := base64.RawURLEncoding.DecodeString(data)
		if err != nil {
			_ = c.AbortWithError(400, err)
			return
		}
		_ = c.Request.Body.Close()
		now := time.Now()
		_, err = s3c.PutObjectWithContext(c, &s3.PutObjectInput{
			Bucket: bucket,
			Key: aws.String(fmt.Sprintf("uploads/%d/%02d/%02d/%s",
				now.Year(), now.Month(), now.Day(), filename)),
			Body:            bytes.NewReader(dataBytes),
			ContentType:     aws.String(contentType),
			ContentEncoding: aws.String(contentEncoding),
			Metadata:        metadata,
		})
		if err != nil {
			_ = c.AbortWithError(500, err)
			return
		}
		c.AbortWithStatus(200)
	}
}
