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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	user := &domain.User{
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

	// Verify the user was created and has generated IDs
	if user.ID == 0 {
		t.Error("Expected ID to be set after creation")
	}
	if user.PublicID == "" {
		t.Error("Expected PublicID to be generated")
	}

	// Verify the user was created in database
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Create a user using the repository (which generates public_id)
	now := time.Now()
	user := &domain.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Test GetByID
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Create a user using the repository (which generates public_id)
	now := time.Now()
	user := &domain.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	ctx = context.Background()
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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Create a user using the repository (which generates public_id)
	now := time.Now()
	user := &domain.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Create a user using the repository (which generates public_id)
	now := time.Now()
	user := &domain.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Prepare updated user
	updatedUser := &domain.User{
		ID:           user.ID,
		Email:        "updated@example.com", // Changed email
		Username:     "updateduser",         // Changed username
		PasswordHash: "new_hash",            // Changed password hash
		Role:         domain.RoleAdmin,      // Changed role
		CreatedAt:    now,
		UpdatedAt:    now.Add(1 * time.Hour), // Changed updated time
		IsActive:     false,                  // Changed active status
	}

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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Create a user using the repository (which generates public_id)
	now := time.Now()
	user := &domain.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Create test users using the repository (which generates public_id)
	now := time.Now()
	users := []*domain.User{
		{Email: "user1@example.com", Username: "user1", PasswordHash: "hash1", Role: domain.RoleUser, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{Email: "user2@example.com", Username: "user2", PasswordHash: "hash2", Role: domain.RoleUser, CreatedAt: now, UpdatedAt: now, IsActive: true},
		{Email: "user3@example.com", Username: "user3", PasswordHash: "hash3", Role: domain.RoleAdmin, CreatedAt: now, UpdatedAt: now, IsActive: true},
	}

	ctx := context.Background()
	for _, user := range users {
		err = repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create test user: %v", err)
		}
	}

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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Create a user using the repository (which generates public_id)
	now := time.Now()
	user := &domain.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

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
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	// Create a user using the repository (which generates public_id)
	now := time.Now()
	user := &domain.User{
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		CreatedAt:    now,
		UpdatedAt:    now,
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

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

func TestSQLiteUserRepository_GetByPublicID(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	user := &domain.User{
		PublicID:     "test-public-id",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		PostCount:    5,
		CommentCount: 10,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Test GetByPublicID
	retrieved, err := repo.GetByPublicID(ctx, user.PublicID)
	if err != nil {
		t.Fatalf("Failed to get user by public ID: %v", err)
	}

	if retrieved.PublicID != user.PublicID {
		t.Errorf("Expected PublicID %s, got %s", user.PublicID, retrieved.PublicID)
	}
	if retrieved.Email != user.Email {
		t.Errorf("Expected Email %s, got %s", user.Email, retrieved.Email)
	}
	if retrieved.Username != user.Username {
		t.Errorf("Expected Username %s, got %s", user.Username, retrieved.Username)
	}
	if retrieved.PostCount != user.PostCount {
		t.Errorf("Expected PostCount %d, got %d", user.PostCount, retrieved.PostCount)
	}
	if retrieved.CommentCount != user.CommentCount {
		t.Errorf("Expected CommentCount %d, got %d", user.CommentCount, retrieved.CommentCount)
	}

	// Test non-existent public ID
	_, err = repo.GetByPublicID(ctx, "non-existent")
	if err != domain.ErrUserNotFound {
		t.Errorf("Expected domain.ErrUserNotFound, got %v", err)
	}
}

func TestSQLiteUserRepository_Count(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)
	ctx := context.Background()

	// Test count with no users
	count, err := repo.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count 0, got %d", count)
	}

	// Create some users
	for i := 0; i < 3; i++ {
		user := &domain.User{
			PublicID:     "test-public-id-" + string(rune(i+'0')),
			Email:        "test" + string(rune(i+'0')) + "@example.com",
			Username:     "testuser" + string(rune(i+'0')),
			PasswordHash: "hashed_password",
			Role:         domain.RoleUser,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
			IsActive:     true,
		}
		err = repo.Create(ctx, user)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Test count with users
	count, err = repo.Count(ctx)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}
}

func TestSQLiteUserRepository_IncrementPostCount(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	user := &domain.User{
		PublicID:     "test-public-id",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		PostCount:    5,
		CommentCount: 10,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Increment post count
	err = repo.IncrementPostCount(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to increment post count: %v", err)
	}

	// Verify the count was incremented
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if retrieved.PostCount != 6 {
		t.Errorf("Expected PostCount 6, got %d", retrieved.PostCount)
	}
}

func TestSQLiteUserRepository_DecrementPostCount(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	user := &domain.User{
		PublicID:     "test-public-id",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		PostCount:    5,
		CommentCount: 10,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Decrement post count
	err = repo.DecrementPostCount(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to decrement post count: %v", err)
	}

	// Verify the count was decremented
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if retrieved.PostCount != 4 {
		t.Errorf("Expected PostCount 4, got %d", retrieved.PostCount)
	}

	// Test decrementing to zero (should not go below 0)
	userZero := &domain.User{
		PublicID:     "test-public-id-zero",
		Email:        "testzero@example.com",
		Username:     "testuserzero",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		PostCount:    0,
		CommentCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	err = repo.Create(ctx, userZero)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = repo.DecrementPostCount(ctx, userZero.ID)
	if err != nil {
		t.Fatalf("Failed to decrement post count: %v", err)
	}

	retrievedZero, err := repo.GetByID(ctx, userZero.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if retrievedZero.PostCount != 0 {
		t.Errorf("Expected PostCount 0, got %d", retrievedZero.PostCount)
	}
}

func TestSQLiteUserRepository_IncrementCommentCount(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	user := &domain.User{
		PublicID:     "test-public-id",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		PostCount:    5,
		CommentCount: 10,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Increment comment count
	err = repo.IncrementCommentCount(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to increment comment count: %v", err)
	}

	// Verify the count was incremented
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if retrieved.CommentCount != 11 {
		t.Errorf("Expected CommentCount 11, got %d", retrieved.CommentCount)
	}
}

func TestSQLiteUserRepository_DecrementCommentCount(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Create the users table
	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	repo := NewSQLiteUserRepository(db)

	user := &domain.User{
		PublicID:     "test-public-id",
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		PostCount:    5,
		CommentCount: 10,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	ctx := context.Background()
	err = repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// Decrement comment count
	err = repo.DecrementCommentCount(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to decrement comment count: %v", err)
	}

	// Verify the count was decremented
	retrieved, err := repo.GetByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if retrieved.CommentCount != 9 {
		t.Errorf("Expected CommentCount 9, got %d", retrieved.CommentCount)
	}

	// Test decrementing to zero (should not go below 0)
	userZero := &domain.User{
		PublicID:     "test-public-id-zero",
		Email:        "testzero@example.com",
		Username:     "testuserzero",
		PasswordHash: "hashed_password",
		Role:         domain.RoleUser,
		PostCount:    0,
		CommentCount: 0,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}
	err = repo.Create(ctx, userZero)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	err = repo.DecrementCommentCount(ctx, userZero.ID)
	if err != nil {
		t.Fatalf("Failed to decrement comment count: %v", err)
	}

	retrievedZero, err := repo.GetByID(ctx, userZero.ID)
	if err != nil {
		t.Fatalf("Failed to get user: %v", err)
	}
	if retrievedZero.CommentCount != 0 {
		t.Errorf("Expected CommentCount 0, got %d", retrievedZero.CommentCount)
	}
}

func TestSQLiteUserRepository_GetByEmail_IncludesAvatarPath(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	now := time.Now()
	_, err = db.Exec(`INSERT INTO users (
		public_id, email, username, password_hash, avatar_path, role,
		post_count, comment_count, reaction_count, created_at, updated_at, is_active
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"11111111-1111-1111-1111-111111111111",
		"avatar-email@example.com",
		"avataruseremail",
		"hash",
		"avatars/email-user.png",
		domain.RoleUser,
		0, 0, 0,
		now, now,
		1,
	)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	repo := NewSQLiteUserRepository(db)
	user, err := repo.GetByEmail(context.Background(), "avatar-email@example.com")
	if err != nil {
		t.Fatalf("GetByEmail returned error: %v", err)
	}

	if user.AvatarPath != "avatars/email-user.png" {
		t.Fatalf("AvatarPath = %q, want %q", user.AvatarPath, "avatars/email-user.png")
	}

	if user.AvatarURL != domain.AvatarURLPrefix+"avatars/email-user.png" {
		t.Fatalf("AvatarURL = %q, want %q", user.AvatarURL, domain.AvatarURLPrefix+"avatars/email-user.png")
	}
}

func TestSQLiteUserRepository_GetByUsername_IncludesAvatarPath(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		public_id TEXT UNIQUE NOT NULL,
		email TEXT UNIQUE,
		username TEXT UNIQUE,
		password_hash TEXT,
		avatar_path TEXT DEFAULT '',
		role TEXT,
		post_count INTEGER NOT NULL DEFAULT 0,
		comment_count INTEGER NOT NULL DEFAULT 0,
		reaction_count INTEGER NOT NULL DEFAULT 0,
		created_at TIMESTAMP,
		updated_at TIMESTAMP,
		is_active INTEGER
	)`)
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	now := time.Now()
	_, err = db.Exec(`INSERT INTO users (
		public_id, email, username, password_hash, avatar_path, role,
		post_count, comment_count, reaction_count, created_at, updated_at, is_active
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"22222222-2222-2222-2222-222222222222",
		"avatar-username@example.com",
		"avataruserbyname",
		"hash",
		"avatars/username-user.png",
		domain.RoleUser,
		0, 0, 0,
		now, now,
		1,
	)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	repo := NewSQLiteUserRepository(db)
	user, err := repo.GetByUsername(context.Background(), "avataruserbyname")
	if err != nil {
		t.Fatalf("GetByUsername returned error: %v", err)
	}

	if user.AvatarPath != "avatars/username-user.png" {
		t.Fatalf("AvatarPath = %q, want %q", user.AvatarPath, "avatars/username-user.png")
	}

	if user.AvatarURL != domain.AvatarURLPrefix+"avatars/username-user.png" {
		t.Fatalf("AvatarURL = %q, want %q", user.AvatarURL, domain.AvatarURLPrefix+"avatars/username-user.png")
	}
}
