package types

import (
	"cmp"
	"encoding/binary"
	"hash/fnv"
	"math"
	"slices"
)

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

func (s Survey) Hash() uint64 {
	h := fnv.New64a()
	h.Write([]byte(s.CMDR))
	h.Write([]byte(s.Project))
	h.Write([]byte(s.Name))

	sorted := slices.SortedFunc(slices.Values(s.SurveyPoints),
		func(a, b SurveyPoint) int {
			return cmp.Compare(a.SystemName, b.SystemName)
		})

	buf := make([]byte, 8)
	for _, sp := range sorted {
		h.Write([]byte(sp.SystemName))
		binary.LittleEndian.PutUint32(buf, uint32(sp.ZSample))
		h.Write(buf[:4])
		binary.LittleEndian.PutUint32(buf, uint32(sp.Count))
		h.Write(buf[:4])
		binary.LittleEndian.PutUint32(buf, math.Float32bits(sp.MaxDistance))
		h.Write(buf[:4])
	}
	return h.Sum64()
}

type FolderProcessingJob struct {
	ProcID   int
	FolderID int
	GCPID    string
}
