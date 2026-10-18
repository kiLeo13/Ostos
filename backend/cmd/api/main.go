package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
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
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	configuration, err := config.Load(os.LookupEnv)
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
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

	server := &http.Server{
		Addr: configuration.HTTPAddress,
		Handler: router.New(router.Dependencies{Logger: logger, Database: sqlDatabase, BodyLimit: configuration.BodyLimit,
			Accounts: accountService, AccessTokens: accountService, Families: familyService, Shopping: shoppingService, Validator: requestValidator}),
		ReadHeaderTimeout: configuration.ReadTimeout,
		ReadTimeout:       configuration.ReadTimeout,
		WriteTimeout:      configuration.WriteTimeout,
		IdleTimeout:       60 * time.Second,
	}

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
	go func() {
		logger.Info("API listening", "address", configuration.HTTPAddress, "environment", configuration.Environment)
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			logger.Error("API server stopped unexpectedly", "error", serveErr)
			stop()
		}
	}()

	<-shutdownContext.Done()
	gracefulContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(gracefulContext); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
		os.Exit(1)
	}
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
