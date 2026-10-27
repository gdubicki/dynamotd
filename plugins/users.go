package plugins

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	"github.com/gdubicki/dynamotd/dynamotd"
	"github.com/shirou/gopsutil/v4/host"
)

func init() {
	dynamotd.Register(usersLabel, dynamotd.NoSettings{}, Users, "users")
}

const usersLabel = "Logged in"

// for mocking in tests
var getUsers = host.UsersWithContext

// Users shows who else is logged in, if anyone, to not f.e. restart something while
// someone else is working on it. (The session being opened is usually not listed yet.)
func Users(ctx context.Context, _ dynamotd.NoSettings) []dynamotd.Row {
	users, err := getUsers(ctx)
	// there is no utmp file f.e. in containers, which means that nobody is logged in
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	if err != nil {
		return []dynamotd.Row{dynamotd.ErrorRow(usersLabel, fmt.Errorf("error getting logged in users: %w", err))}
	}
	if len(users) == 0 {
		return nil
	}

	// group the sessions by user, keeping the order
	var names []string
	sessions := map[string]int{}
	for _, user := range users {
		if sessions[user.User] == 0 {
			names = append(names, user.User)
		}
		sessions[user.User]++
	}

	var value []dynamotd.ColorString
	for i, name := range names {
		if i > 0 {
			value = append(value, dynamotd.ValueDescription(", "))
		}
		value = append(value, dynamotd.ValueNeutral(name))
		if sessions[name] > 1 {
			value = append(value, dynamotd.ValueDescription(fmt.Sprintf(" (%d sessions)", sessions[name])))
		}
	}

	return []dynamotd.Row{{
		Label: dynamotd.SingleColorLabel(usersLabel),
		Value: dynamotd.ToColorText(value...),
	}}
}
