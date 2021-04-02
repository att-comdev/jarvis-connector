package services_test

import (
	"encoding/json"
	"github.com/att-comdev/jarvis-connector/services"
	"github.com/att-comdev/jarvis-connector/types"
	"net/url"
	"testing"
)

func TestGerritCheckerServiceImpl_PendingChecksByScheme(t *testing.T) {
	// Arrange
	gerritServerMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			return []byte{}, nil
		},
		getPathFn: nil,
		getFn: func(inputURL *url.URL) ([]byte, error) {
			obj := []*types.PendingChecksInfo{{
				PatchSet: &types.CheckablePatchSetInfo{
					Repository:   "myRepo",
					ChangeNumber: 1,
					PatchSetID:   1,
				},
				PendingChecks: map[string]*types.PendingCheckInfo{
					"jarvis": {
						State: services.UnsetString,
					},
				},
			}}
			body, err := json.Marshal(&obj)
			if err != nil {
				t.Errorf("Received error setting up TestGerritCheckerServiceImpl_PendingChecksByScheme function: %v", err)
			}
			body = append([]byte(")]}'"), body...)

			return body, nil
		},
		initFn:   nil,
		getURLFn: func() url.URL {
			u, err := url.Parse("https://website.com/")
			if err != nil {
				t.Errorf("Received error setting up TestGerritCheckerServiceImpl_PendingChecksByScheme function: %v", err)
			}
			return *u
		},
		getRepoRootFn: nil,
	}

	services.GerritServer = gerritServerMock

	// Act
	result, err := services.GerritChecker.PendingChecksByScheme("jarvis")

	// Assert
	if err != nil {
		t.Errorf("resulting error expected to be nil, received: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 result, got: %d", len(result))
	}
}

func TestGerritCheckerServiceImpl_ExecuteCheck(t *testing.T) {
	// Arrange
	gerritServerMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			obj := types.CheckInfo{
				Repository:    "myRepo",
				ChangeNumber:  1,
				PatchSetID:    1,
				CheckerUUID:   "",
				State:         "",
				Message:       "",
				Started:       types.Timestamp{},
				Finished:      types.Timestamp{},
				Created:       types.Timestamp{},
				Updated:       types.Timestamp{},
				CheckerName:   "",
				CheckerStatus: "",
				Blocking:      nil,
			}
			body, err := json.Marshal(&obj)
			if err != nil {
				t.Errorf("Received error setting up TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
			}
			body = append([]byte(")]}'"), body...)
			return body, err
		},
		getPathFn: nil,
		getFn:    nil,
		initFn:   nil,
		getURLFn: nil,
		getRepoRootFn: func() string {
			return "https://website.com/"
		},
	}

	eventListenerServerMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			// TODO validate input
			return []byte{}, nil
		},
		getPathFn:     nil,
		getFn:         nil,
		initFn:        nil,
		getURLFn:      nil,
		getRepoRootFn: nil,
	}

	pc := types.PendingChecksInfo{
		PatchSet: &types.CheckablePatchSetInfo{
			Repository:   "myRepo",
			ChangeNumber: 1,
			PatchSetID:   1,
		},
		PendingChecks: map[string]*types.PendingCheckInfo{
			"jarvis:jarvispipeline-061bc62acb425af5bc8a4689221eed5781831ecc": {
				State: services.StatusIrrelevant.StatusString,
			},
		},
	}
	services.GerritServer = gerritServerMock
	services.EventListenerServer = eventListenerServerMock

	// Act
	err := services.GerritChecker.ExecuteCheck(&pc)

	// Assert
	if err != nil {
		t.Errorf("resulting error expected to be nil, received: %v", err)
	}

}

func TestGerritCheckerServiceImpl_CheckerPrefix(t *testing.T) {
	// Arrange
	testData := []struct{
		UUID string
		expected string
		result string
	} {
		{
			UUID: "jarvis:jarvispipeline-02c832b7ceb8bbe829ac5adfaf547d9675e9edfd",
			expected: "jarvispipeline",
		}, {
			UUID: "jarvis:jarvispipeline-061bc62acb425af5bc8a4689221eed5781831ecc",
			expected: "jarvispipeline",
		}, {
			UUID: "jarvis:jarvispipeline-2dfc9f62a11ade0762e86c37180be489463e8440",
			expected: "jarvispipeline",
		}, {
			UUID: "7f1c982e835a68959859b5d3da2b8e4b3af30b31",
			expected: "",
		}, {
			UUID: "",
			expected: "",
		},
	}
	// Act
	for _, test := range testData {
		test.result, _ = services.GerritChecker.CheckerPrefix(test.UUID)
		if test.result != test.expected {
			t.Errorf(
				"Test input: %s does not produce expected result. expected: %s result: %s",
				test.UUID, test.expected, test.result)
		}
	}
}