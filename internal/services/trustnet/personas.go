package trustnet

import (
	"sync"

	"github.com/google/uuid"
)

// Persona represents a user identity/profile variation
type Persona struct {
	ID          string
	UserDID     string
	Name        string // e.g. "Professional", "Personal", "Gaming"
	DisplayName string // e.g. "Alex Echo", "Dad", "NightOwl42"
	Description string // e.g. "Product Lead", "Just me", "Gaming"
	Icon        string // emoji icon for the persona
	IsDefault   bool
}

// PersonaVisibility tracks which personas a specific contact can see
type PersonaVisibility struct {
	UserDID    string
	ContactDID string
	PersonaIDs map[string]bool // personaID -> visible
}

// PersonaService manages personas and per-contact visibility
type PersonaService struct {
	mu         sync.RWMutex
	personas   map[string]*Persona            // personaID -> persona
	byUser     map[string][]string            // userDID -> []personaID
	visibility map[string]*PersonaVisibility  // "userDID:contactDID" -> visibility
	circles    *CircleService
}

// NewPersonaService creates a new persona service
func NewPersonaService(circles *CircleService) *PersonaService {
	return &PersonaService{
		personas:   make(map[string]*Persona),
		byUser:     make(map[string][]string),
		visibility: make(map[string]*PersonaVisibility),
		circles:    circles,
	}
}

// CreatePersona creates a new persona for a user
func (s *PersonaService) CreatePersona(userDID, name, displayName, description, icon string, isDefault bool) (*Persona, error) {
	if name == "" || displayName == "" {
		return nil, ErrPersonaNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check for duplicate name
	for _, id := range s.byUser[userDID] {
		if s.personas[id].Name == name {
			return nil, ErrPersonaDuplicate
		}
	}

	// If this is first persona, make it default
	if len(s.byUser[userDID]) == 0 {
		isDefault = true
	}

	persona := &Persona{
		ID:          uuid.New().String(),
		UserDID:     userDID,
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Icon:        icon,
		IsDefault:   isDefault,
	}

	s.personas[persona.ID] = persona
	s.byUser[userDID] = append(s.byUser[userDID], persona.ID)
	return persona, nil
}

// GetPersona retrieves a persona by ID
func (s *PersonaService) GetPersona(personaID string) (*Persona, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.personas[personaID]
	if !ok {
		return nil, ErrPersonaNotFound
	}
	return p, nil
}

// GetUserPersonas returns all personas for a user
func (s *PersonaService) GetUserPersonas(userDID string) []*Persona {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Persona
	for _, id := range s.byUser[userDID] {
		if p, ok := s.personas[id]; ok {
			result = append(result, p)
		}
	}
	return result
}

// GetDefaultPersona returns the default persona for a user
func (s *PersonaService) GetDefaultPersona(userDID string) *Persona {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, id := range s.byUser[userDID] {
		p := s.personas[id]
		if p.IsDefault {
			return p
		}
	}
	return nil
}

// RemovePersona removes a persona (cannot remove default)
func (s *PersonaService) RemovePersona(personaID, userDID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.personas[personaID]
	if !ok {
		return ErrPersonaNotFound
	}

	if p.UserDID != userDID {
		return ErrPersonaNotFound
	}

	if p.IsDefault {
		return ErrPersonaDefaultOnly
	}

	delete(s.personas, personaID)

	// Remove from user's list
	ids := s.byUser[userDID]
	for i, id := range ids {
		if id == personaID {
			s.byUser[userDID] = append(ids[:i], ids[i+1:]...)
			break
		}
	}

	return nil
}

// SetPersonaVisibility sets which personas a contact can see
func (s *PersonaService) SetPersonaVisibility(userDID, contactDID string, personaIDs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := visKey(userDID, contactDID)
	vis := &PersonaVisibility{
		UserDID:    userDID,
		ContactDID: contactDID,
		PersonaIDs: make(map[string]bool),
	}
	for _, id := range personaIDs {
		vis.PersonaIDs[id] = true
	}
	s.visibility[key] = vis
}

// GetVisiblePersonas returns the personas a contact can see, respecting circle tier rules
func (s *PersonaService) GetVisiblePersonas(userDID, contactDID string) []*Persona {
	s.mu.RLock()
	defer s.mu.RUnlock()

	allPersonas := s.byUser[userDID]
	if len(allPersonas) == 0 {
		return nil
	}

	perms := s.circles.GetPermissions(userDID, contactDID)

	// Inner circle sees all personas
	if perms.SeeAllPersonas {
		var result []*Persona
		for _, id := range allPersonas {
			if p, ok := s.personas[id]; ok {
				result = append(result, p)
			}
		}
		return result
	}

	// Trusted sees selected personas (from explicit visibility settings)
	if perms.SeeSelectedPersonas {
		key := visKey(userDID, contactDID)
		vis, hasVis := s.visibility[key]

		var result []*Persona
		for _, id := range allPersonas {
			p := s.personas[id]
			if p == nil {
				continue
			}
			// Always include default
			if p.IsDefault {
				result = append(result, p)
				continue
			}
			// Include if explicitly granted
			if hasVis && vis.PersonaIDs[id] {
				result = append(result, p)
			}
		}
		return result
	}

	// Acquaintance / Public sees default persona only
	if perms.SeeDefaultPersona {
		for _, id := range allPersonas {
			p := s.personas[id]
			if p != nil && p.IsDefault {
				return []*Persona{p}
			}
		}
	}

	return nil
}

// TogglePersonaForContact toggles a specific persona's visibility for a contact
func (s *PersonaService) TogglePersonaForContact(userDID, contactDID, personaID string, visible bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Verify persona exists
	if _, ok := s.personas[personaID]; !ok {
		return ErrPersonaNotFound
	}

	key := visKey(userDID, contactDID)
	vis, ok := s.visibility[key]
	if !ok {
		vis = &PersonaVisibility{
			UserDID:    userDID,
			ContactDID: contactDID,
			PersonaIDs: make(map[string]bool),
		}
		s.visibility[key] = vis
	}

	if visible {
		vis.PersonaIDs[personaID] = true
	} else {
		delete(vis.PersonaIDs, personaID)
	}

	return nil
}

func visKey(userDID, contactDID string) string {
	return userDID + ":" + contactDID
}
