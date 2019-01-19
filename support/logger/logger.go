package logger

import (
	"os"
)

// Logger is the base logger interface
type Logger interface {
	// basic messages, appends a newline (\n) after each entry
	Info(msg string)

	// basic messages, can be custom formatted, similar to fmt.Printf. User needs to add a \n if they want a newline after the log entry
	Infof(msg string, args ...interface{})

	// error messages, indicates to the logger that these messages can be handled differently (different color, special format, email alerts, etc.). The type of logger will determine what to do with these messages. The logger should NOT panic on these messages. Appends a newline (\n) after each entry.
	Error(msg string)

	// error messages, indicates to the logger that these messages can be handled differently (different color, special format, email alerts, etc.). The type of logger will determine what to do with these messages. The logger should NOT panic on these messages. User needs to add a \n if they want a newline after the log entry.
	Errorf(msg string, args ...interface{})
}

// Fatal is a convenience method that can be used with any Logger to log a fatal error
func Fatal(l Logger, e error) {
	l.Info("")
	l.Errorf("%s", e)
	os.Exit(1)
}

/*
Notes:
	What to do with log lines followed by panic? Should those just go to logger.Fatal and exit?
	What to do with utils.PriceAsFloat? If you have it return an error many of the referencing functions have to parse multiple errors.
		It's really bloated and there's no struct to drop a logger into
		Passing a logger into each function here seems ridiculous
		Overall this file gets nasty with the new logger framework

Files that need framework conversion:
	support/networking
		server.go - pending decision on what to do with panics
	suport/utils
		functions.go - pendimg decision on what to do with all the functions error cascading


Files ready for find-replace:
	api/
		priceFeed.go
	plugins/
		exchangeFeed.go
		ccxtExchange.go
		deleteSideStrategy.go
		fillLogger.go
		fillTracker.go
		intervalTimeController.go
		krakenExchange.go
		mirrorStrategy.go
		sdex.go
		sellSideStrategy.go
		staticSpreadLevelProvider.go
	support/monitoring/
		metricsEndpoint.go
		pagerDuty.go
	support/sdk
		ccxt.go
	support/utils
		configs.go
	terminator/
		terminator.go
	trader/
		trader.go

Files fully converted to new logger framework, i.e. don't have log output lines to replace:
	api/
		assets.go
	model/
		assets.go
		botKey.go
	cmd/
		trade.go
	plugins/
		cmcFeed.go
		fiatFeed.go
		fixedFeed.go
		balancedLevelProvider.go

Files that don't need converting (_test and non-.go files not listed):
main.go
	accounting/pnl/
		pnl.go (not actually Kelp)
	cmd/
		exchanges.go (wouldn't log this)
		root.go
		strategies.go (wouldn't log this)
		version.go
	api/
		alert.go
		exchange.go
		level.go
		strategy.go
		timeController.go
	model/
		dates.go
		number.go
		orderbook.go
		tradingPair.go
	plugins/
		balancedStrategy.go
		buysellStrategy.go
		composeStrategy.go
		deleteStrategy.go
		factory.go
		priceFeed.go
		sellStrategy.go
	support/monitoring
		factory.go
		metrics.go
		metricsRecorder.go
	support/networking
		endpoint.go
		network.go
		parser.go
	support/utils
		config.go
	terminator/
		config.go
	trader/
		config.go
*/
