#!/bin/bash

echo "Logmojo Alert Testing Script"
echo "=================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Create test directories
echo -e "${BLUE}Setting up test environment...${NC}"
mkdir -p ./demo_logs/test-app
mkdir -p ./demo_logs/app-one
mkdir -p ./demo_logs/app-two

# Function to add log entry with timestamp
add_log() {
    local file=$1
    local message=$2
    echo "$(date '+%Y-%m-%d %H:%M:%S') $message" >> "$file"
}

echo -e "${YELLOW}Test 1: Log Pattern Alerts${NC}"
echo "Generating error log entries..."

# Test App errors
add_log "./demo_logs/test-app/error.log" "ERROR: Database connection timeout after 30 seconds"
add_log "./demo_logs/test-app/error.log" "FATAL: Unable to allocate memory for buffer"
add_log "./demo_logs/test-app/error.log" "CRITICAL: Disk space critically low"
add_log "./demo_logs/test-app/error.log" "WARNING: High CPU usage detected"

# App One errors
add_log "./demo_logs/app/error.log" "ERROR: Authentication failed for user 'admin'"
add_log "./demo_logs/app/error.log" "ERROR: Invalid configuration parameter 'max_connections'"

# App Two errors  
add_log "./demo_logs/app-two/error.log" "FATAL: Service startup failed"
add_log "./demo_logs/app-two/error.log" "ERROR: Network connection refused"

echo -e "${GREEN}✓ Log pattern test entries created${NC}"

echo -e "${YELLOW}Test 2: Exception Detection${NC}"
echo "Generating exception entries..."

# Python exceptions
add_log "./demo_logs/test-app/error.log" "Traceback (most recent call last):"
add_log "./demo_logs/test-app/error.log" "  File \"app.py\", line 42, in process_data"
add_log "./demo_logs/test-app/error.log" "    result = data['missing_key']"
add_log "./demo_logs/test-app/error.log" "KeyError: 'missing_key'"

# Java exceptions
add_log "./demo_logs/app/error.log" "Exception in thread \"main\" java.lang.NullPointerException"
add_log "./demo_logs/app/error.log" "    at com.example.UserService.validateUser(UserService.java:45)"
add_log "./demo_logs/app/error.log" "    at com.example.AuthController.login(AuthController.java:23)"

# PHP exceptions
add_log "./demo_logs/app-two/error.log" "PHP Fatal error: Uncaught Error: Call to undefined function mysql_connect()"
add_log "./demo_logs/app-two/error.log" "Stack trace:"
add_log "./demo_logs/app-two/error.log" "#0 /var/www/html/database.php(15): connect_db()"

# JavaScript exceptions
add_log "./demo_logs/test-app/error.log" "Uncaught TypeError: Cannot read property 'length' of undefined"
add_log "./demo_logs/test-app/error.log" "    at processArray (main.js:34:12)"
add_log "./demo_logs/test-app/error.log" "    at Object.handleRequest (app.js:89:5)"

# Go panics
add_log "./demo_logs/app/error.log" "panic: runtime error: invalid memory address or nil pointer dereference"
add_log "./demo_logs/app/error.log" "goroutine 1 [running]:"
add_log "./demo_logs/app/error.log" "main.processRequest(0x0, 0x0, 0x0)"

# Ruby exceptions
add_log "./demo_logs/app-two/error.log" "Traceback (most recent call last):"
add_log "./demo_logs/app-two/error.log" "    2: from /app/main.rb:15:in '<main>'"
add_log "./demo_logs/app-two/error.log" "    1: from /app/user.rb:8:in 'find_user'"
add_log "./demo_logs/app-two/error.log" "/app/user.rb:8:in 'find': undefined method 'name' for nil:NilClass (NoMethodError)"

echo -e "${GREEN}✓ Exception detection test entries created${NC}"

echo -e "${YELLOW}Test 3: System Metric Stress Test${NC}"
echo "Starting CPU stress test (30 seconds)..."

# CPU stress test
stress_cpu() {
    echo "Generating CPU load..."
    # Create multiple background processes to stress CPU
    for i in {1..4}; do
        yes > /dev/null &
        PIDS+=($!)
    done
    
    echo "CPU stress running for 30 seconds..."
    sleep 30
    
    # Kill stress processes
    for pid in "${PIDS[@]}"; do
        kill $pid 2>/dev/null
    done
    
    echo -e "${GREEN}✓ CPU stress test completed${NC}"
}

# Memory stress test
stress_memory() {
    echo "Starting memory stress test..."
    python3 -c "
import time
print('Allocating memory...')
data = []
try:
    for i in range(500):
        data.append(' ' * 1024 * 1024)  # 1MB chunks
        if i % 50 == 0:
            print(f'Allocated {i}MB')
        time.sleep(0.1)
except KeyboardInterrupt:
    pass
print('Memory stress test completed')
" &
    MEMORY_PID=$!
    
    sleep 20
    kill $MEMORY_PID 2>/dev/null
    echo -e "${GREEN}✓ Memory stress test completed${NC}"
}

# Run stress tests
stress_cpu
stress_memory

echo -e "${YELLOW}Test 4: Additional Log Scenarios${NC}"

# Security-related logs
add_log "./demo_logs/test-app/error.log" "SECURITY: Failed login attempt from IP 192.168.1.100"
add_log "./demo_logs/test-app/error.log" "SECURITY: Suspicious file access detected"
add_log "./demo_logs/app/error.log" "AUTH_ERROR: Invalid JWT token provided"

# Performance issues
add_log "./demo_logs/app-two/error.log" "PERFORMANCE: Query execution time exceeded 5000ms"
add_log "./demo_logs/app/error.log" "TIMEOUT: Request timeout after 30 seconds"

# System issues
add_log "./demo_logs/test-app/error.log" "SYSTEM: Disk I/O error on /dev/sda1"
add_log "./demo_logs/app-two/error.log" "NETWORK: Connection pool exhausted"

echo -e "${GREEN}✓ Additional test scenarios created${NC}"

echo ""
echo -e "${BLUE}Test Summary:${NC}"
echo "📁 Log files created with test data:"
echo "   - ./demo_logs/test-app/error.log"
echo "   - ./demo_logs/app/error.log" 
echo "   - ./demo_logs/app-two/error.log"
echo ""
echo -e "${BLUE}Next Steps:${NC}"
echo "1. Open http://localhost:7005/alerts"
echo "2. Create alert rules for each test type"
echo "3. Configure email settings in .env file"
echo "4. Wait 30-60 seconds for monitoring to detect issues"
echo "5. Check Alert History tab for triggered alerts"
echo "6.  Check your email for notifications"
echo ""
echo -e "${YELLOW}Sample Alert Rules to Create:${NC}"
echo "• Log Pattern: Pattern='ERROR|FATAL|CRITICAL', App='Test App'"
echo "• Exception Detection: Type='Exception Detection', App='App One'"
echo "• CPU Alert: Condition='CPU Usage High', Threshold=1.0"
echo "• Memory Alert: Condition='Memory Usage High', Threshold=50.0"
echo ""
echo -e "${GREEN}🎉 Alert testing setup complete!${NC}"