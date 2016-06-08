package main

import "time"

const (
	// CaptchaLength is the minimum captcha string length.
	CaptchaLength = 3
	// CaptchaVariation will add *up to* CaptchaVariation to the CaptchaLength
	CaptchaVariation = 2
	// CollectNum triggers a garbage collection routine after X captchas are created.
	CollectNum = 100
	Expiration = 10 * time.Minute
	StdWidth   = 240
	StdHeight  = 90
)
