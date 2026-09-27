package storage

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	PrefixJadwal  = "jadwal"
	PrefixEkspor  = "ekspor"
	PrefixDokumen = "dokumen"
)

type Options struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type Client struct {
	api    *minio.Client
	bucket string
}

func New(opts Options) (*Client, error) {
	if opts.Endpoint == "" || opts.AccessKey == "" || opts.SecretKey == "" || opts.Bucket == "" {
		return nil, fmt.Errorf("konfigurasi minio belum lengkap")
	}
	api, err := minio.New(opts.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(opts.AccessKey, opts.SecretKey, ""),
		Secure: opts.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("klien minio: %w", err)
	}
	return &Client{api: api, bucket: opts.Bucket}, nil
}

func (c *Client) EnsureBucket(ctx context.Context) error {
	ada, err := c.api.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("cek bucket minio: %w", err)
	}
	if ada {
		return nil
	}
	if err := c.api.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("buat bucket minio: %w", err)
	}
	return nil
}

func (c *Client) Put(ctx context.Context, key string, body io.Reader, size int64, contentType string) error {
	if _, err := ParseKey(key); err != nil {
		return err
	}
	if size < 0 {
		return fmt.Errorf("ukuran objek tidak valid")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := c.api.PutObject(ctx, c.bucket, key, body, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("unggah objek minio: %w", err)
	}
	return nil
}

func (c *Client) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	if _, err := ParseKey(key); err != nil {
		return nil, err
	}
	if _, err := c.api.StatObject(ctx, c.bucket, key, minio.StatObjectOptions{}); err != nil {
		return nil, fmt.Errorf("ambil objek minio: %w", err)
	}
	obj, err := c.api.GetObject(ctx, c.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("ambil objek minio: %w", err)
	}
	return obj, nil
}

func (c *Client) Presign(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if _, err := ParseKey(key); err != nil {
		return "", err
	}
	if expiry <= 0 {
		return "", fmt.Errorf("masa tautan unduh wajib lebih dari nol")
	}
	tautan, err := c.api.PresignedGetObject(ctx, c.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("tautan unduh minio: %w", err)
	}
	return tautan.String(), nil
}

func Key(prefix, name string) (string, error) {
	prefix = strings.Trim(prefix, "/")
	name = strings.Trim(name, "/")
	if prefix == "" || name == "" {
		return "", fmt.Errorf("kunci objek wajib punya prefix dan nama")
	}
	if strings.Contains(prefix, "..") || strings.Contains(name, "..") || strings.Contains(name, "/") {
		return "", fmt.Errorf("kunci objek tidak valid")
	}
	return prefix + "/" + name, nil
}

func ParseKey(key string) (string, error) {
	key = strings.Trim(key, "/")
	bagian := strings.Split(key, "/")
	if len(bagian) != 2 {
		return "", fmt.Errorf("kunci objek wajib berbentuk prefix/nama")
	}
	return Key(bagian[0], bagian[1])
}
