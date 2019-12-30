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
	"github.com/stellar/kelp/database"
	"github.com/stellar/kelp/model"
	"github.com/stellar/kelp/support/postgresdb"
	"github.com/stellar/kelp/support/utils"
)

const marketIdHashLength = 10

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

func makeMarketID(exchangeName string, baseAsset string, quoteAsset string) string {
	idString := fmt.Sprintf("%s_%s_%s", exchangeName, baseAsset, quoteAsset)
	h := sha256.New()
	h.Write([]byte(idString))
	sha256Hash := fmt.Sprintf("%x", h.Sum(nil))
	return sha256Hash[0:marketIdHashLength]
}

// makeTradingMarket makes a market along with the ID field
func makeTradingMarket(exchangeName string, baseAsset string, quoteAsset string) *tradingMarket {
	sha256HashPrefix := makeMarketID(exchangeName, baseAsset, quoteAsset)
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
func MakeFillDBWriter(db *sql.DB, assetDisplayFn model.AssetDisplayFn, exchangeName string) api.FillHandler {
	return &FillDBWriter{
		db:             db,
		assetDisplayFn: assetDisplayFn,
		exchangeName:   exchangeName,
	}
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
	rows, e := f.db.Query(database.SqlQueryMarketsById, marketId)
	if e != nil {
		return nil, fmt.Errorf("could not execute sql select query (%s) for marketId (%s): %s", database.SqlQueryMarketsById, marketId, e)
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
	sqlInsert := fmt.Sprintf(database.SqlMarketsInsertTemplate,
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
	dateString := date.Format(postgresdb.TimestampFormatString)

	market, e := f.fetchOrRegisterMarket(trade)
	if e != nil {
		return fmt.Errorf("cannot fetch or register market for trade (txid=%s): %s", txid, e)
	}

	sqlInsert := fmt.Sprintf(database.SqlTradesInsertTemplate,
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
