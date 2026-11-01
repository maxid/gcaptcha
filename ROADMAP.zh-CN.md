[English](ROADMAP.md) | 简体中文

# 路线图

- [x] 协议兼容规范（来自前端契约）
- [x] 滑块图片生成与 challenge 模型
- [x] 百分比轨迹校验（0.02 容差）
- [x] Redis store（原子一次性 take，既有兼容布局 `::captcha::` key）+ 内存 store
- [ ] 旋转/点选 challenge 类型（按前端实际使用再开，见协议文档 §8）
- [x] 行为校验（`WithTrackPolicy`，opt-in；默认与既有实现的位置校验一致）
- [x] 二次校验（`WithSecondaryVerification`：120s TTL，`secondary:` key）
- [ ] 安全评审
- [ ] v1.0 API 稳定化
