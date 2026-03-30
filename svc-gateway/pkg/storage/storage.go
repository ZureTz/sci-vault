package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3            *s3.Client
	privateBucket string
	publicBucket  string
}

func NewClient(endpoint, accessKey, secretKey, privateBucket, publicBucket string, useSSL bool) *Client {
	cfg := aws.Config{
		Region: "ap-east-1",
		Credentials: aws.NewCredentialsCache(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	slog.Info("connected to rustfs storage", "endpoint", endpoint,
		"privateBucket", privateBucket, "publicBucket", publicBucket)

	return &Client{
		s3:            client,
		privateBucket: privateBucket,
		publicBucket:  publicBucket,
	}
}

func (c *Client) bucket(isPrivate bool) string {
	if isPrivate {
		return c.privateBucket
	}
	return c.publicBucket
}

// EnsureBuckets creates both buckets if they don't exist and sets the public bucket policy.
func (c *Client) EnsureBuckets(ctx context.Context) error {
	if err := c.ensureBucket(ctx, c.privateBucket); err != nil {
		return fmt.Errorf("private bucket %q: %w", c.privateBucket, err)
	}
	if err := c.ensureBucket(ctx, c.publicBucket); err != nil {
		return fmt.Errorf("public bucket %q: %w", c.publicBucket, err)
	}
	if err := c.setPublicReadPolicy(ctx, c.publicBucket); err != nil {
		return fmt.Errorf("set public policy on %q: %w", c.publicBucket, err)
	}
	return nil
}

func (c *Client) ensureBucket(ctx context.Context, name string) error {
	_, err := c.s3.HeadBucket(ctx, &s3.HeadBucketInput{
		Bucket: aws.String(name),
	})
	if err == nil {
		return nil
	}
	_, err = c.s3.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(name),
	})
	return err
}

func (c *Client) setPublicReadPolicy(ctx context.Context, name string) error {
	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/*", name)},
			},
		},
	}
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return err
	}
	_, err = c.s3.PutBucketPolicy(ctx, &s3.PutBucketPolicyInput{
		Bucket: aws.String(name),
		Policy: aws.String(string(policyJSON)),
	})
	return err
}

func (c *Client) PutObject(ctx context.Context, key string, body io.Reader, contentType string, isPrivate bool) error {
	input := &s3.PutObjectInput{
		Bucket: aws.String(c.bucket(isPrivate)),
		Key:    aws.String(key),
		Body:   body,
	}
	if contentType != "" {
		input.ContentType = aws.String(contentType)
	}
	_, err := c.s3.PutObject(ctx, input)
	return err
}

func (c *Client) GetObject(ctx context.Context, key string, isPrivate bool) (io.ReadCloser, error) {
	output, err := c.s3.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucket(isPrivate)),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

func (c *Client) DeleteObject(ctx context.Context, key string, isPrivate bool) error {
	_, err := c.s3.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket(isPrivate)),
		Key:    aws.String(key),
	})
	return err
}

func (c *Client) ListObjects(ctx context.Context, prefix string, isPrivate bool) ([]string, error) {
	output, err := c.s3.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucket(isPrivate)),
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, err
	}
	keys := make([]string, 0, len(output.Contents))
	for _, obj := range output.Contents {
		keys = append(keys, *obj.Key)
	}
	return keys, nil
}

func (c *Client) PrivateBucket() string { return c.privateBucket }
func (c *Client) PublicBucket() string  { return c.publicBucket }
func (c *Client) S3() *s3.Client        { return c.s3 }

// PublicObjectURL returns the proxy URL for an object in the public bucket.
// Nginx/Vite proxies /assets/ to the rustfs public bucket.
func (c *Client) PublicObjectURL(key string) string {
	return fmt.Sprintf("/assets/%s", key)
}

// PrivateObjectURL returns a presigned proxy URL for an object in the private bucket.
// The S3 host and bucket prefix are stripped and replaced with /private/, which
// Nginx/Vite proxies to the rustfs private bucket. The presign signature is preserved
// so rustfs can still validate the request.
// filename sets the Content-Disposition header so browsers save the file with that name.
func (c *Client) PrivateObjectURL(ctx context.Context, key string, expiry time.Duration, filename string) (string, error) {
	presignClient := s3.NewPresignClient(c.s3)
	input := &s3.GetObjectInput{
		Bucket: aws.String(c.privateBucket),
		Key:    aws.String(key),
	}
	if filename != "" {
		input.ResponseContentDisposition = aws.String(fmt.Sprintf(`attachment; filename="%s"`, filename))
	}
	req, err := presignClient.PresignGetObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})
	if err != nil {
		return "", err
	}
	u, err := url.Parse(req.URL)
	if err != nil {
		return "", fmt.Errorf("failed to parse presigned URL: %w", err)
	}
	// Strip /{bucket} prefix from path, replace with /private
	path := strings.TrimPrefix(u.Path, "/"+c.privateBucket)
	return "/private" + path + "?" + u.RawQuery, nil
}
