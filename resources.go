package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

type Resources struct {
	inputFile      string
	outputFile     string
	outputVarsFile string
}

func (r Resources) ProcessData() {

	if r.inputFile == "" || r.outputFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	rawData := GetRaw(r.inputFile)

	var file *os.File
	var varFile *os.File

	if !FileExists(r.outputFile) {
		file = CreateFile(r.outputFile)
	}

	if !FileExists(r.outputVarsFile) {
		varFile = CreateFile(r.outputVarsFile)
	}

	defer file.Close()
	defer varFile.Close()

	// Generic interface to read the file into
	var f interface{}
	err := json.Unmarshal(rawData, &f)
	if err != nil {
		fmt.Println("Error parsing JSON: ", err)
	}

	// Fetch the top level properties from the json
	m := f.(map[string]interface{})

	var resources []interface{}

	// Fetch the resources attribute from the json
	if resourcesMap, ok := m["resources"]; ok {
		// Fetch the properties from the properties map
		resources = resourcesMap.([]interface{})
	} else {
		fmt.Println("Cannot process the input file")
		os.Exit(1)
	}

	s := "resource-config:\n"
	WriteContents(file, s)

	for _, item := range resources {
		value := item.(map[string]interface{})
		if int(value["instances_best_fit"].(float64)) != 0 {
			k := strings.ReplaceAll(fmt.Sprintf("%v", value["identifier"]), "-", "_")
			s := fmt.Sprintf("  %v:\n", value["identifier"])
			WriteContents(file, s)

			s = fmt.Sprintf("    instances: ((%v))\n", fmt.Sprintf("%v", k)+"_instances")
			v := fmt.Sprintf("%s: %v\n", fmt.Sprintf("%v", k)+"_instances", value["instances_best_fit"])

			WriteContents(file, s)
			WriteContents(varFile, v)

			s = fmt.Sprintf("    instance_type:\n      id: ((%v))\n", fmt.Sprintf("%v", k)+"_instance_type")
			v = fmt.Sprintf("%s: %v\n", fmt.Sprintf("%v", k)+"_instance_type", value["instance_type_best_fit"])
			WriteContents(file, s)
			WriteContents(varFile, v)

			if _, ok := value["persistent_disk_mb"]; ok {
				s := fmt.Sprintf("    persistent_disk:\n      size_mb: ((%v))\n", fmt.Sprintf("%v", k)+"_persistent_disk_size_mb")
				v := fmt.Sprintf("%s: \"%v\"\n", fmt.Sprintf("%v", k)+"_persistent_disk_size_mb", value["persistent_disk_best_fit"])
				WriteContents(file, s)
				WriteContents(varFile, v)
			}
		}
	}
}
