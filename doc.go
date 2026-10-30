// Package gcaptcha implements a self-contained slider captcha: image
// generation, one-time server-side validation and pluggable storage.
//
// Wire behavior follows the documented slider-captcha contract: a background
// image plus a slider template, a challenge id, and percentage validation with
// a 0.02 tolerance. Protocol state and product semantics stay out of this package.
package gcaptcha
