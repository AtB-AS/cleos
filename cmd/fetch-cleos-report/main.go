package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"

	"golang.org/x/oauth2/clientcredentials"

	"github.com/atb-as/cleos/pkg/cleos"
)

//goland:noinspection GoUnhandledErrorResult
func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "usage: %s [flags] template_id id_after first_ordered_date\navailable flags:\n", os.Args[0])
		flag.PrintDefaults()
	}
	env := flag.String("e", "staging", "Environment [dev|staging|prod]")
	ts := flag.Int("t", 10, "Timeout in seconds")
	flag.Parse()

	if len(flag.Args()) < 3 {
		flag.Usage()
		os.Exit(1)
	}

	templateId, err := strconv.Atoi(flag.Arg(0))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse template_id: %v\n", err)
		os.Exit(1)
	}

	idAfter, err := strconv.Atoi(flag.Arg(1))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse id_after: %v\n", err)
		os.Exit(1)
	}

	firstOrderedDate, err := time.Parse("2006-01-02", flag.Arg(2))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse first_ordered_date: %v\n", err)
		os.Exit(1)
	}

	var tokenURL, audience, basePath string
	switch *env {
	case "prod":
		tokenURL = cleos.TokenURLProd
		audience = cleos.AudienceProd
		basePath = cleos.BasePathProd
	case "dev":
		tokenURL = cleos.TokenURLDev
		audience = cleos.AudienceDev
		basePath = cleos.BasePathDev
	case "staging":
		fallthrough
	default:
		tokenURL = cleos.TokenURLStaging
		audience = cleos.AudienceStaging
		basePath = cleos.BasePathStaging
	}

	credentials := clientcredentials.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		TokenURL:     tokenURL,
		EndpointParams: url.Values{
			"audience": {audience},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*ts)*time.Second)
	defer cancel()

	svc := cleos.NewService(credentials.Client(ctx), basePath)
	report, err := svc.NextReport(ctx, strconv.Itoa(templateId), strconv.Itoa(idAfter), firstOrderedDate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	file, err := os.Create(path.Join(cwd, report.Filename))
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	_, err = file.Write(report.Content)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stdout, "Successfully wrote file: %s with report ID %s\n", file.Name(), report.ID)
}
