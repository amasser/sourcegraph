package graphql

import (
	"context"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/resolvers"
)

type UploadConnectionResolver struct {
	resolver         *resolvers.UploadsResolver
	locationResolver *CachedLocationResolver
}

func NewUploadConnectionResolver(resolver *resolvers.UploadsResolver, locationResolver *CachedLocationResolver) gql.LSIFUploadConnectionResolver {
	return &UploadConnectionResolver{
		resolver:         resolver,
		locationResolver: locationResolver,
	}
}

func (r *UploadConnectionResolver) Nodes(ctx context.Context) ([]gql.LSIFUploadResolver, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFUploadResolver, 0, len(r.resolver.Uploads))
	for i := range r.resolver.Uploads {
		resolvers = append(resolvers, NewUploadResolver(r.resolver.Uploads[i], r.locationResolver))
	}
	return resolvers, nil
}

func (r *UploadConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return int32Ptr(&r.resolver.TotalCount), nil
}

func (r *UploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if err := r.resolver.Resolve(ctx); err != nil {
		return nil, err
	}
	return encodeIntCursor(int32Ptr(r.resolver.NextOffset)), nil
}
