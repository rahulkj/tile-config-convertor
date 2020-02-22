package main

import (
	"bytes"
	"flag"
	"os"
)

type NetworksAndAZs struct {
	outputFile     string
	outputVarsFile string
}

func (nz NetworksAndAZs) ProcessData() {

	if nz.outputFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var file *os.File
	var varFile *os.File

	if !FileExists(nz.outputFile) {
		file = CreateFile(nz.outputFile)
	}

	if !FileExists(nz.outputVarsFile) {
		varFile = CreateFile(nz.outputVarsFile)
	}

	defer file.Close()
	defer varFile.Close()

	var buf bytes.Buffer
	buf.WriteString("network-properties:\n")
	buf.WriteString("  network: ((network_name))\n")
	buf.WriteString("    name:\n")
	buf.WriteString("  service-network:\n")
	buf.WriteString("    name: ((service_network_name))\n")
	buf.WriteString("  other_availability_zones:\n")
	buf.WriteString("  - name: ((az_1_name))\n")
	buf.WriteString("  - name: ((az_2_name))\n")
	buf.WriteString("  singleton_availability_zone:\n")
	buf.WriteString("    name: ((singleton_availability_zone_name))\n")

	WriteContents(file, buf.String())

	var varbuf bytes.Buffer
	varbuf.WriteString("network_name: \n")
	varbuf.WriteString("service_network_name: \n")
	varbuf.WriteString("az_1_name: \n")
	varbuf.WriteString("az_2_name: \n")
	varbuf.WriteString("singleton_availability_zone_name: \n")
	WriteContents(varFile, varbuf.String())
}
