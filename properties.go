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
			if strings.HasPrefix(kv, "_") {
				kv = strings.Replace(kv, "_", "", 1)
			}

			if nodeData["type"] == "rsa_cert_credentials" {
				var configBuf bytes.Buffer
				var varsBuf bytes.Buffer
				configBuf, varsBuf = handleCert(4, kv, "value: \n", configBuf, varsBuf)
				WriteContents(file, configBuf.String())
				WriteContents(varFile, varsBuf.String())
			} else if nodeData["type"] == "secret" {
				var buf bytes.Buffer
				buf.WriteString("    value: \n")
				buf.WriteString(fmt.Sprintf("      secret: ((%v))\n", kv+"_secret"))
				WriteContents(file, buf.String())
				WriteContents(varFile, fmt.Sprintf("%v: \n", kv+"_secret"))
			} else if nodeData["type"] == "simple_credentials" {
				var buf bytes.Buffer
				buf.WriteString("    value: \n")
				buf.WriteString(fmt.Sprintf("      identity: ((%v))\n", kv+"_identity"))
				buf.WriteString(fmt.Sprintf("      password: ((%v))\n", kv+"_password"))
				WriteContents(file, buf.String())
				WriteContents(varFile, fmt.Sprintf("%v: \n", kv+"_identity"))
				WriteContents(varFile, fmt.Sprintf("%v: \n", kv+"_password"))
			} else if nodeData["type"] == "multi_select_options" {
				var configBuf bytes.Buffer
				var varBuf bytes.Buffer
				configBuf, varBuf = handleMultiSelectOptions(kv, nodeData)
				WriteContents(file, configBuf.String())
				WriteContents(varFile, varBuf.String())
			} else if nodeData["type"] == "collection" {
				var configBuf bytes.Buffer
				var varsBuf bytes.Buffer
				configBuf, varsBuf = handleCollections(kv, nodeData)
				WriteContents(file, configBuf.String())
				WriteContents(varFile, varsBuf.String())
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
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: \n", kv)
				}
				WriteContents(file, s)
				WriteContents(varFile, v)
			} else if nodeData["type"] == "boolean" {
				value := nodeData["value"]
				s := fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
				v := fmt.Sprintf("%s: %v\n", kv, value)
				WriteContents(file, s)
				WriteContents(varFile, v)
			} else {
				var s string
				var v string
				value := nodeData["value"]
				if value != nil {
					s = fmt.Sprintf("%svalue: ((%v))\n", getPaddedString(4), kv)
					v = fmt.Sprintf("%s: \"%v\"\n", kv, value)
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

func handleCert(padding int, prefix string, firstLine string, configBuf bytes.Buffer, varsBuf bytes.Buffer) (bytes.Buffer, bytes.Buffer) {
	s := getPaddedString(padding) + firstLine
	configBuf.WriteString(s)

	paddedString := getPaddedString(padding + 2)
	s = paddedString + "private_key_pem: ((" + prefix + "_private_key_pem)) \n"
	configBuf.WriteString(s)

	s = paddedString + "cert_pem: ((" + prefix + "_cert_pem)) \n"
	configBuf.WriteString(s)

	varsBuf.WriteString(fmt.Sprintf("%s: \n", prefix+"_private_key_pem"))
	varsBuf.WriteString(fmt.Sprintf("%s: \n", prefix+"_cert_pem"))

	return configBuf, varsBuf
}

func handleMultiSelectOptions(kv string, nodeData map[string]interface{}) (bytes.Buffer, bytes.Buffer) {
	var configBuf bytes.Buffer
	var varBuf bytes.Buffer

	configBuf.WriteString("    value: \n")

	value := nodeData["value"]
	valueType := reflect.TypeOf(value)
	if valueType != nil {
		switch valueType.Kind() {
		case reflect.Slice:
			value := nodeData["value"].([]interface{})
			for _, item := range value {
				s := fmt.Sprintf("%s- %s\n", getPaddedString(4), item)
				configBuf.WriteString(s)
			}
		case reflect.String:
			s := fmt.Sprintf("%s- ((%s))\n", getPaddedString(4), kv)
			configBuf.WriteString(s)
			varBuf.WriteString(fmt.Sprintf("%s: %s\n", kv, value))
		}
	} else {
		s := fmt.Sprintf("%s- ((%s))\n", getPaddedString(4), kv)
		configBuf.WriteString(s)
		varBuf.WriteString(fmt.Sprintf("%s: \n", kv))
	}
	return configBuf, varBuf
}

func handleCollections(kv string, nodeData map[string]interface{}) (bytes.Buffer, bytes.Buffer) {
	var configBuf bytes.Buffer
	var varsBuf bytes.Buffer
	value := nodeData["value"].([]interface{})

	configBuf.WriteString("    value: \n")

	for _, item := range value {
		arrayAdded := false
		for innerKey, innerVal := range item.(map[string]interface{}) {
			typeAssertedInnerValue := innerVal.(map[string]interface{})
			innerValueType := typeAssertedInnerValue["type"]
			var innerkv = strings.ReplaceAll(strings.ReplaceAll(strings.Replace(strings.Replace(innerKey, ".properties.", "", 1), ".", "", 1), ".", "_"), "-", "_")
			var s string
			if !arrayAdded {
				if innerValueType == "rsa_cert_credentials" {
					s = fmt.Sprintf("- %s:\n", innerKey)
					configBuf, varsBuf = handleCert(4, kv+innerkv, s, configBuf, varsBuf)
				} else if innerValueType == "secret" {
					configBuf.WriteString(fmt.Sprintf("%s- %s:\n", getPaddedString(4), innerKey))
					configBuf.WriteString(fmt.Sprintf("        secret: ((%v))\n", kv+innerkv+"_secret"))
					varsBuf.WriteString(fmt.Sprintf("%s: \n", kv+innerkv+"_secret"))
				} else {
					configBuf.WriteString(fmt.Sprintf("%s- %s: ((%v)) \n", getPaddedString(4), innerKey, kv+innerkv))
					varsBuf.WriteString(fmt.Sprintf("%s: %v\n", kv+innerkv, typeAssertedInnerValue["value"]))
				}
				arrayAdded = true
			} else {
				if innerValueType == "rsa_cert_credentials" {
					s = fmt.Sprintf("%s:\n", innerKey)
					configBuf, varsBuf = handleCert(6, innerKey, s, configBuf, varsBuf)
				} else if innerValueType == "secret" {
					configBuf.WriteString(fmt.Sprintf("%s%s:\n", getPaddedString(6), innerKey))
					configBuf.WriteString(fmt.Sprintf("        secret: ((%v))\n", kv+innerkv+"_secret"))
					varsBuf.WriteString(fmt.Sprintf("%s: \n", kv+innerkv+"_secret"))
				} else {
					configBuf.WriteString(fmt.Sprintf("%s%s: ((%v)) \n", getPaddedString(6), innerKey, kv+innerkv+"_secret"))
					varsBuf.WriteString(fmt.Sprintf("%s: %v\n", kv+innerkv, typeAssertedInnerValue["value"]))
				}
			}
		}
		arrayAdded = false
	}
	return configBuf, varsBuf
}

func getPaddedString(count int) string {
	var s string
	for i := 0; i < count; i++ {
		s += " "
	}
	return s
}
