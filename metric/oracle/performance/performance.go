package oracle

type sga struct {
	databufferHitRatio      int // Hitrate in the data buffer cache 0% .. 100% (98 :, 95 :)
	libraryCacheGetHitRatio int // Hitrate in the Library Cache (Gets) 0% .. 100% (98 :, 95 :)
	libraryCachePinhitRatio int //Hitrate in the Library Cache (Pins) 0% .. 100% (98 :, 95 :)
	libraryCacheReloads     int // Reload rate in library cache n / sec (10,10)
	dictionaryCacheHitRatio int //Hitrate in the Dictionary Cache
	latchesHitRatio         int // Hitrate of the latches 0% .. 100% (98 :, 95 :)
	sharedPoolReloads       int // Reload rate in the shared pool 0% .. 100% (1, 10)
	sharedPoolFree          int //  Free storage in the shared pool

}

type pga struct {
	inMemorySortRatio int //Percentage of sorts in memory
}

type invalidObjects int  // Number of defective objects, indices, partitions
type staleStatistics int // Number of objects with obsolete Optimizer statistics

type flash struct {
	recoveryAreaUsage int //Used space in the Flash Recovery Area 0% .. 100% (90, 98)
	recoveryAreaFree  int // Free space in the Flash Recovery Area 0% .. 100% (5 :, 2 :)

}
type dataFileIOTraffic int //Number of IO operations of data files per second n / sec (1000, 5000)
type dataFilesExisting int //Percentage of the maximum possible data files 0% .. 100% (80, 90)
type softParseRatio int    //The proportion of soft-parse calls 0% .. 100%
var switchInterval int     // Interval between RedoLog File Switches 0..n seconds
var retryRatio int         //Retry rate in the RedoLog buffer 0% .. 100% (1, 10)
var redoIOTraffic int      //Redo IO in MB / sec
type RollbackSegment struct {
	rollHeaderContention int // Rollback Segment Header Contention 0% .. 100% (1, 2)
}
