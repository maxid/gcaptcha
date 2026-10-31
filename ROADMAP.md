# Roadmap

- [x] Protocol compatibility spec (from frontend contract)
- [x] Slider image generation and challenge model
- [x] Percentage track verification (0.02 tolerance)
- [x] Redis store (atomic one-time take, Java `::captcha::` key layout) + in-memory store
- [ ] Rotate / word-click challenge types（按前端实际使用再开，见协议文档 §8）
- [x] Behavioural verification (`WithTrackPolicy`, opt-in; default matches Java position-only)
- [x] Secondary verification (`WithSecondaryVerification`: 120s TTL, `secondary:` key)
- [ ] Security review
- [ ] v1.0 API stabilization
