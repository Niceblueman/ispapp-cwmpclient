package soap

import (
	"context"
	"regexp"
	"strconv"
	"strings"

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
	results, err := executer.Execute(context.Background(), "uci", "show Device")
	if err != nil {
		e.Body.Fault = &Fault{
			FaultCode:   "101",
			FaultString: "Failed to execute command",
		}
		return
	}
	paramsRegex := regexp.MustCompile("^(" + strings.Join(params, "|") + ")(\\..*)?$")
	lines := strings.Split(string(results.Raw), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1]) // Remove quotes if present
		if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			value = strings.Trim(value, "'")
		} else if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}
		if paramsRegex.MatchString(key) {
			// Determine the appropriate TR-069 type for this parameter
			tr069Type := StringTypeToTR069StandersType(key)

			// Create the value with proper type conversion
			paramValue := ValueToTR069Standers(Value{
				Type:    tr069Type,
				Content: value,
			})

			e.Body.GetParameterValuesResponse.ParameterList.Parameters = append(e.Body.GetParameterValuesResponse.ParameterList.Parameters, ParameterValueStruct{
				Name:  key,
				Value: paramValue,
			})
		}
	}
}
func StringTypeToTR069StandersType(parameterPath string) string {
	// Check each type rule pattern against the parameter path
	for pattern, tr069Type := range TR069ParameterTypeRules {
		if matched, _ := regexp.MatchString(pattern, parameterPath); matched {
			return tr069Type
		}
	}
	// Default to string type if no pattern matches
	return TR069TypeString
}
func ValueToTR069Standers(value Value) Value {
	// convert the value to the TR-069 standard formats
	content := strings.TrimSpace(value.Content)

	// If type is already set correctly, return as-is
	if value.Type != "" && value.Type != "string" {
		return value
	}

	// Auto-detect type based on content if not specified or is generic "string"
	detectedType := TR069TypeString // default

	// Check for boolean values
	if lowerContent := strings.ToLower(content); lowerContent != "" {
		if _, isBool := BooleanValues[lowerContent]; isBool {
			detectedType = TR069TypeBoolean
			// Normalize boolean representation
			if BooleanValues[lowerContent] {
				content = "true"
			} else {
				content = "false"
			}
		}
	}

	// Check for integer values
	if detectedType == TR069TypeString && content != "" {
		if _, err := strconv.Atoi(content); err == nil {
			// Check if it could be unsigned int (non-negative)
			if num, _ := strconv.Atoi(content); num >= 0 {
				detectedType = TR069TypeUnsignedInt
			} else {
				detectedType = TR069TypeInt
			}
		}
	}

	// Check for float values
	if detectedType == TR069TypeString && content != "" {
		if _, err := strconv.ParseFloat(content, 64); err == nil {
			// Contains decimal point, treat as float
			if strings.Contains(content, ".") {
				detectedType = TR069TypeFloat
			}
		}
	}

	// Check for datetime patterns (ISO 8601)
	if detectedType == TR069TypeString && content != "" {
		// Simple regex for ISO 8601 datetime format
		if matched, _ := regexp.MatchString(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`, content); matched {
			detectedType = TR069TypeDateTime
		}
	}

	return Value{
		Type:    detectedType,
		Content: content,
	}
}
