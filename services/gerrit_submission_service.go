package services

import (
	"encoding/json"
	"fmt"
	"log"
	"path"
	"strconv"

	"github.com/att-comdev/jarvis-connector/types"
)

var (
	GerritSubmitter gerritSubmissionService = &GerritSubmissionServiceImpl{}
)

type gerritSubmissionService interface {
	PendingSubmit() ([]*types.PendingSubmitInfo, error)
	ExecuteSubmit(patchset *types.PendingSubmitInfo) error
	PostLock(patchset *types.PendingSubmitInfo) error
	CallMergePipeline(patchset *types.PendingSubmitInfo) error
}

type GerritSubmissionServiceImpl struct{}

// PendingSubmit queries and returns all gerrit changes that are pending submission by Jarvis
func (g *GerritSubmissionServiceImpl) PendingSubmit() ([]*types.PendingSubmitInfo, error) {
	u := GerritServer.GetURL()

	u.Path = path.Join(u.Path, "a/changes/") + "/"
	q := u.Query()

	q.Add("o", "CURRENT_REVISION")
	q.Add("o", "SUBMITTABLE")
	q.Add("o", "LABELS")
	q.Add("q", "status:open")
	u.RawQuery = q.Encode()

	content, err := GerritServer.Get(&u)
	if err != nil {
		return []*types.PendingSubmitInfo{}, err
	}

	var out []*types.PendingSubmitInfo
	if err := types.Unmarshal(content, &out); err != nil {
		return []*types.PendingSubmitInfo{}, err
	}

	var patchsets []*types.PendingSubmitInfo
	for _, obj := range out {
		// Ignore merge conflicts, patchsets without required labels, and patchsets currently being handled by Jarvis
		if obj.Mergeable && obj.Subittable && obj.Labels["Jarvis-Lock"].Approved.AccountID == 0 {
			patchsets = append(patchsets, obj)
		}
	}

	return patchsets, nil
}

// ExecuteSubmit locks the patchset and sends a request to the Jarvis-System Event listener to trigger the merge
// pipeline
func (g *GerritSubmissionServiceImpl) ExecuteSubmit(patchset *types.PendingSubmitInfo) error {
	if err := g.PostLock(patchset); err != nil {
		log.Printf("PostLock Error: %v", err)
		return err
	}

	if err := g.CallMergePipeline(patchset); err != nil {
		log.Printf("CallMergePipeline Error: %v", err)
		return err
	}

	return nil
}

// ExecuteSubmit locks the patchset by adding the 'Jarvis-Lock' label
func (g *GerritSubmissionServiceImpl) PostLock(patchset *types.PendingSubmitInfo) error {
	u := GerritServer.GetURL()
	u.Path = path.Join(u.Path, fmt.Sprintf(
		"a/changes/%s/revisions/%s/review",
		patchset.ChangeID,
		strconv.Itoa(patchset.Revisions[patchset.CurrentRevision].Number)), // The patchset number of current revision
	) + "/"
	lockPayload := types.LockPayload{
		Label: types.LabelPayload{
			JarvisLock: "+1",
		},
	}
	headers := []types.Header{{
		Key:   "Content-Type",
		Value: "application/json",
	}}

	body, err := json.Marshal(lockPayload)
	if err != nil {
		log.Printf("json.Marshal Error: %v", err)
	}
	_, err = GerritServer.PostPath(u.Path, headers, body)
	if err != nil {
		log.Printf("PostPath Path Error: %v", err)
	}

	return nil
}

// CallMergePipeline sends a request to the Jarvis-System Event listener to trigger the merge pipeline
func (g *GerritSubmissionServiceImpl) CallMergePipeline(patchset *types.PendingSubmitInfo) error {
	checkerUUID, err := g.getChecker(patchset.Project)
	if err != nil {
		log.Printf("error finding relevant checker UUID: %v", err)
	}

	data := types.TektonMergePayload{
		RepoRoot:       GerritServer.GetRepoRoot(),
		Project:        patchset.Project,
		ChangeNumber:   strconv.Itoa(patchset.ChangeNumber),
		PatchSetNumber: strconv.Itoa(patchset.Revisions[patchset.CurrentRevision].Number),
		CheckerUUID:    checkerUUID,
	}

	headers := []types.Header{{
		Key:   "Content-Type",
		Value: "application/json",
	}, {
		Key:   "X-Jarvis",
		Value: "merge",
	}}

	body, err := json.Marshal(data) //nolint
	_, err = EventListenerServer.PostPath("", headers, body)

	if err != nil {
		log.Printf("error sending request to Gerrit: %v", err)
	}

	return nil
}

// getChecker returns the checker UUID associated with a given repository
// Warning: This method assumes only one checker exists per repository.
func (g *GerritSubmissionServiceImpl) getChecker(repository string) (string, error) {
	headers := []types.Header{{
		Key:   "Content-Type",
		Value: "application/json",
	}}

	content, err := GerritServer.GetPath("a/plugins/checks/checkers/", headers)

	if err != nil {
		log.Printf("getCheker Error: %v", err)
	}
	var out []*types.CheckerInfo
	if err := types.Unmarshal(content, &out); err != nil {
		return "", err
	}

	for i := 0; i < len(out); i++ {
		if out[i].Repository == repository {
			return out[i].UUID, nil
		}
	}

	return "", nil
}
