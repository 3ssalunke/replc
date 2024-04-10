package s3

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func SaveToS3(key string, filePath string, content string) error {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("", "", "")), config.WithRegion(""))
	if err != nil {
		log.Printf("unable to load AWS SDK config, %v", err)
		return fmt.Errorf("unable to load AWS SDK config, %v", err)
	}

	client := s3.NewFromConfig(awsConfig)

	bucket := "replc"
	destination := fmt.Sprintf("%s/%s", key, filePath)
	client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &bucket,
		Key: &destination,
		Body: ,
	})
}
