package squad

var recvHandlers map[string]func(source *RemoteSquadMember, content *string) error

func SetHandler(header string, handler func(*RemoteSquadMember, *string) error) {
	if recvHandlers == nil {
		recvHandlers = make(map[string]func(source *RemoteSquadMember, content *string) error)
	}
	recvHandlers[header] = handler
}
