package practice

type Notifier interface {
	Send(message string) error
}

type BaseNotifier struct {
	Channel string
}

func (b BaseNotifier) Label(message string) string {
	return "[" + b.Channel + "] " + message
}

type EmailNotifier struct {
	BaseNotifier
	Address  string
	Messages []string
}

var _ Notifier = (*EmailNotifier)(nil)

type PushNotifier struct {
	BaseNotifier
	DeviceID string
	Messages []string
}

var _ Notifier = (*PushNotifier)(nil)

func (n *EmailNotifier) Send(message string) error {
	n.Messages = append(n.Messages, n.Label(message))
	return nil
}

func (n *PushNotifier) Send(message string) error {
	n.Messages = append(n.Messages, n.Label(message))
	return nil
}

func Dispatch(all []Notifier, message string) error {
	for _, n := range all {
		err := n.Send(message)
		if err != nil {
			return err
		}
	}
	return nil
}

func EmailAddress(n Notifier) (string, bool) {
	email, ok := n.(*EmailNotifier)
	if !ok || email == nil {
		return "", false
	}
	return email.Address, true
}

func Describe(n Notifier) string {
	switch v := n.(type) {
	case *EmailNotifier:
		if v != nil {
			return "email: " + v.Address
		}
	case *PushNotifier:
		if v != nil {
			return "push: " + v.DeviceID
		}
	}
	return "unknown"
}
