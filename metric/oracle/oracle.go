package oracle

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/qiniu/log"
	"github.com/qiniu/logkit/metric"
	. "github.com/qiniu/logkit/utils/models"
	"encoding/json"
)

const (
	TypeMetricSystem  = "oracle"
	MetricSystemUsage = "oracle数据库"

	// Config 中的字段
	ConfigDSN    = "dsn"
	ConfigLibDir = "lib_dir"

	ConfigSession    = "session"
	ConfigTablespace = "tablespace"
	ConfigHitRatio   = "hitratio"
	ConfigEventWait  = "eventwait"
	ConfigPhysicalIO = "physicalIO"
	ConfigLogicalIO  = "logicalIO"

	DefaultLibDir = "oci"
)

// KeySystemUsages TypeMetricSystem的字段名称
var KeyOracleUsages = []KeyValue{}

// ConfigDiskUsages TypeMetricDisk config 中的字段描述
var ConfigOracleUsages = []KeyValue{
	{ConfigDSN, "oracle数据库的连接字符串（DSN）"},
	{ConfigLibDir, "oracle数据库连接动态库地址（LD_LIBRARY_PATH）"},
	{ConfigSession, "收集session相关的信息（session_related）"},
	{ConfigTablespace, "收集tablespace相关的信息（tablespace_related）"},
	{ConfigHitRatio, "收集hit ratio相关的信息（hitratio_related）"},
	{ConfigEventWait, "收集event wait相关的信息（eventwait_related）"},
	{ConfigPhysicalIO, "收集physical IO相关的信息（physicalIO_related）"},
	{ConfigLogicalIO, "收集logical IO相关的信息（logicalIO_related）"},
}

type Oracle struct {
	db     *sql.DB
	prefix string
	DSN    string `json:"dsn"`     //user:password@host:port/sid?param1=value1&param2=value2
	LibDir string `json:"lib_dir"` //LD_LIBRARY_PATH
	//UseBundledOCI    bool `json:"use_bundled_oci"`
	EnableSessionStats    bool `json:"session"`
	EnableTablespaceStats bool `json:"tablespace"`
	EnableHitRatioStats   bool `json:"hitratio"`
	EnableEventWaitStats  bool `json:"eventwait"`
	EnablePhysicalIOStats bool `json:"physicalIO"`
	EnableLogicalIOStats  bool `json:"logicalIO"`
}

func (o *Oracle) Collect() (datas []map[string]interface{}, err error) {
	defer HandleCrash()
	err = o.initDB()
	if err != nil {
		return nil, err
	}
	defer o.Close()

	stats := make([]map[string]interface{}, 0)
	// --seesion stat
	if o.EnableSessionStats {
		stats = append(stats, o.fetchSessionStats())
	}
	if o.EnableHitRatioStats {
		stats = append(stats, o.fetchHitratioStats())
	}
	if o.EnableEventWaitStats {
		stats = append(stats, o.fetchEventwaitStats())
	}
	if o.EnablePhysicalIOStats {
		stats = append(stats, o.fetchPhysicalIOStats())
	}
	if o.EnableLogicalIOStats {
		stats = append(stats, o.fetchLogicalIOStats())
	}
	datas = append(datas, stats...)

	// tablespace stats
	if o.EnableTablespaceStats {
		if tsStats, err := o.getTablespaceStats(); err == nil {
			datas = append(datas, tsStats...)
		} else {
			log.Errorf("get oracle tablespace stats failed, error: %v", err.Error())
		}
	}
		fmt.Printf("===============\n %+v \n======",datas)

	js,_ := json.Marshal(datas)
	f, _ := os.OpenFile("/tmp/test.json", os.O_CREATE|os.O_WRONLY, DefaultFilePerm)
	f.Write(js)
	f.Sync()
	f.Close()
	return datas, nil
}

func (_ *Oracle) Name() string {
	return TypeMetricSystem
}

func (_ *Oracle) Usages() string {
	return MetricSystemUsage
}

func (_ *Oracle) Tags() []string {
	return []string{}
}

func (o *Oracle) Config() map[string]interface{} {
	configOptions := make([]Option, 0)
	for _, val := range ConfigOracleUsages {
		option := Option{
			KeyName:       val.Key,
			ChooseOnly:    true,
			ChooseOptions: []interface{}{"true", "false"},
			Default:       true,
			DefaultNoUse:  true,
			Description:   val.Value,
			Type:          metric.ConfigTypeBool,
		}
		switch val.Key {
		case ConfigDSN:
			option.ChooseOnly = false
			option.Default = "sys/password@127.0.0.1:1521?as=sysdba"
			option.Type = metric.ConsifTypeString
		case ConfigLibDir:
			option.ChooseOnly = false
			option.Default = "oci"
			option.Type = metric.ConsifTypeString
		}
		configOptions = append(configOptions, option)
	}
	config := map[string]interface{}{
		metric.OptionString:     configOptions,
		metric.AttributesString: KeyOracleUsages,
	}
	return config
}
func (o *Oracle) Close() {
	if o.db != nil {
		o.db.Close()
	}
}

func HandleCrash() {
	if r := recover(); r != nil {
		log.Error("Recovered in f", r)
	}
}
func merge(datas []map[string]interface{}) map[string]interface{} {
	ret := make(map[string]interface{})
	for _, data := range datas {
		for key, value := range data {
			ret[key] = value
		}

	}
	return ret
}
func (o *Oracle) initDB() error {
	os.Setenv("NLS_LANG", "")
	if len(o.DSN) == 0 {
		return fmt.Errorf("invalid oracle connection string")
	}
	var err error
	o.db, err = sql.Open("oci8", o.DSN)
	if err != nil {
		return fmt.Errorf("sql open error: %v", err.Error())
	}
	return nil
}
func init() {
	//metric.Add(TypeMetricSystem, func() metric.Collector {
	//	return &Oracle{}
	//})
}
