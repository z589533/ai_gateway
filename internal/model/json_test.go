package model

import (
	"database/sql/driver"
	"testing"
)

func TestStringListValueAndScan(t *testing.T) {
	value, err := StringList{"chat:completions", "models:read"}.Value()
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := value.(driver.Value); !ok {
		t.Fatalf("value does not implement driver.Value")
	}

	var scanned StringList
	if err := scanned.Scan(value); err != nil {
		t.Fatal(err)
	}
	if len(scanned) != 2 || scanned[0] != "chat:completions" {
		t.Fatalf("unexpected scanned scopes: %#v", scanned)
	}
}

func TestTableNames(t *testing.T) {
	if (Tenant{}).TableName() != "tenants" {
		t.Fatal("tenant table name mismatch")
	}
	if (APIKey{}).TableName() != "api_keys" {
		t.Fatal("api key table name mismatch")
	}
	if (UsageRecord{}).TableName() != "usage_records" {
		t.Fatal("usage table name mismatch")
	}
}
