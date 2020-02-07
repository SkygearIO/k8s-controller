package controllers

import "time"

var (
	VerificationCooldown time.Duration = 60 * time.Second
	VerificationTimeout  time.Duration = 5 * time.Second
	PollInterval         time.Duration = 10 * time.Second
)
