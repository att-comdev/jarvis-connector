package controllers

import (
	"crypto/sha1" //nolint
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/att-comdev/jarvis-connector/services"
	"github.com/att-comdev/jarvis-connector/types"
)

const (
	checkerScheme = "jarvis"
)

var (
	Checker checkerController = &CheckerControllerImpl{}
)

type checkerController interface {
	PostChecker(repo, prefix string, update bool) (*types.CheckerInfo, error)
	ListCheckers() error
}

type CheckerControllerImpl struct{}

// PostChecker creates or changes a checker. It sets up a checker on the given repo, for the given prefix.
func (controller *CheckerControllerImpl) PostChecker(repo, prefix string, update bool) (*types.CheckerInfo, error) {
	hash := sha1.New() //nolint
	hash.Write([]byte(repo)) //nolint

	uuid := fmt.Sprintf("%s:%s-%x", checkerScheme, prefix, hash.Sum(nil))
	in := types.CheckerInput{
		UUID:        uuid,
		Name:        prefix,
		Description: "check source code formatting.",
		URL:         "",
		Repository:  repo,
		Status:      "ENABLED",
		Blocking:    []string{},
		Query:       "status:open",
	}

	body, err := json.Marshal(&in)
	if err != nil {
		return nil, err
	}

	path := "a/plugins/checks/checkers/"
	if update {
		path += uuid
	}
	headers := []types.Header{{
		Key:   "Content-Type",
		Value: "application/json",
	}}
	content, err := services.GerritServer.PostPath(path, headers, body)
	if err != nil {
		return nil, err
	}

	out := types.CheckerInfo{}
	if err := types.Unmarshal(content, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// ListCheckers returns all the checkers for our scheme.
func (controller *CheckerControllerImpl) ListCheckers() error {
	headers := []types.Header{{
		Key:   "Content-Type",
		Value: "application/json",
	}}
	c, err := services.GerritServer.GetPath("a/plugins/checks/checkers/", headers)
	if err != nil {
		log.Fatalf("ListCheckers: %v", err)
	}

	var out []*types.CheckerInfo
	if err := types.Unmarshal(c, &out); err != nil {
		return err
	}

	filtered := out[:0]
	for _, o := range out {
		if !strings.HasPrefix(o.UUID, checkerScheme+":") {
			continue
		}
		if _, ok := services.GerritChecker.CheckerPrefix(o.UUID); !ok {
			continue
		}

		filtered = append(filtered, o)
	}

	for _, obj := range filtered {
		marshalledJSON, err := json.Marshal(obj)
		if err != nil {
			log.Printf("error marshaling CheckerInfo: %v", err)
		}
		log.Printf("%v \n", string(marshalledJSON))
	}

	return nil
}
