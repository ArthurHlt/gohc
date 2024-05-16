package gohc

import (
	"fmt"
	"sync"
)

type Chains struct {
	hcs        []HealthChecker
	inParallel bool
	requireAll bool
}

func NewChains(inParallel bool, requireAll bool, hcs ...HealthChecker) *Chains {
	return &Chains{
		hcs:        hcs,
		inParallel: inParallel,
		requireAll: requireAll,
	}
}

func (c *Chains) Check(host string) error {
	if len(c.hcs) == 0 {
		return nil
	}
	if c.inParallel {
		return c.checkInParallel(host)
	}
	return c.checkInSeries(host)
}

func (c *Chains) checkInSeries(host string) error {
	var resultErr string
	oneSucceed := false
	for _, hc := range c.hcs {
		err := hc.Check(host)
		if err == nil {
			oneSucceed = true
			continue
		}
		if err != nil && c.requireAll {
			return fmt.Errorf("error on healthcheck '%v' for host '%s' : %w", hc, host, err)
		}
		if err != nil {
			resultErr = fmt.Sprintf("%s- %v: %s\n", resultErr, hc, err)
		}
	}
	if !oneSucceed {
		return fmt.Errorf("errors on healthchecks for host '%s':\n%s", host, resultErr)
	}
	return nil
}

func (c *Chains) checkInParallel(host string) error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error)
	for _, hc := range c.hcs {
		wg.Add(1)
		go func(host string, hc HealthChecker) {
			defer wg.Done()

			err := hc.Check(host)
			if err != nil {
				errCh <- fmt.Errorf("%v: %w", hc, err)
			}
		}(host, hc)
	}
	var resultErr string
	done := make(chan struct{})
	nbErrors := 0
	go func() {
		defer close(done)
		for err := range errCh {
			nbErrors++
			resultErr = resultErr + "- " + err.Error() + "\n"
		}
	}()
	wg.Wait()
	close(errCh)
	<-done
	if resultErr != "" && (c.requireAll || nbErrors == len(c.hcs)) {
		return fmt.Errorf("errors on healthchecks for host '%s':\n%s", host, resultErr)
	}
	return nil
}
