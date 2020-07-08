package functions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/atb-as/cleos/pkg/cleos"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/api/cloudscheduler/v1"
)

var (
	projectID  = os.Getenv("GCP_PROJECT")
	bucketID   = os.Getenv("BUCKET_ID")
	jobID      = os.Getenv("SCHEDULED_JOB_ID")
	templateID = os.Getenv("CLEOS_TEMPLATE_ID")

	clientID     = os.Getenv("CLIENT_ID")
	clientSecret = os.Getenv("CLIENT_SECRET")

	appEnv = os.Getenv("APP_ENV")

	defaultDate = time.Date(2020, 01, 01, 0, 0, 0, 0, time.UTC)

	cleosService     *cleos.Service
	schedulerService *cloudscheduler.Service
	storageClient    *storage.Client
)

// PubSubMessage is the payload of a Pub/Sub event
type PubSubMessage struct {
	Data []byte `json:"data"`
}

// JobDescription represents properties of the scheduled job
type JobDescription struct {
	PreviousReportID string `json:"previousReportId"`
}

// Initialize service variables in init because they may survive between
// function invocations and add to overall function latency
func init() {
	var err error
	schedulerService, err = cloudscheduler.NewService(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

// DailyClearing is triggered by pubsub with a payload of JobDescription. It
// fetches the most recent CLEOS clearing report and uploads it to a cloud
// storage bucket. After successful upload it updates the scheduled job that
// triggers it to include the most recent report ID in its payload
func DailyClearing(ctx context.Context, m PubSubMessage) error {
	var tokenURL, audience, apiBasePath string
	switch appEnv {
	case "prod":
		tokenURL = cleos.TokenURLProd
		audience = cleos.AudienceProd
		apiBasePath = cleos.BasePathProd
	case "staging":
		tokenURL = cleos.TokenURLStaging
		audience = cleos.AudienceStaging
		apiBasePath = cleos.BasePathStaging
	default:
		tokenURL = cleos.TokenURLDev
		audience = cleos.TokenURLDev
		apiBasePath = cleos.BasePathDev
	}

	creds := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     tokenURL,
		EndpointParams: url.Values{
			"audience": {audience},
		},
	}
	cleosService = cleos.NewService(creds.Client(context.Background()), apiBasePath)

	var job JobDescription
	err := json.Unmarshal(m.Data, &job)
	if err != nil {
		return err
	}

	report, err := cleosService.ClearingReport(ctx, templateID, job.PreviousReportID, defaultDate)
	if err != nil {
		return err
	}

	if err := uploadToCloudStorage(ctx, report); err != nil {
		return err
	}

	if err := updateScheduledPayload(report.ReportID); err != nil {
		return err
	}

	return nil
}

// uploadToCloudStorage writes the current report into a cloud storage bucket
func uploadToCloudStorage(ctx context.Context, report *cleos.ClearingReportResponse) error {
	bucket := storageClient.Bucket(bucketID)

	o := bucket.Object(report.Filename)
	w := o.NewWriter(ctx)
	w.ContentType = report.ContentType
	_, err := w.Write(report.Content)
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	return nil
}

// updateScheduledPayload updates the payload of the next scheduled pubsub
// message to the trigger channel to match the most recent report ID
func updateScheduledPayload(reportID string) error {
	var jobDesc = JobDescription{
		PreviousReportID: reportID,
	}

	job, err := schedulerService.Projects.Locations.Jobs.Get(jobID).Do()
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(jobDesc)
	job.PubsubTarget.Data = base64.StdEncoding.EncodeToString(payload)

	if _, err := schedulerService.Projects.Locations.Jobs.Patch(jobID, job).Do(); err != nil {
		return err
	}

	return nil
}
