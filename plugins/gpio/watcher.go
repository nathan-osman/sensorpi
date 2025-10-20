package gpio

import "periph.io/x/conn/v3/gpio"

type gpioWatcher struct {
	pin      gpio.PinIO
	edgeChan chan any
}

func (w *gpioWatcher) run() {
	defer close(w.edgeChan)
	for {
		ok := w.pin.WaitForEdge(-1)
		if !ok {
			break
		}
		select {
		case w.edgeChan <- nil:
		default:
		}
	}
}

func newGpioWatcher(p gpio.PinIO) *gpioWatcher {
	w := &gpioWatcher{
		pin:      p,
		edgeChan: make(chan any),
	}
	go w.run()
	return w
}

func (w *gpioWatcher) Close() {

	// In() will cause WaitForEdge() to return false, triggering run() to
	// stop; then simply wait for the edgeChan to close, draining any other
	// values being sent

	w.pin.In(gpio.Float, gpio.NoEdge)
	ok := true
	for ok {
		_, ok = <-w.edgeChan
	}
}
