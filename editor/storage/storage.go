package storage

import (
	"context"
	"io/ioutil"
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

// UploadObject ...
func (str *Storage) UploadObject(bucket string, obj []byte, mtype string, name string) (string, error) {
	ctx := context.Background()
	wc := str.client.Bucket(bucket).Object(name).NewWriter(ctx)
	wc.ContentType = mtype
	if _, err := wc.Write(obj); err != nil {
		log.Fatal(err)
		return "", err
	}
	if err := wc.Close(); err != nil {
		log.Fatal(err)
		return "", err
	}
	log.Println("Uploaded= " + name)
	return name, nil
}

func (str *Storage) Read(bucket, object string) ([]byte, error) {
	ctx := context.Background()
	// [START download_file]
	rc, err := str.client.Bucket(bucket).Object(object).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}
	return data, nil
	// [END download_file]
}
