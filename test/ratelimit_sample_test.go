package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"snapp/db"
	"snapp/handlers"
	"snapp/limiters"

	"bou.ke/monkey"

	"github.com/stretchr/testify/assert"
)

func mockNowSample(t time.Time) {
	monkey.Patch(time.Now, func() time.Time {
		return t
	})
}

func testResponseByIpSample(t *testing.T,
	uri string,
	handler http.Handler,
	remoteAddr string,
	responseCode int,
	expectedBody string) {
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	req.RemoteAddr = remoteAddr

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)
	//fmt.Println("responseCode", responseCode, rr.Code)
	//fmt.Println("responseCode", rr.Header().Get("Content-Type"))
	assert.Equal(t, responseCode, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	//fmt.Println("responseCode", rr.Body.String())
	assert.JSONEq(t, expectedBody, rr.Body.String())
}

func testOKByIpSample(t *testing.T, handler http.Handler, remoteAddr string) {
	testResponseByIpSample(
		t,
		"/",
		handler,
		remoteAddr,
		http.StatusOK,
		`{"ok": true}`)
}

func testTooManyRequestsByIpSample(t *testing.T, handler http.Handler, remoteAddr string) {
	testResponseByIpSample(
		t,
		"/",
		handler,
		remoteAddr,
		http.StatusTooManyRequests,
		`{"error": "too many requests"}`)
}

func TestRateLimitByIpSample(t *testing.T) {
	mockNowSample(time.Date(2016, 1, 1, 0, 0, 0, 0, time.UTC))

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Root)
	handler := limiters.ByIp(mux, 1, 3)
	begin := time.Now()
	testOKByIpSample(t, handler, "127.0.0.1:48765")
	fmt.Println("xxxx", time.Since(begin))
	testOKByIpSample(t, handler, "127.0.0.1:48765")
	fmt.Println("xxxx", time.Since(begin))
	testOKByIpSample(t, handler, "127.0.0.1:48765")
	fmt.Println("xxxx", time.Since(begin))

	testTooManyRequestsByIpSample(t, handler, "127.0.0.1:48765")
	fmt.Println("xxxx", time.Since(begin))
}

func testResponseByAppKeySample(t *testing.T,
	uri string,
	handler http.Handler,
	appKey string,
	responseCode int,
	expectedBody string) {
	req, _ := http.NewRequest(http.MethodGet, uri, nil)
	req.Header.Set("X-App-Key", appKey)

	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, responseCode, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	assert.JSONEq(t, expectedBody, rr.Body.String())
}

func testOKByAppKeySample(t *testing.T, handler http.Handler, appKey string) {
	testResponseByAppKeySample(
		t,
		"/",
		handler,
		appKey,
		http.StatusOK,
		`{"ok": true}`)
}

func testTooManyRequestsByAppKeySample(t *testing.T, handler http.Handler, appKey string) {
	testResponseByAppKeySample(
		t,
		"/",
		handler,
		appKey,
		http.StatusTooManyRequests,
		`{"error": "too many requests"}`)
}

func TestRateLimitByAppKeySample(t *testing.T) {
	db.GetConnection().Exec(`
		DROP TABLE IF EXISTS app_keys;
		CREATE TABLE app_keys
		(
		  id   BIGSERIAL PRIMARY KEY,
		  key  VARCHAR(255)
		);
		INSERT INTO app_keys
		VALUES
		  (1,  'sample_key_1'),
		  (2,  'sample_key_2');
	`)

	mockNowSample(time.Date(2017, 1, 1, 0, 0, 0, 0, time.UTC))

	mux := http.NewServeMux()
	mux.HandleFunc("/", handlers.Root)
	handler := limiters.ByAppKey(mux, 2, 3)

	testOKByAppKeySample(t, handler, "sample")
	testOKByAppKeySample(t, handler, "sample")
	testOKByAppKeySample(t, handler, "sample")

	// +2 Seconds
	mockNowSample(time.Date(2017, 1, 1, 0, 0, 2, 0, time.UTC))

	testOKByAppKeySample(t, handler, "sample")
	testOKByAppKeySample(t, handler, "sample")
	testOKByAppKeySample(t, handler, "sample")
	testOKByAppKeySample(t, handler, "sample")
}
