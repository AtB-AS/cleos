package upload_gcs_object_sftp

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
var signer ssh.Signer

// Initialize global variables that may survive function invocations.
func init() {
	var err error
	ctx := context.Background()

	privKey, err := accessSecret(ctx, os.Getenv("SECRET_NAME"))
	if err != nil {
		log.Fatal(err)
	}

	signer, err = ssh.ParsePrivateKey(privKey)
	if err != nil {
		log.Fatal(err)
	}

	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Fatal(err)
	}
}

// accessSecret accesses the payload of secret using Secret Manager.
func accessSecret(ctx context.Context, secret string) ([]byte, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	res, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: secret,
	})
	if err != nil {
		return nil, err
	}

	return res.Payload.Data, nil
}

// UploadGCSObjectToSFTP retrieves the contents of a Google Cloud Storage
// Object identified by e and uploads it to an SFTP endpoint.
func UploadGCSObjectToSFTP(ctx context.Context, e GCSEvent) error {
	username := os.Getenv("SSH_USER")
	config := ssh.ClientConfig{
		User:            username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	addr := os.Getenv("SSH_HOST")
	client, err := ssh.Dial("tcp", addr, &config)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpClient.Close()

	reader, err := storageClient.Bucket(e.Bucket).Object(e.Name).NewReader(ctx)
	if err != nil {
		return err
	}
	defer reader.Close()

	writer, err := sftpClient.Create(path.Join(os.Getenv("SFTP_DIR"), e.Name))
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.ReadFrom(reader)
	if err != nil {
		return err
	}

	return nil
}
