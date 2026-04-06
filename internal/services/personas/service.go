package personas

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrMasterIdentityNotFound = errors.New("master identity not found")
	ErrPersonaNotFound        = errors.New("persona not found")
	ErrPersonaLimitExceeded   = errors.New("persona limit exceeded")
	ErrDuplicateUsername      = errors.New("username already in use")
	ErrDuplicateDisplayName   = errors.New("display name already in use")
	ErrUnauthorized           = errors.New("unauthorized")
	ErrAccessDenied           = errors.New("access denied")
	ErrInvalidPersonaCategory = errors.New("invalid persona category")
	ErrInsufficientTrustLevel = errors.New("insufficient trust level")
	ErrGrantNotFound          = errors.New("grant not found")
	ErrGrantExpired           = errors.New("grant expired")
)

type PersonaService struct {
	masters    map[string]*MasterIdentity
	personas   map[string]*Persona
	byUsername map[string]string
	grants     map[string]*AccessGrant
	usageCount map[string]int
}

func NewPersonaService() *PersonaService {
	return &PersonaService{
		masters:    make(map[string]*MasterIdentity),
		personas:   make(map[string]*Persona),
		byUsername: make(map[string]string),
		grants:     make(map[string]*AccessGrant),
		usageCount: make(map[string]int),
	}
}

func (ps *PersonaService) CreateMasterIdentity(userID, did, masterKey string, trustLevel string) (*MasterIdentity, error) {
	if _, exists := ps.masters[userID]; exists {
		return nil, fmt.Errorf("master identity already exists for user %s", userID)
	}

	master := &MasterIdentity{
		UserID:     userID,
		DID:        did,
		MasterKey:  masterKey,
		TrustScore: 0,
		CoreVerification: VerificationStatus{
			KYCVerified:   false,
			PhoneVerified: false,
			EmailVerified: false,
		},
		Personas:  make([]string, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ps.masters[userID] = master
	ps.usageCount[userID] = 0

	return master, nil
}

func (ps *PersonaService) GetMasterIdentity(userID string) (*MasterIdentity, error) {
	master, exists := ps.masters[userID]
	if !exists {
		return nil, ErrMasterIdentityNotFound
	}
	return master, nil
}

func (ps *PersonaService) CreatePersona(
	userID, displayName, username, avatar, bio string,
	category PersonaCategory,
	trustLevel string,
) (*Persona, error) {
	master, err := ps.GetMasterIdentity(userID)
	if err != nil {
		return nil, err
	}

	limits, exists := PersonaLimits[trustLevel]
	if !exists {
		return nil, ErrInsufficientTrustLevel
	}

	// Check username uniqueness BEFORE checking limit
	if _, exists := ps.byUsername[username]; exists {
		return nil, ErrDuplicateUsername
	}

	if ps.usageCount[userID] >= limits.MaxPersonas {
		return nil, ErrPersonaLimitExceeded
	}

	if !isValidCategory(category, limits.CustomCategories) {
		return nil, ErrInvalidPersonaCategory
	}

	personaID := fmt.Sprintf("persona_%s_%d", userID, len(master.Personas))
	derivationPath := fmt.Sprintf("m/867530'/%d'/0'", len(master.Personas))

	persona := &Persona{
		PersonaID:         personaID,
		MasterDID:         master.DID,
		UserID:            userID,
		DisplayName:       displayName,
		Username:          username,
		Avatar:            avatar,
		Bio:               bio,
		Category:          category,
		Status:            "active",
		CreatedAt:         time.Now(),
		LastActiveAt:      time.Now(),
		SigningKey:        generateKey(derivationPath + "/0"),
		EncryptionKey:     generateKey(derivationPath + "/1"),
		DerivationPath:    derivationPath,
		DefaultVisibility: VisibilityContacts,
		Discoverability:   true,
		Privacy:           DefaultPrivacySettings(),
		Notifications:     DefaultNotificationSettings(),
		Features:          DefaultFeatureSettings(),
		AccessListIDs:     make([]string, 0),
		Badges:            make([]string, 0),
		Credentials:       make([]string, 0),
	}

	ps.personas[personaID] = persona
	ps.byUsername[username] = personaID
	ps.usageCount[userID]++
	master.Personas = append(master.Personas, personaID)
	master.UpdatedAt = time.Now()

	return persona, nil
}

func (ps *PersonaService) GetPersona(personaID string) (*Persona, error) {
	persona, exists := ps.personas[personaID]
	if !exists {
		return nil, ErrPersonaNotFound
	}
	return persona, nil
}

func (ps *PersonaService) GetUserPersonas(userID string) ([]*Persona, error) {
	_, err := ps.GetMasterIdentity(userID)
	if err != nil {
		return nil, err
	}

	var result []*Persona
	for _, persona := range ps.personas {
		if persona.UserID == userID {
			result = append(result, persona)
		}
	}
	return result, nil
}

func (ps *PersonaService) DeletePersona(personaID string) error {
	persona, err := ps.GetPersona(personaID)
	if err != nil {
		return err
	}

	master, err := ps.GetMasterIdentity(persona.UserID)
	if err != nil {
		return err
	}

	delete(ps.personas, personaID)
	delete(ps.byUsername, persona.Username)
	ps.usageCount[persona.UserID]--

	newPersonas := make([]string, 0)
	for _, pid := range master.Personas {
		if pid != personaID {
			newPersonas = append(newPersonas, pid)
		}
	}
	master.Personas = newPersonas
	master.UpdatedAt = time.Now()

	return nil
}

func (ps *PersonaService) UpdatePersonaProfile(
	personaID, displayName, bio, status string,
) (*Persona, error) {
	persona, err := ps.GetPersona(personaID)
	if err != nil {
		return nil, err
	}

	persona.DisplayName = displayName
	persona.Bio = bio
	persona.Status = status
	persona.LastActiveAt = time.Now()

	return persona, nil
}

func (ps *PersonaService) GrantAccess(
	personaID, contactID string,
	permissions AccessPermissions,
) (*AccessGrant, error) {
	persona, err := ps.GetPersona(personaID)
	if err != nil {
		return nil, err
	}

	grantID := fmt.Sprintf("grant_%s_%s_%d", personaID, contactID, time.Now().Unix())

	grant := &AccessGrant{
		GrantID:     grantID,
		PersonaID:   personaID,
		ContactID:   contactID,
		GrantedAt:   time.Now(),
		GrantedBy:   persona.UserID,
		Permissions: permissions,
		Revocable:   true,
	}

	ps.grants[grantID] = grant
	persona.AccessListIDs = append(persona.AccessListIDs, grantID)

	return ps.grants[grantID], nil
}

func (ps *PersonaService) RevokeAccess(grantID string) error {
	grant, exists := ps.grants[grantID]
	if !exists {
		return ErrGrantNotFound
	}

	if !grant.Revocable {
		return ErrUnauthorized
	}

	persona, err := ps.GetPersona(grant.PersonaID)
	if err != nil {
		return err
	}

	delete(ps.grants, grantID)

	newAccessList := make([]string, 0)
	for _, gid := range persona.AccessListIDs {
		if gid != grantID {
			newAccessList = append(newAccessList, gid)
		}
	}
	persona.AccessListIDs = newAccessList

	return nil
}

func (ps *PersonaService) CheckAccess(contactID, personaID string) (bool, *AccessGrant, error) {
	persona, err := ps.GetPersona(personaID)
	if err != nil {
		return false, nil, err
	}

	for _, grantID := range persona.AccessListIDs {
		grant, exists := ps.grants[grantID]
		if !exists {
			continue
		}
		if grant.ContactID == contactID {
			if grant.ExpiresAt != nil && grant.ExpiresAt.Before(time.Now()) {
				return false, nil, ErrGrantExpired
			}
			return true, grant, nil
		}
	}

	return false, nil, nil
}

func (ps *PersonaService) GetContactVisiblePersonas(userID, contactID string) ([]*Persona, error) {
	personas, err := ps.GetUserPersonas(userID)
	if err != nil {
		return nil, err
	}

	var visible []*Persona
	for _, persona := range personas {
		if persona.DefaultVisibility == VisibilityEveryone {
			visible = append(visible, persona)
			continue
		}

		allowed, _, _ := ps.CheckAccess(contactID, persona.PersonaID)
		if allowed {
			visible = append(visible, persona)
		}
	}

	return visible, nil
}

func (ps *PersonaService) UpdatePrivacySettings(personaID string, settings PersonaPrivacySettings) (*Persona, error) {
	persona, err := ps.GetPersona(personaID)
	if err != nil {
		return nil, err
	}

	persona.Privacy = settings
	return persona, nil
}

func (ps *PersonaService) UpdateNotificationSettings(personaID string, settings PersonaNotificationSettings) (*Persona, error) {
	persona, err := ps.GetPersona(personaID)
	if err != nil {
		return nil, err
	}

	persona.Notifications = settings
	return persona, nil
}

func (ps *PersonaService) PublicPersonaInfo(personaID string) (*Persona, error) {
	persona, err := ps.GetPersona(personaID)
	if err != nil {
		return nil, err
	}

	limited := &Persona{
		PersonaID:    persona.PersonaID,
		DisplayName:  persona.DisplayName,
		Avatar:       persona.Avatar,
		Username:     persona.Username,
		Category:     persona.Category,
		Status:       persona.Status,
		CreatedAt:    persona.CreatedAt,
		LastActiveAt: persona.LastActiveAt,
		MessageCount: persona.MessageCount,
	}

	return limited, nil
}

func isValidCategory(cat PersonaCategory, allowCustom bool) bool {
	validCategories := []PersonaCategory{
		CategoryProfessional,
		CategoryPersonal,
		CategoryFamily,
		CategoryGaming,
		CategoryDating,
		CategoryCreative,
		CategoryAnonymous,
	}

	for _, valid := range validCategories {
		if cat == valid {
			return true
		}
	}

	if allowCustom && cat == CategoryCustom {
		return true
	}

	return false
}

func generateKey(derivationPath string) string {
	return fmt.Sprintf("key_%s_%d", derivationPath, time.Now().UnixNano())
}
