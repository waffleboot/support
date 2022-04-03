package app

import (
	"encoding/json"
	"net/url"
	"os"

	"github.com/google/uuid"
)

type API interface {
	Init(string) error
	ContextAPI() ContextAPI
	Close() error
}

type ContextAPI interface {
	CurrentContext() *AppContext
	UseContext(string)
	MarkDirty()
	Save() error
}

type AppContext struct {
	Name     string
	Host     *url.URL
	Services []uuid.UUID
}

type config struct {
	CurrentContext string       `json:"current,omitempty"`
	Contexts       []appContext `json:"contexts,omitempty"`
}

type appContext struct {
	Name       string   `json:"name"`
	Host       string   `json:"host,omitempty"`
	ServiceIDs []string `json:"services,omitempty"`
}

type api struct {
	*contextAPI
}

func NewAPI() API {
	return &api{contextAPI: &contextAPI{}}
}

func (a *api) Init(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		if os.IsNotExist(err) {
			a.contextAPI.filename = filename
			a.contextAPI.dirty = true
			return nil
		}
		return err
	}
	defer f.Close()

	config := &config{}
	if err := json.NewDecoder(f).Decode(config); err != nil {
		return err
	}

	var currentContext *AppContext

	contexts := make([]*AppContext, 0, len(config.Contexts))

	for i := range config.Contexts {

		host, err := url.Parse(config.Contexts[i].Host)
		if err != nil {
			return err
		}

		services := make([]uuid.UUID, 0, len(config.Contexts[i].ServiceIDs))
		for j := range config.Contexts[i].ServiceIDs {
			id, err := uuid.Parse(config.Contexts[i].ServiceIDs[j])
			if err != nil {
				return err
			}
			services = append(services, id)
		}

		appCtx := &AppContext{
			Host:     host,
			Name:     config.Contexts[i].Name,
			Services: services,
		}

		if config.Contexts[i].Name == config.CurrentContext {
			currentContext = appCtx
		}

		contexts = append(contexts, appCtx)
	}

	a.contextAPI = &contextAPI{
		currentContext: currentContext,
		filename:       filename,
		contexts:       contexts,
	}

	return nil
}

func (a *api) ContextAPI() ContextAPI {
	return a.contextAPI
}

func (a *api) Close() error {
	return a.contextAPI.Save()
}

type contextAPI struct {
	dirty          bool
	filename       string
	currentContext *AppContext
	contexts       []*AppContext
}

func (a *contextAPI) CurrentContext() *AppContext {
	return a.currentContext
}

func (a *contextAPI) UseContext(name string) {
	for i := range a.contexts {
		if a.contexts[i].Name == name {
			a.currentContext = a.contexts[i]
			return
		}
	}
	a.currentContext = &AppContext{
		Name: name,
	}
	a.contexts = append(a.contexts, a.currentContext)
}

func (a *contextAPI) MarkDirty() {
	a.dirty = true
}

func (a *contextAPI) Save() error {
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

	contexts := make([]appContext, 0, len(a.contexts))
	for i := range a.contexts {

		ids := make([]string, 0, len(a.contexts[i].Services))
		for j := range a.contexts[i].Services {
			ids = append(ids, a.contexts[i].Services[j].String())
		}

		var host string
		if a.contexts[i].Host != nil {
			host = a.contexts[i].Host.String()
		}

		contexts = append(contexts, appContext{
			Name:       a.contexts[i].Name,
			Host:       host,
			ServiceIDs: ids,
		})
	}

	var currentContext string
	if a.currentContext != nil {
		currentContext = a.currentContext.Name
	}

	config := &config{
		CurrentContext: currentContext,
		Contexts:       contexts,
	}

	if err := json.NewEncoder(f).Encode(config); err != nil {
		return err
	}

	return os.Rename(backup, a.filename)
}
