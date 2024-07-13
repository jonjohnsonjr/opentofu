// Copyright (c) The OpenTofu Authors
// SPDX-License-Identifier: MPL-2.0
// Copyright (c) 2023 HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dag

import "maps"

// Set is a set data structure.
type Set map[interface{}]interface{}

// Hashable is the interface used by set to get the hash code of a value.
// If this isn't given, then the value of the item being added to the set
// itself is used as the comparison value.
type Hashable interface {
	Hashcode() interface{}
}

// hashcode returns the hashcode used for set elements.
func hashcode(v interface{}) interface{} {
	return v
}

// Add adds an item to the set
func (s Set) Add(v interface{}) {
	s[v] = v
}

// Delete removes an item from the set.
func (s Set) Delete(v interface{}) {
	delete(s, v)
}

// Include returns true/false of whether a value is in the set.
func (s Set) Include(v interface{}) bool {
	_, ok := s[v]
	return ok
}

// Intersection computes the set intersection with other.
func (s Set) Intersection(other Set) Set {
	result := make(Set)
	if s == nil || other == nil {
		return result
	}
	// Iteration over a smaller set has better performance.
	if other.Len() < s.Len() {
		s, other = other, s
	}
	for _, v := range s {
		if other.Include(v) {
			result.Add(v)
		}
	}
	return result
}

// Difference returns a set with the elements that s has but
// other doesn't.
func (s Set) Difference(other Set) Set {
	if other == nil || other.Len() == 0 {
		return s.Copy()
	}

	result := make(Set)
	for k, v := range s {
		if _, ok := other[k]; !ok {
			result.Add(v)
		}
	}

	return result
}

// Filter returns a set that contains the elements from the receiver
// where the given callback returns true.
func (s Set) Filter(cb func(interface{}) bool) Set {
	result := make(Set)

	for _, v := range s {
		if cb(v) {
			result.Add(v)
		}
	}

	return result
}

// Len is the number of items in the set.
func (s Set) Len() int {
	return len(s)
}

// List returns the list of set elements.
func (s Set) List() []interface{} {
	if s == nil {
		return nil
	}

	r := make([]interface{}, 0, len(s))
	for _, v := range s {
		r = append(r, v)
	}

	return r
}

// Copy returns a shallow copy of the set.
func (s Set) Copy() Set {
	return maps.Clone(s)
}
