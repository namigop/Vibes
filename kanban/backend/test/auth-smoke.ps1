# Exercises the authentication and per-user isolation contract against a
# running API. Exits non-zero on the first failed assertion.
#
#   $env:KANBAN_API = 'http://localhost:8080/api/v1'; ./test/auth-smoke.ps1

$ErrorActionPreference = 'Stop'
$base = if ($env:KANBAN_API) { $env:KANBAN_API } else { 'http://localhost:8081/api/v1' }

$failures = @()
$checks = 0

function Check($name, $cond, $detail = '') {
    $script:checks++
    if ($cond) {
        Write-Output "  PASS  $name"
    } else {
        Write-Output "  FAIL  $name  $detail"
        $script:failures += $name
    }
}

function Api($method, $path, $body = $null, $token = $null) {
    $params = @{ Method = $method; Uri = "$base$path"; TimeoutSec = 20 }
    if ($body -ne $null) {
        $params['Body'] = ($body | ConvertTo-Json -Depth 6)
        $params['ContentType'] = 'application/json'
    }
    if ($token) { $params['Headers'] = @{ Authorization = "Bearer $token" } }
    try {
        $r = Invoke-WebRequest @params -UseBasicParsing
        $json = $null
        if ($r.Content) { $json = $r.Content | ConvertFrom-Json }
        return @{ Status = [int]$r.StatusCode; Body = $json; Raw = $r.Content }
    } catch {
        $content = $_.ErrorDetails.Message
        $json = $null
        if ($content) { try { $json = $content | ConvertFrom-Json } catch { } }
        return @{ Status = [int]$_.Exception.Response.StatusCode; Body = $json; Raw = $content }
    }
}

$demoPassword = if ($env:DEMO_PASSWORD) { $env:DEMO_PASSWORD } else { 'kanban-demo' }

Write-Output "targeting $base"
Write-Output "demo password: $demoPassword"

# ---------------------------------------------------------------------------
Write-Output "`n=== 1. Protected routes reject anonymous callers ==="
$anon = Api GET '/boards'
Check 'GET /boards without a token is 401' ($anon.Status -eq 401) "got $($anon.Status)"
Check '401 uses the unauthorized slug' ($anon.Body.error -eq 'unauthorized') "got '$($anon.Body.error)'"
Check '401 includes a human-readable message' ([bool]$anon.Body.message) "got '$($anon.Body.message)'"

foreach ($probe in @(
    @{ m = 'POST';   p = '/boards';                    b = @{ name = 'x' } },
    @{ m = 'GET';    p = '/auth/me';                   b = $null },
    @{ m = 'DELETE'; p = '/boards/11111111-1111-1111-1111-111111111111'; b = $null }
)) {
    $res = Api $probe.m $probe.p $probe.b
    Check "$($probe.m) $($probe.p) without a token is 401" ($res.Status -eq 401) "got $($res.Status)"
}

# ---------------------------------------------------------------------------
Write-Output "`n=== 2. Login ==="
$bad = Api POST '/auth/login' @{ username = 'alice'; password = 'definitely-wrong' }
Check 'wrong password is 401' ($bad.Status -eq 401) "got $($bad.Status)"
Check 'wrong password is deliberately vague' ($bad.Body.message -eq 'invalid username or password') "got '$($bad.Body.message)'"

$ghost = Api POST '/auth/login' @{ username = 'no-such-person'; password = 'whatever' }
Check 'unknown user gives the identical message' ($ghost.Body.message -eq $bad.Body.message) "got '$($ghost.Body.message)' vs '$($bad.Body.message)'"

$empty = Api POST '/auth/login' @{ username = ''; password = '' }
Check 'empty credentials is 400' ($empty.Status -eq 400) "got $($empty.Status)"

$alice = Api POST '/auth/login' @{ username = 'alice'; password = $demoPassword }
Check 'login as alice is 200' ($alice.Status -eq 200) "got $($alice.Status) $($alice.Raw)"
Check 'login returns a token' ([bool]$alice.Body.token) 'no token in response'
Check 'login returns the user' ($alice.Body.user.username -eq 'alice') "got $($alice.Body.user.username)"
Check 'login response never leaks password_hash' (-not ($alice.Raw -match 'password_hash'))

$aliceToken = $alice.Body.token
$aliceId = $alice.Body.user.id

$caseInsensitive = Api POST '/auth/login' @{ username = 'ALICE'; password = $demoPassword }
Check 'login is case-insensitive on username' ($caseInsensitive.Status -eq 200) "got $($caseInsensitive.Status)"

# ---------------------------------------------------------------------------
Write-Output "`n=== 3. /auth/me ==="
$me = Api GET '/auth/me' $null $aliceToken
Check '/auth/me is 200' ($me.Status -eq 200) "got $($me.Status)"
Check '/auth/me returns the caller' ($me.Body.id -eq $aliceId -and $me.Body.username -eq 'alice')
Check '/auth/me never includes password_hash' (-not ($me.Raw -match 'password_hash'))

$garbage = Api GET '/auth/me' $null 'not-a-real-token'
Check '/auth/me with a garbage token is 401' ($garbage.Status -eq 401) "got $($garbage.Status)"

# ---------------------------------------------------------------------------
Write-Output "`n=== 4. Per-user board isolation ==="
$aliceBoards = Api GET '/boards' $null $aliceToken
Check 'alice sees her boards' ($aliceBoards.Status -eq 200 -and $aliceBoards.Body.Count -gt 0) "got $($aliceBoards.Status) count=$($aliceBoards.Body.Count)"
$aliceBoardIds = @($aliceBoards.Body | ForEach-Object { $_.id })

$bobsBoards = Api POST '/auth/login' @{ username = 'bob'; password = $demoPassword }
Check 'login as bob is 200' ($bobsBoards.Status -eq 200) "got $($bobsBoards.Status)"
$bobToken = $bobsBoards.Body.token
$bobId = $bobsBoards.Body.user.id

Check 'alice and bob are distinct accounts' ($aliceId -ne $bobId) 'both resolved to the same id'

$bobBoards = Api GET '/boards' $null $bobToken
Check 'bob sees his own boards' ($bobBoards.Body.Count -gt 0) "got $($bobBoards.Body.Count)"

$overlap = @($bobBoards.Body.id | Where-Object { $aliceBoardIds -contains $_ })
Check 'the two users share no boards' ($overlap.Count -eq 0) "overlap: $($overlap -join ',')"

# The core guarantee: another user's board must be invisible, not merely denied.
$crossRead = Api GET "/boards/$($aliceBoardIds[0])" $null $bobToken
Check "bob GETting alice's board is 404" ($crossRead.Status -eq 404) "got $($crossRead.Status)"
Check 'the 404 does not confirm the board exists' ($crossRead.Body.error -eq 'not_found') "got '$($crossRead.Body.error)'"

$crossRename = Api PATCH "/boards/$($aliceBoardIds[0])" @{ name = 'stolen' } $bobToken
Check "bob PATCHing alice's board is 404" ($crossRename.Status -eq 404) "got $($crossRename.Status)"

$crossDelete = Api DELETE "/boards/$($aliceBoardIds[0])" $null $bobToken
Check "bob DELETEing alice's board is 404" ($crossDelete.Status -eq 404) "got $($crossDelete.Status)"
Check "alice's board survived bob's delete attempt" ((Api GET "/boards/$($aliceBoardIds[0])" $null $aliceToken).Status -eq 200)

# ---------------------------------------------------------------------------
Write-Output "`n=== 5. Column and card isolation ==="
$aliceDetail = Api GET "/boards/$($aliceBoardIds[0])" $null $aliceToken
Check 'alice can read her own board' ($aliceDetail.Status -eq 200) "got $($aliceDetail.Status)"
$aliceCol = $aliceDetail.Body.columns[0].id
$aliceCard = $aliceDetail.Body.cards[0].id

$colRead = Api GET "/boards/$($aliceBoardIds[0])/columns" $null $bobToken
Check "bob listing alice's columns is 404" ($colRead.Status -eq 404) "got $($colRead.Status)"

$cardPatch = Api PATCH "/cards/$aliceCard" @{ title = 'stolen card' } $bobToken
Check "bob PATCHing alice's card is 404" ($cardPatch.Status -eq 404) "got $($cardPatch.Status)"

$cardDelete = Api DELETE "/cards/$aliceCard" $null $bobToken
Check "bob DELETEing alice's card is 404" ($cardDelete.Status -eq 404) "got $($cardDelete.Status)"

$cardMove = Api POST "/cards/$aliceCard/move" @{ column_id = $aliceCol; position = 0 } $bobToken
Check "bob moving alice's card is 404" ($cardMove.Status -eq 404) "got $($cardMove.Status)"

$cardStill = Api GET "/boards/$($aliceBoardIds[0])" $null $aliceToken
$survivor = @($cardStill.Body.cards | Where-Object { $_.id -eq $aliceCard })
Check 'alice card is untouched' ($survivor.Count -eq 1 -and $survivor[0].title -ne 'stolen card')

# ---------------------------------------------------------------------------
Write-Output "`n=== 6. Owner can still do everything ==="
$newBoard = Api POST '/boards' @{ name = 'Auth Smoke Board' } $aliceToken
Check 'alice can create a board' ($newBoard.Status -eq 201) "got $($newBoard.Status)"
Check 'new board is owned by alice' ($newBoard.Body.owner_id -eq $aliceId) "got $($newBoard.Body.owner_id)"
Check 'new board is seeded with 4 columns' ((Api GET "/boards/$($newBoard.Body.id)" $null $aliceToken).Body.columns.Count -eq 4)

$cleanup = Api DELETE "/boards/$($newBoard.Body.id)" $null $aliceToken
Check 'alice can delete her own board' ($cleanup.Status -eq 204) "got $($cleanup.Status)"

# ---------------------------------------------------------------------------
Write-Output "`n=== 7. Legacy account owns pre-auth boards ==="
$legacy = Api POST '/auth/login' @{ username = 'legacy'; password = $demoPassword }
if ($legacy.Status -eq 200) {
    $legacyToken = $legacy.Body.token
    $legacyBoards = Api GET '/boards' $null $legacyToken
    Check 'legacy can log in' $true
    Check 'legacy sees the boards that predate ownership' ($legacyBoards.Body.Count -gt 0) "got $($legacyBoards.Body.Count)"
    Check 'legacy owns none of the persona boards' ((@($legacyBoards.Body.id | Where-Object { $aliceBoardIds -contains $_ })).Count -eq 0)
} else {
    Check 'legacy can log in' $false "got $($legacy.Status) $($legacy.Raw)"
}

# ---------------------------------------------------------------------------
Write-Output "`n=== 8. Token hygiene ==="
$emptyAuth = Api GET '/boards' $null ''
Check 'empty Authorization header is 401' ($emptyAuth.Status -eq 401) "got $($emptyAuth.Status)"

$basicAuth = Api GET '/boards' $null 'not-even-bearer'
Check 'a non-Bearer scheme is 401' ($basicAuth.Status -eq 401) "got $($basicAuth.Status)"

# A token signed with a different secret must not be accepted.
$forged = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIwMDAwMDAwMC0wMDAwLTAwMDAtMDAwMC0wMDAwMDAwMDAwMDAiLCJpc3MiOiJrYW5iYW4tYXBpIiwiYXVkIjoia2FuYmFuLXdlYiIsImV4cCI6NDEwMjQ0NDgwMH0.forgedsignature'
$forgedRes = Api GET '/boards' $null $forged
Check 'a forged token is 401' ($forgedRes.Status -eq 401) "got $($forgedRes.Status)"

Write-Output "`n================================"
Write-Output "checks run: $checks"
if ($failures.Count -eq 0) {
    Write-Output 'ALL CHECKS PASSED'
    exit 0
} else {
    Write-Output "FAILURES ($($failures.Count)):"
    $failures | ForEach-Object { Write-Output "  - $_" }
    exit 1
}