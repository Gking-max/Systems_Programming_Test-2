-- +migrate Up
-- Create database if it doesn't exist
CREATE DATABASE IF NOT EXISTS feedback_db;
USE feedback_db;

-- Create feedback table
CREATE TABLE IF NOT EXISTS feedback (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) NOT NULL,
    rating INT NOT NULL,
    comments TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    CONSTRAINT chk_rating CHECK (rating >= 1 AND rating <= 5)
);

-- Create indexes for better query performance
CREATE INDEX idx_email ON feedback(email);
CREATE INDEX idx_rating ON feedback(rating);
CREATE INDEX idx_created_at ON feedback(created_at);

-- +migrate Down
-- Drop indexes first
DROP INDEX idx_email ON feedback;
DROP INDEX idx_rating ON feedback;
DROP INDEX idx_created_at ON feedback;

-- Drop the table
DROP TABLE IF EXISTS feedback;

-- Drop the database (optional - uncomment if you want to remove the database entirely)
-- DROP DATABASE IF EXISTS feedback_db;