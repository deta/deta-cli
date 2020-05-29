package auth

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cognitoidentityprovider"
)

const (
	detaDir       = ".deta"
	authTokenPath = ".deta/tokens"
)

var (
	// set with Makefile during compilation
	cognitoClientID string
	accessKeyID     string
	accessKeySecret string
	userpoolRegion  string
)

// Manager manages aws cognito authentication
type Manager struct {
	cip *cognitoidentityprovider.CognitoIdentityProvider
}

// NewManager returns a new auth Manager
func NewManager() *Manager {
	return &Manager{}
}

// starts a new aws session for cognito identity provider
func (m *Manager) newSession() error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(userpoolRegion),
		Credentials: credentials.NewStaticCredentials(accessKeyID, accessKeySecret, ""),
	})
	if err != nil {
		return err
	}
	m.cip = cognitoidentityprovider.New(sess)
	return nil
}

// stores tokens in file ~/.deta/creds
func (m *Manager) storeTokens(tokens *cognitoidentityprovider.AuthenticationResultType) error {
	// TODO: windows compatibility
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	detaDirPath := filepath.Join(home, detaDir)
	err = os.MkdirAll(detaDirPath, 0660)
	if err != nil {
		return err
	}

	tokensFilePath := filepath.Join(home, authTokenPath)
	f, err := os.OpenFile(tokensFilePath, os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		return err
	}
	defer f.Close()

	/*
		Tokens file is written:
		access_token
		id_token
		refresh_token
		expiration_time
	*/
	_, err = f.WriteString(fmt.Sprintf("%s\n", *tokens.AccessToken))
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("%s\n", *tokens.IdToken))
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("%s\n", *tokens.RefreshToken))
	if err != nil {
		return err
	}
	_, err = f.WriteString(fmt.Sprintf("%s\n", m.expiresInToTimestamp(*tokens.ExpiresIn)))
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
func (m *Manager) Login(username, password string) error {
	err := m.newSession()
	if err != nil {
		return err
	}
	o, err := m.cip.InitiateAuth(&cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: aws.String("USER_PASSWORD_AUTH"),
		ClientId: aws.String(cognitoClientID),
	})
	if err != nil {
		return err
	}
	err = m.storeTokens(o.AuthenticationResult)
	if err != nil {
		return err
	}
	return nil
}
