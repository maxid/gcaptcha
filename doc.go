// Package gcaptcha implements a self-contained slider captcha: image
// generation, one-time server-side validation and pluggable storage.
//
// Wire behavior follows the platform's documented gcaptcha contract: a
// background image plus a slider template, a challenge id, and percentage
// validation with the same 0.02 tolerance as the Java implementation. Protocol
// state and product semantics stay out of this package.
package gcaptcha
