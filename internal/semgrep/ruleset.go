package semgrep

var (
	// Global variable to store the rulesets
	Rulesets map[string]Ruleset
)

type Ruleset struct {
	Name string `gorm:"type:text"`
}

// Register the default rulesets
func init() {
	Rulesets = map[string]Ruleset{
		"default":       {Name: "default"},
		"owasp-top-ten": {Name: "owasp-top-ten"},
		"cwe-top-25":    {Name: "cwe-top-25"},
		"java":          {Name: "java"},
		"javascript":    {Name: "javascript"},
		"nodejs":        {Name: "nodejs"},
		"php":           {Name: "php"},
		"python":        {Name: "python"},
		"react":         {Name: "react"},
		"typescript":    {Name: "typescript"},
	}
}

// URL returns the URL for the ruleset used by the --config parameter in Semgrep
func (r *Ruleset) URL() string {
	return "p/" + r.Name
}
