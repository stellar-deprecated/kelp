package model

import (
	"crypto/sha1"
	"fmt"
	"log"
	"strings"

	hProtocol "github.com/stellar/go/protocols/horizon"
)

const botDataKeyPrefix = "b/"
const native = "native"

// BotKey is a unique key to identify a bot
type BotKey struct {
	AssetBaseCode    string
	AssetBaseIssuer  string
	AssetQuoteCode   string
	AssetQuoteIssuer string

	// uninitialized
	cachedKey  *string
	cachedHash *string
}

// String impl
func (b BotKey) String() string {
	return fmt.Sprintf("BotKey(key=%s, hash=%s)", b.Key(), b.Hash())
}

// MakeSortedBotKey makes a BotKey by sorting the passed in assets
func MakeSortedBotKey(assetA hProtocol.Asset, assetB hProtocol.Asset) *BotKey {
	var assetBaseCode, assetBaseIssuer, assetQuoteCode, assetQuoteIssuer string
	if assetA.Type == native && assetB.Type == native {
		log.Fatal("invalid asset types, both cannot be native")
	} else if assetA.Type == native {
		assetBaseCode = native
		assetBaseIssuer = ""
		assetQuoteCode = assetB.Code
		assetQuoteIssuer = assetB.Issuer
	} else if assetB.Type == native {
		assetBaseCode = native
		assetBaseIssuer = ""
		assetQuoteCode = assetA.Code
		assetQuoteIssuer = assetA.Issuer
	} else if strings.Compare(assetA.Code+"/"+assetA.Issuer, assetB.Code+"/"+assetB.Issuer) <= 0 {
		assetBaseCode = assetA.Code
		assetBaseIssuer = assetA.Issuer
		assetQuoteCode = assetB.Code
		assetQuoteIssuer = assetB.Issuer
	} else {
		assetBaseCode = assetB.Code
		assetBaseIssuer = assetB.Issuer
		assetQuoteCode = assetA.Code
		assetQuoteIssuer = assetA.Issuer
	}

	return &BotKey{
		AssetBaseCode:    assetBaseCode,
		AssetBaseIssuer:  assetBaseIssuer,
		AssetQuoteCode:   assetQuoteCode,
		AssetQuoteIssuer: assetQuoteIssuer,
	}
}

// IsBotKey checks whether a given string is a BotKey
func IsBotKey(key string) bool {
	return strings.HasPrefix(key, botDataKeyPrefix)
}

// SplitDataKey splits the data key on the account into the hash and part
func SplitDataKey(key string) (string, string) {
	keyParts := strings.Split(key, "/")
	// [0] is "b" so don't include
	hash := keyParts[1]
	part := keyParts[2]
	return hash, part
}

// HashWithPrefix returns the hash prefixed with "b/"
func (b *BotKey) HashWithPrefix() string {
	return botDataKeyPrefix + b.Hash()
}

// FullKey returns the full key to be used in the manageData operation
func (b *BotKey) FullKey(part int) string {
	return fmt.Sprintf("%s/%d", b.HashWithPrefix(), part)
}

// Key returns the unique key string for this BotKey
func (b *BotKey) Key() string {
	if b.cachedKey != nil {
		return *b.cachedKey
	}

	key := fmt.Sprintf("%s/%s/%s/%s", b.AssetBaseCode, b.AssetBaseIssuer, b.AssetQuoteCode, b.AssetQuoteIssuer)
	b.cachedKey = &key
	return key
}

// Hash returns the hash of the underlying key
func (b *BotKey) Hash() string {
	if b.cachedHash != nil {
		return *b.cachedHash
	}

	h := sha1.New()
	_, e := h.Write([]byte(b.Key()))
	if e != nil {
		log.Fatal(e)
	}
	bs := h.Sum(nil)
	hash := fmt.Sprintf("%x", bs)

	b.cachedHash = &hash
	return hash
}
