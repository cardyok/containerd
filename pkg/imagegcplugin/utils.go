package imagegcplugin

// stringSet represents string set.
type stringSet map[string]struct{}

// newStringSet returns new string set.
func newStringSet(keys ...string) stringSet {
	s := stringSet{}
	s.inserts(keys...)
	return s
}

// insert updates string set.
func (s stringSet) inserts(keys ...string) {
	for _, key := range keys {
		s[key] = struct{}{}
	}
}

// contains return true if string set contains the key.
func (s stringSet) contains(key string) bool {
	_, ok := s[key]
	return ok
}