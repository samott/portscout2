package types

type PortName struct {
	Category string
	Name     string
}

type PortInfo struct {
	Name    PortName
	Version string
}
