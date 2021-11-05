package build

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBuild(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Package: github.com/stellar/go/build")
}

// ExampleTransactionBuilder creates and signs a simple transaction, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
//
// It uses the transaction builder system
func ExampleTransactionBuilder() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		Payment(
			Destination{"GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"},
			NativeAmount{"50"},
		),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAALSRpLtCLv2eboZlEiHDSGR6Hb+zZL92fbSdNpObeE0EAAAAAAAAAAB3NZQAAAAAAAAAAARtDMfAAAABA2oIeQxoJl53RMRWFeLB865zcky39f2gf2PmUubCuJYccEePRSrTC8QQrMOgGwD8a6oe8dgltvezdDsmmXBPyBw==
}

// ExampleCreateAccount creates a transaction to fund a new stallar account with a balance. It then
// encodes the transaction into a base64 string capable of being submitted to stellar-core. It uses
// the transaction builder system.
func ExampleCreateAccount() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		CreateAccount(
			Destination{"GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"},
			NativeAmount{"50"},
		),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAAAAAAALSRpLtCLv2eboZlEiHDSGR6Hb+zZL92fbSdNpObeE0EAAAAAHc1lAAAAAAAAAAABG0Mx8AAAAEDiHQEurLITX87zmkEi8Rrcf5wGp1JrLnSDoTJiN+yNjJZVF3WcBJgoGyIJ3NJo+tNmTqALVrJziiiZGdoukxcN
}

// ExampleBumpSequence creates a transaction to bump sequence of a given account. It then
// encodes the transaction into a base64 string capable of being submitted to stellar-core. It uses
// the transaction builder system.
func ExampleBumpSequence() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		BumpSequence(
			BumpTo(5),
		),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAsAAAAAAAAABQAAAAAAAAABG0Mx8AAAAEAZkMhF/4tin/eiAw+yj13iQ9xApg0wadSyIMdk2D73RvwvgxwqfAGF6qmL1d8qqrjClZ7Vi0MYBDPxOQchwMQA
}

// ExamplePayment creates and signs a native-asset Payment, encodes it into a base64 string capable of
// being submitted to stellar-core. It uses the transaction builder system.
func ExamplePayment() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		Payment(
			Destination{"GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"},
			NativeAmount{"50"},
		),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAEAAAAALSRpLtCLv2eboZlEiHDSGR6Hb+zZL92fbSdNpObeE0EAAAAAAAAAAB3NZQAAAAAAAAAAARtDMfAAAABA2oIeQxoJl53RMRWFeLB865zcky39f2gf2PmUubCuJYccEePRSrTC8QQrMOgGwD8a6oe8dgltvezdDsmmXBPyBw==
}

// ExamplePathPayment creates and signs a simple transaction with PathPayment operation, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExamplePathPayment() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		Payment(
			Destination{"GBDT3K42LOPSHNAEHEJ6AVPADIJ4MAR64QEKKW2LQPBSKLYD22KUEH4P"},
			CreditAmount{"USD", "GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA", "50"},
			PayWith(CreditAsset("EUR", "GCPZJ3MJQ3GUGJSBL6R3MLYZS6FKVHG67BPAINMXL3NWNXR5S6XG657P"), "100").
				Through(Asset{Native: true}).
				Through(CreditAsset("BTC", "GAHJZHVKFLATAATJH46C7OK2ZOVRD47GZBGQ7P6OCVF6RJDCEG5JMQBQ")),
		),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAIAAAABRVVSAAAAAACflO2Jhs1DJkFfo7YvGZeKqpze+F4ENZde22bePZeubwAAAAA7msoAAAAAAEc9q5pbnyO0BDkT4FXgGhPGAj7kCKVbS4PDJS8D1pVCAAAAAVVTRAAAAAAALSRpLtCLv2eboZlEiHDSGR6Hb+zZL92fbSdNpObeE0EAAAAAHc1lAAAAAAIAAAAAAAAAAUJUQwAAAAAADpyeqirBMAJpPzwvuVrLqxHz5shND7/OFUvopGIhupYAAAAAAAAAARtDMfAAAABA5xuIJu/KGKQRuDrdkzNsR4HjT6wX464SHZ/yvYwVb/AkAyyfeMLDNhgKbBxQMWc3Uo5fTst1UHldC+jYNeAhCQ==
}

// ExampleSetOptions creates and signs a simple transaction with SetOptions operation, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleSetOptions() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		SetOptions(
			InflationDest("GCT7S5BA6ZC7SV7GGEMEYJTWOBYTBOA7SC4JEYP7IAEDG7HQNIWKRJ4G"),
			SetAuthRequired(),
			SetAuthRevocable(),
			SetAuthImmutable(),
			ClearAuthRequired(),
			ClearAuthRevocable(),
			ClearAuthImmutable(),
			MasterWeight(1),
			SetThresholds(2, 3, 4),
			HomeDomain("stellar.org"),
			AddSigner("GC6DDGPXVWXD5V6XOWJ7VUTDYI7VKPV2RAJWBVBHR47OPV5NASUNHTJW", 5),
		),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAUAAAABAAAAAKf5dCD2RflX5jEYTCZ2cHEwuB+QuJJh/0AIM3zwaiyoAAAAAQAAAAcAAAABAAAABwAAAAEAAAABAAAAAQAAAAIAAAABAAAAAwAAAAEAAAAEAAAAAQAAAAtzdGVsbGFyLm9yZwAAAAABAAAAALwxmfetrj7X13WT+tJjwj9VPrqIE2DUJ48+59etBKjTAAAABQAAAAAAAAABG0Mx8AAAAECZF17pOfZcyc7YJXMyx++PMydIvL6g2yZcPDY8h4+tmlz+3rsE6uuX0R6xfgNnuMntvK4YMmaOvp4DvaZMMNoA
}

// ExampleSetOptions_manyOperations creates and signs a simple transaction with many SetOptions operations, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleSetOptions_manyOperations() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		InflationDest("GCT7S5BA6ZC7SV7GGEMEYJTWOBYTBOA7SC4JEYP7IAEDG7HQNIWKRJ4G"),
		SetAuthRequired(),
		SetAuthRevocable(),
		SetAuthImmutable(),
		ClearAuthRequired(),
		ClearAuthRevocable(),
		ClearAuthImmutable(),
		MasterWeight(1),
		SetThresholds(2, 3, 4),
		HomeDomain("stellar.org"),
		RemoveSigner("GC6DDGPXVWXD5V6XOWJ7VUTDYI7VKPV2RAJWBVBHR47OPV5NASUNHTJW"),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAETAAAAAAAAAABAAAAAAAAAAAAAAALAAAAAAAAAAUAAAABAAAAAKf5dCD2RflX5jEYTCZ2cHEwuB+QuJJh/0AIM3zwaiyoAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAAAAAAAAQAAAAEAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAAAAAAAAQAAAAIAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAAAAAAAAQAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAABAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAABAAAAAgAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAABAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAEAAAABAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABQAAAAAAAAAAAAAAAAAAAAAAAAABAAAAAgAAAAEAAAADAAAAAQAAAAQAAAAAAAAAAAAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAALc3RlbGxhci5vcmcAAAAAAAAAAAAAAAAFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAABAAAAALwxmfetrj7X13WT+tJjwj9VPrqIE2DUJ48+59etBKjTAAAAAAAAAAAAAAABG0Mx8AAAAEAOXsLbFo3e8fpqyeZEHGP9o/IrQDQRyof+DA1EeUkvUGbNhy57xXcpMhZpRtwXThWBYx4za4q+TRrnoZQtezgN
}

// ExampleChangeTrust creates and signs a simple transaction with ChangeTrust operation, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleChangeTrust() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		Trust("USD", "GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA", Limit("100.25")),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABVVNEAAAAAAAtJGku0Iu/Z5uhmUSIcNIZHodv7Nkv3Z9tJ02k5t4TQQAAAAA7wO+gAAAAAAAAAAEbQzHwAAAAQOIy19X38Y3jcFzvhDsmXu6iDzrzb4iwfS2NAq9GGAFiRJUGoFX85vKtlNcXzQppF4X8oIMNPEb74fuZE/N+GAE=
}

// ExampleChangeTrust_maxLimit creates and signs a simple transaction with ChangeTrust operation (maximum limit), and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleChangeTrust_maxLimit() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		Trust("USD", "GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAYAAAABVVNEAAAAAAAtJGku0Iu/Z5uhmUSIcNIZHodv7Nkv3Z9tJ02k5t4TQX//////////AAAAAAAAAAEbQzHwAAAAQJQC6R3RqNaw5rOmaxqpAE0lD5onM/njn9I2RVlhtS2SGi2Z7xm65USYVWXTJFVqTCfTwwu+QXFcOuqgJjVtHAk=
}

// ExampleRemoveTrust creates and signs a simple transaction with ChangeTrust operation (remove trust), and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleRemoveTrust() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	operationSource := "GCVJCNUHSGKOTBBSXZJ7JJZNOSE2YDNGRLIDPMQDUEQWJQSE6QZSDPNU"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		RemoveTrust(
			"USD",
			"GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA",
			SourceAccount{operationSource},
		),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAQAAAACqkTaHkZTphDK+U/SnLXSJrA2mitA3sgOhIWTCRPQzIQAAAAYAAAABVVNEAAAAAAAtJGku0Iu/Z5uhmUSIcNIZHodv7Nkv3Z9tJ02k5t4TQQAAAAAAAAAAAAAAAAAAAAEbQzHwAAAAQD5FeGBEwJyeauK+WKfcxYBeKw62EtCqvC0p9Z+1cY32fKQ+5Jz9uE1LaDsHW5NurtStKcUTiG5j2qNDf1QpYgw=
}

// ExampleManageOffer creates and signs a simple transaction with ManageOffer operations, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleManageOffer() {
	rate := Rate{
		Selling: NativeAsset(),
		Buying:  CreditAsset("USD", "GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"),
		Price:   Price("125.12"),
	}

	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		CreateOffer(rate, "20"),
		UpdateOffer(rate, "40", OfferID(2)),
		DeleteOffer(rate, OfferID(1)),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAABLAAAAAAAAAABAAAAAAAAAAAAAAADAAAAAAAAAAMAAAAAAAAAAVVTRAAAAAAALSRpLtCLv2eboZlEiHDSGR6Hb+zZL92fbSdNpObeE0EAAAAAC+vCAAAADDgAAAAZAAAAAAAAAAAAAAAAAAAAAwAAAAAAAAABVVNEAAAAAAAtJGku0Iu/Z5uhmUSIcNIZHodv7Nkv3Z9tJ02k5t4TQQAAAAAX14QAAAAMOAAAABkAAAAAAAAAAgAAAAAAAAADAAAAAAAAAAFVU0QAAAAAAC0kaS7Qi79nm6GZRIhw0hkeh2/s2S/dn20nTaTm3hNBAAAAAAAAAAAAAAw4AAAAGQAAAAAAAAABAAAAAAAAAAEbQzHwAAAAQBfosk+t8qpULHP4ppNX2xVPih8lmnbHFZdeuxSP6pgpCCX05S7zZ4PsjVQY2nOnLru6mBTc1r8So+vxHs3FXAc=
}

// ExampleCreatePassiveOffer creates and signs a simple transaction with CreatePassiveOffer operation, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleCreatePassiveOffer() {
	rate := Rate{
		Selling: NativeAsset(),
		Buying:  CreditAsset("USD", "GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"),
		Price:   Price("125.12"),
	}

	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		CreatePassiveOffer(rate, "20"),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAQAAAAAAAAAAVVTRAAAAAAALSRpLtCLv2eboZlEiHDSGR6Hb+zZL92fbSdNpObeE0EAAAAAC+vCAAAADDgAAAAZAAAAAAAAAAEbQzHwAAAAQHv/1xLn+ArfIUoWjn3V0zVka6tulqMYx4zJZhGqdmTw8iCXY0ZtHS+y+7YGgR3vM1DpKOdvWTmhee+sCXIppQA=
}

// ExampleAccountMerge creates and signs a simple transaction with AccountMerge operation, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleAccountMerge() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		AccountMerge(
			Destination{"GBDT3K42LOPSHNAEHEJ6AVPADIJ4MAR64QEKKW2LQPBSKLYD22KUEH4P"},
		),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAgAAAAARz2rmlufI7QEORPgVeAaE8YCPuQIpVtLg8MlLwPWlUIAAAAAAAAAARtDMfAAAABAh3qZrP5T9Xg0LdzwOLx/eA/B7bzj+8j+s9eXNuu7/Ldch7I6kW5iYz6Vfy32FVnKNtoykToB7nQY2o2vo1tqAw==
}

// ExampleInflation creates and signs a simple transaction with Inflation operation, and then
// encodes it into a base64 string capable of being submitted to stellar-core.
func ExampleInflation() {
	seed := "SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"
	tx, err := Transaction(
		SourceAccount{seed},
		Sequence{1},
		TestNetwork,
		Inflation(),
	)

	if err != nil {
		fmt.Println(err)
		return
	}

	txe, err := tx.Sign(seed)
	if err != nil {
		fmt.Println(err)
		return
	}

	txeB64, err := txe.Base64()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("tx base64: %s", txeB64)
	// Output: tx base64: AAAAADZY/nWY0gx6beMpf4S8Ur0qHsjA8fbFtBzBx1cbQzHwAAAAZAAAAAAAAAABAAAAAAAAAAAAAAABAAAAAAAAAAkAAAAAAAAAARtDMfAAAABAzzDG4V7KzynWY0ER/V4HH0WgDvl3hrIizDcKW3qEQY4Ib3yXufVvdbzsET/Dj5js5dgDkcYgikHwRCpqi/J8BQ==
}
