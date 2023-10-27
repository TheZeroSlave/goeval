package goeval

type stringQ struct {
	str string
	i   int
}

func (s *stringQ) next() (byte, bool) {
	if s.i >= len(s.str) {
		return 0, false
	}

	ch := s.str[s.i]
	s.i += 1
	return ch, true
}

func (s *stringQ) rollback() {
	s.i -= 1
}

func (s *stringQ) hasNext() bool {
	return s.i+1 < len(s.str)
}

func (s *stringQ) ok() bool {
	return s.i < len(s.str)
}
