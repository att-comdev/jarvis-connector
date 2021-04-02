package controllers_test

import (
	"net/url"

	"github.com/att-comdev/jarvis-connector/services"
	"github.com/att-comdev/jarvis-connector/types"
)

type serverServiceMock struct {
	postPathFn    func(pathing string, headers []types.Header, content []byte) ([]byte, error)
	getPathFn     func(pathing string, headers []types.Header) ([]byte, error)
	getFn         func(u *url.URL) ([]byte, error)
	initFn        func(url url.URL, authenticator services.Authenticator, testPath string)
	getURLFn      func() url.URL
	getRepoRootFn func() string
}

func (s serverServiceMock) GetPath(pathing string, headers []types.Header) ([]byte, error) {
	return s.getPathFn(pathing, headers)
}

func (s serverServiceMock) PostPath(pathing string, headers []types.Header, content []byte) ([]byte, error) {
	return s.postPathFn(pathing, headers, content)
}

func (s serverServiceMock) Get(u *url.URL) ([]byte, error) {
	return s.getFn(u)
}

func (s serverServiceMock) Init(url url.URL, authenticator services.Authenticator, testPath string) {
	s.initFn(url, authenticator, testPath)
}

func (s serverServiceMock) GetURL() url.URL {
	return s.getURLFn()
}

func (s serverServiceMock) GetRepoRoot() string {
	return s.getRepoRootFn()
}

type checkerServiceMock struct {
	pendingChecksBySchemeFn func(scheme string) ([]*types.PendingChecksInfo, error)
	executeCheckFn          func(pc *types.PendingChecksInfo) error
	checkerPrefixFn         func(uuid string) (string, bool)
}

func (c checkerServiceMock) PendingChecksByScheme(scheme string) ([]*types.PendingChecksInfo, error) {
	return c.pendingChecksBySchemeFn(scheme)
}

func (c checkerServiceMock) ExecuteCheck(pc *types.PendingChecksInfo) error {
	return c.executeCheckFn(pc)
}

func (c checkerServiceMock) CheckerPrefix(uuid string) (string, bool) {
	return c.checkerPrefixFn(uuid)
}