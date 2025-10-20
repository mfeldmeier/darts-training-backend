package services

import (
	"darts-training-app/internal/config"
	"darts-training-app/internal/models"
	"errors"
	"strings"
	"sync"

	"github.com/MicahParks/keyfunc"
	"github.com/rs/zerolog/log"
	"github.com/go-resty/resty/v2"
)

const oidcURLPart = "/.well-known/openid-configuration"

type AuthManager struct {
	configuration         *config.Config
	restClient            *resty.Client
	jwks                  *keyfunc.JWKS
	oidc                  *models.OpenIDConfiguration
	oidcMutex             sync.Mutex
	tokenEndpointResponse *models.TokenEndpointResponse
}

func NewAuthManager(configuration *config.Config) *AuthManager {
	authManager := &AuthManager{
		configuration: configuration,
		restClient:    resty.New(),
	}

	err := authManager.loadJWKS()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load OIDC from the authentication provider!")
		return nil
	}

	return authManager
}

func (m *AuthManager) GetJWKS() (*keyfunc.JWKS, error) {
	if m.jwks == nil {
		if err := m.loadJWKS(); err != nil {
			return nil, err
		}
	}
	return m.jwks, nil
}

func (m *AuthManager) InvalidateClientCredential() {
	m.tokenEndpointResponse = nil
}

func (m *AuthManager) GetClientCredential() (string, error) {
	if m.tokenEndpointResponse == nil {
		err := m.refreshClientCredential()
		if err != nil || m.tokenEndpointResponse == nil {
			return "", errors.New("no client credential")
		}
	}
	return m.tokenEndpointResponse.AccessToken, nil
}

func (m *AuthManager) refreshClientCredential() error {
	if err := m.ensureOIDC(); err != nil {
		return err
	}

	tokenEndpointResponse, err := m.callAuthProviderTokenEndpoint()
	if tokenEndpointResponse == nil || err != nil {
		log.Error().Err(err).Msg("Failed to load JWT token from the authentication provider")
		return err
	}

	if tokenEndpointResponse.TokenType != "Bearer" {
		log.Error().Msg("Got invalid token type from the authentication provider")
		return errors.New("invalid token type")
	}

	m.tokenEndpointResponse = tokenEndpointResponse

	return nil
}

func (m *AuthManager) loadJWKS() error {
	err := m.ensureOIDC()
	if err != nil {
		return err
	}

	jwks, err := keyfunc.Get(m.oidc.JwksURI, keyfunc.Options{
		Client:              m.restClient.GetClient(),
		RefreshErrorHandler: m.refreshErrorHandler,
		RefreshUnknownKID:   true,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWKS from the authentication provider!")
		return err
	}

	m.jwks = jwks
	return nil
}

func (m *AuthManager) refreshErrorHandler(err error) {
	log.Error().Err(err).Msg("Failed to get JWKS from the authentication provider!")
}

func (m *AuthManager) callAuthProviderOIDCEndpoint() (*models.OpenIDConfiguration, error) {
	response, err := m.restClient.R().
		SetHeader("Content-Type", "application/json").
		SetResult(&models.OpenIDConfiguration{}).
		Get(strings.TrimRight(m.configuration.OidcBaseURL, "/") + oidcURLPart)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get OIDC from the authentication provider")
		return nil, err
	}

	if !response.IsSuccess() {
		log.Error().Err(err).Msgf("Failed to get OIDC from the authentication provider: %v", response.Error())
		return nil, err
	}

	oidc := response.Result().(*models.OpenIDConfiguration)

	return oidc, nil
}

func (m *AuthManager) callAuthProviderTokenEndpoint() (*models.TokenEndpointResponse, error) {
	response, err := m.restClient.R().
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		SetHeader("Cache-Control", "no-cache").
		SetAuthScheme("Basic").
		SetAuthToken(m.configuration.ClientCredentialAuthHeaderValue).
		SetResult(&models.TokenEndpointResponse{}).
		SetFormData(map[string]string{"grant_type": "client_credentials"}).
		Post(m.oidc.TokenEndpoint)

	if err != nil {
		log.Error().Err(err).Msg("Failed to get JWT token from the authentication provider's token endpoint")
		return nil, err
	}

	if !response.IsSuccess() {
		log.Error().Err(err).Msgf("Failed to get JWT token from the authentication provider's token endpoint: %v", response.Status())
		return nil, err
	}

	tokenEndpointResponse := response.Result().(*models.TokenEndpointResponse)

	return tokenEndpointResponse, nil
}

func (m *AuthManager) ensureOIDC() error {
	if m.oidc == nil {
		oidc, err := m.callAuthProviderOIDCEndpoint()
		if err != nil {
			log.Error().Err(err).Msg("Failed to load OIDC")
			return err
		}

		m.updateOIDC(oidc)
	}

	return nil
}

func (m *AuthManager) updateOIDC(oidc *models.OpenIDConfiguration) {
	m.oidcMutex.Lock()
	m.oidc = oidc
	m.oidcMutex.Unlock()
}
