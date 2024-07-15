package outil

import "errors"

type Option[T any] struct {
	valid bool
	value T
}

func (opt *Option[T]) Set(v T) {
	opt.valid = true
	opt.value = v
}

func (opt *Option[T]) Unset() {
	opt.valid = false
}

func (opt *Option[T]) IsSet() bool {
	return opt.valid
}

func (opt *Option[T]) Get() (T, error){
	if opt.valid {
		return opt.value, nil
	} else {
		// return the default value of T and error
		return *new(T), errors.New("value is unset")
	}
}

// Returns the value regardless of whether it is known to be set
// 
// ONLY use when sure that value is set
func (opt *Option[T]) StrongGet() T {
	return opt.value
}