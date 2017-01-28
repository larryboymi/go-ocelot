package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ocelotconsulting/go-ocelot/mocks"
)

var apiUnderTest *http.ServeMux
var respRec *httptest.ResponseRecorder
var req *http.Request
var err error

func setup(t *testing.T) {
	ctrl := gomock.NewController(t)
	//mux router with added question routes
	apiUnderTest = New(mocks.NewMockRepository(ctrl)).Mux()

	//The response recorder used to record HTTP responses
	respRec = httptest.NewRecorder()
}

func TestMuxNonExistingRoute404(t *testing.T) {
	setup(t)
	//Testing get of non existent route
	req, err = http.NewRequest("GET", "/api/v1/badroute", nil)
	if err != nil {
		t.Fatal("Creating 'GET /api/v1/badroute' request failed!")
	}

	apiUnderTest.ServeHTTP(respRec, req)

	if respRec.Code != http.StatusNotFound {
		t.Fatal("Server error: Returned ", respRec.Code, " instead of ", http.StatusNotFound)
	}
}
