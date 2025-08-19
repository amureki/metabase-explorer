package api

import (
	"encoding/json"
	"testing"
)

func TestDatabase_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": 1,
		"name": "Sample Database",
		"engine": "postgres"
	}`

	var db Database
	err := json.Unmarshal([]byte(jsonData), &db)
	if err != nil {
		t.Fatalf("Failed to unmarshal Database: %v", err)
	}

	if db.ID != 1 {
		t.Errorf("Database.ID = %d, want 1", db.ID)
	}
	if db.Name != "Sample Database" {
		t.Errorf("Database.Name = %s, want 'Sample Database'", db.Name)
	}
	if db.Engine != "postgres" {
		t.Errorf("Database.Engine = %s, want 'postgres'", db.Engine)
	}
}

func TestTable_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": 100,
		"name": "users",
		"display_name": "Users Table",
		"schema": "public",
		"description": "User account information",
		"fields": []
	}`

	var table Table
	err := json.Unmarshal([]byte(jsonData), &table)
	if err != nil {
		t.Fatalf("Failed to unmarshal Table: %v", err)
	}

	if table.ID != 100 {
		t.Errorf("Table.ID = %d, want 100", table.ID)
	}
	if table.Name != "users" {
		t.Errorf("Table.Name = %s, want 'users'", table.Name)
	}
	if table.DisplayName != "Users Table" {
		t.Errorf("Table.DisplayName = %s, want 'Users Table'", table.DisplayName)
	}
	if table.Schema != "public" {
		t.Errorf("Table.Schema = %s, want 'public'", table.Schema)
	}
	if table.Description != "User account information" {
		t.Errorf("Table.Description = %s, want 'User account information'", table.Description)
	}
	if table.Fields == nil {
		t.Error("Table.Fields should not be nil")
	}
}

func TestField_JSONUnmarshal(t *testing.T) {
	jsonData := `{
		"id": 200,
		"name": "email",
		"display_name": "Email Address",
		"description": "User's email address",
		"base_type": "type/Text",
		"effective_type": "type/Email",
		"semantic_type": "type/Email",
		"database_type": "varchar",
		"table_id": 100,
		"position": 2,
		"active": true,
		"preview_display": true,
		"visibility_type": "normal"
	}`

	var field Field
	err := json.Unmarshal([]byte(jsonData), &field)
	if err != nil {
		t.Fatalf("Failed to unmarshal Field: %v", err)
	}

	if field.ID != 200 {
		t.Errorf("Field.ID = %d, want 200", field.ID)
	}
	if field.Name != "email" {
		t.Errorf("Field.Name = %s, want 'email'", field.Name)
	}
	if field.DisplayName != "Email Address" {
		t.Errorf("Field.DisplayName = %s, want 'Email Address'", field.DisplayName)
	}
	if field.Description != "User's email address" {
		t.Errorf("Field.Description = %s, want 'User's email address'", field.Description)
	}
	if field.BaseType != "type/Text" {
		t.Errorf("Field.BaseType = %s, want 'type/Text'", field.BaseType)
	}
	if field.EffectiveType != "type/Email" {
		t.Errorf("Field.EffectiveType = %s, want 'type/Email'", field.EffectiveType)
	}
	if field.SemanticType != "type/Email" {
		t.Errorf("Field.SemanticType = %s, want 'type/Email'", field.SemanticType)
	}
	if field.DatabaseType != "varchar" {
		t.Errorf("Field.DatabaseType = %s, want 'varchar'", field.DatabaseType)
	}
	if field.TableID != 100 {
		t.Errorf("Field.TableID = %d, want 100", field.TableID)
	}
	if field.Position != 2 {
		t.Errorf("Field.Position = %d, want 2", field.Position)
	}
	if !field.Active {
		t.Error("Field.Active should be true")
	}
	if !field.PreviewDisplay {
		t.Error("Field.PreviewDisplay should be true")
	}
	if field.Visibility != "normal" {
		t.Errorf("Field.Visibility = %s, want 'normal'", field.Visibility)
	}
}

func TestSchema(t *testing.T) {
	schema := Schema{
		Name:       "public",
		TableCount: 5,
	}

	if schema.Name != "public" {
		t.Errorf("Schema.Name = %s, want 'public'", schema.Name)
	}
	if schema.TableCount != 5 {
		t.Errorf("Schema.TableCount = %d, want 5", schema.TableCount)
	}
}
