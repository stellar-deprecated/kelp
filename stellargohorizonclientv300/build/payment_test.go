package build

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stellar/go/xdr"
)

var _ = Describe("Payment Mutators", func() {

	var (
		subject PaymentBuilder
		mut     interface{}

		address = "GAXEMCEXBERNSRXOEKD4JAIKVECIXQCENHEBRVSPX2TTYZPMNEDSQCNQ"
		bad     = "foo"
	)

	Describe("Payment", func() {
		JustBeforeEach(func() {
			subject = PaymentBuilder{}
			subject.Mutate(mut)
		})

		Describe("CreditAmount", func() {
			Context("AlphaNum4", func() {
				BeforeEach(func() {
					mut = CreditAmount{"USD", address, "50.0"}
				})
				It("sets the asset properly", func() {
					Expect(subject.P.Amount).To(Equal(xdr.Int64(500000000)))
					Expect(subject.P.Asset.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum4))
					Expect(subject.P.Asset.AlphaNum4.AssetCode).To(Equal(xdr.AssetCode4{'U', 'S', 'D', 0}))
					var aid xdr.AccountId
					aid.SetAddress(address)
					Expect(subject.P.Asset.AlphaNum4.Issuer.MustEd25519()).To(Equal(aid.MustEd25519()))
					Expect(subject.P.Asset.AlphaNum12).To(BeNil())
				})
				It("succeeds", func() {
					Expect(subject.Err).NotTo(HaveOccurred())
				})
			})

			Context("AlphaNum12", func() {
				BeforeEach(func() {
					mut = CreditAmount{"ABCDEF", address, "50.0"}
				})
				It("sets the asset properly", func() {
					Expect(subject.P.Amount).To(Equal(xdr.Int64(500000000)))
					Expect(subject.P.Asset.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum12))
					Expect(subject.P.Asset.AlphaNum4).To(BeNil())
					Expect(subject.P.Asset.AlphaNum12.AssetCode).To(Equal(xdr.AssetCode12{'A', 'B', 'C', 'D', 'E', 'F', 0, 0, 0, 0, 0, 0}))
					var aid xdr.AccountId
					aid.SetAddress(address)
					Expect(subject.P.Asset.AlphaNum12.Issuer.MustEd25519()).To(Equal(aid.MustEd25519()))
				})
				It("succeeds", func() {
					Expect(subject.Err).NotTo(HaveOccurred())
				})
			})

			Context("issuer invalid", func() {
				BeforeEach(func() {
					mut = CreditAmount{"USD", bad, "50.0"}
				})

				It("failed", func() {
					Expect(subject.Err).To(HaveOccurred())
				})
			})

			Context("amount invalid", func() {
				BeforeEach(func() {
					mut = CreditAmount{"ABCDEF", address, "test"}
				})

				It("failed", func() {
					Expect(subject.Err).To(HaveOccurred())
				})
			})

			Context("asset code length invalid", func() {
				Context("empty", func() {
					BeforeEach(func() {
						mut = CreditAmount{"", address, "50.0"}
					})

					It("failed", func() {
						Expect(subject.Err.Error()).To(ContainSubstring("Asset code length is invalid"))
					})
				})

				Context("too long", func() {
					BeforeEach(func() {
						mut = CreditAmount{"1234567890123", address, "50.0"}
					})

					It("failed", func() {
						Expect(subject.Err.Error()).To(ContainSubstring("Asset code length is invalid"))
					})
				})
			})
		})

		Describe("Destination", func() {
			Context("using a valid stellar address", func() {
				BeforeEach(func() { mut = Destination{address} })

				It("succeeds", func() {
					Expect(subject.Err).NotTo(HaveOccurred())
				})

				It("sets the destination to the correct xdr.AccountId", func() {
					var aid xdr.AccountId
					aid.SetAddress(address)
					Expect(subject.P.Destination.MustEd25519()).To(Equal(aid.MustEd25519()))
				})
			})

			Context("using an invalid value", func() {
				BeforeEach(func() { mut = Destination{bad} })
				It("failed", func() { Expect(subject.Err).To(HaveOccurred()) })
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

		Describe("NativeAmount", func() {
			BeforeEach(func() { mut = NativeAmount{"101"} })
			It("sets the starting balance properly", func() {
				Expect(subject.P.Asset.Type).To(Equal(xdr.AssetTypeAssetTypeNative))
				Expect(subject.P.Amount).To(Equal(xdr.Int64(1010000000)))
			})
			It("succeeds", func() {
				Expect(subject.Err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("PathPayment", func() {
		JustBeforeEach(func() {
			subject = PaymentBuilder{}
			subject.Mutate(PayWith(CreditAsset("EUR", "GCPZJ3MJQ3GUGJSBL6R3MLYZS6FKVHG67BPAINMXL3NWNXR5S6XG657P"), "100").
				Through(NativeAsset()).
				Through(CreditAsset("BTC", "GAHJZHVKFLATAATJH46C7OK2ZOVRD47GZBGQ7P6OCVF6RJDCEG5JMQBQ")))
			subject.Mutate(mut)
		})

		Describe("Destination", func() {
			Context("using a valid stellar address", func() {
				BeforeEach(func() { mut = Destination{address} })

				It("succeeds", func() {
					Expect(subject.Err).NotTo(HaveOccurred())
				})

				It("sets the destination to the correct xdr.AccountId", func() {
					var aid xdr.AccountId
					aid.SetAddress(address)
					Expect(subject.PP.Destination.MustEd25519()).To(Equal(aid.MustEd25519()))
				})
			})

			Context("using an invalid value", func() {
				BeforeEach(func() { mut = Destination{bad} })
				It("failed", func() { Expect(subject.Err).To(HaveOccurred()) })
			})
		})

		Describe("Destination: Asset and Amount", func() {
			Context("native", func() {
				BeforeEach(func() {
					mut = NativeAmount{"50"}
				})
				It("sets the fields properly", func() {
					Expect(subject.PP.DestAmount).To(Equal(xdr.Int64(500000000)))
					Expect(subject.PP.DestAsset.Type).To(Equal(xdr.AssetTypeAssetTypeNative))
					Expect(subject.PP.DestAsset.AlphaNum4).To(BeNil())
					Expect(subject.PP.DestAsset.AlphaNum12).To(BeNil())
				})
				It("succeeds", func() {
					Expect(subject.Err).NotTo(HaveOccurred())
				})
			})

			Context("AlphaNum4", func() {
				BeforeEach(func() {
					mut = CreditAmount{"USD", address, "50"}
				})
				It("sets the asset properly", func() {
					Expect(subject.PP.DestAmount).To(Equal(xdr.Int64(500000000)))
					Expect(subject.PP.DestAsset.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum4))
					Expect(subject.PP.DestAsset.AlphaNum4.AssetCode).To(Equal(xdr.AssetCode4{'U', 'S', 'D', 0}))
					var aid xdr.AccountId
					aid.SetAddress(address)
					Expect(subject.PP.DestAsset.AlphaNum4.Issuer.MustEd25519()).To(Equal(aid.MustEd25519()))
					Expect(subject.PP.DestAsset.AlphaNum12).To(BeNil())
				})
				It("succeeds", func() {
					Expect(subject.Err).NotTo(HaveOccurred())
				})
			})

			Context("AlphaNum12", func() {
				BeforeEach(func() {
					mut = CreditAmount{"ABCDEF", address, "50"}
				})
				It("sets the asset properly", func() {
					Expect(subject.PP.DestAmount).To(Equal(xdr.Int64(500000000)))
					Expect(subject.PP.DestAsset.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum12))
					Expect(subject.PP.DestAsset.AlphaNum4).To(BeNil())
					Expect(subject.PP.DestAsset.AlphaNum12.AssetCode).To(Equal(xdr.AssetCode12{'A', 'B', 'C', 'D', 'E', 'F', 0, 0, 0, 0, 0, 0}))
					var aid xdr.AccountId
					aid.SetAddress(address)
					Expect(subject.PP.DestAsset.AlphaNum12.Issuer.MustEd25519()).To(Equal(aid.MustEd25519()))
				})
				It("succeeds", func() {
					Expect(subject.Err).NotTo(HaveOccurred())
				})
			})

			Context("issuer invalid", func() {
				BeforeEach(func() {
					mut = CreditAmount{"ABCDEF", bad, "50"}
				})

				It("failed", func() {
					Expect(subject.Err).To(HaveOccurred())
				})
			})

			Context("amount invalid", func() {
				BeforeEach(func() {
					mut = CreditAmount{"ABCDEF", address, "test"}
				})

				It("failed", func() {
					Expect(subject.Err).To(HaveOccurred())
				})
			})

			Context("asset code length invalid", func() {
				Context("empty", func() {
					BeforeEach(func() {
						mut = CreditAmount{"", address, "50.0"}
					})

					It("failed", func() {
						Expect(subject.Err.Error()).To(ContainSubstring("Asset code length is invalid"))
					})
				})

				Context("too long", func() {
					BeforeEach(func() {
						mut = CreditAmount{"1234567890123", address, "50.0"}
					})

					It("failed", func() {
						Expect(subject.Err.Error()).To(ContainSubstring("Asset code length is invalid"))
					})
				})
			})
		})
	})
})
