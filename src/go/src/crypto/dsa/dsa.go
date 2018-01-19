// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package dsa implements the Digital Signature Algorithm, as defined in FIPS 186-3.
//
// The DSA operations in this package are not implemented using constant-time algorithms.
package dsa

import (
	"errors"
	"io"
	"math/big"
)

// Parameters represents the domain parameters for a key. These parameters can
// be shared across many keys. The bit length of Q must be a multiple of 8.
type Parameters struct {
	P, Q, G *big.Int
}

// PublicKey represents a DSA public key.
type PublicKey struct {
	Parameters
	Y *big.Int
}

// PrivateKey represents a DSA private key.
type PrivateKey struct {
	PublicKey
	X *big.Int
}

// ErrInvalidPublicKey results when a public key is not usable by this code.
// FIPS is quite strict about the format of DSA keys, but other code may be
// less so. Thus, when using keys which may have been generated by other code,
// this error must be handled.
var ErrInvalidPublicKey = errors.New("crypto/dsa: invalid public key")

// ParameterSizes is a enumeration of the acceptable bit lengths of the primes
// in a set of DSA parameters. See FIPS 186-3, section 4.2.
type ParameterSizes int

const (
	L1024N160 ParameterSizes = iota
	L2048N224
	L2048N256
	L3072N256
)

// numMRTests is the number of Miller-Rabin primality tests that we perform. We
// pick the largest recommended number from table C.1 of FIPS 186-3.
const numMRTests = 64

// GenerateParameters puts a random, valid set of DSA parameters into params.
// This function can take many seconds, even on fast machines.
func GenerateParameters(params *Parameters, rand io.Reader, sizes ParameterSizes) error {
	// This function doesn't follow FIPS 186-3 exactly in that it doesn't
	// use a verification seed to generate the primes. The verification
	// seed doesn't appear to be exported or used by other code and
	// omitting it makes the code cleaner.

	var L, N int
	switch sizes {
	case L1024N160:
		L = 1024
		N = 160
	case L2048N224:
		L = 2048
		N = 224
	case L2048N256:
		L = 2048
		N = 256
	case L3072N256:
		L = 3072
		N = 256
	default:
		return errors.New("crypto/dsa: invalid ParameterSizes")
	}

	qBytes := make([]byte, N/8)
	pBytes := make([]byte, L/8)

	q := new(big.Int)
	p := new(big.Int)
	rem := new(big.Int)
	one := new(big.Int)
	one.SetInt64(1)

GeneratePrimes:
	for {
		if _, err := io.ReadFull(rand, qBytes); err != nil {
			return err
		}

		qBytes[len(qBytes)-1] |= 1
		qBytes[0] |= 0x80
		q.SetBytes(qBytes)

		if !q.ProbablyPrime(numMRTests) {
			continue
		}

		for i := 0; i < 4*L; i++ {
			if _, err := io.ReadFull(rand, pBytes); err != nil {
				return err
			}

			pBytes[len(pBytes)-1] |= 1
			pBytes[0] |= 0x80

			p.SetBytes(pBytes)
			rem.Mod(p, q)
			rem.Sub(rem, one)
			p.Sub(p, rem)
			if p.BitLen() < L {
				continue
			}

			if !p.ProbablyPrime(numMRTests) {
				continue
			}

			params.P = p
			params.Q = q
			break GeneratePrimes
		}
	}

	h := new(big.Int)
	h.SetInt64(2)
	g := new(big.Int)

	pm1 := new(big.Int).Sub(p, one)
	e := new(big.Int).Div(pm1, q)

	for {
		g.Exp(h, e, p)
		if g.Cmp(one) == 0 {
			h.Add(h, one)
			continue
		}

		params.G = g
		return nil
	}
}

// GenerateKey generates a public&private key pair. The Parameters of the
// PrivateKey must already be valid (see GenerateParameters).
func GenerateKey(priv *PrivateKey, rand io.Reader) error {
	if priv.P == nil || priv.Q == nil || priv.G == nil {
		return errors.New("crypto/dsa: parameters not set up before generating key")
	}

	x := new(big.Int)
	xBytes := make([]byte, priv.Q.BitLen()/8)

	for {
		_, err := io.ReadFull(rand, xBytes)
		if err != nil {
			return err
		}
		x.SetBytes(xBytes)
		if x.Sign() != 0 && x.Cmp(priv.Q) < 0 {
			break
		}
	}

	priv.X = x
	priv.Y = new(big.Int)
	priv.Y.Exp(priv.G, x, priv.P)
	return nil
}

// fermatInverse calculates the inverse of k in GF(P) using Fermat's method.
// This has better constant-time properties than Euclid's method (implemented
// in math/big.Int.ModInverse) although math/big itself isn't strictly
// constant-time so it's not perfect.
func fermatInverse(k, P *big.Int) *big.Int {
	two := big.NewInt(2)
	pMinus2 := new(big.Int).Sub(P, two)
	return new(big.Int).Exp(k, pMinus2, P)
}

// Sign signs an arbitrary length hash (which should be the result of hashing a
// larger message) using the private key, priv. It returns the signature as a
// pair of integers. The security of the private key depends on the entropy of
// rand.
//
// Note that FIPS 186-3 section 4.6 specifies that the hash should be truncated
// to the byte-length of the subgroup. This function does not perform that
// truncation itself.
//
// Be aware that calling Sign with an attacker-controlled PrivateKey may
// require an arbitrary amount of CPU.
func Sign(rand io.Reader, priv *PrivateKey, hash []byte) (r, s *big.Int, err error) {
	// FIPS 186-3, section 4.6

	n := priv.Q.BitLen()
	if priv.Q.Sign() <= 0 || priv.P.Sign() <= 0 || priv.G.Sign() <= 0 || priv.X.Sign() <= 0 || n&7 != 0 {
		err = ErrInvalidPublicKey
		return
	}
	n >>= 3

	var attempts int
	for attempts = 10; attempts > 0; attempts-- {
		k := new(big.Int)
		buf := make([]byte, n)
		for {
			_, err = io.ReadFull(rand, buf)
			if err != nil {
				return
			}
			k.SetBytes(buf)
			// priv.Q must be >= 128 because the test above
			// requires it to be > 0 and that
			//    ceil(log_2(Q)) mod 8 = 0
			// Thus this loop will quickly terminate.
			if k.Sign() > 0 && k.Cmp(priv.Q) < 0 {
				break
			}
		}

		kInv := fermatInverse(k, priv.Q)

		r = new(big.Int).Exp(priv.G, k, priv.P)
		r.Mod(r, priv.Q)

		if r.Sign() == 0 {
			continue
		}

		z := k.SetBytes(hash)

		s = new(big.Int).Mul(priv.X, r)
		s.Add(s, z)
		s.Mod(s, priv.Q)
		s.Mul(s, kInv)
		s.Mod(s, priv.Q)

		if s.Sign() != 0 {
			break
		}
	}

	// Only degenerate private keys will require more than a handful of
	// attempts.
	if attempts == 0 {
		return nil, nil, ErrInvalidPublicKey
	}

	return
}

// Verify verifies the signature in r, s of hash using the public key, pub. It
// reports whether the signature is valid.
//
// Note that FIPS 186-3 section 4.6 specifies that the hash should be truncated
// to the byte-length of the subgroup. This function does not perform that
// truncation itself.
func Verify(pub *PublicKey, hash []byte, r, s *big.Int) bool {
	// FIPS 186-3, section 4.7

	if pub.P.Sign() == 0 {
		return false
	}

	if r.Sign() < 1 || r.Cmp(pub.Q) >= 0 {
		return false
	}
	if s.Sign() < 1 || s.Cmp(pub.Q) >= 0 {
		return false
	}

	w := new(big.Int).ModInverse(s, pub.Q)

	n := pub.Q.BitLen()
	if n&7 != 0 {
		return false
	}
	z := new(big.Int).SetBytes(hash)

	u1 := new(big.Int).Mul(z, w)
	u1.Mod(u1, pub.Q)
	u2 := w.Mul(r, w)
	u2.Mod(u2, pub.Q)
	v := u1.Exp(pub.G, u1, pub.P)
	u2.Exp(pub.Y, u2, pub.P)
	v.Mul(v, u2)
	v.Mod(v, pub.P)
	v.Mod(v, pub.Q)

	return v.Cmp(r) == 0
}
