package constants

const (
	WorkerCount    = 10
	HeaderCount    = 6
	CLIWorkerCount = 100
	ARGS           = 3
)

var CsvHeader = []string{
	"URL", "HTML Version", "Title",
	"H1", "H2", "H3", "H4", "H5", "H6",
	"Internal Links", "External Links", "Inaccessible Links",
	"Has Login Form",
}
