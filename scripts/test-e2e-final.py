#!/usr/bin/env python3
import urllib.request
import json
import time

base = "http://127.0.0.1:18473/api/v1"

print("=== Juvia E2E Test on Clean Server ===\n")

# Check setup status
resp = urllib.request.urlopen(f"{base}/setup/status")
setup_data = json.loads(resp.read())
print(f"Setup required: {setup_data['data']['setup_required']}")

if setup_data['data']['setup_required']:
    # Run first-run
    data = json.dumps({"username": "admin", "password": "TestPass123!", "email": "admin@test.local"}).encode()
    req = urllib.request.Request(f"{base}/setup/first-run", data=data, headers={"Content-Type": "application/json"})
    resp = urllib.request.urlopen(req)
    print(f"First-run: {json.loads(resp.read())}\n")

# Login
data = json.dumps({"username": "admin", "password": "TestPass123!"}).encode()
req = urllib.request.Request(f"{base}/auth/login", data=data, headers={"Content-Type": "application/json"})
resp = urllib.request.urlopen(req)
result = json.loads(resp.read())
token = result["data"]["access_token"]
print(f"Login: OK (token: {token[:20]}...)\n")

# Create website
domain = f"test{int(time.time())}.com"
print(f"Creating website: {domain}")
data = json.dumps({"domain": domain, "php_version": "8.2", "web_server": "nginx"}).encode()
req = urllib.request.Request(f"{base}/websites", data=data, headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
try:
    resp = urllib.request.urlopen(req)
    result = json.loads(resp.read())
    print(f"Create website: {result}\n")
    website_id = result["data"]["website_id"]
except urllib.error.HTTPError as e:
    print(f"Create website error: {e.read()}\n")
    website_id = None

# List websites
print("Listing websites:")
req = urllib.request.Request(f"{base}/websites", headers={"Authorization": f"Bearer {token}"})
resp = urllib.request.urlopen(req)
result = json.loads(resp.read())
print(f"Total websites: {result['meta']['total']}")
for w in result['data']:
    print(f"  - {w['domain']} (id={w['id']})")

print("\n=== Test Complete ===")