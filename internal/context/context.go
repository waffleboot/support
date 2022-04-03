package context

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/waffleboot/cloud/internal/app"
)

type ContextAPI struct {
	contextsHolder
	filename string
	dirty    bool
}

var validate *validator.Validate = validator.New()

type ContextApiParams struct {
	UseConfig     string `validate:"required"`
	UseContext    string
	UseContextNum string
}

func NewContextAPI(req ContextApiParams) (app.ContextAPI, error) {
	if err := validate.Struct(req); err != nil {
		return nil, err
	}

	f, err := os.Open(req.UseConfig)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
		var contexts contextsHolder
		return &ContextAPI{
			dirty:          contexts.setcurrent(req.UseContext),
			filename:       req.UseConfig,
			contextsHolder: contexts,
		}, nil
	}
	defer f.Close()

	config := &dtoConfig{}

	if err := json.NewDecoder(f).Decode(config); err != nil {
		return nil, err
	}

	if err := validate.Struct(config); err != nil {
		return nil, err
	}

	api, err := conv2domain(config)
	if err != nil {
		return nil, err
	}

	api.filename = req.UseConfig

	api.dirty = api.setcurrent(req.UseContext)

	return api, nil
}

func (a *ContextAPI) CurrentContext() *app.AppContext {
	return a.currentContext
}

func (a *ContextAPI) AddService(id uuid.UUID) error {
	if a.currentContext == nil {
		return errors.New("need context")
	}
	a.currentContext.Services = append(a.currentContext.Services, id)
	a.dirty = true
	return nil
}

func (a *ContextAPI) DelService(id uuid.UUID) error {
	if a.currentContext == nil {
		return errors.New("need context")
	}
	services := a.currentContext.Services
	for i := range services {
		if services[i].String() == id.String() {
			if len(services) == 1 {
				a.currentContext.Services = nil
				a.dirty = true
				return nil
			}
			last := len(services) - 1
			services[i] = services[last]
			a.currentContext.Services = services[:last]
			a.dirty = true
			return nil
		}
	}
	return errors.New("not found")
}

func (a *ContextAPI) MarkDirty() {
	a.dirty = true
}

func (a *ContextAPI) Contexts() []*app.AppContext {
	return a.contexts
}

func (a *ContextAPI) Close() error {
	if !a.dirty {
		return nil
	}
	a.dirty = false

	backup := a.filename + ".bkp"
	f, err := os.Create(backup)
	if err != nil {
		os.Remove(backup)
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(conv2dto(a)); err != nil {
		return err
	}

	return os.Rename(backup, a.filename)
}

type contextsHolder struct {
	currentContext *app.AppContext
	contexts       []*app.AppContext
}

func (h *contextsHolder) init(contexts []*app.AppContext, current string) {
	h.contexts = contexts
	h.setcurrent(current)
}

func (h *contextsHolder) setcurrent(name string) bool {
	if name == "" {
		return false
	}
	if h.currentContext != nil && h.currentContext.Name == name {
		return false
	}
	for i := range h.contexts {
		if h.contexts[i].Name == name {
			h.currentContext = h.contexts[i]
			return true
		}
	}
	h.currentContext = &app.AppContext{
		Name: name,
	}
	h.contexts = append(h.contexts, h.currentContext)
	return true
}
