package build

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("TransactionEnvelope Mutators:", func() {

	var (
		subject TransactionEnvelopeBuilder
		mut     TransactionEnvelopeMutator
		err     error
	)

	BeforeEach(func() { subject = TransactionEnvelopeBuilder{} })
	JustBeforeEach(func() { err = subject.Mutate(mut) })

	Describe("TransactionBuilder", func() {
		Context("that is valid", func() {
			BeforeEach(func() {
				var err2 error
				mut, err2 = Transaction(Sequence{10})
				Expect(err2).NotTo(HaveOccurred())
			})
			It("succeeds", func() { Expect(err).NotTo(HaveOccurred()) })
			It("sets the TX", func() { Expect(subject.E.V0.Tx.SeqNum).To(BeEquivalentTo(10)) })
		})
	})

	Describe("Sign", func() {
		Context("with a valid key", func() {
			BeforeEach(func() {
				subject.MutateTX(SourceAccount{"SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"}, TestNetwork)
				mut = Sign{"SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"}
			})

			It("succeeds", func() { Expect(err).NotTo(HaveOccurred()) })
			It("adds a signature to the envelope", func() {
				Expect(subject.E.V0.Signatures).To(HaveLen(1))
			})
		})

		Context("with an invalid key", func() {
			BeforeEach(func() { mut = Sign{""} })

			It("fails", func() {
				Expect(err).To(HaveOccurred())
			})
		})
	})

})
