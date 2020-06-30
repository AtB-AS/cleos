package main

import (
	"cleos/pkg/cleos"
	"context"
	"golang.org/x/oauth2/clientcredentials"
	"log"
	"net/url"
	"os"
	"time"
)

func main() {
	creds := clientcredentials.Config{
		ClientID:     os.Getenv("CLIENT_ID"),
		ClientSecret: os.Getenv("CLIENT_SECRET"),
		TokenURL:     cleos.TokenURLStaging,
		EndpointParams: url.Values{
			"audience": {cleos.AudienceStaging},
		},
	}
	s := cleos.NewService(creds.Client(context.Background()), cleos.BasePathStaging)

	date := time.Date(2020, 06, 29, 0, 0, 0, 0, time.UTC)
	res, err := s.ClearingReport(context.Background(), "275", "11661", date)
	if err != nil {
		log.Fatal(err)
	}

	file, err := os.Create(res.Filename)
	if err != nil {
		log.Fatal(err)
	}
	n, err := file.Write(res.Content)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("wrote %d bytes to %s", n, file.Name())
}
