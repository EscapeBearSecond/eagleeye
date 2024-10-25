package license

import (
	"strings"
	"testing"

	"github.com/EscapeBearSecond/falcon/internal/license"
	"github.com/stretchr/testify/assert"
)

func TestParseAndVerify(t *testing.T) {
	l1 := strings.NewReader(`{
	"id": "cs32gb9tr70jiv1j5d8g",
	"subject": "cursec_license",
	"issuer": "cursec",
	"issued_at": "2024-10-09",
	"audience": "curescan_test",
	"hardware": "cda5e21c204d81653a078ca0ab4c775b8db0b0be42edecaadd420b6d9fc60d09dedbf091cc929a0a0c777123a8",
	"expires_at": "2024-12-31",
	"signature": "3597fe80a2a147e04582ed7e3fb23128557c3c6fe983930e1763514682a3e8c3"
}`)

	l2 := strings.NewReader(`{
	"id": "cs32gs9tr70jlhqg0cs0",
	"subject": "cursec_license",
	"issuer": "cursec",
	"issued_at": "2024-10-09",
	"audience": "curescan_test",
	"hardware": "5be1c2845c6dbd6de44211fa0cf9528b41d2958e343ff660017ff6e506faa1bf4417c47cd54727be377db26063",
	"expires_at": "2024-12-31",
	"signature": "751615bc2c4de1f94c44e3312632e20343e0431577be0b2a77b1aa99f22cf2d6"
}`)

	l3 := strings.NewReader(`{
	"id": "cs32gv9tr70jm42u9jq0",
	"subject": "cursec_license",
	"issuer": "cursec",
	"issued_at": "2024-10-09",
	"audience": "curescan_test",
	"hardware": "b1abc120f24da0391ccbcc05ac2765948fff325b80b2ad046b4485625e2560d37046554be02854d482a50b7c16",
	"expires_at": "2024-09-30",
	"signature": "cb231378f7ecc9d249eee4cfe98c401e25677bfe186b36221a0e0220ccd55291"
}`)

	{
		err := VerifyFromReader(l1)
		assert.NoError(t, err)
	}

	{
		err := VerifyFromReader(l2)
		assert.ErrorIs(t, err, license.ErrHardwareMismatch)
	}

	{
		err := VerifyFromReader(l3)
		assert.ErrorIs(t, err, license.ErrExpired)
	}
}
