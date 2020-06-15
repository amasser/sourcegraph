package graphql

import (
	"context"
	"testing"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	resolvermocks "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers/mocks"
)

func TestDefinitions(t *testing.T) {
	mockResolver := resolvermocks.NewMockQueryResolver()
	resolver := NewQueryResolver(mockResolver, NewCachedLocationResolver())

	args := &gql.LSIFQueryPositionArgs{Line: 10, Character: 15}
	if _, err := resolver.Definitions(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockResolver.DefinitionsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockResolver.DefinitionsFunc.History()))
	}
	if val := mockResolver.DefinitionsFunc.History()[0].Arg1; val != 10 {
		t.Fatalf("unexpected line. want=%d have=%d", 10, val)
	}
	if val := mockResolver.DefinitionsFunc.History()[0].Arg2; val != 15 {
		t.Fatalf("unexpected character. want=%d have=%d", 15, val)
	}
}

func TestReferences(t *testing.T) {
	mockResolver := resolvermocks.NewMockQueryResolver()
	resolver := NewQueryResolver(mockResolver, NewCachedLocationResolver())

	offset := int32(25)
	cursor := "" // TODO

	args := &gql.LSIFPagedQueryPositionArgs{
		LSIFQueryPositionArgs: gql.LSIFQueryPositionArgs{
			Line:      10,
			Character: 15,
		},
		ConnectionArgs: graphqlutil.ConnectionArgs{First: &offset},
		After:          &cursor,
	}

	if _, err := resolver.References(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	// TODO - assert calls
}

func TestHover(t *testing.T) {
	mockResolver := resolvermocks.NewMockQueryResolver()
	resolver := NewQueryResolver(mockResolver, NewCachedLocationResolver())

	args := &gql.LSIFQueryPositionArgs{Line: 10, Character: 15}
	if _, err := resolver.Hover(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockResolver.HoverFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockResolver.HoverFunc.History()))
	}
	if val := mockResolver.HoverFunc.History()[0].Arg1; val != 10 {
		t.Fatalf("unexpected line. want=%d have=%d", 10, val)
	}
	if val := mockResolver.HoverFunc.History()[0].Arg2; val != 15 {
		t.Fatalf("unexpected character. want=%d have=%d", 15, val)
	}
}

func TestDiagnostics(t *testing.T) {
	mockResolver := resolvermocks.NewMockQueryResolver()
	resolver := NewQueryResolver(mockResolver, NewCachedLocationResolver())

	offset := int32(25)
	args := &gql.LSIFDiagnosticsArgs{
		ConnectionArgs: graphqlutil.ConnectionArgs{First: &offset},
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockResolver.DiagnosticsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockResolver.DiagnosticsFunc.History()))
	}
	if val := mockResolver.DiagnosticsFunc.History()[0].Arg1; val != 25 {
		t.Fatalf("unexpected limit. want=%d have=%d", 25, val)
	}
}

func TestDiagnosticsDefaultLimit(t *testing.T) {
	mockResolver := resolvermocks.NewMockQueryResolver()
	resolver := NewQueryResolver(mockResolver, NewCachedLocationResolver())

	args := &gql.LSIFDiagnosticsArgs{
		ConnectionArgs: graphqlutil.ConnectionArgs{},
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}

	if len(mockResolver.DiagnosticsFunc.History()) != 1 {
		t.Fatalf("unexpected call count. want=%d have=%d", 1, len(mockResolver.DiagnosticsFunc.History()))
	}
	if val := mockResolver.DiagnosticsFunc.History()[0].Arg1; val != DefaultDiagnosticsPageSize {
		t.Fatalf("unexpected limit. want=%d have=%d", DefaultDiagnosticsPageSize, val)
	}
}

func TestDiagnosticsDefaultIllegalLimit(t *testing.T) {
	mockResolver := resolvermocks.NewMockQueryResolver()
	resolver := NewQueryResolver(mockResolver, NewCachedLocationResolver())

	offset := int32(-1)
	args := &gql.LSIFDiagnosticsArgs{
		ConnectionArgs: graphqlutil.ConnectionArgs{First: &offset},
	}

	if _, err := resolver.Diagnostics(context.Background(), args); err != ErrIllegalLimit {
		t.Fatalf("unexpected error. want=%q have=%q", ErrIllegalLimit, err)
	}
}
