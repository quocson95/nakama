package server

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	grpcgw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
)

// recv request from another node
// for handler multi node nakama
func (s *ApiServer) RpcProxyNodeHttp(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()
	httpKey := queryParams.Get("http_key")
	w.Header().Set("content-type", "application/json")
	if httpKey == "" || httpKey != s.config.GetRuntime().HTTPKey {
		w.WriteHeader(http.StatusUnauthorized)
		_, err := w.Write(httpKeyInvalidBytes)
		if err != nil {
			s.logger.Debug("Error writing response to client", zap.Error(err))
		}
		return
	}

	start := time.Now()
	var success bool
	var recvBytes, sentBytes int
	var err error
	var id string

	// After this point the RPC will be captured in metrics.
	defer func() {
		s.metrics.ApiRpc(id, time.Since(start), int64(recvBytes), int64(sentBytes), !success)
	}()
	if r.Method != "POST" {
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	id = strings.ToLower("proxy-node")
	fn := s.runtime.Rpc(id)
	if fn == nil {
		// No function registered for this ID.
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		sentBytes, err = w.Write(rpcFunctionNotFoundBytes)
		if err != nil {
			s.logger.Debug("Error writing response to client", zap.Error(err))
		}
		return
	}
	var payload string
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		if err.Error() == "http: request body too large" {
			w.Header().Set("content-type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			sentBytes, err = w.Write(requestBodyTooLargeBytes)
			if err != nil {
				s.logger.Debug("Error writing response to client", zap.Error(err))
			}
			return
		}

		// Other error reading request body.
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		sentBytes, err = w.Write(internalServerErrorBytes)
		if err != nil {
			s.logger.Debug("Error writing response to client", zap.Error(err))
		}
		return
	}
	recvBytes = len(b)
	payload = string(b)
	queryParams.Del("http_key")
	uid := ""

	clientIP, clientPort := extractClientAddressFromRequest(s.logger, r)

	// Extract http headers
	headers := make(map[string][]string)
	for k, v := range r.Header {
		if k == "Grpc-Timeout" {
			continue
		}
		headers[k] = make([]string, 0, len(v))
		for _, h := range v {
			headers[k] = append(headers[k], h)
		}
	}

	var username string
	var vars map[string]string
	var expiry int64
	// Execute the function.
	result, fnErr, code := fn(r.Context(), headers, queryParams, uid, username, vars, expiry, "", clientIP, clientPort, "", payload)
	if fnErr != nil {
		response, _ := json.Marshal(map[string]interface{}{"error": fnErr, "message": fnErr.Error(), "code": code})
		w.Header().Set("content-type", "application/json")
		w.WriteHeader(grpcgw.HTTPStatusFromCode(code))
		sentBytes, err = w.Write(response)
		if err != nil {
			s.logger.Debug("Error writing response to client", zap.Error(err))
		}
		return
	}
	response := []byte(result)
	if contentType := r.Header["Content-Type"]; len(contentType) > 0 {
		// Assume the request input content type is the same as the expected response.
		w.Header().Set("content-type", contentType[0])
	} else {
		// Don't know payload content-type.
		w.Header().Set("content-type", "text/plain")
	}
	w.WriteHeader(http.StatusOK)
	sentBytes, err = w.Write(response)
	if err != nil {
		s.logger.Debug("Error writing response to client", zap.Error(err))
		return
	}
	success = true
}
