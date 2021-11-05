package build

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stellar/go/xdr"
)

var _ = Describe("BumpSequenceBuilder Mutators", func() {

	var (
		subject BumpSequenceBuilder
		mut     interface{}

		address = "GAXEMCEXBERNSRXOEKD4JAIKVECIXQCENHEBRVSPX2TTYZPMNEDSQCNQ"
		bad     = "foo"
	)

	JustBeforeEach(func() {
		subject = BumpSequenceBuilder{}
		subject.Mutate(mut)
	})

	Describe("BumpTo", func() {
		BeforeEach(func() { mut = BumpTo(9223372036854775807) })

		It("succeeds", func() {
			Expect(subject.Err).NotTo(HaveOccurred())
		})

		It("sets the value to the correct xdr.SequenceNumber", func() {
			Expect(subject.BS.BumpTo).To(Equal(xdr.SequenceNumber(9223372036854775807)))
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
