package controller

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	Cfg "github.com/3ssalunke/replc/config"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Storage struct {
	Client *s3.Client
}

func NewS3Storage(cfg Cfg.Config) (*S3Storage, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO(), config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3.Accesskey, cfg.S3.Secretkey, "")), config.WithRegion(cfg.S3.Region))
	if err != nil {
		log.Printf("unable to load AWS SDK config, %v", err)
		return nil, fmt.Errorf("unable to load AWS SDK config, %v", err)
	}

	client := s3.NewFromConfig(awsConfig)

	return &S3Storage{
		Client: client,
	}, nil
}

func (c *Controller) CopyObjects(language string, replId string, storage *S3Storage) error {
	bucket := "replc"
	sourceFolder := fmt.Sprintf("boilerplates/%s", language)
	destinationFolder := fmt.Sprintf("replcs/%s", replId)

	return listObjects(storage.Client, bucket, sourceFolder, destinationFolder)
}

func listObjects(client *s3.Client, bucket, sourceFolder, destinationFolder string) error {
	paginator := s3.NewListObjectsV2Paginator(client, &s3.ListObjectsV2Input{
		Bucket: &bucket,
		Prefix: &sourceFolder,
	})

	// Wait group to synchronize Go routines
	var wg sync.WaitGroup
	errCh := make(chan error)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.Background())
		if err != nil {
			log.Printf("unable to list objects: %v", err)
			return fmt.Errorf("unable to list objects: %v", err)
		}

		for _, obj := range page.Contents {
			// Get the key (file path) relative to the source folder
			key := strings.TrimPrefix(*obj.Key, sourceFolder)

			// If the object is a folder, recursively list objects inside it
			if strings.HasSuffix(key, "/") {
				err := listObjects(client, bucket, *obj.Key, destinationFolder)
				if err != nil {
					return err
				}
				continue
			}

			wg.Add(1)
			// Copy the object to the destination folder maintaining the directory structure
			go func(obj types.Object) {
				defer wg.Done()

				source := fmt.Sprintf("%s/%s", bucket, *obj.Key)
				destination := fmt.Sprintf("%s%s", destinationFolder, key)
				_, err := client.CopyObject(context.TODO(), &s3.CopyObjectInput{
					Bucket:     &bucket,
					CopySource: &source,
					Key:        &destination,
				})
				if err != nil {
					log.Printf("unable to copy object %s: %v", *obj.Key, err)
					errCh <- fmt.Errorf("unable to copy object %s: %v", *obj.Key, err)
				}

				log.Printf("Object copied: %s\n", *obj.Key)
			}(obj)
		}
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return err
	}

	return nil
}
