package s3

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// Client handles S3 operations
type Client struct {
	bucket   string
	region   string
	uploader *s3manager.Uploader
}

// NewClient creates a new S3 client
// Credentials are loaded from environment variables or AWS credentials file
// AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN (optional)
func NewClient(bucket, region string) (*Client, error) {
	// Create an AWS session using the SDK's default credential chain, which
	// resolves credentials in order from:
	//   1. Environment variables (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, AWS_SESSION_TOKEN)
	//   2. The shared credentials/config files (~/.aws/credentials, ~/.aws/config)
	//   3. An IAM role (ECS task role, EC2 instance profile, etc.)
	// SharedConfigEnable ensures profiles in the shared config file are honored.
	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create AWS session: %w", err)
	}

	uploader := s3manager.NewUploader(sess)

	return &Client{
		bucket:   bucket,
		region:   region,
		uploader: uploader,
	}, nil
}

// UploadImage uploads an image to S3
func (c *Client) UploadImage(filename string, data []byte, contentType string) error {
	// Ensure filename doesn't start with /
	filename = strings.TrimPrefix(filename, "/")

	// Create upload input
	input := &s3manager.UploadInput{
		Bucket:      aws.String(c.bucket),
		Key:         aws.String(fmt.Sprintf("memes/%s", filename)),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
		ACL:         aws.String("public-read"), // Make images publicly readable
		Metadata: map[string]*string{
			"source": aws.String("giggles-ai-crawler"),
		},
	}

	// Upload to S3
	_, err := c.uploader.Upload(input)
	if err != nil {
		return fmt.Errorf("failed to upload to S3: %w", err)
	}

	return nil
}

// GetPublicURL returns the public URL for an uploaded image
func (c *Client) GetPublicURL(filename string) string {
	filename = strings.TrimPrefix(filename, "/")
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/memes/%s", c.bucket, c.region, filename)
}

// LoadCredentialsFromFile loads credentials from a file (for local development)
// Format: key=value pairs, one per line
// Example:
//
//	AWS_ACCESS_KEY_ID=your_key
//	AWS_SECRET_ACCESS_KEY=your_secret
func LoadCredentialsFromFile(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read credentials file: %w", err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		os.Setenv(key, value)
	}

	return nil
}
