package models

import "time"

// Skill represents a skill entity stored in the database.
type Skill struct {
    Key         string    `json:"key"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    Logo        string    `json:"logo"`
    Tags        []string  `json:"tags"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// SkillInput captures payloads for create/update requests.
type SkillInput struct {
    Key         string   `json:"key"`
    Name        string   `json:"name"`
    Description string   `json:"description"`
    Logo        string   `json:"logo"`
    Tags        []string `json:"tags"`
}
