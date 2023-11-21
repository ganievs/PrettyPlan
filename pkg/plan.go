package prettyplan

import (
	"encoding/json"
	"io"

	"github.com/google/go-cmp/cmp"
	tfjson "github.com/hashicorp/terraform-json"
)

type PlanData struct {
	CreatedAddresses  map[string]string
	UpdatedAddresses  map[string]string
	DeletedAddresses  map[string]string
	ReplacedAddresses map[string]string
	MovedAddresses    map[string]string
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

func SerializePlan(plan *tfjson.Plan) (err error) {
	planData := new(PlanData)

	for _, resource := range plan.ResourceChanges {
		if resource.Change.Actions.NoOp() || resource.Change.Actions.Read() {
			continue
		}

		switch {
		case resource.Change.Actions.Create():
			diff := cmp.Diff(resource.Change.Before, resource.Change.After)
			planData.CreatedAddresses = make(map[string]string)
			planData.CreatedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		case resource.Change.Actions.Update():
			diff := cmp.Diff(resource.Change.Before, resource.Change.After)
			planData.UpdatedAddresses = make(map[string]string)
			planData.UpdatedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		case resource.Change.Actions.Delete():
			diff := cmp.Diff(resource.Change.Before, resource.Change.After)
			planData.DeletedAddresses = make(map[string]string)
			planData.DeletedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		case resource.Change.Actions.Replace():
			diff := cmp.Diff(resource.Change.Before, resource.Change.After)
			planData.ReplacedAddresses = make(map[string]string)
			planData.ReplacedAddresses[resource.Address] = diff
			planData.ResourceChanges = append(planData.ResourceChanges, diff)
		}
	}

	return err
}
