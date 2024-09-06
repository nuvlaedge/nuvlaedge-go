package release

type Release interface {
	GetComposeFiles(fileNames []string, workDir string) ([]string, error)
}
