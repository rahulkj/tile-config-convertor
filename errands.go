package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Errands struct {
	inputFile      string
	outputFile     string
	outputVarsFile string
}

func (e Errands) ProcessData() {
	if e.inputFile == "" || e.outputFile == "" || e.outputVarsFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	rawData := GetRaw(e.inputFile)

	var file *os.File
	var varFile *os.File

	if !FileExists(e.outputFile) {
		file = CreateFile(e.outputFile)
	}

	if !FileExists(e.outputVarsFile) {
		varFile = CreateFile(e.outputVarsFile)
	}

	defer file.Close()
	defer varFile.Close()

	// Generic interface to read the file into
	var f interface{}
	err := json.Unmarshal(rawData, &f)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}

	// Fetch the top level errands from the json
	m := f.(map[string]interface{})

	var v []interface{}

	// Fetch the errands attribute from the json
	if errandsMap, ok := m["errands"]; ok {
		// Fetch the errands from the errands map
		v = errandsMap.([]interface{})
	} else {
		fmt.Println("Cannot process the input file")
		os.Exit(1)
	}

	s := "errand-config:\n"
	WriteContents(file, s)

	for k := range v {
		node := v[k]

		nodeData := node.(map[string]interface{})
		var buf bytes.Buffer

		k := strings.ReplaceAll(fmt.Sprintf("%v", nodeData["name"]), "-", "_")

		s := fmt.Sprintf("  %s: \n", nodeData["name"])
		buf.WriteString(s)

		v := ""
		if nodeData["post_deploy"] == true || nodeData["post_deploy"] == false {
			s = fmt.Sprintf("    %s: ((%s))\n", "post-deploy-state", k+"_post_deploy_state")
			v = fmt.Sprintf("%s: %t\n", k+"_post_deploy_state", nodeData["post_deploy"])
			buf.WriteString(s)
			WriteContents(varFile, v)
		} else if nodeData["pre_delete"] == true || nodeData["pre_delete"] == false {
			v = fmt.Sprintf("%s: %t\n", k+"_pre_delete_state", nodeData["pre_delete"])
			s = fmt.Sprintf("    %s: ((%s))\n", "pre-delete-state", k+"_pre_delete_state")
			WriteContents(varFile, v)
			buf.WriteString(s)
		}

		WriteContents(file, buf.String())
	}
}
