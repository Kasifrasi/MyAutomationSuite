package report

func GetTemplateBytesDirect(path string) ([]byte, error) {
	return templateFiles.ReadFile(path)
}
