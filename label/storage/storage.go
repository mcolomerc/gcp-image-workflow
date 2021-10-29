package storage

import (
	"context"
	"fmt"
	"log"

	"cloud.google.com/go/storage"
)

type Storage struct {
	client *storage.Client
}

func New() *Storage {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
		return nil
	}

	return &Storage{
		client: client,
	}
}

// copyFile copies an object into specified bucket.
func (str *Storage) CopyFile(srcBucket, srcObject, dstBucket, dstObject string) error {
	ctx := context.Background()

	src := str.client.Bucket(srcBucket).Object(srcObject)
	dst := str.client.Bucket(dstBucket).Object(dstObject)
	if _, err := dst.CopierFrom(src).Run(ctx); err != nil {
		return fmt.Errorf("Object(%q).CopierFrom(%q).Run: %v", dstObject, srcObject, err)
	}
	return nil
}
