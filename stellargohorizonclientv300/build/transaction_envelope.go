package build

import (
	"bytes"
	"encoding/base64"
	"fmt"

	"github.com/stellar/go/keypair"
	"github.com/stellar/go/support/errors"
	"github.com/stellar/go/xdr"
)

// TransactionEnvelopeMutator is a interface that wraps the
// MutateTransactionEnvelope operation.  types may implement this interface to
// specify how they modify an xdr.TransactionEnvelope object
type TransactionEnvelopeMutator interface {
	MutateTransactionEnvelope(*TransactionEnvelopeBuilder) error
}

// TransactionEnvelopeBuilder helps you build a TransactionEnvelope
type TransactionEnvelopeBuilder struct {
	E *xdr.TransactionEnvelope
	// TODO support fee bump transactions

	child *TransactionBuilder
}

func (b *TransactionEnvelopeBuilder) Init() {
	if b.E == nil {
		b.E = &xdr.TransactionEnvelope{
			Type: xdr.EnvelopeTypeEnvelopeTypeTxV0,
			V0:   &xdr.TransactionV0Envelope{},
		}
	}

	if b.child == nil {
		// TODO support fee bump transactions
		b.child = &TransactionBuilder{TX: &b.E.V0.Tx}
	}
}

// Mutate applies the provided TransactionEnvelopeMutators to this builder's
// envelope
func (b *TransactionEnvelopeBuilder) Mutate(muts ...TransactionEnvelopeMutator) error {
	b.Init()

	for i, m := range muts {
		err := m.MutateTransactionEnvelope(b)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("mutator:%d failed", i))
		}
	}

	return nil
}

// MutateTX runs Mutate on the underlying transaction using the provided
// mutators.
func (b *TransactionEnvelopeBuilder) MutateTX(muts ...TransactionMutator) error {
	b.Init()

	err := b.child.Mutate(muts...)
	if err != nil {
		return err
	}

	return nil
}

// Bytes encodes the builder's underlying envelope to XDR
func (b *TransactionEnvelopeBuilder) Bytes() ([]byte, error) {
	var txBytes bytes.Buffer
	_, err := xdr.Marshal(&txBytes, b.E)
	if err != nil {
		return nil, errors.Wrap(err, "marshal xdr failed")
	}

	return txBytes.Bytes(), nil
}

// Base64 returns a string which is the xdr-then-base64-encoded form
// of the builder's underlying transaction envelope
func (b *TransactionEnvelopeBuilder) Base64() (string, error) {
	bs, err := b.Bytes()
	if err != nil {
		return "", errors.Wrap(err, "get raw bytes failed")
	}

	return base64.StdEncoding.EncodeToString(bs), nil
}

// ------------------------------------------------------------
//
//   Mutator implementations
//
// ------------------------------------------------------------

// MutateTransactionEnvelope adds a signature to the provided envelope
func (m Sign) MutateTransactionEnvelope(txe *TransactionEnvelopeBuilder) error {
	hash, err := txe.child.Hash()
	if err != nil {
		return errors.Wrap(err, "hash tx failed")
	}

	kp, err := keypair.Parse(m.Seed)
	if err != nil {
		return errors.Wrap(err, "parse failed")
	}

	sig, err := kp.SignDecorated(hash[:])
	if err != nil {
		return errors.Wrap(err, "sign tx failed")
	}

	txe.E.V0.Signatures = append(txe.E.V0.Signatures, sig)
	return nil
}

// MutateTransactionEnvelope for TransactionBuilder causes the underylying
// transaction to be set as the provided envelope's Tx field
func (m *TransactionBuilder) MutateTransactionEnvelope(txe *TransactionEnvelopeBuilder) error {
	// TODO support fee bump transactions
	txe.E.V0.Tx = *m.TX
	newChild := *m
	txe.child = &newChild
	m.TX = &txe.E.V0.Tx
	return nil
}
