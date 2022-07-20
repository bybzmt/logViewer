package find

type builtin struct {
	f *Matcher
}

func NewClient() *builtin {
	return new(builtin)
}

func (s *builtin) Glob(pattern string) (out []string, err error) {
	return glob(pattern)
}

func (s *builtin) Open(m *MatchParam) (err error) {
	s.f, err = newMatcher(m)
	return
}

func (s *builtin) Read() (data []byte, err error) {
	if s.f == nil {
		return nil, notOpenFile
	}

	return s.f.Match()
}

func (s *builtin) Ping() (err error) {
	return nil
}

func (s *builtin) Close() (err error) {
	if s.f != nil {
		s.f.Close()
		s.f = nil
	}
	return nil
}
