package trustnet

import "errors"

var (
	// Circle errors
	ErrCircleNotFound      = errors.New("trust circle not found")
	ErrContactNotFound     = errors.New("contact not found in any circle")
	ErrContactAlreadyInCircle = errors.New("contact already in this circle")
	ErrCircleFull          = errors.New("trust circle is at maximum capacity")
	ErrCannotAddSelf       = errors.New("cannot add yourself to a circle")
	ErrInvalidCircleTier   = errors.New("invalid circle tier")

	// Endorsement errors
	ErrEndorsementSelfEndorse   = errors.New("cannot endorse yourself")
	ErrEndorsementDuplicate     = errors.New("already endorsed this user in this category")
	ErrEndorsementInsufficientTrust = errors.New("insufficient trust score to endorse")
	ErrEndorsementRateLimited   = errors.New("endorsement daily limit reached")
	ErrEndorsementNotFound      = errors.New("endorsement not found")
	ErrEndorsementNotOwner      = errors.New("only the endorser can revoke")
	ErrEndorsementCooldown      = errors.New("endorsement revocation in cooldown period")

	// Dispute errors
	ErrDisputeRateLimited   = errors.New("can only file 1 dispute per 90 days")
	ErrDisputeNotFound      = errors.New("dispute not found")
	ErrDisputeAlreadyResolved = errors.New("dispute already resolved")
	ErrDisputeInvalidType   = errors.New("invalid dispute type")
	ErrJurorIneligible      = errors.New("juror does not meet requirements")
	ErrJurorAlreadyVoted    = errors.New("juror has already voted")
	ErrJurorConflict        = errors.New("juror has connection to dispute parties")

	// Anti-sybil errors
	ErrSybilDetected        = errors.New("potential sybil account detected")
	ErrDeviceClusterDetected = errors.New("device cluster detected")

	// Rate limit errors
	ErrTrustOpRateLimited = errors.New("trust operation rate limit exceeded")

	// Contact request errors
	ErrRequestAlreadySent   = errors.New("contact request already sent")
	ErrRequestNotFound      = errors.New("contact request not found")
	ErrRequestAlreadyHandled = errors.New("contact request already accepted or declined")
	ErrRequestToSelf        = errors.New("cannot send contact request to yourself")
	ErrAlreadyContacts      = errors.New("already in contacts")
	ErrRequestDeclined      = errors.New("contact request was declined")

	// Persona errors
	ErrPersonaNotFound    = errors.New("persona not found")
	ErrPersonaDuplicate   = errors.New("persona already exists")
	ErrPersonaDefaultOnly = errors.New("cannot remove default persona")

	// Discovery errors
	ErrQRCodeInvalid   = errors.New("invalid echo QR code")
	ErrQRCodeExpired   = errors.New("QR code has expired")
	ErrUserNotFound    = errors.New("user not found")
	ErrSearchTooShort  = errors.New("search query too short")
)
