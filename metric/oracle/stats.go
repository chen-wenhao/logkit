package oracle

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/qiniu/log"
)

// metric queries
const (
	SysstatViewQuery = `SELECT name,value FROM v$sysstat`

	TableSpaceQuery            = `SELECT tablespace_name,status FROM dba_tablespaces`
	TableSpaceUsedPercentQuery = `
	SELECT A.tablespace_name,round(((1-(A.total)/B.total)*100),2) as used_percent
	FROM (select tablespace_name,sum(bytes) total
	FROM dba_free_space group by tablespace_name) A,
	(select tablespace_name,sum(bytes) total
	FROM dba_data_files group by tablespace_name) B
	WHERE A.tablespace_name=B.tablespace_name`
)

func (o *Oracle) fetchSessionStats() map[string]interface{} {
	sessionCollectors := []string{
		"maxsession","session","session_system","session_active","session_inactive",
		"dbversion","maxprocs","procnum","blocking_sessions","uptime",
	}
	return o.getMetricByKeys(sessionCollectors,sqlMap)
}
func (o *Oracle) fetchHitratioStats() map[string]interface{} {
	hitRatioCollectors := []string{
		"buffercache_hitratio","librarycache_hitratio","dictionarycache_hitratio",
		"librarycache_hitratio_body","librarycache_hitratio_table_proc",
		"librarycache_hitratio_sqlarea","librarycache_hitratio_trigger",
		"pinhitratio_body","pinhitratio_sqlarea",
		"pinhitratio_trigger","pinhitratio_table_proc",
		"redo_log_allocation_ratio",
	}
	return o.getMetricByKeys(hitRatioCollectors,sqlMap)
}
func (o *Oracle)fetchEventwaitStats() map[string]interface{}{
	eventWaitCollectors := []string{
		"waits_directpath_read","waits_controlfileio","waits_logwrite",
		"waits_multiblock_read","waits_singleblock_read",
		"waits_sqlnet","waits_file_io","waits_latchfree",
	}
	return o.getMetricByKeys(eventWaitCollectors,sqlMap)
}
func (o *Oracle)fetchPhysicalIOStats() map[string]interface{}{
	physicalIOCollectors := []string{
		"physicalio_datafile_reads","physicalio_datafile_writes","physicalio_redo_writes",
	}
	return o.getMetricByKeys(physicalIOCollectors,sqlMap)
}

func (o *Oracle)fetchLogicalIOStats() map[string]interface{}{
	logicalIOCollectors := []string{
		"db_block_changes","db_consistent_gets","db_block_gets",
	}
	return o.getMetricByKeys(logicalIOCollectors,sqlMap)
}
func (o *Oracle) fetchSysstats() (retsSlice []map[string]interface{}, err error) {
	return Query(o.db, SysstatViewQuery, "")
}

func (o *Oracle) getTablespaceStats() (retsSlice []map[string]interface{}, err error) {
	// TODO: more stats on tablespace
	return Query(o.db, TableSpaceUsedPercentQuery, "")
}

func (o *Oracle) getMetricBySqlMap(sqlMap map[string]string) map[string]interface{} {
	rets := make(map[string]interface{})
	for k, sqlStr := range sqlMap {
		if ret, err := QueryVar(o.db, sqlStr); err == nil {
			rets[k] = ret
		} else {
			log.Errorf("fialed to exec sql: %v error: %v",k, err.Error())
		}

	}
	return rets
}
func (o *Oracle) getMetricByKeys(keys []string,sqlMap map[string]string) map[string]interface{} {
	rets := make(map[string]interface{})
	for _, key := range keys {
		if sqlStr,exists := sqlMap[key];exists {
			if ret, err := QueryVar(o.db, sqlStr); err == nil {
				rets[key] = ret
			} else {
				log.Errorf("fialed to exec sql: %v error: %v",key, err.Error())
			}
		}
	}
	return rets
}

// Query specified var
func QueryVar(db *sql.DB, sql string) ( interface{}, error) {
	rets, err := Query(db,sql,"")
	if err != nil {
		return nil, err
	}
	if len(rets) != 0 && len(rets[0]) !=0 {
		for _, v :=  range rets[0]{
			return v,nil
		}
	}
	return nil, fmt.Errorf("No row return")
}

// prefix for output key
func Query(db *sql.DB, sql string, prefix string) (retsSlice []map[string]interface{}, err error) {
	rows, err := db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("query db failed, err: %v", err.Error())
	}
	defer rows.Close()
	fields, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	scanRets:= initScans(len(fields), rows)
	for rows.Next() {
		if err := rows.Scan(scanRets...); err != nil {
			log.Errorf("Scan error: %v",err.Error())
			//continue
		}
		tempScanMap := make(map[string]interface{})
		for index, field := range fields {
			tempScanMap[formatKey(prefix, strings.ToLower(field))] =
				reflect.Indirect(reflect.ValueOf(scanRets[index])).Interface()
		}
		retsSlice = append(retsSlice, tempScanMap)
	}
	return retsSlice, nil
}

func formatKey(prefix string, args ...string) string {
	if len(prefix) == 0 {
		return strings.Join(args, "_")
	}
	return prefix + "_" + strings.Join(args, "_")
}

func initScans(length int, rows *sql.Rows) (scanArgs []interface{}) {
	tps, err := rows.ColumnTypes()
	if err != nil {
		log.Error(err)
	}
	if len(tps) != length {
		log.Errorf("getInitScans length is %v not equal to columetypes %v", length, len(tps))
	}
	scanArgs = make([]interface{}, length)
	for i, v := range tps {
		scantype := v.ScanType().String()
		switch scantype {
		case "int64":
			scanArgs[i] = new(int64)
		case "float64":
			scanArgs[i] = new(float64)
		case "[]byte":
			scanArgs[i] = new(interface{})
		case "[]string":
			scanArgs[i] = new(interface{})
		case "time.Time":
			scanArgs[i] = new(time.Time)
		default:
			scanArgs[i] = new(interface{})
		}
	}
	return
}
