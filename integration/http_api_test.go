//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/dariomba/screen-go/internal/adapters/openapi"
	"github.com/stretchr/testify/suite"
)

type APITestSuite struct {
	APIBaseSuite
}

func (s *APITestSuite) TestCreateJob_HappyPath() {
	resp, err := http.Post(
		s.Server.URL+"/v1/job",
		"application/json",
		strings.NewReader(`{"url":"https://github.com","format":"png"}`),
	)

	s.Require().NoError(err)
	s.Equal(http.StatusAccepted, resp.StatusCode)

	var body openapi.CreateJobResponse
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.NotEmpty(body.JobID)
	s.Equal(openapi.CreateJobResponseStatusPending, body.Status)
	s.NotEmpty(body.StatusURL)
}

func (s *APITestSuite) TestCreateJob_InvalidURL() {
	resp, err := http.Post(
		s.Server.URL+"/v1/job",
		"application/json",
		strings.NewReader(`{"url":"not-a-url"}`),
	)

	s.Require().NoError(err)
	s.Equal(http.StatusUnprocessableEntity, resp.StatusCode)

	var body openapi.ErrorResponse
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&body))
	s.NotEmpty(body.Error)
}

func (s *APITestSuite) TestCreateJob_MissingURL() {
	resp, err := http.Post(
		s.Server.URL+"/v1/job",
		"application/json",
		strings.NewReader(`{}`),
	)

	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *APITestSuite) TestCreateJob_InvalidFormat() {
	resp, err := http.Post(
		s.Server.URL+"/v1/job",
		"application/json",
		strings.NewReader(`{"url":"https://github.com","format":"gif"}`),
	)

	s.Require().NoError(err)
	s.Equal(http.StatusBadRequest, resp.StatusCode)
}

func (s *APITestSuite) TestGetJobStatus_NotFound() {
	resp, err := http.Get(s.Server.URL + "/v1/job/non-existent-id")

	s.Require().NoError(err)
	s.Equal(http.StatusNotFound, resp.StatusCode)
}

func (s *APITestSuite) TestGetJobStatus_Found() {
	// create job first
	resp, err := http.Post(
		s.Server.URL+"/v1/job",
		"application/json",
		strings.NewReader(`{"url":"https://github.com"}`),
	)
	s.Require().NoError(err)

	var created openapi.CreateJobResponse
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&created))

	// poll status
	resp, err = http.Get(s.Server.URL + "/v1/job/" + created.JobID)
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.StatusCode)

	var job openapi.Job
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&job))
	s.Equal(created.JobID, job.JobID)
}

func (s *APITestSuite) TestFullJobLifecycle() {
	// submit job
	resp, err := http.Post(
		s.Server.URL+"/v1/job",
		"application/json",
		strings.NewReader(`{"url":"https://github.com","format":"png"}`),
	)
	s.Require().NoError(err)
	s.Equal(http.StatusAccepted, resp.StatusCode)

	var created openapi.CreateJobResponse
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&created))

	s.Require().Eventually(func() bool {
		resp, err := http.Get(s.Server.URL + "/v1/job/" + created.JobID)
		if err != nil {
			return false
		}
		var job openapi.Job
		if err := json.NewDecoder(resp.Body).Decode(&job); err != nil {
			return false
		}
		return job.Status == openapi.JobStatusDone
	}, 10*time.Second, 100*time.Millisecond)
}

func (s *APITestSuite) TestFullJobLifecycle_InvalidURL() {
	// submit a job with a URL that will fail at the processor level
	// (valid format but unreachable/non-existent host)
	resp, err := http.Post(
		s.Server.URL+"/v1/job",
		"application/json",
		strings.NewReader(`{"url":"http://this-host-does-not-exist-screen-go-test.invalid"}`),
	)
	s.Require().NoError(err)
	s.Equal(http.StatusAccepted, resp.StatusCode)

	var created openapi.CreateJobResponse
	s.Require().NoError(json.NewDecoder(resp.Body).Decode(&created))

	s.Require().Eventually(func() bool {
		resp, _ := http.Get(s.Server.URL + "/v1/job/" + created.JobID)
		var job openapi.Job
		json.NewDecoder(resp.Body).Decode(&job)
		return job.Status == openapi.JobStatusFailed
	}, 30*time.Second, 500*time.Millisecond)
}

func TestAPITestSuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}
