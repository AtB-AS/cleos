package upload_report_sftp

import (
	"context"
	"log"
	"os"
	"path"

	secretmanager "cloud.google.com/go/secretmanager/apiv1beta1"
	"cloud.google.com/go/storage"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1beta1"
)

var storageClient *storage.Client
var sftpClient *sftp.Client

func init() {
	var err error
	ctx := context.Background()
	key, err := getPrivateKey(ctx, os.Getenv("KEY_NAME"))
	if err != nil {
		log.Fatal(err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal(err)
	}

	username := os.Getenv("SSH_USER")
	config := ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	addr := os.Getenv("SSH_HOST")
	client, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		log.Fatal(err)
	}

	sftpClient, err = sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}

	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

func getPrivateKey(ctx context.Context, keyName string) ([]byte, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	secret, err := client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{
		Name: keyName,
	})
	if err != nil {
		return nil, err
	}

	return []byte(secret.String()), nil
}

func UploadCLEOSReportToSFTP(ctx context.Context, e GCSEvent) error {
	reader, err := storageClient.Bucket(e.Bucket).Object(e.Name).NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := sftpClient.Open(path.Join(os.Getenv("SFTP_DIR"), e.Name))
	if err != nil {
		return err
	}

	_, err = writer.ReadFrom(reader)
	if err != nil {
		return err
	}

	return nil
}
