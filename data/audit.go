package data

import "time"

// AuditSystemUser ...
const AuditSystemUser = "<system>"

// AuditEventContextType ...
type AuditEventContextType string

const (
	// AuditEventContextUser ...
	AuditEventContextUser AuditEventContextType = "user"
)

// AuditEvent ...
type AuditEvent string

// AuditReason ...
type AuditReason string

const (
	// AuditEventUserCreated ...
	AuditEventUserCreated AuditEvent = "user_created"
	// AuditEventVerificationEmailSent ...
	AuditEventVerificationEmailSent AuditEvent = "verification_email_sent"
	// AuditEventUserLoginOK ...
	AuditEventUserLoginOK AuditEvent = "user_login_ok"
	// AuditEventUserLoginFailed ...
	AuditEventUserLoginFailed AuditEvent = "user_login_failed"
	// AuditEventPasswordChangeOK ...
	AuditEventPasswordChangeOK AuditEvent = "password_change_ok"
	// AuditEventPasswordChangeFailed ...
	AuditEventPasswordChangeFailed AuditEvent = "password_change_failed"
	// AuditEventUserRolesUpdated ...
	AuditEventUserRolesUpdated AuditEvent = "user_roles_updated"

	// AuditReasonNone ...
	AuditReasonNone AuditReason = ""
	// AuditReasonUserNotFound ...
	AuditReasonUserNotFound AuditReason = "user_not_found"
	// AuditReasonUserInactive ...
	AuditReasonUserInactive AuditReason = "user_inactive"
	// AuditReasonInvalidPassword ...
	AuditReasonInvalidPassword AuditReason = "invalid_password"
	// AuditReasonPasswordChangeRequired ...
	AuditReasonPasswordChangeRequired AuditReason = "password_change_required"
)

type auditEvent struct {
	UserID      string                `bson:"user_id"`
	Created     time.Time             `bson:"created"`
	ContextType AuditEventContextType `bson:"context_type"`
	Context     string                `bson:"context"`
	Event       AuditEvent            `bson:"event"`
	Reason      AuditReason           `bson:"reason"`
}

func (m *MongoDB) createAuditEvent(userID string, contextType AuditEventContextType, context string, event AuditEvent, reason AuditReason) error {
	sess := m.New()
	defer sess.Close()

	e := auditEvent{
		UserID:      userID,
		Created:     time.Now(),
		ContextType: contextType,
		Context:     context,
		Event:       event,
		Reason:      reason,
	}

	return m.DB("florence").C("audit").Insert(&e)
}
