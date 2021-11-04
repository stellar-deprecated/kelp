package build

import (
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/stellar/go/xdr"
)

var _ = Describe("ClearData", func() {
	var (
		subject ManageDataBuilder
		name    string
	)

	JustBeforeEach(func() {
		subject = ClearData(name)
	})

	Context("Valid name", func() {
		BeforeEach(func() {
			name = "my data"
		})

		It("succeeds", func() {
			Expect(subject.Err).ToNot(HaveOccurred())
			Expect(subject.MD.DataName).To(Equal(xdr.String64("my data")))
			Expect(subject.MD.DataValue).To(BeNil())
		})
	})

	Context("Long key", func() {
		BeforeEach(func() { name = strings.Repeat("a", 65) })

		It("errors", func() {
			Expect(subject.Err).To(HaveOccurred())
		})
	})

	Context("empty key", func() {
		BeforeEach(func() { name = "" })

		It("errors", func() {
			Expect(subject.Err).To(HaveOccurred())
		})
	})
})

var _ = Describe("SetData", func() {
	var (
		subject ManageDataBuilder
		name    string
		value   []byte
	)

	JustBeforeEach(func() {
		subject = SetData(name, value)
	})

	Context("Valid name and value", func() {
		BeforeEach(func() {
			name = "my data"
			value = []byte{0xFF, 0xFF}
		})

		It("succeeds", func() {
			Expect(subject.Err).ToNot(HaveOccurred())
			Expect(subject.MD.DataName).To(Equal(xdr.String64("my data")))
			Expect(*subject.MD.DataValue).To(Equal(xdr.DataValue([]byte{0xFF, 0xFF})))
		})
	})

	Context("empty value", func() {
		BeforeEach(func() {
			name = "some name"
			value = []byte{}
		})

		It("succeeds", func() {
			Expect(subject.Err).ToNot(HaveOccurred())
			Expect(subject.MD.DataName).To(Equal(xdr.String64("some name")))
		})
	})

	Context("Long key", func() {
		BeforeEach(func() {
			name = strings.Repeat("a", 65)
			value = []byte{}
		})

		It("errors", func() {
			Expect(subject.Err).To(HaveOccurred())
		})
	})

	Context("empty key", func() {
		BeforeEach(func() {
			name = ""
			value = []byte{}
		})

		It("errors", func() {
			Expect(subject.Err).To(HaveOccurred())
		})
	})

	Context("nil value", func() {
		BeforeEach(func() {
			name = "some name"
			value = nil
		})

		It("errors", func() {
			Expect(subject.Err).To(HaveOccurred())
		})
	})

	Context("Long value", func() {
		BeforeEach(func() {
			name = "some name"
			value = []byte(strings.Repeat("a", 65))
		})

		It("errors", func() {
			Expect(subject.Err).To(HaveOccurred())
		})
	})
})

var _ = Describe("ManageData Mutators", func() {

	var (
		subject ManageDataBuilder
		mut     interface{}

		address = "GAXEMCEXBERNSRXOEKD4JAIKVECIXQCENHEBRVSPX2TTYZPMNEDSQCNQ"
		bad     = "foo"
	)

	JustBeforeEach(func() {
		subject = ManageDataBuilder{}
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
})
