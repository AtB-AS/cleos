# Cloud functions for fetching CLEOS clearing reports
This directory contains a Go package suitable for deployment to Google Cloud 
Functions

## FetchCLEOSReport
FetchCLEOSReport is triggered by Cloud Scheduler via this Cloud PubSub message:

````json
{
  "data": {
   "previousReportId": "123"
  }
}
````
 
It tries to fetch all available CLEOS reports generated after `previousReportId`
, and uploads them to a cloud storage bucket. If successful it updates the 
scheduled job's payload with the most recent report ID.

In case of failure, the `previousReportId` parameter will stay unchanged, and the
function will pick up where it previously failed on the next invocation.

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
`gcloud functions deploy {FUNCTION_NAME} --region=europe-west1 --runtime=go113 --entry-point=FetchCLEOSReport --trigger-topic={TRIGGER_TOPIC}`

### Manual invocations
You can trigger a run by publishing a message to the trigger topic. To make the
function fetch all reports, beginning from the first published report:

````shell script
$ gcloud pubsub topics publish $TRIGGER_TOPIC --message='{"previousReportId": 0}'
````