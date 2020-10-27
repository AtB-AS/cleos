package ftp

import (
	"context"
	"fmt"
	"os"
	"path"
	"time"

	"cloud.google.com/go/functions/metadata"
	"cloud.google.com/go/storage"
	"github.com/jlaffaye/ftp"
	"golang.org/x/sync/errgroup"
)

var storageClient *storage.Client

const (
	timeout     = 5 * time.Second
	maxEventAge = 24 * time.Hour
)

// Initialize global variables that may survive function invocations.
func init() {
	var err error
	ctx := context.Background()

	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "storage.NewClient: %v", err)
		os.Exit(1)
	}
}

// UploadGCSObjectToFTP retrieves the contents of a Google Cloud Storage
// Object identified by e and uploads it to an FTP endpoint.
func UploadGCSObjectToFTP(ctx context.Context, e GCSEvent) error {
	// Give up delivery if the event is too old.
	// TODO: Notify?
	meta, err := metadata.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("metadata.FromContext: %v", err)
	}
	if time.Since(meta.Timestamp) > maxEventAge {
		fmt.Fprintf(
			os.Stderr,
			"event timeout: halting retries for event %q, regarding object creation of: %q",
			meta.EventID, fmt.Sprintf("gs://%s/%s", e.Bucket, e.Name),
		)
		return nil
	}

	ctx, done := context.WithTimeout(ctx, timeout)
	defer done()

	bucketReader, err := storageClient.Bucket(e.Bucket).Object(e.Name).NewReader(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := bucketReader.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "close: %v\n", err)
		}
	}()

	addr := os.Getenv("FTP_HOST")
	conn, err := ftp.Dial(addr, ftp.DialWithContext(ctx))
	if err != nil {
		return err
	}
	defer conn.Quit()

	var g errgroup.Group
	g.Go(func() error {
		env := os.Getenv("APP_ENV")
		pw := os.Getenv("FTP_PASSWORD")
		user := os.Getenv("FTP_USER")
		if err := conn.Login(user, pw); err != nil {
			return err
		}

		dirs, err := conn.List(".")
		if err != nil {
			return err
		}

		var found bool
		for _, d := range dirs {
			if d.Name == env {
				found = true
			}
		}

		if !found {
			if err := conn.MakeDir(env); err != nil {
				return err
			}
		}

		if err := conn.Stor(path.Join(env, e.Name), contextAwareReader{
			bucketReader, ctx,
		}); err != nil {
			return err
		}

		return nil
	})

	return g.Wait()
}
