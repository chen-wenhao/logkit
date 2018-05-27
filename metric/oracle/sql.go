package oracle

var sqlMap = map[string]string{
	"archivelog_switch": archivelog_switch,
	"uptime":            uptime,

	"db_block_gets":          db_block_gets,
	"db_consistent_gets":     db_consistent_gets,
	"dbphysicalreads":      dbphysicalreads,
	"db_block_changes":       db_block_changes,
	"buffercache_hitratio": buffercache_hitratio,

	"dictionarycache_hitratio": dictionarycache_hitratio,

	"librarycache_hitratio":            librarycache_hitratio,
	"librarycache_hitratio_body":       hitratio_body,
	"librarycache_hitratio_table_proc": hitratio_table_proc,
	"librarycache_hitratio_sqlarea":    hitratio_sqlarea,
	"librarycache_hitratio_trigger":    hitratio_trigger,

	"miss_latch":       miss_latch,
	"latches_hitratio": latches_hitratio,

	"pga_aggregate_target": pga_aggregate_target,
	"pga": pga,

	"physicalio_datafile_reads":    physicalio_datafile_reads,
	"physicalio_datafile_writes":   physicalio_datafile_writes,
	"physicalio_redo_writes":       physicalio_redo_writes,

	"pinhitratio_body":       pinhitratio_body,
	"pinhitratio_sqlarea":    pinhitratio_sqlarea,
	"pinhitratio_trigger":    pinhitratio_trigger,
	"pinhitratio_table_proc": pinhitratio_table_proc,

	"sharedpool_dict_cache": sharedpool_dict_cache,
	"sharedpool_free_mem":   sharedpool_free_mem,
	"sharedpool_sql_area":   sharedpool_sql_area,
	"sharedpool_misc":       sharedpool_misc,
	"sharedpool_lib_cache":  sharedpool_lib_cache,

	"sga_buffer_cache": sga_buffer_cache,
	"sga_shared_pool":  sga_shared_pool,
	"sga_fixed":        sga_fixed,
	"sga_java_pool":    sga_java_pool,
	"sga_large_pool":   sga_large_pool,
	"sga_log_buffer":   sga_log_buffer,

	"dbversion":        dbversion,
	"maxprocs":         maxprocs,
	"procnum":          procnum,
	"maxsession":       maxsession,
	"session":          session,
	"session_system":   session_system,
	"session_active":   session_active,
	"session_inactive": session_inactive,

	"waits_directpath_read":  waits_directpath_read,
	"waits_controlfileio":    waits_controlfileio,
	"waits_logwrite":         waits_logwrite,
	"waits_multiblock_read":  waits_multiblock_read,
	"waits_singleblock_read": waits_singleblock_read,
	"waits_sqlnet":           waits_sqlnet,
	"waits_file_io":          waits_file_io,
	"waits_latchfree":          waits_latchfree,

	"blocking_sessions_full": blocking_sessions_full,
	"blocking_sessions":      blocking_sessions,

	"redo_log_allocation_ratio":      redo_log_allocation_ratio,
}

const (
	archivelog_switch = `
        select count(*)
        from gv$log_history
        where first_time >= (sysdate - 1 / 24)
    `

	uptime = `
        select to_char ( (sysdate - startup_time) * 86400, 'FM99999999999999990')
        from gv$instance
    `
	db_block_gets = `
        select sum (value)
        from gv$sysstat
        where name = 'db block gets'
    `
	db_block_changes = `
        select sum (value)
        from gv$sysstat
        where name = 'db block changes'
`
	db_consistent_gets = `
        select sum (value)
        from gv$sysstat
        where name = 'consistent gets'
`
	dbphysicalreads = `
        select sum (value)
        from gv$sysstat
        where name = 'physical reads'
`
	buffercache_hitratio = `
        select round(( sum (case name when 'consistent gets' then value else 0 end)
        + sum (case name when 'db block gets' then value else 0 end)
        - sum (case name when 'physical reads' then value else 0 end))
        / ( sum (case name when 'consistent gets' then value else 0 end)
        + sum (case name when 'db block gets' then value else 0 end))
        * 100,2)
        from gv$sysstat
`
	librarycache_hitratio = `
	SELECT ROUND(sum(pinhits)/sum(pins)*100,2)
	FROM gv$librarycache
`
	hitratio_body = `
        SELECT ROUND(gethitratio * 100,2)
        FROM gv$librarycache
        WHERE namespace = 'BODY'
`
	hitratio_sqlarea = `
        SELECT ROUND(gethitratio * 100,2)
        FROM gv$librarycache
        WHERE namespace = 'SQL AREA'
`
	hitratio_trigger = `
        SELECT ROUND(gethitratio * 100,2)
        FROM gv$librarycache
        WHERE namespace = 'TRIGGER'
`
	hitratio_table_proc = `
        SELECT ROUND(gethitratio * 100,2)
        FROM gv$librarycache
        WHERE namespace = 'TABLE/PROCEDURE'
`
	miss_latch = `
        select sum (misses) from gv$latch
`
	pga_aggregate_target = `
        select value
        from gv$pgastat
        where name = 'aggregate PGA target parameter'
`
	pga = `
        select value
        from gv$pgastat
        where name = 'total PGA inuse'
`
	physicalio_datafile_reads = `
        select sum (value)
        from gv$sysstat
        where name = 'physical reads direct'
`
	physicalio_datafile_writes = `
        select sum (value)
        from gv$sysstat
        where name = 'physical writes direct'
`
	physicalio_redo_writes = `
        select sum (value)
        from gv$sysstat
        where name = 'redo writes'
`
	pinhitratio_body = `
        select round(pins / (pins + reloads) * 100,2)
        from gv$librarycache
        where namespace = 'BODY'
`
	pinhitratio_sqlarea = `
        select round(pins / (pins + reloads) * 100,2)
        from gv$librarycache
        where namespace = 'SQL AREA'
`
	pinhitratio_trigger = `
        select round(pins / (pins + reloads) * 100,2)
        from gv$librarycache
        where namespace = 'TRIGGER'
`
	pinhitratio_table_proc = `
        select round(pins / (pins + reloads) * 100,2)
        from gv$librarycache
        where namespace = 'TABLE/PROCEDURE'
`
	sharedpool_dict_cache = `
        select bytes
        from gv$sgastat
        where pool = 'shared pool' and name = 'dictionary cache'
   `
	sharedpool_free_mem = `
        select bytes
        from gv$sgastat
        where pool = 'shared pool' and name = 'free memory'
`
	sharedpool_lib_cache = `
        select bytes
        from gv$sgastat
        where pool = 'shared pool' and name = 'library cache'
    `
	sharedpool_sql_area = `
        select bytes
        from gv$sgastat
        where pool = 'shared pool' and name = 'sql area'
    `
	sharedpool_misc = `
        select sum (bytes)
        from gv$sgastat
        where pool = 'shared pool'
        and name not in ('library cache'
        , 'dictionary cache'
        , 'free memory'
        , 'sql area')
`
	maxprocs = `
        select value from gv$parameter where name = 'processes'
`
	procnum = `
        select count (*) from gv$process
`
	maxsession = `
        select value from gv$parameter where name = 'sessions'
`
	session = `
        select count (*) from gv$session
`
	session_system = `
        select count (*)
        from gv$session
        where type = 'BACKGROUND'
`
	session_active = `
        select count (*)
        from gv$session
        where type != 'BACKGROUND' and status = 'ACTIVE'
`
	session_inactive = `
        select count (*)
        from gv$session
        where type != 'BACKGROUND' and status = 'INACTIVE'
`
	sga_buffer_cache = `
        select sum (bytes)
        from gv$sgastat
        where name in ('db_block_buffers', 'buffer_cache')
`
	sga_fixed = `
        select sum (bytes)
        from gv$sgastat
        where name = 'fixed_sga'
`
	sga_java_pool = `
        select sum (bytes)
        from gv$sgastat
        where pool = 'java pool'
`
	sga_large_pool = `
        select sum (bytes)
        from gv$sgastat
        where pool = 'large pool'
`
	sga_shared_pool = `
        select sum (bytes)
        from gv$sgastat
        where pool = 'shared pool'
`
	sga_log_buffer = `
        select sum (bytes)
        from gv$sgastat
        where name = 'log_buffer'
`
	waits_directpath_read = `
        select total_waits
        from gv$system_event
        where event = 'direct path read'
`
	waits_file_io = `
        select nvl (sum (total_waits), 0)
        from gv$system_event
        where event in ('file identify', 'file open')
`
	waits_controlfileio = `
        select sum (total_waits)
        from gv$system_event
        where event in ('control file sequential read'
        , 'control file single write'
        , 'control file parallel write')
`
	waits_logwrite = `
        select sum (total_waits)
        from gv$system_event
        where event in ('log file single write', 'log file parallel write')
`
	waits_logsync = `
        select sum(total_waits)
        from gv$system_event
        where event = 'log file sync'
`
	waits_multiblock_read = `
        select sum (total_waits)
        from gv$system_event
        where event = 'db file scattered read'
`
	waits_singleblock_read = `
        select sum (total_waits)
        from gv$system_event
        where event = 'db file sequential read'
`
	waits_sqlnet = `
        select sum (total_waits)
        from gv$system_event
        where event in ('SQL*Net message to client'
        , 'SQL*Net message to dblink'
        , 'SQL*Net more data to client'
        , 'SQL*Net more data to dblink'
        , 'SQL*Net break/reset to client'
        , 'SQL*Net break/reset to dblink')
`
	waits_latchfree = `
        select sum (total_waits)
        from gv$system_event
        where event = 'latch free'
`
	blocking_sessions = `
select count(*)
  from (
            select rootid
              from (
                          select level lvl
                               , connect_by_root (inst_id || '.' || sid) rootid
                               , seconds_in_wait
                            from gv$session
                      start with blocking_session is null
                      connect by nocycle prior inst_id = blocking_instance
                                     and prior sid = blocking_session
                   )
             where lvl > 1
          group by rootid
            having sum(seconds_in_wait) > 300
       )
        `
	blocking_sessions_full = `
    select listagg(lpad(' ', (level - 1) * 4)
           || 'INST_ID         :  '
           || inst_id
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'SERVICE_NAME    :  '
           || service_name
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'SID,SERIAL      :  '
           || sid
           || ','
           || serial#
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'USERNAME        :  '
           || username
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'OSUSER          :  '
           || osuser
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'MACHINE         :  '
           || machine
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'PROGRAM         :  '
           || program
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'MODULE          :  '
           || module
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'SQL_ID          :  '
           || sql_id
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'EVENT           :  '
           || event
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'SECONDS_IN_WAIT :  '
           || seconds_in_wait
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'STATE           :  '
           || state
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || 'STATUS          :  '
           || status
           || chr(10)
           || lpad(' ', (level - 1) * 4)
           || '========================='
           , chr(10)) within group (order by level) as blocking_sess_info
      from (
                  select inst_id || '.' || sid id
                       , case
                            when blocking_instance is not null
                            then
                               blocking_instance || '.' || blocking_session
                         end
                            parent_id
                       , inst_id
                       , service_name
                       , sid
                       , serial#
                       , username
                       , osuser
                       , machine
                       , program
                       , module
                       , sql_id
                       , event
                       , seconds_in_wait
                       , state
                       , status
                       , level lvl
                       , connect_by_isleaf isleaf
                       , connect_by_root (inst_id || '.' || sid) rootid
                    from gv$session
              start with blocking_session is null
              connect by nocycle prior inst_id = blocking_instance
                             and prior sid = blocking_session
           )
     where lvl || isleaf <> '11'
       and rootid in
              (
                   select rootid
                     from (
                                 select level lvl
                                      , connect_by_root (inst_id || '.' || sid) rootid
                                      , seconds_in_wait
                                   from gv$session
                             start with blocking_session is null
                             connect by nocycle prior inst_id = blocking_instance
                                            and prior sid = blocking_session
                          )
                    where lvl > 1
                 group by rootid
                   having sum(seconds_in_wait) > 300
              )
connect by nocycle prior id = parent_id
start with parent_id is null
    `
	dbversion = `
        select banner
        from gv$version
        where banner like '%Oracle Database%'
`
	tablespaces = `
            select name ts from gv$tablespace
        `
	wait_classes = `
            select distinct wait_class as class
            from gv$system_event
        `
	ts_usage_pct = `
        query = q{
            select tablespace_name ts, round(used_percent, 5) pct
            from dba_tablespace_usage_metrics
       ` //不适用于自动增长ts
	ts_usage_bytes = `
            select ta.tablespace_name as ts, ta.used_space * tb.block_size as bytes
            from dba_tablespace_usage_metrics ta
            join dba_tablespaces tb on ta.tablespace_name = tb.tablespace_name
        `
	waits_ms = `
          select ta.wait_class as class, sum(ta.total_waits) as waits_ms
          from gv$system_event ta
          where event not in ('SQL*Net message from client', 'pipe get')
          group by ta.wait_class
       `
	dictionarycache_hitratio = `
	SELECT ROUND((sum(gets-getmisses-usage-fixed))/sum(gets)*100,2) "Data dictionary cache"
	FROM v$rowcache
	`
	latches_hitratio = `
	SELECT ROUND((1-(misses/gets))*100,2)
	FROM v$latch
	WHERE name in ('library cache', 'shared pool')
	`
	redo_log_allocation_ratio = `
	SELECT ROUND(100*(a.value/b.value),2)
	FROM v$sysstat a, v$sysstat b
	WHERE a.name = 'redo log space requests' AND b.name = 'redo entries'
	`
)

