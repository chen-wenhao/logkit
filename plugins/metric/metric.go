package metric

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"bytes"
	"github.com/qiniu/log"
	"github.com/qiniu/logkit/metric"
)

const (
	DefaultConfDir     = "plugins/metric/conf"
	DefaultBinDir      = "plugins/metric/bin"

)

var DefaultConfSuffixs = [2]string{".json", ".conf"}

type Plugin struct {
	PluginName string                 `json:"name"`
	PluginType string                 `json:"type"` //cmd | rpc
	Version    string                 `json:"version"`
	ConfigData map[string]interface{} `json:"data"` // pass to external
	Spec       `json:"spec"`
}
type Spec struct {
	Addr string `json:"addr"`
	Cmd  string `json:"cmd"`
	Env string `json:"env"`
}

type CMDPlugin struct {
	Plugin
}
// TODO: to implement rpc plugins
func (p *CMDPlugin) Name() string {
	out , err := p.execCmd("name")
	if err != nil {
		log.Errorf("Exec name cmd failed, err: %s", err.Error())
		return ""
	}
	return out.String()
}
func (p *CMDPlugin) Usages() string {
	out , err := p.execCmd("usages")
	if err != nil {
		log.Errorf("Exec usages cmd failed, err: %s", err.Error())
		return ""
	}
	return out.String()
}
func (p *CMDPlugin) Tags() []string {
	out , err := p.execCmd("tags")
	if err != nil {
		log.Errorf("Exec tags cmd failed, err: %s", err.Error())
		return nil
	}
	return strings.Split(strings.TrimRight(strings.TrimLeft(out.String(),"["),"]"),",")
}
func (p *CMDPlugin) Config() (config map[string]interface{}) {
	out , err := p.execCmd("config")
	if err != nil {
		log.Errorf("Exec config cmd failed, err: %s", err.Error())
		return nil
	}
	err = json.Unmarshal(out.Bytes(), &config)
	if err != nil {
		log.Errorf("Exec config cmd failed, err: %s", err.Error())
		return nil
	}
	return config
}
func (p *CMDPlugin) Collect() (datas []map[string]interface{}, err error) {
	out , err := p.execCmd("collect")
	if err != nil {
		return nil, fmt.Errorf("Exec collect failed, err: %s", err.Error())
	}
	err = json.Unmarshal(out.Bytes(), &datas)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal output data failed, error: %v",err.Error())
	}
	return
}
func (p *CMDPlugin) SyncConfig(config map[string]interface{}) error {
	log.Debugf("SyncConfig with data %v",config)
	p.ConfigData = config
	return nil
}

func (p *CMDPlugin)execCmd(arg ...string)(*bytes.Buffer,error){
	cmd := exec.Command(p.Cmd, arg...)
	cmd.Env = strings.Split(p.Env,";")
	// stdin prepare
	if p.ConfigData !=nil {
		stdindata, err := json.Marshal(p.ConfigData)
		if err != nil {
			return nil, fmt.Errorf("Marshal stdin data failed, error: %v",err.Error())
		}
		cmd.Stdin = bytes.NewReader(stdindata)
	}
	// stdout
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("run cmd failed, err: %v", err.Error())
	}
	log.Debugf("Collect original data with %s",out.String())
	return &out,nil
}

func ProbeMetricPlugins(confDir string) error {
	if confDir == "" {
		confDir = DefaultConfDir
	}
	files, err := GetPluginConf(confDir)
	if len(files) == 0 || err != nil {
		return fmt.Errorf("no external plugins found in %s", err)
	}

	sort.Strings(files)
	for _, confFile := range files {
		p, err := PluginFromFile(confFile)
		if err != nil {
			log.Errorf("read config file failed, error: %s", err.Error())
			continue
		}
		log.Debugf("probe a metric plugin,[%s]",p.PluginName)
		p.Cmd = confDir +"/../bin/" + p.Cmd
		if _, err := os.Stat(p.Cmd);err !=nil {
			log.Errorf("plugin [%s] is invalid, error %s",p.PluginName,err.Error())
			continue
		}
		registry(p.PluginName,p)
	}

	return nil
}

func ProbeMetricPluginsOnce() error{
	rootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("get executable file path error %v", err)
	}
	confDir := filepath.Join(rootDir, DefaultConfDir)
	if err := ProbeMetricPlugins(confDir); err != nil {
		log.Errorf("probe plugins failed, error: %v", err.Error())
		return err
	}
	return nil
	
}
// for dynamically load plugins
func LoopProbeMetricPlugins(exitchan chan struct{}) error {
	rootDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return fmt.Errorf("get executable file path error %v", err)
	}
	confDir := filepath.Join(rootDir, DefaultConfDir)
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-exitchan:
			return nil
		case <-ticker.C:
			if err := ProbeMetricPlugins(confDir); err != nil {
				log.Errorf("probe plugins failed, error: %v", err.Error())
			}

		}
	}
	return nil
}
func registry(name string, p metric.Collector) {
	metric.Add(name, func() metric.Collector {
		return p
	})
}

func isConfFile(file string) bool {
	suffixs := DefaultConfSuffixs
	for _, str := range suffixs {
		if ok := strings.HasSuffix(strings.ToLower(file), str); !ok {
			continue
		}
		return true
	}
	return false
}

func GetPluginConf(dirPth string) (files []string, err error) {
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)

	for _, fi := range dir {
		if isConfFile(fi.Name()) {
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}

	return files, nil
}
//  当前只支持cmd 类型 plugin
func PluginFromFile(filename string) (*CMDPlugin, error) {
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("error reading %s: %s", filename, err)
	}
	var p CMDPlugin
	if err := json.Unmarshal(bytes, &p); err != nil {
		return nil, fmt.Errorf("error parsing configuration: %s", err)
	}
	return &p, nil
}
