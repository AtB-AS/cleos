#!/bin/bash

case $1 in
  atb-staging)
    PROJECT="atb-mobility-platform-staging"
    BUCKET_ID="amp-staging-usage-report"
    APP_ENV="staging"
    FTP_HOST="ftp.atb.no:21"
    FTP_USER="ftp_entur"
    ;;

  atb-prod)
    PROJECT="atb-mobility-platform"
    BUCKET_ID="amp-prod-usage-report"
    APP_ENV="prod"
    FTP_HOST="ftp.atb.no:21"
    FTP_USER="ftp_entur"
    ;;

  *)
    echo "Usage: $0 atb-staging|atb-prod"
    exit 1
  ;;
esac

rm -rf vendor
go mod vendor
rm -rf vendor/cloud.google.com/go/functions/metadata/
gcloud beta functions deploy UploadCarnetReportFtp \
  --project=$PROJECT \
  --region=europe-west1 \
  --entry-point=UploadGCSObjectToFTP \
  --runtime=go113 \
  --trigger-bucket=$BUCKET_ID \
  --set-env-vars=APP_ENV=$APP_ENV,FTP_HOST=$FTP_HOST,FTP_USER=$FTP_USER \
  --set-secrets 'FTP_PASSWORD=ftp-report-password:latest'