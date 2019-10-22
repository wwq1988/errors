package errors

import (
	"fmt"
	"strings"
	"sync"

	"github.com/wwq1988/errors/stack"
)

// timeout 判断是否超时接口
type timeout interface {
	Timeout() bool
}

// temporary 判断是否临时接口
type temporary interface {
	Temporary() bool
}

type stackError struct {
	fields stack.Fields
	once   sync.Once
	stacks []string
	err    error
}

// New 初始化stackError
func New(msg string, args ...interface{}) error {
	return NewEx(2, fmt.Errorf(msg, args...), nil)
}

// NewWithFields 初始化stackError带域存储
func NewWithFields(msg string, fields stack.Fields) error {
	return NewEx(2, fmt.Errorf(msg), fields)
}

// NewWithField 初始化stackError带kv
func NewWithField(msg string, key string, val interface{}) error {
	fs := stack.New().Set(key, val)
	return NewEx(2, fmt.Errorf(msg), fs)
}

// NewEx 初始化stackError,带堆栈深度和域存储
func NewEx(depth int, err error, fields stack.Fields) error {
	stackFrame := stack.Get(depth)
	se, ok := err.(*stackError)
	if !ok {
		if fields == nil {
			fields = stack.New()
		}
		return &stackError{
			err:    err,
			fields: fields,
			stacks: []string{stackFrame},
		}
	}
	se.stacks = append(se.stacks, stackFrame)
	if se.fields == nil {
		se.fields = stack.New()
	}
	if fields != nil {
		se.fields.Merge(fields)
	}
	return se
}

// WithField 携带域存储
func (s *stackError) WithField(key string, val interface{}) {
	s.fields.Set(key, val)
}

// WithField 携带域存储
func (s *stackError) WithFields(fields stack.Fields) {
	s.fields.Merge(fields)
}

// Fields 获取域存储
func (s *stackError) Fields() stack.Fields {
	s.once.Do(s.fillStackField)
	return s.fields
}

func (s *stackError) fillStackField() {
	s.fields.Set("stack", s.stack())
}

func (s *stackError) stack() string {
	if len(s.stacks) == 0 {
		return ""
	}
	return strings.Join(s.stacks, ";")
}

func (s *stackError) cleanFields() {
	s.fields = stack.New()
}

func (s *stackError) Unwrap() error {
	return s.err
}

func (s *stackError) Is(err error) bool {
	return s.Unwrap() == Unwrap(err)
}

func (s *stackError) Error() string {
	return s.err.Error()
}

// Trace 追踪错误
func Trace(err error) error {
	if err == nil {
		return nil
	}
	return NewEx(2, err, nil)
}

// TraceWithFields 追踪错误带域存储
func TraceWithFields(err error, fields stack.Fields) error {
	return TraceWithFieldsEx(err, fields, 2)
}

// TraceWithFieldsEx 追踪错误带域存储和堆栈深度
func TraceWithFieldsEx(err error, fields stack.Fields, depth int) error {
	if err == nil {
		return nil
	}
	return NewEx(depth+1, err, fields)
}

// TraceWithFieldEx 追踪错误带kv和堆栈深度
func TraceWithFieldEx(err error, key string, val interface{}, depth int) error {
	fs := stack.New()
	fs.Set(key, val)
	return TraceWithFieldsEx(err, fs, depth+1)
}

// TraceWithField 追踪错误带kv
func TraceWithField(err error, key string, val interface{}) error {
	return TraceWithFieldEx(err, key, val, 2)

}

// Is 判断err是否相同
func Is(src, dst error) bool {
	return Unwrap(src) == Unwrap(dst)
}

// Unwrap 返回原始err
func Unwrap(err error) error {
	if stackErr, ok := err.(*stackError); ok {
		return stackErr.Unwrap()
	}
	return err
}

// IsTimeout 判断是否超时
func IsTimeout(err error) bool {
	timeoutErr, ok := err.(timeout)
	return ok && timeoutErr.Timeout()
}

// IsTemporary 判断是否是临时错误
func IsTemporary(err error) bool {
	temporaryErr, ok := err.(temporary)
	return ok && temporaryErr.Temporary()
}

// Fields 获取域
func Fields(err error) stack.Fields {
	if stackErr, ok := err.(*stackError); ok {
		return stackErr.Fields()
	}
	return stack.New()
}
