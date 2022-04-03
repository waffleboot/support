package main

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/google/uuid"
	"github.com/urfave/cli/v2"
	"github.com/waffleboot/cloud/internal/app"

	app_context "github.com/waffleboot/cloud/internal/context"
)

var (
	FlagContext = &cli.StringFlag{
		Name:    "context",
		Aliases: []string{"c"},
	}
	FlagConfig = &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"f"},
		Value:   "cloud.json",
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

var _api app.ContextAPI

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	app := &cli.App{
		Name: "cloud",
		Before: func(ctx *cli.Context) error {
			api, err := app_context.NewContextAPI(app_context.ContextApiParams{
				UseConfig:  ctx.String(FlagConfig.Name),
				UseContext: ctx.String(FlagContext.Name),
			})
			if err != nil {
				return err
			}
			_api = api
			return nil
		},
		After: func(ctx *cli.Context) error {
			if _api == nil {
				return nil
			}
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
				Action:          db,
				Subcommands: []*cli.Command{
					{
						Name:   "list",
						Action: dbList,
						Before: func(ctx *cli.Context) error {
							return dbNeedContext(ctx)
						},
					},
					{
						Name: "create",
						Flags: []cli.Flag{
							FlagHost,
						},
						Action: dbCreate,
						Before: func(ctx *cli.Context) error {
							err := dbNeedContext(ctx)
							if err != nil {
								return err
							}
							return dbNeedHost(ctx)
						},
					},
					{
						Name: "delete",
						Flags: []cli.Flag{
							FlagHost,
						},
						Action: dbDelete,
						Before: func(ctx *cli.Context) error {
							err := dbNeedContext(ctx)
							if err != nil {
								return err
							}
							return dbNeedHost(ctx)
						},
						// Before: func(ctx *cli.Context) error {
						// 	if err := beforeDBCommands(ctx); err != nil {
						// 		return err
						// 	}
						// 	appCtx := _api.ContextAPI().CurrentContext()
						// 	if len(appCtx.Services) == 0 {
						// 		return errors.New("no services to delete")
						// 	}
						// 	return nil
						// },
					},
				},
			},
		},
	}

	return app.RunContext(ctx, os.Args)
}

// func initAPI(ctx *cli.Context) error {
// 	config := strings.TrimSpace(ctx.String(FlagConfig.Name))
// 	if err := _api.Init(config); err != nil {
// 		return err
// 	}
// 	useCtx := strings.TrimSpace(ctx.String(FlagContext.Name))
// 	if useCtx != "" {
// 		_api.ContextAPI().UseContext(useCtx)
// 		_api.ContextAPI().MarkDirty()
// 	}
// 	return nil
// }

// func beforeDBCommands(ctx *cli.Context) error {
// 	appCtx := _api.ContextAPI().CurrentContext()

// 	host := strings.TrimSpace(ctx.String(FlagHost.Name))
// 	if host == "" {
// 		if appCtx == nil {
// 			return errors.New("need host and context")
// 		}
// 		urlHost := appCtx.Host
// 		if urlHost == nil {
// 			return fmt.Errorf("need host for %s context", appCtx.Name)
// 		}
// 		return nil
// 	}

// 	if appCtx == nil {
// 		return errors.New("need context")
// 	}

// 	urlHost, err := url.Parse(host)
// 	if err != nil {
// 		return err
// 	}
// 	appCtx.Host = urlHost

// 	return nil
// }

func dbNeedContext(ctx *cli.Context) error {
	if _api.CurrentContext() == nil {
		return errors.New("need context")
	}
	return nil
}

func dbNeedHost(ctx *cli.Context) error {
	host := ctx.String(FlagHost.Name)
	if host != "" {
		u, err := url.Parse(host)
		if err != nil {
			return err
		}
		_api.CurrentContext().Host = u
		_api.MarkDirty()
	}
	if _api.CurrentContext().Host == nil {
		return errors.New("need host")
	}
	return nil
}

func db(ctx *cli.Context) error {
	for _, ctx := range _api.Contexts() {
		if ctx.Host == nil {
			fmt.Println(ctx.Name)
			continue
		}
		fmt.Println(ctx.Name, ctx.Host)
	}
	return nil
}

func dbList(ctx *cli.Context) error {
	currentContext := _api.CurrentContext()
	for i := range currentContext.Services {
		fmt.Println(currentContext.Services[i])
	}
	return nil
}

func dbCreate(ctx *cli.Context) error {
	return _api.AddService(uuid.New())
}

func dbDelete(ctx *cli.Context) error {
	if argID := ctx.Args().First(); argID != "" {
		id, err := uuid.Parse(argID)
		if err != nil {
			return err
		}
		return _api.DelService(id)
	}
	services := _api.CurrentContext().Services
	for i := range services {
		if err := _api.DelService(services[i]); err != nil {
			return err
		}
	}
	return nil
}
