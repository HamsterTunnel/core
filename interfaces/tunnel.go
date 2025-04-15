package interfaces

type Tunnel interface {
	Start(userPort, clienPort string) error
	Stop() error
}
