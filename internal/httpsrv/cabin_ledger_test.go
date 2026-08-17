package httpsrv_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"microgrid-dispatch/internal/domain/cabin"
	"microgrid-dispatch/internal/httpsrv"
)

// listCabins reads the cabin ledger through the API.
func listCabins(t *testing.T, srv *httpsrv.Server) []cabin.Cabin {
	t.Helper()
	req := httptest.NewRequest("GET", "/api/cabins", &bytes.Buffer{})
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list cabins: code = %d (body %s)", rec.Code, rec.Body.String())
	}
	var out []cabin.Cabin
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("decode cabin list: %v (body %s)", err, rec.Body.String())
	}
	return out
}

func registerCabin(t *testing.T, srv *httpsrv.Server, id, name string, capacity float64) {
	t.Helper()
	code, body := do(t, srv, "POST", "/api/cabins", map[string]any{
		"id": id, "name": name, "capacity_kwh": capacity,
	})
	if code != http.StatusCreated {
		t.Fatalf("register cabin %s: %d %v", id, code, body)
	}
}

// TestHTTPCabinLedgerListsRegisteredCabins covers the cabin ledger listing: every
// cabin that was registered must show up in the list with its own attributes.
func TestHTTPCabinLedgerListsRegisteredCabins(t *testing.T) {
	srv, _ := newServer(t)

	if got := listCabins(t, srv); len(got) != 0 {
		t.Fatalf("cabin list = %+v, want empty before any registration", got)
	}

	registerCabin(t, srv, "cabin-1", "1号构网型储能舱", 1250)
	registerCabin(t, srv, "cabin-2", "2号构网型储能舱", 2000)

	got := listCabins(t, srv)
	if len(got) != 2 {
		t.Fatalf("cabin list has %d entries, want 2: %+v", len(got), got)
	}
	byID := map[string]cabin.Cabin{}
	for _, c := range got {
		byID[c.ID] = c
	}
	first, ok := byID["cabin-1"]
	if !ok {
		t.Fatalf("cabin-1 missing from the ledger list: %+v", got)
	}
	if first.Name != "1号构网型储能舱" || first.CapacityKWh != 1250 || !first.Powered {
		t.Fatalf("cabin-1 entry = %+v, want name/capacity/powered as registered", first)
	}
	second, ok := byID["cabin-2"]
	if !ok {
		t.Fatalf("cabin-2 missing from the ledger list: %+v", got)
	}
	if second.CapacityKWh != 2000 {
		t.Fatalf("cabin-2 capacity = %v, want 2000", second.CapacityKWh)
	}
}

// TestHTTPCabinLedgerGrowsWithEachRegistration covers incremental registration:
// the list length must track the number of registered cabins.
func TestHTTPCabinLedgerGrowsWithEachRegistration(t *testing.T) {
	srv, _ := newServer(t)

	for i, id := range []string{"cabin-a", "cabin-b", "cabin-c"} {
		registerCabin(t, srv, id, "舱区"+id, 1000+float64(i))
		got := listCabins(t, srv)
		if len(got) != i+1 {
			t.Fatalf("after registering %s: list has %d entries, want %d: %+v", id, len(got), i+1, got)
		}
	}
}

// TestHTTPCabinLedgerReflectsLockState covers the operational view: once a drill
// takes the cabin work lock, the ledger listing must show the holding crew, and
// the single-cabin lookup must agree with the list.
func TestHTTPCabinLedgerReflectsLockState(t *testing.T) {
	srv, _ := newServer(t)
	setupCabinAndPlan(t, srv)

	code, body := do(t, srv, "POST", "/api/drills", map[string]any{
		"plan_id": "plan-1", "cabin_id": "cabin-1", "team_id": "crew-A",
	})
	if code != http.StatusCreated {
		t.Fatalf("apply drill: %d %v", code, body)
	}
	drillID := body["id"].(string)
	if code, body := do(t, srv, "POST", "/api/drills/"+drillID+"/approve", map[string]any{
		"supervisor_id": "sup",
	}); code != http.StatusOK {
		t.Fatalf("approve: %d %v", code, body)
	}

	got := listCabins(t, srv)
	if len(got) != 1 {
		t.Fatalf("cabin list has %d entries, want 1: %+v", len(got), got)
	}
	if got[0].LockHolder != "crew-A" {
		t.Fatalf("cabin list lock_holder = %q, want crew-A", got[0].LockHolder)
	}

	code, single := do(t, srv, "GET", "/api/cabins/cabin-1", nil)
	if code != http.StatusOK {
		t.Fatalf("get cabin: %d %v", code, single)
	}
	if single["lock_holder"] != "crew-A" {
		t.Fatalf("single cabin lock_holder = %v, want crew-A", single["lock_holder"])
	}
}
