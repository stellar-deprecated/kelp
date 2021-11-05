package build

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stellar/go/xdr"
)

var _ = Describe("ManageOffer", func() {

	Describe("ManageOfferBuilder", func() {
		var (
			subject ManageOfferBuilder
			mut     interface{}

			address = "GAXEMCEXBERNSRXOEKD4JAIKVECIXQCENHEBRVSPX2TTYZPMNEDSQCNQ"
			bad     = "foo"

			rate = Rate{
				Selling: CreditAsset("EUR", "GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"),
				Buying:  NativeAsset(),
				Price:   Price("41.265"),
			}
		)

		JustBeforeEach(func() {
			subject = ManageOfferBuilder{}
			subject.Mutate(mut)
		})

		Describe("CreateOffer", func() {
			Context("creates offer properly", func() {
				It("sets values properly", func() {
					builder := CreateOffer(rate, "20")

					Expect(builder.MO.Amount).To(Equal(xdr.Int64(200000000)))

					Expect(builder.MO.Selling.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum4))
					Expect(builder.MO.Selling.AlphaNum4.AssetCode).To(Equal(xdr.AssetCode4{'E', 'U', 'R', 0}))
					var aid xdr.AccountId
					aid.SetAddress(rate.Selling.Issuer)
					Expect(builder.MO.Selling.AlphaNum4.Issuer.MustEd25519()).To(Equal(aid.MustEd25519()))
					Expect(builder.MO.Selling.AlphaNum12).To(BeNil())

					Expect(builder.MO.Buying.Type).To(Equal(xdr.AssetTypeAssetTypeNative))
					Expect(builder.MO.Buying.AlphaNum4).To(BeNil())
					Expect(builder.MO.Buying.AlphaNum12).To(BeNil())

					Expect(builder.MO.Price.N).To(Equal(xdr.Int32(8253)))
					Expect(builder.MO.Price.D).To(Equal(xdr.Int32(200)))

					Expect(builder.MO.OfferId).To(Equal(xdr.Int64(0)))
				})
			})
		})

		Describe("UpdateOffer", func() {
			Context("updates the offer properly", func() {
				It("sets values properly", func() {
					builder := UpdateOffer(rate, "100", 5)

					Expect(builder.MO.Amount).To(Equal(xdr.Int64(1000000000)))

					Expect(builder.MO.Selling.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum4))
					Expect(builder.MO.Selling.AlphaNum4.AssetCode).To(Equal(xdr.AssetCode4{'E', 'U', 'R', 0}))
					var aid xdr.AccountId
					aid.SetAddress(rate.Selling.Issuer)
					Expect(builder.MO.Selling.AlphaNum4.Issuer.MustEd25519()).To(Equal(aid.MustEd25519()))
					Expect(builder.MO.Selling.AlphaNum12).To(BeNil())

					Expect(builder.MO.Buying.Type).To(Equal(xdr.AssetTypeAssetTypeNative))
					Expect(builder.MO.Buying.AlphaNum4).To(BeNil())
					Expect(builder.MO.Buying.AlphaNum12).To(BeNil())

					Expect(builder.MO.Price.N).To(Equal(xdr.Int32(8253)))
					Expect(builder.MO.Price.D).To(Equal(xdr.Int32(200)))

					Expect(builder.MO.OfferId).To(Equal(xdr.Int64(5)))
				})
			})
		})

		Describe("DeleteOffer", func() {
			Context("deletes the offer properly", func() {
				It("sets values properly", func() {
					builder := DeleteOffer(rate, 10)

					Expect(builder.MO.Amount).To(Equal(xdr.Int64(0)))

					Expect(builder.MO.Selling.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum4))
					Expect(builder.MO.Selling.AlphaNum4.AssetCode).To(Equal(xdr.AssetCode4{'E', 'U', 'R', 0}))
					var aid xdr.AccountId
					aid.SetAddress(rate.Selling.Issuer)
					Expect(builder.MO.Selling.AlphaNum4.Issuer.MustEd25519()).To(Equal(aid.MustEd25519()))
					Expect(builder.MO.Selling.AlphaNum12).To(BeNil())

					Expect(builder.MO.Buying.Type).To(Equal(xdr.AssetTypeAssetTypeNative))
					Expect(builder.MO.Buying.AlphaNum4).To(BeNil())
					Expect(builder.MO.Buying.AlphaNum12).To(BeNil())

					Expect(builder.MO.Price.N).To(Equal(xdr.Int32(8253)))
					Expect(builder.MO.Price.D).To(Equal(xdr.Int32(200)))

					Expect(builder.MO.OfferId).To(Equal(xdr.Int64(10)))
				})
			})
		})

		Describe("SourceAccount", func() {
			Context("using a valid stellar address", func() {
				BeforeEach(func() { mut = SourceAccount{address} })

				It("succeeds", func() {
					Expect(subject.Err).NotTo(HaveOccurred())
				})

				It("sets the destination to the correct xdr.AccountId", func() {
					var aid xdr.AccountId
					aid.SetAddress(address)
					Expect(subject.O.SourceAccount.MustEd25519()).To(Equal(aid.MustEd25519()))
				})
			})

			Context("using an invalid value", func() {
				BeforeEach(func() { mut = SourceAccount{bad} })
				It("failed", func() { Expect(subject.Err).To(HaveOccurred()) })
			})
		})
	})

	Describe("CreatePassiveOfferBuilder", func() {
		var (
			subject ManageOfferBuilder
			mut     interface{}

			rate = Rate{
				Selling: CreditAsset("EUR", "GAWSI2JO2CF36Z43UGMUJCDQ2IMR5B3P5TMS7XM7NUTU3JHG3YJUDQXA"),
				Buying:  NativeAsset(),
				Price:   Price("41.265"),
			}
		)

		JustBeforeEach(func() {
			subject = ManageOfferBuilder{}
			subject.Mutate(mut)
		})

		Describe("CreatePassiveOffer", func() {
			Context("creates offer properly", func() {
				It("sets values properly", func() {
					builder := CreatePassiveOffer(rate, "20")

					Expect(builder.PO.Amount).To(Equal(xdr.Int64(200000000)))

					Expect(builder.PO.Selling.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum4))
					Expect(builder.PO.Selling.AlphaNum4.AssetCode).To(Equal(xdr.AssetCode4{'E', 'U', 'R', 0}))
					var aid xdr.AccountId
					aid.SetAddress(rate.Selling.Issuer)
					Expect(builder.PO.Selling.AlphaNum4.Issuer.MustEd25519()).To(Equal(aid.MustEd25519()))
					Expect(builder.PO.Selling.AlphaNum12).To(BeNil())

					Expect(builder.PO.Buying.Type).To(Equal(xdr.AssetTypeAssetTypeNative))
					Expect(builder.PO.Buying.AlphaNum4).To(BeNil())
					Expect(builder.PO.Buying.AlphaNum12).To(BeNil())

					Expect(builder.PO.Price.N).To(Equal(xdr.Int32(8253)))
					Expect(builder.PO.Price.D).To(Equal(xdr.Int32(200)))
				})
			})
		})
	})
})
