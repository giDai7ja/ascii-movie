package play

import (
	"github.com/gabe565/ascii-telnet-go/internal/server"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "play",
		Short: "Play the movie locally",
		RunE:  run,
	}
	server.PlayFlags(cmd.Flags())
	return cmd
}

func run(cmd *cobra.Command, args []string) (err error) {
	handler, err := server.New(cmd.Flags(), false)
	if err != nil {
		return err
	}

	return handler.ServeAscii(cmd.OutOrStdout())
}
