# Cloud functions for fetching CLEOS clearing reports
This directory contains a Go package suitable for deployment to Google Cloud Functions

## DailyClearing
DailyClearing is triggered by Cloud PubSub. It tries to fetch all available reports using the `previousReportId`
parameter in the PubSub message's payload, and uploads them to a cloud storage bucket. If successful it updates
the scheduled job that triggered it with the most recent report ID.

### Configuration

#### Environment
DailyClearing expects to find these environment variables:
- `APP_ENV`: The CLEOS environment to communicate with. Possible values are `prod`, `staging` and `dev`.
- `BUCKET_ID`: The bucket to place reports in.
- `CLEOS_TEMPLATE_ID`: The CLEOS template ID to fetch.
- `CLIENT_ID`: The client id used for authenticating with CLEOS.
- `CLIENT_SECRET`: The client secret used for authenticating with CLEOS.
- `SCHEDULED_JOB_ID`: The id of the scheduled job to update. Example value: `projects/{PROJECT_ID}/locations/{LOCATION}/jobs/{JOB_NAME}`

#### Deployment
`gcloud functions deploy {FUNCTION_NAME} --region=europe-west1 --runtime=go113 --entry-point=DailyClearing --trigger-topic={TRIGGER_TOPIC}`
