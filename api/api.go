package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/deta/deta-cli/runtime"
)

// injects X-Resource-Addr header from account and region
func (c *DetaClient) injectResourceHeader(headers map[string]string, account, region string) {
	resAddr := fmt.Sprintf("aws:%s:%s", account, region)
	encoded := base64.StdEncoding.EncodeToString([]byte(resAddr))
	headers["X-Resource-Addr"] = encoded
}

// DeployRequest deploy program request
type DeployRequest struct {
	ProgramID   string            `json:"pid"`
	Changes     map[string]string `json:"change"`
	Deletions   []string          `json:"delete"`
	BinaryFiles map[string]string `json:"binary"`
	Account     string            `json:"-"`
	Region      string            `json:"-"`
}

// DeployResponse deploy program response
type DeployResponse struct {
	ProgramID string `json:"program_id"`
}

// Deploy sends deploy request
func (c *DetaClient) Deploy(r *DeployRequest) (*DeployResponse, error) {
	headers := make(map[string]string)
	c.injectResourceHeader(headers, r.Account, r.Region)

	i := &requestInput{
		Path:      fmt.Sprintf("/%s/", patcherPath),
		Method:    "POST",
		Headers:   headers,
		Body:      r,
		NeedsAuth: true,
	}
	o, err := c.request(i)
	if err != nil {
		return nil, err
	}
	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to deploy: %v", msg)
	}

	var resp DeployResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy: %v", err)
	}
	return &resp, nil
}

// NewProgramRequest request to create a new program
type NewProgramRequest struct {
	Space   int64   `json:"spaceID"`
	Project string  `json:"project"`
	Group   string  `json:"group"`
	Name    string  `json:"name"`
	Runtime string  `json:"runtime"`
	Fork    *string `json:"fork"`
}

// NewProgramResponse response to a new program request
type NewProgramResponse struct {
	ID      string   `json:"id"`
	Space   int64    `json:"space"`
	Runtime string   `json:"runtime"`
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Project string   `json:"project"`
	Account string   `json:"account"`
	Region  string   `json:"region"`
	Deps    []string `json:"deps"`
	Envs    []string `json:"envs"`
	Public  bool     `json:"http_auth"`
	Visor   string   `json:"log_level"`
}

// NewProgram sends a new program request
func (c *DetaClient) NewProgram(r *NewProgramRequest) (*NewProgramResponse, error) {
	i := &requestInput{
		Path:      fmt.Sprintf("/%s/", "programs"),
		Method:    "POST",
		NeedsAuth: true,
		Body:      *r,
	}

	o, err := c.request(i)
	if err != nil {
		return nil, err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to create new program: %v", msg)
	}

	var resp NewProgramResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to create new program: %v", err)
	}
	return &resp, nil
}

// DownloadProgramRequest download program request
type DownloadProgramRequest struct {
	ProgramID string
	Runtime   string
	Account   string
	Region    string
}

// DownloadProgramResponse download program response
type DownloadProgramResponse struct {
	ZipFile []byte
}

// DownloadProgram download all program files
func (c *DetaClient) DownloadProgram(req *DownloadProgramRequest) (*DownloadProgramResponse, error) {
	headers := make(map[string]string)
	c.injectResourceHeader(headers, req.Account, req.Region)

	i := &requestInput{
		Path:      fmt.Sprintf("/%s/archives/%s", viewerPath, req.ProgramID),
		Method:    "GET",
		Headers:   headers,
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return nil, err
	}
	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to download micro: %v", msg)
	}
	return &DownloadProgramResponse{
		ZipFile: o.Body,
	}, nil
}

// ListSpaceItem an item in ListSpacesResponse
type ListSpaceItem struct {
	SpaceID int64  `json:"spaceID"`
	Name    string `json:"name"`
	Role    string `json:"role"`
}

// ListSpacesResponse response to list spaces request
type ListSpacesResponse []ListSpaceItem

// ListSpaces send list a spaces request
func (c *DetaClient) ListSpaces() (ListSpacesResponse, error) {
	i := &requestInput{
		Path:      fmt.Sprintf("/%s/", "spaces"),
		Method:    "GET",
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return nil, err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to list spaces: %v", msg)
	}
	var resp ListSpacesResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to list spaces: %v", err)
	}
	return resp, nil
}

// UpdateProgNameRequest request to update program name
type UpdateProgNameRequest struct {
	ProgramID string `json:"-"`
	Name      string `json:"name"`
}

// UpdateProgName update program name
func (c *DetaClient) UpdateProgName(req *UpdateProgNameRequest) error {

	i := &requestInput{
		Path:      fmt.Sprintf("/programs/%s", req.ProgramID),
		Method:    "PATCH",
		Body:      req,
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return fmt.Errorf("failed to update program name: %s", msg)
	}
	return nil
}

// UpdateProgEnvsRequest request to update program envs
type UpdateProgEnvsRequest struct {
	ProgramID string
	Account   string
	Region    string
	Vars      map[string]*string
}

// UpdateProgEnvs update program environment variables
func (c *DetaClient) UpdateProgEnvs(req *UpdateProgEnvsRequest) error {
	headers := make(map[string]string)
	c.injectResourceHeader(headers, req.Account, req.Region)

	i := &requestInput{
		Path:      fmt.Sprintf("/programs/%s/envs", req.ProgramID),
		Headers:   headers,
		NeedsAuth: true,
		Method:    "PATCH",
		Body:      req.Vars,
	}

	o, err := c.request(i)
	if err != nil {
		return err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return fmt.Errorf("failed to update env vars: %s", msg)
	}
	return nil
}

// UpdateProgRuntimeRequest request to update program runtime
type UpdateProgRuntimeRequest struct {
	ProgramID string `json:"-"`
	Runtime   string `json:"runtime"`
}

// UpdateProgRuntime update program runtime
func (c *DetaClient) UpdateProgRuntime(req *UpdateProgRuntimeRequest) error {
	i := &requestInput{
		Path:      fmt.Sprintf("/programs/%s/runtime", req.ProgramID),
		Method:    "PATCH",
		Body:      req,
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return fmt.Errorf("failed to update program runtime: %s", msg)
	}
	return nil
}

// UpdateProgDepsRequest request to update program dependencies
type UpdateProgDepsRequest struct {
	ProgramID string `json:"program_id"`
	Command   string `json:"command"`
}

// UpdateProgDepsResponse response to update program dependencies request
type UpdateProgDepsResponse struct {
	Output   string `json:"output"`
	HasError bool   `json:"-"` // if the output has error
}

// UpdateProgDeps update program dependencies
func (c *DetaClient) UpdateProgDeps(req *UpdateProgDepsRequest) (*UpdateProgDepsResponse, error) {
	i := &requestInput{
		Path:      fmt.Sprintf("/%s/commands", pigeonPath),
		Method:    "POST",
		NeedsAuth: true,
		Body:      req,
	}

	o, err := c.request(i)
	if err != nil {
		return nil, err
	}

	if o.Status != 200 {
		// 209 is used for special case for this request
		if o.Status != 209 {
			msg := o.Error.Message
			if msg == "" {
				msg = o.Error.Errors[0]
			}
			return nil, fmt.Errorf("failed to update dependencies: %s", msg)
		}
	}

	var resp UpdateProgDepsResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, err
	}
	if o.Status == 209 {
		resp.HasError = true
	}
	return &resp, nil
}

// UpdateAuthRequest request to update http auth for a program
type UpdateAuthRequest struct {
	ProgramID string `json:"-"`
	AuthValue bool   `json:"http_auth"`
}

// UpdateAuth update http auth (enable or disable) for a program
func (c *DetaClient) UpdateAuth(req *UpdateAuthRequest) error {
	i := &requestInput{
		Path:      fmt.Sprintf("/programs/%s/api", req.ProgramID),
		Method:    "PATCH",
		Body:      req,
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return fmt.Errorf("failed to update program auth: %v", msg)
	}
	return nil
}

// CreateAPIKeyRequest request to create an api key for a program
type CreateAPIKeyRequest struct {
	ProgramID   string `json:"program_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// CreateAPIKeyResponse response to create api key request
type CreateAPIKeyResponse struct {
	Name        string `json:"name"`
	Description string `json:"description,omitmepty"`
	Prefix      string `json:"prefix"`
	APIKey      string `json:"api_key"`
	Created     string `json:"created"`
}

// CreateAPIKey create an api key for your program
func (c *DetaClient) CreateAPIKey(req *CreateAPIKeyRequest) (*CreateAPIKeyResponse, error) {
	i := &requestInput{
		Path:      fmt.Sprintf("/api_keys/"),
		Method:    "POST",
		Body:      req,
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return nil, err
	}

	if o.Status != 201 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to create an api key: %v", msg)
	}

	var resp CreateAPIKeyResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeleteAPIKeyRequest request to delete an api key
type DeleteAPIKeyRequest struct {
	ProgramID string
	Name      string
}

// DeleteAPIKey delete an api key
func (c *DetaClient) DeleteAPIKey(req *DeleteAPIKeyRequest) error {
	i := &requestInput{
		Path:      fmt.Sprintf("/api_keys/%s/%s", req.ProgramID, req.Name),
		Method:    "DELETE",
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return fmt.Errorf("failed to delete api key: %v", msg)
	}
	return nil
}

// UpdateVisorModeRequest request to update visor mode
type UpdateVisorModeRequest struct {
	ProgramID string `json:"-"`
	Mode      string `json:"log_level"`
}

// UpdateVisorMode updates the visor mode for a program
func (c *DetaClient) UpdateVisorMode(req *UpdateVisorModeRequest) error {
	i := &requestInput{
		Path:      fmt.Sprintf("/programs/%s/log-level", req.ProgramID),
		Body:      req,
		NeedsAuth: true,
		Method:    "PATCH",
	}

	o, err := c.request(i)
	if err != nil {
		return err
	}
	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return fmt.Errorf("failed to update visor mode: %v", msg)
	}
	return nil
}

// GetProjectsRequest request to get your projects
type GetProjectsRequest struct {
	SpaceID int64
}

// GetProjectsItem an item in get projects response
type GetProjectsItem struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Created     string `json:"created"`
}

// GetProjectsResponse response to get projects request
type GetProjectsResponse struct {
	Projects []*GetProjectsItem `json:"projects"`
}

// GetProjects gets projects
func (c *DetaClient) GetProjects(req *GetProjectsRequest) (*GetProjectsResponse, error) {
	i := &requestInput{
		Path:      fmt.Sprintf("/spaces/%d/projects", req.SpaceID),
		Method:    "GET",
		NeedsAuth: true,
	}
	o, err := c.request(i)
	if err != nil {
		return nil, err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to get projects: %v", msg)
	}

	var resp GetProjectsResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetProgDetailsRequest request to get program details
type GetProgDetailsRequest struct {
	Program string
	Project string
	Space   int64
}

// GetProgDetailsResponse response to get program details
type GetProgDetailsResponse struct {
	ID      string   `json:"id"`
	Space   int64    `json:"space"`
	Runtime string   `json:"runtime"`
	Name    string   `json:"name"`
	Path    string   `json:"path"`
	Project string   `json:"project"`
	Account string   `json:"account"`
	Region  string   `json:"region"`
	Deps    []string `json:"deps"`
	Envs    []string `json:"envs"`
	Public  bool     `json:"public"`
	Visor   string   `json:"log_level"`
}

// GetProgDetails get program details
func (c *DetaClient) GetProgDetails(req *GetProgDetailsRequest) (*GetProgDetailsResponse, error) {
	i := &requestInput{
		Path:      fmt.Sprintf("/spaces/%d/projects/%s/programs/%s", req.Space, req.Project, req.Program),
		Method:    "GET",
		NeedsAuth: true,
	}
	o, err := c.request(i)
	if err != nil {
		return nil, err
	}
	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to get details: %v", msg)
	}
	var resp GetProgDetailsResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// InvokeProgRequest request to invoke a program
type InvokeProgRequest struct {
	ProgramID string `json:"-"`
	Action    string `json:"action,omitempty"`
	Body      string `json:"body,omitempty"`
}

// InvokeProgResponse response to invoke program request
type InvokeProgResponse struct {
	Logs    string `json:"logs"`
	Payload string `json:"payload"`
}

// InvokeProgram invoke lambda program
func (c *DetaClient) InvokeProgram(req *InvokeProgRequest) (*InvokeProgResponse, error) {
	i := &requestInput{
		Path:      fmt.Sprintf("/invocations/%s", req.ProgramID),
		Method:    "POST",
		Body:      req,
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return nil, err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to invoke program: %v", msg)
	}

	var resp InvokeProgResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// AddScheduleRequest request to add schedule/cron to a program
type AddScheduleRequest struct {
	ProgramID  string `json:"program_id"`
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

// AddSchedule add a schedule/cron to a program
func (c *DetaClient) AddSchedule(req *AddScheduleRequest) error {
	i := &requestInput{
		Path:      fmt.Sprintf("/schedules/"),
		Method:    "POST",
		NeedsAuth: true,
		Body:      req,
	}

	o, err := c.request(i)
	if err != nil {
		return err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return fmt.Errorf("failed to add schedule: %v", msg)
	}
	return nil
}

// DeleteScheduleRequest request to delete a schedule/cron from a program
type DeleteScheduleRequest struct {
	ProgramID string
}

// DeleteSchedule delete a schedule from a program
func (c *DetaClient) DeleteSchedule(req *DeleteScheduleRequest) error {
	i := &requestInput{
		Path:      fmt.Sprintf("/schedules/%s", req.ProgramID),
		Method:    "DELETE",
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return err
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return fmt.Errorf("failed to delete schedule: %v", msg)
	}
	return nil
}

// GetScheduleRequest request to get a schedule for a program
type GetScheduleRequest struct {
	ProgramID string
}

// GetScheduleResponse response to get schedule request
type GetScheduleResponse struct {
	ProgramID  string `json:"id"`
	Type       string `json:"type"`
	Expression string `json:"expression"`
}

// GetSchedule  get a schedule for a program
func (c *DetaClient) GetSchedule(req *GetScheduleRequest) (*GetScheduleResponse, error) {
	i := &requestInput{
		Path:      fmt.Sprintf("/schedules/%s", req.ProgramID),
		Method:    "GET",
		NeedsAuth: true,
	}

	o, err := c.request(i)
	if err != nil {
		return nil, err
	}

	if o.Status == 404 {
		return nil, nil
	}

	if o.Status != 200 {
		msg := o.Error.Message
		if msg == "" {
			msg = o.Error.Errors[0]
		}
		return nil, fmt.Errorf("failed to get schedule: %v", msg)
	}

	var resp GetScheduleResponse
	err = json.Unmarshal(o.Body, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetUserInfoResponse response to GetUserInfo request
type GetUserInfoResponse struct {
	DefaultSpace     int64
	DefaultSpaceName string
	DefaultProject   string
}

// GetUserInfo gets user info
func (c *DetaClient) GetUserInfo() (*GetUserInfoResponse, error) {
	resp, err := c.ListSpaces()
	if err != nil {
		return nil, err
	}

	return &GetUserInfoResponse{
		DefaultSpace:     resp[0].SpaceID,
		DefaultSpaceName: resp[0].Name,
		DefaultProject:   runtime.DefaultProject,
	}, nil
}

// GetLogsRequest to a micro
type GetLogsRequest struct {
	ProgramID string
	Start     int64
	End       int64
	LastToken string
}

// LogType is a single log record from api
type LogType struct {
	Timestamp int64  `json:"timestamp"`
	Log       string `json:"log"`
}

// GetLogsResponse from a micro
type GetLogsResponse struct {
	LastToken string    `json:"last_token"`
	Logs      []LogType `json:"logs"`
}

func (c *DetaClient) GetLogs(req *GetLogsRequest) (*GetLogsResponse, error) {
	r := &requestInput{
		Path:      fmt.Sprintf("/programs/%s/logs", req.ProgramID),
		Method:    "GET",
		NeedsAuth: true,
		QueryParams: map[string]string{
			"start":      fmt.Sprint(req.Start),
			"end":        fmt.Sprint(req.End),
			"last_token": req.LastToken,
		},
	}

	res, err := c.request(r)
	if err != nil {
		return nil, err
	}

	if res.Status != 200 {
		msg := res.Error.Message
		if msg == "" {
			msg = res.Error.Errors[0]
		}

		return nil, fmt.Errorf("failed to get logs: %v", msg)
	}

	result := &GetLogsResponse{
		Logs: make([]LogType, 0),
	}
	err = json.Unmarshal(res.Body, result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
