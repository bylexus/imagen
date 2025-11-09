package cli

import (
	"flag"
	"strings"

	"github.com/bylexus/imagen/pkg/server"
)

// ServeCommand handles the 'serve' command
type ServeCommand struct {
	listen string
}

// Execute runs the serve command
func (c *ServeCommand) Execute(args []string) error {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	fs.StringVar(&c.listen, "listen", ":3000", "Listen address(es), comma-separated")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Parse listen addresses
	addresses := strings.Split(c.listen, ",")
	for i, addr := range addresses {
		addresses[i] = strings.TrimSpace(addr)
	}

	// Create and start server
	srv := server.NewServer(addresses)
	return srv.Start()
}
