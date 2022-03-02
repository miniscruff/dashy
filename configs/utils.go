package configs

import (
	"strings"
	"os"
)

func stringOrEnvVar(value string) string {
	if strings.HasPrefix(value, "env:") {
		return os.Getenv(value[4:])
	}
	return value
}
