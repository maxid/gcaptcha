[English](README.md) | 简体中文

# gcaptcha

Go 的行为验证码服务端库：滑块挑战生成、
基于百分比的轨迹校验与一次性令牌。

> 状态：**开发快照**（v0.0.0-dev）。滑块验证码已实现；
> 旋转/点选、轨迹加密与二次校验属于路线图事项。

## 特性

- 滑块挑战：纹理背景 + 抠图模板，`data:image/png;base64` URL
- 带平台兼容 0.02 容差的百分比校验
- 一次性挑战（在 Redis 中 `GET`+`DEL`，或使用内存存储）
- 可插拔的 `Store`；内置 Redis 存储

## 用法

```go
store := gcaptcha.NewRedisStore("127.0.0.1:6379", "", 0, "")
svc := gcaptcha.New(store) // 600x360 background, 110px piece, 20s TTL

challenge, err := svc.Generate(ctx)          // return challenge to the client
ok, err := svc.Validate(ctx, challenge.ID, track) // single use
```

`track` 与前端载荷保持一致（`bgImageWidth`、`trackList[].x/...`）；
百分比为 `(last.x - first.x) / bgImageWidth`。

## 安装

```bash
go get github.com/maxid/gcaptcha
```

## 许可证

MIT。见 [LICENSE](LICENSE)。
