package common

// FileMissingError is returned when a file is missing
type FileMissingError struct {
	FileName string
}

func NewFileMissingError(fileName string) *FileMissingError {
	return &FileMissingError{FileName: fileName}
}

func (e *FileMissingError) Error() string {
	return "File " + e.FileName + " is missing"
}

// FileOpenError is returned when a file cannot be opened
type FileOpenError struct {
	FileName string
}

func NewFileOpenError(fileName string) *FileOpenError {
	return &FileOpenError{FileName: fileName}
}

func (e *FileOpenError) Error() string {
	return "Error opening file " + e.FileName + ": "
}
