package build

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stellar/go/xdr"
)

var _ = Describe("ChangeTrustBuilder Mutators", func() {

	var (
		subject ChangeTrustBuilder
		mut     interface{}

		address = "GAXEMCEXBERNSRXOEKD4JAIKVECIXQCENHEBRVSPX2TTYZPMNEDSQCNQ"
		bad     = "foo"
	)

	JustBeforeEach(func() {
		subject = ChangeTrustBuilder{}
		subject.Mutate(mut)
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

	Describe("Line", func() {
		Context("AssetTypeCreditAlphanum4", func() {
			BeforeEach(func() {
				mut = CreditAsset("USD", address)
			})

			It("sets Asset properly", func() {
				Expect(subject.CT.Line.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum4))
				Expect(subject.CT.Line.AlphaNum4.AssetCode).To(Equal(xdr.AssetCode4{'U', 'S', 'D', 0}))
				Expect(subject.CT.Line.AlphaNum12).To(BeNil())
			})
		})

		Context("AssetTypeCreditAlphanum12", func() {
			BeforeEach(func() {
				mut = CreditAsset("ABCDEF", address)
			})

			It("sets Asset properly", func() {
				Expect(subject.CT.Line.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum12))
				Expect(subject.CT.Line.AlphaNum4).To(BeNil())
				Expect(subject.CT.Line.AlphaNum12.AssetCode).To(Equal(xdr.AssetCode12{'A', 'B', 'C', 'D', 'E', 'F', 0, 0, 0, 0, 0, 0}))
			})
		})

		Context("asset invalid", func() {
			Context("native", func() {
				BeforeEach(func() {
					mut = NativeAsset()
				})

				It("failed", func() {
					Expect(subject.Err.Error()).To(ContainSubstring("Native asset not allowed"))
				})
			})

			Context("empty", func() {
				BeforeEach(func() {
					mut = CreditAsset("", address)
				})

				It("failed", func() {
					Expect(subject.Err.Error()).To(ContainSubstring("Asset code length is invalid"))
				})
			})

			Context("too long", func() {
				BeforeEach(func() {
					mut = CreditAsset("1234567890123", address)
				})

				It("failed", func() {
					Expect(subject.Err.Error()).To(ContainSubstring("Asset code length is invalid"))
				})
			})
		})

		Context("issuer invalid", func() {
			BeforeEach(func() {
				mut = CreditAsset("USD", bad)
			})

			It("failed", func() {
				Expect(subject.Err).To(HaveOccurred())
			})
		})
	})

	Describe("Limit", func() {
		Context("sets limit properly", func() {
			BeforeEach(func() {
				mut = Limit("20")
			})

			It("sets limit value properly", func() {
				Expect(subject.CT.Limit).To(Equal(xdr.Int64(200000000)))
			})
		})

		Context("sets max limit properly", func() {
			BeforeEach(func() {
				mut = MaxLimit
			})

			It("sets limit value properly", func() {
				Expect(subject.CT.Limit).To(Equal(xdr.Int64(9223372036854775807)))
			})
		})
	})
})
