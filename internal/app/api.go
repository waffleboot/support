package app

import (
	"net/url"

	"github.com/google/uuid"
)

type ContextAPI interface {
	CurrentContext() *AppContext
	Contexts() []*AppContext
	AddService(uuid.UUID) error
	DelService(uuid.UUID) error
	MarkDirty()
	Close() error
}

type AppContext struct {
	Name     string
	Host     *url.URL
	Services []uuid.UUID
}
