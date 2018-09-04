package platform

import (
	"net/http"
	"time"
)

type RequestStats struct {
	QueueingTime   time.Duration
	ProcessingTime time.Duration
	TotalTime      time.Duration
}

type RequestHeaders struct {
	ShopId int `json:"shop_id"`
}

type ResponseHeaders struct {
	HttpStatus int
}

type HttpRequest struct {
	httpReq  *http.Request
	httpResp http.ResponseWriter

	ResponseHeaders
	RequestStats
	RequestHeaders
}
