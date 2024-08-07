package sheets

import (
	"errors"
	"google.golang.org/api/sheets/v4"
)

type Sheet struct {
	id            string
	service       ServiceInterface
	googleService *sheets.Service
}

func NewSheet(id string) *Sheet {
	return &Sheet{
		id:            id,
		service:       nil,
		googleService: nil,
	}
}

func NewSheetByService(id string, service ServiceInterface) (*Sheet, error) {
	if service == nil {
		return nil, errors.New("internal is nil\n")
	}
	if !service.IsAuth() {
		return nil, errors.New("internal is not authorize\n")
	}

	return &Sheet{
		id:            id,
		service:       service,
		googleService: service.Internal(),
	}, nil
}

func (s *Sheet) SetService(service ServiceInterface) error {
	if !service.IsAuth() {
		return errors.New("internal is not authorized\n")
	}

	s.service = service
	s.googleService = service.Internal()

	return nil
}

func (s *Sheet) Update(pageName string, startIndex string, data [][]interface{}) error {
	if s.service == nil {
		return errors.New("there is no connection to internal\n")
	}

	vr := &sheets.ValueRange{
		Values: data,
	}

	_, err := s.googleService.Spreadsheets.Values.Update(s.id, pageName+"!"+startIndex, vr).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}

	return nil
}

func (s *Sheet) Append(pageName string, startIndex string, data [][]interface{}) error {
	if s.service == nil {
		return errors.New("there is no connection to internal\n")
	}

	valueRange := &sheets.ValueRange{
		Values: data,
	}

	_, err := s.googleService.Spreadsheets.Values.Append(s.id, pageName+"!"+startIndex, valueRange).ValueInputOption("RAW").Do()
	if err != nil {
		return err
	}

	return nil
}

func (s *Sheet) Clear(pageName string, startIndex string) error {
	if s.service == nil {
		return errors.New("there is no connection to internal\n")
	}

	_, err := s.googleService.Spreadsheets.Values.Clear(s.id, pageName+"!"+startIndex, &sheets.ClearValuesRequest{}).Do()
	if err != nil {
		return err
	}

	return nil
}
