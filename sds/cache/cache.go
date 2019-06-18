package cache

import (
	"bytes"
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"golang.org/x/crypto/acme/autocert"
	"io/ioutil"
)

const prefix = "envoy-letsencrypt/cache/"

type s3Cache struct {
	bucket         *string
	region         *string
	letsEncryptUrl string
	s3c            *s3.S3
}

func (this *s3Cache) Get(ctx context.Context, key string) ([]byte, error) {
	pvtKeyRes, err := this.s3c.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: this.bucket,
		Key:    aws.String(this.createFullKey(key)),
	})

	if err != nil {
		if err.(awserr.Error).Code() == "NoSuchKey" {
			return nil, autocert.ErrCacheMiss
		} else {
			return nil, err
		}
	}

	body, err := ioutil.ReadAll(pvtKeyRes.Body)
	if err != nil {
		panic(err)
	}

	return body, nil
}

func (this *s3Cache) Put(ctx context.Context, key string, data []byte) error {
	_, err := this.s3c.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:               this.bucket,
		Key:                  aws.String(this.createFullKey(key)),
		Body:                 bytes.NewReader(data),
		ServerSideEncryption: aws.String(s3.ServerSideEncryptionAwsKms),
	})
	return err
}

func (this *s3Cache) Delete(ctx context.Context, key string) error {
	_, err := this.s3c.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: this.bucket,
		Key:    aws.String(this.createFullKey(key)),
	})
	return err
}

func (this *s3Cache) createFullKey(key string) string {
	return prefix + this.letsEncryptUrl + "/" + key
}

func New(region, bucket *string, letsEncryptUrl string) autocert.Cache {
	sess := session.Must(session.NewSession(&aws.Config{Region: region}))
	s3c := s3.New(sess)
	return &s3Cache{
		s3c:            s3c,
		bucket:         bucket,
		region:         region,
		letsEncryptUrl: letsEncryptUrl,
	}
}
