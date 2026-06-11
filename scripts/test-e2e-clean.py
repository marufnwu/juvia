#!/usr/bin/env python3
import urllib.request
import json

base = "http://127.0.0.1:18473/api/v1"

# Test login
data = json.dumps({"username": "admin", "password": "TestPass123!"}).encode()
req = urllib.request.Request(f"{base}/auth/login", data=data, headers={"Content-Type": "application/json"})
try:
    resp = urllib.request.urlopen(req)
    result = json.loads(resp.read())
    print("Login:", result)
    token = result["data"]["access_token"]
except urllib.error.HTTPError as e:
    print("Login error:", e.read())
    exit(1)

# Test create website
import time
domain = f"test{int(time.time())}.com"
print(f"\nCreating website: {domain}")
data = json.dumps({"domain": domain, "php_version": "8.2", "web_server": "nginx"}).encode()
req = urllib.request.Request(f"{base}/websites", data=data, headers={"Content-Type": "application/json", "Authorization": f"Bearer {token}"})
try:
    resp = urllib.request.urlopen(req)
    result = json.loads(resp.read())
    print("Create website:", result)
except urllib.error.HTTPError as e:
    print("Create website error:", e.read())

# List websites
print("\nListing websites:")
req = urllib.request.Request(f"{base}/websites", headers={"Authorization": f"Bearer {token}"})
try:
    resp = urllib.request.urlopen(req)
    result = json.loads(resp.read())
    print(json.dumps(result, indent=2))
except urllib.error.HTTPError as e:
    print("List websites error:", e.read())