-- PostgreSQL queries to update user_settings webhook configurations
-- Replace the placeholders with your actual values

-- ============================================
-- Option 1: Update by User ID (UUID)
-- ============================================
-- Replace 'YOUR_USER_ID_HERE' with the actual UUID of the user
UPDATE user_settings
SET 
    telegram_bot_token = 'YOUR_TELEGRAM_BOT_TOKEN_HERE',
    telegram_chat_id = 'YOUR_TELEGRAM_CHAT_ID_HERE',
    discord_webhook_url = 'YOUR_DISCORD_WEBHOOK_URL_HERE',
    updated_at = NOW()
WHERE user_id = 'YOUR_USER_ID_HERE';

-- ============================================
-- Option 2: Update by User Email (via JOIN)
-- ============================================
-- Replace 'user@example.com' with the actual email
UPDATE user_settings us
SET 
    telegram_bot_token = 'YOUR_TELEGRAM_BOT_TOKEN_HERE',
    telegram_chat_id = 'YOUR_TELEGRAM_CHAT_ID_HERE',
    discord_webhook_url = 'YOUR_DISCORD_WEBHOOK_URL_HERE',
    updated_at = NOW()
FROM users u
WHERE us.user_id = u.id
  AND u.email = 'user@example.com';

-- ============================================
-- Option 3: Update only specific fields (leave others unchanged)
-- ============================================
-- Update only Telegram settings
UPDATE user_settings
SET 
    telegram_bot_token = 'YOUR_TELEGRAM_BOT_TOKEN_HERE',
    telegram_chat_id = 'YOUR_TELEGRAM_CHAT_ID_HERE',
    updated_at = NOW()
WHERE user_id = 'YOUR_USER_ID_HERE';

-- Update only Discord webhook
UPDATE user_settings
SET 
    discord_webhook_url = 'YOUR_DISCORD_WEBHOOK_URL_HERE',
    updated_at = NOW()
WHERE user_id = 'YOUR_USER_ID_HERE';

-- ============================================
-- Option 4: Update multiple users at once
-- ============================================
-- Update all users (use with caution!)
UPDATE user_settings
SET 
    telegram_bot_token = 'YOUR_TELEGRAM_BOT_TOKEN_HERE',
    telegram_chat_id = 'YOUR_TELEGRAM_CHAT_ID_HERE',
    discord_webhook_url = 'YOUR_DISCORD_WEBHOOK_URL_HERE',
    updated_at = NOW()
WHERE user_id IN (
    'USER_ID_1',
    'USER_ID_2',
    'USER_ID_3'
);

-- ============================================
-- Option 5: Set to NULL (clear the values)
-- ============================================
UPDATE user_settings
SET 
    telegram_bot_token = NULL,
    telegram_chat_id = NULL,
    discord_webhook_url = NULL,
    updated_at = NOW()
WHERE user_id = 'YOUR_USER_ID_HERE';

-- ============================================
-- Helper: Find your User ID by email
-- ============================================
SELECT id, email, name 
FROM users 
WHERE email = 'user@example.com';

-- ============================================
-- Helper: View current webhook settings
-- ============================================
SELECT 
    us.user_id,
    u.email,
    u.name,
    us.telegram_bot_token,
    us.telegram_chat_id,
    us.discord_webhook_url,
    us.notification_webhook,
    us.updated_at
FROM user_settings us
JOIN users u ON us.user_id = u.id
WHERE u.email = 'user@example.com';

-- ============================================
-- Example with actual values (template)
-- ============================================
-- Example for a specific user:
-- UPDATE user_settings
-- SET 
--     telegram_bot_token = '1234567890:ABCdefGHIjklMNOpqrsTUVwxyz',
--     telegram_chat_id = '-1001234567890',
--     discord_webhook_url = 'https://discord.com/api/webhooks/1234567890/abcdefghijklmnopqrstuvwxyz',
--     notification_webhook = true,  -- Enable webhook notifications
--     updated_at = NOW()
-- WHERE user_id = '550e8400-e29b-41d4-a716-446655440000';

