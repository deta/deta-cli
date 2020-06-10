package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	detaDir       = ".deta"
	authTokenPath = ".deta/tokens"
)

var (
	// set with Makefile during compilation
	loginURL string

	// port to start local server for login
	localServerPort int
)

// CognitoToken aws congito tokens
type CognitoToken struct {
	AccessToken  string `json:"access_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	Expires      string `json:"expires"`
}

// Manager manages aws cognito authentication
type Manager struct {
	srv       *http.Server
	tokenChan chan *CognitoToken
	errChan   chan error
}

// NewManager returns a new auth Manager
func NewManager() *Manager {
	srv := &http.Server{}
	return &Manager{
		tokenChan: make(chan *CognitoToken, 1),
		errChan:   make(chan error, 1),
		srv:       srv,
	}
}

// stores tokens in file ~/.deta/tokens
func (m *Manager) storeTokens(tokens *CognitoToken) error {
	expiresIn, err := m.expiresInFromToken(tokens.AccessToken)
	if err != nil {
		return err
	}
	tokens.Expires = expiresIn

	marshalled, err := json.Marshal(tokens)
	if err != nil {
		return err
	}

	// TODO: windows compatibility
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	detaDirPath := filepath.Join(home, detaDir)
	err = os.MkdirAll(detaDirPath, 0760)
	if err != nil {
		return err
	}

	tokensFilePath := filepath.Join(home, authTokenPath)

	err = ioutil.WriteFile(tokensFilePath, marshalled, 0660)
	if err != nil {
		return err
	}
	return nil
}

type tokenPayload struct {
	ExpiresIn int64 `json:"exp"`
}

// pulls token expire time from token, time is in seconds since Unix epoch
func (m *Manager) expiresInFromToken(accessToken string) (string, error) {
	tokenParts := strings.Split(accessToken, ".")
	if len(tokenParts) != 3 {
		return "", fmt.Errorf("access token is of invalid format")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return "", err
	}

	var payload tokenPayload
	err = json.Unmarshal(decoded, &payload)
	if err != nil {
		return "", err
	}
	e := payload.ExpiresIn
	if e == 0 {
		return "", fmt.Errorf("No expire time found in access token")
	}
	return fmt.Sprintf("%d", e), nil
}

// GetTokens retrieves the tokens from storage
func (m *Manager) GetTokens() (*CognitoToken, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil
	}

	tokensFilePath := filepath.Join(home, authTokenPath)
	f, err := os.Open(tokensFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var tokens CognitoToken
	err = json.Unmarshal(contents, &tokens)
	if err != nil {
		return nil, err
	}
	return &tokens, nil
}

// Login logs in to the user pool and stores the tokens
func (m *Manager) Login() error {
	err := m.useFreePort()
	if err != nil {
		return err
	}
	fmt.Println("Please, log in from the web page. Waiting..")
	err = m.openLoginPage()
	if err != nil {
		return err
	}
	err = m.retrieveTokens()
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) openLoginPage() error {
	loginURL = fmt.Sprintf("%s/%d", loginURL, localServerPort)
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", loginURL).Start()
	case "windows":
		return exec.Command("rundll32", "url.dll,FileProtocolHandler", loginURL).Start()
	case "darwin":
		return exec.Command("open", loginURL).Start()
	default:
		return fmt.Errorf("unsupported platform")
	}
}

func (m *Manager) tokenHandler(w http.ResponseWriter, r *http.Request) {
	// notify manager error channel of the error and return 500
	serverError := func(w http.ResponseWriter, err error) {
		m.errChan <- err
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}

	u, err := url.Parse(loginURL)
	if err != nil {
		serverError(w, err)
	}

	// CORS
	host := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	w.Header().Set("Access-Control-Allow-Origin", host)
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allowe-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var tokens CognitoToken
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		serverError(w, err)
	}
	err = json.Unmarshal(body, &tokens)
	if err != nil {
		serverError(w, err)
	}

	// provide tokens on token channel and return 200
	m.tokenChan <- &tokens
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// starts a local server
func (m *Manager) startLocalServer() {
	http.HandleFunc("/tokens", m.tokenHandler)

	m.srv.Addr = fmt.Sprintf(":%d", localServerPort)
	err := m.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		m.errChan <- err
	}
}

//  uses a free TCP port
func (m *Manager) useFreePort() error {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()
	localServerPort = l.Addr().(*net.TCPAddr).Port
	return nil
}

// shuts the server down
func (m *Manager) shutdownServer() {
	// returns an error but ignoring for now
	m.srv.Shutdown(context.Background())
}

// starts local server to retrieve tokens from login page
// shuts down the server on receiving the tokens
func (m *Manager) retrieveTokens() error {
	go m.startLocalServer()
	select {
	case err := <-m.errChan:
		m.shutdownServer()
		return err
	case tokens := <-m.tokenChan:
		if err := m.storeTokens(tokens); err != nil {
			m.shutdownServer()
			return err
		}
		m.shutdownServer()
		return nil
	}
}
