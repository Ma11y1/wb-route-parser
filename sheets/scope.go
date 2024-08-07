package sheets

import "google.golang.org/api/sheets/v4"

type Scope string

const (
	DRIVE_SCOPE           Scope = sheets.DriveScope
	DRIVE_FILE_SCOPE      Scope = sheets.DriveFileScope
	DRIVE_READONLY_SCOPE  Scope = sheets.DriveReadonlyScope
	SHEETS_ALL_SCOPE      Scope = sheets.SpreadsheetsScope
	SHEETS_READONLY_SCOPE Scope = sheets.SpreadsheetsReadonlyScope
)
