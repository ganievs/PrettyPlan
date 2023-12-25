package main

import (
	"fmt"
	"os"

	"prettyplan/pkg/converter"
)

func main() {
	os.Exit(run())
}

func run() int {
	planData, err := converter.DecodePlan(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot parse input as Terraform plan JSON: %v", err)
		return 1
	}
	parsed, err := converter.ConvertPlan(planData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot render: %v", err)
		return 1
	}
	fmt.Println(parsed)
	return 0
}
