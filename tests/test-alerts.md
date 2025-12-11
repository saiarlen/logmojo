# Alert Testing Guide

## Prerequisites

1. **Configure Email Settings** in `.env`:
```bash
MONITOR_NOTIFIERS_EMAIL_ENABLED=true
MONITOR_NOTIFIERS_EMAIL_SMTP_HOST=smtp.gmail.com
MONITOR_NOTIFIERS_EMAIL_SMTP_PORT=587
MONITOR_NOTIFIERS_EMAIL_USERNAME=your-email@gmail.com
MONITOR_NOTIFIERS_EMAIL_PASSWORD=your-app-password
MONITOR_NOTIFIERS_EMAIL_TO=test@example.com
```

2. **Start the application**:
```bash
go run .
```

3. **Access the web interface**: http://localhost:7005

## Test 1: System Metric Alerts

### CPU High Alert
1. Go to **Alerts** page → **Create Alert Rule**
2. Fill in:
   - **Name**: "High CPU Test"
   - **Type**: "System Metric Threshold"
   - **Condition**: "CPU Usage High"
   - **Threshold**: 1.0 (very low to trigger easily)
   - **Severity**: "High"
   - **Enable email**: ✓

3. **Trigger the alert** using CPU stress:
```bash
# macOS/Linux - CPU stress test
yes > /dev/null &
yes > /dev/null &
yes > /dev/null &
yes > /dev/null &

# Kill after testing
killall yes
```

### Memory High Alert
1. Create rule:
   - **Name**: "High Memory Test"
   - **Condition**: "Memory Usage High"
   - **Threshold**: 50.0
   - **Email**: ✓

2. **Trigger with memory stress**:
```bash
# Create memory pressure
python3 -c "
import time
data = []
for i in range(1000):
    data.append(' ' * 1024 * 1024)  # 1MB chunks
    time.sleep(0.1)
"
```

### Disk Space Alert
1. Create rule:
   - **Name**: "Low Disk Space Test"
   - **Condition**: "Disk Space Low"
   - **Threshold**: 90.0 (trigger when less than 90% free)
   - **Email**: ✓

## Test 2: Log Pattern Alerts

### Create Test Log Files
```bash
# Create test error logs
mkdir -p ./demo_logs/test-app
echo "$(date) ERROR: Database connection failed" >> ./demo_logs/test-app/error.log
echo "$(date) FATAL: Application crashed unexpectedly" >> ./demo_logs/test-app/error.log
```

### Update config.yaml
Add test app to config:
```yaml
apps:
  - name: "Test App"
    service_name: "test-app"
    logs:
      - name: "Error Log"
        path: "./demo_logs/test-app/error.log"
```

### Create Log Pattern Rule
1. **Name**: "Error Pattern Test"
2. **Type**: "Log Pattern Match"
3. **Log Pattern**: `ERROR|FATAL|CRITICAL`
4. **App Filter**: "Test App"
5. **Log Source Filter**: "Error Log"
6. **Email**: ✓

### Trigger Log Pattern Alert
```bash
# Add new error entries
echo "$(date) ERROR: Authentication failed for user admin" >> ./demo_logs/test-app/error.log
echo "$(date) CRITICAL: System overload detected" >> ./demo_logs/test-app/error.log
```

## Test 3: Exception Detection Alerts

### Create Exception Rule
1. **Name**: "Exception Detection Test"
2. **Type**: "Exception Detection"
3. **App Filter**: "Test App"
4. **Email**: ✓

### Generate Test Exceptions
```bash
# Python exception
echo "$(date) Traceback (most recent call last):" >> ./demo_logs/test-app/error.log
echo "  File \"app.py\", line 42, in main" >> ./demo_logs/test-app/error.log
echo "    result = divide_by_zero()" >> ./demo_logs/test-app/error.log
echo "ZeroDivisionError: division by zero" >> ./demo_logs/test-app/error.log

# Java exception
echo "$(date) Exception in thread \"main\" java.lang.NullPointerException" >> ./demo_logs/test-app/error.log
echo "    at com.example.App.main(App.java:15)" >> ./demo_logs/test-app/error.log

# PHP exception
echo "$(date) PHP Fatal error: Uncaught Error: Call to undefined function" >> ./demo_logs/test-app/error.log

# JavaScript exception
echo "$(date) Uncaught TypeError: Cannot read property 'length' of undefined" >> ./demo_logs/test-app/error.log
```

## Test 4: Service Status Alerts

### Create Service Status Rule
1. **Name**: "Service Status Test"
2. **Type**: "Service Status Change"
3. **Email**: ✓

*Note: This requires systemd services and will trigger when services start/stop/fail*

## Testing Email Notifications

### Gmail App Password Setup
1. Enable 2FA on Gmail
2. Generate App Password: Google Account → Security → App passwords
3. Use the generated password in `MONITOR_NOTIFIERS_EMAIL_PASSWORD`

### Test Email Configuration
```bash
# Test SMTP connection
curl -v --url 'smtps://smtp.gmail.com:465' --ssl-reqd \
  --mail-from 'your-email@gmail.com' \
  --mail-rcpt 'test@example.com' \
  --user 'your-email@gmail.com:your-app-password'
```

## Automated Testing Script

Run this script to trigger multiple alerts:
```bash
#!/bin/bash
echo "Starting Alert Tests..."

# 1. Generate log errors
echo "$(date) ERROR: Test error message" >> ./demo_logs/test-app/error.log
echo "$(date) FATAL: Test fatal error" >> ./demo_logs/test-app/error.log

# 2. Generate exceptions
echo "$(date) Traceback: Test Python exception" >> ./demo_logs/test-app/error.log
echo "$(date) Exception: Test Java exception" >> ./demo_logs/test-app/error.log

# 3. CPU stress (run for 30 seconds)
echo "Generating CPU load..."
yes > /dev/null &
PID1=$!
yes > /dev/null &
PID2=$!

sleep 30

kill $PID1 $PID2
echo "CPU stress test completed"

echo "Check your email and the Alerts page for notifications!"
```

## Verification Steps

1. **Check Alert History**: Go to Alerts → Alert History tab
2. **Check Email**: Look for emails with subject "[SEVERITY] Alert: Rule Name"
3. **Check Logs**: Monitor application logs for alert triggers
4. **Check Database**: 
   ```bash
   sqlite3 monitor.db "SELECT * FROM alerts ORDER BY timestamp DESC LIMIT 10;"
   ```

## Troubleshooting

### No Emails Received
- Check SMTP credentials
- Verify firewall/network settings
- Check spam folder
- Test with different email provider

### Alerts Not Triggering
- Check rule is enabled
- Verify thresholds are appropriate
- Check application logs for errors
- Ensure monitoring intervals are working

### Log Pattern Not Matching
- Test regex patterns online
- Check file permissions
- Verify log file paths in config
- Check app/log filter settings