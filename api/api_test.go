package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/ocelotconsulting/go-ocelot/mocks"
	"github.com/ocelotconsulting/go-ocelot/types"
)

var apiUnderTest *http.ServeMux
var respRec *httptest.ResponseRecorder
var req *http.Request
var err error

func setupRoutes() map[string]types.Route {
	routes := make(map[string]types.Route)
	routes["test"] = types.Route{
		ID:         "test",
		ProxiedURL: "test.ocelot.com",
		TargetPort: 8080,
	}
	return routes
}

func setup(t *testing.T) {
	ctrl := gomock.NewController(t)

	repoMock := mocks.NewMockRepository(ctrl)
	routes := setupRoutes()
	repoMock.EXPECT().Routes().Return(routes).AnyTimes()
	//mux router with added question routes
	apiUnderTest = New(repoMock).Mux()

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

func TestMuxEcho(t *testing.T) {
	setup(t)
	//Testing get of non existent route
	req, err = http.NewRequest("GET", "/api/v1/echo", nil)
	if err != nil {
		t.Fatal("Creating 'GET /api/v1/echo' request failed!")
	}

	apiUnderTest.ServeHTTP(respRec, req)

	if respRec.Code != http.StatusOK {
		t.Fatal("Server error: Returned ", respRec.Code, " instead of ", http.StatusOK)
	}
}

func TestMuxGetRoutes(t *testing.T) {
	setup(t)
	//Testing get of non existent route
	req, err = http.NewRequest("GET", "/api/v1/routes", nil)
	if err != nil {
		t.Fatal("Creating 'GET /api/v1/routes' request failed!")
	}

	apiUnderTest.ServeHTTP(respRec, req)

	if respRec.Code != http.StatusMovedPermanently {
		t.Fatal("Server error: Returned ", respRec.Code, " instead of ", http.StatusMovedPermanently)
	}

	req, err = http.NewRequest("GET", respRec.HeaderMap["Location"][0], nil)
	respRec = httptest.NewRecorder()

	apiUnderTest.ServeHTTP(respRec, req)

	var respRoute map[string]types.Route
	json.NewDecoder(respRec.Body).Decode(&respRoute)

	if respRec.Code != http.StatusOK {
		t.Fatal("Server error: Returned ", respRec.Code, " instead of ", http.StatusOK)
	}

	if _, ok := respRoute["test"]; !ok {
		t.Fatal("Test route not present in routes.")
	}

}

func TestMuxGetTestRoute(t *testing.T) {
	setup(t)
	//Testing get of non existent route
	req, err = http.NewRequest("GET", "/api/v1/routes/test", nil)
	if err != nil {
		t.Fatal("Creating 'GET /api/v1/routes/test' request failed!")
	}

	apiUnderTest.ServeHTTP(respRec, req)

	if respRec.Code != http.StatusOK {
		t.Fatal("Server error: Returned ", respRec.Code, " instead of ", http.StatusOK)
	}

	var respRoute types.Route
	json.NewDecoder(respRec.Body).Decode(&respRoute)

	if respRoute.ID != "test" {
		t.Fatal("Server error: Returned invalid route ", respRoute.ID, " instead of ", "test")
	}

}
