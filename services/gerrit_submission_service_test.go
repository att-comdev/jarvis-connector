package services_test

import (
	"encoding/json"
	"github.com/att-comdev/jarvis-connector/services"
	"github.com/att-comdev/jarvis-connector/types"
	"net/url"
	"testing"
)

func TestGerritSubmissionServiceImpl_PendingSubmit(t *testing.T) {
	// Arrange
	serverMock := serverServiceMock{}
	serverMock.getURLFn = func() url.URL {
		mockedURL, err := url.Parse("https://website.com")
		if err != nil {
			t.Errorf("Received error setting up TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
		}
		return *mockedURL
	}
	revisions := map[string]types.Revision{}

	labels := map[string]types.Label{}

	serverMock.getFn = func(u *url.URL) ([]byte, error) {
		// TODO verify Input
		out := []*types.PendingSubmitInfo{{
			ID:              "ID-1",
			Project:         "MyProject",
			Branch:          "Main",
			ChangeNumber:    10,
			Status:          "Status",
			Mergeable:       true,
			Subittable:      true,
			CurrentRevision: "I3657f951abfbb0eb7a959cf57951597fcbc27167",
			Revisions:       revisions,
			Labels:          labels,
		}}
		body, err := json.Marshal(&out)
		if err != nil {
			t.Errorf("Received error setting up TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
		}
		body = append([]byte(")]}'"), body...)
		return body, err
	}

	services.GerritServer = serverMock

	// Act
	result, err := services.GerritSubmitter.PendingSubmit()

	// Assert
	if err != nil {
		t.Errorf("Received error from TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("result length is expected to be 1")
	}
}

func TestGerritSubmissionServiceImpl_PendingSubmit2(t *testing.T) {
	// Arrange
	serverMock := serverServiceMock{}
	serverMock.getURLFn = func() url.URL {
		mockedURL, err := url.Parse("https://website.com")
		if err != nil {
			t.Errorf("Received error setting up TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
		}
		return *mockedURL
	}
	revisions := map[string]types.Revision{}

	labels := map[string]types.Label{}
	lockedLabel := map[string]types.Label{
		"Jarvis-Lock": {
			Approved: types.Approval{
				AccountID: 1000000,
			},
			Optional: false,
		},
	}

	serverMock.getFn = func(u *url.URL) ([]byte, error) {
		// TODO verify Input
		out := []*types.PendingSubmitInfo{{
			ID:              "ID-1",
			Project:         "MyProject",
			Branch:          "Main",
			ChangeNumber:    10,
			Status:          "Status",
			Mergeable:       true,
			Subittable:      true,
			CurrentRevision: "I3657f951abfbb0eb7a959cf57951597fcbc27167",
			Revisions:       revisions,
			Labels:          labels,
		}, {
			ID:              "ID-2",
			Project:         "MyProject",
			Branch:          "Main",
			ChangeNumber:    11,
			Status:          "Status",
			Mergeable:       false,
			Subittable:      true,
			CurrentRevision: "I750130646e7fd31c1c8c7f0db0843b27c66ea2f2",
			Revisions:       revisions,
			Labels:          labels,
		}, {
			ID:              "ID-3",
			Project:         "MyProject",
			Branch:          "Main",
			ChangeNumber:    12,
			Status:          "Status",
			Mergeable:       true,
			Subittable:      false,
			CurrentRevision: "I3d1f704f9baa62901bd540aaa48698846f582782",
			Revisions:       revisions,
			Labels:          labels,
		}, {
			ID:              "ID-4",
			Project:         "MyProject",
			Branch:          "Main",
			ChangeNumber:    12,
			Status:          "Status",
			Mergeable:       true,
			Subittable:      false,
			CurrentRevision: "Iee1a1fc901063a131a688e9212e9ec3d433d141f",
			Revisions:       revisions,
			Labels:          lockedLabel,
		}}
		body, err := json.Marshal(&out)
		if err != nil {
			t.Errorf("Received error setting up TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
		}
		body = append([]byte(")]}'"), body...)
		return body, err
	}

	services.GerritServer = serverMock

	// Act
	result, err := services.GerritSubmitter.PendingSubmit()

	// Assert
	if err != nil {
		t.Errorf("Received error from TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("result length is expected to be 1")
	}
}

func TestGerritSubmissionServiceImpl_ExecuteSubmit(t *testing.T) {
	// Arrange
	u, _ := url.Parse("https://website.com")
	gerritServerMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			return []byte{}, nil
		},
		getPathFn: func(pathing string, headers []types.Header) ([]byte, error) {
			obj := []*types.CheckerInfo{{
				UUID:        "",
				Name:        "",
				Description: "",
				URL:         "",
				Repository:  "",
				Status:      "",
				Blocking:    nil,
				Query:       "",
				Created:     types.Timestamp{},
				Updated:     types.Timestamp{},
			}}
			body, err := json.Marshal(&obj)
			if err != nil {
				t.Errorf("Received error setting up TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
			}
			body = append([]byte(")]}'"), body...)

			return body, nil
		},
		getFn:    nil,
		initFn:   nil,
		getURLFn: func() url.URL {
			return *u
		},
		getRepoRootFn: func() string {
			return "https://website.com/"
		},
	}

	eventListenerServerMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			return []byte{}, nil
		},
		getPathFn:     nil,
		getFn:         nil,
		initFn:        nil,
		getURLFn:      nil,
		getRepoRootFn: nil,
	}

	revisions := map[string]types.Revision{
		"I3657f951abfbb0eb7a959cf57951597fcbc27167": {
			Number: 1,
		},
	}

	testingPatchset := &types.PendingSubmitInfo{
		ID:              "ID-1",
		Project:         "MyProject",
		Branch:          "Main",
		ChangeNumber:    10,
		ChangeID:        "realisticValue",
		Status:          "Status",
		Mergeable:       true,
		Subittable:      true,
		CurrentRevision: "I3657f951abfbb0eb7a959cf57951597fcbc27167",
		Revisions:       revisions,
		Labels:          nil,
	}

	services.GerritServer = gerritServerMock
	services.EventListenerServer = eventListenerServerMock

	// Act
	err := services.GerritSubmitter.ExecuteSubmit(testingPatchset)

	// Assert
	if err != nil {
		t.Errorf("resulting error expected to be nil, received: %v", err)
	}
}

func TestGerritSubmissionServiceImpl_PostLock(t *testing.T) {
	// Arrange
	u, _ := url.Parse("https://website.com")
	serverMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			// TODO verify Input
			return []byte(")]}'"), nil
		},
		getPathFn: nil,
		getFn:     nil,
		initFn:    nil,
		getURLFn: func() url.URL {
			return *u
		},
		getRepoRootFn: nil,
	}

	revisions := map[string]types.Revision{
		"I3657f951abfbb0eb7a959cf57951597fcbc27167": {
			Number: 1,
		},
	}

	testingPatchset := &types.PendingSubmitInfo{
		ID:              "ID-1",
		Project:         "MyProject",
		Branch:          "Main",
		ChangeNumber:    10,
		ChangeID:        "realisticValue",
		Status:          "Status",
		Mergeable:       true,
		Subittable:      true,
		CurrentRevision: "I3657f951abfbb0eb7a959cf57951597fcbc27167",
		Revisions:       revisions,
		Labels:          nil,
	}

	services.GerritServer = serverMock

	// Act
	err := services.GerritSubmitter.PostLock(testingPatchset)

	// Assert
	if err != nil {
		t.Errorf("resulting error expected to be nil, received: %v", err)
	}
}

func TestGerritSubmissionServiceImpl_CallMergePipeline(t *testing.T) {
	// Arrange
	gerritServerMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			return []byte{}, nil
		},
		getPathFn: func(pathing string, headers []types.Header) ([]byte, error) {
			obj := []*types.CheckerInfo{{
				UUID:        "",
				Name:        "",
				Description: "",
				URL:         "",
				Repository:  "",
				Status:      "",
				Blocking:    nil,
				Query:       "",
				Created:     types.Timestamp{},
				Updated:     types.Timestamp{},
			}}
			body, err := json.Marshal(&obj)
			if err != nil {
				t.Errorf("Received error setting up TestGerritSubmissionServiceImpl_PendingSubmit function: %v", err)
			}
			body = append([]byte(")]}'"), body...)

			return body, nil
		},
		getFn:    nil,
		initFn:   nil,
		getURLFn: nil,
		getRepoRootFn: func() string {
			return "https://website.com/"
		},
	}

	eventListenerServerMock := serverServiceMock{
		postPathFn: func(pathing string, headers []types.Header, content []byte) ([]byte, error) {
			return []byte{}, nil
		},
		getPathFn:     nil,
		getFn:         nil,
		initFn:        nil,
		getURLFn:      nil,
		getRepoRootFn: nil,
	}

	revisions := map[string]types.Revision{
		"I3657f951abfbb0eb7a959cf57951597fcbc27167": {
			Number: 1,
		},
	}

	testingPatchset := &types.PendingSubmitInfo{
		ID:              "ID-1",
		Project:         "MyProject",
		Branch:          "Main",
		ChangeNumber:    10,
		ChangeID:        "realisticValue",
		Status:          "Status",
		Mergeable:       true,
		Subittable:      true,
		CurrentRevision: "I3657f951abfbb0eb7a959cf57951597fcbc27167",
		Revisions:       revisions,
		Labels:          nil,
	}

	services.GerritServer = gerritServerMock
	services.EventListenerServer = eventListenerServerMock
	// Act
	err := services.GerritSubmitter.CallMergePipeline(testingPatchset)

	// Assert
	if err != nil {
		t.Errorf("resulting error expected to be nil, received: %v", err)
	}
}
