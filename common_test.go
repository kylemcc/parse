package parse

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

type ctxT struct {
	ts1            *httptest.Server
	oldHost1       string
	oldHttpClient1 *http.Client

	ts2            *httptest.Server
	oldHost2       string
	oldHttpClient2 *http.Client
}

var ctx = ctxT{}

func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *httptest.Server) {
	ts1 := httptest.NewTLSServer(handler)
	ctx.ts1 = ts1

	_url1, err := url.Parse(ts1.URL)
	if err != nil {
		panic(err)
	}

	ctx.oldHost1 = apps["app_id"].parseHost
	ctx.oldHttpClient1 = apps["app_id"].httpClient
	apps["app_id"].parseHost = _url1.Host
	apps["app_id"].httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	ts2 := httptest.NewTLSServer(handler)
	ctx.ts2 = ts2

	_url2, err := url.Parse(ts2.URL)
	if err != nil {
		panic(err)
	}

	ctx.oldHost2 = apps["app_id_2"].parseHost
	ctx.oldHttpClient2 = apps["app_id_2"].httpClient
	apps["app_id_2"].parseHost = _url2.Host
	apps["app_id_2"].httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	return ts1, ts2
}

func teardownTestServer() {
	ctx.ts1.Close()
	ctx.ts2.Close()

	apps["app_id"].parseHost = ctx.oldHost1
	apps["app_id"].httpClient = ctx.oldHttpClient1
	apps["app_id_2"].parseHost = ctx.oldHost2
	apps["app_id_2"].httpClient = ctx.oldHttpClient2
}

func TestMain(m *testing.M) {
	Initialize("app_id", "rest_key", "master_key")
	ServerURL("https://api.parse.com/1/")

	Initialize("app_id_2", "rest_key_2", "master_key_2")
	ServerURL("https://api.parse.com/2/")
	os.Exit(m.Run())
}
