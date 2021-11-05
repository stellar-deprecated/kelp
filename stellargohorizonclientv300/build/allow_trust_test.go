package build

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stellar/go/xdr"
)

var _ = Describe("AllowTrustBuilder Mutators", func() {

	var (
		subject AllowTrustBuilder
		mut     interface{}

		address = "GAXEMCEXBERNSRXOEKD4JAIKVECIXQCENHEBRVSPX2TTYZPMNEDSQCNQ"
		bad     = "foo"
	)

	JustBeforeEach(func() {
		subject = AllowTrustBuilder{}
		subject.Mutate(mut)
	})

	Describe("Trustor", func() {
		Context("using a valid stellar address", func() {
			BeforeEach(func() { mut = Trustor{address} })

			It("succeeds", func() {
				Expect(subject.Err).NotTo(HaveOccurred())
			})

			It("sets the destination to the correct xdr.AccountId", func() {
				var aid xdr.AccountId
				aid.SetAddress(address)
				Expect(subject.AT.Trustor.MustEd25519()).To(Equal(aid.MustEd25519()))
			})
		})

		Context("using an invalid value", func() {
			BeforeEach(func() { mut = Trustor{bad} })
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

	Describe("AllowTrustAsset", func() {
		Context("AssetTypeCreditAlphanum4", func() {
			BeforeEach(func() {
				mut = AllowTrustAsset{"USD"}
			})

			It("sets Asset properly", func() {
				Expect(subject.AT.Asset.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum4))
				Expect(*subject.AT.Asset.AssetCode4).To(Equal(xdr.AssetCode4{'U', 'S', 'D', 0}))
				Expect(subject.AT.Asset.AssetCode12).To(BeNil())
			})
		})

		Context("AssetTypeCreditAlphanum12", func() {
			BeforeEach(func() {
				mut = AllowTrustAsset{"ABCDEF"}
			})

			It("sets Asset properly", func() {
				Expect(subject.AT.Asset.Type).To(Equal(xdr.AssetTypeAssetTypeCreditAlphanum12))
				Expect(subject.AT.Asset.AssetCode4).To(BeNil())
				Expect(*subject.AT.Asset.AssetCode12).To(Equal(xdr.AssetCode12{'A', 'B', 'C', 'D', 'E', 'F', 0, 0, 0, 0, 0, 0}))
			})
		})

		Context("asset code length invalid", func() {
			Context("empty", func() {
				BeforeEach(func() {
					mut = AllowTrustAsset{""}
				})

				It("failed", func() {
					Expect(subject.Err.Error()).To(ContainSubstring("Asset code length is invalid"))
				})
			})

			Context("too long", func() {
				BeforeEach(func() {
					mut = AllowTrustAsset{"1234567890123"}
				})

				It("failed", func() {
					Expect(subject.Err.Error()).To(ContainSubstring("Asset code length is invalid"))
				})
			})
		})
	})

	Describe("Authorize", func() {
		Context("when equal true", func() {
			BeforeEach(func() {
				mut = Authorize{uint32(xdr.TrustLineFlagsAuthorizedFlag)}
			})

			It("sets authorize flag properly", func() {
				Expect(subject.AT.Authorize).To(Equal(xdr.Uint32(xdr.TrustLineFlagsAuthorizedFlag)))
			})
		})

		Context("when equal false", func() {
			BeforeEach(func() {
				subject.AT.Authorize = 0
				mut = Authorize{0}
			})

			It("sets authorize flag properly", func() {
				Expect(subject.AT.Authorize).To(Equal(xdr.Uint32(0)))
			})
		})
	})
})
