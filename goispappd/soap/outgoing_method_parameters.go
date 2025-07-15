package soap

import (
	"regexp"
	"strings"

	"github.com/Niceblueman/goispappd/internal/commands"
	"github.com/Niceblueman/goispappd/internal/exec"
)

func (e *RequestEnvelope) LoadParametersValues(params []string) {
	// Initialize the GetParameterValuesResponse
	executer := exec.NewExecutor(exec.ExecConfig{})
	e.Body.GetParameterValuesResponse = &GetParameterValuesResponse{
		ParameterList: ParameterList{
			Parameters: make([]ParameterValueStruct, 0, len(params)),
		},
	}
	// strings.Join(params, "|") parma could ne "Device.DeviceInfo.*" so we need to match all parmasvalues keys that start with "Device.DeviceInfo."
	paramsRegex := regexp.MustCompile("^(" + strings.Join(params, "|") + ")(\\..*)?$")
	for key, getter := range commands.ParametersValues {
		if paramsRegex.MatchString(key) {
			if result, err := getter(executer, nil); err == nil && result.Success {
				value := Value{
					Type:    "string",
					Content: string(result.Raw),
				}
				e.Body.GetParameterValuesResponse.ParameterList.Parameters = append(e.Body.GetParameterValuesResponse.ParameterList.Parameters, ParameterValueStruct{
					Name:  key,
					Value: value,
				})
			}
		}
	}
}
