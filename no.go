package gohc

type NoHealthCheck struct {
}

func NewNoHealthCheck() *NoHealthCheck {
	return &NoHealthCheck{}
}

func (h *NoHealthCheck) Check(host string) error {
	return nil
}

func (h *NoHealthCheck) String() string {
	return "NoHealthCheck"
}
