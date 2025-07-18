// Package muci implements a basic parser and handler for OpenWrt UCI configuration files.
// It supports loading a configuration file, retrieving values with Get,
// updating or adding values with Set, and saving the configuration back to disk.
//
// The UCI file format (as described on the OpenWrt documentation)
// is assumed to have directives like:
//
//	config <type> '<name>'
//	  option <key> '<value>'
//	  list <key> '<value>'
//
// Comments (lines starting with "#") and blank lines are ignored.
//
// This is a starting point. Depending on your needs you may wish
// to expand this package with more advanced type support, validation,
// or error handling.
package uci

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// Section represents a UCI configuration section.
type Section struct {
	SectionType string              // e.g. "interface", "firewall"
	Name        string              // The section identifier (if provided)
	Options     map[string]string   // Singular options (option keyword)
	Lists       map[string][]string // List options (list keyword) that may have multiple values
}

// UCIConfig represents the configuration loaded from a UCI file.
type UCIConfig struct {
	Filename string
	Package  string // The package name (e.g. "Device")
	Sections []*Section
}

// UCI holds a collection of UCI configuration files,
// typically keyed by a config name (e.g. "network" for /etc/config/network).
type UCI struct {
	Configs map[string]*UCIConfig
}

// NewUCI creates a new UCI instance.
func NewUCI() *UCI {
	return &UCI{
		Configs: make(map[string]*UCIConfig),
	}
}

// LoadConfig loads a UCI configuration file from the given file path,
// parses it, and returns a UCIConfig representation.
// If the file doesn't exist, an empty config is created.
// If packageName is provided, it will be set as the root package name.
func LoadConfig(filePath string, packageName *string) (*UCIConfig, error) {
	cfg := &UCIConfig{
		Filename: filePath,
		Package:  "",
		Sections: make([]*Section, 0),
	}

	// Check if file exists
	file, err := os.Open(filePath)
	if err != nil {
		// If file doesn't exist, create empty config and proceed
		if os.IsNotExist(err) {
			// Set package name if provided
			if packageName != nil && *packageName != "" {
				cfg.Package = *packageName
			}

			return cfg, nil
		}
		return nil, fmt.Errorf("failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var currentSection *Section

	// Precompile regular expressions for each directive.
	// package: package <name>
	packageRe := regexp.MustCompile(`^package\s+(\S+)`)
	// config: config <type> "<name>" (with quotes for name)
	configRe := regexp.MustCompile(`^config\s+(\S+)(?:\s+"([^"]+)")?`)
	// option: option <key> <value> (without quotes)
	optionRe := regexp.MustCompile(`^option\s+(\S+)\s+(.+)`)
	// list: list <key> <value> (without quotes)
	listRe := regexp.MustCompile(`^list\s+(\S+)\s+(.+)`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if matches := packageRe.FindStringSubmatch(line); matches != nil {
			// Set the package name
			cfg.Package = matches[1]
		} else if matches := configRe.FindStringSubmatch(line); matches != nil {
			// Begin a new section.
			sectionType := matches[1]
			sectionName := ""
			if len(matches) > 2 && matches[2] != "" {
				sectionName = matches[2]
			}
			currentSection = &Section{
				SectionType: sectionType,
				Name:        sectionName,
				Options:     make(map[string]string),
				Lists:       make(map[string][]string),
			}
			cfg.Sections = append(cfg.Sections, currentSection)
		} else if matches := optionRe.FindStringSubmatch(line); matches != nil {
			if currentSection == nil {
				return nil, fmt.Errorf("found option outside of a section: %s", line)
			}
			optName := matches[1]
			optValue := matches[2]
			currentSection.Options[optName] = optValue
		} else if matches := listRe.FindStringSubmatch(line); matches != nil {
			if currentSection == nil {
				return nil, fmt.Errorf("found list outside of a section: %s", line)
			}
			listKey := matches[1]
			listValue := matches[2]
			currentSection.Lists[listKey] = append(currentSection.Lists[listKey], listValue)
		} else {
			// Optionally, handle additional directives as needed.
			// For now, unrecognized lines are ignored.
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file %s: %v", filePath, err)
	}

	// Set package name if provided
	if packageName != nil && *packageName != "" {
		cfg.Package = *packageName
	}

	return cfg, nil
}

// Get retrieves the value of an option from a given section within the UCIConfig.
// The function looks for a section identified by its type only (since sections don't have names in the new format).
// It first checks for a singular option. If not found, it then looks for a list option.
// The return type is interface{}: it can be either a string (for option)
// or a []string (for a list). An error is returned if the section or key is not found.
func (cfg *UCIConfig) Get(sectionType, key string) (interface{}, error) {
	for _, sec := range cfg.Sections {
		if sec.SectionType == sectionType {
			// Check singular options first.
			if val, ok := sec.Options[key]; ok {
				return val, nil
			}
			// Then check list options.
			if listVal, ok := sec.Lists[key]; ok {
				return listVal, nil
			}
			return nil, fmt.Errorf("key %s not found in section '%s'", key, sectionType)
		}
	}
	return nil, fmt.Errorf("section '%s' not found", sectionType)
}

// Set updates or adds an option value in a specified section within the UCIConfig.
// The caller can specify whether the option should be handled as a list (isList == true)
// or as a singular option (isList == false). If the section is not found, it will be created.
func (cfg *UCIConfig) Set(sectionType, key, value string, isList bool) error {
	for _, sec := range cfg.Sections {
		if sec.SectionType == sectionType {
			if isList {
				// Append to an existing list or create a new one.
				sec.Lists[key] = append(sec.Lists[key], value)
			} else {
				sec.Options[key] = value
			}
			return nil
		}
	}

	// Section doesn't exist, create it
	newSection := &Section{
		SectionType: sectionType,
		Name:        sectionType, // Use sectionType as name too
		Options:     make(map[string]string),
		Lists:       make(map[string][]string),
	}

	if isList {
		newSection.Lists[key] = append(newSection.Lists[key], value)
	} else {
		newSection.Options[key] = value
	}

	cfg.Sections = append(cfg.Sections, newSection)
	return nil
}

// addSection function to add section in the configuration file
// if section doesnt exit we add one
func (cfg *UCIConfig) AddSection(sectionType, sectionName string) error {
	for _, sec := range cfg.Sections {
		if sec.SectionType == sectionType && sec.Name == sectionName {
			return fmt.Errorf("section '%s' of type '%s' already exists", sectionName, sectionType)
		}
	}
	newSection := &Section{
		SectionType: sectionType,
		Name:        sectionName,
		Options:     make(map[string]string),
		Lists:       make(map[string][]string),
	}
	cfg.Sections = append(cfg.Sections, newSection)
	return nil
}

// Save writes the UCIConfig back to its original file.
// It will rewrite the configuration based on the current in-memory state.
// Use caution as this overwrites the original file.
func (cfg *UCIConfig) Save() error {
	file, err := os.Create(cfg.Filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s for writing: %v", cfg.Filename, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	// Write package header if it exists
	if cfg.Package != "" {
		_, err := fmt.Fprintf(writer, "package %s\n", cfg.Package)
		if err != nil {
			return err
		}
	}

	for _, sec := range cfg.Sections {
		// Write the config header with quotes around the name
		if sec.Name != "" {
			_, err := fmt.Fprintf(writer, "config %s \"%s\"\n", sec.SectionType, sec.Name)
			if err != nil {
				return err
			}
		} else {
			_, err := fmt.Fprintf(writer, "config %s\n", sec.SectionType)
			if err != nil {
				return err
			}
		}

		// Write singular options without quotes
		for opt, val := range sec.Options {
			_, err := fmt.Fprintf(writer, "\toption %s %s\n", opt, val)
			if err != nil {
				return err
			}
		}
		// Write list options without quotes
		for key, listValues := range sec.Lists {
			for _, val := range listValues {
				_, err := fmt.Fprintf(writer, "\tlist %s %s\n", key, val)
				if err != nil {
					return err
				}
			}
		}
		// Separate sections with a blank line.
		_, err = writer.WriteString("\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}
