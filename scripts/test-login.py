#!/usr/bin/env python3
import urllib.request
import json

base = "http://127.0.0.1:18473/api/v1"

# Test health
resp = urllib.request.urlopen(f"{base}/health")
print("Health:", json.loads(resp.read()))

# Test login
data = json.dumps({"username": "admin", "password": "testpass123"}).encode()
req = urllib.request.Request(f"{base}/auth/login", data=data, headers={"Content-Type": "application/json"})
try:
    resp = urllib.request.urlopen(req)
    print("Login:", json.loads(resp.read()))
except urllib.error.HTTPError as e:
    print("Login error:", e.read())