package p2p

type HandShakeFunc func(any) error

func NOPHandShake(any) error {
	return nil
}
