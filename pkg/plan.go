package prettyplan

import (
	"encoding/json"
	"io"

	tfjson "github.com/hashicorp/terraform-json"
)

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
