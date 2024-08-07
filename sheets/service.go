package sheets

import (
	"context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
	"net/http"
	"os"
)

type Service struct {
	httpClient *http.Client
	internal   *sheets.Service
	isAuth     bool
}

func NewService() *Service {
	return &Service{
		httpClient: nil,
		internal:   nil,
		isAuth:     false,
	}
}

func (s *Service) Auth(ctx context.Context, credentials string, scope Scope) error {
	file, err := os.ReadFile(credentials)
	if err != nil {
		return err
	}

	config, err := google.JWTConfigFromJSON(file, string(scope))
	if err != nil {
		return err
	}

	s.httpClient = config.Client(ctx)
	s.internal, err = sheets.NewService(ctx, option.WithHTTPClient(s.httpClient))
	if err != nil {
		return err
	}

	s.isAuth = true

	return nil
}

func (s *Service) Close() {
	if s.httpClient != nil {
		if transport, ok := s.httpClient.Transport.(*http.Transport); ok {
			transport.CloseIdleConnections()
		}
	}
	s.httpClient = nil
	s.internal = nil
	s.isAuth = false
}

func (s *Service) Internal() *sheets.Service {
	return s.internal
}

func (s *Service) IsAuth() bool {
	return s.isAuth
}
