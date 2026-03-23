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
	SystemName  string
	ZSample     int
	Count       int
	MaxDistance float32
}

// Normalize deduplicates SurveyPoints by system name using the
// supplied z-coordinate map (name → actual Z coordinate).  When
// the same system appears more than once, the point whose ZSample
// is closest to the system's real Z is kept; the others are
// returned so the caller can log them.
func (s *Survey) Normalize(
	systemZ map[string]float32,
) []SurveyPoint {
	type entry struct {
		idx  int
		dist float32
	}
	best := make(map[string]entry, len(s.SurveyPoints))
	for i, sp := range s.SurveyPoints {
		z, ok := systemZ[sp.SystemName]
		var dist float32
		if ok {
			diff := z - float32(sp.ZSample)
			dist = float32(math.Abs(float64(diff)))
		} else {
			dist = math.MaxFloat32
		}
		if prev, seen := best[sp.SystemName]; !seen ||
			dist < prev.dist {
			best[sp.SystemName] = entry{i, dist}
		}
	}

	kept := make([]SurveyPoint, 0, len(best))
	dropped := make([]SurveyPoint, 0)
	for i, sp := range s.SurveyPoints {
		if best[sp.SystemName].idx == i {
			kept = append(kept, sp)
		} else {
			dropped = append(dropped, sp)
		}
	}
	s.SurveyPoints = kept
	return dropped
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
