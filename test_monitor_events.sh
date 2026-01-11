#!/bin/bash
# Test script for DeviceMonitor events
# Run this after starting a session with Device Monitor enabled

DEVICE_ID=${1:-$(adb devices | grep -v "List" | grep "device$" | head -1 | cut -f1)}

if [ -z "$DEVICE_ID" ]; then
    echo "No device found!"
    exit 1
fi

echo "Testing DeviceMonitor events on device: $DEVICE_ID"
echo "================================================"

# 1. Activity events
echo ""
echo "[1/6] Testing activity_start / activity_displayed..."
adb -s $DEVICE_ID shell am start -n com.android.settings/.Settings
sleep 3

echo ""
echo "[2/6] Testing activity switch (resume)..."
adb -s $DEVICE_ID shell am start -n com.android.calculator2/.Calculator 2>/dev/null || \
adb -s $DEVICE_ID shell am start -a android.intent.action.DIAL
sleep 3

# 2. Screen events
echo ""
echo "[3/6] Testing screen_change (off)..."
adb -s $DEVICE_ID shell input keyevent KEYCODE_POWER
sleep 3

echo ""
echo "[4/6] Testing screen_change (on)..."
adb -s $DEVICE_ID shell input keyevent KEYCODE_WAKEUP
sleep 1
adb -s $DEVICE_ID shell input keyevent 82
sleep 3

# 3. Battery events
echo ""
echo "[5/6] Testing battery_change (low battery warning)..."
adb -s $DEVICE_ID shell dumpsys battery set level 10
sleep 5

echo ""
echo "[6/6] Resetting battery..."
adb -s $DEVICE_ID shell dumpsys battery reset
sleep 2

# 4. App stop event
echo ""
echo "[Bonus] Testing app_stop..."
adb -s $DEVICE_ID shell am force-stop com.android.settings
sleep 2

echo ""
echo "================================================"
echo "Test complete! Check the Session Timeline for events:"
echo "  - device: battery_change, screen_change"
echo "  - app: activity_start, activity_displayed, activity_resume, app_stop"
echo ""
echo "Note: network_change requires toggling WiFi/airplane mode manually"
