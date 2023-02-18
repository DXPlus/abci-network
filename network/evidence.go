package network

import (
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"math"

	bls "github.com/chuwt/chia-bls-go"

	ecrypto "github.com/ethereum/go-ethereum/crypto"
)

const EvidenceMaxSize = 10

type Authorize struct {
	// BLS public key
	BLSPublicKey []byte
	// Signature for BLS public key
	SignatureForBLS []byte
	// ECDSA Public key
	ESDSAPublicKey []byte
}

type TransmissionEvidence struct {
	OwnerAuthorize Authorize
	PublicKeyList  [][]byte
	ShareList      [][]byte // 0~1
	Signature      []byte
	Count          int
}

type EvidenceManagement struct {
	PublicKey  bls.PublicKey
	PrivateKey bls.PrivateKey
	// Transmission incentive strategy
}

func NewEvidenceManagement(privateKeyBytes []byte) *EvidenceManagement {
	privateKey := bls.KeyFromBytes(privateKeyBytes)
	return &EvidenceManagement{
		PublicKey:  privateKey.GetPublicKey(),
		PrivateKey: privateKey,
	}
}

func (em *EvidenceManagement) VerifyEvidence(te TransmissionEvidence) (bool, error) {
	if len(te.PublicKeyList) > EvidenceMaxSize {
		return false, errors.New("over max evidence size")
	}

	// Verify authorize sturct
	//   Verify ECDSA signature
	dataHash := sha256.Sum256(te.OwnerAuthorize.BLSPublicKey)
	isRight := ecrypto.VerifySignature(te.OwnerAuthorize.ESDSAPublicKey, dataHash[:], te.OwnerAuthorize.SignatureForBLS)

	if !isRight {
		return false, errors.New("ecdsa signature verify failed")
	}

	// Verify BLS Aggregate Signature
	asm := bls.AugSchemeMPL{}
	isRight = asm.AggregateVerify(te.PublicKeyList, te.ShareList, te.Signature)

	if !isRight {
		return false, errors.New("aggregate verify failed")
	}

	return true, nil
}

func (em *EvidenceManagement) AddEvidence(te *TransmissionEvidence) error {
	// Get share for this evidence
	share := em.getShare()

	// Sign share with own privite key
	asm := bls.AugSchemeMPL{}
	sign := asm.Sign(em.PrivateKey, Float32ToByte(share))

	// Aggregate old sign and own sign
	newSign, err := asm.Aggregate(te.Signature, sign)
	if err != nil {
		return err
	}

	// Update evidence
	te.Signature = newSign
	te.PublicKeyList = append(te.PublicKeyList, em.PublicKey.Bytes())
	te.ShareList = append(te.ShareList, Float32ToByte(share))

	return nil
}

// Get share
func (em *EvidenceManagement) getShare() float32 {
	return 0.2
}

func Float32ToByte(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}
