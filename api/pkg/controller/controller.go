package controller

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/lukemarsden/helix/api/pkg/filestore"
	"github.com/lukemarsden/helix/api/pkg/model"
	"github.com/lukemarsden/helix/api/pkg/store"
	"github.com/lukemarsden/helix/api/pkg/types"
	"github.com/rs/zerolog/log"
)

type ControllerOptions struct {
	Store     store.Store
	Filestore filestore.FileStore
	// this is an "env" prefix like "dev"
	// the user prefix is handled inside the controller
	// (see getFilestorePath)
	FilePrefixGlobal string
	// this is a golang template that is used to prefix the user
	// path in the filestore - it is passed Owner and OwnerType values
	// write me an example FilePrefixUser as a go template
	// e.g. "users/{{.Owner}}"
	FilePrefixUser string
	// a static path used to denote what sub-folder job results live in
	FilePrefixResults string
}

type Controller struct {
	Ctx                context.Context
	Options            ControllerOptions
	SessionUpdatesChan chan *types.Session
	// the backlog of sessions that need a GPU
	sessionQueue    []*types.Session
	sessionQueueMtx sync.Mutex

	// keep a map of instantiated models so we can ask it about memory
	// the models package looks after instantiating this for us
	models map[types.ModelName]model.Model
}

func NewController(
	ctx context.Context,
	options ControllerOptions,
) (*Controller, error) {
	if options.Store == nil {
		return nil, fmt.Errorf("store is required")
	}
	if options.Filestore == nil {
		return nil, fmt.Errorf("filestore is required")
	}
	models, err := model.GetModels()
	if err != nil {
		return nil, err
	}
	controller := &Controller{
		Ctx:                ctx,
		Options:            options,
		SessionUpdatesChan: make(chan *types.Session),
		sessionQueue:       []*types.Session{},
		models:             models,
	}
	return controller, nil
}

func (c *Controller) Initialize() error {
	err := c.loadSessionQueues(c.Ctx)
	if err != nil {
		return err
	}
	return nil
}

// this should be run in a go-routine
func (c *Controller) StartLooping() {
	for {
		select {
		case <-c.Ctx.Done():
			return
		case <-time.After(10 * time.Second):
			err := c.loop(c.Ctx)
			if err != nil {
				log.Error().Msgf("error in controller loop: %s", err.Error())
				debug.PrintStack()
			}
		}
	}
}

func (c *Controller) loop(ctx context.Context) error {
	return nil
}
