package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
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
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	cidp "github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

const (
	detaDir            = ".deta"
	authTokenPath      = ".deta/tokens"
	detaAccessTokenEnv = "DETA_ACCESS_TOKEN"
)

var (
	// set with Makefile during compilation
	loginURL        string
	cognitoClientID string
	cognitoRegion   string

	// port to start local server for login
	localServerPort int

	// ErrRefreshTokenInvalid refresh token invalid
	ErrRefreshTokenInvalid = errors.New("refresh token is invalid")
	// ErrNoAuthTokenFound no access token found
	ErrNoAuthTokenFound = errors.New("no auth token found")
	// ErrInvalidAccessToken invalid access token
	ErrInvalidAccessToken = errors.New("invalid access token")
)

// Token aws cognito token or access keys
type Token struct {
	AccessToken     string `json:"access_token"`
	IDToken         string `json:"id_token"`
	RefreshToken    string `json:"refresh_token"`
	Expires         int64  `json:"expires"`
	DetaAccessToken string `json:"deta_access_token"`
}

// Manager manages aws cognito authentication
type Manager struct {
	bearerAuth bool
	srv        *http.Server
	tokenChan  chan *Token
	errChan    chan error
}

// NewManager returns a new auth Manager
func NewManager() *Manager {
	srv := &http.Server{}

	return &Manager{
		bearerAuth: true,
		tokenChan:  make(chan *Token, 1),
		errChan:    make(chan error, 1),
		srv:        srv,
	}
}

// stores tokens in file ~/.deta/tokens
func (m *Manager) storeTokens(tokens *Token) error {
	expiresIn, err := m.expiresFromToken(tokens.AccessToken)
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
	Expires int64 `json:"exp"`
}

// pulls token expire time from token, time is in seconds since Unix epoch
func (m *Manager) expiresFromToken(accessToken string) (int64, error) {
	tokenParts := strings.Split(accessToken, ".")
	if len(tokenParts) != 3 {
		return 0, fmt.Errorf("access token is of invalid format")
	}

	decoded, err := base64.RawURLEncoding.DecodeString(tokenParts[1])
	if err != nil {
		return 0, err
	}

	var payload tokenPayload
	err = json.Unmarshal(decoded, &payload)
	if err != nil {
		return 0, err
	}
	e := payload.Expires
	if e == 0 {
		return 0, fmt.Errorf("no expire time found in access token")
	}
	return e, nil
}

// checks if token is already expired
func (m *Manager) isTokenExpired(token *Token) bool {
	if !m.bearerAuth {
		return false
	}
	unixTime := time.Now().Unix()
	return token.Expires < unixTime
}

// getTokens retrieves the tokens from storage or env var
func (m *Manager) getTokens() (*Token, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, nil
	}

	tokensFilePath := filepath.Join(home, authTokenPath)
	f, err := os.Open(tokensFilePath)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	defer f.Close()

	// ignoring errors here
	// as we fall back to retrieving acces token from env
	// if not found in env then will finally return an error
	var tokens Token
	contents, _ := ioutil.ReadAll(f)
	json.Unmarshal(contents, &tokens)

	// first priority to aws access token
	if tokens.AccessToken != "" {
		return &tokens, nil
	}

	// if no congito token
	m.bearerAuth = false

	// check the env first for deta access token
	detaAccessToken := os.Getenv(detaAccessTokenEnv)

	// if not in env, check from tokens retreived from file
	if detaAccessToken == "" {
		detaAccessToken = tokens.DetaAccessToken
	}

	// if still no token found, return err
	if detaAccessToken == "" {
		return nil, ErrNoAuthTokenFound
	}
	return &Token{
		DetaAccessToken: detaAccessToken,
	}, nil
}

// refreshes the tokens
func (m *Manager) refreshTokens() (*Token, error) {
	tokens, err := m.getTokens()
	if err != nil {
		return nil, err
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cognitoRegion),
		Credentials: credentials.AnonymousCredentials,
	})
	if err != nil {
		return nil, err
	}

	idp := cidp.New(sess)
	o, err := idp.InitiateAuth(&cidp.InitiateAuthInput{
		AuthFlow: aws.String("REFRESH_TOKEN_AUTH"),
		AuthParameters: map[string]*string{
			"REFRESH_TOKEN": aws.String(tokens.RefreshToken),
		},
		ClientId: aws.String(cognitoClientID),
	})
	if err != nil {
		var aerr awserr.Error
		if errors.As(err, &aerr) {
			if aerr.Code() == cidp.ErrCodeNotAuthorizedException {
				return nil, ErrRefreshTokenInvalid
			}
		}
		return nil, err
	}

	authResult := o.AuthenticationResult
	if authResult == nil {
		return nil, fmt.Errorf("failed to refresh auth tokens")
	}

	newTokens := &Token{
		AccessToken:  *authResult.AccessToken,
		IDToken:      *authResult.IdToken,
		RefreshToken: tokens.RefreshToken,
	}
	err = m.storeTokens(newTokens)
	if err != nil {
		return nil, err
	}
	return newTokens, nil
}

// GetTokens gets tokens from env var or local storage if not expired
// else refreshes the tokens first and then provides the new tokens
func (m *Manager) GetTokens() (*Token, error) {
	tokens, err := m.getTokens()
	if err != nil {
		return nil, err
	}

	if !m.isTokenExpired(tokens) {
		return tokens, nil
	}

	newTokens, err := m.refreshTokens()
	if err != nil {
		return nil, err
	}
	return newTokens, err
}

// Login logs in to the user pool and stores the tokens
func (m *Manager) Login() error {
	err := m.useFreePort()
	if err != nil {
		return err
	}
	err = m.openLoginPage()
	if err != nil {
		fmt.Println("Failed to open the login page, open the following link in your browser:")
		fmt.Println(loginURL)
	}
	err = m.retrieveTokensFromServer()
	if err != nil {
		return err
	}
	return nil
}

// IsBearerAuth check if token auth type is bearer
func (m *Manager) IsBearerAuth() bool {
	return m.bearerAuth
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
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var tokens Token
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
func (m *Manager) retrieveTokensFromServer() error {
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
