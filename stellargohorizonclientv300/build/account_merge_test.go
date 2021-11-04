package build

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stellar/go/xdr"
)

var _ = Describe("AccountMergeBuilder Mutators", func() {
	var (
		subject AccountMergeBuilder
		mut     interface{}

		address = "GAXEMCEXBERNSRXOEKD4JAIKVECIXQCENHEBRVSPX2TTYZPMNEDSQCNQ"
		bad     = "foo"
	)

	JustBeforeEach(func() {
		subject = AccountMergeBuilder{}
		subject.Mutate(mut)
	})

	Describe("Destination", func() {
		Context("using a valid stellar address", func() {
			BeforeEach(func() { mut = Destination{address} })

			It("succeeds", func() {
				Expect(subject.Err).NotTo(HaveOccurred())
			})

			It("sets the destination to the correct xdr.MuxedAccount", func() {
				var muxed xdr.MuxedAccount
				muxed.SetAddress(address)
				Expect(subject.Destination.Type).To(Equal(muxed.Type))
				Expect(subject.Destination.Ed25519).To(Equal(muxed.Ed25519))
				Expect(subject.Destination.Med25519).To(Equal(muxed.Med25519))
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
				var muxed xdr.MuxedAccount
				muxed.SetAddress(address)
				Expect(subject.O.SourceAccount.Type).To(Equal(muxed.Type))
				Expect(subject.O.SourceAccount.Ed25519).To(Equal(muxed.Ed25519))
				Expect(subject.O.SourceAccount.Med25519).To(Equal(muxed.Med25519))
			})
		})

		Context("using an invalid value", func() {
			BeforeEach(func() { mut = SourceAccount{bad} })
			It("failed", func() { Expect(subject.Err).To(HaveOccurred()) })
		})
	})

	//
})
