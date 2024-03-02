package conf

import (
	"fmt"
	"testing"
)

func TestGetConfigFilePath(t *testing.T) {
	fmt.Println(GetConfigFilePath())
}
