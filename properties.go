package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
)

type Properties struct {
	inputFile      string
	outputFile     string
	outputVarsFile string
}

func (p Properties) ProcessData() {

	if p.inputFile == "" || p.outputFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	rawData := GetRaw(p.inputFile)

	var file *os.File
	var varFile *os.File

	if !FileExists(p.outputFile) {
		file = CreateFile(p.outputFile)
	}

	if !FileExists(p.outputVarsFile) {
		varFile = CreateFile(p.outputVarsFile)
	}

	defer file.Close()
	defer varFile.Close()

	// Generic interface to read the file into
	var f interface{}
	err1 := json.Unmarshal(rawData, &f)
	if err1 != nil {
		fmt.Println("Error parsing JSON: ", err1)
	}

	// Fetch the top level properties from the json
	m := f.(map[string]interface{})

	var v map[string]interface{}

	// Fetch the properties attribute from the json
	if propertiesMap, ok := m["properties"]; ok {
		// Fetch the properties from the properties map
		v = propertiesMap.(map[string]interface{})
	} else {
		fmt.Println("Cannot process the input file")
		os.Exit(1)
	}

	var keys []string
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	s := "product-properties:\n"
	WriteContents(file, s)

	// To perform the opertion you want
	for _, k := range keys {
		node := v[k]

		nodeData := node.(map[string]interface{})

		if nodeData["configurable"] == true {
			s := fmt.Sprintf("  %s:", k)
			length := len(k)
			totalPadding := 100
			if nodeData["optional"] == false {
				s = fmt.Sprintf("%s\n", s)
			} else {
				s = fmt.Sprintf("%s%s%s\n", s, getPaddedString(totalPadding-length), "# OPTIONAL")
			}

			WriteContents(file, s)

			var kv = strings.ReplaceAll(strings.ReplaceAll(strings.Replace(k, ".properties.", "", 1), ".", "_"), "-", "_")

			if nodeData["type"] == "rsa_cert_credentials" {
				var buf bytes.Buffer
				buf = handleCert(4, "value: \n", buf)
				WriteContents(file, buf.String())
			} else if nodeData["type"] == "secret" {
				var buf bytes.Buffer
				buf.WriteString("    value: \n")
				buf.WriteString("      secret: \n")
				WriteContents(file, buf.String())
			} else if nodeData["type"] == "simple_credentials" {
				var buf bytes.Buffer
				buf.WriteString("    value: \n")
				buf.WriteString("      identity: \n")
				buf.WriteString("      password: \n")
				WriteContents(file, buf.String())
			} else if nodeData["type"] == "multi_select_options" {
				var buf bytes.Buffer
				buf = handleMultiSelectOptions(nodeData)
				WriteContents(file, buf.String())
			} else if nodeData["type"] == "collection" {
				var buf bytes.Buffer
				buf = handleCollections(nodeData)
				WriteContents(file, buf.String())
			} else if nodeData["type"] == "integer" {
				var s string
				var v string
				value := nodeData["value"]
				switch value.(type) {
				case float64:
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: %v\n", kv, int(value.(float64)))
				case float32:
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: %v\n", kv, int(value.(float32)))
				case int64:
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: %v\n", kv, value.(int64))
				case int32:
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: %v\n", kv, value.(int32))
				case int:
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: %v\n", kv, value.(int32))
				default:
					s = fmt.Sprintf("%svalue: \n", getPaddedString(4))
					v = fmt.Sprintf("%s: \n", kv)
				}
				WriteContents(file, s)
				WriteContents(varFile, v)
			} else {
				var s string
				var v string
				value := nodeData["value"]
				if value != nil {
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: %v\n", kv, value)
				} else {
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: \n", kv)
				}
				WriteContents(file, s)
				WriteContents(varFile, v)
			}
		}
	}
}

func handleCert(padding int, firstLine string, buf bytes.Buffer) bytes.Buffer {
	s := getPaddedString(padding) + firstLine
	buf.WriteString(s)

	paddedString := getPaddedString(padding + 2)
	s = paddedString + "private_key_pem: \n"
	buf.WriteString(s)

	s = paddedString + "cert_pem: \n"
	buf.WriteString(s)

	return buf
}

func handleMultiSelectOptions(nodeData map[string]interface{}) bytes.Buffer {
	var buf bytes.Buffer
	buf.WriteString("    value: \n")
	value := nodeData["value"]
	valueType := reflect.TypeOf(value)
	if valueType != nil {
		switch valueType.Kind() {
		case reflect.Slice:
			value := nodeData["value"].([]interface{})
			for _, item := range value {
				s := fmt.Sprintf("%s- %s\n", getPaddedString(4), item)
				buf.WriteString(s)
			}
		case reflect.String:
			s := fmt.Sprintf("%s- %s\n", getPaddedString(4), value)
			buf.WriteString(s)
		}
	}
	return buf
}

func handleCollections(nodeData map[string]interface{}) bytes.Buffer {
	var buf bytes.Buffer
	value := nodeData["value"].([]interface{})

	buf.WriteString("    value: \n")

	for _, item := range value {
		arrayAdded := false
		for innerKey, innerVal := range item.(map[string]interface{}) {
			typeAssertedInnerValue := innerVal.(map[string]interface{})
			innerValueType := typeAssertedInnerValue["type"]
			var s string
			if !arrayAdded {
				if innerValueType == "rsa_cert_credentials" {
					s = fmt.Sprintf("- %s:\n", innerKey)
					buf = handleCert(4, s, buf)
				} else if innerValueType == "secret" {
					s = fmt.Sprintf("%s- %s:\n", getPaddedString(4), innerKey)
					buf.WriteString(s)
					buf.WriteString("        secret: \n")
				} else {
					s = fmt.Sprintf("%s- %s: %v \n", getPaddedString(4), innerKey, typeAssertedInnerValue["value"])
					buf.WriteString(s)
				}
				arrayAdded = true
			} else {
				if innerValueType == "rsa_cert_credentials" {
					s = fmt.Sprintf("%s:\n", innerKey)
					buf = handleCert(6, s, buf)
				} else if innerValueType == "secret" {
					s = fmt.Sprintf("%s%s:\n", getPaddedString(6), innerKey)
					buf.WriteString(s)
					buf.WriteString("        secret: \n")
				} else {
					s = fmt.Sprintf("%s%s: %v \n", getPaddedString(6), innerKey, typeAssertedInnerValue["value"])
					buf.WriteString(s)
				}
			}
		}
		arrayAdded = false
	}
	return buf
}

func getPaddedString(count int) string {
	var s string
	for i := 0; i < count; i++ {
		s += " "
	}
	return s
}
