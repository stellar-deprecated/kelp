package query

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/stellar/kelp/support/utils"

	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/plugins"
	"github.com/stellar/kelp/support/logger"
	"github.com/stellar/kelp/trader"
)

// Server is a query server with which the trade command will serve information about an actively running bot
type Server struct {
	l            logger.Logger
	strategyName string
	strategy     api.Strategy
	botConfig    trader.BotConfig
	client       *horizon.Client
	sdex         *plugins.SDEX
	exchangeShim api.ExchangeShim
	tradingPair  *model.TradingPair
}

// MakeServer is a factory method
func MakeServer(
	l logger.Logger,
	strategyName string,
	strategy api.Strategy,
	botConfig trader.BotConfig,
	client *horizon.Client,
	sdex *plugins.SDEX,
	exchangeShim api.ExchangeShim,
	tradingPair *model.TradingPair,
) *Server {
	return &Server{
		l:            l,
		strategyName: strategyName,
		strategy:     strategy,
		botConfig:    botConfig,
		client:       client,
		sdex:         sdex,
		exchangeShim: exchangeShim,
		tradingPair:  tradingPair,
	}
}

// StartIPC kicks off the Server which reads from Stdin and writes to Stdout, this should be run in a new goroutine
func (s *Server) StartIPC() error {
	pipeRead := os.NewFile(uintptr(3), "pipe_read")
	pipeWrite := os.NewFile(uintptr(4), "pipe_write")

	scanner := bufio.NewScanner(pipeRead)
	s.l.Infof("waiting for IPC command...\n")
	for scanner.Scan() {
		command := scanner.Text()
		s.l.Infof("...received IPC command: %s\n", command)
		output, e := s.executeCommandIPC(command)
		if e != nil {
			return fmt.Errorf("error while executing IPC Command ('%s'): %s", command, e)
		}
		if !strings.HasSuffix(output, "\n") {
			output += "\n"
		}

		output += utils.IPCBoundary + "\n"
		s.l.Infof("responding to IPC command ('%s') with output: %s", command, output)
		_, e = pipeWrite.WriteString(output)
		if e != nil {
			return fmt.Errorf("error while writing output to pipeWrite (name=%s; fd=%v): %s", pipeWrite.Name(), pipeWrite.Fd(), e)
		}
		s.l.Infof("waiting for next IPC command...\n")
	}

	if e := scanner.Err(); e != nil {
		return fmt.Errorf("error while reading commands in query server: %s", e)
	}
	return nil
}

func (s *Server) executeCommandIPC(cmd string) (string, error) {
	cmd = strings.TrimSpace(cmd)

	switch cmd {
	case "":
		return "", nil
	case "getBotInfo":
		output, e := s.getBotInfo()
		if e != nil {
			return "", fmt.Errorf("unable to get bot info: %s", e)
		}

		outputBytes, e := json.MarshalIndent(output, "", "  ")
		if e != nil {
			return "", fmt.Errorf("unable to marshall output to JSON: %s", e)
		}
		return string(outputBytes), nil
	default:
		// don't do anything if the input is an incorrect command because we take input from standard in
		return "", nil
	}
}
