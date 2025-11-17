package adapters

import (
	"context"
	"database/sql"
	"forum/internal/modules/user/domain"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3" // Import SQLite driver
)

func TestSQLiteUserRepository_Create(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Errorf("Create returned error: %v", err)
	}

	// Verify the user was created
	var id int
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", user.Email).Scan(&id)
	if err != nil {
		t.Errorf("User was not created in database: %v", err)
	}
	if id != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, id)
	}
}

func TestSQLiteUserRepository_Get(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Insert a user directly for testing
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	_, err = db.Exec("INSERT INTO users (id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	ctx := context.Background()
	result, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Errorf("Get returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected a user, got nil")
	}

	if result.ID != user.ID {
		t.Errorf("Expected ID %d, got %d", user.ID, result.ID)
	}
	if result.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, result.Email)
	}
	if result.Username != user.Username {
		t.Errorf("Expected Username %s, got %s", user.Username, result.Username)
	}
	if result.Role != user.Role {
		t.Errorf("Expected Role %s, got %s", user.Role, result.Role)
	}
	if result.IsActive != user.IsActive {
		t.Errorf("Expected IsActive %v, got %v", user.IsActive, result.IsActive)
	}
}

func TestSQLiteUserRepository_GetByEmail(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Insert a user directly for testing
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	_, err = db.Exec("INSERT INTO users (id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	ctx := context.Background()
	result, err := repo.GetByEmail(ctx, user.Email)
	if err != nil {
		t.Errorf("GetByEmail returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected a user, got nil")
	}

	if result.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, result.Email)
	}
	if result.Username != user.Username {
		t.Errorf("Expected Username %s, got %s", user.Username, result.Username)
	}
}

func TestSQLiteUserRepository_GetByUsername(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Insert a user directly for testing
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	_, err = db.Exec("INSERT INTO users (id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	ctx := context.Background()
	result, err := repo.GetByUsername(ctx, user.Username)
	if err != nil {
		t.Errorf("GetByUsername returned error: %v", err)
	}

	if result == nil {
		t.Fatal("Expected a user, got nil")
	}

	if result.Username != user.Username {
		t.Errorf("Expected Username %s, got %s", user.Username, result.Username)
	}
	if result.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, result.Email)
	}
}

func TestSQLiteUserRepository_Update(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Insert a user directly for testing
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	_, err = db.Exec("INSERT INTO users (id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Prepare updated user
	updatedUser := &domain.User{
		ID:           1,
		Email:        "updated@example.com", // Changed email
		Username:     "updateduser",         // Changed username
		PasswordHash: "new_hash",            // Changed password hash
		Role:         domain.RoleAdmin,      // Changed role
		CreatedAt:    now,
		UpdatedAt:    now.Add(1 * time.Hour), // Changed updated time
		IsActive:     false,                 // Changed active status
	}

	ctx := context.Background()
	err = repo.Update(ctx, updatedUser)
	if err != nil {
		t.Errorf("Update returned error: %v", err)
	}

	// Verify the update in the database
	var email, username string
	var isActive int // Boolean is stored as integer in SQLite
	err = db.QueryRow("SELECT email, username, is_active FROM users WHERE id = ?", updatedUser.ID).Scan(&email, &username, &isActive)
	if err != nil {
		t.Errorf("Failed to query updated user: %v", err)
	}

	if email != "updated@example.com" {
		t.Errorf("Expected email 'updated@example.com', got '%s'", email)
	}
	if username != "updateduser" {
		t.Errorf("Expected username 'updateduser', got '%s'", username)
	}
	if isActive != 0 { // False in SQLite is 0
		t.Errorf("Expected isActive 0 (false), got %d", isActive)
	}
}

func TestSQLiteUserRepository_Delete(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Insert a user directly for testing
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	_, err = db.Exec("INSERT INTO users (id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	ctx := context.Background()
	err = repo.Delete(ctx, user.ID)
	if err != nil {
		t.Errorf("Delete returned error: %v", err)
	}

	// Verify the user was deleted
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE id = ?", user.ID).Scan(&count)
	if err != nil {
		t.Errorf("Failed to query deleted user: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 users after deletion, got %d", count)
	}
}

func TestSQLiteUserRepository_List(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Insert test users directly for testing
	now := time.Now()
	users := []*domain.User{
		{ID: 1, Email: "user1@example.com", Username: "user1", PasswordHash: "hash1", Role: domain.RoleUser, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{ID: 2, Email: "user2@example.com", Username: "user2", PasswordHash: "hash2", Role: domain.RoleUser, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{ID: 3, Email: "user3@example.com", Username: "user3", PasswordHash: "hash3", Role: domain.RoleAdmin, CreatedAt: now, UpdatedAt: now, IsActive: true},
	}

	for _, user := range users {
		_, err = db.Exec("INSERT INTO users (id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			user.ID,
			user.Email,
			user.Username,
			user.PasswordHash,
			user.Role,
			user.CreatedAt,
			user.UpdatedAt,
			user.IsActive,
		)
		if err != nil {
			t.Fatalf("Failed to insert test user: %v", err)
		}
	}

	ctx := context.Background()
	result, err := repo.List(ctx, 0, 0) // Get all users
	if err != nil {
		t.Errorf("List returned error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("Expected 3 users, got %d", len(result))
	}

	// Verify we got the right users
	expectedEmails := map[string]bool{
		"user1@example.com": true,
		"user2@example.com": true,
		"user3@example.com": true,
	}

	for _, user := range result {
		if !expectedEmails[user.Email] {
			t.Errorf("Unexpected user email: %s", user.Email)
		}
	}
}

func TestSQLiteUserRepository_ExistsByEmail(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Insert a user directly for testing
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	_, err = db.Exec("INSERT INTO users (id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	ctx := context.Background()
	
	// Check for existing email
	exists, err := repo.ExistsByEmail(ctx, user.Email)
	if err != nil {
		t.Errorf("ExistsByEmail returned error: %v", err)
	}
	if !exists {
		t.Error("Expected email to exist")
	}
	
	// Check for non-existing email
	exists, err = repo.ExistsByEmail(ctx, "nonexistent@example.com")
	if err != nil {
		t.Errorf("ExistsByEmail returned error: %v", err)
	}
	if exists {
		t.Error("Expected email to not exist")
	}
}

func TestSQLiteUserRepository_ExistsByUsername(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		role TEXT,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active BOOLEAN
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Insert a user directly for testing
	now := time.Now()
	user := &domain.User{
		ID:           1,
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	_, err = db.Exec("INSERT INTO users (id, email, username, password_hash, role, created_at, updated_at, is_active) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		user.ID,
		user.Email,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
	)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	ctx := context.Background()
	
	// Check for existing username
	exists, err := repo.ExistsByUsername(ctx, user.Username)
	if err != nil {
		t.Errorf("ExistsByUsername returned error: %v", err)
	}
	if !exists {
		t.Error("Expected username to exist")
	}
	
	// Check for non-existing username
	exists, err = repo.ExistsByUsername(ctx, "nonexistent")
	if err != nil {
		t.Errorf("ExistsByUsername returned error: %v", err)
	}
	if exists {
		t.Error("Expected username to not exist")
	}
}