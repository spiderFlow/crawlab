package utils

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
)

func PrintLogoWithWelcomeInfo() {
	printLogo()
	printWelcomeInfo()
}

func printLogo() {
	figure.NewColorFigure("Crawlab", "slant", "blue", true).Print()
	fmt.Println()
}

func printWelcomeInfo() {
	fmt.Println("Distributed web crawling platform for efficient, scalable data extraction.")
	fmt.Println("For more information, please refer to the following resources:")
	fmt.Println("- Website:        https://crawlab.cn")
	fmt.Println("- Documentation:  https://docs.crawlab.cn")
	fmt.Println("- GitHub:         https://github.com/crawlab-team/crawlab")
	if IsMaster() {
		fmt.Println("Visit https://localhost:8080 for the web ui, once the server is ready.")
	}
	fmt.Println()
}
