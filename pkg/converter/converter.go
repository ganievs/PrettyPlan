package converter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
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

func ConvertPlan(plan *tfjson.Plan) (planData PlanData, err error) {
	for _, resource := range plan.ResourceChanges {
		if resource.Change.Actions.NoOp() || resource.Change.Actions.Read() {
			continue
		}

		switch {
		case resource.Change.Actions.Create():
			after, err := formatJson("+", resource.Change.After)
			if err != nil {
				fmt.Println("Error:", err)
			}
			planData.CreatedAddresses = make(map[string]interface{})
			planData.CreatedAddresses[resource.Address] = string(after)
			planData.ResourceChanges = append(planData.ResourceChanges, string(after))

		case resource.Change.Actions.Update():
			before, err := formatJson(" ", resource.Change.Before)
			if err != nil {
				fmt.Println("Error:", err)
			}
			after, err := formatJson(" ", resource.Change.After)
			if err != nil {
				fmt.Println("Error:", err)
			}
			diff := cmp.Diff(string(before), string(after))
			planData.UpdatedAddresses = make(map[string]interface{})
			planData.UpdatedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)

		case resource.Change.Actions.Delete():
			before, err := formatJson("-", resource.Change.Before)
			if err != nil {
				fmt.Println("Error:", err)
			}
			planData.DeletedAddresses = make(map[string]interface{})
			planData.DeletedAddresses[resource.Address] = string(before)
			planData.ResourceChanges = append(planData.ResourceChanges, string(before))

		case resource.Change.Actions.Replace():
			before, err := formatJson(" ", resource.Change.Before)
			if err != nil {
				fmt.Println("Error:", err)
			}
			after, err := formatJson(" ", resource.Change.After)
			if err != nil {
				fmt.Println("Error:", err)
			}
			diff := cmp.Diff(string(before), string(after))
			planData.ReplacedAddresses = make(map[string]interface{})
			planData.ReplacedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		}
	}

	return planData, err
}

func ConvertPlanTest(plan *tfjson.Plan) (planData string, err error) {
	var outputBuilder strings.Builder

	outputBuilder.WriteString("Terraform will perform the following actions:\n\n")

	for _, resource := range plan.ResourceChanges {
		if resource.Change.Actions.NoOp() || resource.Change.Actions.Read() {
			continue
		}

		switch {
		case resource.Change.Actions.Create():
			outputBuilder.WriteString(fmt.Sprintf("%s will be %s\n", resource.Address, "create"))
			outputBuilder.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", resource.Type, resource.Name))
			afterJSON, _ := json.MarshalIndent(resource.Change.After, "+", "  ")
			after := bytes.Trim(afterJSON, "\x00")
			// after, err := removeNullValues("+", afterJSON)
			// if err != nil {
			// 	fmt.Println("Error:", err)
			// }
			outputBuilder.WriteString(fmt.Sprintf("%s\n", string(after)))
		case resource.Change.Actions.Update():
			outputBuilder.WriteString(fmt.Sprintf("%s will be %s\n", resource.Address, "update"))
			// outputBuilder.WriteString(fmt.Sprintf("%s resource \"%s\" \"%s\" {\n", changeColor, resource.Type, resource.Name))
			beforeJSON, _ := json.MarshalIndent(resource.Change.Before, " ", "  ")
			before, err := formatJson("-", beforeJSON)
			if err != nil {
				fmt.Println("Error:", err)
			}
			after, err := formatJson("+", resource.Change.After)
			if err != nil {
				fmt.Println("Error:", err)
			}
			outputBuilder.WriteString(fmt.Sprintf("%s\n", string(before)))
			outputBuilder.WriteString(fmt.Sprintf("%s\n", string(after)))
		case resource.Change.Actions.Delete():
			outputBuilder.WriteString(fmt.Sprintf("%s will be %s\n", resource.Address, "delete"))
			outputBuilder.WriteString(fmt.Sprintf("resource \"%s\" \"%s\" {\n", resource.Type, resource.Name))
			beforeJSON, _ := json.MarshalIndent(resource.Change.Before, " ", "  ")
			before, err := formatJson("-", beforeJSON)
			if err != nil {
				fmt.Println("Error:", err)
			}
			outputBuilder.WriteString(fmt.Sprintf("%s\n", string(before)))
		}

		outputBuilder.WriteString("}\n\n")
	}

	// fmt.Println(outputBuilder.String())

	planData = outputBuilder.String()
	return planData, err
}

func removeNullValuesFromMap(data map[string]interface{}) {
	for key, value := range data {

		// Remove Null values
		if value == nil {
			delete(data, key)
			continue
		}

		// If the value is a map, recursively remove null values from it.
		if reflect.TypeOf(value).Kind() == reflect.Map {
			if subMap, ok := value.(map[string]interface{}); ok {
				removeNullValuesFromMap(subMap)
			}
		}

		// If the value is a slice, check each element.
		if reflect.TypeOf(value).Kind() == reflect.Slice {
			if subSlice, ok := value.([]interface{}); ok {
				for i, item := range subSlice {
					if itemMap, ok := item.(map[string]interface{}); ok {
						removeNullValuesFromMap(itemMap)
						subSlice[i] = itemMap
					}
				}
			}
		}
	}
}

func formatJson(ident string, jsonData interface{}) ([]byte, error) {
	if mapData, ok := jsonData.(map[string]interface{}); ok {
		removeNullValuesFromMap(mapData)
		return json.MarshalIndent(mapData, ident, "   ")
	}

	return nil, nil
}
