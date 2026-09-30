# Exercises every endpoint in docs/API.md against a running API.
# Exits non-zero on the first failed assertion.

$ErrorActionPreference = 'Stop'
$base = 'http://localhost:8081/api/v1'
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

function Api($method, $path, $body = $null) {
    $params = @{ Method = $method; Uri = "$base$path"; TimeoutSec = 20 }
    if ($body -ne $null) {
        $params['Body'] = ($body | ConvertTo-Json -Depth 6)
        $params['ContentType'] = 'application/json'
    }
    try {
        $r = Invoke-WebRequest @params -UseBasicParsing
        $code = [int]$r.StatusCode
        $json = $null
        if ($r.Content) { $json = $r.Content | ConvertFrom-Json }
        return @{ Status = $code; Body = $json; Raw = $r }
    } catch {
        # PowerShell 7 surfaces errors as HttpResponseMessage, so the body
        # comes from ErrorDetails rather than a response stream.
        $code = [int]$_.Exception.Response.StatusCode
        $content = $_.ErrorDetails.Message
        $json = $null
        if ($content) { try { $json = $content | ConvertFrom-Json } catch { } }
        return @{ Status = $code; Body = $json; Raw = $content }
    }
}

# Positions must always be a gap-free 0..n-1 run within their parent.
function CheckGapFree($items, $label) {
    $positions = @($items | ForEach-Object { $_.position } | Sort-Object)
    $expected = @(0..($positions.Count - 1))
    $ok = ($positions.Count -eq $expected.Count) -and (-not (Compare-Object $positions $expected))
    Check "$label positions are gap-free 0..n-1" $ok "got: $($positions -join ',')"
}

Write-Output "`n=== 0. Cleanup from previous runs ==="
# An interrupted earlier run can leave a smoke board behind, which would make
# the final board count assertion fail for reasons unrelated to the API.
$stale = Api GET '/boards'
foreach ($b in @($stale.Body)) {
    if ($b.name -like 'Smoke*') {
        $null = Api DELETE "/boards/$($b.id)"
        Write-Output "  removed stale board: $($b.name)"
    }
}
$baseline = (Api GET '/boards').Body.Count
Write-Output "  baseline boards: $baseline"

Write-Output "`n=== 1. Health ==="
$h = Api GET '/health'
Check 'health 200 ok' ($h.Status -eq 200 -and $h.Body.status -eq 'ok' -and $h.Body.db -eq 'ok') "got $($h.Status)"

Write-Output "`n=== 2. Board CRUD ==="
$created = Api POST '/boards' @{ name = 'Smoke Test Board' }
Check 'create board 201' ($created.Status -eq 201) "got $($created.Status)"
$boardId = $created.Body.id
Check 'create board returns name' ($created.Body.name -eq 'Smoke Test Board')

$renamed = Api PATCH "/boards/$boardId" @{ name = 'Smoke Renamed' }
Check 'rename board 200' ($renamed.Status -eq 200 -and $renamed.Body.name -eq 'Smoke Renamed') "got $($renamed.Status) $($renamed.Body | ConvertTo-Json -Compress)"

$detail = Api GET "/boards/$boardId"
Check 'board detail 200' ($detail.Status -eq 200) "got $($detail.Status)"
Check 'new board seeded 4 default columns' ($detail.Body.columns.Count -eq 4) "got $($detail.Body.columns.Count)"
Check 'default column names correct' (($detail.Body.columns.name -join ',') -eq 'To Do,In Progress,Review,Done') "got $($detail.Body.columns.name -join ',')"
Check 'new board has zero cards' ($detail.Body.cards.Count -eq 0)
Check 'columns are gap-free' ((@($detail.Body.columns.position | Sort-Object) -join ',') -eq '0,1,2,3')

Write-Output "`n=== 3. Custom column (deal stage) ==="
$stage = Api POST "/boards/$boardId/columns" @{ name = 'Negotiation' }
Check 'create custom column 201' ($stage.Status -eq 201) "got $($stage.Status)"
$negId = $stage.Body.id
Check 'custom column appended at end' ($stage.Body.position -eq 4) "got $($stage.Body.position)"

$renamedCol = Api PATCH "/columns/$negId" @{ name = 'Contracting' }
Check 'rename column 200' ($renamedCol.Status -eq 200 -and $renamedCol.Body.name -eq 'Contracting')

Write-Output "`n=== 4. Column reorder ==="
$cm = Api POST "/columns/$negId/move" @{ position = 0 }
Check 'move column 200' ($cm.Status -eq 200) "got $($cm.Status)"
Check 'move column returns full reordered list' ($cm.Body.Count -eq 5) "got $($cm.Body.Count)"
Check 'moved column is now first' ($cm.Body[0].id -eq $negId) "got $($cm.Body[0].name)"
Check 'column order gap-free after move' ((@($cm.Body.position | Sort-Object) -join ',') -eq '0,1,2,3,4') "got $((@($cm.Body.position | Sort-Object)) -join ',')"

$todoCol = ($cm.Body | Where-Object { $_.name -eq 'To Do' }).id
$inProgCol = ($cm.Body | Where-Object { $_.name -eq 'In Progress' }).id
$reviewCol = ($cm.Body | Where-Object { $_.name -eq 'Review' }).id
$doneCol = ($cm.Body | Where-Object { $_.name -eq 'Done' }).id
$contractCol = $negId

Write-Output "`n=== 5. Card CRUD ==="
$c1 = Api POST "/columns/$todoCol/cards" @{ title = 'Card A'; description = 'first'; assignee = 'Erik'; priority = 'high'; due_date = '2026-10-15' }
Check 'create card 201' ($c1.Status -eq 201) "got $($c1.Status)"
Check 'due_date round-trips as calendar date' ($c1.Body.due_date -eq '2026-10-15') "got '$($c1.Body.due_date)'"
$c2 = Api POST "/columns/$todoCol/cards" @{ title = 'Card B'; priority = 'low' }
$c3 = Api POST "/columns/$todoCol/cards" @{ title = 'Card C'; priority = 'urgent'; assignee = 'Priya' }
# No priority supplied at all, so the server default must apply.
$cDefault = Api POST "/columns/$todoCol/cards" @{ title = 'Card Default' }
Check 'omitted priority defaults to medium' ($cDefault.Body.priority -eq 'medium') "got '$($cDefault.Body.priority)'"
Api DELETE "/cards/$($cDefault.Body.id)" | Out-Null
Check 'null due_date stays null' ($null -eq $c2.Body.due_date)

$c4 = Api POST "/columns/$todoCol/cards" @{ title = 'Card D' }
$upd = Api PATCH "/cards/$($c4.Body.id)" @{ title = 'Card D updated'; description = 'now with detail'; assignee = 'Marco'; priority = 'urgent'; due_date = '2026-12-31' }
Check 'update card 200' ($upd.Status -eq 200) "got $($upd.Status)"
Check 'update applies all fields' ($upd.Body.title -eq 'Card D updated' -and $upd.Body.priority -eq 'urgent' -and $upd.Body.due_date -eq '2026-12-31' -and $upd.Body.assignee -eq 'Marco')

Write-Output "`n=== 6. Card move: across columns ==="
$before = Api GET "/boards/$boardId"
$todoBefore = @($before.Body.cards | Where-Object { $_.column_id -eq $todoCol })
Check '4 cards in To Do' ($todoBefore.Count -eq 4) "got $($todoBefore.Count)"
CheckGapFree $todoBefore 'To Do'

# Move Card A (index 0) into 'Review' at index 0.
$move = Api POST "/cards/$($c1.Body.id)/move" @{ column_id = $reviewCol; position = 0 }
Check 'move card across columns 200' ($move.Status -eq 200) "got $($move.Status)"
Check 'moved card reports new column' ($move.Body.column_id -eq $reviewCol)
Check 'moved card settled at position 0' ($move.Body.position -eq 0) "got $($move.Body.position)"

$after = Api GET "/boards/$boardId"
$todoAfter = @($after.Body.cards | Where-Object { $_.column_id -eq $todoCol })
$reviewAfter = @($after.Body.cards | Where-Object { $_.column_id -eq $reviewCol })
Check 'source column lost the card' ($todoAfter.Count -eq 3) "got $($todoAfter.Count)"
Check 'destination column gained the card' ($reviewAfter.Count -eq 1) "got $($reviewAfter.Count)"
CheckGapFree $todoAfter 'To Do after move'
CheckGapFree $reviewAfter 'Review after move'

Write-Output "`n=== 7. Card move: within column ==="
# To Do now holds B(0) C(1) D(2). Move D to the front.
$move2 = Api POST "/cards/$($c4.Body.id)/move" @{ column_id = $todoCol; position = 0 }
Check 'reorder within column 200' ($move2.Status -eq 200)
Check 'reordered card landed at 0' ($move2.Body.position -eq 0)

$after2 = Api GET "/boards/$boardId"
$todoNow = @($after2.Body.cards | Where-Object { $_.column_id -eq $todoCol } | Sort-Object position)
Check 'reorder put Card D first' ($todoNow[0].id -eq $c4.Body.id) "got $($todoNow[0].title)"
CheckGapFree $todoNow 'To Do after reorder'

Write-Output "`n=== 8. Move clamping (out-of-range positions) ==="
$clamp = Api POST "/cards/$($c2.Body.id)/move" @{ column_id = $doneCol; position = 999 }
Check 'huge position is clamped, not rejected' ($clamp.Status -eq 200 -and $clamp.Body.position -eq 0) "got $($clamp.Status) pos=$($clamp.Body.position)"
$neg = Api POST "/cards/$($c2.Body.id)/move" @{ column_id = $todoCol; position = -5 }
Check 'negative position is clamped' ($neg.Status -eq 200 -and $neg.Body.position -eq 0) "got $($neg.Status) pos=$($neg.Body.position)"

Write-Output "`n=== 9. Validation and error handling ==="
$e1 = Api POST "/columns/$todoCol/cards" @{ title = '   ' }
Check 'blank title 400 validation_failed' ($e1.Status -eq 400 -and $e1.Body.error -eq 'validation_failed') "got $($e1.Status) $($e1.Raw)"
$e2 = Api POST "/columns/$todoCol/cards" @{ title = 'x'; priority = 'catastrophic' }
Check 'bad priority 400' ($e2.Status -eq 400) "got $($e2.Status) $($e2.Raw)"
$e3 = Api POST "/columns/$todoCol/cards" @{ title = 'x'; due_date = '15-10-2026' }
Check 'malformed due_date 400' ($e3.Status -eq 400) "got $($e3.Status) $($e3.Raw)"
$e4 = Api GET "/boards/not-a-uuid"
Check 'malformed uuid 400' ($e4.Status -eq 400) "got $($e4.Status)"
$e5 = Api GET '/boards/11111111-1111-1111-1111-111111111111'
Check 'unknown board 404 not_found' ($e5.Status -eq 404 -and $e5.Body.error -eq 'not_found') "got $($e5.Status) $($e5.Raw)"
$e6 = Api DELETE '/cards/11111111-1111-1111-1111-111111111111'
Check 'delete unknown card 404' ($e6.Status -eq 404) "got $($e6.Status)"
$e7 = Api POST '/boards' @{ name = '' }
Check 'blank board name 400' ($e7.Status -eq 400) "got $($e7.Status)"

Write-Output "`n=== 10. Cascade delete ==="
$del = Api DELETE "/columns/$contractCol"
Check 'delete column 204' ($del.Status -eq 204) "got $($del.Status)"
$after3 = Api GET "/boards/$boardId"
Check 'column gone' (-not ($after3.Body.columns.id -contains $contractCol))

$delCard = Api DELETE "/cards/$($c3.Body.id)"
Check 'delete card 204' ($delCard.Status -eq 204) "got $($delCard.Status)"

$delBoard = Api DELETE "/boards/$boardId"
Check 'delete board 204' ($delBoard.Status -eq 204) "got $($delBoard.Status)"
$after4 = Api GET "/boards/$boardId"
Check 'deleted board now 404' ($after4.Status -eq 404) "got $($after4.Status)"

$boardsNow = Api GET '/boards'
Check 'board count back to baseline after cleanup' ($boardsNow.Body.Count -eq $baseline) "got $($boardsNow.Body.Count), expected $baseline"
Check 'seeded boards still present' (@($boardsNow.Body.name) -contains 'Product Roadmap' -and @($boardsNow.Body.name) -contains 'Sales Pipeline') "got $((@($boardsNow.Body.name)) -join ',')"

Write-Output "`n=== 11. CORS preflight ==="
try {
    $pre = Invoke-WebRequest -Method OPTIONS -Uri "$base/boards" -Headers @{ Origin = 'http://localhost:3000' } -UseBasicParsing -TimeoutSec 15
    Check 'preflight returns 204' ([int]$pre.StatusCode -eq 204) "got $($pre.StatusCode)"
    Check 'preflight echoes allowed origin' ($pre.Headers['Access-Control-Allow-Origin'] -eq 'http://localhost:3000') "got $($pre.Headers['Access-Control-Allow-Origin'])"
} catch {
    Check 'preflight 204' $false $_.Exception.Message
}

Write-Output "`n================================"
Write-Output "checks run: $checks"
if ($failures.Count -eq 0) {
    Write-Output "ALL CHECKS PASSED"
    exit 0
} else {
    Write-Output "FAILURES ($($failures.Count)):"
    $failures | ForEach-Object { Write-Output "  - $_" }
    exit 1
}
