package s3

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func SaveToS3(key string, filePath string, content string) error {
	accessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	accessSecret := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, accessSecret, "")), config.WithRegion(region))
	if err != nil {
		log.Printf("unable to load AWS SDK config, %v", err)
		return fmt.Errorf("unable to load AWS SDK config, %v", err)
	}

	client := s3.NewFromConfig(awsConfig)

	bucket := os.Getenv("S3_BUCKET")
	if bucket == "" {
		bucket = "replc"
	}
	destination := fmt.Sprintf("%s/%s", key, filePath)
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &destination,
		Body:   strings.NewReader(content),
	})

	return err
}
