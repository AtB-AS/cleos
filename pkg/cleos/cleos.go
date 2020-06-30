package cleos

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"regexp"
	"time"
)

const (
	BasePathDev     = "https://api.dev.entur.io/cleos-reporting/api/v1"
	BasePathStaging = "https://api.staging.entur.io/cleos-reporting/api/v1"
	BasePathProd    = "https://api.entur.io/cleos-reporting/api/v1"

	TokenURLDev     = "https://partner.dev.entur.org/oauth/token"
	TokenURLStaging = "https://partner.staging.entur.org/oauth/token"
	TokenURLProd    = "https://partner.entur.org/oauth/token"

	AudienceDev     = "https://cleos.dev.entur.io"
	AudienceStaging = "https://cleos.staging.entur.io"
	AudienceProd    = "https://cleos.entur.io"

	ErrUnknownReportType ErrCleos = "supplied report type is unknown"
	ErrAllDownloaded     ErrCleos = "all available reports downloaded, retry later"
	ErrNotGenerated      ErrCleos = "report is being generated, retry later"
	ErrUnauthorized      ErrCleos = "unauthorized"
	ErrForbidden         ErrCleos = "forbidden"
	ErrConflict          ErrCleos = "report failed execution, contact support"
	ErrGone              ErrCleos = "no future reports on this templateID will be generated, stop the job or update the templateID"
	ErrUnknownStatus     ErrCleos = "unknown status"

	dateLayout = "2006-01-02"
)

type ErrCleos string

func (e ErrCleos) Error() string { return string(e) }

type ClearingReportResponse struct {
	Content     []byte
	ContentType string
	Filename    string
	ReportID    string
}

type Service struct {
	basePath string
	client   *http.Client
}

func NewService(client *http.Client, basePath string) *Service {
	return &Service{
		client:   client,
		basePath: basePath,
	}
}

func (s *Service) ClearingReport(ctx context.Context, templateID, IDAfter string, firstOrderedDate time.Time) (*ClearingReportResponse, error) {
	d := firstOrderedDate.Format(dateLayout)
	q := url.Values{
		"templateId":       []string{templateID},
		"idAfter":          []string{IDAfter},
		"firstOrderedDate": []string{d},
	}
	endpoint := fmt.Sprintf("%s/partner-reports/report/next/content?%s", s.basePath, q.Encode())

	req, err := s.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	res := ClearingReportResponse{}
	err = s.do(req, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (s *Service) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (s *Service) do(req *http.Request, v interface{}) error {
	res, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		switch res.StatusCode {
		case http.StatusAccepted:
			return ErrAllDownloaded
		case http.StatusNoContent:
			return ErrNotGenerated
		case http.StatusUnauthorized:
			return ErrUnauthorized
		case http.StatusForbidden:
			return ErrForbidden
		case http.StatusConflict:
			return ErrConflict
		case http.StatusGone:
			return ErrGone
		default:
			return ErrUnknownStatus
		}
	}

	if v == nil {
		return nil
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	contentDisposition := res.Header.Get("Content-Disposition")
	// Work around an issue where the returned Content-Disposition header values
	// are unquoted
	re := regexp.MustCompile("=(.+)$")
	quoted := re.ReplaceAllString(contentDisposition, "=\"$1\"")
	_, params, err := mime.ParseMediaType(quoted)
	if err != nil {
		return err
	}
	reportId := res.Header.Get("X-Entur-Report-Id")
	contentType := res.Header.Get("Content-Type")

	switch t := v.(type) {
	case *ClearingReportResponse:
		t.Content = body
		t.Filename = params["filename"]
		t.ReportID = reportId
		t.ContentType = contentType
	default:
		return ErrUnknownReportType
	}

	return nil
}
