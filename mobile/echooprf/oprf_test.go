package echooprf_test

import (
	"testing"

	"github.com/thechadcromwell/echoapp/internal/services/contacts"
	"github.com/thechadcromwell/echoapp/mobile/echooprf"
)

func TestMobileOPRF_MatchesContactsService(t *testing.T) {
	keyHex, err := contacts.GenerateOPRFKeyHex()
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONTACT_OPRF_KEY", keyHex)
	oprfSvc, err := contacts.NewOPRFService()
	if err != nil {
		t.Fatal(err)
	}
	svc := contacts.NewService(nil)
	svc.SetOPRF(oprfSvc)

	client := echooprf.NewClient()
	phone := "+1 555-0100"
	blindResult, err := client.BlindPhones([]string{phone})
	if err != nil {
		t.Fatal(err)
	}
	evaluated, err := svc.OPRFEvaluate(blindResult.Blinded)
	if err != nil {
		t.Fatal(err)
	}
	got, err := client.FinalizePhones(blindResult.SessionID, evaluated)
	if err != nil {
		t.Fatal(err)
	}
	want, err := svc.DiscoveryKey(phone)
	if err != nil {
		t.Fatal(err)
	}
	if got[0] != want {
		t.Fatalf("mobile client %s != server %s", got[0], want)
	}
}
