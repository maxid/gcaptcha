# gcaptcha

A behavior-captcha server library in Go, compatible with the tianai-captcha protocol used by the XDMP frontend.

> Status: **experimental scaffold** (v0.0.0-dev). This is a **clean-room** implementation: no third-party source code is copied.

## Planned features

- Slider / rotate / click captcha challenge generation
- Track encryption and similarity verification
- Redis-backed one-time tokens
- Secondary verification support

## Install

```bash
go get github.com/maxid/gcaptcha
```

## License

MIT. See [LICENSE](LICENSE).
