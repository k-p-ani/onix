/*
  Onix Config Manager - Pilot
  Copyright (c) 2018-2021 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/

package core

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type AK struct {
	Data      string `json:"ak"`
	Signature string `json:"s"`
}

func (ak *AK) String() string {
	b, _ := json.Marshal(ak)
	return string(b[:])
}

type AKInfo struct {
	HostUUID   string    `json:"host_uuid"`
	MacAddress string    `json:"mac_address"`
	CtlURI     string    `json:"ctl_uri"`
	Expiry     time.Time `json:"expiry"`
	VerifyKey  string    `json:"verify_key"`
}

func AkExist() bool {
	_, err := os.Stat(AkFile())
	return err == nil
}

func TkExist() bool {
	_, err := os.Stat(TkFile())
	return err == nil
}

func readAKey(ak AK) (*AKInfo, error) {
	if valid, err := verify(ak.Data, ak.Signature); !valid {
		return nil, fmt.Errorf("signature verification failed: %s\n", err)
	}
	data, err := decrypt(sk, ak.Data, iv)
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt activation key: %s\n", err)
	}
	akinfo := new(AKInfo)
	err = json.Unmarshal([]byte(data), akinfo)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal activation key: %s\n", err)
	}
	return akinfo, nil
}

func loadAKey(path string) (*AK, error) {
	path = Abs(path)
	keyBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read activation key file: %s\n", err)
	}
	d, err := hex.DecodeString(string(keyBytes[:]))
	if err != nil {
		return nil, fmt.Errorf("cannot decode activation key: %s\n", err)
	}
	ak := new(AK)
	err = json.Unmarshal(d, ak)
	if err != nil {
		return nil, fmt.Errorf("cannot unmarshal activation key: %s\n", err)
	}
	return ak, err
}
