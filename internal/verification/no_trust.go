package verification

type NoTrust struct{}

func (t *NoTrust) Verify(msg string) error {
	return nil
}
