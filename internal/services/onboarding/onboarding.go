package onboarding

import (
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
)

// OnboardingStep represents a step in the onboarding flow
type OnboardingStep string

const (
	StepCarousel     OnboardingStep = "carousel"
	StepWelcome      OnboardingStep = "welcome"
	StepPhoneEntry   OnboardingStep = "phone_entry"
	StepOTP          OnboardingStep = "otp_verification"
	StepPasskey      OnboardingStep = "passkey_setup"
	StepRecovery     OnboardingStep = "recovery_setup"
	StepProfile      OnboardingStep = "profile_setup"
	StepTrustDash    OnboardingStep = "trust_dashboard"
	StepComplete     OnboardingStep = "complete"
)

// StepStatus tracks the completion state of each step
type StepStatus string

const (
	StatusPending   StepStatus = "pending"
	StatusActive    StepStatus = "active"
	StatusCompleted StepStatus = "completed"
	StatusSkipped   StepStatus = "skipped"
)

// RegistrationMethod indicates how the user is registering
type RegistrationMethod string

const (
	RegistrationPhone        RegistrationMethod = "phone"
	RegistrationVerifiableID RegistrationMethod = "verifiable_id"
)

// Profile holds the user's profile information collected during onboarding
type Profile struct {
	DisplayName string
	Username    string
	Bio         string
	AvatarURL   string
}

// SecuritySetupItem represents an item on the trust dashboard checklist
type SecuritySetupItem struct {
	Label     string
	Completed bool
	Action    string // empty if completed
	Points    int    // trust score points for completing
}

// OnboardingSession tracks a user's progress through the onboarding flow
type OnboardingSession struct {
	ID                 string
	PhoneNumber        string
	UserID             string
	DID                string
	RegistrationMethod RegistrationMethod
	Steps              map[OnboardingStep]StepStatus
	CurrentStep        OnboardingStep
	Profile            *Profile
	CreatedAt          time.Time
	UpdatedAt          time.Time
	CompletedAt        *time.Time
	CarouselViewed     bool
	SkippedSteps       []OnboardingStep
}

const (
	SessionExpiry      = 24 * time.Hour
	MaxDisplayNameLen  = 50
	MaxUsernameLen     = 30
	MaxBioLen          = 150
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// OnboardingService coordinates the multi-step registration flow
type OnboardingService struct {
	mu       sync.Mutex
	sessions map[string]*OnboardingSession // sessionID -> session
	phone    *PhoneVerificationService
	passkey  *PasskeyService
	recovery *RecoveryService

	// Track taken usernames
	usernames map[string]bool
}

// NewOnboardingService creates a new onboarding service with its dependencies
func NewOnboardingService(phone *PhoneVerificationService, passkey *PasskeyService, recovery *RecoveryService) *OnboardingService {
	return &OnboardingService{
		sessions:  make(map[string]*OnboardingSession),
		phone:     phone,
		passkey:   passkey,
		recovery:  recovery,
		usernames: make(map[string]bool),
	}
}

// StartSession begins a new onboarding flow
func (s *OnboardingService) StartSession(method RegistrationMethod) (*OnboardingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	session := &OnboardingSession{
		ID:                 uuid.New().String(),
		RegistrationMethod: method,
		CurrentStep:        StepCarousel,
		Steps: map[OnboardingStep]StepStatus{
			StepCarousel:   StatusActive,
			StepWelcome:    StatusPending,
			StepPhoneEntry: StatusPending,
			StepOTP:        StatusPending,
			StepPasskey:    StatusPending,
			StepRecovery:   StatusPending,
			StepProfile:    StatusPending,
			StepTrustDash:  StatusPending,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	s.sessions[session.ID] = session
	return session, nil
}

// GetSession retrieves an onboarding session
func (s *OnboardingService) GetSession(sessionID string) (*OnboardingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if time.Since(session.CreatedAt) > SessionExpiry {
		return nil, ErrSessionExpired
	}

	return session, nil
}

// CompleteCarousel marks the intro carousel as viewed
func (s *OnboardingService) CompleteCarousel(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return err
	}

	session.CarouselViewed = true
	session.Steps[StepCarousel] = StatusCompleted
	session.Steps[StepWelcome] = StatusActive
	session.CurrentStep = StepWelcome
	session.UpdatedAt = time.Now()
	return nil
}

// SkipCarousel skips the intro carousel
func (s *OnboardingService) SkipCarousel(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return err
	}

	session.Steps[StepCarousel] = StatusSkipped
	session.Steps[StepWelcome] = StatusActive
	session.CurrentStep = StepWelcome
	session.UpdatedAt = time.Now()
	return nil
}

// StartPhoneEntry transitions to the phone entry step
func (s *OnboardingService) StartPhoneEntry(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return err
	}

	if session.Steps[StepWelcome] != StatusActive && session.Steps[StepWelcome] != StatusCompleted {
		return ErrStepNotReady
	}

	session.Steps[StepWelcome] = StatusCompleted
	session.Steps[StepPhoneEntry] = StatusActive
	session.CurrentStep = StepPhoneEntry
	session.RegistrationMethod = RegistrationPhone
	session.UpdatedAt = time.Now()
	return nil
}

// SubmitPhone sends a verification code to the phone number
func (s *OnboardingService) SubmitPhone(sessionID, phoneNumber string) (*PhoneVerification, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.CurrentStep != StepPhoneEntry {
		return nil, ErrStepNotReady
	}

	verification, err := s.phone.SendCode(phoneNumber)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	session.PhoneNumber = phoneNumber
	session.Steps[StepPhoneEntry] = StatusCompleted
	session.Steps[StepOTP] = StatusActive
	session.CurrentStep = StepOTP
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	return verification, nil
}

// VerifyOTP verifies the OTP code and advances to passkey setup
func (s *OnboardingService) VerifyOTP(sessionID, code string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	if session.CurrentStep != StepOTP {
		return ErrStepNotReady
	}

	if err := s.phone.VerifyCode(session.PhoneNumber, code); err != nil {
		return err
	}

	s.mu.Lock()
	session.Steps[StepOTP] = StatusCompleted
	session.Steps[StepPasskey] = StatusActive
	session.CurrentStep = StepPasskey
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	return nil
}

// SetupPasskey initiates passkey creation for the session
func (s *OnboardingService) SetupPasskey(sessionID string) (*PasskeyChallenge, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.CurrentStep != StepPasskey {
		return nil, ErrStepNotReady
	}

	// Use session ID as the temporary user ID for challenge
	challenge, err := s.passkey.CreateChallenge(session.ID)
	if err != nil {
		return nil, err
	}

	return challenge, nil
}

// CompletePasskey registers the passkey credential and generates DID
func (s *OnboardingService) CompletePasskey(sessionID, challengeID, credentialID string, publicKey []byte, passkeyType PasskeyType, deviceInfo string) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	if session.CurrentStep != StepPasskey {
		return ErrStepNotReady
	}

	cred, err := s.passkey.RegisterCredential(challengeID, credentialID, publicKey, passkeyType, deviceInfo)
	if err != nil {
		return err
	}

	s.mu.Lock()
	session.DID = GenerateDID(cred.PublicKey)
	session.UserID = cred.UserID
	session.Steps[StepPasskey] = StatusCompleted
	session.Steps[StepRecovery] = StatusActive
	session.CurrentStep = StepRecovery
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	return nil
}

// SkipPasskey skips passkey setup (flagged as incomplete)
func (s *OnboardingService) SkipPasskey(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return err
	}

	if session.CurrentStep != StepPasskey {
		return ErrStepNotReady
	}

	session.Steps[StepPasskey] = StatusSkipped
	session.SkippedSteps = append(session.SkippedSteps, StepPasskey)
	session.Steps[StepRecovery] = StatusActive
	session.CurrentStep = StepRecovery
	session.UpdatedAt = time.Now()
	return nil
}

// SetupRecoveryEmail adds an email recovery method
func (s *OnboardingService) SetupRecoveryEmail(sessionID, email string) (*RecoveryMethod, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.CurrentStep != StepRecovery {
		return nil, ErrStepNotReady
	}

	userID := session.ID // Use session ID until proper user ID is assigned
	method, err := s.recovery.AddEmail(userID, email)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	session.Steps[StepRecovery] = StatusCompleted
	session.Steps[StepProfile] = StatusActive
	session.CurrentStep = StepProfile
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	return method, nil
}

// SetupRecoveryWallet adds a wallet recovery method
func (s *OnboardingService) SetupRecoveryWallet(sessionID, walletAddress string) (*RecoveryMethod, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.CurrentStep != StepRecovery {
		return nil, ErrStepNotReady
	}

	userID := session.ID
	method, err := s.recovery.AddWallet(userID, walletAddress)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	session.Steps[StepRecovery] = StatusCompleted
	session.Steps[StepProfile] = StatusActive
	session.CurrentStep = StepProfile
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	return method, nil
}

// SetupRecoveryContact adds a trusted contact recovery method
func (s *OnboardingService) SetupRecoveryContact(sessionID, contactDID string) (*RecoveryMethod, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	if session.CurrentStep != StepRecovery {
		return nil, ErrStepNotReady
	}

	userID := session.ID
	method, err := s.recovery.AddTrustedContact(userID, contactDID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	session.Steps[StepRecovery] = StatusCompleted
	session.Steps[StepProfile] = StatusActive
	session.CurrentStep = StepProfile
	session.UpdatedAt = time.Now()
	s.mu.Unlock()

	return method, nil
}

// SkipRecovery skips recovery method setup (flagged as incomplete)
func (s *OnboardingService) SkipRecovery(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return err
	}

	if session.CurrentStep != StepRecovery {
		return ErrStepNotReady
	}

	session.Steps[StepRecovery] = StatusSkipped
	session.SkippedSteps = append(session.SkippedSteps, StepRecovery)
	session.Steps[StepProfile] = StatusActive
	session.CurrentStep = StepProfile
	session.UpdatedAt = time.Now()
	return nil
}

// SetupProfile saves the user's profile information
func (s *OnboardingService) SetupProfile(sessionID string, profile *Profile) error {
	if profile.DisplayName == "" {
		return ErrDisplayNameRequired
	}
	if len(profile.DisplayName) > MaxDisplayNameLen {
		return ErrDisplayNameTooLong
	}
	if len(profile.Bio) > MaxBioLen {
		return ErrBioTooLong
	}
	if profile.Username != "" {
		if !usernameRegex.MatchString(profile.Username) {
			return ErrUsernameInvalid
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return err
	}

	if session.CurrentStep != StepProfile {
		return ErrStepNotReady
	}

	// Check username availability
	if profile.Username != "" {
		if s.usernames[profile.Username] {
			return ErrUsernameTaken
		}
		s.usernames[profile.Username] = true
	}

	session.Profile = profile
	session.Steps[StepProfile] = StatusCompleted
	session.Steps[StepTrustDash] = StatusActive
	session.CurrentStep = StepTrustDash
	session.UpdatedAt = time.Now()
	return nil
}

// SkipProfile skips profile setup with minimal defaults
func (s *OnboardingService) SkipProfile(sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return err
	}

	if session.CurrentStep != StepProfile {
		return ErrStepNotReady
	}

	session.Steps[StepProfile] = StatusSkipped
	session.SkippedSteps = append(session.SkippedSteps, StepProfile)
	session.Steps[StepTrustDash] = StatusActive
	session.CurrentStep = StepTrustDash
	session.UpdatedAt = time.Now()
	return nil
}

// GetSecuritySetup returns the trust dashboard security checklist
func (s *OnboardingService) GetSecuritySetup(sessionID string) ([]SecuritySetupItem, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	items := []SecuritySetupItem{
		{
			Label:     "Phone verified",
			Completed: session.Steps[StepOTP] == StatusCompleted,
			Points:    5,
		},
		{
			Label:     "Passkey enabled",
			Completed: session.Steps[StepPasskey] == StatusCompleted,
			Action:    "Enable",
			Points:    5,
		},
		{
			Label:     "Recovery method",
			Completed: session.Steps[StepRecovery] == StatusCompleted,
			Action:    "Add",
			Points:    10,
		},
		{
			Label:     "Identity verified",
			Completed: false, // Requires KYC which is post-onboarding
			Action:    "+50",
			Points:    50,
		},
	}

	// Clear action for completed items
	for i := range items {
		if items[i].Completed {
			items[i].Action = ""
		}
	}

	return items, nil
}

// GetTrustScore calculates the initial trust score based on completed onboarding steps
func (s *OnboardingService) GetTrustScore(sessionID string) (int, string, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return 0, "", err
	}

	score := 0
	if session.Steps[StepOTP] == StatusCompleted {
		score += 5
	}
	if session.Steps[StepPasskey] == StatusCompleted {
		score += 5
	}
	if session.Steps[StepRecovery] == StatusCompleted {
		score += 5
	}

	level := "Newcomer"
	if score >= 20 {
		level = "Basic"
	}

	return score, level, nil
}

// CompleteOnboarding marks the onboarding flow as complete
func (s *OnboardingService) CompleteOnboarding(sessionID string) (*OnboardingSession, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, err := s.getSessionLocked(sessionID)
	if err != nil {
		return nil, err
	}

	// Must have at least completed phone verification
	if session.Steps[StepOTP] != StatusCompleted {
		return nil, ErrStepNotReady
	}

	now := time.Now()
	session.CompletedAt = &now
	session.CurrentStep = StepComplete
	session.UpdatedAt = now
	return session, nil
}

// GetSkippedSteps returns steps that were skipped during onboarding
func (s *OnboardingService) GetSkippedSteps(sessionID string) ([]OnboardingStep, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	return session.SkippedSteps, nil
}

// HasNudgeCard returns true if the user should see a "Complete Setup" nudge on home
func (s *OnboardingService) HasNudgeCard(sessionID string) (bool, error) {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return false, err
	}
	return len(session.SkippedSteps) > 0, nil
}

// CheckUsernameAvailability checks if a username is available
func (s *OnboardingService) CheckUsernameAvailability(username string) (bool, error) {
	if !usernameRegex.MatchString(username) {
		return false, ErrUsernameInvalid
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	return !s.usernames[username], nil
}

// getSessionLocked retrieves a session (must be called with mu held)
func (s *OnboardingService) getSessionLocked(sessionID string) (*OnboardingSession, error) {
	session, ok := s.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}

	if time.Since(session.CreatedAt) > SessionExpiry {
		return nil, ErrSessionExpired
	}

	return session, nil
}
