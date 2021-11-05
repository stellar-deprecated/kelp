package build

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stellar/go/xdr"
)

var _ = Describe("CreateAccountBuilder Mutators", func() {

	var (
		subject CreateAccountBuilder
		mut     interface{}

		address = "GAXEMCEXBERNSRXOEKD4JAIKVECIXQCENHEBRVSPX2TTYZPMNEDSQCNQ"
		bad     = "foo"
	)

	JustBeforeEach(func() {
		subject = CreateAccountBuilder{}
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
				Expect(subject.CA.Destination.MustEd25519()).To(Equal(aid.MustEd25519()))
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
			Expect(subject.CA.StartingBalance).To(Equal(xdr.Int64(1010000000)))
		})
		It("succeeds", func() {
			Expect(subject.Err).NotTo(HaveOccurred())
		})
	})
})
