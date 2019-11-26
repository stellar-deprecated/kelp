package plugins

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/postgresdb"
	"github.com/stellar/kelp/support/utils"
)

const dateFormatString = "2006/01/02 15:04:05 MST"
const sqlMarketsTableCreate = "CREATE TABLE IF NOT EXISTS markets (market_id TEXT PRIMARY KEY, exchange_name TEXT NOT NULL, base TEXT NOT NULL, quote TEXT NOT NULL)"
const sqlMarketsInsertTemplate = "INSERT INTO markets (market_id, exchange_name, base, quote) VALUES ('%s', '%s', '%s', '%s')"
const sqlFetchMarketById = "SELECT market_id, exchange_name, base, quote FROM markets WHERE market_id = $1 LIMIT 1"
const sqlTradesTableCreate = "CREATE TABLE IF NOT EXISTS trades (market_id TEXT NOT NULL, txid TEXT NOT NULL, date_utc TIMESTAMP WITHOUT TIME ZONE NOT NULL, action TEXT NOT NULL, type TEXT NOT NULL, counter_price DOUBLE PRECISION NOT NULL, base_volume DOUBLE PRECISION NOT NULL, counter_cost DOUBLE PRECISION NOT NULL, fee DOUBLE PRECISION NOT NULL, PRIMARY KEY (market_id, txid))"
const sqlTradesInsertTemplate = "INSERT INTO trades (market_id, txid, date_utc, action, type, counter_price, base_volume, counter_cost, fee) VALUES ('%s', '%s', '%s', '%s', '%s', %.15f, %.15f, %.15f, %.15f)"
const marketIdHashLength = 10

var sqlIndexes = []string{
	"CREATE INDEX IF NOT EXISTS date ON trades (market_id, date_utc)",
}

type tradingMarket struct {
	ID           string
	ExchangeName string
	BaseAsset    string
	QuoteAsset   string
}

func (t *tradingMarket) equals(other tradingMarket) bool {
	if t.ExchangeName != other.ExchangeName {
		return false
	} else if t.BaseAsset != other.BaseAsset {
		return false
	} else if t.QuoteAsset != other.QuoteAsset {
		return false
	}
	return true
}

// FillDBWriter is a FillHandler that writes fills to a SQL database
type FillDBWriter struct {
	db             *sql.DB
	assetDisplayFn model.AssetDisplayFn
	exchangeName   string

	// uninitialized
	market *tradingMarket
}

// makeTradingMarket makes a market along with the ID field
func makeTradingMarket(exchangeName string, baseAsset string, quoteAsset string) *tradingMarket {
	idString := fmt.Sprintf("%s_%s_%s", exchangeName, baseAsset, quoteAsset)
	h := sha256.New()
	h.Write([]byte(idString))
	sha256Hash := fmt.Sprintf("%x", h.Sum(nil))
	sha256HashPrefix := sha256Hash[0:marketIdHashLength]

	return &tradingMarket{
		ID:           sha256HashPrefix,
		ExchangeName: exchangeName,
		BaseAsset:    baseAsset,
		QuoteAsset:   quoteAsset,
	}
}

// String is the Stringer method
func (m *tradingMarket) String() string {
	return fmt.Sprintf("tradingMarket[ID=%s, ExchangeName=%s, BaseAsset=%s, QuoteAsset=%s]", m.ID, m.ExchangeName, m.BaseAsset, m.QuoteAsset)
}

var _ api.FillHandler = &FillDBWriter{}

// MakeFillDBWriter is a factory method
func MakeFillDBWriter(postgresDbConfig *postgresdb.Config, assetDisplayFn model.AssetDisplayFn, exchangeName string) (api.FillHandler, error) {
	dbCreated, e := postgresdb.CreateDatabaseIfNotExists(postgresDbConfig)
	if e != nil {
		if strings.Contains(e.Error(), "connect: connection refused") {
			utils.PrintErrorHintf("ensure your postgres database is available on %s:%d, or remove the 'POSTGRES_DB' config from your trader config file\n", postgresDbConfig.GetHost(), postgresDbConfig.GetPort())
		}
		return nil, fmt.Errorf("error when creating database from config (%+v), created=%v: %s", *postgresDbConfig, dbCreated, e)
	}
	if dbCreated {
		log.Printf("created database '%s'", postgresDbConfig.GetDbName())
	} else {
		log.Printf("did not create db '%s' because it already exists", postgresDbConfig.GetDbName())
	}

	db, e := sql.Open("postgres", postgresDbConfig.MakeConnectString())
	if e != nil {
		return nil, fmt.Errorf("could not open database: %s", e)
	}
	// don't defer db.Close() here becuase we want it open for the life of the application for now

	e = postgresdb.CreateTableIfNotExists(db, sqlTradesTableCreate)
	if e != nil {
		return nil, fmt.Errorf("could not create trades table: %s", e)
	}
	e = postgresdb.CreateTableIfNotExists(db, sqlMarketsTableCreate)
	if e != nil {
		return nil, fmt.Errorf("could not create markets table: %s", e)
	}

	for i, sqlIndexCreate := range sqlIndexes {
		var statement *sql.Stmt
		statement, e = db.Prepare(sqlIndexCreate)
		if e != nil {
			return nil, fmt.Errorf("could not prepare sql statement to create index (%s) (i=%d): %s", sqlIndexCreate, i, e)
		}
		_, e = statement.Exec()
		if e != nil {
			return nil, fmt.Errorf("could not execute sql statement to create index (%s) (i=%d): %s", sqlIndexCreate, i, e)
		}
	}

	fdbw := &FillDBWriter{
		db:             db,
		assetDisplayFn: assetDisplayFn,
		exchangeName:   exchangeName,
	}
	log.Printf("made FillDBWriter with db config: %s\n", postgresDbConfig.MakeConnectString())
	return fdbw, nil
}

func (f *FillDBWriter) fetchOrRegisterMarket(trade model.Trade) (*tradingMarket, error) {
	if f.market != nil {
		return f.market, nil
	}

	txid := utils.CheckedString(trade.TransactionID)
	baseAssetString, e := f.assetDisplayFn(trade.Pair.Base)
	if e != nil {
		return nil, fmt.Errorf("bot is not configured to recognize the base asset from this trade (txid=%s), base asset = %s, error: %s", txid, string(trade.Pair.Base), e)
	}
	quoteAssetString, e := f.assetDisplayFn(trade.Pair.Quote)
	if e != nil {
		return nil, fmt.Errorf("bot is not configured to recognize the quote asset from this trade (txid=%s), quote asset = %s, error: %s", txid, string(trade.Pair.Quote), e)
	}

	market := makeTradingMarket(f.exchangeName, baseAssetString, quoteAssetString)
	fetchedMarket, e := f.fetchMarketFromDb(market.ID)
	if e != nil {
		return nil, fmt.Errorf("error while fetching market (ID=%s) from db: %s", market.ID, e)
	}

	if fetchedMarket == nil {
		e = f.registerMarket(market)
		if e != nil {
			return nil, fmt.Errorf("unable to register market: %s", market.String())
		}
		log.Printf("registered market in db: %s", market.String())
	} else if !market.equals(*fetchedMarket) {
		return nil, fmt.Errorf("fetched market (%s) was different from computed market (%s)", *fetchedMarket, *market)
	}

	f.market = market
	return market, nil
}

func (f *FillDBWriter) fetchMarketFromDb(marketId string) (*tradingMarket, error) {
	rows, e := f.db.Query(sqlFetchMarketById, marketId)
	if e != nil {
		return nil, fmt.Errorf("could not execute sql select query (%s) for marketId (%s): %s", sqlFetchMarketById, marketId, e)
	}
	defer rows.Close()

	for rows.Next() {
		var market tradingMarket
		e = rows.Scan(&market.ID, &market.ExchangeName, &market.BaseAsset, &market.QuoteAsset)
		if e != nil {
			return nil, fmt.Errorf("could not scan row into tradingMarket struct: %s", e)
		}

		log.Printf("fetched market from db: %s", market.String())
		return &market, nil
	}

	return nil, nil
}

func (f *FillDBWriter) registerMarket(market *tradingMarket) error {
	sqlInsert := fmt.Sprintf(sqlMarketsInsertTemplate,
		market.ID,
		market.ExchangeName,
		market.BaseAsset,
		market.QuoteAsset,
	)

	_, e := f.db.Exec(sqlInsert)
	if e != nil {
		// duplicate insert should return an error
		return fmt.Errorf("could not execute sql insert values statement (%s): %s", sqlInsert, e)
	}

	return nil
}

// HandleFill impl.
func (f *FillDBWriter) HandleFill(trade model.Trade) error {
	txid := utils.CheckedString(trade.TransactionID)
	timeSeconds := trade.Timestamp.AsInt64() / 1000
	date := time.Unix(timeSeconds, 0).UTC()
	dateString := date.Format(dateFormatString)

	market, e := f.fetchOrRegisterMarket(trade)
	if e != nil {
		return fmt.Errorf("cannot fetch or register market for trade (txid=%s): %s", txid, e)
	}

	sqlInsert := fmt.Sprintf(sqlTradesInsertTemplate,
		market.ID,
		txid,
		dateString,
		trade.OrderAction.String(),
		trade.OrderType.String(),
		f.checkedFloat(trade.Price),
		f.checkedFloat(trade.Volume),
		f.checkedFloat(trade.Cost),
		f.checkedFloat(trade.Fee),
	)
	_, e = f.db.Exec(sqlInsert)
	if e != nil {
		if strings.Contains(e.Error(), "duplicate key value violates unique constraint \"trades_pkey\"") {
			log.Printf("trying to reinsert trade (txid=%s) to db, ignore and continue\n", txid)
			return nil
		}

		// return an error on any other errors
		return fmt.Errorf("could not execute sql insert values statement (%s): %s", sqlInsert, e)
	}

	log.Printf("wrote trade (txid=%s) to db\n", txid)
	return nil
}

func (f *FillDBWriter) checkedFloat(n *model.Number) interface{} {
	if n == nil {
		return nil
	}
	return n.AsFloat()
}
