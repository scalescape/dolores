package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/rs/zerolog/log"
	"github.com/scalescape/dolores/config"
)

type credentials struct {
	APIToken string
	ID       string
}

type MonartClient struct {
	ctx  context.Context //nolint:containedctx
	cred credentials
	cli  *http.Client
}

var ErrMethodUndefined = errors.New("method not yet implemented")

func (s MonartClient) Init(ctx context.Context, bucket string, cfg Configuration) error {
	return ErrMethodUndefined
}

func (s MonartClient) UploadSecrets(ec EncryptedConfig) error {
	data, err := json.Marshal(ec)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	req, err := http.NewRequest(http.MethodPut, s.serverURL("secrets"), bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("unable to build upload request: %w", err)
	}
	if _, err := s.call(req, nil); err != nil {
		return err
	}
	return nil
}

type PublicKeyRequest struct {
	Environment string `json:"environment"`
}

func (s MonartClient) GetOrgPublicKeys(env string) (OrgPublicKeys, error) {
	pkreq := PublicKeyRequest{Environment: env}
	data, err := json.Marshal(pkreq)
	if err != nil {
		return OrgPublicKeys{}, fmt.Errorf("failed to marshal config: %w", err)
	}
	req, err := http.NewRequest(http.MethodGet, s.serverURL("secrets/recipients"), bytes.NewReader(data))
	if err != nil {
		return OrgPublicKeys{}, fmt.Errorf("unable to build public key request: %w", err)
	}
	var result OrgPublicKeys
	if _, err := s.call(req, &result); err != nil {
		return OrgPublicKeys{}, err
	}
	return result, nil
}

func (s MonartClient) FetchSecrets(fetchReq FetchSecretRequest) ([]byte, error) {
	data, err := json.Marshal(fetchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal fetch secret request: %w", err)
	}
	req, err := http.NewRequest(http.MethodGet, s.serverURL("secrets"), bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("unable to build fetch config request: %w", err)
	}
	result := new(FetchSecretResponse)
	if _, err := s.call(req, &result); err != nil {
		return nil, err
	}
	sec, err := base64.StdEncoding.DecodeString(result.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 response: %w", err)
	}
	return sec, nil
}

func (s MonartClient) GetSecretList(cfg SecretListConfig) ([]SecretObject, error) {
	path := fmt.Sprintf("environment/%s/secrets", cfg.Environment)
	req, err := http.NewRequest(http.MethodGet, s.serverURL(path), nil)
	if err != nil {
		return nil, fmt.Errorf("unable to build Get SecretList request: %w", err)
	}
	result := new(SecretListResponse)
	if _, err := s.call(req, &result); err != nil {
		return nil, err
	}
	return result.Secrets, nil
}

func (s MonartClient) serverURL(path string) string {
	return "https://relyonmetrics.com/api/" + path
}

func (s MonartClient) call(req *http.Request, dest any) (*http.Response, error) {
	req.Header.Add("User-ID", s.cred.ID)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", s.cred.APIToken))
	resp, err := s.cli.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call server: %w", err)
	}
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Error().Msgf("server failed with status: %d", resp.StatusCode)
		return nil, fmt.Errorf("server failed with response: %d %w", resp.StatusCode, err)
	}
	if dest != nil {
		if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func NewMonart(ctx context.Context, cfg *config.Monart) MonartClient {
	cred := credentials{APIToken: cfg.APIToken, ID: cfg.ID}
	return MonartClient{cli: http.DefaultClient, cred: cred, ctx: ctx}
}
