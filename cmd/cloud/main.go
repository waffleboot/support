package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"github.com/waffleboot/cloud/internal/app"
)

var (
	FlagContext = &cli.StringFlag{
		Name: "context",
	}
	FlagConfig = &cli.StringFlag{
		Name:  "config",
		Value: "cloud.json",
	}
	FlagHost = &cli.StringFlag{
		Name: "host",
	}
)

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
	}
}

var _api app.API

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	_api = app.NewAPI()

	app := &cli.App{
		Name:   "cloud",
		Before: initAPI,
		After: func(ctx *cli.Context) error {
			return _api.Close()
		},
		Flags: []cli.Flag{
			FlagContext,
			FlagConfig,
		},
		Commands: []*cli.Command{
			{
				Name:            "db",
				SkipFlagParsing: true,
				Before:          beforeDBCommands,
				Subcommands: []*cli.Command{
					{
						Name: "create",
						Flags: []cli.Flag{
							FlagHost,
						},
						Action: dbCreate,
						Before: beforeDBCommands,
					},
					{
						Name: "delete",
						Flags: []cli.Flag{
							FlagHost,
						},
						Action: dbDelete,
						Before: func(ctx *cli.Context) error {
							if err := beforeDBCommands(ctx); err != nil {
								return err
							}
							appCtx := _api.ContextAPI().CurrentContext()
							if len(appCtx.Services) == 0 {
								return errors.New("no services to delete")
							}
							return nil
						},
					},
				},
			},
		},
	}

	return app.RunContext(ctx, os.Args)
}

func initAPI(ctx *cli.Context) error {
	config := strings.TrimSpace(ctx.String(FlagConfig.Name))
	if err := _api.Init(config); err != nil {
		return err
	}
	useCtx := strings.TrimSpace(ctx.String(FlagContext.Name))
	if useCtx != "" {
		_api.ContextAPI().UseContext(useCtx)
		_api.ContextAPI().MarkDirty()
	}
	return nil
}

func beforeDBCommands(ctx *cli.Context) error {
	appCtx := _api.ContextAPI().CurrentContext()

	host := strings.TrimSpace(ctx.String(FlagHost.Name))
	if host == "" {
		if appCtx == nil {
			return errors.New("need host and context")
		}
		urlHost := appCtx.Host
		if urlHost == nil {
			return fmt.Errorf("need host for %s context", appCtx.Name)
		}
		return nil
	}

	if appCtx == nil {
		return errors.New("need context")
	}

	urlHost, err := url.Parse(host)
	if err != nil {
		return err
	}
	appCtx.Host = urlHost

	return nil
}

func dbCreate(ctx *cli.Context) error {
	id := uuid.New()
	arr := _api.ContextAPI().CurrentContext().Services
	arr = append(arr, id)
	_api.ContextAPI().CurrentContext().Services = arr
	_api.ContextAPI().MarkDirty()
	return nil
}

func dbDelete(ctx *cli.Context) error {
	// arr := _api.ContextAPI().CurrentContext().ServiceIDs
	return nil
}
