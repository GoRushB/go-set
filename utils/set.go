package utils

import (
	"fmt"
	"reflect"
	"sync"
)

type set struct {
	mutex     sync.RWMutex
	itemType  reflect.Type
	safe      bool
	m         map[Any]int64
	l         []Any
	removeNum int64
}

func NewSet(singleton Any, unsafe ...bool) (*set, error) {
	typ := reflect.TypeOf(singleton)
	s := &set{
		itemType: typ,
		safe:     len(unsafe) == 0 || !unsafe[0],
		m:        make(map[Any]int64),
		l:        make([]Any, 0),
	}
	canEqual := true
	func() {
		defer func() {
			if err := recover(); err != nil {
				canEqual = false
			}
		}()
		_ = singleton == singleton
	}()
	if !canEqual {
		return nil, fmt.Errorf("type %s cannot be equal", typ)
	}
	if typ.Kind() == reflect.Ptr {
		return nil, fmt.Errorf("type %s is ptr", typ)
	}
	return s, nil
}

func newEmptySetByType(typ reflect.Type) *set {
	return &set{
		itemType: typ,
		safe:     true,
		m:        make(map[Any]int64),
		l:        make([]Any, 0),
	}
}

func (s *set) TryLock() func() {
	if s == nil || !s.safe {
		return func() {}
	}
	s.mutex.Lock()
	return s.mutex.Unlock
}

func (s *set) TryRLock() func() {
	if s == nil || !s.safe {
		return func() {}
	}
	s.mutex.RLock()
	return s.mutex.RUnlock
}

func (s *set) Reset() *set {
	if s == nil {
		return s
	}
	defer s.TryLock()()
	return s.reset()
}

func (s *set) reset() *set {
	s.m = make(map[Any]int64)
	s.l = make([]Any, 0)
	s.removeNum = 0
	s.safe = true
	return s
}

func (s *set) Clone(safe ...bool) *set {
	if s == nil {
		return nil
	}
	s.TryRLock()()
	return s.clone(safe...)
}

func (s *set) clone(safe ...bool) *set {
	safe = append(safe, s.safe)
	size := s.len()
	newSet := &set{
		itemType:  s.itemType,
		m:         make(map[Any]int64, size),
		l:         make([]Any, size),
		removeNum: 0,
		safe:      safe[0],
	}
	for v, i := range s.m {
		newSet.m[v] = i
		newSet.l[i] = v
	}
	return newSet
}

func (s *set) Len() int64 {
	if s == nil {
		return 0
	}
	defer s.TryRLock()()
	return s.len()
}

func (s *set) len() int64 {
	return int64(len(s.m))
}

func (s *set) Empty() bool {
	if s == nil {
		return true
	}
	defer s.TryRLock()()
	return s.empty()
}

func (s *set) empty() bool {
	return s.len() == 0
}

func (s *set) IsExist(item Any) bool {
	if s == nil || reflect.TypeOf(item) != s.itemType {
		return false
	}
	defer s.TryRLock()()
	_, exist := s.isExist(item)
	return exist
}

func (s *set) isExist(item Any) (int64, bool) {
	index, exist := s.m[item]
	return index, exist
}

func (s *set) Add(list ...Any) *set {
	if s == nil || len(list) == 0 {
		return s
	}
	typ := reflect.TypeOf(list[0])
	isSlice := typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array
	if typ != s.itemType && (isSlice && typ.Elem() != s.itemType) {
		return s
	}
	defer s.TryLock()()
	for i := range list {
		if list[i] == nil {
			continue
		}
		if !isSlice {
			s.add(list[i])
		} else {
			val := reflect.ValueOf(list[i])
			size := val.Len()
			for j := 0; j < size; j++ {
				s.add(val.Index(j).Interface())
			}
		}
	}
	return s
}

func (s *set) add(item Any) *set {
	if _, exist := s.isExist(item); exist {
		return s
	}
	s.m[item] = int64(len(s.l))
	s.l = append(s.l, item)
	return s
}

func (s *set) Remove(list ...Any) *set {
	const maxRemoveNum = 1024
	if s == nil || len(list) == 0 {
		return s
	}
	typ := reflect.TypeOf(list[0])
	isSlice := typ.Kind() == reflect.Slice || typ.Kind() == reflect.Array
	if typ != s.itemType && (isSlice && typ.Elem() != s.itemType) {
		return s
	}
	defer s.TryLock()()
	for i := range list {
		if list[i] == nil {
			continue
		}
		if !isSlice {
			s.remove(list[i])
		} else {
			val := reflect.ValueOf(list[i])
			size := val.Len()
			for j := 0; j < size; j++ {
				s.remove(val.Index(j).Interface())
			}
		}
	}
	if s.removeNum >= maxRemoveNum {
		newList := make([]Any, s.len())
		var index int64
		for _, v := range s.l {
			if v == nil {
				continue
			}
			newList[index] = v
			s.m[v] = index
			index++
		}
		s.l = newList
		s.removeNum = 0
	}
	return s
}

func (s *set) remove(item Any) *set {
	index, exist := s.isExist(item)
	if !exist {
		return s
	}
	delete(s.m, item)
	s.l[index] = nil
	s.removeNum++
	return s
}

func (s *set) IsSubsetOf(super *set) bool {
	if s == nil || super == nil || s.itemType != super.itemType {
		return false
	}
	defer s.TryRLock()()
	defer super.TryRLock()()
	if s.empty() {
		return true
	}
	if s.len() > super.len() {
		return false
	}
	for v := range s.m {
		if _, exist := super.isExist(v); !exist {
			return false
		}
	}
	return true
}

func (s *set) IsSupersetOf(sub *set) bool {
	return sub.IsSubsetOf(s)
}

func (s *set) Overlaps(other *set) bool {
	if s == nil || other == nil || s.itemType != other.itemType {
		return false
	}
	defer s.TryRLock()()
	defer other.TryRLock()()
	if s.empty() || other.empty() {
		return false
	}
	long, short := s, other
	if long.len() < short.len() {
		long, short = short, long
	}
	for v := range short.m {
		if _, exist := long.isExist(v); exist {
			return true
		}
	}
	return false
}

func (s *set) OverlapsData(other *set) *set {
	if s == nil || other == nil || s.itemType != other.itemType {
		return nil
	}
	defer s.TryRLock()()
	defer other.TryRLock()()
	res := newEmptySetByType(s.itemType)
	if s.empty() || other.empty() {
		return res
	}
	long, short := s, other
	if long.len() < short.len() {
		long, short = short, long
	}
	for v := range short.m {
		if _, exist := long.isExist(v); exist {
			res.add(v)
		}
	}
	return res
}

func (s *set) ToList() AnyList {
	if s == nil {
		return make(AnyList, 0)
	}
	defer s.TryRLock()()
	return s.toList()
}

func (s *set) toList() AnyList {
	list := make(AnyList, s.len())
	var index int64
	for _, v := range s.l {
		if v == nil {
			continue
		}
		list[index] = v
		index++
	}
	return list
}
