package mrpack

const (
	common string = "overrides"
	client string = "client-overrides"
	server string = "server-overrides"
)

func OverrideDirsClient() []string {
	return []string{common, client}
}

func OverrideDirsServer() []string {
	return []string{common, server}
}
