package converter

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/morpho-app/lexsync/internal/utils"
	"go.uber.org/zap"
)

// ProcessLexiconJSON reads a JSON file and generates corresponding Kotlin code
func ProcessLexiconJSON(jsonFilePath string, sugar *zap.SugaredLogger) (string, error) {
	sugar.Infof("Processing JSON file: %s", jsonFilePath)

	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		sugar.Errorf("Failed to read JSON file %s: %v", jsonFilePath, err)
		return "", fmt.Errorf("failed to read JSON file %s: %v", jsonFilePath, err)
	}

	var lexiconData map[string]interface{}
	if err := json.Unmarshal(jsonData, &lexiconData); err != nil {
		sugar.Errorf("Failed to unmarshal JSON file %s: %v", jsonFilePath, err)
		return "", fmt.Errorf("failed to unmarshal JSON file %s: %v", jsonFilePath, err)
	}

	defs, ok := lexiconData["defs"].(map[string]interface{})
	if !ok {
		sugar.Errorf("No 'defs' key found in %s", jsonFilePath)
		return "", fmt.Errorf("no 'defs' key found in %s", jsonFilePath)
	}

	if len(defs) == 0 {
		sugar.Errorf("'defs' is empty in %s", jsonFilePath)
		return "", fmt.Errorf("'defs' is empty in %s", jsonFilePath)
	}

	var kotlinClasses strings.Builder

	// Collect and sort definition names
	defNames := make([]string, 0, len(defs))
	for defName := range defs {
		defNames = append(defNames, defName)
	}
	sort.Strings(defNames)

	for i, defName := range defNames {
		defValue := defs[defName]
		defMap, ok := defValue.(map[string]interface{})
		if !ok {
			sugar.Warnf("Definition %s is not a valid object. Skipping.", defName)
			continue
		}

		className := utils.ToCamelCase(defName, true) // PascalCase for class names
		kotlinCode, err := GenerateKotlinDataClass(className, defMap, sugar)
		if err != nil {
			sugar.Warnf("Error generating Kotlin class for %s: %v", defName, err)
			continue
		}

		kotlinClasses.WriteString(kotlinCode)

		// Add a newline between classes except after the last one
		if i < len(defNames)-1 {
			kotlinClasses.WriteString("\n")
		}
	}

	return kotlinClasses.String(), nil
}

// GenerateKotlinDataClass generates Kotlin data class code from a JSON definition
func GenerateKotlinDataClass(className string, defMap map[string]interface{}, sugar *zap.SugaredLogger) (string, error) {
	properties, ok := defMap["properties"].(map[string]interface{})
	if !ok {
		sugar.Warnf("No 'properties' found for class %s. Generating empty class.", className)
		return fmt.Sprintf("data class %s()\n", className), nil
	}

	// Handle required fields
	requiredFields := map[string]bool{}
	if required, exists := defMap["required"].([]interface{}); exists {
		for _, field := range required {
			if fieldStr, ok := field.(string); ok {
				requiredFields[fieldStr] = true
			}
		}
	}

	var fields []string
	var enums []string

	// Collect and sort property names
	propNames := make([]string, 0, len(properties))
	for propName := range properties {
		propNames = append(propNames, propName)
	}
	sort.Strings(propNames)

	for _, propName := range propNames {
		propValue := properties[propName]
		propMap, ok := propValue.(map[string]interface{})
		if !ok {
			sugar.Warnf("Property %s in class %s is not a valid object. Skipping.", propName, className)
			continue
		}

		// Determine the Kotlin type and JSON type
		kotlinType, jsonType := MapJSONTypeToKotlin(propMap, sugar)
		isRequired := requiredFields[propName]

		// Make the type nullable if not required
		if !isRequired {
			kotlinType += "?"
		}

		// Handle default values (if any)
		defaultValue := ""
		if defVal, exists := propMap["default"]; exists {
			defaultValue = FormatDefaultValue(defVal, jsonType, sugar)
		}

		// Generate Kotlin field with optional default value
		camelCasePropName := utils.ToCamelCase(propName, false) // camelCase for property names
		fields = append(fields, fmt.Sprintf("    val %s: %s%s", camelCasePropName, kotlinType, defaultValue))
	}

	// Assemble enums if any
	enumSection := ""
	if len(enums) > 0 {
		enumSection = strings.Join(enums, "\n") + "\n\n"
	}

	// Assemble the Kotlin data class
	kotlinClass := fmt.Sprintf("%sdata class %s(\n%s\n)\n", enumSection, className, strings.Join(fields, ",\n"))
	return kotlinClass, nil
}

func FormatDefaultValue(defVal interface{}, jsonType string, sugar *zap.SugaredLogger) string {
	switch jsonType {
	case "string":
		if v, ok := defVal.(string); ok {
			return fmt.Sprintf(" = \"%s\"", v)
		}
	case "boolean":
		if v, ok := defVal.(bool); ok {
			return fmt.Sprintf(" = %t", v)
		}
	case "integer":
		if v, ok := defVal.(float64); ok {
			return fmt.Sprintf(" = %d", int(v))
		}
	case "number", "float", "double":
		if v, ok := defVal.(float64); ok {
			return fmt.Sprintf(" = %f", v)
		}
	case "array":
		// Handle default arrays if necessary
		sugar.Warn("Default values for arrays are not supported yet.")
	case "object":
		// Handle default objects if necessary
		sugar.Warn("Default values for objects are not supported yet.")
	default:
		sugar.Warnf("Unsupported JSON type '%s' for default value", jsonType)
	}
	return ""
}

// GenerateKotlinEnum generates a Kotlin enum class from a list of values
func GenerateKotlinEnum(enumName string, values []string, sugar *zap.SugaredLogger) string {
	if len(values) == 0 {
		sugar.Warnf("No enum values provided for %s. Skipping enum generation.", enumName)
		return ""
	}

	var enumEntries []string
	for _, val := range values {
		enumEntries = append(enumEntries, val)
	}
	return fmt.Sprintf("enum class %s {\n    %s\n}\n", enumName, strings.Join(enumEntries, ",\n    "))
}

// MapJSONTypeToKotlin maps JSON types to Kotlin types.
func MapJSONTypeToKotlin(propMap map[string]interface{}, sugar *zap.SugaredLogger) (kotlinType string, jsonType string) {
	if propType, exists := propMap["type"].(string); exists {
		jsonType = propType
		switch propType {
		case "string":
			kotlinType = "String"
		case "integer":
			kotlinType = "Int"
		case "number", "float":
			kotlinType = "Float"
		case "double":
			kotlinType = "Double"
		case "boolean":
			kotlinType = "Boolean"
		case "array":
			items, hasItems := propMap["items"].(map[string]interface{})
			if hasItems {
				itemType, _ := MapJSONTypeToKotlin(items, sugar)
				kotlinType = fmt.Sprintf("List<%s>", itemType)
			} else {
				kotlinType = "List<Any>"
			}
		case "object":
			kotlinType = "Map<String, Any>"
		default:
			sugar.Warnf("Unknown JSON type '%s', defaulting to 'Any'", propType)
			kotlinType = "Any"
		}
	} else if ref, exists := propMap["$ref"].(string); exists {
		// Handle references
		className := ExtractClassNameFromRef(ref)
		kotlinType = className
		jsonType = "object" // Assuming references are to objects
	} else {
		sugar.Warn("No 'type' or '$ref' found in property")
		kotlinType = "Any"
		jsonType = "any"
	}
	return
}

func ExtractClassNameFromRef(ref string) string {
	// Handle references in the format "#/defs/ClassName"
	parts := strings.Split(ref, "/")
	className := parts[len(parts)-1]
	className = utils.ToCamelCase(className, true)
	return className
}
