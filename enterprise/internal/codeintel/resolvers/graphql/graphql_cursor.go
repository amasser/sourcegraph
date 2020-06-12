package resolvers

import (
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func encodeCursor(val *string) *graphqlutil.PageInfo {
	if val != nil {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(*val)))
	}

	return graphqlutil.HasNextPage(false)
}

func decodeCursor(val *string) (string, error) {
	if val == nil {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(*val)
	if err != nil {
		return "", err
	}

	return string(decoded), nil
}

func encodeIntCursor(val *int) *graphqlutil.PageInfo {
	var str string
	if val != nil {
		str = fmt.Sprintf("%d", *val)
	}

	return encodeCursor(&str)
}

func decodeIntCursor(val *string) (int, error) {
	cursor, err := decodeCursor(val)
	if err == nil || cursor == "" {
		return 0, err
	}

	return strconv.Atoi(string(cursor))
}
