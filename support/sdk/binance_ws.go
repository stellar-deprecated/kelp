package sdk

type EventBinance struct {
	Name string `json:"e"`
}

type EventExecutionReport struct {
	EventBinance
	EventTime                    int64       `json:"E"` // "E": 1499405658658,           //Event time
	Symbol                       string      `json:"s"` // "s": "ETHBTC",                //Symbol
	ClientOrderID                string      `json:"c"` // "c": "mUvoqJxFIILMdfAW5iGSOW",  //Client order ID
	Side                         string      `json:"S"` // "S": "BUY",                    // Side
	OrderType                    string      `json:"o"` // "o": "LIMIT",                  // Order type
	TimeInForce                  string      `json:"f"` // "f": "GTC",                    // Time in force
	OrderQuantity                string      `json:"q"` // "q": "1.00000000",             // Order quantity
	OrderPrice                   string      `json:"p"` // "p": "0.10264410",             // Order price
	StopPrice                    string      `json:"P"` // "P": "0.00000000",             // Stop price
	IceberQuantity               string      `json:"F"` // "F": "0.00000000",             // Iceberg quantity
	OrderListID                  int64       `json:"g"` // "g": -1,                       // OrderListId
	OriginalClientID             interface{} `json:"C"` // "C": null,                     // Original client order ID; This is the ID of the order being canceled
	CurrentExecutionType         string      `json:"x"` // "x": "NEW",                    // Current execution type
	CurrentOrderStatus           string      `json:"X"` // "X": "NEW",                    // Current order status
	OrderRejectReason            string      `json:"r"` // "r": "NONE",                   // Order reject reason; will be an error code.
	OrderID                      int64       `json:"i"` // "i": 4293153,                  // Order ID
	LastExecutedQuantity         string      `json:"l"` // "l": "0.00000000",             // Last executed quantity
	CumulativeFillerQuantity     string      `json:"z"` // "z": "0.00000000",             // Cumulative filled quantity
	LastExecutedPrice            string      `json:"L"` // "L": "0.00000000",             // Last executed price
	ComissionAmount              string      `json:"n"` // "n": "0",                      // Commission amount
	ComissionAsset               interface{} `json:"N"` // "N": null,                     // Commission asset
	TransactionTime              int64       `json:"T"` // "T": 1499405658657,            // Transaction time
	TradeID                      int64       `json:"t"` // "t": -1,                       // Trade ID
	Ignore                       int64       `json:"I"` // "I": 8641984,                  // Ignore
	IsTherOrderOnBook            bool        `json:"w"` // "w": true,                     // Is the order on the book?
	IsTheTradeMarkerSide         bool        `json:"m"` // "m": false,                    // Is this trade the maker side?
	IsIgnore                     bool        `json:"M"` // "M": false,                    // Ignore
	OrderCreationTime            int64       `json:"O"` // "O": 1499405658657,            // Order creation time
	CumulativeQuoteAssetQuantity string      `json:"Z"` // "Z": "0.00000000",             // Cumulative quote asset transacted quantity
	LastQuoteAssetQuantity       string      `json:"Y"` // "Y": "0.00000000",              // Last quote asset transacted quantity (i.e. lastPrice * lastQty)
	QuoteOrderQuantity           string      `json:"Q"` // "Q": "0.00000000"              // Quote Order Qty
}
