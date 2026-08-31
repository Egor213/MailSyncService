package mail

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"mail-sync-service/internal/entity"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
)

type IMAPClient struct {
	client *imapclient.Client
}

func NewIMAPClient() *IMAPClient {
	return &IMAPClient{}
}

func (c *IMAPClient) Connect(ctx context.Context, server string, port int, useTLS bool) error {
	addr := fmt.Sprintf("%s:%d", server, port)
	var (
		client *imapclient.Client
		err    error
	)

	if useTLS {
		options := &imapclient.Options{
			TLSConfig: &tls.Config{
				InsecureSkipVerify: false,
			},
		}
		client, err = imapclient.DialTLS(addr, options)
	} else {
		client, err = imapclient.DialInsecure(addr, nil)
	}
	if err != nil {
		return fmt.Errorf("connect to %s: %w", addr, err)
	}
	c.client = client
	return nil
}

func (c *IMAPClient) Authenticate(ctx context.Context, mailbox *entity.Mailbox) error {
	if c.client == nil {
		return fmt.Errorf("client not connected")
	}

	if mailbox.AuthTypeID == entity.AuthTypeToID[entity.AuthTypeOAuth2] {
		authString := c.buildXOAUTH2String(mailbox.Email, mailbox.AccessToken)
		saslClient := &xoauth2Client{authString: authString}
		if err := c.client.Authenticate(saslClient); err != nil {
			return fmt.Errorf("oauth2 auth failed: %w", err)
		}
	} else {
		return fmt.Errorf("plain auth not implemented yet")
	}
	return nil
}

func (c *IMAPClient) buildXOAUTH2String(email, accessToken string) string {
	authStr := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", email, accessToken)
	return base64.StdEncoding.EncodeToString([]byte(authStr))
}

type xoauth2Client struct {
	authString string
}

func (x *xoauth2Client) Start() (mech string, initial []byte, err error) {
	return "XOAUTH2", []byte(x.authString), nil
}

func (x *xoauth2Client) Next(challenge []byte) ([]byte, error) {
	if len(challenge) > 0 {
		return []byte{}, nil
	}
	return nil, nil
}

func (c *IMAPClient) ListMessages(ctx context.Context, folder string, lastUID string) ([]*MessageMeta, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	selectCmd := c.client.Select(folder, nil)
	if _, err := selectCmd.Wait(); err != nil {
		return nil, fmt.Errorf("select folder %s: %w", folder, err)
	}

	var criteria imap.SearchCriteria
	if lastUID != "" {
		var uid uint32
		if _, err := fmt.Sscan(lastUID, &uid); err == nil && uid > 0 {
			uidSet := imap.UIDSet{}
			uidSet.AddRange(imap.UID(uid+1), 0)
			criteria.UID = []imap.UIDSet{uidSet}
		}
	}

	searchCmd := c.client.UIDSearch(&criteria, nil)
	searchData, err := searchCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("uid search failed: %w", err)
	}
	uids := searchData.AllUIDs()
	if len(uids) == 0 {
		return []*MessageMeta{}, nil
	}

	var result []*MessageMeta
	const batchSize = 50

	for i := 0; i < len(uids); i += batchSize {
		end := i + batchSize
		if end > len(uids) {
			end = len(uids)
		}
		batch := uids[i:end]

		uidSet := imap.UIDSet{}
		for _, uid := range batch {
			uidSet.AddNum(uid)
		}

		fetchOptions := &imap.FetchOptions{
			UID:      true,
			Envelope: true,
			Flags:    true,
			BodySection: []*imap.FetchItemBodySection{
				{
					Specifier:    imap.PartSpecifierHeader,
					HeaderFields: []string{"Subject", "From", "To", "Date"},
				},
			},
		}

		fetchCmd := c.client.Fetch(uidSet, fetchOptions)
		msgBufs, err := fetchCmd.Collect()
		if err != nil {
			return nil, fmt.Errorf("fetch collect failed: %w", err)
		}

		for _, msgBuf := range msgBufs {
			meta := &MessageMeta{
				UID:    fmt.Sprintf("%d", msgBuf.UID),
				Folder: folder,
			}
			if msgBuf.Envelope != nil {
				meta.Subject = msgBuf.Envelope.Subject
				if len(msgBuf.Envelope.From) > 0 {
					addr := msgBuf.Envelope.From[0]
					if addr.Host != "" {
						meta.From = fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
					} else {
						meta.From = addr.Mailbox
					}
				}
				if len(msgBuf.Envelope.To) > 0 {
					addr := msgBuf.Envelope.To[0]
					if addr.Host != "" {
						meta.To = fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
					} else {
						meta.To = addr.Mailbox
					}
				}
				meta.Date = msgBuf.Envelope.Date.Format(time.RFC3339)
			}
			meta.Seen = false
			for _, flag := range msgBuf.Flags {
				if flag == imap.FlagSeen {
					meta.Seen = true
				}
			}
			meta.Flags = flagsToStrings(msgBuf.Flags)

			result = append(result, meta)
		}
	}
	return result, nil
}

func (c *IMAPClient) FetchFullMessage(ctx context.Context, folder, uid string) (*MessageMeta, error) {
	if c.client == nil {
		return nil, fmt.Errorf("client not connected")
	}

	selectCmd := c.client.Select(folder, nil)
	if _, err := selectCmd.Wait(); err != nil {
		return nil, fmt.Errorf("select folder %s: %w", folder, err)
	}

	var uidNum uint32
	if _, err := fmt.Sscan(uid, &uidNum); err != nil {
		return nil, fmt.Errorf("invalid uid: %s", uid)
	}

	uidSet := imap.UIDSet{}
	uidSet.AddNum(imap.UID(uidNum))

	fetchOptions := &imap.FetchOptions{
		UID: true,
		BodySection: []*imap.FetchItemBodySection{
			{Specifier: imap.PartSpecifierNone},
		},
	}

	fetchCmd := c.client.Fetch(uidSet, fetchOptions)
	msgBufs, err := fetchCmd.Collect()
	if err != nil {
		return nil, fmt.Errorf("fetch body failed: %w", err)
	}
	if len(msgBufs) == 0 {
		return nil, fmt.Errorf("message not found")
	}
	msgBuf := msgBufs[0]

	meta := &MessageMeta{
		UID:    fmt.Sprintf("%d", msgBuf.UID),
		Folder: folder,
	}

	if bodySection := msgBuf.FindBodySection(fetchOptions.BodySection[0]); bodySection != nil {
		rawBody := string(bodySection)
		plain, html := extractTextAndHTML(rawBody)
		meta.BodyText = plain
		meta.BodyHTML = html
	}

	return meta, nil
}

func (c *IMAPClient) Close() error {
	if c.client != nil {
		return c.client.Close()
	}
	return nil
}

func flagsToStrings(flags []imap.Flag) []string {
	strs := make([]string, len(flags))
	for i, f := range flags {
		strs[i] = string(f)
	}
	return strs
}

func extractTextAndHTML(raw string) (plain, html string) {
	headerEnd := strings.Index(raw, "\r\n\r\n")
	if headerEnd == -1 {
		headerEnd = strings.Index(raw, "\n\n")
		if headerEnd == -1 {
			return raw, ""
		}
		headerEnd += 2
	} else {
		headerEnd += 4
	}

	headers := raw[:headerEnd]
	body := raw[headerEnd:]

	contentType := ""
	for _, line := range strings.Split(headers, "\n") {
		if strings.HasPrefix(strings.ToLower(line), "content-type:") {
			contentType = strings.TrimSpace(line[len("content-type:"):])
			break
		}
	}

	if !strings.HasPrefix(strings.ToLower(contentType), "multipart/") {
		if strings.Contains(strings.ToLower(contentType), "text/html") {
			return "", body
		}
		return body, ""
	}

	boundary := ""
	for _, part := range strings.Split(contentType, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(strings.ToLower(part), "boundary=") {
			boundary = strings.Trim(part[len("boundary="):], `"`)
			break
		}
	}
	if boundary == "" {
		return body, ""
	}

	parts := strings.Split(body, "--"+boundary)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || part == "--" {
			continue
		}
		if strings.HasSuffix(part, "--") {
			part = strings.TrimSuffix(part, "--")
		}

		partHeaderEnd := strings.Index(part, "\r\n\r\n")
		if partHeaderEnd == -1 {
			partHeaderEnd = strings.Index(part, "\n\n")
			if partHeaderEnd == -1 {
				continue
			}
			partHeaderEnd += 2
		} else {
			partHeaderEnd += 4
		}
		partHeaders := part[:partHeaderEnd]
		partBody := part[partHeaderEnd:]

		partContentType := ""
		for _, line := range strings.Split(partHeaders, "\n") {
			if strings.HasPrefix(strings.ToLower(line), "content-type:") {
				partContentType = strings.TrimSpace(line[len("content-type:"):])
				break
			}
		}

		if strings.Contains(strings.ToLower(partContentType), "text/plain") {
			plain = partBody
		} else if strings.Contains(strings.ToLower(partContentType), "text/html") {
			html = partBody
		}
	}

	return plain, html
}
