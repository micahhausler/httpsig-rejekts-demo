package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/aoliveti/curling"
	"github.com/common-fate/httpsig/signer"
	"github.com/micahhausler/rejekts-eu-2025/cmd"
	"github.com/micahhausler/rejekts-eu-2025/gh"
	"github.com/micahhausler/rejekts-eu-2025/transport"
)

func main() {
	keyFile := flag.String("key", "", "path to private key")
	host := flag.String("host", "localhost", "host to connect to")
	port := flag.Int("port", 8080, "port to connect to")
	username := flag.String("username", "", "GitHub username to authenticate as")
	execute := flag.Bool("execute", false, "execute the request")
	logLevel := cmd.LevelFlag(slog.LevelInfo)
	flag.Var(&logLevel, "log-level", "log level")
	flag.Parse()

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.Level(logLevel),
		AddSource: slog.Level(logLevel) == slog.LevelDebug,
	})))

	addr := fmt.Sprintf("http://%s:%d", *host, *port)

	keyData, err := os.ReadFile(*keyFile)
	if err != nil {
		slog.Error("failed to read key file", "error", err)
		os.Exit(1)
	}

	algorithm, err := gh.NewGHSigner(keyData)
	if err != nil {
		slog.Error("failed to create signer", "error", err)
		os.Exit(1)
	}

	signer := gh.NewRequestSigner(signer.Transport{
		KeyID: algorithm.KeyID(),
		Tag:   "foo",
		Alg:   algorithm,
		CoveredComponents: []string{
			"@method",
			"@target-uri",
			"content-type",
			// "content-length",
			// "content-digest",
			"x-github-username",
		},
		BaseTransport: transport.NewTransportWithFallbackHeaders(http.DefaultTransport, http.Header{
			"Content-Type": []string{"application/json"},
		}),
		OnDeriveSigningString: func(ctx context.Context, stringToSign string) {
			slog.Debug("signing string", "string", stringToSign)
		},
	})

	req, err := http.NewRequest(http.MethodPost, addr, nil)
	req.Header.Add("X-GitHub-Username", *username)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		slog.Error("Failed to create request", "error", err.Error())
		os.Exit(1)
	}
	req2, err := signer.SignRequest(req)
	if err != nil {
		slog.Error("Failed to sign request", "error", err.Error())
		os.Exit(1)
	}

	cmd, err := curling.NewFromRequest(req2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(cmd)
	fmt.Println()

	if *execute {
		client := http.DefaultClient
		resp, err := client.Do(req2)
		if err != nil {
			slog.Error("Failed to execute request", "error", err.Error())
			os.Exit(1)
		}

		resBytes, err := httputil.DumpResponse(resp, true)
		if err != nil {
			slog.Error("failed to dump response", "error", err)
			os.Exit(1)
		}

		fmt.Println(string(resBytes))
	}
}
