package types

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"
)

var jsonPrefix = []byte(")]}'")

type CheckerInput struct {
	UUID        string   `json:"uuid"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	URL         string   `json:"url"`
	Repository  string   `json:"repository"`
	Status      string   `json:"status"`
	Blocking    []string `json:"blocking"`
	Query       string   `json:"query"`
}

// Gerrit doesn't use the format with "T" in the middle, so must
// define a custom serializer.

const timeLayout = "2006-01-02 15:04:05.000000000"

type Timestamp time.Time

func (ts *Timestamp) String() string {
	return ((time.Time)(*ts)).String()
}

var _ = (json.Marshaler)((*Timestamp)(nil))

func (ts *Timestamp) MarshalJSON() ([]byte, error) {
	t := (*time.Time)(ts)
	return []byte("\"" + t.Format(timeLayout) + "\""), nil
}

var _ = (json.Unmarshaler)((*Timestamp)(nil))

func (ts *Timestamp) UnmarshalJSON(b []byte) error {
	b = bytes.TrimPrefix(b, []byte{'"'})
	b = bytes.TrimSuffix(b, []byte{'"'})
	t, err := time.Parse(timeLayout, string(b))
	if err != nil {
		return err
	}
	*ts = Timestamp(t)
	return nil
}

type CheckerInfo struct {
	UUID        string `json:"uuid"`
	Name        string
	Description string
	URL         string `json:"url"`
	Repository  string `json:"repository"`
	Status      string
	Blocking    []string  `json:"blocking"`
	Query       string    `json:"query"`
	Created     Timestamp `json:"created"`
	Updated     Timestamp `json:"updated"`
}

func (info *CheckerInfo) String() string {
	out, err := json.Marshal(info)
	if err != nil {
		log.Printf("error marshaling CheckerInfo: %v", err)
	}
	return string(out)
}

// Unmarshal unmarshalls Gerrit JSON, stripping the security prefix.
func Unmarshal(content []byte, dest interface{}) error {
	if !bytes.HasPrefix(content, jsonPrefix) {
		if len(content) > 100 {
			content = content[:100]
		}
		bodyStr := string(content)

		return fmt.Errorf("prefix %q not found, got %s", jsonPrefix, bodyStr)
	}

	content = bytes.TrimPrefix(content, jsonPrefix)
	return json.Unmarshal(content, dest)
}

type PendingCheckInfo struct {
	State string
}

type CheckablePatchSetInfo struct {
	Repository   string
	ChangeNumber int `json:"change_number"`
	PatchSetID   int `json:"patch_set_id"`
}

func (in *CheckablePatchSetInfo) String() string {
	out, err := json.Marshal(in)
	if err != nil {
		log.Printf("error marshaling CheckablePatchSetInfo: %v", err)
	}
	return string(out)
}

type PendingChecksInfo struct {
	PatchSet      *CheckablePatchSetInfo       `json:"patch_set"`
	PendingChecks map[string]*PendingCheckInfo `json:"pending_checks"`
}

func (info *PendingCheckInfo) String() string {
	out, err := json.Marshal(info)
	if err != nil {
		log.Printf("error marshaling PendingCheckInfo: %v", err)
	}
	return string(out)
}

type CheckInput struct {
	CheckerUUID string     `json:"checker_uuid"`
	State       string     `json:"state"`
	Message     string     `json:"message"`
	URL         string     `json:"url"`
	Started     *Timestamp `json:"started"`
}

func (in *CheckInput) String() string {
	out, err := json.Marshal(in)
	if err != nil {
		log.Printf("error marshaling CheckInput: %v", err)
	}
	return string(out)
}

type CheckInfo struct {
	Repository    string    `json:"repository"`
	ChangeNumber  int       `json:"change_number"`
	PatchSetID    int       `json:"patch_set_id"`
	CheckerUUID   string    `json:"checker_uuid"`
	State         string    `json:"state"`
	Message       string    `json:"message"`
	Started       Timestamp `json:"started"`
	Finished      Timestamp `json:"finished"`
	Created       Timestamp `json:"created"`
	Updated       Timestamp `json:"updated"`
	CheckerName   string    `json:"checker_name"`
	CheckerStatus string    `json:"checker_status"`
	Blocking      []string  `json:"blocking"`
}

// TektonListenerPayload to be received by trigger
type TektonListenerPayload struct {
	RepoRoot       string `json:"repoRoot"`
	Project        string `json:"project"`
	ChangeNumber   string `json:"changeNumber"`
	PatchSetNumber int    `json:"patchSetNumber"`
	CheckerUUID    string `json:"checkerUUID"`
}

type Header struct {
	Key   string
	Value string
}

type PendingSubmitInfo struct {
	ID              string              `json:"id"`
	Project         string              `json:"project"`
	Branch          string              `json:"branch"`
	Hashtags        []string            `json:"hashtags"`
	ChangeID        string              `json:"change_id"`
	ChangeNumber    int                 `json:"_number"`
	Subject         string              `json:"subject"`
	Status          string              `json:"status"`
	Created         Timestamp           `json:"created"`
	Updated         Timestamp           `json:"updated"`
	SubmitType      string              `json:"submit_type"`
	Mergeable       bool                `json:"mergeable"`
	Subittable      bool                `json:"submittable"`
	CurrentRevision string              `json:"current_revision"`
	Revisions       map[string]Revision `json:"revisions"`
	Labels          map[string]Label    `json:"labels"`
	RevisionNumber  int
}

type Revision struct {
	Kind    string    `json:"kind"`
	Number  int       `json:"_number"`
	Created Timestamp `json:"created"`
	Ref     string    `json:"ref"`
}

type Label struct {
	Approved Approval `json:"approved"`
	Optional bool     `json:"optional"`
}

type Approval struct {
	AccountID int `json:"_account_id"`
}

type LockPayload struct {
	Label LabelPayload `json:"labels"`
}

type LabelPayload struct {
	JarvisLock string `json:"Jarvis-Lock"`
}

type TektonMergePayload struct {
	RepoRoot       string `json:"repoRoot"`
	Project        string `json:"project"`
	ChangeNumber   string `json:"changeNumber"`
	PatchSetNumber string `json:"patchSetNumber"`
	CheckerUUID    string `json:"checkerUUID"`
}
