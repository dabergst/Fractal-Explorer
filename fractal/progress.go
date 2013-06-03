// Daniel Bergström
// dabergst@kth.se

package fractal

import (
	"sync"
)

type progress struct {
	isFinished        bool
	requestedElements int
	finishedElements  int
	sync.Mutex
}

func (s *progress) newRequest(requestedElements int) {
	s.isFinished = false
	s.requestedElements = requestedElements
	s.finishedElements = 0
}

func (s *progress) IsFinished() bool {
	s.Lock()
	defer s.Unlock()
	return s.isFinished
}

func (s *progress) elementFinished() {
	s.Lock()
	defer s.Unlock()
	s.finishedElements++
	if s.finishedElements == s.requestedElements {
		s.isFinished = true
	}
}

func (s *progress) GetProgress() float64 {
	s.Lock()
	defer s.Unlock()
	return float64(s.finishedElements) / float64(s.requestedElements)
}
