package semgrep

import "encoding/json"

type Result struct {
	CheckID string `json:"check_id"`
	Extra   struct {
		Lines    string `json:"lines"`
		Message  string `json:"message"`
		Metadata struct {
			Asvs struct {
				ControlID  string `json:"control_id"`
				ControlURL string `json:"control_url"`
				Section    string `json:"section"`
				Version    string `json:"version"`
			} `json:"asvs"`
			Category           string              `json:"category" `
			Confidence         string              `json:"confidence"`
			Cwe                StringOrStringSlice `json:"cwe"`
			Impact             string              `json:"impact"`
			Likelihood         string              `json:"likelihood"`
			Owasp              StringOrStringSlice `json:"owasp" `
			References         StringOrStringSlice `json:"references"`
			VulnerabilityClass StringOrStringSlice `json:"vulnerability_class"`
		} `json:"metadata"`
	} `json:"extra"`
	Path string `json:"path"`
}

type StringOrStringSlice struct {
	Value []string
	Set   bool
}

// The UnmarshalJSON method on StringOrStringSlice will parse the JSON as either a single string or an array of strings into a slice of strings
func (s *StringOrStringSlice) UnmarshalJSON(b []byte) error {
	var strVal string
	var arrVal []string

	// Try to unmarshal into a single string first
	err := json.Unmarshal(b, &strVal)
	if err == nil {
		// No error, create a array with a single value
		arrVal = []string{strVal}
	} else {
		// Try to unmarshall into a slice of strings
		err = json.Unmarshal(b, &arrVal)
		if err != nil {
			// Both unmarshall were unsuccessful
			return err
		}
	}

	// Set the value of s to the unmarshalled value which will always be a slice
	s.Value = arrVal
	s.Set = true
	return nil
}
