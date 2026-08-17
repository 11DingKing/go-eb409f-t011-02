# BENZHI_README

## 项目说明

- 项目：11DingKing/go-eb409f-t011-02
- 项目用途：一个自包含的 Go 后端服务，用于在并网/离网切换窗口内统筹构网型储能舱区的 黑启动演练与日常运维，确保黑启动演练、储能舱巡检、设备异常隔离与维修工单 四条流程互不干扰。
- Go 工具链：`golang:1.26`
- 前端工具链：无

## 标准构建、运行和测试命令

进入容器后执行：

```bash
# 编译
cd '/app' && GOTOOLCHAIN=local go build ./...

# 启动
cd '/app' && GOTOOLCHAIN=local go run ./cmd/server

# 测试
cd '/app' && GOTOOLCHAIN=local go test ./...
```

## Docker 构建和进入容器

```bash
chmod +x build_benzhi_docker.sh
./build_benzhi_docker.sh benzhi-task-42-amd64 linux/amd64
./build_benzhi_docker.sh benzhi-task-42-arm64 linux/arm64
docker run -it benzhi-task-42-amd64:latest
docker run -it --platform linux/arm64 benzhi-task-42-arm64:latest
```

## 题目验证命令

1. 预期退出码 0：`go test -timeout=120s ./internal/httpsrv/ -run "TestHTTPCabinLedgerListsRegisteredCabins|TestHTTPCabinLedgerGrowsWithEachRegistration|TestHTTPCabinLedgerReflectsLockState" -count=1 -v`
2. 预期退出码 0：`go test -buildvcs=false -count=1 ./...`
3. 预期退出码 0：`GOTOOLCHAIN=local go build -buildvcs=false ./... && GOTOOLCHAIN=local go vet ./...`

## Bug 复现

Bug 现象、触发步骤和完整错误信息见 `BUG_REPRO.md`。
