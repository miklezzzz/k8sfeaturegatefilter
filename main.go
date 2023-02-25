package main

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"bytes"
	"reflect"
	"strings"
	"flag"
	"regexp"
)

const (
	regExpPattern string = "%s.feature gate"
)

var (
	FeatureGatesPtr *string
	PathToManifestPtr *string
)

func parse(data interface{}, re *regexp.Regexp) bool {
	switch reflect.ValueOf(data).Kind() {
	case reflect.Map:
		m, ok := data.(map[interface{}]interface{})
		if !ok {
			fmt.Printf("want type map[string]interface{};  got %T", data)
		}
		for k := range m {
			if k == "description" {
				if re.MatchString(reflect.ValueOf(m[k]).String()) {
					return true
				}
			}
			if res := parse(m[k], re); res {
				delete(m, k)
			}
			data = m
		}
	case reflect.Slice,reflect.Array:
		l, ok := data.([]interface{})
		if !ok {
			fmt.Printf("want type []interface{};  got %T", data)
		}
		for i := range l {
			if res := parse(l[i], re); res {
				l[i] = l[len(l)-1]
				data = l[:len(l)-1]
			}
		}
	case reflect.String:
	}
	return false
}

func init() {
        FeatureGatesPtr = flag.String("gates", "", "Kubernetes Feature Gates to exclude, delimited by a comma")
        PathToManifestPtr = flag.String("file", "./crd.yaml", "Path to the manifest")
        flag.Parse()
}

func main() {
	var data map[interface{}]interface{}

	buffer := bytes.NewBuffer(nil)
	file, err := os.Open(*PathToManifestPtr)
	if err != nil {
		panic(err.Error())
	}
	defer file.Close()

	io.Copy(buffer, file)
	err = yaml.Unmarshal(buffer.Bytes(), &data)

	if err != nil {
		fmt.Printf(err.Error())
	}

	if len(*FeatureGatesPtr) != 0 {
		gates := strings.Split(*FeatureGatesPtr, ",")
		gatesRegEx := gates[0]

		for i:=1;i<len(gates);i++ {
			gatesRegEx = fmt.Sprintf("%s|%s", gatesRegEx, gates[i])
		}

		gatesRegEx = fmt.Sprintf("(%s)", gatesRegEx)
		filter := regexp.MustCompile(fmt.Sprintf("%s", fmt.Sprintf(regExpPattern, gatesRegEx)))
		parse(data, filter)
	}

	d, err := yaml.Marshal(&data)

	if err != nil {
		fmt.Printf(err.Error())
	}

	fmt.Printf("---\n%s\n\n", string(d))
}
