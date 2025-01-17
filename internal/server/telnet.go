package server

import (
	"context"
	"errors"
	"io"
	"net"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/gabe565/ascii-movie/internal/movie"
	"github.com/gabe565/ascii-movie/internal/server/telnet"
	flag "github.com/spf13/pflag"
)

var telnetListeners uint8

type TelnetServer struct {
	MovieServer
}

func NewTelnet(flags *flag.FlagSet) TelnetServer {
	return TelnetServer{MovieServer: NewMovieServer(flags, TelnetFlagPrefix)}
}

func (s *TelnetServer) Listen(ctx context.Context, m *movie.Movie) error {
	s.Log.WithField("address", s.Address).Info("Starting Telnet server")

	listen, err := net.Listen("tcp", s.Address)
	if err != nil {
		return err
	}
	defer func(listen net.Listener) {
		_ = listen.Close()
	}(listen)

	var serveGroup sync.WaitGroup
	serveCtx, serveCancel := context.WithCancel(context.Background())
	defer serveCancel()

	go func() {
		telnetListeners += 1
		defer func() {
			telnetListeners -= 1
		}()

		for {
			conn, err := listen.Accept()
			if err != nil {
				select {
				case <-ctx.Done():
					return
				default:
					s.Log.WithError(err).Error("Failed to accept connection")
					continue
				}
			}

			serveGroup.Add(1)
			go func() {
				defer serveGroup.Done()
				s.Handler(serveCtx, conn, m)
			}()
		}
	}()

	<-ctx.Done()
	s.Log.Info("Stopping Telnet server")
	defer s.Log.Info("Stopped Telnet server")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	go func() {
		serveCancel()
		serveGroup.Wait()
		shutdownCancel()
	}()
	<-shutdownCtx.Done()

	return listen.Close()
}

func (s *TelnetServer) Handler(ctx context.Context, conn net.Conn, m *movie.Movie) {
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	remoteIP := RemoteIp(conn.RemoteAddr().String())
	logger := s.Log.WithField("remote_ip", remoteIP)

	id, err := serverInfo.StreamConnect("telnet", remoteIP)
	if err != nil {
		logger.Error(err)
		_, _ = conn.Write([]byte(ErrorText(err) + "\n"))
		return
	}
	defer serverInfo.StreamDisconnect(id)

	inR, inW := io.Pipe()
	outR, outW := io.Pipe()
	defer func() {
		_ = outR.Close()
		_ = inR.Close()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	player := movie.NewPlayer(m, logger)
	program := tea.NewProgram(
		player,
		tea.WithInput(inR),
		tea.WithOutput(outW),
		tea.WithFPS(30),
	)

	if timeout != 0 {
		go func() {
			timer := time.NewTimer(timeout)
			defer timer.Stop()
			select {
			case <-timer.C:
				cancel()
			case <-ctx.Done():
			}
		}()
	}

	go func() {
		<-ctx.Done()
		program.Send(movie.Quit())
	}()

	go func() {
		// Proxy output to client
		_, _ = io.Copy(conn, outR)
		cancel()
		_, _ = io.Copy(io.Discard, outR)
	}()

	go func() {
		// Proxy input to program
		_ = telnet.Proxy(conn, inW)
		cancel()
	}()

	if _, err := program.Run(); err != nil && !errors.Is(err, tea.ErrProgramKilled) {
		logger.WithError(err).Error("Stream failed")
	}

	program.Kill()
}
