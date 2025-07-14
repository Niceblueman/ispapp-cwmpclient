package device

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/Niceblueman/goispappd/soap"
)

// Compare efficiently compares two Device structs and returns differences in SetParameterValues format
// Time complexity: O(n) where n is the number of fields in the Device struct
// Space complexity: O(d) where d is the number of differences found
// Uses parameter indexer approach to ensure valid TR-069 parameter paths
func (d *Device) Compare(other *Device) *soap.SetParameterValues {
	if d == nil || other == nil {
		return nil
	}

	// Pre-allocate slice with reasonable capacity to minimize allocations
	var differences []ParameterValueStruct

	// Use reflection to compare structs efficiently
	currentVal := reflect.ValueOf(d).Elem()
	otherVal := reflect.ValueOf(other).Elem()
	currentType := currentVal.Type()

	// Compare all fields recursively with proper TR-069 path validation
	compareStructFields(currentVal, otherVal, currentType, "Device", &differences, make(map[reflect.Type]bool))

	if len(differences) == 0 {
		return nil
	}

	// Build SetParameterValues response
	result := &soap.SetParameterValues{
		ParameterList: struct {
			Params []struct {
				Name  string `xml:"cwmp:Name"`
				Value string `xml:"cwmp:Value"`
			} `xml:"cwmp:ParameterValueStruct"`
		}{},
		ParameterKey: fmt.Sprintf("Compare_%d_changes", len(differences)),
	}

	// Convert our internal format to SOAP format
	result.ParameterList.Params = make([]struct {
		Name  string `xml:"cwmp:Name"`
		Value string `xml:"cwmp:Value"`
	}, len(differences))

	for i, diff := range differences {
		result.ParameterList.Params[i].Name = diff.Name
		result.ParameterList.Params[i].Value = diff.Value
	}

	return result
}

// ParameterValueStruct represents a parameter difference internally
type ParameterValueStruct struct {
	Name  string
	Value string
}

// compareStructFields recursively compares struct fields with optimal performance
// Uses parameter indexer approach for proper field name resolution and circular reference prevention
func compareStructFields(currentVal, otherVal reflect.Value, structType reflect.Type, pathPrefix string, differences *[]ParameterValueStruct, visited map[reflect.Type]bool) {
	// Prevent infinite recursion (similar to parameter indexer)
	if visited[structType] {
		return
	}
	visited[structType] = true
	defer func() { delete(visited, structType) }()

	// Single pass to compare all fields
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip unexported fields for performance
		if !field.IsExported() {
			continue
		}

		currentFieldVal := currentVal.Field(i)
		otherFieldVal := otherVal.Field(i)

		// Build parameter path using proper field name resolution
		fieldPath := buildFieldPath(pathPrefix, field.Name, field.Type)

		// Compare fields based on type with type-specific optimizations
		compareFieldValues(currentFieldVal, otherFieldVal, fieldPath, field.Type, differences, visited)
	}
}

// buildFieldPath constructs valid TR-069 parameter paths using parameter indexer logic
func buildFieldPath(pathPrefix, fieldName string, fieldType reflect.Type) string {
	// Handle plural form conversions for array fields (like parameter indexer)
	if fieldType.Kind() == reflect.Slice {
		// Convert singular to plural for array fields (same as parameter indexer logic)
		switch fieldName {
		case "Interface":
			fieldName = "Interfaces"
		case "AccessPoint":
			fieldName = "AccessPoints"
		case "Radio":
			fieldName = "Radios"
		case "SSID":
			fieldName = "SSIDs"
		case "AssociatedDevice":
			fieldName = "AssociatedDevices"
		case "IPv4Address":
			fieldName = "IPv4Addresses"
		case "Router":
			fieldName = "Routers"
		case "IPv4Forwarding":
			fieldName = "IPv4Forwardings"
			// Add more mappings as needed based on the Device struct
		}
	}

	// Handle global Device prefix (like parameter indexer does)
	// If pathPrefix is "Device", don't add extra prefix for the first level
	if pathPrefix == "Device" {
		return "Device." + fieldName
	}

	// Build path efficiently using strings.Builder
	var pathBuilder strings.Builder
	pathBuilder.Grow(len(pathPrefix) + len(fieldName) + 1) // Pre-allocate capacity
	pathBuilder.WriteString(pathPrefix)
	pathBuilder.WriteByte('.')
	pathBuilder.WriteString(fieldName)
	return pathBuilder.String()
}

// compareFieldValues efficiently compares individual field values
// Optimized for common types to minimize reflection overhead
func compareFieldValues(currentVal, otherVal reflect.Value, paramPath string, fieldType reflect.Type, differences *[]ParameterValueStruct, visited map[reflect.Type]bool) {
	// Handle nil pointers efficiently
	if currentVal.Kind() == reflect.Ptr || otherVal.Kind() == reflect.Ptr {
		if currentVal.IsNil() && otherVal.IsNil() {
			return
		}
		if currentVal.IsNil() || otherVal.IsNil() {
			addDifference(differences, paramPath, getStringValue(otherVal))
			return
		}
		// Dereference pointers
		currentVal = currentVal.Elem()
		otherVal = otherVal.Elem()
		fieldType = fieldType.Elem()
	}

	// Fast path for common primitive types
	switch fieldType.Kind() {
	case reflect.String:
		if currentVal.String() != otherVal.String() {
			addDifference(differences, paramPath, otherVal.String())
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if currentVal.Int() != otherVal.Int() {
			addDifference(differences, paramPath, strconv.FormatInt(otherVal.Int(), 10))
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if currentVal.Uint() != otherVal.Uint() {
			addDifference(differences, paramPath, strconv.FormatUint(otherVal.Uint(), 10))
		}
	case reflect.Float32, reflect.Float64:
		if currentVal.Float() != otherVal.Float() {
			addDifference(differences, paramPath, strconv.FormatFloat(otherVal.Float(), 'f', -1, 64))
		}
	case reflect.Bool:
		if currentVal.Bool() != otherVal.Bool() {
			addDifference(differences, paramPath, strconv.FormatBool(otherVal.Bool()))
		}
	case reflect.Slice:
		compareSliceFields(currentVal, otherVal, paramPath, differences, visited)
	case reflect.Struct:
		// Recursively compare nested structs
		compareStructFields(currentVal, otherVal, fieldType, paramPath, differences, visited)
	default:
		// Fallback for other types using string representation
		currentStr := getStringValue(currentVal)
		otherStr := getStringValue(otherVal)
		if currentStr != otherStr {
			addDifference(differences, paramPath, otherStr)
		}
	}
}

// compareSliceFields efficiently compares slice fields with index-based parameter names
// Uses sparse array approach similar to parameter indexer for proper TR-069 indexing
func compareSliceFields(currentVal, otherVal reflect.Value, pathPrefix string, differences *[]ParameterValueStruct, visited map[reflect.Type]bool) {
	currentLen := currentVal.Len()
	otherLen := otherVal.Len()
	maxLen := currentLen
	if otherLen > maxLen {
		maxLen = otherLen
	}

	// Pre-allocate string builder for index paths
	var pathBuilder strings.Builder
	baseLen := len(pathPrefix) + 10 // Reserve space for index

	elementType := reflect.TypeOf((*interface{})(nil)).Elem()
	if currentLen > 0 {
		elementType = currentVal.Index(0).Type()
	} else if otherLen > 0 {
		elementType = otherVal.Index(0).Type()
	}

	// Handle struct elements with Index field (sparse array approach like parameter indexer)
	if elementType.Kind() == reflect.Struct {
		if indexField, hasIndex := elementType.FieldByName("Index"); hasIndex && indexField.Type.Kind() == reflect.Int {
			compareSliceWithIndexField(currentVal, otherVal, pathPrefix, differences, visited)
			return
		}
	}

	// Traditional array comparison for non-indexed elements
	for i := 0; i < maxLen; i++ {
		pathBuilder.Reset()
		pathBuilder.Grow(baseLen)
		pathBuilder.WriteString(pathPrefix)
		pathBuilder.WriteByte('.')
		pathBuilder.WriteString(strconv.Itoa(i + 1)) // TR-069 uses 1-based indexing
		indexPath := pathBuilder.String()

		if i >= currentLen {
			// New element in other slice
			if i < otherLen {
				otherElem := otherVal.Index(i)
				addDifference(differences, indexPath, getStringValue(otherElem))
			}
		} else if i >= otherLen {
			// Element removed from other slice - could add deletion logic here
			// For now, we'll skip removed elements as SetParameterValues typically sets new values
		} else {
			// Compare existing elements
			currentElem := currentVal.Index(i)
			otherElem := otherVal.Index(i)
			compareFieldValues(currentElem, otherElem, indexPath, currentElem.Type(), differences, visited)
		}
	}
}

// compareSliceWithIndexField handles slices of structs with Index fields (like TR-069 sparse arrays)
func compareSliceWithIndexField(currentVal, otherVal reflect.Value, pathPrefix string, differences *[]ParameterValueStruct, visited map[reflect.Type]bool) {
	// Build maps of existing elements by their Index field values
	currentMap := make(map[int]reflect.Value)
	otherMap := make(map[int]reflect.Value)

	// Populate current elements map
	for i := 0; i < currentVal.Len(); i++ {
		elem := currentVal.Index(i)
		if elem.Kind() == reflect.Struct {
			indexField := elem.FieldByName("Index")
			if indexField.IsValid() && indexField.Kind() == reflect.Int {
				index := int(indexField.Int())
				currentMap[index] = elem
			}
		}
	}

	// Populate other elements map
	for i := 0; i < otherVal.Len(); i++ {
		elem := otherVal.Index(i)
		if elem.Kind() == reflect.Struct {
			indexField := elem.FieldByName("Index")
			if indexField.IsValid() && indexField.Kind() == reflect.Int {
				index := int(indexField.Int())
				otherMap[index] = elem
			}
		}
	}

	// Find all unique indices
	allIndices := make(map[int]bool)
	for index := range currentMap {
		allIndices[index] = true
	}
	for index := range otherMap {
		allIndices[index] = true
	}

	// Compare elements by their TR-069 indices
	for index := range allIndices {
		indexPath := fmt.Sprintf("%s.%d", pathPrefix, index)

		currentElem, hasCurrentElem := currentMap[index]
		otherElem, hasOtherElem := otherMap[index]

		if !hasCurrentElem && hasOtherElem {
			// New element in other
			addDifference(differences, indexPath, getStringValue(otherElem))
		} else if hasCurrentElem && hasOtherElem {
			// Compare existing elements
			compareFieldValues(currentElem, otherElem, indexPath, currentElem.Type(), differences, visited)
		}
		// Skip removed elements (hasCurrentElem && !hasOtherElem)
	}
}

// addDifference efficiently adds a difference to the slice
// Inlined for performance in hot path
func addDifference(differences *[]ParameterValueStruct, name, value string) {
	*differences = append(*differences, ParameterValueStruct{
		Name:  name,
		Value: value,
	})
}

// getStringValue efficiently converts reflect.Value to string representation
// Optimized for common types to avoid expensive reflection calls
func getStringValue(val reflect.Value) string {
	if !val.IsValid() {
		return ""
	}

	// Handle nil pointers
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return ""
		}
		val = val.Elem()
	}

	// Fast path for common types
	switch val.Kind() {
	case reflect.String:
		return val.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(val.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(val.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(val.Float(), 'f', -1, 64)
	case reflect.Bool:
		return strconv.FormatBool(val.Bool())
	default:
		// Fallback using fmt.Sprintf for complex types
		return fmt.Sprintf("%v", val.Interface())
	}
}
