/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-2020 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
package sign

import (
	"testing"
)

func TestGenerateKeys(t *testing.T) {
	// creates a new PGP object
	p := NewPGP("gatblau/boot", "an artisan pgp key for digital signatures", "onix@gatblau.org", 2048)

	// saves private and public keys
	err := p.SaveKeyPair(".", "pub.key", "priv.key")
	if err != nil {
		t.FailNow()
	}
}

func TestLoadKeySignAndVerify(t *testing.T) {
	// load the private key for signing
	priv, err := LoadPGP("priv.key")
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	// sign the message
	signature, err := priv.Sign([]byte("Hello World"))
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	// load the public key for verification of the signature
	pub, err := LoadPGP("pub.key")
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
	// verify the signature
	err = pub.Verify([]byte("Hello World"), signature)
	if err != nil {
		t.Fatal(err)
		t.FailNow()
	}
}
