package version

var (
	version  = "dev"
	branch   = "main"
	date     = "unknown"
	revision = "unknown"
)

func Version() string {
	return version
}

func Branch() string {
	return branch
}

func Date() string {
	return date
}

func Revision() string {
	return revision
}
