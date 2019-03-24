

```
cat ./test.http | vegeta attack -format=http -duration=30s -lazy -rate=1000 | tee results.bin | \
  vegeta report

with cuckoo

Requests      [total, rate]            30000, 1000.02
Duration      [total, attack, wait]    29.999490432s, 29.999322s, 168.432µs
Latencies     [mean, 50, 95, 99, max]  235.216µs, 179.43µs, 234.477µs, 356.812µs, 36.749449ms
Bytes In      [total, mean]            509658, 16.99
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:30000


without cuckoo

Requests      [total, rate]            30000, 1000.01
Duration      [total, attack, wait]    29.99998436s, 29.999624s, 360.36µs
Latencies     [mean, 50, 95, 99, max]  5.290563ms, 343.184µs, 23.486891ms, 49.801748ms, 157.038286ms
Bytes In      [total, mean]            509658, 16.99
Bytes Out     [total, mean]            0, 0.00
Success       [ratio]                  100.00%
Status Codes  [code:count]             200:30000
```