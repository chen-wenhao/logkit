package oracle

type tabelspace struct {
	usage int  //per
	free int   // per
	fragmentation int // Free Space Fragmentation Index
	iobalanc  int  //IO distribution under the data files of a tablespace
	remainingTime int //Number of days remaining until a tablespace is 100% occupied. The growth rate is calculated from the data of the last 30 days. (using the -lookback parameter, other periods can be specified)
canAllocate bool //Checks if there is still enough space for the next extent in the tablespace
}

var sql = `select
A.tablespace_name,round((1-(A.total)/B.total)*100,4) used_percent
from (select tablespace_name,sum(bytes) total
from dba_free_space group by tablespace_name) A,
(select tablespace_name,sum(bytes) total
from dba_data_files group by tablespace_name) B
where A.tablespace_name=B.tablespace_name;`