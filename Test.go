package main

import (
	"fmt"
	"github.com/Ericwyn/TransUtils/conf"
	"github.com/Ericwyn/TransUtils/strutils"
)

func main() {
	testFormatInput()
}

func testFormatInput() {
	str := `/**
 * The service class that manages LocationProviders and issues location
 * updates and alerts.
 */`

	conf.InitConfig()
	fmt.Println(strutils.FormatInputBoxText(str))

}
