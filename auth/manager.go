package auth

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	detaDir         = ".deta"
	authTokenPath   = ".deta/tokens.txt"
	localServerPort = ":9000"
)

var (
	// set with Makefile during compilation
	loginURL string
)

// aws congito tokens
type cognitoToken struct {
	AccessToken  string `json:"acess_token"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // in seconds
}

// Manager manages aws cognito authentication
type Manager struct {
	srv       *http.Server
	tokenChan chan *cognitoToken
	errChan   chan error
}

// NewManager returns a new auth Manager
func NewManager() *Manager {
	srv := &http.Server{
		Addr: localServerPort,
	}
	return &Manager{
		tokenChan: make(chan *cognitoToken, 1),
		errChan:   make(chan error, 1),
		srv:       srv,
	}
}

// stores tokens in file ~/.deta/creds
func (m *Manager) storeTokens(tokens *cognitoToken) error {
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
	f, err := os.OpenFile(tokensFilePath, os.O_CREATE|os.O_WRONLY, 0760)
	if err != nil {
		return err
	}
	defer f.Close()

	/*
		Tokens file is written as:
		access_token
		id_token
		refresh_token
		expiration_time
	*/
	_, err = f.WriteString(fmt.Sprintf("%s\n", tokens.AccessToken))
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("%s\n", tokens.IDToken))
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("%s\n", tokens.RefreshToken))
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("%s\n", m.expiresInToTimestamp(tokens.ExpiresIn)))
	if err != nil {
		return err
	}
	return nil
}

// covert expires in to timestamp(RFC3339) calculated from current time
func (m *Manager) expiresInToTimestamp(expiresIn int64) string {
	timestamp := time.Now().Add(time.Duration(expiresIn) * time.Second)
	return timestamp.Format(time.RFC3339)
}

// GetAccessToken retrieves the access token from storage
func (m *Manager) GetAccessToken() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", nil
	}
	tokensFilePath := filepath.Join(home, detaDir, authTokenPath)
	f, err := os.Open(tokensFilePath)
	if err != nil {
		return "", err
	}
	reader := bufio.NewReader(f)
	accessToken, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	accessToken = strings.TrimSuffix(accessToken, "\n")
	return "", nil
}

// Login logs in to the user pool and stores the tokens
func (m *Manager) Login() error {
	fmt.Println("Please, log in from the web page. Waiting..")
	err := m.openLoginPage()
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

	var tokens cognitoToken
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
	err := m.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		m.errChan <- err
	}
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
