// Code generated by github.com/efritz/go-mockgen 0.1.0; DO NOT EDIT.

package mocks

import (
	"context"
	api "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api"
	client "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"sync"
)

// MockCodeIntelAPI is a mock implementation of the CodeIntelAPI interface
// (from the package
// github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/api)
// used for unit testing.
type MockCodeIntelAPI struct {
	// DefinitionsFunc is an instance of a mock function object controlling
	// the behavior of the method Definitions.
	DefinitionsFunc *CodeIntelAPIDefinitionsFunc
	// DiagnosticsFunc is an instance of a mock function object controlling
	// the behavior of the method Diagnostics.
	DiagnosticsFunc *CodeIntelAPIDiagnosticsFunc
	// FindClosestDumpsFunc is an instance of a mock function object
	// controlling the behavior of the method FindClosestDumps.
	FindClosestDumpsFunc *CodeIntelAPIFindClosestDumpsFunc
	// HoverFunc is an instance of a mock function object controlling the
	// behavior of the method Hover.
	HoverFunc *CodeIntelAPIHoverFunc
	// ReferencesFunc is an instance of a mock function object controlling
	// the behavior of the method References.
	ReferencesFunc *CodeIntelAPIReferencesFunc
}

// NewMockCodeIntelAPI creates a new mock of the CodeIntelAPI interface. All
// methods return zero values for all results, unless overwritten.
func NewMockCodeIntelAPI() *MockCodeIntelAPI {
	return &MockCodeIntelAPI{
		DefinitionsFunc: &CodeIntelAPIDefinitionsFunc{
			defaultHook: func(context.Context, string, int, int, int) ([]api.ResolvedLocation, error) {
				return nil, nil
			},
		},
		DiagnosticsFunc: &CodeIntelAPIDiagnosticsFunc{
			defaultHook: func(context.Context, string, int, int, int) ([]api.ResolvedDiagnostic, int, error) {
				return nil, 0, nil
			},
		},
		FindClosestDumpsFunc: &CodeIntelAPIFindClosestDumpsFunc{
			defaultHook: func(context.Context, int, string, string, bool, string) ([]store.Dump, error) {
				return nil, nil
			},
		},
		HoverFunc: &CodeIntelAPIHoverFunc{
			defaultHook: func(context.Context, string, int, int, int) (string, client.Range, bool, error) {
				return "", client.Range{}, false, nil
			},
		},
		ReferencesFunc: &CodeIntelAPIReferencesFunc{
			defaultHook: func(context.Context, int, string, int, api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error) {
				return nil, api.Cursor{}, false, nil
			},
		},
	}
}

// NewMockCodeIntelAPIFrom creates a new mock of the MockCodeIntelAPI
// interface. All methods delegate to the given implementation, unless
// overwritten.
func NewMockCodeIntelAPIFrom(i api.CodeIntelAPI) *MockCodeIntelAPI {
	return &MockCodeIntelAPI{
		DefinitionsFunc: &CodeIntelAPIDefinitionsFunc{
			defaultHook: i.Definitions,
		},
		DiagnosticsFunc: &CodeIntelAPIDiagnosticsFunc{
			defaultHook: i.Diagnostics,
		},
		FindClosestDumpsFunc: &CodeIntelAPIFindClosestDumpsFunc{
			defaultHook: i.FindClosestDumps,
		},
		HoverFunc: &CodeIntelAPIHoverFunc{
			defaultHook: i.Hover,
		},
		ReferencesFunc: &CodeIntelAPIReferencesFunc{
			defaultHook: i.References,
		},
	}
}

// CodeIntelAPIDefinitionsFunc describes the behavior when the Definitions
// method of the parent MockCodeIntelAPI instance is invoked.
type CodeIntelAPIDefinitionsFunc struct {
	defaultHook func(context.Context, string, int, int, int) ([]api.ResolvedLocation, error)
	hooks       []func(context.Context, string, int, int, int) ([]api.ResolvedLocation, error)
	history     []CodeIntelAPIDefinitionsFuncCall
	mutex       sync.Mutex
}

// Definitions delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockCodeIntelAPI) Definitions(v0 context.Context, v1 string, v2 int, v3 int, v4 int) ([]api.ResolvedLocation, error) {
	r0, r1 := m.DefinitionsFunc.nextHook()(v0, v1, v2, v3, v4)
	m.DefinitionsFunc.appendCall(CodeIntelAPIDefinitionsFuncCall{v0, v1, v2, v3, v4, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the Definitions method
// of the parent MockCodeIntelAPI instance is invoked and the hook queue is
// empty.
func (f *CodeIntelAPIDefinitionsFunc) SetDefaultHook(hook func(context.Context, string, int, int, int) ([]api.ResolvedLocation, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Definitions method of the parent MockCodeIntelAPI instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *CodeIntelAPIDefinitionsFunc) PushHook(hook func(context.Context, string, int, int, int) ([]api.ResolvedLocation, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CodeIntelAPIDefinitionsFunc) SetDefaultReturn(r0 []api.ResolvedLocation, r1 error) {
	f.SetDefaultHook(func(context.Context, string, int, int, int) ([]api.ResolvedLocation, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CodeIntelAPIDefinitionsFunc) PushReturn(r0 []api.ResolvedLocation, r1 error) {
	f.PushHook(func(context.Context, string, int, int, int) ([]api.ResolvedLocation, error) {
		return r0, r1
	})
}

func (f *CodeIntelAPIDefinitionsFunc) nextHook() func(context.Context, string, int, int, int) ([]api.ResolvedLocation, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeIntelAPIDefinitionsFunc) appendCall(r0 CodeIntelAPIDefinitionsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CodeIntelAPIDefinitionsFuncCall objects
// describing the invocations of this function.
func (f *CodeIntelAPIDefinitionsFunc) History() []CodeIntelAPIDefinitionsFuncCall {
	f.mutex.Lock()
	history := make([]CodeIntelAPIDefinitionsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeIntelAPIDefinitionsFuncCall is an object that describes an invocation
// of method Definitions on an instance of MockCodeIntelAPI.
type CodeIntelAPIDefinitionsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 int
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []api.ResolvedLocation
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CodeIntelAPIDefinitionsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CodeIntelAPIDefinitionsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// CodeIntelAPIDiagnosticsFunc describes the behavior when the Diagnostics
// method of the parent MockCodeIntelAPI instance is invoked.
type CodeIntelAPIDiagnosticsFunc struct {
	defaultHook func(context.Context, string, int, int, int) ([]api.ResolvedDiagnostic, int, error)
	hooks       []func(context.Context, string, int, int, int) ([]api.ResolvedDiagnostic, int, error)
	history     []CodeIntelAPIDiagnosticsFuncCall
	mutex       sync.Mutex
}

// Diagnostics delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockCodeIntelAPI) Diagnostics(v0 context.Context, v1 string, v2 int, v3 int, v4 int) ([]api.ResolvedDiagnostic, int, error) {
	r0, r1, r2 := m.DiagnosticsFunc.nextHook()(v0, v1, v2, v3, v4)
	m.DiagnosticsFunc.appendCall(CodeIntelAPIDiagnosticsFuncCall{v0, v1, v2, v3, v4, r0, r1, r2})
	return r0, r1, r2
}

// SetDefaultHook sets function that is called when the Diagnostics method
// of the parent MockCodeIntelAPI instance is invoked and the hook queue is
// empty.
func (f *CodeIntelAPIDiagnosticsFunc) SetDefaultHook(hook func(context.Context, string, int, int, int) ([]api.ResolvedDiagnostic, int, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Diagnostics method of the parent MockCodeIntelAPI instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *CodeIntelAPIDiagnosticsFunc) PushHook(hook func(context.Context, string, int, int, int) ([]api.ResolvedDiagnostic, int, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CodeIntelAPIDiagnosticsFunc) SetDefaultReturn(r0 []api.ResolvedDiagnostic, r1 int, r2 error) {
	f.SetDefaultHook(func(context.Context, string, int, int, int) ([]api.ResolvedDiagnostic, int, error) {
		return r0, r1, r2
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CodeIntelAPIDiagnosticsFunc) PushReturn(r0 []api.ResolvedDiagnostic, r1 int, r2 error) {
	f.PushHook(func(context.Context, string, int, int, int) ([]api.ResolvedDiagnostic, int, error) {
		return r0, r1, r2
	})
}

func (f *CodeIntelAPIDiagnosticsFunc) nextHook() func(context.Context, string, int, int, int) ([]api.ResolvedDiagnostic, int, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeIntelAPIDiagnosticsFunc) appendCall(r0 CodeIntelAPIDiagnosticsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CodeIntelAPIDiagnosticsFuncCall objects
// describing the invocations of this function.
func (f *CodeIntelAPIDiagnosticsFunc) History() []CodeIntelAPIDiagnosticsFuncCall {
	f.mutex.Lock()
	history := make([]CodeIntelAPIDiagnosticsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeIntelAPIDiagnosticsFuncCall is an object that describes an invocation
// of method Diagnostics on an instance of MockCodeIntelAPI.
type CodeIntelAPIDiagnosticsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 int
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []api.ResolvedDiagnostic
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 int
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CodeIntelAPIDiagnosticsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CodeIntelAPIDiagnosticsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2}
}

// CodeIntelAPIFindClosestDumpsFunc describes the behavior when the
// FindClosestDumps method of the parent MockCodeIntelAPI instance is
// invoked.
type CodeIntelAPIFindClosestDumpsFunc struct {
	defaultHook func(context.Context, int, string, string, bool, string) ([]store.Dump, error)
	hooks       []func(context.Context, int, string, string, bool, string) ([]store.Dump, error)
	history     []CodeIntelAPIFindClosestDumpsFuncCall
	mutex       sync.Mutex
}

// FindClosestDumps delegates to the next hook function in the queue and
// stores the parameter and result values of this invocation.
func (m *MockCodeIntelAPI) FindClosestDumps(v0 context.Context, v1 int, v2 string, v3 string, v4 bool, v5 string) ([]store.Dump, error) {
	r0, r1 := m.FindClosestDumpsFunc.nextHook()(v0, v1, v2, v3, v4, v5)
	m.FindClosestDumpsFunc.appendCall(CodeIntelAPIFindClosestDumpsFuncCall{v0, v1, v2, v3, v4, v5, r0, r1})
	return r0, r1
}

// SetDefaultHook sets function that is called when the FindClosestDumps
// method of the parent MockCodeIntelAPI instance is invoked and the hook
// queue is empty.
func (f *CodeIntelAPIFindClosestDumpsFunc) SetDefaultHook(hook func(context.Context, int, string, string, bool, string) ([]store.Dump, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// FindClosestDumps method of the parent MockCodeIntelAPI instance inovkes
// the hook at the front of the queue and discards it. After the queue is
// empty, the default hook function is invoked for any future action.
func (f *CodeIntelAPIFindClosestDumpsFunc) PushHook(hook func(context.Context, int, string, string, bool, string) ([]store.Dump, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CodeIntelAPIFindClosestDumpsFunc) SetDefaultReturn(r0 []store.Dump, r1 error) {
	f.SetDefaultHook(func(context.Context, int, string, string, bool, string) ([]store.Dump, error) {
		return r0, r1
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CodeIntelAPIFindClosestDumpsFunc) PushReturn(r0 []store.Dump, r1 error) {
	f.PushHook(func(context.Context, int, string, string, bool, string) ([]store.Dump, error) {
		return r0, r1
	})
}

func (f *CodeIntelAPIFindClosestDumpsFunc) nextHook() func(context.Context, int, string, string, bool, string) ([]store.Dump, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeIntelAPIFindClosestDumpsFunc) appendCall(r0 CodeIntelAPIFindClosestDumpsFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CodeIntelAPIFindClosestDumpsFuncCall
// objects describing the invocations of this function.
func (f *CodeIntelAPIFindClosestDumpsFunc) History() []CodeIntelAPIFindClosestDumpsFuncCall {
	f.mutex.Lock()
	history := make([]CodeIntelAPIFindClosestDumpsFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeIntelAPIFindClosestDumpsFuncCall is an object that describes an
// invocation of method FindClosestDumps on an instance of MockCodeIntelAPI.
type CodeIntelAPIFindClosestDumpsFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 string
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 bool
	// Arg5 is the value of the 6th argument passed to this method
	// invocation.
	Arg5 string
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []store.Dump
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CodeIntelAPIFindClosestDumpsFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4, c.Arg5}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CodeIntelAPIFindClosestDumpsFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1}
}

// CodeIntelAPIHoverFunc describes the behavior when the Hover method of the
// parent MockCodeIntelAPI instance is invoked.
type CodeIntelAPIHoverFunc struct {
	defaultHook func(context.Context, string, int, int, int) (string, client.Range, bool, error)
	hooks       []func(context.Context, string, int, int, int) (string, client.Range, bool, error)
	history     []CodeIntelAPIHoverFuncCall
	mutex       sync.Mutex
}

// Hover delegates to the next hook function in the queue and stores the
// parameter and result values of this invocation.
func (m *MockCodeIntelAPI) Hover(v0 context.Context, v1 string, v2 int, v3 int, v4 int) (string, client.Range, bool, error) {
	r0, r1, r2, r3 := m.HoverFunc.nextHook()(v0, v1, v2, v3, v4)
	m.HoverFunc.appendCall(CodeIntelAPIHoverFuncCall{v0, v1, v2, v3, v4, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefaultHook sets function that is called when the Hover method of the
// parent MockCodeIntelAPI instance is invoked and the hook queue is empty.
func (f *CodeIntelAPIHoverFunc) SetDefaultHook(hook func(context.Context, string, int, int, int) (string, client.Range, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// Hover method of the parent MockCodeIntelAPI instance inovkes the hook at
// the front of the queue and discards it. After the queue is empty, the
// default hook function is invoked for any future action.
func (f *CodeIntelAPIHoverFunc) PushHook(hook func(context.Context, string, int, int, int) (string, client.Range, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CodeIntelAPIHoverFunc) SetDefaultReturn(r0 string, r1 client.Range, r2 bool, r3 error) {
	f.SetDefaultHook(func(context.Context, string, int, int, int) (string, client.Range, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CodeIntelAPIHoverFunc) PushReturn(r0 string, r1 client.Range, r2 bool, r3 error) {
	f.PushHook(func(context.Context, string, int, int, int) (string, client.Range, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *CodeIntelAPIHoverFunc) nextHook() func(context.Context, string, int, int, int) (string, client.Range, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeIntelAPIHoverFunc) appendCall(r0 CodeIntelAPIHoverFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CodeIntelAPIHoverFuncCall objects
// describing the invocations of this function.
func (f *CodeIntelAPIHoverFunc) History() []CodeIntelAPIHoverFuncCall {
	f.mutex.Lock()
	history := make([]CodeIntelAPIHoverFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeIntelAPIHoverFuncCall is an object that describes an invocation of
// method Hover on an instance of MockCodeIntelAPI.
type CodeIntelAPIHoverFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 string
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 int
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 int
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 int
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 string
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 client.Range
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 bool
	// Result3 is the value of the 4th result returned from this method
	// invocation.
	Result3 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CodeIntelAPIHoverFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CodeIntelAPIHoverFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2, c.Result3}
}

// CodeIntelAPIReferencesFunc describes the behavior when the References
// method of the parent MockCodeIntelAPI instance is invoked.
type CodeIntelAPIReferencesFunc struct {
	defaultHook func(context.Context, int, string, int, api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error)
	hooks       []func(context.Context, int, string, int, api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error)
	history     []CodeIntelAPIReferencesFuncCall
	mutex       sync.Mutex
}

// References delegates to the next hook function in the queue and stores
// the parameter and result values of this invocation.
func (m *MockCodeIntelAPI) References(v0 context.Context, v1 int, v2 string, v3 int, v4 api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error) {
	r0, r1, r2, r3 := m.ReferencesFunc.nextHook()(v0, v1, v2, v3, v4)
	m.ReferencesFunc.appendCall(CodeIntelAPIReferencesFuncCall{v0, v1, v2, v3, v4, r0, r1, r2, r3})
	return r0, r1, r2, r3
}

// SetDefaultHook sets function that is called when the References method of
// the parent MockCodeIntelAPI instance is invoked and the hook queue is
// empty.
func (f *CodeIntelAPIReferencesFunc) SetDefaultHook(hook func(context.Context, int, string, int, api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error)) {
	f.defaultHook = hook
}

// PushHook adds a function to the end of hook queue. Each invocation of the
// References method of the parent MockCodeIntelAPI instance inovkes the
// hook at the front of the queue and discards it. After the queue is empty,
// the default hook function is invoked for any future action.
func (f *CodeIntelAPIReferencesFunc) PushHook(hook func(context.Context, int, string, int, api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error)) {
	f.mutex.Lock()
	f.hooks = append(f.hooks, hook)
	f.mutex.Unlock()
}

// SetDefaultReturn calls SetDefaultDefaultHook with a function that returns
// the given values.
func (f *CodeIntelAPIReferencesFunc) SetDefaultReturn(r0 []api.ResolvedLocation, r1 api.Cursor, r2 bool, r3 error) {
	f.SetDefaultHook(func(context.Context, int, string, int, api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error) {
		return r0, r1, r2, r3
	})
}

// PushReturn calls PushDefaultHook with a function that returns the given
// values.
func (f *CodeIntelAPIReferencesFunc) PushReturn(r0 []api.ResolvedLocation, r1 api.Cursor, r2 bool, r3 error) {
	f.PushHook(func(context.Context, int, string, int, api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error) {
		return r0, r1, r2, r3
	})
}

func (f *CodeIntelAPIReferencesFunc) nextHook() func(context.Context, int, string, int, api.Cursor) ([]api.ResolvedLocation, api.Cursor, bool, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()

	if len(f.hooks) == 0 {
		return f.defaultHook
	}

	hook := f.hooks[0]
	f.hooks = f.hooks[1:]
	return hook
}

func (f *CodeIntelAPIReferencesFunc) appendCall(r0 CodeIntelAPIReferencesFuncCall) {
	f.mutex.Lock()
	f.history = append(f.history, r0)
	f.mutex.Unlock()
}

// History returns a sequence of CodeIntelAPIReferencesFuncCall objects
// describing the invocations of this function.
func (f *CodeIntelAPIReferencesFunc) History() []CodeIntelAPIReferencesFuncCall {
	f.mutex.Lock()
	history := make([]CodeIntelAPIReferencesFuncCall, len(f.history))
	copy(history, f.history)
	f.mutex.Unlock()

	return history
}

// CodeIntelAPIReferencesFuncCall is an object that describes an invocation
// of method References on an instance of MockCodeIntelAPI.
type CodeIntelAPIReferencesFuncCall struct {
	// Arg0 is the value of the 1st argument passed to this method
	// invocation.
	Arg0 context.Context
	// Arg1 is the value of the 2nd argument passed to this method
	// invocation.
	Arg1 int
	// Arg2 is the value of the 3rd argument passed to this method
	// invocation.
	Arg2 string
	// Arg3 is the value of the 4th argument passed to this method
	// invocation.
	Arg3 int
	// Arg4 is the value of the 5th argument passed to this method
	// invocation.
	Arg4 api.Cursor
	// Result0 is the value of the 1st result returned from this method
	// invocation.
	Result0 []api.ResolvedLocation
	// Result1 is the value of the 2nd result returned from this method
	// invocation.
	Result1 api.Cursor
	// Result2 is the value of the 3rd result returned from this method
	// invocation.
	Result2 bool
	// Result3 is the value of the 4th result returned from this method
	// invocation.
	Result3 error
}

// Args returns an interface slice containing the arguments of this
// invocation.
func (c CodeIntelAPIReferencesFuncCall) Args() []interface{} {
	return []interface{}{c.Arg0, c.Arg1, c.Arg2, c.Arg3, c.Arg4}
}

// Results returns an interface slice containing the results of this
// invocation.
func (c CodeIntelAPIReferencesFuncCall) Results() []interface{} {
	return []interface{}{c.Result0, c.Result1, c.Result2, c.Result3}
}
