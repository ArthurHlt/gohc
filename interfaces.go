package gohc

type HealthChecker interface {
	Check(host string) error
}
