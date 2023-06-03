package utils

import (
	"fmt"
	"log"

	"github.com/fatih/color"
)

func HandleError(err error) {
	if err != nil {
		log.Fatal(err)
		panic(err)
	}
}

func PrintLn(content string, colAttr color.Attribute) {
	defer color.Set(color.Reset)
	color.Set(colAttr)
	fmt.Println(content)
}
