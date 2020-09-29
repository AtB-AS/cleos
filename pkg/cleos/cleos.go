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
	cleosDateLayout = "2006-01-02"
)

type ErrCleos string

const (
	ErrNotFound         ErrCleos = "a report for the specified template id was not found"
	ErrAllDownloaded    ErrCleos = "all available reports downloaded, retry later"
	ErrNotGenerated     ErrCleos = "report is being generated, retry later"
	ErrUnauthorized     ErrCleos = "unauthorized"
	ErrForbidden        ErrCleos = "forbidden"
	ErrConflict         ErrCleos = "report failed execution, contact support"
	ErrGone             ErrCleos = "no future reports on this templateID will be generated, stop the job or update the templateID"
	ErrUnknownStatus    ErrCleos = "unknown status"
	ErrInvalidArguments ErrCleos = "invalid arguments"
)

func (e ErrCleos) Error() string { return string(e) }

type Report struct {
	ID          string
	Content     []byte
	ContentType string
	Filename    string
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

// NextReport fetches the next available report for templateID generated after
// IDAfter on or after firstOrderedDate
func (s *Service) NextReport(ctx context.Context, templateID, IDAfter string, firstOrderedDate time.Time) (*Report, error) {
	d := firstOrderedDate.Format(cleosDateLayout)
	q := url.Values{
		"templateId":       {templateID},
		"idAfter":          {IDAfter},
		"firstOrderedDate": {d},
	}
	endpoint := fmt.Sprintf("%s/partner-reports/report/next/content?%s", s.basePath, q.Encode())

	req, err := s.newRequest(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := s.do(req)
	if err != nil {
		return nil, err
	}

	filename, err := res.filename()
	if err != nil {
		return nil, err
	}
	return &Report{
		Content:     res.content,
		ContentType: res.contentType(),
		Filename:    filename,
		ID:          res.reportID(),
	}, nil
}

func (s *Service) newRequest(ctx context.Context, method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *cleosResponse) contentType() string {
	return c.headers.Get("Content-Type")
}

func (c *cleosResponse) reportID() string {
	return c.headers.Get("X-Entur-Report-Id")
}

// Extracts the filename key from the Content-Disposition header. If the value
// contains spaces, Cleos will not quote it which violates RFC2183 and makes
// mime.ParseMediaType choke. Work around it by always quoting the value.
func (c *cleosResponse) filename() (string, error) {
	contentDisposition := c.headers.Get("Content-Disposition")
	re := regexp.MustCompile("filename=([^;]+)")
	quoted := re.ReplaceAllString(contentDisposition, "filename=\"$1\"")

	_, params, err := mime.ParseMediaType(quoted)
	if err != nil {
		return "", err
	}

	return params["filename"], nil
}

type cleosResponse struct {
	content []byte
	headers http.Header
}

func (s *Service) do(req *http.Request) (*cleosResponse, error) {
	res, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		switch res.StatusCode {
		case http.StatusNotFound:
			return nil, ErrNotFound
		case http.StatusAccepted:
			return nil, ErrAllDownloaded
		case http.StatusNoContent:
			return nil, ErrNotGenerated
		case http.StatusUnauthorized:
			return nil, ErrUnauthorized
		case http.StatusForbidden:
			return nil, ErrForbidden
		case http.StatusConflict:
			return nil, ErrConflict
		case http.StatusGone:
			return nil, ErrGone
		case http.StatusBadRequest:
			return nil, ErrInvalidArguments
		default:
			return nil, ErrUnknownStatus
		}
	}

	content, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return &cleosResponse{
		content: content,
		headers: res.Header,
	}, nil
}
