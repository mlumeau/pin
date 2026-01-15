package wiring

import (
	"pin/internal/contracts"
	platformserver "pin/internal/platform/server"
)

type Deps struct {
	srv   *platformserver.Server
	repos contracts.Repos
}

func NewDeps(srv *platformserver.Server) Deps {
	return Deps{srv: srv, repos: srv.Repos()}
}
