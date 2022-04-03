package context

import (
	"net/url"

	"github.com/google/uuid"

	"github.com/waffleboot/cloud/internal/app"
)

type dtoConfig struct {
	CurrentContext string       `json:"current-context,omitempty"`
	Contexts       []dtoContext `json:"contexts,omitempty"`
}

type dtoContext struct {
	Name     string      `json:"name" validate:"required"`
	Host     string      `json:"host,omitempty"`
	Services []uuid.UUID `json:"services,omitempty"`
}

func conv2dto(a *ContextAPI) *dtoConfig {
	contexts := make([]dtoContext, len(a.contexts))
	for i := range a.contexts {
		ctx := a.contexts[i]
		host := ""
		if ctx.Host != nil {
			host = ctx.Host.String()
		}
		contexts[i] = dtoContext{
			Host:     host,
			Name:     ctx.Name,
			Services: ctx.Services,
		}
	}
	if a.currentContext == nil {
		return &dtoConfig{
			Contexts: contexts,
		}
	}
	return &dtoConfig{
		Contexts:       contexts,
		CurrentContext: a.currentContext.Name,
	}
}

func conv2domain(dtoConfig *dtoConfig) (*ContextAPI, error) {
	contexts := make([]*app.AppContext, len(dtoConfig.Contexts))
	for i := range dtoConfig.Contexts {
		ctx := dtoConfig.Contexts[i]
		host, err := url.Parse(ctx.Host)
		if err != nil {
			return nil, err
		}
		contexts[i] = &app.AppContext{
			Host:     host,
			Name:     ctx.Name,
			Services: ctx.Services,
		}
	}
	api := &ContextAPI{}
	api.init(contexts, dtoConfig.CurrentContext)
	return api, nil
}
