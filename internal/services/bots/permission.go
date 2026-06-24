package bots

import "errors"

// ErrPermissionDenied is returned when a bot lacks a granted permission.
var ErrPermissionDenied = errors.New("bot permission denied")

// Authorize checks that userDID granted permission to botDID and the bot is active.
func Authorize(store *InstallStore, userDID, botDID string, perm Permission) error {
	if store == nil {
		return ErrPermissionDenied
	}
	inst, ok := store.Get(userDID, botDID)
	if !ok || !inst.Active {
		return ErrPermissionDenied
	}
	for _, g := range inst.GrantedPermissions {
		if g == perm {
			return nil
		}
	}
	return ErrPermissionDenied
}

// ValidateGrant ensures granted permissions are a subset of manifest requirements.
func ValidateGrant(manifest Manifest, granted []Permission) bool {
	if len(granted) == 0 {
		return false
	}
	required := make(map[Permission]bool, len(manifest.RequiredPermissions))
	for _, p := range manifest.RequiredPermissions {
		required[p] = true
	}
	for _, g := range granted {
		if !required[g] {
			return false
		}
	}
	// All required permissions must be granted.
	for p := range required {
		found := false
		for _, g := range granted {
			if g == p {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
