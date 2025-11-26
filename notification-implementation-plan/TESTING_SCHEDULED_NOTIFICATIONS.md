# Testing Scheduled Notifications Manually

This guide shows you how to test scheduled notifications without waiting for the actual due dates.

## Method 1: Create a Notification with Near-Future Schedule (Recommended)

### Step 1: Create a Test Debt List

```bash
# First, get your JWT token by logging in
TOKEN=$(curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-email@example.com",
    "password": "your-password"
  }' | jq -r '.data.token')

# Create a debt list
DEBT_LIST_ID=$(curl -X POST http://localhost:8080/api/v1/debts \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "contact_id": "your-contact-id",
    "total_amount": 1000.00,
    "currency": "Php",
    "debt_type": "owed_to_me",
    "due_date": "2025-12-31T00:00:00Z",
    "installment_plan": "onetime"
  }' | jq -r '.data.id')
```

### Step 2: Create a Notification Scheduled for 2 Minutes from Now

```bash
# Calculate time 2 minutes from now (adjust as needed)
SCHEDULED_TIME=$(date -u -d '+2 minutes' '+%Y-%m-%dT%H:%M:%SZ')

# Create a notification scheduled for near future
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"debt_list_id\": \"$DEBT_LIST_ID\",
    \"notification_type\": \"email\",
    \"recipient_type\": \"user\",
    \"message\": \"Test scheduled notification\",
    \"scheduled_for\": \"$SCHEDULED_TIME\",
    \"schedule_type\": \"reminder\"
  }"
```

### Step 3: Verify the Notification Was Created

```bash
# Get all notifications for the debt list
curl -X GET "http://localhost:8080/api/v1/notifications/debt-lists/$DEBT_LIST_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.'
```

### Step 4: Wait and Check Status

The worker processes notifications every minute (configurable via `NOTIFICATION_WORKER_INTERVAL`). After 2 minutes:

```bash
# Check notification status
curl -X GET "http://localhost:8080/api/v1/notifications/{notification-id}" \
  -H "Authorization: Bearer $TOKEN" | jq '.data.status'
```

Expected status changes:
- Initially: `"pending"`
- After worker processes: `"sent"` (if successful) or `"failed"` (if error)

## Method 2: Schedule Notifications for a Debt List

### Use the Schedule Endpoint

```bash
# Schedule notifications for a debt list (creates notifications based on due date)
curl -X POST "http://localhost:8080/api/v1/notifications/schedule" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"debt_list_id\": \"$DEBT_LIST_ID\"
  }"
```

This creates notifications at:
- 7 days before due date
- 3 days before due date  
- 1 day before due date

**Note:** For testing, create a debt list with a due date 1-2 days in the future to see notifications scheduled soon.

## Method 3: Direct Database Manipulation (Advanced)

If you need to test with very specific times, you can directly update the database:

```sql
-- Update a notification's scheduled time to 1 minute from now
UPDATE notifications 
SET scheduled_for = NOW() + INTERVAL '1 minute',
    next_run_at = NOW() + INTERVAL '1 minute',
    status = 'pending'
WHERE id = 'your-notification-id';
```

## Method 4: Test with Immediate Schedule

Create a notification scheduled for the current time or past time (worker will process immediately):

```bash
# Schedule for 10 seconds from now
SCHEDULED_TIME=$(date -u -d '+10 seconds' '+%Y-%m-%dT%H:%M:%SZ')

curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"debt_list_id\": \"$DEBT_LIST_ID\",
    \"notification_type\": \"email\",
    \"recipient_type\": \"user\",
    \"message\": \"Immediate test notification\",
    \"scheduled_for\": \"$SCHEDULED_TIME\",
    \"schedule_type\": \"manual\"
  }"
```

## Monitoring the Worker

### Check Worker Logs

The worker logs when it processes notifications. Look for:

```
INFO Starting notification worker
INFO Processing pending notifications
INFO Notification sent successfully notification_id=...
```

### Check Notification Status via API

```bash
# Get all pending notifications
curl -X GET "http://localhost:8080/api/v1/notifications" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.data[] | select(.status == "pending")'
```

## Testing Different Notification Types

### Email Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"debt_list_id\": \"$DEBT_LIST_ID\",
    \"notification_type\": \"email\",
    \"recipient_type\": \"user\",
    \"message\": \"Test email notification\",
    \"scheduled_for\": \"$SCHEDULED_TIME\"
  }"
```

### SMS Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"debt_list_id\": \"$DEBT_LIST_ID\",
    \"notification_type\": \"sms\",
    \"recipient_type\": \"user\",
    \"message\": \"Test SMS notification\",
    \"scheduled_for\": \"$SCHEDULED_TIME\"
  }"
```

### Webhook Notification

```bash
curl -X POST http://localhost:8080/api/v1/notifications \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"debt_list_id\": \"$DEBT_LIST_ID\",
    \"notification_type\": \"webhook\",
    \"webhook_type\": \"slack\",
    \"recipient_type\": \"user\",
    \"message\": \"Test webhook notification\",
    \"scheduled_for\": \"$SCHEDULED_TIME\"
  }"
```

## Quick Test Script

Save this as `test_notification.sh`:

```bash
#!/bin/bash

# Configuration
API_URL="http://localhost:8080/api/v1"
EMAIL="your-email@example.com"
PASSWORD="your-password"

# Login
echo "Logging in..."
TOKEN=$(curl -s -X POST "$API_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}" \
  | jq -r '.data.token')

if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo "❌ Login failed"
  exit 1
fi

echo "✅ Logged in successfully"

# Get first debt list
DEBT_LIST_ID=$(curl -s -X GET "$API_URL/debts" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data[0].id')

if [ "$DEBT_LIST_ID" == "null" ] || [ -z "$DEBT_LIST_ID" ]; then
  echo "❌ No debt lists found. Create one first."
  exit 1
fi

echo "✅ Using debt list: $DEBT_LIST_ID"

# Schedule notification for 1 minute from now
SCHEDULED_TIME=$(date -u -d '+1 minute' '+%Y-%m-%dT%H:%M:%SZ')
echo "📅 Scheduling notification for: $SCHEDULED_TIME"

RESPONSE=$(curl -s -X POST "$API_URL/notifications" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"debt_list_id\": \"$DEBT_LIST_ID\",
    \"notification_type\": \"email\",
    \"recipient_type\": \"user\",
    \"message\": \"Test scheduled notification\",
    \"scheduled_for\": \"$SCHEDULED_TIME\",
    \"schedule_type\": \"reminder\"
  }")

NOTIFICATION_ID=$(echo $RESPONSE | jq -r '.data.id')

if [ "$NOTIFICATION_ID" == "null" ] || [ -z "$NOTIFICATION_ID" ]; then
  echo "❌ Failed to create notification"
  echo $RESPONSE | jq '.'
  exit 1
fi

echo "✅ Notification created: $NOTIFICATION_ID"
echo "⏳ Waiting 70 seconds for worker to process..."

# Wait and check status
sleep 70

STATUS=$(curl -s -X GET "$API_URL/notifications/$NOTIFICATION_ID" \
  -H "Authorization: Bearer $TOKEN" \
  | jq -r '.data.status')

echo "📊 Notification status: $STATUS"

if [ "$STATUS" == "sent" ]; then
  echo "✅ Notification sent successfully!"
elif [ "$STATUS" == "failed" ]; then
  echo "❌ Notification failed. Check logs."
else
  echo "⏳ Notification still pending. Worker may need more time."
fi
```

Make it executable and run:

```bash
chmod +x test_notification.sh
./test_notification.sh
```

## Troubleshooting

### Worker Not Processing Notifications

1. **Check if worker is running:**
   - Look for "Starting notification worker" in logs
   - Check `NOTIFICATION_WORKER_INTERVAL` environment variable

2. **Check notification status:**
   ```bash
   # Should show pending notifications
   curl -X GET "http://localhost:8080/api/v1/notifications" \
     -H "Authorization: Bearer $TOKEN" \
     | jq '.data[] | select(.status == "pending")'
   ```

3. **Verify scheduled time:**
   ```bash
   # Check if scheduled_for is in the past or near future
   curl -X GET "http://localhost:8080/api/v1/notifications/{id}" \
     -H "Authorization: Bearer $TOKEN" \
     | jq '.data.scheduled_for'
   ```

### Notifications Stuck in Pending

- Ensure worker is running
- Check that `scheduled_for` time has passed
- Verify `enabled` field is `true`
- Check worker logs for errors

### Testing Email/SMS Delivery

For email/SMS to actually send, you need valid credentials:
- **Email:** Set `SMTP_USERNAME`, `SMTP_PASSWORD`, etc.
- **SMS:** Set `TWILIO_ACCOUNT_SID`, `TWILIO_AUTH_TOKEN`, etc.

Without valid credentials, notifications will be created but sending will fail.

## Environment Variables for Testing

Set these for faster testing:

```bash
# Process notifications every 10 seconds (instead of 1 minute)
export NOTIFICATION_WORKER_INTERVAL=10s

# Restart your server to apply changes
```

This makes testing much faster!

