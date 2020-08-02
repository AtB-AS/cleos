package functions

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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
	projectID    = os.Getenv("GCP_PROJECT")
	bucketHandle = os.Getenv("BUCKET_ID")
	jobID        = os.Getenv("SCHEDULED_JOB_ID")
	templateID   = os.Getenv("CLEOS_TEMPLATE_ID")

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

type config struct {
	tokenURL    string
	audience    string
	apiBasePath string
}

type jobDescription struct {
	PreviousReportID string `json:"previousReportId"`
}

// Initialize service variables in init because they may survive between
// function invocations and add to overall function latency
func init() {
	var err error

	cfg := configFromEnvironment(appEnv)
	creds := clientcredentials.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		TokenURL:     cfg.tokenURL,
		EndpointParams: url.Values{
			"audience": {cfg.audience},
		},
	}
	cleosService = cleos.NewService(creds.Client(context.Background()), cfg.apiBasePath)

	schedulerService, err = cloudscheduler.NewService(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	storageClient, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatal(err)
	}

}
func configFromEnvironment(env string) config {
	var cfg config
	switch env {
	case "prod":
		cfg.tokenURL = cleos.TokenURLProd
		cfg.audience = cleos.AudienceProd
		cfg.apiBasePath = cleos.BasePathProd
	case "staging":
		cfg.tokenURL = cleos.TokenURLStaging
		cfg.audience = cleos.AudienceStaging
		cfg.apiBasePath = cleos.BasePathStaging
	default:
		cfg.tokenURL = cleos.TokenURLDev
		cfg.audience = cleos.AudienceDev
		cfg.apiBasePath = cleos.BasePathDev
	}

	return cfg
}

// DailyClearing is triggered by pubsub with a payload of JobDescription. It
// fetches the most recent CLEOS clearing report and uploads it to a cloud
// storage bucket. After successful upload it updates the scheduled job that
// triggers it to include the most recent report ID in its payload
func DailyClearing(ctx context.Context, m PubSubMessage) error {
	var job jobDescription
	var currentID string
	err := json.Unmarshal(m.Data, &job)
	if err != nil {
		return err
	}

	currentID = job.PreviousReportID
	for {
		start := time.Now()
		report, err := cleosService.NextReport(ctx, templateID, currentID, defaultDate)
		if err != nil {
			if err == cleos.ErrAllDownloaded || err == cleos.ErrNotGenerated {
				break
			}
			return err
		}
		w := newBucketWriter(
			ctx,
			bucketHandle,
			fmt.Sprintf("%s_%s", report.ReportID, report.Filename),
			report.ContentType)
		if _, err := w.Write(report.Content); err != nil {
			return err
		}
		if err := w.Close(); err != nil {
			return err
		}

		log.Printf("successfully fetched report %s (%s) in %s", report.ReportID, report.Filename, time.Since(start))
		currentID = report.ReportID
	}

	if err = updateScheduledPayload(currentID); err != nil {
		return err
	}
	return nil
}

func newBucketWriter(ctx context.Context, bucketHandle, name, contentType string) io.WriteCloser {
	bucket := storageClient.Bucket(bucketHandle)
	w := bucket.Object(name).NewWriter(ctx)
	w.ContentType = contentType
	return w
}

// updateScheduledPayload updates the payload of the next scheduled pubsub
// message to the trigger channel to match the most recent report ID
func updateScheduledPayload(reportID string) error {
	var jobDesc = jobDescription{
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
