package gatewayv2

import (
	"strings"
	"testing"
)

// CODEQL-ATTEMPT(run-20261126T052024Z item 12, go/sql-injection): Oracle DDL
// takes no bind parameters for identifiers or secrets, so the rotated secret
// must ride literally inside the double-quoted IDENTIFIED BY clause (with the
// quote/control/length guards in validateOracleRotateParams as the breakout
// defense). A bind-variable rewrite produces DDL Oracle rejects at runtime.
// This test FAILS under that rewrite (statement no longer carries the
// secret), documenting the keep.
func TestAlterPasswordStatementCarriesSecretLiterally(t *testing.T) {
	p := sqlRotateParams{TargetUsername: "APP_USER", NewPassword: "NewPw_123"}
	p.Dialect = "oracle"
	p.Username = "ROTATOR"

	got, err := alterPasswordStatement(p, p.TargetUsername, p.Username, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `IDENTIFIED BY "NewPw_123"`) {
		t.Fatalf("rotation DDL must carry the new secret literally (Oracle DDL takes no binds): %q", got)
	}
}
