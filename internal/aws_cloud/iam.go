// Copyright (c) https://github.com/fd008, all rights reserved.
// SPDX-License-Identifier: MPL-2.0

package awscloud

import (
	"reflect" // To compare principals and resources accurately
	"sort"    // For consistent string slice sorting
)

// Policy represents an AWS IAM policy document
type Policy struct {
	Version   string      `json:"Version"`
	Statement []Statement `json:"Statement"`
}

// Statement represents a single IAM policy statement
type Statement struct {
	Effect    string      `json:"Effect"`
	Principal Principal   `json:"Principal,omitempty"`
	Action    interface{} `json:"Action"`   // Can be string or []string
	Resource  interface{} `json:"Resource"` // Can be string or []string
	Condition interface{} `json:"Condition,omitempty"`
	Sid       string      `json:"Sid,omitempty"` // Optional statement ID
}

// Principal represents the entity allowed or denied access
type Principal struct {
	AWS           interface{} `json:"AWS,omitempty"` // Can be string or []string
	Federated     interface{} `json:"Federated,omitempty"`
	Service       interface{} `json:"Service,omitempty"` // Can be string or []string
	CanonicalUser string      `json:"CanonicalUser,omitempty"`
}

// MergePolicies merges two IAM policies based on principal and resource.
// Actions are only merged if both principal and resource match.
// Otherwise, statements are added individually.
func MergePolicies(policy1, policy2 Policy) Policy {
	mergedPolicy := policy1

	for _, stmt2 := range policy2.Statement {
		foundMatch := false
		for i, stmt1 := range mergedPolicy.Statement {
			// Compare principals AND resources
			if reflect.DeepEqual(stmt1.Principal, stmt2.Principal) &&
				reflect.DeepEqual(stmt1.Resource, stmt2.Resource) {
				// Principals and Resources match, merge actions and conditions
				mergedPolicy.Statement[i].Action = mergeStringOrStringSlice(stmt1.Action, stmt2.Action)
				mergedPolicy.Statement[i].Condition = mergeConditions(stmt1.Condition, stmt2.Condition) // Add this line
				foundMatch = true
				break
			}
		}

		// If no matching principal AND resource were found, add the statement from policy2
		if !foundMatch {
			mergedPolicy.Statement = append(mergedPolicy.Statement, stmt2)
		}
	}

	return mergedPolicy
}

// Helper function to merge string or string slice values (Actions/Resources)
func mergeStringOrStringSlice(val1, val2 interface{}) interface{} {
	// Handle nil values gracefully
	if val1 == nil && val2 == nil {
		return nil
	}
	if val1 == nil {
		return val2
	}
	if val2 == nil {
		return val1
	}

	var s1, s2 []string

	// Convert interfaces to string slices for processing
	s1 = interfaceToStringSlice(val1)
	s2 = interfaceToStringSlice(val2)

	// Create a map to store unique values
	uniqueValues := make(map[string]bool)
	for _, s := range s1 {
		uniqueValues[s] = true
	}
	for _, s := range s2 {
		uniqueValues[s] = true
	}

	// Convert back to a slice
	var merged []string
	for val := range uniqueValues {
		merged = append(merged, val)
	}

	// Sort for consistent output, important for DeepEqual on resources later
	sort.Strings(merged)

	// If there's only one item, return it as a string to match IAM policy JSON format
	if len(merged) == 1 {
		return merged[0]
	}

	return merged
}

// Helper to convert interface{} to []string
func interfaceToStringSlice(val interface{}) []string {
	var result []string
	if str, ok := val.(string); ok {
		result = []string{str}
	} else if slice, ok := val.([]interface{}); ok {
		for _, v := range slice {
			if str, ok := v.(string); ok {
				result = append(result, str)
			}
		}
	} else if slice, ok := val.([]string); ok {
		result = slice
	}
	return result
}

// mergeConditions merges two IAM condition blocks.
// This is a simplified example, real-world merging of conditions can be highly complex
// due to various operators (StringEquals, NumericGreaterThan, ForAllValues:StringEquals, etc.)
// and nested structures.
//
// For this example, we assume conditions are represented as map[string]interface{}
// and merge them by combining keys. Conflicts will overwrite.
func mergeConditions(cond1, cond2 interface{}) interface{} {
	// If both are nil, return nil
	if cond1 == nil && cond2 == nil {
		return nil
	}
	// If one is nil, return the other
	if cond1 == nil {
		return cond2
	}
	if cond2 == nil {
		return cond1
	}

	// Attempt to convert conditions to map[string]interface{}
	map1, ok1 := cond1.(map[string]interface{})
	map2, ok2 := cond2.(map[string]interface{})

	// If both are maps, proceed with merging them
	if ok1 && ok2 {
		mergedMap := make(map[string]interface{})
		// Copy all from map1
		for k, v := range map1 {
			mergedMap[k] = v
		}
		// Copy or overwrite from map2
		for k, v := range map2 {
			// **Here's a simplified merge strategy: Overwrite on conflict.**
			// In a real-world scenario, you'd need more sophisticated logic
			// for different condition operators (e.g., merging arrays for ForAllValues).
			mergedMap[k] = v
		}
		return mergedMap
	}

	// If conditions are not both maps (e.g., they're strings, or different types)
	// it's difficult to merge them meaningfully without more specific logic.
	// For this example, we'll prioritize cond1.
	// You might want to handle this case more carefully, potentially returning an error
	// or employing a different conflict resolution strategy based on your needs.
	return cond1
}
