package types

type Survey struct {
	CMDR         string
	Project      string
	Name         string
	SurveyPoints []SurveyPoint
}

type SurveyPoint struct {
	X          float32
	Y          float32
	Z          float32
	EDSMID     int64
	SystemName string
	ZSample    int
	Count      int
	MaxDistance float32
}

type FolderProcessingJob struct {
	ProcID   int
	FolderID int
	GCPID    string
}
