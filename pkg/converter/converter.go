package converter

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/go-cmp/cmp"
	tfjson "github.com/hashicorp/terraform-json"
)

type PlanData struct {
	CreatedAddresses  map[string]interface{}
	UpdatedAddresses  map[string]interface{}
	DeletedAddresses  map[string]interface{}
	ReplacedAddresses map[string]interface{}
	MovedAddresses    map[string]interface{}
	ResourceChanges   []string
}

func DecodePlan(input io.Reader) (plan *tfjson.Plan, err error) {
	err = json.NewDecoder(input).Decode(&plan)
	if err != nil {
		return plan, err
	}

	if err := plan.Validate(); err != nil {
		return plan, err
	}

	return plan, err
}

func ConvertPlanOld(plan *tfjson.Plan) (planData PlanData, err error) {
	for _, resource := range plan.ResourceChanges {
		if resource.Change.Actions.NoOp() || resource.Change.Actions.Read() {
			continue
		}

		switch {
		case resource.Change.Actions.Create():
			beforeJSON, _ := json.MarshalIndent(resource.Change.Before, "", "  ")
			afterJSON, _ := json.MarshalIndent(resource.Change.After, "", "  ")
			diff := cmp.Diff(string(beforeJSON), string(afterJSON))
			planData.CreatedAddresses = make(map[string]interface{})
			planData.CreatedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		case resource.Change.Actions.Update():
			beforeJSON, _ := json.MarshalIndent(resource.Change.Before, "", "  ")
			afterJSON, _ := json.MarshalIndent(resource.Change.After, "", "  ")
			diff := cmp.Diff(string(beforeJSON), string(afterJSON))
			planData.UpdatedAddresses = make(map[string]interface{})
			planData.UpdatedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		case resource.Change.Actions.Delete():
			beforeJSON, _ := json.MarshalIndent(resource.Change.Before, "", "  ")
			afterJSON, _ := json.MarshalIndent(resource.Change.After, "", "  ")
			diff := cmp.Diff(string(beforeJSON), string(afterJSON))
			planData.DeletedAddresses = make(map[string]interface{})
			planData.DeletedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		case resource.Change.Actions.Replace():
			beforeJSON, _ := json.MarshalIndent(resource.Change.Before, "", "  ")
			afterJSON, _ := json.MarshalIndent(resource.Change.After, "", "  ")
			diff := cmp.Diff(string(beforeJSON), string(afterJSON))
			planData.ReplacedAddresses = make(map[string]interface{})
			planData.ReplacedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		}
	}

	return planData, err
}

func ConvertPlan(plan *tfjson.Plan) (planData string, err error) {
	var outputBuilder strings.Builder

	outputBuilder.WriteString("Terraform will perform the following actions:\n\n")

	for _, resource := range plan.ResourceChanges {
		if resource.Change.Actions.NoOp() || resource.Change.Actions.Read() {
			continue
		}

		var changeColor string
		switch {
		case resource.Change.Actions.Create():
			changeColor = "\033[32m" // Green
			outputBuilder.WriteString(fmt.Sprintf("\033[1m# %s\033[0m will be %s\n", resource.Address, "create"))
			outputBuilder.WriteString(fmt.Sprintf("%s resource \"%s\" \"%s\" {\n", changeColor, resource.Type, resource.Name))
			afterJSON, _ := json.MarshalIndent(resource.Change.After, "+", "  ")
			outputBuilder.WriteString(fmt.Sprintf("  %s+ %s\n", changeColor, string(afterJSON)))
		case resource.Change.Actions.Update():
			changeColor = "\033[33m" // Yellow
			outputBuilder.WriteString(fmt.Sprintf("\033[1m# %s\033[0m will be %s\n", resource.Address, "update"))
			// outputBuilder.WriteString(fmt.Sprintf("%s resource \"%s\" \"%s\" {\n", changeColor, resource.Type, resource.Name))
			beforeJSON, _ := json.MarshalIndent(resource.Change.Before, "-", "  ")
			afterJSON, _ := json.MarshalIndent(resource.Change.After, "+", "  ")
			outputBuilder.WriteString(fmt.Sprintf("  %s~ before = %s\n", changeColor, string(beforeJSON)))
			outputBuilder.WriteString(fmt.Sprintf("  %s~ after = %s\n", changeColor, string(afterJSON)))
		case resource.Change.Actions.Delete():
			changeColor = "\033[31m" // Red
			outputBuilder.WriteString(fmt.Sprintf("\033[1m# %s\033[0m will be %s\n", resource.Address, "delete"))
			outputBuilder.WriteString(fmt.Sprintf("%s resource \"%s\" \"%s\" {\n", changeColor, resource.Type, resource.Name))
			beforeJSON, _ := json.MarshalIndent(resource.Change.Before, "-", "  ")
			outputBuilder.WriteString(fmt.Sprintf("  %s- %s\n", changeColor, string(beforeJSON)))
		}

		outputBuilder.WriteString("}\n\n")
	}

	// fmt.Println(outputBuilder.String())

	planData = outputBuilder.String()
	return planData, err
}
