package sheets

import (
	"google.golang.org/api/sheets/v4"
)

type ServiceInterface interface {
	Internal() *sheets.Service
	IsAuth() bool
	Close()
}
