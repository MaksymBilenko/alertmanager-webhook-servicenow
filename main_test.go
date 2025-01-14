package main

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/alertmanager/template"
	"github.com/stretchr/testify/mock"
)

type MockedSnClient struct {
	mock.Mock
}

func (mock *MockedSnClient) CreateIncident(incidentParam Incident) (Incident, error) {
	args := mock.Called(incidentParam)
	return args.Get(0).(Incident), args.Error(1)
}

func (mock *MockedSnClient) GetIncidents(params map[string]string) ([]Incident, error) {
	args := mock.Called(params)
	return args.Get(0).([]Incident), args.Error(1)
}

func (mock *MockedSnClient) UpdateIncident(incidentParam Incident, sysID string) (Incident, error) {
	args := mock.Called(incidentParam, sysID)
	return args.Get(0).(Incident), args.Error(1)
}

func TestLoadSnClient_OK(t *testing.T) {
	loadConfig("config/servicenow_example.yml")
	_, err := loadSnClient()
	if err != nil {
		t.Fatal(err)
	}
}

func TestWebhookHandler_Firing_DoNotExists_OK(t *testing.T) {
	loadConfig("config/servicenow_example.yml")
	incidentUpdateFields = map[string]bool{}
	snClientMock := new(MockedSnClient)
	serviceNow = snClientMock
	snClientMock.On("GetIncidents", mock.Anything).Return([]Incident{}, nil)
	snClientMock.On("CreateIncident", mock.Anything).Run(func(args mock.Arguments) {
		incident := args.Get(0).(Incident)
		if len(incident) == 0 {
			t.Errorf("Wrong incident len: got %v, do not want %v", len(incident), 0)
		}
	}).Return(Incident{}, nil)
	snClientMock.On("UpdateIncident", mock.Anything, mock.Anything).Return(Incident{}, errors.New("Update should not be called"))

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test/alertmanager_firing.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	expected := `{"Status":200,"Message":"Success"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandler_Firing_Exists_Create_OK(t *testing.T) {
	loadConfig("config/servicenow_example.yml")
	snClientMock := new(MockedSnClient)
	serviceNow = snClientMock
	snClientMock.On("GetIncidents", mock.Anything).Return([]Incident{Incident{"state": "6", "number": "INC42", "sys_id": "42"}}, nil)
	snClientMock.On("CreateIncident", mock.Anything).Return(Incident{}, nil)
	snClientMock.On("UpdateIncident", mock.Anything, mock.Anything).Return(Incident{}, errors.New("Update should not be called"))

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test/alertmanager_firing.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	expected := `{"Status":200,"Message":"Success"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandler_Firing_Exists_Update_OK(t *testing.T) {
	loadConfig("config/servicenow_example.yml")
	incidentUpdateFields = map[string]bool{
		"comments": true,
	}
	snClientMock := new(MockedSnClient)
	serviceNow = snClientMock
	snClientMock.On("GetIncidents", mock.Anything).Return([]Incident{Incident{"state": "1", "number": "INC42", "sys_id": "42"}}, nil)
	snClientMock.On("CreateIncident", mock.Anything).Return(Incident{}, errors.New("Create should not be called"))
	snClientMock.On("UpdateIncident", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		incident := args.Get(0).(Incident)
		if len(incident) != 1 {
			t.Errorf("Wrong incident len: got %v, want %v", len(incident), 1)
		}
	}).Return(Incident{}, nil)

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test/alertmanager_firing.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	expected := `{"Status":200,"Message":"Success"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandler_Resolved_DoNotExists_OK(t *testing.T) {
	loadConfig("config/servicenow_example.yml")
	snClientMock := new(MockedSnClient)
	serviceNow = snClientMock
	snClientMock.On("GetIncidents", mock.Anything).Return([]Incident{}, nil)
	snClientMock.On("CreateIncident", mock.Anything).Return(Incident{}, errors.New("Create should not be called"))
	snClientMock.On("UpdateIncident", mock.Anything, mock.Anything).Return(Incident{}, errors.New("Update should not be called"))

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test/alertmanager_resolved.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	expected := `{"Status":200,"Message":"Success"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandler_Resolved_Exists_OK(t *testing.T) {
	loadConfig("config/servicenow_example.yml")
	snClientMock := new(MockedSnClient)
	serviceNow = snClientMock
	snClientMock.On("GetIncidents", mock.Anything).Return([]Incident{Incident{"state": "7", "number": "INC42", "sys_id": "42"}}, nil)
	snClientMock.On("CreateIncident", mock.Anything).Return(Incident{}, errors.New("Create should not be called"))
	snClientMock.On("UpdateIncident", mock.Anything, mock.Anything).Return(Incident{}, nil)

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test/alertmanager_resolved.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusOK)
	}

	expected := `{"Status":200,"Message":"Success"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandler_BadRequest(t *testing.T) {
	loadConfig("config/servicenow_example.yml")

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", nil)

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusBadRequest)
	}

	expected := `{"Status":400,"Message":"EOF"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestWebhookHandler_InternalServerError(t *testing.T) {
	loadConfig("config/servicenow_example.yml")
	snClientMock := new(MockedSnClient)
	serviceNow = snClientMock
	snClientMock.On("GetIncidents", mock.Anything).Return([]Incident{}, nil)
	snClientMock.On("CreateIncident", mock.Anything).Return(Incident{}, errors.New("Error"))

	// Load a simple example of a body coming from AlertManager
	data, err := ioutil.ReadFile("test/alertmanager_firing.json")
	if err != nil {
		t.Fatal(err)
	}

	// Create a request to pass to the handler
	req := httptest.NewRequest("GET", "/webhook", bytes.NewReader(data))

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(webhook)

	// Test the handler with the request and record the result
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Wrong status code: got %v, want %v", status, http.StatusInternalServerError)
	}

	// Check the response body
	expected := `{"Status":500,"Message":"Error"}`
	if rr.Body.String() != expected {
		t.Errorf("Unexpected body: got %v, want %v", rr.Body.String(), expected)
	}
}

func TestApplyTemplate_emptyText(t *testing.T) {
	data := template.Data{}
	text := ""
	result, err := applyTemplate("name", text, data)
	if err != nil {
		t.Fatal(err)
	}
	expected := ""
	if result != expected {
		t.Errorf("Unexpected result: got %v, want %v", result, expected)
	}
}

func TestApplyTemplate_OK(t *testing.T) {
	data := template.Data{
		Status: "firing",
		CommonAnnotations: map[string]string{
			"error": "my error",
		},
	}
	text := "Status is {{.Status}} and error is {{.CommonAnnotations.error}}"
	result, err := applyTemplate("name", text, data)
	if err != nil {
		t.Fatal(err)
	}
	expected := "Status is firing and error is my error"
	if result != expected {
		t.Errorf("Unexpected result: got %v, want %v", result, expected)
	}
}

func TestApplyIncidentTemplate_Range(t *testing.T) {
	data := template.Data{
		CommonAnnotations: map[string]string{
			"error":   "a",
			"warning": "b",
		},
	}
	incident := Incident{
		"description": "{{ range $key, $val := .CommonAnnotations}}{{ $key }}:{{ $val }} {{end}}",
	}
	err := applyIncidentTemplate(incident, data)
	if err != nil {
		t.Fatal(err)
	}
	result := incident["description"]
	expected := "error:a warning:b "

	if result != expected {
		t.Errorf("Unexpected result: got %v, want %v", result, expected)
	}
}
