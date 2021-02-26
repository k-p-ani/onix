/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-2021 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
package cmd

import (
	"github.com/gatblau/onix/artisan/crypto"
	"github.com/spf13/cobra"
)

// list local packages
type PGPGenCmd struct {
	cmd     *cobra.Command
	bitSize *int   // the key bit size
	path    string // the file path to the keys
	prefix  string // the filename of the keys
	name    string // the key identity filename
	comment string // the key identity comment
	email   string // the key identity email
}

func NewPGPGenCmd() *PGPGenCmd {
	c := &PGPGenCmd{
		cmd: &cobra.Command{
			Use:   "gen",
			Short: "generate a new PGP/RSA key pair",
			Long:  `PGP/RSA keys are used to sign and verify signatures of packages or encrypt/decrypt files`,
		},
	}
	c.bitSize = c.cmd.Flags().IntP("size", "s", 2048, "The bit size of the generated key pair, defaults to s=2048 \nOther common sizes are 1024, 3072 and 4096. \nAny size is possible.")
	c.cmd.Flags().StringVarP(&c.path, "path", "p", ".", "The path of the generated key pair, defaults to the current directory \".\"")
	c.cmd.Flags().StringVarP(&c.prefix, "prefix", "x", "id", "The prefix given to the generated key pair, defaults to 'id' which produces: id_rsa_key.pem (private key) and id_rsa_pub.pem (public key).\nIf specified, the naming is [name]_rsa_key.pem (private key) and [name]_rsa_pub.pem (public key)")
	c.cmd.Flags().StringVarP(&c.name, "name", "n", "", "The identity name of the generated key pair, if not specified the hostname where the generation takes place is used.")
	c.cmd.Flags().StringVarP(&c.comment, "comment", "c", "", "The identity comment in the generated key pair, if not specified a default comment is used.")
	c.cmd.Flags().StringVarP(&c.email, "email", "e", "", "The identity email given to the generated key pair, if not specified the an autogenerated email is used.")
	c.cmd.Run = c.Run
	return c
}

func (b *PGPGenCmd) Run(cmd *cobra.Command, args []string) {
	crypto.GeneratePGPKeys(b.path, b.prefix, b.name, b.comment, b.email, *b.bitSize)
}
