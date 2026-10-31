English | [简体中文](README.zh-CN.md)

# gcaptcha

A behavior-captcha server library in Go: slider challenge generation,
percentage-based track verification and one-time tokens.

> Status: **development snapshot** (v0.0.0-dev). Slider captcha is implemented;
> rotate/click, track encryption and secondary verification are roadmap items.

## Features

- Slider challenge: textured background + cut-out template, `data:image/png;base64` URLs
- Percentage validation with the platform-compatible 0.02 tolerance
- One-time challenges (`GET`+`DEL` in Redis, or an in-memory store)
- Pluggable `Store`; Redis store included

## Usage

```go
store := gcaptcha.NewRedisStore("127.0.0.1:6379", "", 0, "")
svc := gcaptcha.New(store) // 600x360 background, 110px piece, 20s TTL

challenge, err := svc.Generate(ctx)          // return challenge to the client
ok, err := svc.Validate(ctx, challenge.ID, track) // single use
```

`track` mirrors the frontend payload (`bgImageWidth`, `trackList[].x/...`);
the percentage is `(last.x - first.x) / bgImageWidth`.

## Install

```bash
go get github.com/maxid/gcaptcha
```

## License

MIT. See [LICENSE](LICENSE).
