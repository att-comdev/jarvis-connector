package controllers_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/att-comdev/jarvis-connector/cmd/connector/controllers"
	"github.com/att-comdev/jarvis-connector/services"
	"github.com/att-comdev/jarvis-connector/types"
)

func TestCheckerControllerImpl_PostChecker(t *testing.T) {
	// Arrange
	repo := "myrepo"
	prefix := "jarvis"

	expected := types.CheckerInput{
		UUID:        "This value cannot be predicted with a test, when checker is being created",
		Name:        prefix,
		Description: "check source code formatting.",
		URL:         "",
		Repository:  repo,
		Status:      "ENABLED",
		Blocking:    []string{},
		Query:       "status:open",
	}

	serverMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			out := types.CheckerInfo{
				UUID:        "This value cannot be predicted with a test, when checker is being created\"",
				Name:        prefix,
				Description: "check source code formatting.",
				URL:         "",
				Repository:  repo,
				Status:      "ENABLED",
				Blocking:    []string{},
				Query:       "status:open",
			}
			body, err := json.Marshal(&out)
			body = append([]byte(")]}'"), body...)
			return body, err
		},
		getPathFn:     nil,
		getFn:         nil,
		initFn:        nil,
		getURLFn:      nil,
		getRepoRootFn: nil,
	}

	services.GerritServer = serverMock

	// Act
	actual, err := controllers.Checker.PostChecker(repo, prefix, false)

	// Assert
	if err != nil {
		t.Errorf("Received error from TestPostChecker_shouldSendInputToGerritServer function: %v", err)
	}
	if actual.Name != expected.Name {
		t.Errorf("actual.Name: (%v), does not match expected.Name: (%v)", actual.Name, expected.Name)
	}
}

func TestCheckerControllerImpl_ListCheckers(t *testing.T) {
	// Arrange
	serverMock := serverServiceMock{
		postPathFn: nil,
		getPathFn: func(pathing string, headers []types.Header) ([]byte, error) {
			out := []*types.CheckerInfo{{
				UUID:        "jarvis:UUID",
				Name:        "Name",
				Description: "Description",
				URL:         "gerrit.jarvis.local",
				Repository:  "Repository",
				Status:      "Status",
				Blocking:    nil,
				Query:       "Query",
				Created:     types.Timestamp(time.Now()),
				Updated:     types.Timestamp(time.Now()),
			}}
			body, err := json.Marshal(&out)
			body = append([]byte(")]}'"), body...)
			return body, err
		},
		getFn:         nil,
		initFn:        nil,
		getURLFn:      nil,
		getRepoRootFn: nil,
	}
	checkerMock := checkerServiceMock{
		pendingChecksBySchemeFn: nil,
		executeCheckFn:          nil,
		checkerPrefixFn: func(scheme string) (string, bool) {
			return "", true
		},
	}

	services.GerritServer = serverMock
	services.GerritChecker = checkerMock

	// Act
	err := controllers.Checker.ListCheckers()

	// Assert
	if err != nil {
		t.Errorf("resulting error expected to be nil, received: %v", err)
	}
}
