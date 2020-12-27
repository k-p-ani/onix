/*
  Onix Config Manager - Artisan
  Copyright (c) 2018-2020 by www.gatblau.org
  Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
  Contributors to this project, hereby assign copyright in this code to the project,
  to be licensed under the same terms as the rest of the code.
*/
package cmd

func InitialiseRootCmd() *RootCmd {
	rootCmd := NewRootCmd()
	buildCmd := NewBuildCmd()
	artefactsCmd := NewListCmd()
	pushCmd := NewPushCmd()
	rmCmd := NewRmCmd()
	tagCmd := NewTagCmd()
	serveCmd := NewServeCmd()
	versionCmd := NewVersionCmd()
	runCmd := NewRunCmd()
	mergeCmd := NewMergeCmd()
	pullCmd := NewPullCmd()
	openCmd := NewOpenCmd()
	pgpCmd := InitialisePGPCommand()
	pipeCmd := InitialisePipeCommand()
	rootCmd.Cmd.AddCommand(
		buildCmd.cmd,
		artefactsCmd.cmd,
		pushCmd.cmd,
		rmCmd.cmd,
		tagCmd.cmd,
		serveCmd.cmd,
		versionCmd.cmd,
		runCmd.cmd,
		mergeCmd.cmd,
		pullCmd.cmd,
		openCmd.cmd,
		pgpCmd.cmd,
		pipeCmd.cmd,
	)
	return rootCmd
}

func InitialisePipeCommand() *PipeCmd {
	pipeCmd := NewPipeCmd()
	pipeSetupCmd := NewPipeAppCmd()
	pipeCmd.cmd.AddCommand(pipeSetupCmd.cmd)
	return pipeCmd
}

func InitialisePGPCommand() *PGPCmd {
	pgpCmd := NewPGPCmd()
	pgpGenCmd := NewPGPGenCmd()
	pgpImportCmd := NewPGPImportCmd()
	pgpExportCmd := NewPGPExportCmd()
	pgpEncryptCmd := NewPGPEncryptCmd()
	pgpDecryptCmd := NewPGPDecryptCmd()
	pgpCmd.cmd.AddCommand(pgpGenCmd.cmd, pgpImportCmd.cmd, pgpExportCmd.cmd, pgpEncryptCmd.cmd, pgpDecryptCmd.cmd)
	return pgpCmd
}