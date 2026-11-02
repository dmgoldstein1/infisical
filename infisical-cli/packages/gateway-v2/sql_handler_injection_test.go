package gatewayv2

import (
	"strings"
	"testing"
)

// CodeQL go/sql-injection: Oracle DDL has no bind parameters for identifiers
// or passwords, so alterPasswordStatement interpolates them. The sink is safe
// only because every interpolated value is rejected when it contains a
// double quote (the sole breakout character inside "..."). Pin that.
func TestAlterPasswordStatementRejectsQuoteBreakout(t *testing.T) {
	oracle := func(target, password, current string) sqlRotateParams {
		return sqlRotateParams{
			sqlTestParams:  sqlTestParams{Dialect: "oracle", Password: current},
			TargetUsername: target,
			NewPassword:    password,
		}
	}
	malicious := []sqlRotateParams{
		oracle(`a" OR "1"="1`, "secret", ""),
		oracle("app", `x" IDENTIFIED BY "p`, ""),
		oracle("app", "secret", `old" REPLACE "x`),
		oracle(`a"; DROP TABLE users; --`, "secret", ""),
	}
	for i, p := range malicious {
		if _, err := alterPasswordStatement(p, p.TargetUsername, p.TargetUsername, true); err == nil {
			t.Errorf("case %d: quote-bearing input was accepted", i)
		}
	}

	// Legit rotation values — including quotes' neighbors that must keep
	// working — still produce a single ALTER USER statement.
	legit, err := alterPasswordStatement(oracle("app_user", `s3;cret 'sq' \ pass`, ""), "app_user", "session_user", false)
	if err != nil {
		t.Fatalf("legit rotation rejected: %v", err)
	}
	if !strings.HasPrefix(legit, `ALTER USER "app_user" IDENTIFIED BY "`) {
		t.Fatalf("unexpected statement: %q", legit)
	}
}
