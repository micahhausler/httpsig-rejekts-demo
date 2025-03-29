package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/common-fate/httpsig"
	"github.com/common-fate/httpsig/inmemory"
	"github.com/common-fate/httpsig/sigparams"
	"github.com/micahhausler/rejekts-eu-2025/cmd"
	"github.com/micahhausler/rejekts-eu-2025/pkg/attributes"
	"github.com/micahhausler/rejekts-eu-2025/pkg/gh"
	flag "github.com/spf13/pflag"
)

func main() {
	port := flag.Int("port", 8080, "port to listen on")
	authority := flag.String("authority", "localhost:8080", "authority to listen on")
	scheme := flag.String("scheme", "https", "scheme to listen on")
	logLevel := cmd.LevelFlag(slog.LevelInfo)
	flag.Var(&logLevel, "log-level", "log level")
	flag.Parse()
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     slog.Level(logLevel),
		AddSource: slog.Level(logLevel) == slog.LevelDebug,
	})))

	addr := fmt.Sprintf("0.0.0.0:%d", *port)

	headerName := "X-GitHub-Username"
	keyDir, err := gh.NewDynamicGitHubKeyDirectory(headerName)
	if err != nil {
		slog.Error("failed to create key directory", "error", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()

	verifier := httpsig.Middleware(httpsig.MiddlewareOpts{
		NonceStorage: inmemory.NewNonceStorage(),
		KeyDirectory: keyDir,
		Tag:          "foo",
		Scheme:       *scheme,
		Authority:    *authority,
		OnValidationError: func(ctx context.Context, err error) {
			slog.Error("validation error", "error", err)
		},
		Validation: &sigparams.ValidateOpts{
			ForbidClientSideAlg: false,
			BeforeDuration:      time.Minute,
			RequiredCoveredComponents: map[string]bool{
				"@method":           true,
				"@target-uri":       true,
				"x-github-username": true,
				// "content-length": true,
				// "content-digest": true,
			},
			RequireNonce: true,
		},
		OnDeriveSigningString: func(ctx context.Context, stringToSign string) {
			slog.Debug("string to sign", "string", stringToSign)
		},
	})
	mux.Handle("/healthz", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	mux.Handle("/hello", keyDir.KeyFetcher(verifier(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawAttribute := httpsig.AttributesFromContext(r.Context())
		if rawAttribute == nil {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Signature verified, no attributes found")
			defer slog.Info("no attributes found")
			return
		}

		attr, ok := rawAttribute.(attributes.User)
		if !ok {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Signature verified, but attributes are not of type attributes.User")
			defer slog.Error("Attributes are not of type attributes.User")
			return
		}
		defer slog.Info("request", "username", attr.Username)
		fmt.Fprintf(w, `{"message": "hello, %s!"}`, attr.Username)
	}))))

	slog.Info("starting server", "address", addr)
	err = http.ListenAndServe(addr, mux)
	if err != nil {
		slog.Error("failed to start server", "error", err)
		os.Exit(1)
	}
}
