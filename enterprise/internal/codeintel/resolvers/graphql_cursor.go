package resolvers

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

// TODO - should put the b64 stuff here
func encodeCursor(val string) *graphqlutil.PageInfo {
	if val != "" {
		return graphqlutil.NextPageCursor(val)
	}

	return graphqlutil.HasNextPage(false)
}

func encodeIntCursor(val *int) *graphqlutil.PageInfo {
	if val != nil {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", val))))
	}

	return graphqlutil.HasNextPage(false)
}

func decodeIntCursor(val *string) (int, error) {
	if val == nil {
		return 0, nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return 0, err
	}

	v, _ := strconv.Atoi(string(decoded))
	return v, nil
}
