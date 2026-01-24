package wiring

import (
	"pin/internal/contracts"
	pinserver "pin/internal/platform/server"
)

type Deps struct {
	srv   *pinserver.Server
	repos contracts.Repos
}

// NewDeps constructs a new deps.
func NewDeps(srv *pinserver.Server) Deps {
	return Deps{srv: srv, repos: srv.Repos()}
}
