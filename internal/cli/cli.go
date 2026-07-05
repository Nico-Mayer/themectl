package cli

import (
	"github.com/nico-mayer/themectl-cli/internal/config"
	"github.com/nico-mayer/themectl-cli/internal/engine"
	"github.com/nico-mayer/themectl-cli/internal/theme"
	urfaveCli "github.com/urfave/cli/v3"
)

func New(cfg config.Config, store *theme.Store, engine *engine.Engine) *urfaveCli.Command {
	return &urfaveCli.Command{
		Commands: []*urfaveCli.Command{
			listCmd(store),
			setCmd(cfg, store, engine),
		},
	}
}
