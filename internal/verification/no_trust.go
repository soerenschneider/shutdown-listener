package verification

// NoTrust accepts all messages that are being received.
type NoTrust struct{}

func (t *NoTrust) Verify(msg string) error {
	return nil
}
