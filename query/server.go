package query

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

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
	var e error
	for {
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			command := scanner.Text()
			output, e := s.executeCommandIPC(command)
			if e != nil {
				return fmt.Errorf("error while executing IPC Command ('%s'): %s", command, e)
			}

			if len(output) > 0 {
				if !strings.HasSuffix(output, "\n") {
					output += "\n"
				}

				_, e = os.Stdout.WriteString(output)
				if e != nil {
					return fmt.Errorf("error while writing output to Stdout (name=%s): %s", os.Stdout.Name(), e)
				}
			}
		}

		if e = scanner.Err(); e != nil {
			return fmt.Errorf("error while reading commands in query server: %s", e)
		}

		time.Sleep(5 * time.Second)
	}
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
