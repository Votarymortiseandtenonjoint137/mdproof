# API Testing with curl + jq

Build a REST API, start it, exercise every endpoint, and tear it down — all in Markdown. This is mdproof's sweet spot: `curl` for requests, `jq:` assertions for structured validation, persistent session for stateful workflows.

## Steps

### Step 1: Build the test server

A tiny Go HTTP server with `/health` and `/items` CRUD endpoints.

```bash
cd /workspace && go build -o /tmp/testserver ./runbooks/fixtures/testserver
echo "build ok"
```

Expected:

- build ok
- exit_code: 0

### Step 2: Start server and health check

Start in background, save PID for cleanup. Wait for readiness, then verify the health endpoint returns structured JSON.

```bash
/tmp/testserver &
echo $! > /tmp/testserver.pid
sleep 1
curl -sf http://localhost:18080/health
```

Expected:

- jq: .status == "ok"

### Step 3: Create first item (POST)

POST JSON to create a resource. Verify the response includes an auto-generated ID and the submitted name.

```bash
curl -s -X POST http://localhost:18080/items \
  -H "Content-Type: application/json" \
  -d '{"name":"mdproof"}'
```

Expected:

- jq: .id == 1
- jq: .name == "mdproof"

### Step 4: Create second item (POST)

IDs auto-increment. Second item gets ID 2.

```bash
curl -s -X POST http://localhost:18080/items \
  -H "Content-Type: application/json" \
  -d '{"name":"runbook"}'
```

Expected:

- jq: .id == 2
- jq: .name == "runbook"

### Step 5: List all items (GET)

Retrieve the full collection. Verify count and ordering with `jq:` assertions.

```bash
curl -s http://localhost:18080/items
```

Expected:

- jq: . | length == 2
- jq: .[0].name == "mdproof"
- jq: .[1].name == "runbook"

### Step 6: Test invalid input (error handling)

Send malformed JSON. The server should return HTTP 400 with an error message. Good tests cover unhappy paths too.

```bash
HTTP_CODE=$(curl -s -o /tmp/err_body -w '%{http_code}' -X POST http://localhost:18080/items -H "Content-Type: application/json" -d 'not valid json')
echo "HTTP_STATUS:$HTTP_CODE"
cat /tmp/err_body
```

Expected:

- HTTP_STATUS:400
- error

### Step 7: Reset state (DELETE)

Clear all items and verify the collection is empty. Proves stateful operations work across the session.

```bash
curl -s -X DELETE http://localhost:18080/items
curl -s http://localhost:18080/items
```

Expected:

- jq: . | length == 0

### Step 8: Cleanup

Kill the server and remove temp files. Always clean up — even in tests.

```bash
kill $(cat /tmp/testserver.pid) 2>/dev/null
rm -f /tmp/testserver /tmp/testserver.pid
echo "cleanup done"
```

Expected:

- cleanup done
