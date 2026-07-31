package types

import "testing"

func TestSliceMembers(t *testing.T) {
	slice := NewSlice(Int32Type)

	t.Run("length", func(t *testing.T) {
		member, ok := slice.LookupMember("length")
		if !ok {
			t.Fatal("expected slice length member")
		}
		if member.Name != "length" {
			t.Fatalf("expected member name length, got %q", member.Name)
		}
		if member.Kind != MemberProperty {
			t.Fatalf("expected length to be a property, got %v", member.Kind)
		}
		if len(member.Parameters) != 0 {
			t.Fatalf("expected length to have no parameters, got %d", len(member.Parameters))
		}
		if len(member.ReturnTypes) != 1 || !IsTypesEqual(member.ReturnTypes[0], Int64Type) {
			t.Fatalf("expected length to have type int64, got %v", member.ReturnTypes)
		}
		if member.MutatesReceiver {
			t.Fatal("expected length not to mutate its receiver")
		}
	})

	t.Run("append", func(t *testing.T) {
		member, ok := slice.LookupMember("append")
		if !ok {
			t.Fatal("expected slice append member")
		}
		if member.Name != "append" {
			t.Fatalf("expected member name append, got %q", member.Name)
		}
		if member.Kind != MemberMethod {
			t.Fatalf("expected append to be a method, got %v", member.Kind)
		}
		if len(member.Parameters) != 1 || !IsTypesEqual(member.Parameters[0], Int32Type) {
			t.Fatalf("expected append parameter int32, got %v", member.Parameters)
		}
		if len(member.ReturnTypes) != 0 {
			t.Fatalf("expected append to return void, got %v", member.ReturnTypes)
		}
		if !member.MutatesReceiver {
			t.Fatal("expected append to mutate its receiver")
		}
	})

	t.Run("unknown", func(t *testing.T) {
		if _, ok := slice.LookupMember("unknown"); ok {
			t.Fatal("expected unknown slice member lookup to fail")
		}
	})
}

func TestArrayMembers(t *testing.T) {
	array := NewArray(StringType, 3)

	t.Run("length", func(t *testing.T) {
		member, ok := array.LookupMember("length")
		if !ok {
			t.Fatal("expected array length member")
		}
		if member.Name != "length" {
			t.Fatalf("expected member name length, got %q", member.Name)
		}
		if member.Kind != MemberProperty {
			t.Fatalf("expected length to be a property, got %v", member.Kind)
		}
		if len(member.ReturnTypes) != 1 || !IsTypesEqual(member.ReturnTypes[0], Int64Type) {
			t.Fatalf("expected length to have type int64, got %v", member.ReturnTypes)
		}
		if member.MutatesReceiver {
			t.Fatal("expected length not to mutate its receiver")
		}
	})

	t.Run("toSlice", func(t *testing.T) {
		member, ok := array.LookupMember("toSlice")
		if !ok {
			t.Fatal("expected array toSlice member")
		}
		if member.Name != "toSlice" {
			t.Fatalf("expected member name toSlice, got %q", member.Name)
		}
		if member.Kind != MemberMethod {
			t.Fatalf("expected toSlice to be a method, got %v", member.Kind)
		}
		if len(member.Parameters) != 0 {
			t.Fatalf("expected toSlice to have no parameters, got %d", len(member.Parameters))
		}
		if len(member.ReturnTypes) != 1 || member.ReturnTypes[0].Name() != "[]string" {
			t.Fatalf("expected toSlice to return []string, got %v", member.ReturnTypes)
		}
		if member.MutatesReceiver {
			t.Fatal("expected toSlice not to mutate its receiver")
		}
	})

	t.Run("unknown", func(t *testing.T) {
		if _, ok := array.LookupMember("unknown"); ok {
			t.Fatal("expected unknown array member lookup to fail")
		}
	})
}

func TestCollectionTypesProvideMembers(t *testing.T) {
	var _ MemberProvider = NewSlice(Int32Type)
	var _ MemberProvider = NewArray(Int32Type, 1)
}
