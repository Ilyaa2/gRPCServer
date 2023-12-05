package util

import "context"

func GetReqIDFromContext(ctx context.Context) string {
	reqID, ok := ctx.Value("req-id").(string)
	if !ok {
		reqID = "none"
	}
	return reqID
}
