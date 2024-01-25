package main

import (
	"context"
	"fmt"
	"os"

	"github.com/base-org/blob-archiver/api/flags"
	"github.com/base-org/blob-archiver/api/metrics"
	"github.com/base-org/blob-archiver/api/service"
	"github.com/base-org/blob-archiver/common/beacon"
	"github.com/base-org/blob-archiver/common/storage"
	opservice "github.com/ethereum-optimism/optimism/op-service"
	"github.com/ethereum-optimism/optimism/op-service/cliapp"
	oplog "github.com/ethereum-optimism/optimism/op-service/log"
	opmetrics "github.com/ethereum-optimism/optimism/op-service/metrics"
	"github.com/ethereum/go-ethereum/log"
	"github.com/urfave/cli/v2"
)

var (
	Version   = "v0.0.1"
	GitCommit = ""
	GitDate   = ""
)

func main() {
	oplog.SetupDefaults()

	app := cli.NewApp()
	app.Flags = cliapp.ProtectFlags(flags.Flags)
	app.Version = opservice.FormatVersion(Version, GitCommit, GitDate, "")
	app.Name = "blob-api"
	app.Usage = "API service for Ethereum blobs"
	app.Description = "Service for fetching blob sidecars from a datastore"
	app.Action = cliapp.LifecycleCmd(Main())

	err := app.Run(os.Args)
	if err != nil {
		log.Crit("Application failed", "message", err)
	}
}

// Main is the entrypoint into the API.
// This method returns a cliapp.LifecycleAction, to create an op-service CLI-lifecycle-managed API Server.
func Main() cliapp.LifecycleAction {
	return func(cliCtx *cli.Context, closeApp context.CancelCauseFunc) (cliapp.Lifecycle, error) {
		cfg := flags.ReadConfig(cliCtx)
		if err := cfg.Check(); err != nil {
			return nil, fmt.Errorf("config check failed: %w", err)
		}

		l := oplog.NewLogger(oplog.AppOut(cliCtx), cfg.LogConfig)
		oplog.SetGlobalLogHandler(l.GetHandler())
		opservice.ValidateEnvVars(flags.EnvVarPrefix, flags.Flags, l)

		registry := opmetrics.NewRegistry()
		m := metrics.NewMetrics(registry)

		storageClient, err := storage.NewStorage(cfg.StorageConfig, l)
		if err != nil {
			return nil, err
		}

		beaconClient, err := beacon.NewBeaconClient(context.Background(), cfg.BeaconConfig)
		if err != nil {
			return nil, err
		}

		l.Info("Initializing API Service")
		return service.NewAPIService(l, storageClient, beaconClient, cfg, registry, m), nil
	}
}