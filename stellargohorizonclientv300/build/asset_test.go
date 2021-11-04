package build

import (
	"testing"

	"github.com/stellar/go/xdr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsset_ToXDR(t *testing.T) {
	var (
		issuer        xdr.AccountId
		issuerAddress = "GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"
	)
	require.NoError(t, issuer.SetAddress(issuerAddress))

	cases := []struct {
		Name        string
		Asset       Asset
		Expected    xdr.Asset
		ExpectedErr string
	}{
		{
			Name:     "Native",
			Asset:    NativeAsset(),
			Expected: xdr.Asset{Type: xdr.AssetTypeAssetTypeNative},
		},
		{
			Name:  "Alphanum4",
			Asset: CreditAsset("USD", issuerAddress),
			Expected: xdr.Asset{
				Type: xdr.AssetTypeAssetTypeCreditAlphanum4,
				AlphaNum4: &xdr.AssetAlphaNum4{
					AssetCode: xdr.AssetCode4{0x55, 0x53, 0x44, 0x00}, //USD
					Issuer:    issuer,
				},
			},
		},
		{
			Name:  "Alphanum12",
			Asset: CreditAsset("SCOTTBUCKS", issuerAddress),
			Expected: xdr.Asset{
				Type: xdr.AssetTypeAssetTypeCreditAlphanum12,
				AlphaNum12: &xdr.AssetAlphaNum12{
					AssetCode: [12]byte{
						0x53, 0x43, 0x4f, 0x54,
						0x54, 0x42, 0x55, 0x43,
						0x4b, 0x53, 0x00, 0x00,
					}, //SCOTTBUCKS
					Issuer: issuer,
				},
			},
		},
		{
			Name:        "bad issuer",
			Asset:       CreditAsset("USD", "FUNK"),
			ExpectedErr: "strkey is 4 bytes long; minimum valid length is 5",
		},
		{
			Name:        "bad code",
			Asset:       CreditAsset("", issuerAddress),
			ExpectedErr: "Asset code length is invalid",
		},
	}

	for _, kase := range cases {
		actual, err := kase.Asset.ToXDR()

		if kase.ExpectedErr != "" {
			if assert.Error(t, err, ("no expected error in case: " + kase.Name)) {
				assert.EqualError(t, err, kase.ExpectedErr)
			}
			continue
		}

		if assert.NoError(t, err, "unexpected error in case: %s", kase.Name) {
			assert.Equal(t, kase.Expected, actual, "invalid xdr result")
		}
	}
}

func TestAsset_MustXDR(t *testing.T) {
	// good
	assert.NotPanics(t, func() {
		NativeAsset().MustXDR()
	})

	// bad
	assert.Panics(t, func() {
		CreditAsset("USD", "BONK").MustXDR()
	})
}
