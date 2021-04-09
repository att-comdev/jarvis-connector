package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/att-comdev/jarvis-connector/types"
)

// checkerScheme is the scheme by which we are registered in the Gerrit server.
const (
	checkerScheme = "jarvis"
)

// errIrrelevant is a marker error value used for checks that don't apply for a change.
var (
	GerritChecker gerritCheckerService = &GerritCheckerServiceImpl{}
	errIrrelevant                      = errors.New("irrelevant")
)

type gerritCheckerService interface {
	PendingChecksByScheme(scheme string) ([]*types.PendingChecksInfo, error)
	ExecuteCheck(pc *types.PendingChecksInfo) error
	CheckerPrefix(uuid string) (string, bool)
}

type GerritCheckerServiceImpl struct{}

// Returns Checks that are pending execution and are associated with the scheme provided
func (g *GerritCheckerServiceImpl) PendingChecksByScheme(scheme string) ([]*types.PendingChecksInfo, error) {
	u := GerritServer.GetURL()

	// The trailing '/' handling is really annoying.
	u.Path = path.Join(u.Path, "a/plugins/checks/checks.pending/") + "/"

	q := "scheme:" + scheme
	u.RawQuery = "query=" + q

	content, err := GerritServer.Get(&u)

	if err != nil {
		return nil, err
	}

	var out []*types.PendingChecksInfo
	if err := types.Unmarshal(content, &out); err != nil {
		return nil, err
	}

	return out, nil
}

// ExecuteCheck executes the pending checks specified in the argument.
func (g *GerritCheckerServiceImpl) ExecuteCheck(pc *types.PendingChecksInfo) error {
	log.Println("checking", pc)

	repository := pc.PatchSet.Repository
	changeID := strconv.Itoa(pc.PatchSet.ChangeNumber)
	psID := pc.PatchSet.PatchSetID
	for uuid := range pc.PendingChecks {
		now := types.Timestamp(time.Now())
		checkInput := types.CheckInput{
			CheckerUUID: uuid,
			State:       StatusRunning.String(),
			Message:     "Jarvis about to submit job to tekton",
			Started:     &now,
		}
		log.Printf("posted %s", &checkInput)
		_, err := g.postCheck(changeID, psID, &checkInput)
		if err != nil {
			return err
		}

		var status StatusService
		msg := ""
		url := ""
		lang, ok := g.CheckerPrefix(uuid)

		if !ok {
			return fmt.Errorf("uuid %q had unknown prefix", uuid)
		}

		msgs, details, err := g.checkChange(uuid, repository, changeID, psID, lang)
		if err == errIrrelevant { //nolint
			status = StatusIrrelevant
		} else if err != nil {
			status = StatusFail
			log.Printf("failed in attempt to schedule checkChange(%s, %s, %d, %q): %v", uuid, changeID, psID, lang, err)
		} else if len(msgs) != 0 {
			status = StatusSuccessful
		} else {
			status = StatusFail
			log.Printf("message empty for checkChange(%s, %s, %d, %q): %v", uuid, changeID, psID, lang, err)
		}
		url = details
		msg = strings.Join(msgs, ", ")
		if len(msg) > 1000 {
			msg = msg[:995] + "..."
		}

		log.Printf("status %s for lang %s on %v", status, lang, pc.PatchSet)
		checkInput = types.CheckInput{
			CheckerUUID: uuid,
			State:       status.String(),
			Message:     msg,
			URL:         url,
			Started:     &types.Timestamp{},
		}
		log.Printf("posted %s", &checkInput)

		if _, err := g.postCheck(changeID, psID, &checkInput); err != nil {
			return err
		}
	}
	return nil
}

// checkerPrefix extracts the prefix to check for from a checker UUID.
func (g *GerritCheckerServiceImpl) CheckerPrefix(uuid string) (string, bool) {
	uuid = strings.TrimPrefix(uuid, checkerScheme+":")
	fields := strings.Split(uuid, "-")
	if len(fields) != 2 {
		return "", false
	}
	return fields[0], true
}

// PostCheck posts a single check result onto a change.
func (g *GerritCheckerServiceImpl) postCheck(changeID string, psID int, input *types.CheckInput) (*types.CheckInfo, error) { //nolint
	headers := []types.Header{{
		Key:   "Content-Type",
		Value: "application/json",
	}}
	body, err := json.Marshal(input)
	if err != nil {
		return nil, err
	}

	res, err := GerritServer.PostPath(fmt.Sprintf("a/changes/%s/revisions/%d/checks/", changeID, psID), headers, body)
	if err != nil {
		return nil, err
	}

	var out types.CheckInfo
	if err := types.Unmarshal(res, &out); err != nil {
		return nil, err
	}

	return &out, nil
}

// checkChange checks a (change, patchset) for correct formatting in the given prefix. It returns
// a list of complaints, or the errIrrelevant error if there is nothing to do.
func (g *GerritCheckerServiceImpl) checkChange(uuid string, repository string, changeID string, psID int, prefix string) ([]string, string, error) { //nolint
	log.Printf("checkChange(%s, %d, %q)", changeID, psID, prefix)

	headers := []types.Header{{
		Key:   "Content-Type",
		Value: "application/json",
	}, {
		Key:   "X-Jarvis",
		Value: "create",
	}}
	data := types.TektonListenerPayload{
		RepoRoot:       GerritServer.GetRepoRoot(),
		Project:        repository,
		ChangeNumber:   changeID,
		PatchSetNumber: psID,
		CheckerUUID:    uuid,
	}
	body, err := json.Marshal(data)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("body: %v", body)

	_, err = EventListenerServer.PostPath("", headers, body)
	if err != nil {
		log.Fatal(err)
	}

	var messages []string
	messages = append(messages, "Job has been submitted to tekton")
	var details = ""
	return messages, details, nil
}
