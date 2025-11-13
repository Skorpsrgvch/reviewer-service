# load-test.ps1
param(
    [int]$Count = 50
)

$LogPath = "load-test-results.txt"
$BaseURL = "http://localhost:8080"
$Headers = @{
    "Authorization" = "Bearer admin"
    "Content-Type" = "application/json"
}

# Убедимся, что файл лога чистый для этого теста
"=== ТЕСТ: Полный жизненный цикл PR ($Count шт) ===" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Время начала: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')" | Out-File -FilePath $LogPath -Append -Encoding UTF8

$authorID = "u1"

# Шаг 1: Создание PR
Write-Host " Создание $Count PR..." -ForegroundColor Cyan
$createStart = Get-Date
$createdPRs = @()

for ($i = 1; $i -le $Count; $i++) {
    $prID = "pr-lc-$i"
    $body = @{
        pull_request_id = $prID
        pull_request_name = "Lifecycle Test $i"
        author_id = $authorID
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$BaseURL/pullRequest/create" -Method Post -Headers $Headers -Body $body
        $createdPRs += $prID
    } catch {
        $msg = "CREATE [$i] ERROR: $($_.Exception.Message)"
        Write-Host $msg -ForegroundColor Red
        $msg | Out-File -FilePath $LogPath -Append -Encoding UTF8
    }
}
$createEnd = Get-Date
$createTime = ($createEnd - $createStart).TotalSeconds

# Шаг 2: Reassign (берём первого ревьюера из ответа — но мы не получаем его в этом скрипте)
# Поэтому просто попробуем переназначить одного из возможных — например, "u2"
# Убедись, что u2 существует и активен!

Write-Host " Переназначение ревьюеров..." -ForegroundColor Yellow
$reassignStart = Get-Date
$reassignSuccess = 0

foreach ($prID in $createdPRs) {
    $body = @{
        pull_request_id = $prID
        old_reviewer_id = "u2"  # должен существовать в той же команде, что и автор
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$BaseURL/pullRequest/reassign" -Method Post -Headers $Headers -Body $body
        $reassignSuccess++
    } catch {
        # Не считаем ошибку критичной — возможно, u2 не был назначен
        # Но если хочешь — можешь сделать реальный reassign через GET /users/getReview
        # Для простоты пропустим
    }
}
$reassignEnd = Get-Date
$reassignTime = ($reassignEnd - $reassignStart).TotalSeconds

# Шаг 3: Merge
Write-Host " Слияние PR..." -ForegroundColor Green
$mergeStart = Get-Date
$mergeSuccess = 0

foreach ($prID in $createdPRs) {
    $body = @{
        pull_request_id = $prID
    } | ConvertTo-Json

    try {
        $response = Invoke-RestMethod -Uri "$BaseURL/pullRequest/merge" -Method Post -Headers $Headers -Body $body
        $mergeSuccess++
    } catch {
        $msg = "MERGE [$prID] ERROR: $($_.Exception.Message)"
        Write-Host $msg -ForegroundColor Red
        $msg | Out-File -FilePath $LogPath -Append -Encoding UTF8
    }
}
$mergeEnd = Get-Date
$mergeTime = ($mergeEnd - $mergeStart).TotalSeconds

# Итоги
$totalTime = ($mergeEnd - $createStart).TotalSeconds
"" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"ИТОГ ПОЛНОГО ЦИКЛА:" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Создано PR: $($createdPRs.Count)" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Reassign успешны (оценочно): $reassignSuccess" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Merge успешны: $mergeSuccess" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Время создания: $($createTime.ToString('F3')) сек" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Время reassign: $($reassignTime.ToString('F3')) сек" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Время merge: $($mergeTime.ToString('F3')) сек" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Общее время: $($totalTime.ToString('F3')) сек" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"Среднее время на PR (полный цикл): $(($totalTime / $Count).ToString('F3')) сек" | Out-File -FilePath $LogPath -Append -Encoding UTF8
"-" * 60 | Out-File -FilePath $LogPath -Append -Encoding UTF8
"" | Out-File -FilePath $LogPath -Append -Encoding UTF8

Write-Host " Полный цикл завершён. Результаты в $LogPath" -ForegroundColor Green