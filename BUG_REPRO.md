# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

舱区台账列表接口永远是空的，帮我修一下。

现象：

1. POST /api/cabins 登记舱区返回 201，登记内容也都对；但紧接着 GET /api/cabins 始终返回空数组 []，不管登记了 1 个还是 3 个舱区。
2. 按 ID 单查是正常的：GET /api/cabins/{id} 能查到刚登记的舱区，名称、容量、上电状态都对；黑启动演练批准后单查还能看到作业锁持有班组。所以数据确实在系统里，只是列表看不到。
3. 前端的舱区台账页面因此一直空白，运维脚本靠列表判空来决定是否初始化默认舱区，结果每次启动都会把默认舱区重新写一遍。
4. 其他台账的列表（黑启动预案列表）是正常的。

复现：登记两个舱区，然后读舱区台账列表，对比列表条数和单查结果。

期望行为：
- 列表要返回全部已登记舱区，条数等于登记数量，每条都带自己的 id、名称、容量和上电状态；
- 每登记一个舱区，列表条数相应增加；
- 演练批准后列表里该舱区的作业锁持有班组要与单查结果一致。

修完请保证 go test -timeout=120s -count=1 ./... 全绿。

## 含 Bug 版本

- 仓库：11DingKing/go-eb409f-t011-02
- 仓库地址：https://github.com/11DingKing/go-eb409f-t011-02.git
- parent SHA：3303aad175d983fab7791ed263c343c837f2ec27

## 复现步骤

```bash
git clone -- https://github.com/11DingKing/go-eb409f-t011-02.git bug-repro
cd bug-repro
git checkout --detach 3303aad175d983fab7791ed263c343c837f2ec27
go test -timeout=120s ./internal/httpsrv/ -run "TestHTTPCabinLedgerListsRegisteredCabins|TestHTTPCabinLedgerGrowsWithEachRegistration|TestHTTPCabinLedgerReflectsLockState" -count=1 -v
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test -timeout=120s ./internal/httpsrv/ -run "TestHTTPCabinLedgerListsRegisteredCabins|TestHTTPCabinLedgerGrowsWithEachRegistration|TestHTTPCabinLedgerReflectsLockState" -count=1 -v
=== RUN   TestHTTPCabinLedgerListsRegisteredCabins
    cabin_ledger_test.go:54: cabin list has 0 entries, want 2: []
--- FAIL: TestHTTPCabinLedgerListsRegisteredCabins (0.01s)
=== RUN   TestHTTPCabinLedgerGrowsWithEachRegistration
    cabin_ledger_test.go:85: after registering cabin-a: list has 0 entries, want 1: []
--- FAIL: TestHTTPCabinLedgerGrowsWithEachRegistration (0.00s)
=== RUN   TestHTTPCabinLedgerReflectsLockState
    cabin_ledger_test.go:112: cabin list has 0 entries, want 1: []
--- FAIL: TestHTTPCabinLedgerReflectsLockState (0.00s)
FAIL
FAIL	microgrid-dispatch/internal/httpsrv	0.066s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test -timeout=120s ./internal/httpsrv/ -run "TestHTTPCabinLedgerListsRegisteredCabins|TestHTTPCabinLedgerGrowsWithEachRegistration|TestHTTPCabinLedgerReflectsLockState" -count=1 -v
=== RUN   TestHTTPCabinLedgerListsRegisteredCabins
    cabin_ledger_test.go:54: cabin list has 0 entries, want 2: []
--- FAIL: TestHTTPCabinLedgerListsRegisteredCabins (0.00s)
=== RUN   TestHTTPCabinLedgerGrowsWithEachRegistration
    cabin_ledger_test.go:85: after registering cabin-a: list has 0 entries, want 1: []
--- FAIL: TestHTTPCabinLedgerGrowsWithEachRegistration (0.00s)
=== RUN   TestHTTPCabinLedgerReflectsLockState
    cabin_ledger_test.go:112: cabin list has 0 entries, want 1: []
--- FAIL: TestHTTPCabinLedgerReflectsLockState (0.00s)
FAIL
FAIL	microgrid-dispatch/internal/httpsrv	0.002s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

通过标准（bugfix）：
1. 定向复现命令在修复后全部通过：
   go test -timeout=120s ./internal/httpsrv/ -run 'TestHTTPCabinLedgerListsRegisteredCabins|TestHTTPCabinLedgerGrowsWithEachRegistration|TestHTTPCabinLedgerReflectsLockState' -count=1 -v
2. 全量回归通过：go test -timeout=120s -count=1 ./...（本仓库全量可跑，未做范围收敛）
3. go build ./... 与 go vet ./... 均为 exit 0
4. linux/amd64 与 linux/arm64 两个架构下上述命令均通过
5. 断言的是公开 HTTP 行为：GET /api/cabins 的条数与每条内容（id/name/capacity_kwh/powered/lock_holder），以及与 GET /api/cabins/{id} 的一致性；断言的是列表完整内容而不只是长度
6. 不得通过修改或跳过测试、放宽断言使结果变绿
