package skill

// Skill describes one discoverable skill directory in a skills repository.
type Skill struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Path        string `json:"path"`
}
