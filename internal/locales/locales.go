package locales

import "fmt"

// Messages is a map of message keys to their localized strings
type Messages map[string]string

// Current holds the active message map (set by Init)
var Current *Messages

// Init sets the current locale based on the nsfw flag
func Init(nsfw bool) {
	if nsfw {
		Current = &NSFW
	} else {
		Current = &SFW
	}
}

// Msg retrieves a message by key and applies fmt.Sprintf with the given args
func Msg(key string, args ...any) string {
	if Current == nil {
		Current = &SFW
	}
	msg, ok := (*Current)[key]
	if !ok {
		return fmt.Sprintf("[missing: %s]", key)
	}
	if len(args) == 0 {
		return msg
	}
	return fmt.Sprintf(msg, args...)
}
