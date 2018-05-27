package main

import (
	"fmt"
	"os"
	"encoding/json"

	"github.com/qiniu/logkit/metric"
	"github.com/qiniu/logkit/plugins/metric/pkg/oracle"
	_ "github.com/mattn/go-oci8"
)

type Command func([]string)

var usageTemplate = `orakit is a tool for collecting oracle metric.

Usage:

	orakit command [arguments]

The commands are:
	name                show orakit name
	config              print config
	collect [arguments] start collect oracle metic
	version             print orakit version
	usages              print usages

`
var ora metric.Collector
var Commands = map[string]func([]string){}
var version = "1.0"

func init() {
	Commands["collect"] = Collect
	Commands["config"] = Config
	Commands["usages"] = Usages
	Commands["name"] = Name
	Commands["version"] = Version
	Commands["tags"] = Tags

}
func main() {
	ora = oracle.New()
	args := os.Args[1:]
	if len(args) > 0 {
		commandToRun, found := commandMatching(args[0])
		if found {
			commandToRun(args[1:])
			return
		}
	}
	fmt.Println(usageTemplate)
}

func commandMatching(name string) (Command, bool) {
	if c, exists := Commands[name]; exists {
		return c, true
	}
	return nil, false
}

func Name(args []string) {
	fmt.Print(ora.Name())
	return

}
func Collect(args []string) {
	datas, err := ora.Collect()
	if err != nil {
		fmt.Printf("< collect failed %s >", err.Error())
	}
	JsonPrint(datas)
	return
}
func Usages(args []string) {
	fmt.Print(ora.Usages())
	return
}
func Config(args []string) {
	JsonPrint(ora.Config())
	return
}
func Version(args []string) {
	fmt.Print(version)
	return
}
func Tags(args []string){
	fmt.Print(ora.Tags())
	return
}
func JsonPrint(obj interface{}) error {
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = os.Stdout.Write(data)
	return err
}
