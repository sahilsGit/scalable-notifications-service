-- Create tables for user notification preferences

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id VARCHAR(36) PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    email VARCHAR(255) NOT NULL,
    global_opt_in BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_username (username),
    UNIQUE KEY unique_email (email)
);

-- Channel preferences per user (email, in-app, push, whatsapp, sms)
CREATE TABLE IF NOT EXISTS user_channel_preferences (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    channel_name VARCHAR(20) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_channel (user_id, channel_name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Event type preferences per user and channel
CREATE TABLE IF NOT EXISTS user_event_preferences (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    channel_name VARCHAR(20) NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_event_channel (user_id, event_type, channel_name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- User contact info for different channels
CREATE TABLE IF NOT EXISTS user_contact_info (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    channel_name VARCHAR(20) NOT NULL,
    contact_value VARCHAR(255) NOT NULL,
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_user_channel_contact (user_id, channel_name),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Insert sample users with global opt-in status
INSERT INTO users (id, username, email, global_opt_in) VALUES 
('user-001', 'user1', 'user1@example.com', TRUE),
('user-002', 'user2', 'user2@example.com', TRUE),
('user-003', 'user3', 'user3@example.com', FALSE);

-- Default channel preferences
INSERT INTO user_channel_preferences (user_id, channel_name, enabled) VALUES 
('user-001', 'email', TRUE),
('user-001', 'in-app', TRUE),
('user-001', 'push', TRUE),
('user-001', 'whatsapp', FALSE),
('user-001', 'sms', FALSE),
('user-002', 'email', TRUE),
('user-002', 'in-app', TRUE),
('user-002', 'push', FALSE),
('user-002', 'whatsapp', TRUE),
('user-002', 'sms', FALSE),
('user-003', 'email', FALSE),
('user-003', 'in-app', FALSE);

-- Event-specific preferences
INSERT INTO user_event_preferences (user_id, event_type, channel_name, enabled) VALUES 
('user-001', 'security_alert', 'email', TRUE),
('user-001', 'security_alert', 'in-app', TRUE),
('user-001', 'security_alert', 'push', TRUE),
('user-001', 'message_received', 'in-app', TRUE),
('user-001', 'message_received', 'push', TRUE),
('user-001', 'message_received', 'email', FALSE),
('user-001', 'like', 'in-app', TRUE),
('user-001', 'like', 'push', FALSE),
('user-002', 'security_alert', 'email', TRUE),
('user-002', 'security_alert', 'in-app', TRUE),
('user-002', 'security_alert', 'whatsapp', TRUE),
('user-002', 'message_received', 'in-app', TRUE),
('user-002', 'message_received', 'whatsapp', TRUE);

-- Contact information
INSERT INTO user_contact_info (user_id, channel_name, contact_value, verified) VALUES 
('user-001', 'email', 'user1@example.com', TRUE),
('user-001', 'push', 'device-token-123', TRUE),
('user-002', 'email', 'user2@example.com', TRUE),
('user-002', 'whatsapp', '+1234567890', TRUE);