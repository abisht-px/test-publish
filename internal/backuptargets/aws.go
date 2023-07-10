package backuptargets

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type AwsS3StorageProvider struct {
	bucketName string
	region     string
	accessKey  string
	secretKey  string
}

func NewAwsS3StorageProvider(bucketName, region, accessKey, secretKey string) *AwsS3StorageProvider {
	return &AwsS3StorageProvider{
		bucketName: bucketName,
		region:     region,
		accessKey:  accessKey,
		secretKey:  secretKey,
	}
}

func (p *AwsS3StorageProvider) GetClient() (*s3.S3, error) {
	s3Credentials := credentials.NewStaticCredentials(p.accessKey, p.secretKey, "")
	awsConfig := aws.NewConfig().WithCredentials(s3Credentials).WithRegion(p.region)
	sess, err := session.NewSession(awsConfig)
	if err != nil {
		return nil, fmt.Errorf("error creating new session for s3 access: %w", err)
	}
	return s3.New(sess), nil
}

func (p *AwsS3StorageProvider) ListObjectsWithPrefix(prefixKey string) (*s3.ListObjectsV2Output, error) {
	s3Client, err := p.GetClient()
	if err != nil {
		return nil, fmt.Errorf("error getting s3 client: %w", err)
	}
	resp, err := s3Client.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(p.bucketName),
		Prefix: aws.String(prefixKey),
	})
	if err != nil {
		return nil, fmt.Errorf("error listing objects from s3 bucket: %w", err)
	}
	return resp, nil
}
