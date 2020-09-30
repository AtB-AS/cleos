# Cloud function for uploading a GCS Object to SFTP
This directory contains a Go package suitable for deployment to Google Cloud 
Functions

## UploadGCSObjectToSFTP
UploadGCSObjectToSFTP is designed to be triggered by Cloud Storage events of
type `google.storage.objects.finalize`
 
It retrieves the contents of the GCS Object and tries to store it on a remote
SFTP endpoint.

Retries should be enabled when deploying to work around transient failures.

### Configuration

#### Environment
UploadGCSObjectToSFTP expects to find these environment variables:
- `SECRET_NAME`: The full path to the Cloud Secret that holds the SSH private
key to use when authenticating with the remote SFTP endpoint. 

    Example value: `projects/my-project/secrets/my-secret/versions/latest`
    
    Creating a secret:
    ````shell script
    $ gcloud secrets create $SECRET_NAME --data-file=id_rsa
    ````
  
- `SSH_USERNAME`: The username of the SSH user to authenticate as.
- `SSH_HOST`: The address of the SFTP endpoint. Example value: `2.tcp.ngrok.io:18745`
- `SFTP_DIR`: Absolute path to the directory on the remote SFTP endpoint to put the GCS Object in.

#### Deployment
```shell script
gcloud functions deploy $FUNCTION_NAME --region=europe-west1 --runtime=go113 ---trigger-resource $TRIGGER_BUCKET --trigger-event google.storage.object.finalize
````
### Manual invocations
````shell script
$ gsutil cp somefile.txt gs://configured_bucket
````
