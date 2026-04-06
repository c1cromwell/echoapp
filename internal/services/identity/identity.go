package identity

// UserIdentity represents a user's identity information
type UserIdentity struct {
	ID                string
	PhoneHash         string
	VerificationLevel int
	Status            string
}

// IdentityService manages user authentication
type IdentityService struct {
	users map[string]*UserIdentity
}

// NewIdentityService creates a new service
func NewIdentityService() *IdentityService {
	return &IdentityService{
		users: make(map[string]*UserIdentity),
	}
}

// RegisterUser creates a new user identity
func (is *IdentityService) RegisterUser(phoneHash string) (*UserIdentity, error) {
	user := &UserIdentity{
		ID:                "user-" + phoneHash[:4],
		PhoneHash:         phoneHash,
		VerificationLevel: 0,
		Status:            "active",
	}
	is.users[user.ID] = user
	return user, nil
}

// GetUser retrieves a user
func (is *IdentityService) GetUser(userID string) (*UserIdentity, error) {
	if user, exists := is.users[userID]; exists {
		return user, nil
	}
	return nil, ErrUserNotFound
}
