#!/usr/bin/env python3
import urllib.request
import json

base = "http://127.0.0.1:18473/api/v1"

# Test setup/first-run
data = json.dumps({"username": "admin", "password": "TestPass123!", "email": "admin@test.local"}).encode()
req = urllib.request.Request(f"{base}/setup/first-run", data=data, headers={"Content-Type": "application/json"})
try:
    resp = urllib.request.urlopen(req)
    print("First-run:", json.loads(resp.read()))
except urllib.error.HTTPError as e:
    print("First-run error:", e.read())