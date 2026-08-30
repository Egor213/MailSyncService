package entity

type OAuthState struct {
	State     string `json:"state"`
	MailboxID string `json:"mailbox_id"`
}
