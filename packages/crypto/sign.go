package crypto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/EGaaS/go-egaas-mvp/packages/consts"
	"github.com/EGaaS/go-egaas-mvp/packages/converter"
)

type signProvider int

const (
	_ECDSA signProvider = iota
)

func Sign(privateKey string, data string) ([]byte, error) {
	if len(data) == 0 {
		log.Debug(SigningEmpty.Error())
	}
	switch signProv {
	case _ECDSA:
		return signECDSA(privateKey, data)
	default:
		return nil, UnknownProviderError
	}
}

func CheckSign(public []byte, data string, signature []byte) (bool, error) {
	if len(public) == 0 {
		log.Debug(CheckingSignEmpty.Error())
	}
	switch signProv {
	case _ECDSA:
		return checkECDSA(public, data, signature)
	default:
		return false, UnknownProviderError
	}
}

// JSSignToBytes converts hex signature which has got from the browser to []byte
func JSSignToBytes(in string) ([]byte, error) {
	r, s, err := parseSign(in)
	if err != nil {
		return nil, err
	}
	return append(converter.FillLeft(r.Bytes()), converter.FillLeft(s.Bytes())...), nil
}

func signECDSA(privateKey string, data string) (ret []byte, err error) {
	var pubkeyCurve elliptic.Curve

	switch ellipticSize {
	case elliptic256:
		pubkeyCurve = elliptic.P256()
	default:
		log.Fatal(UnsupportedCurveSize.Error())
	}

	b, err := hex.DecodeString(privateKey)
	if err != nil {
		return
	}
	bi := new(big.Int).SetBytes(b)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi

	signhash, err := Hash([]byte(data))
	if err != nil {
		log.Fatal(HashingError)
	}
	r, s, err := ecdsa.Sign(crand.Reader, priv, signhash)
	if err != nil {
		return
	}
	ret = append(converter.FillLeft(r.Bytes()), converter.FillLeft(s.Bytes())...)
	return
}

// CheckECDSA checks if forSign has been signed with corresponding to public the private key
func checkECDSA(public []byte, data string, signature []byte) (bool, error) {
	if len(data) == 0 {
		return false, fmt.Errorf("invalid parameters len(data) == 0")
	}
	if len(public) != consts.PubkeySizeLength {
		return false, fmt.Errorf("invalid parameters len(public) = %d", len(public))
	}
	if len(signature) == 0 {
		return false, fmt.Errorf("invalid parameters len(signature) == 0")
	}

	var pubkeyCurve elliptic.Curve
	switch ellipticSize {
	case elliptic256:
		pubkeyCurve = elliptic.P256()
	default:
		log.Fatal(UnsupportedCurveSize)
	}

	hash, err := Hash([]byte(data))
	if err != nil {
		log.Fatal(HashingError)
	}

	pubkey := new(ecdsa.PublicKey)
	pubkey.Curve = pubkeyCurve
	pubkey.X = new(big.Int).SetBytes(public[0 : consts.PubkeySizeLength/2])
	pubkey.Y = new(big.Int).SetBytes(public[consts.PubkeySizeLength/2:])
	r, s, err := parseSign(hex.EncodeToString(signature))
	if err != nil {
		return false, err
	}
	verifystatus := ecdsa.Verify(pubkey, hash, r, s)
	if !verifystatus {
		return false, IncorrectSign
	}
	return true, nil
}

// parseSign converts the hex signature to r and s big number
func parseSign(sign string) (*big.Int, *big.Int, error) {
	var err error

	if len(sign) != 128 {
		return nil, nil, fmt.Errorf(`wrong len of signature %d`, len(sign))
	}

	all, err := hex.DecodeString(sign[:])
	if err != nil {
		return nil, nil, err
	}

	return new(big.Int).SetBytes(all[:32]), new(big.Int).SetBytes(all[32:]), nil
}
