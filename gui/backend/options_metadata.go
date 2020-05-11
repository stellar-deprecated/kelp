package backend

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/nikhilsaraf/go-tools/multithreading"
	"github.com/stellar/kelp/api"
	"github.com/stellar/kelp/support/networking"
	"github.com/stellar/kelp/support/sdk"
	"github.com/stellar/kelp/support/utils"
)

type metadata interface{}

type dropdownType struct {
	Type    string                    `json:"type"`
	Options map[string]dropdownOption `json:"options"`
}

var _ metadata = dropdownType{}

type textType struct {
	Type         string      `json:"type"`
	DefaultValue string      `json:"defaultValue"`
	Subtype      interface{} `json:"subtype"`
}

var _ metadata = textType{}

type dropdownOption struct {
	Value   string   `json:"value"`
	Text    string   `json:"text"`
	Subtype metadata `json:"subtype"`
}

const dropdownString = "dropdown"
const textString = "text"
const cmcListURL = "https://s2.coinmarketcap.com/generated/search/quick_search.json"

var ccxtExchangeNames = map[string]string{
	"bitfinex2":     "Bitfinex [alternate]",
	"hitbtc2":       "Hitbtc [alternate]",
	"_1broker":      "1broker",
	"_1btcxe":       "1btcxe",
	"coinmarketcap": "CoinMarketCap",
}

var ccxtBlacklist = utils.StringSet([]string{
	"_1broker",
	"allcoin",
	"anybits",
	"bibox",
	"bitmex",
	"bitsane",
	"bitz",
	"btctradeim",
	"ccex",
	"coinegg",
	"coingi",
	"cointiger",
	"coolcoin",
	"cryptopia",
	"flowbtc",
	"gatecoin",
	"huobicny",
	"kucoin",
	"liqui",
	"rightbtc",
	"theocean",
	"tidex",
	"wex",
	"xbtce",
	"yunbi",
})

func dropdown(options *dropdownOptionsBuilder) dropdownType {
	return dropdownType{
		Type:    dropdownString,
		Options: options._build(),
	}
}

func text(defaultValue string) textType {
	return textType{
		Type:         textString,
		DefaultValue: defaultValue,
		Subtype:      nil,
	}
}

type dropdownOptionsBuilder struct {
	m map[string]dropdownOption
}

func optionsBuilder() *dropdownOptionsBuilder {
	return &dropdownOptionsBuilder{
		m: map[string]dropdownOption{},
	}
}

func (dob *dropdownOptionsBuilder) ccxtMarket(marketCode string) *dropdownOptionsBuilder {
	return dob._leaf(marketCode, marketCode)
}

func (dob *dropdownOptionsBuilder) krakenMarket(marketCode string, marketName string) *dropdownOptionsBuilder {
	return dob._leaf(marketCode, marketName)
}

func (dob *dropdownOptionsBuilder) coinmarketcap(tickerCode string, currencyName string) *dropdownOptionsBuilder {
	return dob._leaf(fmt.Sprintf("https://api.coinmarketcap.com/v1/ticker/%s/", tickerCode), currencyName)
}

func (dob *dropdownOptionsBuilder) currencylayer(tickerCode string, name string) *dropdownOptionsBuilder {
	value := tickerCode
	text := fmt.Sprintf("%s - %s", tickerCode, name)
	return dob._leaf(value, text)
}

func (dob *dropdownOptionsBuilder) _leaf(value string, text string) *dropdownOptionsBuilder {
	return dob.option(value, text, nil)
}

func (dob *dropdownOptionsBuilder) option(value string, text string, subtype metadata) *dropdownOptionsBuilder {
	dob.m[value] = dropdownOption{
		Value:   value,
		Text:    text,
		Subtype: subtype,
	}
	return dob
}

func (dob *dropdownOptionsBuilder) includeOptions(other *dropdownOptionsBuilder) *dropdownOptionsBuilder {
	for key, value := range other._build() {
		dob.option(key, value.Text, value.Subtype)
	}
	return dob
}

func (dob *dropdownOptionsBuilder) _build() map[string]dropdownOption {
	return dob.m
}

type cmcListItem struct {
	ID     uint32 `json:"id"`
	Name   string `json:"name"`
	Rank   uint32 `json:"rank"`
	Slug   string `json:"slug"`
	Symbol string `json:"symbol"`
}

func fetchCmcSlug2NameMap() (map[string]string, error) {
	m := map[string]string{}

	var list []cmcListItem
	e := networking.JSONRequest(http.DefaultClient, "GET", cmcListURL, "", nil, &list, "error")
	if e != nil {
		return m, fmt.Errorf("error fetching list of currencies from coinmarketcap: %s", e)
	}

	for _, item := range list {
		m[item.Slug] = item.Name + " - " + item.Symbol
	}
	return m, nil
}

func loadOptionsMetadata() (metadata, error) {
	cmcSlug2Name, e := fetchCmcSlug2NameMap()
	if e != nil {
		return nil, fmt.Errorf("cannot load CMC slug2Name map: %s", e)
	}
	coinmarketcapOptions := optionsBuilder()
	for slug, name := range cmcSlug2Name {
		coinmarketcapOptions.coinmarketcap(slug, name)
	}
	log.Printf("loaded %d currencies from coinmarketcap\n", len(cmcSlug2Name))

	totalCcxtExchanges := len(sdk.GetExchangeList())
	log.Printf("loading %d exchanges from ccxt\n", totalCcxtExchanges)
	ccxtOptionsChan := make(chan *dropdownOption, totalCcxtExchanges)
	threadTracker := multithreading.MakeThreadTracker()
	for _, ccxtExchangeName := range sdk.GetExchangeList() {
		if _, ok := ccxtBlacklist[ccxtExchangeName]; ok {
			ccxtOptionsChan <- nil
			continue
		}

		e = threadTracker.TriggerGoroutine(func(inputs []interface{}) {
			ccxtExchangeName := inputs[0].(string)
			ccxtOptionsChan := inputs[1].(chan *dropdownOption)

			displayName := strings.Title(ccxtExchangeName)
			if name, ok := ccxtExchangeNames[ccxtExchangeName]; ok {
				displayName = name
			}
			displayName = displayName + " (via CCXT)"

			c, e := sdk.MakeInitializedCcxtExchange(ccxtExchangeName, api.ExchangeAPIKey{}, []api.ExchangeParam{}, []api.ExchangeHeader{})
			if e != nil {
				// don't block if we are unable to load an exchange
				log.Printf("unable to make ccxt exchange '%s' when trying to load options metadata, continuing: %s\n", ccxtExchangeName, e)
				ccxtOptionsChan <- nil
				return
			}

			marketsBuilder := optionsBuilder()
			for tradingPair := range c.GetMarkets() {
				if strings.Count(tradingPair, "/") != 1 {
					// ignore because the trading pair does not have exactly one '/'
					// do not log because that gets too noisy and is not something that we can fix either
					ccxtOptionsChan <- nil
					return
				}

				marketsBuilder.ccxtMarket(tradingPair)
			}

			ccxtOptionsChan <- &dropdownOption{
				Value:   "ccxt-" + ccxtExchangeName,
				Text:    displayName,
				Subtype: dropdown(marketsBuilder),
			}
		}, []interface{}{ccxtExchangeName, ccxtOptionsChan})
		if e != nil {
			log.Printf("error loading ccxt exchange '%s': %s", ccxtExchangeName, e)
		}
	}
	log.Printf("kicked off initialization of all CCXT instances, waiting for all threads to finish ...")
	threadTracker.Wait()

	close(ccxtOptionsChan)
	ccxtOptions := optionsBuilder()
	for i := 0; i < totalCcxtExchanges; i++ {
		dop := <-ccxtOptionsChan
		if dop == nil {
			continue
		}

		ccxtOptions.option(dop.Value, dop.Text, dop.Subtype)
		log.Printf("added ccxt exchange option '%s' for '%s'", dop.Value, dop.Text)
	}
	log.Printf("... all CCXT instances initialized")

	builder := optionsBuilder().
		option("crypto", "Crypto (CMC)", dropdown(
			coinmarketcapOptions)).
		option("exchange", "Centralized Exchange", dropdown(optionsBuilder().
			option("kraken", "Kraken", dropdown(optionsBuilder().
				krakenMarket("XXLM/ZUSD", "XLM/USD"))).
			includeOptions(ccxtOptions))).
		option("fiat", "Fiat (CurrencyLayer)", dropdown(optionsBuilder().
			currencylayer("AED", "United Arab Emirates Dirham").
			currencylayer("AFN", "Afghan Afghani").
			currencylayer("ALL", "Albanian Lek").
			currencylayer("AMD", "Armenian Dram").
			currencylayer("ANG", "Netherlands Antillean Guilder").
			currencylayer("AOA", "Angolan Kwanza").
			currencylayer("ARS", "Argentine Peso").
			currencylayer("AUD", "Australian Dollar").
			currencylayer("AWG", "Aruban Florin").
			currencylayer("AZN", "Azerbaijani Manat").
			currencylayer("BAM", "Bosnia-Herzegovina Convertible Mark").
			currencylayer("BBD", "Barbadian Dollar").
			currencylayer("BDT", "Bangladeshi Taka").
			currencylayer("BGN", "Bulgarian Lev").
			currencylayer("BHD", "Bahraini Dinar").
			currencylayer("BIF", "Burundian Franc").
			currencylayer("BMD", "Bermudan Dollar").
			currencylayer("BND", "Brunei Dollar").
			currencylayer("BOB", "Bolivian Boliviano").
			currencylayer("BRL", "Brazilian Real").
			currencylayer("BSD", "Bahamian Dollar").
			currencylayer("BTC", "Bitcoin").
			currencylayer("BTN", "Bhutanese Ngultrum").
			currencylayer("BWP", "Botswanan Pula").
			currencylayer("BYR", "Belarusian Ruble").
			currencylayer("BZD", "Belize Dollar").
			currencylayer("CAD", "Canadian Dollar").
			currencylayer("CDF", "Congolese Franc").
			currencylayer("CHF", "Swiss Franc").
			currencylayer("CLF", "Chilean Unit of Account (UF)").
			currencylayer("CLP", "Chilean Peso").
			currencylayer("CNY", "Chinese Yuan").
			currencylayer("COP", "Colombian Peso").
			currencylayer("CRC", "Costa Rican Colón").
			currencylayer("CUC", "Cuban Convertible Peso").
			currencylayer("CUP", "Cuban Peso").
			currencylayer("CVE", "Cape Verdean Escudo").
			currencylayer("CZK", "Czech Republic Koruna").
			currencylayer("DJF", "Djiboutian Franc").
			currencylayer("DKK", "Danish Krone").
			currencylayer("DOP", "Dominican Peso").
			currencylayer("DZD", "Algerian Dinar").
			currencylayer("EGP", "Egyptian Pound").
			currencylayer("ERN", "Eritrean Nakfa").
			currencylayer("ETB", "Ethiopian Birr").
			currencylayer("EUR", "Euro").
			currencylayer("FJD", "Fijian Dollar").
			currencylayer("FKP", "Falkland Islands Pound").
			currencylayer("GBP", "British Pound Sterling").
			currencylayer("GEL", "Georgian Lari").
			currencylayer("GGP", "Guernsey Pound").
			currencylayer("GHS", "Ghanaian Cedi").
			currencylayer("GIP", "Gibraltar Pound").
			currencylayer("GMD", "Gambian Dalasi").
			currencylayer("GNF", "Guinean Franc").
			currencylayer("GTQ", "Guatemalan Quetzal").
			currencylayer("GYD", "Guyanaese Dollar").
			currencylayer("HKD", "Hong Kong Dollar").
			currencylayer("HNL", "Honduran Lempira").
			currencylayer("HRK", "Croatian Kuna").
			currencylayer("HTG", "Haitian Gourde").
			currencylayer("HUF", "Hungarian Forint").
			currencylayer("IDR", "Indonesian Rupiah").
			currencylayer("ILS", "Israeli New Sheqel").
			currencylayer("IMP", "Manx pound").
			currencylayer("INR", "Indian Rupee").
			currencylayer("IQD", "Iraqi Dinar").
			currencylayer("IRR", "Iranian Rial").
			currencylayer("ISK", "Icelandic Króna").
			currencylayer("JEP", "Jersey Pound").
			currencylayer("JMD", "Jamaican Dollar").
			currencylayer("JOD", "Jordanian Dinar").
			currencylayer("JPY", "Japanese Yen").
			currencylayer("KES", "Kenyan Shilling").
			currencylayer("KGS", "Kyrgystani Som").
			currencylayer("KHR", "Cambodian Riel").
			currencylayer("KMF", "Comorian Franc").
			currencylayer("KPW", "North Korean Won").
			currencylayer("KRW", "South Korean Won").
			currencylayer("KWD", "Kuwaiti Dinar").
			currencylayer("KYD", "Cayman Islands Dollar").
			currencylayer("KZT", "Kazakhstani Tenge").
			currencylayer("LAK", "Laotian Kip").
			currencylayer("LBP", "Lebanese Pound").
			currencylayer("LKR", "Sri Lankan Rupee").
			currencylayer("LRD", "Liberian Dollar").
			currencylayer("LSL", "Lesotho Loti").
			currencylayer("LTL", "Lithuanian Litas").
			currencylayer("LVL", "Latvian Lats").
			currencylayer("LYD", "Libyan Dinar").
			currencylayer("MAD", "Moroccan Dirham").
			currencylayer("MDL", "Moldovan Leu").
			currencylayer("MGA", "Malagasy Ariary").
			currencylayer("MKD", "Macedonian Denar").
			currencylayer("MMK", "Myanma Kyat").
			currencylayer("MNT", "Mongolian Tugrik").
			currencylayer("MOP", "Macanese Pataca").
			currencylayer("MRO", "Mauritanian Ouguiya").
			currencylayer("MUR", "Mauritian Rupee").
			currencylayer("MVR", "Maldivian Rufiyaa").
			currencylayer("MWK", "Malawian Kwacha").
			currencylayer("MXN", "Mexican Peso").
			currencylayer("MYR", "Malaysian Ringgit").
			currencylayer("MZN", "Mozambican Metical").
			currencylayer("NAD", "Namibian Dollar").
			currencylayer("NGN", "Nigerian Naira").
			currencylayer("NIO", "Nicaraguan Córdoba").
			currencylayer("NOK", "Norwegian Krone").
			currencylayer("NPR", "Nepalese Rupee").
			currencylayer("NZD", "New Zealand Dollar").
			currencylayer("OMR", "Omani Rial").
			currencylayer("PAB", "Panamanian Balboa").
			currencylayer("PEN", "Peruvian Nuevo Sol").
			currencylayer("PGK", "Papua New Guinean Kina").
			currencylayer("PHP", "Philippine Peso").
			currencylayer("PKR", "Pakistani Rupee").
			currencylayer("PLN", "Polish Zloty").
			currencylayer("PYG", "Paraguayan Guarani").
			currencylayer("QAR", "Qatari Rial").
			currencylayer("RON", "Romanian Leu").
			currencylayer("RSD", "Serbian Dinar").
			currencylayer("RUB", "Russian Ruble").
			currencylayer("RWF", "Rwandan Franc").
			currencylayer("SAR", "Saudi Riyal").
			currencylayer("SBD", "Solomon Islands Dollar").
			currencylayer("SCR", "Seychellois Rupee").
			currencylayer("SDG", "Sudanese Pound").
			currencylayer("SEK", "Swedish Krona").
			currencylayer("SGD", "Singapore Dollar").
			currencylayer("SHP", "Saint Helena Pound").
			currencylayer("SLL", "Sierra Leonean Leone").
			currencylayer("SOS", "Somali Shilling").
			currencylayer("SRD", "Surinamese Dollar").
			currencylayer("STD", "São Tomé and Príncipe Dobra").
			currencylayer("SVC", "Salvadoran Colón").
			currencylayer("SYP", "Syrian Pound").
			currencylayer("SZL", "Swazi Lilangeni").
			currencylayer("THB", "Thai Baht").
			currencylayer("TJS", "Tajikistani Somoni").
			currencylayer("TMT", "Turkmenistani Manat").
			currencylayer("TND", "Tunisian Dinar").
			currencylayer("TOP", "Tongan Paʻanga").
			currencylayer("TRY", "Turkish Lira").
			currencylayer("TTD", "Trinidad and Tobago Dollar").
			currencylayer("TWD", "New Taiwan Dollar").
			currencylayer("TZS", "Tanzanian Shilling").
			currencylayer("UAH", "Ukrainian Hryvnia").
			currencylayer("UGX", "Ugandan Shilling").
			currencylayer("USD", "United States Dollar").
			currencylayer("UYU", "Uruguayan Peso").
			currencylayer("UZS", "Uzbekistan Som").
			currencylayer("VEF", "Venezuelan Bolívar Fuerte").
			currencylayer("VND", "Vietnamese Dong").
			currencylayer("VUV", "Vanuatu Vatu").
			currencylayer("WST", "Samoan Tala").
			currencylayer("XAF", "CFA Franc BEAC").
			currencylayer("XAG", "Silver (troy ounce)").
			currencylayer("XAU", "Gold (troy ounce)").
			currencylayer("XCD", "East Caribbean Dollar").
			currencylayer("XDR", "Special Drawing Rights").
			currencylayer("XOF", "CFA Franc BCEAO").
			currencylayer("XPF", "CFP Franc").
			currencylayer("YER", "Yemeni Rial").
			currencylayer("ZAR", "South African Rand").
			currencylayer("ZMK", "Zambian Kwacha (pre-2013)").
			currencylayer("ZMW", "Zambian Kwacha").
			currencylayer("ZWL", "Zimbabwean Dollar"))).
		option("fixed", "Fixed Value", text("1.0"))
	mdata := dropdown(builder)
	return mdata, nil
}

func (s *APIServer) optionsMetadata(w http.ResponseWriter, r *http.Request) {
	s.writeJsonWithLog(w, s.cachedOptionsMetadata, false)
}
