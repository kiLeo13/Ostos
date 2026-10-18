package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/labstack/echo/v5"
	"ostos/backend/internal/accounts"
	"ostos/backend/internal/config"
	"ostos/backend/internal/families"
	"ostos/backend/internal/platform"
	platformses "ostos/backend/internal/platform/ses"
	mysqlrepository "ostos/backend/internal/repository/mysql"
	"ostos/backend/internal/shopping"
	"ostos/backend/internal/transport/http/router"
	"ostos/backend/internal/transport/http/validation"
)

func main() {
	if len(os.Args) == 2 && os.Args[1] == "healthcheck" {
		if err := runHealthcheck(os.Getenv("OSTOS_TLS_CERT_FILE"), os.Getenv("OSTOS_API_HOSTNAME"), "https://127.0.0.1:8443/health"); err != nil {
			fmt.Fprintln(os.Stderr, "healthcheck failed")
			os.Exit(1)
		}
		return
	}
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	configuration, err := config.Load(os.LookupEnv)
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}
	if configuration.Environment == "production" {
		if err := platform.ValidateServerCertificate(configuration.TLSCertFile, configuration.TLSKeyFile, configuration.APIHostname, time.Now()); err != nil {
			logger.Error("invalid TLS configuration", "error", err)
			os.Exit(1)
		}
	}

	idGenerator, err := platform.NewSnowflakeGenerator(configuration.SnowflakeNode)
	if err != nil {
		logger.Error("initialize Snowflake generator", "error", err)
		os.Exit(1)
	}

	database, sqlDatabase, err := mysqlrepository.Open(configuration.DatabaseDSN)
	if err != nil {
		logger.Error("initialize database", "error", err)
		os.Exit(1)
	}
	defer func() {
		if closeErr := sqlDatabase.Close(); closeErr != nil {
			logger.Error("close database", "error", closeErr)
		}
	}()

	accountRepository := mysqlrepository.NewAccountRepository(database)
	familyRepository := mysqlrepository.NewFamilyRepository(database)
	jwtManager, err := accounts.NewJWTManager(configuration.JWTSecret, configuration.JWTIssuer, configuration.JWTAudience, 30*time.Minute)
	if err != nil {
		logger.Error("initialize JWT manager", "error", err)
		os.Exit(1)
	}
	tokenCipher, err := accounts.NewTokenCipher(configuration.EmailTokenKey)
	if err != nil {
		logger.Error("initialize email token cipher", "error", err)
		os.Exit(1)
	}
	clock := platform.SystemClock{}
	accountService, err := accounts.NewService(accountRepository, idGenerator, clock, accounts.Argon2Hasher{}, jwtManager, tokenCipher)
	if err != nil {
		logger.Error("initialize account service", "error", err)
		os.Exit(1)
	}
	familyService, err := families.NewService(familyRepository, idGenerator, clock)
	if err != nil {
		logger.Error("initialize family service", "error", err)
		os.Exit(1)
	}
	shoppingRepository := mysqlrepository.NewShoppingRepository(database)
	shoppingService, err := shopping.NewService(shoppingRepository, idGenerator, clock)
	if err != nil {
		logger.Error("initialize shopping service", "error", err)
		os.Exit(1)
	}
	requestValidator, err := validation.New()
	if err != nil {
		logger.Error("initialize request validator", "error", err)
		os.Exit(1)
	}

	server := router.New(router.Dependencies{
		Logger: logger, Database: sqlDatabase, BodyLimit: configuration.BodyLimit, CORSOrigins: configuration.CORSOrigins,
		TrustedProxies: configuration.TrustedProxies, Accounts: accountService, AccessTokens: accountService,
		Families: familyService, Shopping: shoppingService, Validator: requestValidator,
	})

	shutdownContext, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if configuration.SESEnabled {
		awsConfiguration, loadErr := awsconfig.LoadDefaultConfig(shutdownContext, awsconfig.WithRegion(configuration.SESRegion))
		if loadErr != nil {
			logger.Error("initialize AWS configuration", "error", loadErr)
			os.Exit(1)
		}
		emailSender, senderErr := platformses.New(sesv2.NewFromConfig(awsConfiguration), configuration.SESFrom)
		if senderErr != nil {
			logger.Error("initialize SES sender", "error", senderErr)
			os.Exit(1)
		}
		worker, workerErr := accounts.NewOutboxWorker(accountRepository, tokenCipher, emailSender, clock, configuration.PublicAppURL)
		if workerErr != nil {
			logger.Error("initialize email outbox", "error", workerErr)
			os.Exit(1)
		}
		go runEmailOutbox(shutdownContext, logger, worker)
	}
	start := echo.StartConfig{
		Address: configuration.HTTPAddress, HideBanner: true, GracefulTimeout: 10 * time.Second,
		BeforeServeFunc: func(httpServer *http.Server) error {
			httpServer.ReadHeaderTimeout = configuration.ReadTimeout
			httpServer.ReadTimeout = configuration.ReadTimeout
			httpServer.WriteTimeout = configuration.WriteTimeout
			httpServer.IdleTimeout = 60 * time.Second
			return nil
		},
	}
	logger.Info("API listening", "address", configuration.HTTPAddress, "environment", configuration.Environment, "tls", configuration.Environment == "production")
	if configuration.Environment == "production" {
		start.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
		err = start.StartTLS(shutdownContext, server, configuration.TLSCertFile, configuration.TLSKeyFile)
	} else {
		err = start.Start(shutdownContext, server)
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
		logger.Error("API server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}

func runHealthcheck(certFile, hostname, endpoint string) error {
	certificate, err := os.ReadFile(certFile)
	if err != nil {
		return err
	}
	roots := x509.NewCertPool()
	if !roots.AppendCertsFromPEM(certificate) {
		return errors.New("invalid healthcheck certificate")
	}
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12, RootCAs: roots, ServerName: hostname,
		}},
	}
	response, err := client.Get(endpoint)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 4096))
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("health endpoint returned %d", response.StatusCode)
	}
	return nil
}

func runEmailOutbox(ctx context.Context, logger *slog.Logger, worker *accounts.OutboxWorker) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for {
				err := worker.DispatchOne(ctx)
				if errors.Is(err, accounts.ErrNoPendingEmail) {
					break
				}
				if err != nil {
					logger.ErrorContext(ctx, "email outbox delivery failed", "error", err)
					break
				}
			}
		}
	}
}
