package main

import (
	"reflect"
	"testing"
)

func TestGetUserByEmail(t *testing.T) {
	initDB()
	defer db.Close()

	email := "ksn@mds.rs"
	got, err := getUserByEmail(email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}

	want := User{
		UserID:       12,
		UserName:     "Kasanov",
		Email:        "ksn@mds.rs",
		UserPassword: "ksn123",
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %+v, want %+v", got, want)
	}
}
