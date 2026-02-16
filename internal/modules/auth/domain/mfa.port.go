package domain

import "context"

type MFAChallenge struct {
	ProvisioningURI string
	Sent            bool
}

type MFAService interface {
	Initiate(ctx context.Context, auth *UserAuth) (MFAChallenge, error)
	Verify(ctx context.Context, auth *UserAuth, code string) error
}
