-- Migration: Add reaction_count column to users table
-- The reaction_count column was missing from the users schema but is 
-- required by the user repository's GetByID query.

-- Add reaction_count column with default value of 0
ALTER TABLE users ADD COLUMN reaction_count INTEGER NOT NULL DEFAULT 0;
